package main

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"net"
	"context"
	"github.com/pkg/errors"
	"fmt"
	"os"
	"github.com/sirupsen/logrus"
)

const (
	REGISTER_PREFIX = "/register"
	ETCD_IP = "172.31.0.245:2379"
)

var (
	endpoints []string
	etcdDialTimeout = 5000
	logName = "./register.log"
)

// 注册节点到etcd： /register/IP地址
type Register struct {
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
	localIP string // 本机IP
	log *logrus.Logger
}

// 获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs []net.Addr
		addr net.Addr
		ipNet *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()	// 192.168.1.1
				return
			}
		}
	}

	err = errors.New("Not found local ip")
	return
}

// 注册到/register/IP, 并自动续租
func (register *Register) keepOnline() {
	var (
		regKey string
		putResp *clientv3.PutResponse
		leaseGrantResp *clientv3.LeaseGrantResponse
		err error
		keepAliveChan <- chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
	)
	cancelCtx, cancelFunc := context.WithCancel(context.TODO())
	// 注册路径
	regKey = REGISTER_PREFIX + "/" + register.localIP
	//获取key
	getResp, err := register.kv.Get(context.TODO(), regKey)
	// 创建租约
	leaseCount := 0
	for {
		if leaseGrantResp, err = register.lease.Grant(cancelCtx, 10); err != nil {
			leaseCount++
			register.log.Warningf("这是第%d次创建租约", leaseCount)
			time.Sleep(1 * time.Second)
		} else {
			register.log.Infoln("创建租约成功")
			break
		}
		if leaseCount >= 3 {
			cancelFunc()
			register.log.Errorln("尝试创建租约3次后，仍然失败")
			os.Exit(1)
		}
	}
	if err != nil || getResp.Count == 0 {
		putCount := 0
		// 注册到etcd
		for {
			if putResp, err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
				putCount++
				register.log.Warningf("这是第%d次注册到etcd", putCount)
				time.Sleep(1 * time.Second)
			} else {
				register.log.Infof("注册成功,Reversion是%d", putResp.Header.Revision)
				break
			}
			if putCount >= 3 {
				cancelFunc()
				register.log.Errorln("尝试注册etcd 3次后，仍然失败")
				os.Exit(2)
			}
		}
	} else {
		keepAliveCount := 0
		for {
			// 自动续租
			if keepAliveChan, err = register.lease.KeepAlive(cancelCtx, leaseGrantResp.ID); err != nil {
				keepAliveCount++
				register.log.Warningf("这是第%d次申请租约", keepAliveCount)
				time.Sleep(1 * time.Second)
			} else {
				register.log.Infof("申请租约成功")
				break
			}
			if keepAliveCount >= 3 {
				cancelFunc()
				register.log.Errorln("尝试申请租约3次后，仍然失败")
				os.Exit(3)
			}
		}
		select {
			case keepAliveResp = <- keepAliveChan:
				if keepAliveResp == nil {
					register.log.Errorln("续约失败")
				} else {
					register.log.Infoln("续约成功")
				}
		}
	}
	//register.kv.Delete(context.TODO(), regKey)
	//fmt.Println(regKey)
	//fmt.Println(register.kv.Get(context.TODO(), regKey))
}

func main() {
	//日志配置
	log := logrus.New()
	logFd, err := os.OpenFile(logName, os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed")
		return
	}
	log.Out = logFd
	log.SetLevel(logrus.DebugLevel)
	defer logFd.Close()

	//etcd配置
	endpoints = append(endpoints, ETCD_IP)
	config := clientv3.Config{
		Endpoints: endpoints, // 集群地址
		DialTimeout: time.Duration(etcdDialTimeout) * time.Millisecond, // 连接超时
		//DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(config)
	fmt.Println("err:", err)
	fmt.Println("client:", client)
	if err != nil {
		log.Warningln("初始化etcd失败")
		return
	}

	// 本机IP
	localIp, err := getLocalIP()
	if err != nil {
		log.Warningln("获取本机IP失败")
		return
	}
	// 得到KV和Lease的API子集
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	fmt.Println("kv:", kv)
	register := &Register{
		client: client,
		kv: kv,
		lease: lease,
		localIP: localIp,
		log: log,
	}

	// 服务注册
	register.keepOnline()
}