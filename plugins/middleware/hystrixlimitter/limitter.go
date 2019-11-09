package hystrixlimitter

import (
	"context"
	"errors"

	"google.golang.org/grpc"
)

/*
 服务端限流
 https://segmentfault.com/a/1190000015347065
*/

var (
	// ErrLimitExceed 超出限流器限制
	ErrLimitExceed = errors.New("Rate limit exceed!")
)

// UnaryHandler 非流式中间件
func (hl *HystrixLimitter) UnaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if hl.Options.FilterOutFunc != nil && !hl.Options.FilterOutFunc(ctx, info.FullMethod) {
		resp, err = handler(ctx, req)
		return
	}

	// 限流
	if hl.Limiter.Allow() == false {
		return nil, ErrLimitExceed
	}

	// 执行下一步
	resp, err = handler(ctx, req)
	return
}

// StreamHandler 流式中间件
func (hl *HystrixLimitter) StreamHandler(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if hl.Options.FilterOutFunc != nil && !hl.Options.FilterOutFunc(stream.Context(), info.FullMethod) {
		err = handler(srv, stream)
		return
	}

	// 限流
	if hl.StreamLimiter.Allow() == false {
		return ErrLimitExceed
	}

	err = handler(srv, stream)
	return
}
