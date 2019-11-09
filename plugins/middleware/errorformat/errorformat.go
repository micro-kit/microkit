package errorformat

import (
	"context"

	"github.com/micro-kit/micro-common/microerror"
	"github.com/micro-kit/microkit/plugins/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/* 返回错误码处理 */

// ErrorFormat 错误码处理
type ErrorFormat struct {
	Options *Options
}

// NewErrorFormat 统一返回错误码处理
func NewErrorFormat(opts ...Option) middleware.Middleware {
	ef := &ErrorFormat{
		Options: new(Options),
	}
	configure(ef, opts...)

	return ef
}

// 配置设置项
func configure(ef *ErrorFormat, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(ef.Options)
	}
}

// UnaryHandler 非流式中间件
func (ef *ErrorFormat) UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if ef.Options.FilterOutFunc != nil && !ef.Options.FilterOutFunc(ctx, info.FullMethod) {
		resp, err = handler(ctx, req)
		return
	}

	// 返回错误重新整理
	defer func() {
		err = ef.serverError(err)
	}()

	// 执行下一步
	resp, err = handler(ctx, req)
	return
}

// StreamHandler 流式中间件
func (ef *ErrorFormat) StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if ef.Options.FilterOutFunc != nil && !ef.Options.FilterOutFunc(stream.Context(), info.FullMethod) {
		err = handler(srv, stream)
		return
	}

	// 返回错误重新整理
	defer func() {
		err = ef.serverError(err)
	}()

	err = handler(srv, stream)
	return
}

// UnaryClient 非流式客户端中间件 grpc.UnaryClientInterceptor
func (ef *ErrorFormat) UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	if ef.Options.FilterOutFunc != nil && !ef.Options.FilterOutFunc(ctx, method) {
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}

	// 执行下一步
	err = invoker(ctx, method, req, reply, cc, opts...)
	return
}

// StreamClient 流式服客户中间件 grpc.StreamClientInterceptor
func (ef *ErrorFormat) StreamClient(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	if ef.Options.FilterOutFunc != nil && !ef.Options.FilterOutFunc(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}

	cs, err = streamer(ctx, desc, cc, method, opts...)
	return
}

// 处理服务端返回错误
func (ef *ErrorFormat) serverError(err error) error {
	if err == nil {
		return nil
	}
	// 断言是否是microerror.MicroError
	microError, ok := err.(*microerror.MicroError)
	if !ok {
		// 不是microerror.MicroError则设置系统错误
		microError = microerror.GetMicroError(microerror.UnknownServerError, err)
	}
	return status.New(codes.Code(microError.Code), microError.Msg).Err()
}
