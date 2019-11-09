package server

import (
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/micro-kit/micro-common/config"
	"github.com/micro-kit/micro-common/logger"
	"github.com/micro-kit/microkit/internal/common"
	"github.com/micro-kit/microkit/plugins/middleware"
	"github.com/micro-kit/microkit/plugins/middleware/errorformat"
	"github.com/micro-kit/microkit/plugins/middleware/hystrixlimitter"
	zap "github.com/micro-kit/microkit/plugins/middleware/logger"
	"github.com/micro-kit/microkit/plugins/middleware/opentracing"
	"github.com/micro-kit/microkit/plugins/middleware/prometheus"
	"github.com/micro-kit/microkit/plugins/register"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

/*
 服务端
 可以创建服务实例，只需传入部分简单配置即可启动服务
*/

var (
	tracerCloser io.Closer // 链路追踪关闭
)

// Server 用于构造grpc服务，中间件加载等通用代码，简化应用创建过程
type Server struct {
	opts *Options
}

// NewDefaultServer 创建加载默认中间件服务
func NewDefaultServer() (*Server, error) {
	serviceName := config.GetSvcName()
	// 存储默认选项
	opts := make([]middleware.Middleware, 0)
	// 日志中间件
	opts = append(opts, zap.NewZapLogger(nil, zap.Logger(logger.Logger)))
	// 链路追踪中间件
	tracer, closer, err := common.NewJaegerTracer(serviceName)
	if err != nil {
		return nil, err
	}
	tracerCloser = closer
	opts = append(opts, opentracing.NewOpentracing(
		opentracing.Logger(logger.Logger),
		opentracing.Tracer(tracer),
	))
	// 监控中间件
	opts = append(opts, prometheus.NewPrometheus(prometheus.Enable(true)))
	// 熔断限流中间件
	opts = append(opts, hystrixlimitter.NewHystrixLimitter(
		hystrixlimitter.Type(hystrixlimitter.HystrixLimitterTypeServer),
		hystrixlimitter.Logger(logger.Logger),
		hystrixlimitter.ServiceName(serviceName),
	))

	return NewServer(Middleware(opts...), MetricsAddress(":19999"))
}

// NewServer 创建一个grpc服务对象
func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		opts: new(Options),
	}
	// 服务端加载日志整理
	opts = append(opts, Middleware(errorformat.NewErrorFormat()))
	// 配置
	configure(s, opts...)
	return s, nil
}

// 配置设置项
func configure(s *Server, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(s.opts)
	}
	// 默认值
	if s.opts.address == "" {
		s.opts.address = config.GetGRPCAddr()
	}
	if s.opts.advertise == "" {
		s.opts.advertise = config.GetGRPCAdvertiseAddr()
	}
	if s.opts.id == "" {
		s.opts.id = config.GetSvcID()
	}
	if s.opts.serviceName == "" {
		s.opts.serviceName = config.GetSvcName()
	}
	if s.opts.writeBufSize <= 0 {
		s.opts.writeBufSize = defaultWriteBufSize
	}
	if s.opts.readBufSize <= 0 {
		s.opts.readBufSize = defaultReadBufSize
	}
	if s.opts.reg == nil {
		// 默认注册到etcdv3
		s.opts.reg = common.NewServerEtcdRegister()
	}
}

// RegisterServer 注册服务回调 - 用于注册服务的grpc服务方法
type RegisterServer func(grpcServer *grpc.Server)

// Serve 启动服务
func (s *Server) Serve(regServer ...RegisterServer) error {
	if len(regServer) == 0 {
		return errors.New("At least one RegisterServer passed in")
	}
	// 监听
	lis, err := net.Listen("tcp", s.opts.address)
	if err != nil {
		return err
	}
	// 组织流式调用和普通调用插件列表
	streamInterceptors := make([]grpc.StreamServerInterceptor, 0)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0)
	for _, v := range s.opts.middlewares {
		streamInterceptors = append(streamInterceptors, v.StreamHandler)
		unaryInterceptors = append(unaryInterceptors, v.UnaryHandler)
	}
	// 创建grpc server并设置中间件
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(middleware.ChainStreamServer(streamInterceptors...)),
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(unaryInterceptors...)),
		grpc.WriteBufferSize(s.opts.writeBufSize),
		grpc.ReadBufferSize(s.opts.readBufSize),
	)
	// 注册grpc服务实现
	for _, regSrv := range regServer {
		regSrv(grpcServer)
	}
	// 注册服务
	err = s.Register()
	if err != nil {
		return err
	}
	// 启动普罗米修斯监控接口
	go s.Metrics()

	logger.Logger.Infow("启动服务", "service_name", s.opts.serviceName, "id", s.opts.id, "advertise", s.opts.advertise)
	// 启动服务
	return grpcServer.Serve(lis)
}

// Register 注册服务
func (s *Server) Register() error {
	return s.opts.reg.Register(&register.Node{
		Id:        s.opts.id,
		Address:   s.opts.address,
		Advertise: s.opts.advertise,
	})
}

// UnRegister 取消注册信息
func (s *Server) UnRegister() error {
	if tracerCloser != nil {
		tracerCloser.Close()
	}
	return s.opts.reg.UnRegister(&register.Node{
		Id:        s.opts.id,
		Address:   s.opts.address,
		Advertise: s.opts.advertise,
	})
}

// Stop 停止服务
func (s *Server) Stop() error {
	err := s.UnRegister()
	if err != nil {
		return err
	}
	return nil
}

// Metrics 普罗米修斯监控信息接口
func (s *Server) Metrics() {
	if s.opts.metricsAddress != "" {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(s.opts.metricsAddress, nil)
	}
}
