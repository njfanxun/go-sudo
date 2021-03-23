package grpc

import (
	"net"
	"strconv"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

/** @Description: Server is a gRPC server wrapper */
type Server struct {
	*grpc.Server
	port int
	lis  net.Listener
	log  *logrus.Logger
}

func NewServer(port int) *Server {
	srv := &Server{
		port: port,
		log:  logrus.New(),
	}
	srv.log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	entry := logrus.NewEntry(srv.log)
	logOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}
	grpcOpts := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_logrus.StreamServerInterceptor(entry, logOpts...),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(entry, logOpts...),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	}
	srv.Server = grpc.NewServer(grpcOpts...)
	return srv
}

/** @Description:  注册grpc的实现类*/
func (s *Server) RegisterService(desc *grpc.ServiceDesc, ss interface{}) {
	s.Server.RegisterService(desc, ss)
}

/** @Description:  server start running */
func (s *Server) Start() {
	var err error
	s.lis, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(s.port))
	if err != nil {
		panic(err)
	}
	go func() {
		err = s.Serve(s.lis)
		if err != nil {
			panic(err)
		}
	}()
	s.log.Infof("[gRPC] server listening on: %s", s.lis.Addr().String())
}

func (s *Server) Stop() {
	s.GracefulStop()
	s.log.Info("[gRPC] server stopping")
}
