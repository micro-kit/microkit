package tpl

import (
	"context"

	"github.com/micro-kit/microkit/plugins/middleware"
	"google.golang.org/grpc"
)

/* 中间件 */

// Tpl 中间件
type Tpl struct {
	Options *Options
}

// NewTpl 创建链路追踪器
func NewTpl(opts ...Option) middleware.Middleware {
	tpl := &Tpl{
		Options: new(Options),
	}
	// 配置
	configure(tpl, opts...)
	return tpl
}

// 配置设置项
func configure(tpl *Tpl, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(tpl.Options)
	}
}

// UnaryHandler 非流式中间件
func (tpl *Tpl) UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if tpl.Options.FilterOutFunc != nil && !tpl.Options.FilterOutFunc(ctx, info.FullMethod) {
		resp, err = handler(ctx, req)
		return
	}

	// 执行下一步
	resp, err = handler(ctx, req)
	return
}

// StreamHandler 流式中间件
func (tpl *Tpl) StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if tpl.Options.FilterOutFunc != nil && !tpl.Options.FilterOutFunc(stream.Context(), info.FullMethod) {
		err = handler(srv, stream)
		return
	}

	err = handler(srv, stream)
	return
}

// UnaryClient 非流式客户端中间件 grpc.UnaryClientInterceptor
func (tpl *Tpl) UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	if tpl.Options.FilterOutFunc != nil && !tpl.Options.FilterOutFunc(ctx, method) {
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}

	// 执行下一步
	err = invoker(ctx, method, req, reply, cc, opts...)
	return
}

// StreamClient 流式服客户中间件 grpc.StreamClientInterceptor
func (tpl *Tpl) StreamClient(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	if tpl.Options.FilterOutFunc != nil && !tpl.Options.FilterOutFunc(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}

	cs, err = streamer(ctx, desc, cc, method, opts...)
	return
}
