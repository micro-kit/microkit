package client

import (
	"time"

	"github.com/micro-kit/microkit/plugins/middleware"
	"github.com/micro-kit/microkit/plugins/register"
)

/* 客户端配置 */

const (
	defaultConnTimeout = 5 * time.Second
)

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	serviceName string                  // 服务名
	middlewares []middleware.Middleware // 中间件
	reg         register.Register       // 注册中间件
	connTimeout time.Duration           // 连接超时
}

// ServiceName 服务名
func ServiceName(serviceName string) Option {
	return func(o *Options) {
		o.serviceName = serviceName
	}
}

// Middleware 服务中间件
func Middleware(middlewares ...middleware.Middleware) Option {
	return func(o *Options) {
		o.middlewares = middlewares
	}
}

// Register 服务注册中间件
func Register(reg register.Register) Option {
	return func(o *Options) {
		o.reg = reg
	}
}

// ConnTimeout 设置连接超时
func ConnTimeout(connTimeout time.Duration) Option {
	return func(o *Options) {
		o.connTimeout = connTimeout
	}
}
