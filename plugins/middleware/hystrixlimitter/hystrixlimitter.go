package hystrixlimitter

import (
	"log"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/micro-kit/microkit/plugins/middleware"
	"golang.org/x/time/rate"
)

// HystrixLimitter 熔断限流中间件
type HystrixLimitter struct {
	Options       *Options
	Limiter       *rate.Limiter // 限流器
	StreamLimiter *rate.Limiter // 流调用 限流器

}

// NewHystrixLimitter 创建熔断限流中间件
func NewHystrixLimitter(opts ...Option) middleware.Middleware {
	hl := &HystrixLimitter{
		Options: new(Options),
	}
	// 配置
	configure(hl, opts...)
	// 未设置日志对象退出
	if hl.Options.Logger == nil {
		log.Fatalln("限流熔断中间件未设置日志对象")
	}
	// 如果未设置类型，直接退出
	if hl.Options.Type == "" {
		hl.Options.Logger.Fatalw("未设置限流熔断中间件类型 server｜client", "options", hl.Options)
	}
	// 如果是服务端-初始化限流器
	if hl.Options.Type == HystrixLimitterTypeServer {
		// 普通rpc调用
		hl.Limiter = rate.NewLimiter(rate.Every(hl.Options.Limiter), hl.Options.LimiterBurst)
		// 流式调用
		hl.StreamLimiter = rate.NewLimiter(rate.Every(hl.Options.StreamLimiter), hl.Options.StreamLimiterBurst)
	}
	if hl.Options.Type == HystrixLimitterTypeClient {
		if hl.Options.ServiceName == "" {
			hl.Options.Logger.Fatalw("服务名不能为空", "options", hl.Options)
		}
		// 熔断器
		hystrix.ConfigureCommand(
			hl.Options.ServiceName, // 熔断器名字，可以用服务名称命名，一个名字对应一个熔断器，对应一份熔断策略
			hystrix.CommandConfig{
				Timeout:                hl.Options.Timeout,                // 超时时间 毫秒
				MaxConcurrentRequests:  hl.Options.MaxConcurrentRequests,  // 最大并发数，超过并发返回错误
				RequestVolumeThreshold: hl.Options.RequestVolumeThreshold, // 请求数量的阀值，用这些数量的请求来计算阀值
				ErrorPercentThreshold:  hl.Options.ErrorPercentThreshold,  // 错误率阀值，达到阀值，启动熔断 百分比
				SleepWindow:            hl.Options.SleepWindow,            // 熔断尝试恢复时间 毫秒
			},
		)
	}
	return hl
}

// 配置设置项
func configure(hl *HystrixLimitter, ops ...Option) {
	// 处理设置参数
	for _, o := range ops {
		o(hl.Options)
	}
	/* 默认值 */
	// 限流
	if hl.Options.Limiter <= 0 {
		hl.Options.Limiter = DefaultLimiter
	}
	if hl.Options.LimiterBurst <= 0 {
		hl.Options.LimiterBurst = DefaultLimiterBurst
	}
	if hl.Options.StreamLimiter <= 0 {
		hl.Options.StreamLimiter = DefaultStreamLimiter
	}
	if hl.Options.StreamLimiterBurst <= 0 {
		hl.Options.StreamLimiterBurst = DefaultStreamLimiterBurst
	}
	// 熔断
	if hl.Options.Timeout <= 0 {
		hl.Options.Timeout = hystrix.DefaultTimeout
	}
	if hl.Options.MaxConcurrentRequests <= 0 {
		hl.Options.MaxConcurrentRequests = hystrix.DefaultMaxConcurrent
	}
	if hl.Options.RequestVolumeThreshold <= 0 {
		hl.Options.RequestVolumeThreshold = hystrix.DefaultVolumeThreshold
	}
	if hl.Options.ErrorPercentThreshold <= 0 {
		hl.Options.ErrorPercentThreshold = hystrix.DefaultErrorPercentThreshold
	}
	if hl.Options.SleepWindow <= 0 {
		hl.Options.SleepWindow = hystrix.DefaultSleepWindow
	}
}
