package tpl

import (
	"github.com/micro-kit/microkit/plugins/middleware"
	"go.uber.org/zap"
)

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	FilterOutFunc middleware.FilterFunc
	Logger        *zap.SugaredLogger
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
