package main

import (
	"registerServer/server"
	"github.com/astaxie/beego/logs"
	"time"
)

func main() {
	// 日志文件初始化
	err := server.InitConf()
	if err != nil {
		return
	}

	// 日志初始化
	err = server.InitLogger()
	if err != nil {
		logs.Error("Init log failed:", err)
		return
	}

	// 备份结果存储初始化
	err = server.InitStore()
	if err != nil {
		logs.Error("Init store failed:", err)
		return
	}

	// http服务初始化
	err = server.InitApiServer()
	if err != nil {
		logs.Error("Init http failed", err)
		return
	}

	// etcd初始化
	err = server.InitEtcd()
	if err != nil {
		logs.Error("Init etcd failed:", err)
		return
	}

	// mongo初始化
	err = server.InitMongo()
	if err != nil {
		logs.Error("Init mongo failed:", err)
		return
	}

	// 备份结果计算初始化
	err = server.InitCalculate()
	if err != nil {
		logs.Error("Init calculate failed:", err)
		return
	}

	// kafka初始化
	err = server.InitKafka()
	if err != nil {
		logs.Error("Init kafka failed:", err)
		return
	}

	// 备份结果比较初始化
	err = server.InitCompare()
	if err != nil {
		logs.Error("Init compare failed:", err)
		return
	}

	// redis初始化
	err = server.InitRedis()
	if err != nil {
		logs.Error("Init redis failed:", err)
		return
	}
	for {
		time.Sleep(1 * time.Second)
	}
	return
}
