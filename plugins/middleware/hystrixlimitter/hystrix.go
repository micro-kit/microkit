package hystrixlimitter

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"google.golang.org/grpc"
)

/*
 客户端熔断 - 如果可以，可以在调用层并发使用 hystrix.Go 处理
 https://segmentfault.com/a/1190000012439580
 https://segmentfault.com/a/1190000015347065
*/

// UnaryClient 非流式客户端中间件 grpc.UnaryClientInterceptor
func (hl *HystrixLimitter) UnaryClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	if hl.Options.FilterOutFunc != nil && !hl.Options.FilterOutFunc(ctx, method) {
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}

	// 熔断
	err = hystrix.Do(hl.Options.ServiceName, func() error {
		// 执行下一步
		return invoker(ctx, method, req, reply, cc, opts...)
	}, func(err error) error {
		// 失败处理逻辑，访问其他资源失败时，或者处于熔断开启状态时，会调用这段逻辑
		// 可以简单构造一个response返回，也可以有一定的策略，比如访问备份资源
		// 也可以直接返回 err，这样不用和远端失败的资源通信，防止雪崩
		return err
	})
	if err != nil {
		return
	}
	return
}

// StreamClient 流式服客户中间件 grpc.StreamClientInterceptor
func (hl *HystrixLimitter) StreamClient(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (cs grpc.ClientStream, err error) {
	if hl.Options.FilterOutFunc != nil && !hl.Options.FilterOutFunc(ctx, method) {
		return streamer(ctx, desc, cc, method, opts...)
	}

	// TODO 流调用，暂时不处理

	cs, err = streamer(ctx, desc, cc, method, opts...)
	return
}
