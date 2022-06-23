package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	kafkaTopic, kafkaBootstrapServers string
	bufferSize, amount, messageSize   int
	registerConsumer                  bool
)

func main() {
	var benchmarkCmd = &cobra.Command{
		Use:   "benchmark",
		Short: "Benchmark kafka publishing and consuming",
		Run:   benchmark,
	}
	benchmarkCmd.PersistentFlags().StringVarP(&kafkaTopic, "topic", "", "default-topic", "topic to use for benchmark")
	benchmarkCmd.PersistentFlags().StringVarP(&kafkaBootstrapServers, "bootstrap-servers", "", "127.0.0.1:9092,", "list of bootstrap servers; ex: 127.0.0.1:9092,")
	benchmarkCmd.PersistentFlags().IntVarP(&bufferSize, "buffer-size", "", 10, "buffer size to use for async publishing")
	benchmarkCmd.PersistentFlags().IntVarP(&amount, "amount", "", 10, "amount of messages to publish for benchmark")
	benchmarkCmd.PersistentFlags().IntVarP(&messageSize, "message-size", "", 1024, "size of message to use for each publish in bytes")
	benchmarkCmd.PersistentFlags().BoolVarP(&registerConsumer, "register-consumer", "", false, "whether to include a consumer or not")

	benchmarkCmd.Execute()
}

func benchmark(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	var cl *kgo.Client

	if registerConsumer {
		c, err := kgo.NewClient(
			kgo.SeedBrokers(strings.Split(kafkaBootstrapServers, ",")...),
			kgo.MaxBufferedRecords(bufferSize),
			kgo.ConsumerGroup("benchmark"),
			kgo.ConsumeTopics(kafkaTopic),
		)
		if err != nil {
			log.Fatal(err)
		}
		cl = c
	} else {
		c, err := kgo.NewClient(
			kgo.SeedBrokers(strings.Split(kafkaBootstrapServers, ",")...),
			kgo.MaxBufferedRecords(bufferSize),
		)
		if err != nil {
			log.Fatal(err)
		}
		cl = c
	}

	var wg sync.WaitGroup

	// Producer Bench
	wg.Add(1)
	go func() {
		produceMessages(ctx, cl, kafkaTopic, amount, &wg)
	}()

	// Consumer Bench
	if registerConsumer {
		wg.Add(1)
		go func() {
			consumeMessages(ctx, cl, amount, &wg)
		}()
	}

	wg.Wait()
	cl.Close()
}

func produceMessages(ctx context.Context, cl *kgo.Client, topic string, amount int, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	for i := 0; i < amount; i++ {
		wg.Add(1)
		v := make([]byte, messageSize)
		iBytes := byte(i)
		v[0] = iBytes
		record := &kgo.Record{
			Topic: topic,
			Value: v,
		}

		cl.Produce(ctx, record, func(_ *kgo.Record, err error) {
			defer wg.Done()
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

func consumeMessages(ctx context.Context, cl *kgo.Client, amount int, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	consumed := 0

	for consumed < amount {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			log.Println(errs)
			return
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			iter.Next()
			consumed++
		}
		consumed++
	}
	log.Println(fmt.Sprintf("Consumed %d messages in %s\n", consumed, time.Now().Sub(start)))
}
