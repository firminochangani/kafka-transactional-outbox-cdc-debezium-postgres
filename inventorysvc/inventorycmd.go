package inventorysvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/firminochangani/kafka-cdc-debezium-postgres/events"
)

func RunInventoryCmd() *cli.Command {
	return &cli.Command{
		Name:        "inventory-svc",
		Description: "Inventory service is responsible for consuming the event OrderPlaced",
		Action:      run,
	}
}

func run(ctx context.Context, command *cli.Command) error {
	consumer, err := NewConsumer(onOrderPlaced)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return consumer.Run(ctx)
	})

	g.Go(func() error {
		log.Println("inventorysvc is running...")
		<-ctx.Done()

		log.Println("stopping inventorysvc...")
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		return consumer.Stop(timeoutCtx)
	})

	return g.Wait()
}

func onOrderPlaced(ctx context.Context, body []byte) error {
	var event events.OrderPlaced
	err := json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	log.Printf("inventorysvc: A new order has been placed: %v\n", event)

	return nil
}

type Consumer struct {
	client  *kgo.Client
	handler ConsumerHandler
}

type ConsumerHandler func(ctx context.Context, body []byte) error

func NewConsumer(handler ConsumerHandler) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(os.Getenv("KAFKA_BROKER_URLS"), ",")...),
		kgo.ConsumeTopics(events.TopicOrders),
		kgo.ConsumerGroup("inventory"),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:  client,
		handler: handler,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("context closed")
		default:
		}

		fetches := c.client.PollFetches(ctx)
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			err := c.handler(context.Background(), record.Value)
			if err != nil {
				log.Printf("consumer: error consuming message '%s' due to: %v\n", string(record.Value), err)
				continue
			}
		}
		err := c.client.CommitUncommittedOffsets(ctx)
		if err != nil {
			return fmt.Errorf("consumer: error committing offsets: %v", err)
		}
	}
}

func (c *Consumer) Stop(ctx context.Context) error {
	err := c.client.LeaveGroupContext(ctx)
	if err != nil {
		return fmt.Errorf("consumer: error leaving consumer group: %w", err)
	}

	c.client.Close()

	return nil
}
