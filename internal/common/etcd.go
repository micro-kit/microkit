package common

import (
	"strings"

	"github.com/micro-kit/micro-common/config"
	"github.com/micro-kit/micro-common/logger"
	"github.com/micro-kit/microkit/plugins/register"
	"github.com/micro-kit/microkit/plugins/register/etcdv3"
)

// NewServerEtcdRegister 获取一个etcd注册中间件 - 配置从环境变量获取
func NewServerEtcdRegister() register.Register {
	etcdAddrs := strings.Split(config.GetETCDAddr(), ";")
	reg := etcdv3.NewRegistry(
		register.Addrs(etcdAddrs...),
		register.Username(config.GetETCDUser()),
		register.Password(config.GetETCDPassword()),
		register.TTL(config.GetRegisterTTL()),
		register.Name(config.GetSvcName()),
		register.Schema(config.GetSchema()),
		register.Logger(logger.Logger),
	)
	return reg
}

// NewClientEtcdRegister 创建客户端注册对象
func NewClientEtcdRegister(svcName string) register.Register {
	etcdAddrs := strings.Split(config.GetETCDAddr(), ";")
	reg := etcdv3.NewRegistry(
		register.Addrs(etcdAddrs...),
		register.Username(config.GetETCDUser()),
		register.Password(config.GetETCDPassword()),
		register.TTL(config.GetRegisterTTL()),
		register.Name(svcName),
		register.Schema(config.GetSchema()),
		register.Logger(logger.Logger),
	)
	return reg
}
