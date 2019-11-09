package middleware

import (
	"context"

	"google.golang.org/grpc"
)

/* grpc中间件 */

// Middleware 通用拦截器中间件
type Middleware interface {
	// 非流式服务端中间件 grpc.UnaryServerInterceptor
	UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	// 流式服务端中间件 grpc.StreamServerInterceptor
	StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
	// 非流式客户端中间件 grpc.UnaryClientInterceptor
	UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	// 流式服客户中间件 grpc.StreamClientInterceptor
	StreamClient(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error)
}

// FilterFunc 过滤函数，如果返回false则中间件不处理
type FilterFunc func(ctx context.Context, fullMethodName string) bool
