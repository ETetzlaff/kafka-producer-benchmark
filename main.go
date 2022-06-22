package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	KafkaTopic            = "KAFKA_TOPIC"
	KafkaBootstrapServers = "KAFKA_BOOTSTRAP_SERVERS"
	BufferSize            = "BUFFER_SIZE"
	Amount                = "AMOUNT"
)

func init() {
	viper.SetDefault(KafkaTopic, "default-topic")
	viper.SetDefault(KafkaBootstrapServers, "127.0.0.1:9092")
	viper.SetDefault(BufferSize, 10)
	viper.SetDefault(Amount, 10)

	viper.BindEnv(KafkaTopic)
	viper.BindEnv(KafkaBootstrapServers)
	viper.BindEnv(BufferSize)
	viper.BindEnv(Amount)
}

func main() {
	ctx := context.Background()

	brokers := viper.GetString(KafkaBootstrapServers)
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.MaxBufferedRecords(viper.GetInt(BufferSize)),
	)
	if err != nil {
		log.Fatal(err)
	}

	topic := viper.GetString(KafkaTopic)
	amount := viper.GetInt(Amount)

	start := time.Now()
	for i := 0; i < amount; i++ {
		record := &kgo.Record{
			Topic: topic,
			Value: []byte(fmt.Sprintf("{\"some_key\": %d}", i)),
		}

		cl.Produce(ctx, record, func(_ *kgo.Record, err error) {
			if err != nil {
				log.Println(err)
			}
		})
	}
	log.Println(
		fmt.Sprintf(
			"Finished %d messages in %s",
			amount,
			time.Now().Sub(start),
		),
	)
}
