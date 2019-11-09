package hystrixlimitter

import (
	"time"

	"github.com/micro-kit/microkit/plugins/middleware"
	"go.uber.org/zap"
)

// 默认值
const (
	// DefaultLimiter 默认3毫秒
	DefaultLimiter = 3 * time.Millisecond
	// DefaultLimiterBurst 默认缓存token数
	DefaultLimiterBurst = 100
	// DefaultStreamLimiter 流调用 默认3毫米
	DefaultStreamLimiter = 3 * time.Millisecond
	// DefaultStreamLimiterBurst 流调用 默认缓存token数
	DefaultStreamLimiterBurst = 3
	// HystrixLimitterTypeServer 服务端
	HystrixLimitterTypeServer = "server"
	// HystrixLimitterTypeClient 客户端
	HystrixLimitterTypeClient = "client"
)

// HystrixLimitterType 客户端还是服务端
type HystrixLimitterType string

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	Type          HystrixLimitterType
	FilterOutFunc middleware.FilterFunc
	Logger        *zap.SugaredLogger
	/* 限流 */
	Limiter            time.Duration // 限流器，多久生成一个token
	StreamLimiter      time.Duration // 流调用 限流器，多久生成一个token
	LimiterBurst       int           // 缓存token数量
	StreamLimiterBurst int           // 流调用 缓存token数量
	/* 熔断 */
	ServiceName            string // 服务名
	Timeout                int    // 单位毫秒
	MaxConcurrentRequests  int    // 最大并发数，超过并发返回错误
	RequestVolumeThreshold int    // 请求数量的阀值，用这些数量的请求来计算阀值
	ErrorPercentThreshold  int    // 错误数量阀值，达到阀值，启动熔断
	SleepWindow            int    // 熔断尝试恢复时间
}

// Type 设置是客户端还是服务端
func Type(typ HystrixLimitterType) Option {
	return func(o *Options) {
		o.Type = typ
	}
}

// Logger 设置日志对象
func Logger(logger *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// FilterOutFunc 设置中间件忽略函数列表
func FilterOutFunc(filterOutFunc middleware.FilterFunc) Option {
	return func(o *Options) {
		o.FilterOutFunc = filterOutFunc
	}
}

/* 限流 */

// Limiter 多久生成一个token
func Limiter(limiter time.Duration) Option {
	if limiter <= 0 {
		limiter = DefaultLimiter
	}
	return func(o *Options) {
		o.Limiter = limiter
	}
}

// LimiterBurst 限流缓存token数量
func LimiterBurst(limiterBurst int) Option {
	if limiterBurst <= 0 {
		limiterBurst = DefaultLimiterBurst
	}
	return func(o *Options) {
		o.LimiterBurst = limiterBurst
	}
}

// StreamLimiter 多久生成一个token
func StreamLimiter(limiter time.Duration) Option {
	if limiter <= 0 {
		limiter = DefaultStreamLimiter
	}
	return func(o *Options) {
		o.StreamLimiter = limiter
	}
}

// StreamLimiterBurst 限流缓存token数量
func StreamLimiterBurst(limiterBurst int) Option {
	if limiterBurst <= 0 {
		limiterBurst = DefaultStreamLimiterBurst
	}
	return func(o *Options) {
		o.StreamLimiterBurst = limiterBurst
	}
}

/* end 限流 */

/* 熔断 */

// ServiceName 服务名
func ServiceName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}

// Timeout 单位毫秒
func Timeout(t int) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// MaxConcurrentRequests 最大并发数，超过并发返回错误
func MaxConcurrentRequests(max int) Option {
	return func(o *Options) {
		o.MaxConcurrentRequests = max
	}
}

// RequestVolumeThreshold 请求数量的阀值，用这些数量的请求来计算阀值
func RequestVolumeThreshold(threshold int) Option {
	return func(o *Options) {
		o.RequestVolumeThreshold = threshold
	}
}

// ErrorPercentThreshold 错误数量阀值，达到阀值，启动熔断
func ErrorPercentThreshold(threshold int) Option {
	return func(o *Options) {
		o.ErrorPercentThreshold = threshold
	}
}

// SleepWindow 熔断尝试恢复时间
func SleepWindow(sleep int) Option {
	return func(o *Options) {
		o.SleepWindow = sleep
	}
}

/* end 熔断 */
