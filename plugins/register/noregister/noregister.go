package noregister

import (
	"github.com/micro-kit/microkit/plugins/register"
	"google.golang.org/grpc/resolver"
)

/* 不需要注册到注册中心的服务 */

// NoRegister 不需要注册到注册中心的服务
type NoRegister struct {
	Options *register.Options // 注册服务配置
}

// NewRegistry 创建一个etcdv3服务注册中间件
func NewRegistry(ops ...register.Option) register.Register {
	return &NoRegister{}
}

// Register 注册一个服务
func (nr *NoRegister) Register(n *register.Node) error {
	return nil
}

// UnRegister 取消注册一个服务
func (nr *NoRegister) UnRegister(n *register.Node) error {
	return nil
}

// GetResolver 获取客户端发现 grpc Resolver对象
func (nr *NoRegister) GetResolver() resolver.Resolver {
	return nr
}

// GetBuilder 获取grpc服务注册Builder对象
func (nr *NoRegister) GetBuilder() resolver.Builder {
	return nr
}

// Close 关闭注册解析器
func (nr *NoRegister) Close() {
}

// ResolveNow 什么也不做
func (nr *NoRegister) ResolveNow(rn resolver.ResolveNowOption) {
}

// Build grpc resolver接口需要 - 创建一个服务注册中心中间件
func (nr *NoRegister) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	return nr, nil
}

// Scheme 相当于服务分类名 - 服务注册根路径
func (nr *NoRegister) Scheme() string {
	schema := nr.Options.Schema
	if schema == "" {
		schema = "default"
	}
	return schema
}
