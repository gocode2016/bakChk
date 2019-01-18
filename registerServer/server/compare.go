package server

import (
	"fmt"
	"time"
	"golang.org/x/net/context"
	"registerServer/common"
	"github.com/gorhill/cronexpr"
	"github.com/astaxie/beego/logs"
	"strings"
)

type Compare struct {

}

var (
	G_compare *Compare
)

func InitCompare() (err error) {
	G_compare = &Compare{}
	go G_compare.execDaily()
	return
}

func (compare *Compare) compareBackupSize() {
	go G_etcd.ListServer()
	for {
		select {
			case server := <- G_store.ClientList:
				getBackupSize(server)
		}
	}
}

//{ "_id" : ObjectId("5c1241b54c2c3da8d7f0741b"), "host" : "10.10.10.20", "backupName" : "20181213.zip", "backupSize" : 1.6, "backupTime" : ISODate("2018-12-13T11:25:41.439Z") }
//{ "_id" : ObjectId("5c1241b54c2c3da8d7f0741c"), "host" : "10.10.20.30", "backupName" : "20181213.zip", "backupSize" : 1.6, "backupTime" : ISODate("2018-12-13T11:25:41.439Z") }
//{ "_id" : ObjectId("5c1241b54c2c3da8d7f0741d"), "host" : "10.10.20.40", "backupName" : "20181213.zip", "backupSize" : 1.6, "backupTime" : ISODate("2018-12-13T11:25:41.439Z") }
func getBackupSize(ip string) {
	var allBackupSize  []map[string][]map[string]float64
	filter := &common.FindByHost{
		Host: ip,
	}
	cur, err := G_mongo.collection.Find(context.TODO(), filter)
	if err != nil {
		logs.Warning("mongo select fail,", err)
	}

	allBackupSizeMap := make(map[string][]map[string]float64)
	//defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var singleBackupSizeMap map[string]float64
		singleBackupSizeMap = make(map[string]float64)
		result := &common.BackupStatistics{}
		if err = cur.Decode(result); err != nil {
			logs.Warning("Decode backup err,", err)
			continue
		}
		if len(result.BackupName) > 0 {
			singleBackupSizeMap[result.BackupName] = result.BackupSize
			allBackupSizeMap[result.Host] = append(allBackupSizeMap[result.Host], singleBackupSizeMap)
		}
	}
	allBackupSize = append(allBackupSize, allBackupSizeMap)
	if len(allBackupSize) > 0 {
		for _, v := range allBackupSize {
			for k1, v1 := range v {
				var flag1, flag2 bool
				var nowBackupSize, yesterdayBackupSize, beforeYesterdayBackupSize float64
				nowBackupName := fmt.Sprintf("%s.zip", common.Now)
				yesterdayBackupName := fmt.Sprintf("%s.zip", common.Yesterday)
				beforeYesterdayBackupName := fmt.Sprintf("%s.zip", common.BeforeYesterday)
				for _, v2 := range v1 {
					for k3, _ := range v2 {
						if strings.TrimSpace(k3) == nowBackupName {
							nowBackupSize,_ = v2[nowBackupName]
						} else if strings.TrimSpace(k3) == yesterdayBackupName {
							yesterdayBackupSize,_ = v2[yesterdayBackupName]
						} else if strings.TrimSpace(k3) == beforeYesterdayBackupName {
							beforeYesterdayBackupSize,_ = v2[beforeYesterdayBackupName]
						}
					}
				}
				if yesterdayBackupSize != 0 && nowBackupSize != 0  && yesterdayBackupSize > nowBackupSize {
					maxPercent := ((yesterdayBackupSize - nowBackupSize)/yesterdayBackupSize) * 100
					if maxPercent > G_config.maxPercent {
						flag1 = true
					}
				}
				if beforeYesterdayBackupSize != 0 && nowBackupSize != 0  && beforeYesterdayBackupSize > nowBackupSize {
					maxPercent := ((beforeYesterdayBackupSize - nowBackupSize)/beforeYesterdayBackupSize) * 100
					if maxPercent > G_config.maxPercent {
						flag2 = true
					}
				}
				if flag1 && flag2 {
					alertMsg := fmt.Sprintf("IP:%s, today's backup size is at least %f of percent smaller than the previous two days.", k1, G_config.maxPercent)
					G_store.AlertChan <- alertMsg
				}
			}

		}
	}
}

func (compare *Compare) execDaily() {
	go func() {
		for {
			expr := cronexpr.MustParse(G_config.executionTimeDaily)
			now := time.Now()
			nextTime := expr.Next(now)
			logs.Info("The next execution time is ", nextTime)
			time.AfterFunc(nextTime.Sub(now), func(){
				compare.compareBackupSize()
			})
			select {
				case <- time.NewTimer(24 * 60 * 60 * time.Second).C:
			}
		}
	}()
}
