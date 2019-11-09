package opentracing

import (
	"github.com/micro-kit/microkit/plugins/middleware"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	FilterOutFunc middleware.FilterFunc
	Logger        *zap.SugaredLogger
	Tracer        opentracing.Tracer
}

// Logger 设置日志对象
func Logger(logger *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// Tracer 链路追踪客户端
func Tracer(tracer opentracing.Tracer) Option {
	return func(o *Options) {
		o.Tracer = tracer
	}
}

// FilterOutFunc 设置中间件忽略函数列表
func FilterOutFunc(filterOutFunc middleware.FilterFunc) Option {
	return func(o *Options) {
		o.FilterOutFunc = filterOutFunc
	}
}
