package register

import (
	"crypto/tls"

	"go.uber.org/zap"
)

/* 服务注册参数 */

// Option 实例值设置
type Option func(*Options)

// Options 注册相关参数
type Options struct {
	Schema    string      // 服务注册根路径
	Name      string      // 服务名
	Addrs     []string    // 服务注册中间件地址
	TTL       int64       // 注册信息生存有效期
	Secure    bool        // 是否启用安全连接
	TLSConfig *tls.Config // tls加密连接配置
	Logger    *zap.SugaredLogger
}

// Addrs 服务注册中间件地址
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// Secure 服务中间件是否使用加密连接
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// TLSConfig TLS 加密连接配置
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// TTL 注册信息生存有效期
func TTL(t int64) Option {
	return func(o *Options) {
		o.TTL = t
	}
}

// Name 设置服务名
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Schema 服务注册根路径
func Schema(s string) Option {
	return func(o *Options) {
		o.Schema = s
	}
}

// Logger 设置日志对象
func Logger(logger *zap.SugaredLogger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}
