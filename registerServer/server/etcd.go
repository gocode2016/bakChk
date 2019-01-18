package server

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"registerServer/common"
	"golang.org/x/net/context"
	"strings"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"github.com/astaxie/beego/logs"
)

type Etcd struct {
	kv clientv3.KV
	client *clientv3.Client
	watcher clientv3.Watcher
}

var G_etcd *Etcd

func InitEtcd() (err error) {
	etcdUri := G_config.etcdUri
	etcdDialTimeout := G_config.etcdDialTimeout
	endpoints := []string{etcdUri}
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		DialTimeout: time.Duration(etcdDialTimeout) * time.Millisecond,
	})
	if err != nil {
		return
	}
	kv := clientv3.NewKV(client)
	watcher := clientv3.NewWatcher(client)
	G_etcd = &Etcd{
		kv: kv,
		client: client,
		watcher: watcher,
	}
	go G_etcd.watchServers()
	go G_etcd.ListServer()
	return
}

func (etcd *Etcd) watchServers() (err error) {
	getResp, err := etcd.kv.Get(context.TODO(), common.SERVER_REGISTER_DIR, clientv3.WithPrefix())
	if err != nil {
		logs.Error("Error getting this prefix,", err)
		return
	}

	for _, kvpair := range getResp.Kvs {
		value := strings.TrimPrefix(string(kvpair.Key), common.SERVER_REGISTER_DIR)
		G_store.PushServer(value)
	}
	//从当前reversion监听变化事件
	go func() {
		watchStartRevision := getResp.Header.Revision + 1
		//监听/cron/jobs目录后续的变化
		watchChan := etcd.watcher.Watch(context.TODO(), common.SERVER_REGISTER_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp := range watchChan {
			for _, watchEvent := range watchResp.Events {
				value := strings.TrimPrefix(string(watchEvent.Kv.Key), common.SERVER_REGISTER_DIR)
				switch watchEvent.Type {
					//PUT事件监听，并反序列化job
					case mvccpb.PUT:
						logs.Info("Etcd add value ", value)
						G_store.PushServer(value)
						G_etcd.ListServer()
					//DELETE事件监听
					case mvccpb.DELETE:
						logs.Info("Etcd del value ", value)
						G_store.DelServer(value)
						G_etcd.ListServer()
				}
			}
		}
	}()
	return
}

func (etcd *Etcd) ListServer() {
	getResp, err := etcd.kv.Get(context.TODO(), common.SERVER_REGISTER_DIR, clientv3.WithPrefix())
	if err != nil {
		logs.Error("Error getting this prefix with listing server,", err)
		return
	}
	G_store.IpList = nil
	for _, kvpair := range getResp.Kvs {
		value := strings.TrimPrefix(string(kvpair.Key), common.SERVER_REGISTER_DIR)
		G_store.ClientList <- value
		G_store.IpList = append(G_store.IpList, &value)
	}
}


