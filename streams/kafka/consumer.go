package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SriHemnath/credit_card_processing/utils/logger"

	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	group    *kafka.Reader
	messages chan kafka.Message
	wg       sync.WaitGroup
}

func NewConsumer(brokers []string, group string, topic string, cmtInterval time.Duration) (*Consumer, error) {

	rdr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                brokers,
		GroupID:                group,
		Topic:                  topic,
		QueueCapacity:          1000,
		MinBytes:               1e3, //1kb scientific notation
		MaxBytes:               1e7, //10mb
		CommitInterval:         cmtInterval,
		WatchPartitionChanges:  true,
		PartitionWatchInterval: 5 * time.Second,
		StartOffset:            kafka.FirstOffset,
		ReadBackoffMin:         50 * time.Millisecond,
		ReadBackoffMax:         500 * time.Millisecond,
		MaxAttempts:            5,
		Dialer: &kafka.Dialer{
			ClientID:  "kafka-go",
			Timeout:   30 * time.Second,
			DualStack: true,
		},
	})

	consumer := &Consumer{
		group:    rdr,
		messages: make(chan kafka.Message),
	}
	return consumer, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	var err error
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			var m kafka.Message
			m, err = c.group.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					err = nil
				}
				return
			}

			logger.Info(fmt.Sprintf("fetched message from topic=%v partition=%v offset=%v", m.Topic, m.Partition, m.Offset))

			select {
			case <-ctx.Done():
				logger.Info("consumer context cancelled, exiting")
				return
			case c.messages <- m:
			}
		}
	}()

	c.wg.Wait()
	logger.Info("consumer exiting")
	return nil
}

func (c *Consumer) Close() error {
	//close kafka reader
	err := c.group.Close()

	//wait for consumer loop to exit
	c.wg.Wait()

	close(c.messages)
	logger.Info("kafka consumer closed")
	return err
}

func (c *Consumer) Messages() <-chan kafka.Message {
	return c.messages
}

func (c *Consumer) CommitOffset(msg kafka.Message) {
	c.group.CommitMessages(context.Background(), msg)
	logger.Info(fmt.Sprintf("committed topic=%v partition=%v offset=%v", msg.Topic, msg.Partition, msg.Offset))
}
