package client

import (
	"errors"
	"io"

	"github.com/micro-kit/micro-common/logger"
	"github.com/micro-kit/micro-common/opentrac"
	"github.com/micro-kit/microkit/internal/common"
	"github.com/micro-kit/microkit/plugins/middleware"
	"github.com/micro-kit/microkit/plugins/middleware/hystrixlimitter"
	zap "github.com/micro-kit/microkit/plugins/middleware/logger"
	"github.com/micro-kit/microkit/plugins/middleware/opentracing"
	"github.com/micro-kit/microkit/plugins/middleware/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

/*
 客户端
 用于创建客户端对象，传入基本配置即可创建一个客户端
*/

// Client 客户端对象
type Client struct {
	opts         *Options
	conn         *grpc.ClientConn // grpc 连接
	tracerCloser io.Closer        // 链路追踪关闭
}

// NewDefaultClient 创建默认客户端对象
func NewDefaultClient(opts ...Option) (*Client, error) {
	c := &Client{
		opts: new(Options),
	}
	// 配置
	configure(c, opts...)
	// 服务名检查
	if c.opts.serviceName == "" {
		return nil, errors.New("Service name must be set")
	}

	// 存储默认选项
	middlewares := make([]middleware.Middleware, 0)
	// 日志中间件
	middlewares = append(middlewares, zap.NewZapLogger(nil, zap.Logger(logger.Logger)))
	// 链路追踪中间件
	tracer, closer, err := opentrac.NewJaegerTracer(c.opts.serviceName)
	if err != nil {
		return nil, err
	}
	c.tracerCloser = closer // 保存链路追踪关闭对象
	middlewares = append(middlewares, opentracing.NewOpentracing(
		opentracing.Logger(logger.Logger),
		opentracing.Tracer(tracer),
	))
	// 监控中间件
	middlewares = append(middlewares, prometheus.NewPrometheus(prometheus.Enable(true)))
	// 熔断限流中间件
	middlewares = append(middlewares, hystrixlimitter.NewHystrixLimitter(
		hystrixlimitter.Type(hystrixlimitter.HystrixLimitterTypeClient),
		hystrixlimitter.Logger(logger.Logger),
		hystrixlimitter.ServiceName(c.opts.serviceName),
	))

	// 再次配置
	opts = append(opts, Middleware(middlewares...))
	configure(c, opts...)

	return c, nil
}

// NewClient 创建客户端对象
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		opts: new(Options),
	}
	// 配置
	configure(c, opts...)
	// 服务名检查
	if c.opts.serviceName == "" {
		return nil, errors.New("Service name must be set")
	}

	return c, nil
}

// 配置设置项
func configure(c *Client, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(c.opts)
	}
	// 默认值
	if c.opts.connTimeout <= 0 {
		c.opts.connTimeout = defaultConnTimeout
	}
	if c.opts.reg == nil {
		c.opts.reg = common.NewClientEtcdRegister(c.opts.serviceName)
	}
}

// NewGrpcClient 用户生成客户端对象
type NewGrpcClient func(*grpc.ClientConn)

// Dial 连接grpc服务端，并构造多客户端
func (c *Client) Dial(clients ...NewGrpcClient) error {
	// 注册中间件获取客户端发现对象
	r := c.opts.reg.GetBuilder()
	resolver.Register(r)
	// grpc连接配置
	opts := make([]grpc.DialOption, 0)
	opts = append(opts,
		grpc.WithBalancerName("round_robin"), // 轮训调用
		grpc.WithInsecure(),                  // 禁用连接安全检查
		grpc.WithTimeout(c.opts.connTimeout), // 连接超时
	)
	// 连接服务
	conn, err := grpc.Dial(r.Scheme()+"://author/"+c.opts.serviceName, opts...)
	if err != nil {
		return err
	}
	c.conn = conn
	// 初始化grpc客户端
	for _, cc := range clients {
		cc(conn)
	}
	return nil
}

// Close 关闭连接 - 服务结束时
func (c *Client) Close() {
	// grpc 连接
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			logger.Logger.Errorw("关闭grpc客户端连接错误", "err", err, "service_name", c.opts.serviceName)
		}
	}
	// 链路追踪
	if c.tracerCloser != nil {
		err := c.tracerCloser.Close()
		if err != nil {
			logger.Logger.Errorw("关闭链路追踪错误", "err", err, "service_name", c.opts.serviceName)
		}
	}
}
