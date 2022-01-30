package config

import "time"

type Config struct {
	Port int
	//Elastic elastic.Config
	Consumer KafkaConsumer
	//Producer KafkaProducer
}

type KafkaConsumer struct {
	Brokers        []string
	Topic          string
	Group          string
	CommitInterval time.Duration
}

type KafkaProducer struct {
	Brokers []string
	Topic   string
}
