package register

import (
	"google.golang.org/grpc/resolver"
)

/* 服务插件定义 */

// Register 服务注册插件
type Register interface {
	Register(n *Node) error         // 注册一个服务
	UnRegister(n *Node) error       // 取消注册一个服务
	GetResolver() resolver.Resolver // 获取服务名对应服务列表
	GetBuilder() resolver.Builder   // 获取服务名对应服务列表
}

// Node 服务节点信息
type Node struct {
	Id        string            `json:"id"`        // 服务id
	Address   string            `json:"address"`   // 服务地址ip或域名+端口
	Advertise string            `json:"advertise"` // 服务注册发现地址
	Metadata  map[string]string `json:"metadata"`  // 服务注册附加信息，可以是系统资源信息 - TODO 当前采用轮训(grpc.WithBalancerName("round_robin"))后期可在此字段做文章
}
