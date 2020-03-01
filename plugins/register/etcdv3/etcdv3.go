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
func (er *EtcdV3) initEtcdCli() (err error) {
	er.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   er.Options.Addrs,
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		er.Options.Logger.Errorw("连接etcd服务错误", "err", err, "addrs", er.Options.Addrs)
		log.Fatalln("连接etcd服务错误", err)
		return err
	}
	return nil
}

// Register 注册一个服务
func (er *EtcdV3) Register(n *register.Node) error {
	if n == nil {
		return errors.New("服务注册信息不能为nil")
	}
	er.node = n
	// 定时器每隔多少秒续期一次
	ticker := time.NewTicker(time.Second * time.Duration(er.Options.TTL))
	// 服务存储地址 - 服务名/服务id
	er.srvKey = er.keyPrefix(n)
	// 协程内定时续期
	go func() {
		for {
			getResp, err := er.cli.Get(context.Background(), er.srvKey)
			if err != nil {
				er.Options.Logger.Warnw("续期服务注册信息，查询错误", "err", err, "srvKey", er.srvKey)
			} else if getResp.Count == 0 {
				err = er.withAlive()
				if err != nil {
					er.Options.Logger.Warnw("续期服务注册信息，续期错误", "err", err, "srvKey", er.srvKey)
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
func (er *EtcdV3) keyPrefix(n *register.Node) string {
	if n == nil {
		return fmt.Sprintf("microkit/services/%s/", er.Options.Name)
	}
	return fmt.Sprintf("microkit/services/%s/%s", er.Options.Name, n.Id)
}

// 续期
func (er *EtcdV3) withAlive() error {
	leaseResp, err := er.cli.Grant(context.Background(), er.Options.TTL)
	if err != nil {
		return err
	}
	// 服务注册信息 - 包含node全部信息
	addr, err := json.Marshal(er.node)
	if err != nil {
		return err
	}

	// 写服务信息
	_, err = er.cli.Put(context.Background(), er.srvKey, string(addr), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	_, err = er.cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}

// UnRegister 取消注册一个服务
func (er *EtcdV3) UnRegister(n *register.Node) error {
	_, err := er.cli.Delete(context.Background(), er.srvKey)
	if err != nil {
		er.Options.Logger.Warnw("删除服务注册信息错误", "err", err, "srvKey", er.srvKey)
	}
	return err
}

// GetResolver 获取客户端发现 grpc Resolver对象
func (er *EtcdV3) GetResolver() resolver.Resolver {
	return er
}

// GetBuilder 获取grpc服务注册Builder对象
func (er *EtcdV3) GetBuilder() resolver.Builder {
	return er
}

// Build grpc resolver接口需要 - 创建一个服务注册中心中间件
func (er *EtcdV3) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	var err error
	if er.cli == nil {
		// 创建etcd客户端对象
		err = er.initEtcdCli()
		if err != nil {
			er.Options.Logger.Fatalw("连接etcd服务错误", "err", err, "addrs", er.Options.Addrs)
			return nil, err
		}
	}
	er.cc = cc

	go er.watch(er.keyPrefix(nil))
	return er, nil
}

// Scheme 相当于服务分类名 - 服务注册根路径
func (er *EtcdV3) Scheme() string {
	schema := er.Options.Schema
	if schema == "" {
		schema = "default"
	}
	return schema
}

// ResolveNow 什么也不做
func (er *EtcdV3) ResolveNow(rn resolver.ResolveNowOption) {
	// log.Println("ResolveNow")
}

// Close 关闭注册解析器
func (er *EtcdV3) Close() {
	er.Options.Logger.Infow("关闭etcd服务解析器")
}

// watch 监听服务变化
func (er *EtcdV3) watch(keyPrefix string) {
	var addrList []resolver.Address
	getResp, err := er.cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		er.Options.Logger.Errorw("监听服务变化错误", "err", err, "keyPrefix", keyPrefix)
	} else {
		for _, kv := range getResp.Kvs {
			sn := new(register.Node)
			err = json.Unmarshal(kv.Value, sn)
			addrList = append(addrList, resolver.Address{Addr: sn.Advertise})
		}
	}
	er.cc.NewAddress(addrList)
	rch := er.cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())
	for n := range rch {
		for _, ev := range n.Events {
			// 解析服务注册的地址信息
			sn := new(register.Node)
			err = json.Unmarshal(ev.Kv.Value, sn)
			if err != nil {
				er.Options.Logger.Errorw("服务注册信息解析错误", "err", err, "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
				continue
			}
			// 客户端发现地址
			addr := sn.Advertise

			switch ev.Type {
			case mvccpb.PUT:
				if !exist(addrList, addr) {
					addrList = append(addrList, resolver.Address{Addr: addr})
					er.cc.NewAddress(addrList)
				}
			case mvccpb.DELETE:
				if s, ok := remove(addrList, addr); ok {
					addrList = s
					er.cc.NewAddress(addrList)
				}
			}
			er.Options.Logger.Infow("监听到etcd服务变化", "type", ev.Type, "key", string(ev.Kv.Key), "value", string(ev.Kv.Value))
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
