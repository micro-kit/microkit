package etcdv3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/micro-kit/microkit/plugins/register"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
)

/* 注册服务到etcd v3 */

// EtcdV3 注册服务到etcd v3 api的服务注册中间件
type EtcdV3 struct {
	Options *register.Options   // 注册服务配置
	cli     *clientv3.Client    // etcdv3 客户端
	srvKey  string              // 服务注册key
	node    *register.Node      // 注册节点信息
	cc      resolver.ClientConn // grpc 客户端
}

// NewRegistry 创建一个etcdv3服务注册中间件
func NewRegistry(ops ...register.Option) register.Register {
	etcdV3 := &EtcdV3{
		Options: new(register.Options),
	}
	// 配置
	configure(etcdV3, ops...)
	// 创建etcd客户端对象
	err := etcdV3.initEtcdCli()
	if err != nil {
		etcdV3.Options.Logger.Fatalw("连接etcd服务错误", "err", err, "addrs", etcdV3.Options.Addrs)
	}
	return etcdV3
}

// 配置设置项
func configure(etcdV3 *EtcdV3, ops ...register.Option) {
	for _, o := range ops {
		o(etcdV3.Options)
	}
	// 默认值
	if len(etcdV3.Options.Addrs) == 0 {
		etcdV3.Options.Addrs = append(etcdV3.Options.Addrs, "127.0.0.1:2379")
	}
	if etcdV3.Options.TTL <= 0 {
		etcdV3.Options.TTL = 10
	}
}

// 初始化etcd服务连接
func (p *EtcdV3) initEtcdCli() (err error) {
	p.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   p.Options.Addrs,
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		p.Options.Logger.Errorw("连接etcd服务错误", "err", err, "addrs", p.Options.Addrs)
		log.Fatalln("连接etcd服务错误", err)
		return err
	}
	return nil
}

// Register 注册一个服务
func (p *EtcdV3) Register(n *register.Node) error {
	if n == nil {
		return errors.New("服务注册信息不能为nil")
	}
	p.node = n
	// 定时器每隔多少秒续期一次
	ticker := time.NewTicker(time.Second * time.Duration(p.Options.TTL))
	// 服务存储地址 - 服务名/服务id
	p.srvKey = p.keyPrefix(n)
	// 协程内定时续期
	go func() {
		for {
			getResp, err := p.cli.Get(context.Background(), p.srvKey)
			if err != nil {
				p.Options.Logger.Warnw("续期服务注册信息，查询错误", "err", err, "srvKey", p.srvKey)
			} else if getResp.Count == 0 {
				err = p.withAlive()
				if err != nil {
					p.Options.Logger.Warnw("续期服务注册信息，续期错误", "err", err, "srvKey", p.srvKey)
				}
			} else {
				// 存在则什么也做
			}
			<-ticker.C
		}
	}()
	return nil
}

// 获取服务存储key前缀
func (p *EtcdV3) keyPrefix(n *register.Node) string {
	if n == nil {
		return fmt.Sprintf("microkit/services/%s/", p.Options.Name)
	}
	return fmt.Sprintf("microkit/services/%s/%s", p.Options.Name, n.Id)
}

// 续期
func (p *EtcdV3) withAlive() error {
	leaseResp, err := p.cli.Grant(context.Background(), p.Options.TTL)
	if err != nil {
		return err
	}
	// 服务注册信息 - 包含node全部信息
	addr, err := json.Marshal(p.node)
	if err != nil {
		return err
	}

	// 写服务信息
	_, err = p.cli.Put(context.Background(), p.srvKey, string(addr), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	_, err = p.cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister 取消注册一个服务
func (p *EtcdV3) UnRegister(n *register.Node) error {
	_, err := p.cli.Delete(context.Background(), p.srvKey)
	if err != nil {
		p.Options.Logger.Warnw("删除服务注册信息错误", "err", err, "srvKey", p.srvKey)
	}
	return err
}

// GetResolver 获取客户端发现 grpc Resolver对象
func (p *EtcdV3) GetResolver() resolver.Resolver {
	return p
}

// GetBuilder 获取grpc服务注册Builder对象
func (p *EtcdV3) GetBuilder() resolver.Builder {
	return p
}

// Build grpc resolver接口需要 - 创建一个服务注册中心中间件
func (p *EtcdV3) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	var err error
	if p.cli == nil {
		// 创建etcd客户端对象
		err = p.initEtcdCli()
		if err != nil {
			p.Options.Logger.Fatalw("连接etcd服务错误", "err", err, "addrs", p.Options.Addrs)
			return nil, err
		}
	}
	p.cc = cc

	go p.watch(p.keyPrefix(nil))
	return p, nil
}

// Scheme 相当于服务分类名 - 服务注册根路径
func (p *EtcdV3) Scheme() string {
	schema := p.Options.Schema
	if schema == "" {
		schema = "default"
	}
	return schema
}

// ResolveNow 什么也不做
func (p *EtcdV3) ResolveNow(rn resolver.ResolveNowOption) {
	// log.Println("ResolveNow")
}

// Close 关闭注册解析器
func (p *EtcdV3) Close() {
	p.Options.Logger.Infow("关闭etcd服务解析器")
}

// watch 监听服务变化
func (p *EtcdV3) watch(keyPrefix string) {
	var addrList []resolver.Address
	getResp, err := p.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		p.Options.Logger.Errorw("监听服务变化错误", "err", err, "keyPrefix", keyPrefix)
	} else {
		for _, kv := range getResp.Kvs {
			sn := new(register.Node)
			err = json.Unmarshal(kv.Value, sn)
			addrList = append(addrList, resolver.Address{Addr: sn.Advertise})
		}
	}
	p.cc.NewAddress(addrList)
	rch := p.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			// 解析服务注册的地址信息
			sn := new(register.Node)
			err = json.Unmarshal(ev.Kv.Value, sn)
			if err != nil {
				p.Options.Logger.Errorw("服务注册信息解析错误", "err", err, "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
				continue
			}
			// 客户端发现地址
			addr := sn.Advertise

			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, addr) {
					addrList = append(addrList, resolver.Address{Addr: addr})
					p.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					addrList = s
					p.cc.NewAddress(addrList)
				}
			}
			p.Options.Logger.Infow("监听到etcd服务变化", "type", ev.Type, "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
		}
	}
}

// 解析器是否已存在
func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

// 删除解析器
func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}
