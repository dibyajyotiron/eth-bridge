package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/eth-bridging/config"
	"github.com/eth-bridging/internal/models"
	"github.com/eth-bridging/internal/services"
	rediscli "github.com/eth-bridging/pkg/redisclient"
	"github.com/go-redis/redis/v8"
)

type RedisStreamConsumer struct {
	ctx        context.Context
	client     rediscli.RedisClient
	streamName string
	groupName  string
	consumerID string
	service    services.BridgeEventService
	done       chan bool
	wg         *sync.WaitGroup
	cfg        *config.Config
}

type NewConsumerInput struct {
	Client     rediscli.RedisClient
	StreamName string
	GroupName  string
	ConsumerID string
	Service    services.BridgeEventService
	Wg         *sync.WaitGroup
	Cfg        *config.Config
}

// NewRedisStreamConsumer creates a new Redis stream consumer that listens
// to a specific Redis stream and processes messages using the provided service.
//
// The function performs the following actions:
// - Creates a new consumer group for the given stream if it doesn't already exist.
// - Returns a new RedisStreamConsumer instance with the provided configuration.
//
// Note:
// In a production environment with multiple pods, it is recommended to use the same `consumerID`
// across all pods. This ensures that events are processed in parallel by the multiple instances
// of the consumer group, rather than each pod processing the same messages.
func NewRedisStreamConsumer(input *NewConsumerInput) *RedisStreamConsumer {
	ctx := context.Background()

	// Create the consumer group if it doesn't exist. The "0" indicates starting from the earliest message.
	err := input.Client.XGroupCreateMkStream(ctx, input.StreamName, input.GroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("Failed to create consumer group: %v", err)
	}

	return &RedisStreamConsumer{
		client:     input.Client,
		streamName: input.StreamName,
		groupName:  input.GroupName,
		consumerID: input.ConsumerID,
		service:    input.Service,
		ctx:        ctx,
		done:       make(chan bool),
		wg:         input.Wg,
		cfg:        input.Cfg,
	}
}

func (r *RedisStreamConsumer) Stop() {
	r.done <- true
	r.wg.Done()
}

// Consume listens for messages from the Redis stream, processes each message,
// decodes it into a BridgeEvent, and saves the event to the database.
// It acknowledges each message after processing.
//
// The method continuously reads from the stream in a loop, handling any errors
// in reading or decoding. If an error occurs during event saving, it logs the error.
//
// It also ensures the stream message is acknowledged once processed.
//
//	Note: Since it is a blocking process, please ensure to call it with `go` keyword
func (r *RedisStreamConsumer) Consume() {
	for {
		select {
		// Stop the consumer gracefully
		case <-r.done:
			log.Println("Stopping consumer gracefully...")
			return
		default:
			entries, err := r.client.XReadGroup(r.ctx, &redis.XReadGroupArgs{
				Group:    r.groupName,
				Consumer: r.consumerID,
				Streams:  []string{r.streamName, ">"},
				Count:    10,
				// Block:    0,
			}).Result()

			if err != nil {
				log.Printf("Error reading from Redis stream: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			r.processStreamEntries(entries)
		}
	}
}

// processStreamEntries Processes the stream messages coming from redis
// And store in DB if processing is success
// Acknowledge the message
func (r *RedisStreamConsumer) processStreamEntries(entries []redis.XStream) {
	for _, stream := range entries {
		for _, message := range stream.Messages {

			var eventMsg models.BridgeEvent
			if err := decodeMessage(message, &eventMsg); err != nil {
				log.Printf("Error decoding message: %v", err)
				r.moveToDLQ(message)
				continue
			}

			event := models.BridgeEvent{
				Token:           eventMsg.Token,
				Amount:          eventMsg.Amount,
				FromChain:       eventMsg.FromChain,
				ToChain:         eventMsg.ToChain,
				Timestamp:       eventMsg.Timestamp,
				TransactionHash: eventMsg.TransactionHash,
			}

			if err := r.service.SaveEvent(&event); err != nil {
				log.Printf("Error saving event: %v", err)
				r.moveToDLQ(message)
			} else {
				log.Printf("Processed event: %+v", event)
			}

			r.client.XAck(r.ctx, r.streamName, r.groupName, message.ID)
		}
	}
}

// moveToDLQ moves the message to a Dead Letter Queue (DLQ)
// !Caution: Ideally should move messages to DLQ that failed processing after multiple retries
func (r *RedisStreamConsumer) moveToDLQ(message redis.XMessage) {
	// Add the message to a DLQ stream (e.g., "bridging_events_dlq")
	err := r.client.XAdd(r.ctx, &redis.XAddArgs{
		Stream: r.cfg.RedisStreamDlq,
		Values: message.Values,
	}).Err()
	if err != nil {
		log.Printf("Error moving message %s to DLQ: %v", message.ID, err)
	}
}

// decodeMessage simply converts redis XMessage into BridgeEvent
func decodeMessage(msg redis.XMessage, eventMsg *models.BridgeEvent) error {
	data, err := json.Marshal(msg.Values)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, eventMsg)
}
