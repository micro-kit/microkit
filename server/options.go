package server

import (
	"github.com/micro-kit/microkit/plugins/middleware"
	"github.com/micro-kit/microkit/plugins/register"
)

/* 服务配置 */

const (
	// 默认读写缓冲区 128KB
	defaultWriteBufSize = 128 * 1024
	defaultReadBufSize  = 128 * 1024
)

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	id          string                  // 服务id
	serviceName string                  // 服务名
	address     string                  // 监听地址
	advertise   string                  // 服务注册地址
	middlewares []middleware.Middleware // 中间件
	reg         register.Register       // 注册中间件

	writeBufSize int // 写缓冲区
	readBufSize  int // 读缓冲区

	metricsAddress string // 普罗米修斯监控信息接口
}

// Address 监听地址
func Address(address string) Option {
	return func(o *Options) {
		o.address = address
	}
}

// Advertise 服务注册地址
func Advertise(advertise string) Option {
	return func(o *Options) {
		o.advertise = advertise
	}
}

// Middleware 服务中间件
func Middleware(middlewares ...middleware.Middleware) Option {
	return func(o *Options) {
		if o.middlewares == nil {
			o.middlewares = make([]middleware.Middleware, 0)
		}
		o.middlewares = append(o.middlewares, middlewares...)
	}
}

// Register 服务注册中间件
func Register(reg register.Register) Option {
	return func(o *Options) {
		o.reg = reg
	}
}

// ID 服务唯一id
func ID(id string) Option {
	return func(o *Options) {
		o.id = id
	}
}

// ServiceName 服务名
func ServiceName(serviceName string) Option {
	return func(o *Options) {
		o.serviceName = serviceName
	}
}

// WriteBufSize 写缓冲区
func WriteBufSize(writeBufSize int) Option {
	return func(o *Options) {
		o.writeBufSize = writeBufSize
	}
}

// ReadBufSize 读缓冲区
func ReadBufSize(readBufSize int) Option {
	return func(o *Options) {
		o.readBufSize = readBufSize
	}
}

// MetricsAddress 普罗米修斯监控信息接口
func MetricsAddress(metricsAddress string) Option {
	return func(o *Options) {
		o.metricsAddress = metricsAddress
	}
}
