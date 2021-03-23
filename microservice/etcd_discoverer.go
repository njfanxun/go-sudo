package microservice

import (
	"context"
	"math/rand"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdDiscoverer struct {
	cli        *clientv3.Client
	prefixPath string
	watcher    clientv3.Watcher
	mu         sync.Mutex
	data       map[string]map[string]string
}

var (
	d                  *EtcdDiscoverer
	onceEtcdDiscoverer sync.Once
)

func InitDiscoverer(cfg clientv3.Config, prefixPath string) {
	onceEtcdDiscoverer.Do(func() {
		var err error
		d = &EtcdDiscoverer{
			prefixPath: prefixPath,
			mu:         sync.Mutex{},
			data:       make(map[string]map[string]string),
		}
		d.cli, err = clientv3.New(cfg)
		if err != nil {
			panic(err)
		}
		d.watcher = clientv3.NewWatcher(d.cli)
		go d.run()
		d.initServicesTable()
	})
}

/** @Description:  初始化读取etcd注册的服务列表*/
func (e *EtcdDiscoverer) initServicesTable() {
	resp, err := e.cli.Get(context.TODO(), e.prefixPath, clientv3.WithPrefix())
	if err != nil {
		logrus.Errorf("发现注册服务错误:%s", err.Error())
		return
	}
	defer e.mu.Unlock()
	e.mu.Lock()
	for _, kv := range resp.Kvs {
		p := string(kv.Key)
		addr := path.Base(p)
		serviceName := path.Base(path.Dir(p))
		e.data[serviceName] = make(map[string]string)
		e.data[serviceName][addr] = string(kv.Value)
	}
}

/** @Description:  开启etcd的服务发现监视 调用时请开启协程*/
func (e *EtcdDiscoverer) run() {
	watchChan := e.watcher.Watch(context.Background(), e.prefixPath, clientv3.WithPrefix())
	for {
		select {
		case resp := <-watchChan:
			for _, event := range resp.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					e.put(string(event.Kv.Key), string(event.Kv.Value))
					break
				case clientv3.EventTypeDelete:
					e.del(string(event.Kv.Key))
					break
				}
			}
		}
	}
}

func (e *EtcdDiscoverer) put(p string, v string) {
	defer e.mu.Unlock()
	e.mu.Lock()
	addr := path.Base(p)
	serviceName := path.Base(path.Dir(p))
	_, found := e.data[serviceName]
	if !found {
		e.data[serviceName] = make(map[string]string)
	}
	e.data[serviceName][addr] = v
}
func (e *EtcdDiscoverer) del(p string) {
	defer e.mu.Unlock()
	e.mu.Lock()
	addr := path.Base(p)
	serviceName := path.Base(path.Dir(p))
	delete(e.data[serviceName], addr)
	if len(e.data[serviceName]) == 0 {
		delete(e.data, serviceName)
	}
}

func Disc() *EtcdDiscoverer {
	if d == nil {
		panic("use etcd discoverer instance must be call initFunc")
	}
	return d
}

//** @Description:  根据前缀开启服务发现*/
func (e *EtcdDiscoverer) GetServiceAddress(serviceName string) (string, error) {
	defer e.mu.Unlock()
	e.mu.Lock()
	m, found := e.data[serviceName]
	if !found || len(m) == 0 {
		return "", errors.Errorf("No %s service registration...", serviceName)
	}
	var list []string
	for addr, _ := range m {
		list = append(list, addr)
	}
	return randomPolicy(list), nil
}

func randomPolicy(list []string) string {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return list[r.Intn(len(list))]
}
