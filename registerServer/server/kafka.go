package server

import (
	"github.com/Shopify/sarama"
	"fmt"
	"github.com/astaxie/beego/logs"
)

type Kafka struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	server string
}

var (
	G_kafka *Kafka
)

func InitKafka() (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	kafkaUri := G_config.kafkaUri
	producer, err := sarama.NewSyncProducer([]string{kafkaUri}, config)
	if err != nil {
		return
	}

	consumer, err := sarama.NewConsumer([]string{kafkaUri}, nil)
	if err != nil {
		return
	}
	G_kafka = &Kafka{
		producer: producer,
		consumer: consumer,
		server: kafkaUri,
	}
	go G_kafka.produce()
	//go G_kafka.consume()
	return
}

func (kafka *Kafka) produce() {
	msg := &sarama.ProducerMessage{}
	for {
		select {
			case message := <- G_store.AlertChan:
				msg.Topic = G_config.kafkaTopic
				msg.Value = sarama.StringEncoder(message)
				_,_,err := kafka.producer.SendMessage(msg)
				if err != nil {
					logs.Error("send message to consumer fail,", err)
					continue
				}
		}
	}

	//循环判断partition是否发送过来成功commited的ACK消息.
	//go func(p sarama.AsyncProducer) {
	//	for{
	//		select {
	//			case  suc := <- p.Successes():
	//				fmt.Println("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
	//			case fail := <- p.Errors():
	//				fmt.Println("err: ", fail.Err)
	//			}
	//	}
	//}(kafka.producer)
	//
	//for {
	//	select {
	//		case message := <- G_store.AlertChan:
	//			msg := &sarama.ProducerMessage{
	//				Topic: G_config.kafkaTopic,
	//			}
	//			//将字符串转化为字节数组
	//			msg.Value = sarama.ByteEncoder(message)
	//			//使用通道发送
	//			kafka.producer.Input() <- msg
	//	}
	//}
}

func (kafka *Kafka) consume() {
	consumer, err := sarama.NewConsumer([]string{kafka.server}, nil)
	if err != nil {
		panic(err)
	}

	partitionList, err := consumer.Partitions(G_config.kafkaTopic)
	if err != nil {
		panic(err)
	}

	for partition := range partitionList {
		pc, err := consumer.ConsumePartition(G_config.kafkaTopic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		defer pc.AsyncClose()
		go func(sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				logs.Info("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				SendEmail(string(msg.Value))
			}
		}(pc)
	}
}

