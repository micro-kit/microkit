package errorformat

import (
	"github.com/micro-kit/microkit/plugins/middleware"
)

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	FilterOutFunc middleware.FilterFunc
}

// FilterOutFunc 设置中间件忽略函数列表
func FilterOutFunc(filterOutFunc middleware.FilterFunc) Option {
	return func(o *Options) {
		o.FilterOutFunc = filterOutFunc
	}
}
