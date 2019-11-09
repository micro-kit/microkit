package prometheus

import (
	"context"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/micro-kit/microkit/plugins/middleware"
	"google.golang.org/grpc"
)

/* 普罗米修斯指标监控中间件 */

// Prometheus 普罗米修斯指标监控中间件
type Prometheus struct {
	Options *Options
}

func NewPrometheus(opts ...Option) middleware.Middleware {
	p := &Prometheus{
		Options: new(Options),
	}
	// 配置
	configure(p, opts...)

	return p
}

// 配置设置项
func configure(p *Prometheus, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(p.Options)
	}
}

/* 服务端拦截器 */

// UnaryHandler 非流式中间件
func (p *Prometheus) UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if p.Options.FilterOutFunc != nil && !p.Options.FilterOutFunc(ctx, info.FullMethod) {
		resp, err = handler(ctx, req)
		return
	}

	// github.com/grpc-ecosystem/go-grpc-prometheus
	resp, err = grpc_prometheus.UnaryServerInterceptor(ctx, req, info, handler)
	return
}

// StreamHandler 流式中间件
func (p *Prometheus) StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if p.Options.FilterOutFunc != nil && !p.Options.FilterOutFunc(stream.Context(), info.FullMethod) {
		err = handler(srv, stream)
		return
	}
	// github.com/grpc-ecosystem/go-grpc-prometheus
	err = grpc_prometheus.StreamServerInterceptor(srv, stream, info, handler)
	return
}

/* 以下客户端拦截器 */

// UnaryClient 非流式客户端中间件 grpc.UnaryClientInterceptor
func (p *Prometheus) UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	if p.Options.FilterOutFunc != nil && !p.Options.FilterOutFunc(ctx, method) {
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}

	// github.com/grpc-ecosystem/go-grpc-prometheus
	err = grpc_prometheus.UnaryClientInterceptor(ctx, method, req, reply, cc, invoker, opts...)
	return
}

// StreamClient 流式服客户中间件 grpc.StreamClientInterceptor
func (p *Prometheus) StreamClient(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	if p.Options.FilterOutFunc != nil && !p.Options.FilterOutFunc(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}
	// github.com/grpc-ecosystem/go-grpc-prometheus
	cs, err = grpc_prometheus.StreamClientInterceptor(ctx, desc, cc, method, streamer, opts...)
	return
}
