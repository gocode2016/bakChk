package server

import (
	"net/http"
	"net"
	"strconv"
	"time"
	"registerServer/common"
	"golang.org/x/net/context"
	"github.com/astaxie/beego/logs"
	"github.com/bitly/go-simplejson"
	"encoding/json"
)

type SingleBackup struct {
	BackupName string	`json:"backupName"`
	BackupSize float64	`json:"backupSize"`
}

type BatchBackup struct {
	IP string `json:"ip"`
	SingleBackup []SingleBackup `json:"singleBackup"`
}

func listBackup(resp http.ResponseWriter, req *http.Request) {
	data, err := G_redis.getData()
	var backupList []BatchBackup
	if data == "Nil" && err == nil {
		if len(backupList) > 0 {
			backupList = nil
		}
		if len(G_store.IpList) > 0 {
			for _, ip := range G_store.IpList {
				var batchBackup BatchBackup
				filter := &common.FindByHost{
					Host: *ip,
				}
				cur, err := G_mongo.collection.Find(context.TODO(), filter)
				if err != nil {
					logs.Warning("mongo select fail,", err)
				}
				defer cur.Close(context.TODO())
				batchBackup.IP = *ip
				for cur.Next(context.TODO()) {
					result := &common.BackupStatistics{}
					if err = cur.Decode(result); err != nil {
						logs.Warning("Decode backup err,", err)
						continue
					}
					var singleBackup SingleBackup
					singleBackup.BackupName = result.BackupName
					singleBackup.BackupSize = result.BackupSize
					batchBackup.SingleBackup = append(batchBackup.SingleBackup, singleBackup)
				}
				backupList = append(backupList, batchBackup)
			}
		}
		bytes, err := common.BuildResponse(0, "success", backupList)
		G_redis.setData(string(bytes))
		if err == nil {
			resp.Write(bytes)
		}
	} else if len(data) != 0 && err == nil {
		var response common.Response
		js, _ := simplejson.NewJson([]byte(data))
		if err != nil {
			logs.Warning("unmarshal backup list fail,", err)
			return
		}
		err := json.Unmarshal([]byte(js.Interface().(string)), &response)
		backupList := response.Data
		//fmt.Println("backupList:", backupList)
		bytes, err := common.BuildResponse(0, "success", backupList)
		if err == nil {
			resp.Write(bytes)
		}
	}

}

func InitApiServer() (err error) {
	//配置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/index", listBackup)

	//静态目录配置
	webDir := http.Dir(G_config.webRoot)
	webHandler := http.FileServer(webDir)
	mux.Handle("/", http.StripPrefix("/", webHandler))

	//启动tcp监听
	listener, err := net.Listen("tcp", ":" + strconv.Itoa(G_config.apiPort))
	if err != nil {
		return
	}
	//创建一个HTTP服务
	httpServer := &http.Server{
		ReadTimeout: time.Duration(G_config.apiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.apiWriteTimeout) * time.Millisecond,
		Handler: mux,
	}

	go httpServer.Serve(listener)
	return
}
