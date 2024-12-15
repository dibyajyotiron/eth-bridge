package producer

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/eth-bridging/internal/models"
	"github.com/go-redis/redis/v8"
)

// Producer defines the interface for publishing events.
type Producer interface {
	// PublishEvent publishes an event to the Redis stream
	PublishEvent(event models.BridgeEvent) error
	Stop()
}

// RedisClient defines the methods used by RedisProducer for testability
type RedisClient interface {
	XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd
}
type RedisProducer struct {
	client RedisClient
	stream string
	done   chan bool
	wg     *sync.WaitGroup
}

func NewRedisProducer(client RedisClient, stream string, wg *sync.WaitGroup) *RedisProducer {
	return &RedisProducer{
		client: client,
		stream: stream,
		done:   make(chan bool),
		wg:     wg,
	}
}

func (p *RedisProducer) Stop() {
	p.done <- true
	p.wg.Done()
}

// PublishEvent publishes an event to the Redis stream
func (p *RedisProducer) PublishEvent(event models.BridgeEvent) error {
	// Check if we received a stop signal before publishing
	select {
	case <-p.done:
		// Stop the producer if stop signal is received
		fmt.Printf("Stop signal received. Publisher will not publish any more events.")
		return nil
	default:
		// Continue publishing the events to redis
	}

	ctx := context.Background()

	// Convert the event to a flat key-value map as that is what will be sent using redis
	eventMap := map[string]interface{}{
		"transactionHash": event.TransactionHash,
		"token":           event.Token,
		"amount":          event.Amount,
		"fromChain":       event.FromChain,
		"toChain":         event.ToChain,
		"timestamp":       event.Timestamp,
	}

	// Add the event to the Redis stream
	_, err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.stream,
		Values: eventMap,
		ID:     "*",
	}).Result()

	if err != nil {
		log.Printf("Error publishing event to Redis stream: %v", err)
		return err
	}

	log.Printf("Event published to Redis stream: %s", p.stream)
	return nil
}
