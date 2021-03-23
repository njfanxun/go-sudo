package microservice

import (
	"context"
	"encoding/json"
	"path"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistrar struct {
	cli       *clientv3.Client
	leaseId   clientv3.LeaseID
	lease     clientv3.Lease
	closeChan chan int
	opt       *ServiceOption
}

/** @Description:  新建etcd的服务注册器*/
func NewEtcdRegistrar(cfg clientv3.Config, opt *ServiceOption) *EtcdRegistrar {
	var err error
	r := &EtcdRegistrar{
		opt:       opt,
		closeChan: make(chan int),
	}
	r.cli, err = clientv3.New(cfg)
	if err != nil {
		panic(err)
	}
	r.lease = clientv3.NewLease(r.cli)
	return r
}

func (r *EtcdRegistrar) RegisterService() error {
	// 请求租约
	r.leaseId = 0
	leaseResponse, err := r.lease.Grant(context.TODO(), 2*int64(r.opt.heartbeat.Seconds()))
	if err != nil {
		err = errors.Wrapf(err, "[Register] lease Grant err:%s", err.Error())
		return err
	}
	// 在etcd中注册服务信息
	// k: /{基础路径}/{服务使用版本}/{服务名}/{服务调用地址}
	// v: 自定的meta信息 以json字符串存储
	addr := r.opt.host + ":" + strconv.Itoa(r.opt.port)
	key := path.Join(r.opt.basePath, r.opt.mode, r.opt.name, addr)
	value, err := json.Marshal(r.opt.metadata)
	if err != nil {
		err = errors.Wrapf(err, "[metadata] register err:%s", err.Error())
		return err
	}
	_, err = r.cli.KV.Put(context.TODO(), key, string(value), clientv3.WithLease(leaseResponse.ID))
	if err != nil {
		err = errors.Wrapf(err, "[Register] register err:%s", err.Error())
		return err
	}
	r.leaseId = leaseResponse.ID
	go func() {
		r.keepAlive()
	}()
	logrus.Infof("[%s service] has been registered in etcd", r.opt.name)
	return nil
}

func (r *EtcdRegistrar) keepAlive() {
	t := time.NewTicker(r.opt.heartbeat)
	for {
		select {
		case <-t.C:
			_, err := r.lease.KeepAliveOnce(context.TODO(), r.leaseId)
			// 如果租约id丢失
			if err == rpctypes.ErrLeaseNotFound {
				t.Stop()
				goto Restart
			}
		case <-r.closeChan:
			t.Stop()
			goto Exit
		}

	}
Restart:
	r.RegisterService()
Exit:
	r.cli.Close()
	close(r.closeChan)

}

// 注销服务
func (r *EtcdRegistrar) DeregisterService() {
	if r.lease != nil {
		_, _ = r.lease.Revoke(context.TODO(), r.leaseId)
	}
	r.closeChan <- 1
	logrus.Infof("[%s service] has been deregistered in etcd", r.opt.name)
}
