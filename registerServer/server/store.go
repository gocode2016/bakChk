package server

import "registerServer/common"

type Store struct {
	ClientChan chan string
	DelClientChan chan string
	AlertChan chan string
	ResultChan chan common.BackupStatistics
	ClientList chan string
	IpList []*string
}

var G_store *Store

func InitStore() (err error) {
	G_store = &Store{
		ClientChan: make(chan string, 1000),
		DelClientChan: make(chan string, 1000),
		AlertChan: make(chan string, 1000),
		ResultChan: make(chan common.BackupStatistics, 1000),
		ClientList: make(chan string, 1000),
	}
	return
}

func (store *Store) PushServer(ip string) {
	store.ClientChan <- ip
}

func (store *Store) DelServer(ip string) {
	store.DelClientChan <- ip
}
