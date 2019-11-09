package prometheus

import (
	"github.com/micro-kit/microkit/plugins/middleware"
)

/* 服务注册参数 */

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	Enable        bool // 是否启用监控
	FilterOutFunc middleware.FilterFunc
}

// Enable 是否启用普罗米修斯指标采集中间件
func Enable(enable bool) Option {
	return func(o *Options) {
		o.Enable = enable
	}
}

// FilterOutFunc 设置中间件忽略函数列表
func FilterOutFunc(filterOutFunc middleware.FilterFunc) Option {
	return func(o *Options) {
		o.FilterOutFunc = filterOutFunc
	}
}
