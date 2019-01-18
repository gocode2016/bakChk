package common

import (
	"time"
	"encoding/json"
)

const (
	SERVER_REGISTER_DIR = "/register/"
	EMAIL_SUBJECT = "BACKUP ALERT!!!"
)

var (
	Now = time.Now().Format("20060102")
	Yesterday = time.Now().AddDate(0, 0, -1).Format("20060102")
	BeforeYesterday = time.Now().AddDate(0, 0, -2).Format("20060102")
)

type FindByHost struct {
	Host string `bson:"host"`
}

type FindByBackupName struct {
	BackupName string `bson:"backupName"`
}

// mongo存放结果
type BackupStatistics struct {
	Host string `json:"host" bson:"host"`                      //IP
	BackupName string `json:"backupName" bson:"backupName"`    //备份名称
	BackupSize float64 `json:"backupSize" bson:"backupSize"`    //备份大小
	BackupTime time.Time `json:"backupTime" bson:"backupTime"` //备份统计时间
}

// HTTP接口应答
type Response struct {
	Errno int `json:"errno"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

// 应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	// 1, 定义一个response
	var (
		response Response
	)

	response.Errno = errno
	response.Msg = msg
	response.Data = data

	// 2, 序列化json
	resp, err = json.Marshal(response)
	return
}

//func RemoveServer(server string, serverList []interface{}) []interface{} {
//	if len(serverList) > 0 {
//		var index int
//		for i, value := range serverList {
//			if strings.TrimSpace(value.(string)) == strings.TrimSpace(server) {
//				index = i
//			}
//		}
//		serverList = append(serverList[:index], serverList[index+1:]...)
//		return serverList
//	}
//	return nil
//}

//func DelChanServer(server string, ClientList chan *string, serverTempList []*string) (chan *string) {
//	if len(ClientList) > 0 {
//		count := 0
//		for {
//			s := <- ClientList
//			if len(ClientList) == 0 {
//				break
//			}
//			if strings.TrimSpace(*s) != strings.TrimSpace(server) {
//				serverTempList[count] = s
//				count++
//			}
//			fmt.Println("3333", serverTempList, len(ClientList))
//		}
//		fmt.Println("2222", serverTempList, len(ClientList))
//		for _, s := range serverTempList {
//			if s != nil {
//				ClientList <- s
//			}
//		}
//		for v := range ClientList {
//			fmt.Println("v:", *v)
//		}
//		return ClientList
//	}
//	return nil
//}
