package server

import (
	"github.com/astaxie/beego/config"
)

type Config struct {
	logLevel string
	logPath string
	etcdUri string
	etcdDialTimeout int
	kafkaUri string
	kafkaTopic string
	mongoUri string
	mongoTimeout int
	mongoDatabase string
	mongoCollection string
	batchResult int
	backupDir string
	maxPercent float64
	smtpServer string
	username string
	password string
	sender string
	receiver string
	executionTimeDaily string
	batchTimeout int
	webRoot string
	staticRoot string
	apiPort int
	apiReadTimeout int
	apiWriteTimeout int
	redisUri string
	redisThread int
	expireTime int
}

var G_config *Config

func InitConf() (err error){
	conf, err := config.NewConfig("ini", "./config/server.conf")
	if err != nil {
		return
	}
	logLevel := conf.String("log::log_level")
	logPath := conf.String("log::log_path")
	etcdUri := conf.String("etcd::etcdUri")
	etcdDialTimeout, err := conf.Int("etcd::timeout")
	if err != nil {
		etcdDialTimeout = 5000
	}
	kafkaUri := conf.String("kafka::kafkaUri")
	kafkaTopic := conf.String("kafka::topic")
	backupDir := conf.String("backup::directory")
	maxPercent, err := conf.Float("backup::maxPercent")
	if err != nil {
		maxPercent = 10
	}
	mongoUri := conf.String("mongodb::mongoUri")
	mongoTimeout, err := conf.Int("mongodb::timeout")
	if err != nil {
		return
	}
	mongoDatabase := conf.String("mongodb::database")
	mongoCollection := conf.String("mongodb::collection")
	batchResult, err := conf.Int("mongodb::batchResult")
	if err != nil {
		return
	}
	smtpServer := conf.String("email::smtpserver")
	username := conf.String("email::username")
	password := conf.String("email::password")
	sender := conf.String("email::sender")
	receiver := conf.String("email::receiver")
	executionTimeDaily := conf.String("crontab::execution_time_daily")
	batchTimeout, err := conf.Int("mongo::batchTimeout")
	if err != nil {
		batchTimeout = 1000
	}
	webRoot := conf.String("http::webRoot")
	staticRoot := conf.String("http::staticRoot")
	apiPort, err := conf.Int("http::apiPort")
	if err != nil {
		apiPort = 8888
	}
	apiReadTimeout, err := conf.Int("http::apiReadTimeout")
	if err != nil {
		apiReadTimeout = 5000
	}
	apiWriteTimeout, err := conf.Int("http::apiWriteTimeout")
	if err != nil {
		apiWriteTimeout = 5000
	}
	redisUri := conf.String("redis::redisUri")
	expireTime, err := conf.Int("redis::expireTime")
	if err != nil {
		expireTime = 600
	}
	redisThread, err := conf.Int("redis::redisThread")
	if err != nil {
		redisThread = 10
	}
	G_config = &Config{
		logLevel: logLevel,
		logPath: logPath,
		etcdUri: etcdUri,
		etcdDialTimeout: etcdDialTimeout,
		kafkaUri: kafkaUri,
		kafkaTopic: kafkaTopic,
		backupDir: backupDir,
		maxPercent: maxPercent,
		mongoUri: mongoUri,
		mongoTimeout: mongoTimeout,
		mongoDatabase: mongoDatabase,
		mongoCollection: mongoCollection,
		batchResult: batchResult,
		smtpServer: smtpServer,
		username: username,
		password: password,
		sender: sender,
		receiver: receiver,
		executionTimeDaily: executionTimeDaily,
		batchTimeout: batchTimeout,
		webRoot: webRoot,
		staticRoot: staticRoot,
		apiPort: apiPort,
		apiReadTimeout: apiReadTimeout,
		apiWriteTimeout: apiWriteTimeout,
		redisUri: redisUri,
		redisThread: redisThread,
		expireTime: expireTime,
	}
	return
}


