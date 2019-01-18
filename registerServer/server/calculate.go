package server

import (
	"fmt"
	"time"
	"os"
	"registerServer/common"
	"github.com/astaxie/beego/logs"
	"strconv"
)

type Calculate struct {

}

var G_calculate *Calculate

func InitCalculate() (err error) {
	G_calculate = &Calculate{}
	go G_calculate.handleServerEvent()
	return
}

func (calculate *Calculate) handleServerEvent() {
	for {
		select {
			case ip := <- G_store.ClientChan:
				fmt.Println("calculate ip:", ip)
				calculate.calculate(ip)
			case ip := <- G_store.DelClientChan:
				alertContent := fmt.Sprintf("%s has been deleted", ip)
				G_store.AlertChan <- alertContent
		}
	}
}

// 备份计算结果发送到结果成功管道，不存在的则发送到报警管道
func (calculate *Calculate) calculate(ip string) {
	now := time.Now().Format("20060102")
	//now := time.Now().AddDate(0, 0, -2).Format("20060102")
	backupDir := G_config.backupDir
	backupName := fmt.Sprintf("%s/%s/%s.zip", backupDir, ip, now)
	if checkBackupExist(backupName) {
		s, _ := os.Stat(backupName)
		backupSize := s.Size()
		newBackupSize := convertToMB(float64(backupSize))
		backupStatistics := common.BackupStatistics{
			Host: ip,
			BackupName: fmt.Sprintf("%s.zip", now),
			BackupSize: newBackupSize,
			BackupTime: time.Now(),
		}
		G_store.ResultChan <- backupStatistics
	} else {
		alertContent := fmt.Sprintf("Not found %s", backupName)
		logs.Warning("Not found %s", backupName)
		G_store.AlertChan <- alertContent
	}
}

// 单位转换
func convertToMB(backupSize float64) (float64) {
	newBackupSize, err := strconv.ParseFloat(fmt.Sprintf("%.2f", backupSize/1024/1024), 64)
	fmt.Println(newBackupSize)
	if err != nil {
		return 0.0
	}
	return 	newBackupSize
}

// 判断备份是否存在
func checkBackupExist(backupName string) bool {
	_, err := os.Stat(backupName)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
