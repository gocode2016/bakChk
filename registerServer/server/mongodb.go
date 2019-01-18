package server

import (
	"golang.org/x/net/context"
	"time"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/astaxie/beego/logs"
	"registerServer/common"
	"fmt"
)

type Mongo struct {
	client *mongo.Client
	collection *mongo.Collection
	autoCommitChan chan *common.BackupStatistics
	results []interface{}
}

var G_mongo *Mongo

func InitMongo() (err error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(G_config.mongoTimeout) * time.Millisecond)
	client, err := mongo.Connect(ctx, G_config.mongoUri)
	collection := client.Database(G_config.mongoDatabase).Collection(G_config.mongoCollection)
	G_mongo = &Mongo{
		client: client,
		collection: collection,
		autoCommitChan: make(chan *common.BackupStatistics, 1000),
	}
	err = G_mongo.deleteTodayBackup()
	if err == nil {
		go G_mongo.handleBackupResult()
	}
	return
}

func (mongo *Mongo) handleBackupResult() {
	var (
		backupStatistics *common.BackupStatistics
		commitTimer *time.Timer
		timeoutBatch *common.BackupStatistics
	)

	for {
		select {
			case result := <- G_store.ResultChan:
				//超时1s就提交
				if backupStatistics == nil {
					backupStatistics = &common.BackupStatistics{}
					commitTimer = time.AfterFunc(
						time.Duration(time.Duration(G_config.batchTimeout) * time.Millisecond),
						func(backupStatistics *common.BackupStatistics) func() {
							return func() {
								mongo.autoCommitChan <- backupStatistics
							}
						}(backupStatistics),
					)
				}
				mongo.results = append(mongo.results, &result)
				//批量结果的条数超过定义的个数就提交
				if len(mongo.results) >= G_config.batchResult {
					mongo.insertBacthResult()
					mongo.results = nil
					commitTimer.Stop()
				}
			case timeoutBatch = <- mongo.autoCommitChan:
				//过期批次已经过期的情况处理
				if timeoutBatch != backupStatistics {
					continue
				}
				mongo.insertBacthResult()
				backupStatistics = nil
		}
	}
}

func (mongo *Mongo) insertBacthResult() {
	res, err := mongo.collection.InsertMany(context.TODO(), mongo.results)
	if err != nil {
		logs.Error("Insert backup result err,", err)
	}
	logs.Info("The Insert ID of this insertion is ", res.InsertedIDs)
}

func (mongo *Mongo) deleteTodayBackup() (err error){
	filter := &common.FindByBackupName{
		BackupName: fmt.Sprintf("%s.zip", common.Now),
	}
	deleteResult, err := mongo.collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		logs.Warning("delete %s failed from mongodb", fmt.Sprintf("%s.zip", common.Now))
		return
	}
	fmt.Println("delete result:", deleteResult.DeletedCount)
	return
}