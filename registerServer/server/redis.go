package server

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/astaxie/beego/logs"
	"time"
)

type Redis struct {
	pool *pool.Pool
}

var G_redis *Redis

func InitRedis() (err error) {
	pool, err := pool.New("tcp", G_config.redisUri, G_config.redisThread)
	if err != nil {
		logs.Warning("Connect redis pool fail,", err)
		return
	}
	G_redis = &Redis{
		pool: pool,
	}
	go func() {
		for {
			//每3秒ping一次保持链接
			pool.Cmd("PING")
			time.Sleep(3 * time.Second)
		}
	}()
	return
}

func (r *Redis) setData(data string) (err error) {
	resp := r.pool.Cmd("SET", "backupList", data, "EX", G_config.expireTime)
	if resp.Err != nil {
		logs.Warning("redis set data fail,", resp.Err)
		return
	}
	return
}

func (r *Redis) getData() (string, error) {
	 resp := r.pool.Cmd("GET", "backupList")
	if resp.Err != nil {
		logs.Warning("redis get data fail,", resp.Err)
		return "", resp.Err
	}
	return resp.String(), nil
}
