package di

import (
	"log"
	"sync"

	"github.com/eth-bridging/config"
	"github.com/eth-bridging/internal/consumer"
	"github.com/eth-bridging/internal/producer"
	"github.com/eth-bridging/internal/repositories"
	"github.com/eth-bridging/internal/services"
	ethereum "github.com/eth-bridging/pkg/go-eth"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Container holds all dependencies for the app
// The purpose of Container is to ensure Dependency Injection(DI)
type Container struct {
	EventService services.BridgeEventService
	Consumer     consumer.RedisStreamConsumer
	Producer     producer.RedisProducer
}

// InitializeContainer initializes the components of the application, including
// the database (PostgreSQL), Redis client, repositories, services, and stream consumers/producers.
//
// It also starts the necessary goroutines for processing incoming events and consuming
// Redis stream messages. This function returns a Container object that contains
// references to the EventService and StreamConsumer.
//
//	Note:
//	  Ideally in production, consumer should run as a separate microservice, for simplicity, we are clubbing both in a single service
func InitializeContainer(cfg *config.Config, ethClient *ethereum.EthereumClient, wg *sync.WaitGroup) *Container {
	// Initialize PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.PostgresURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// Initialize Repository
	eventRepo := repositories.NewBridgeEventRepository(db, cfg)

	// Initialize Service
	eventService := services.NewBridgeEventService(eventRepo, ethClient)

	// Initialize Redis Stream Consumer
	input := &consumer.NewConsumerInput{
		Client:     redisClient,
		StreamName: cfg.RedisStreamName,
		GroupName:  "bridge_group",
		ConsumerID: "consumer_1",
		Service:    eventService,
		Wg:         wg,
		Cfg:        cfg,
	}
	streamConsumer := consumer.NewRedisStreamConsumer(input)

	// Initialize Redis Stream Consumer
	wg.Add(1)
	go streamConsumer.Consume()

	// Initialize Redis Stream Producer
	streamProducer := producer.NewRedisProducer(redisClient, cfg.RedisStreamName, wg)

	// Start processing the incoming bridging events
	// Ideally, in production, processing should be a separate microservice
	wg.Add(1)
	go eventService.ProcessIncomingBridgeEvents(streamProducer)

	return &Container{
		EventService: eventService,
		Consumer:     *streamConsumer,
		Producer:     *streamProducer,
	}
}
