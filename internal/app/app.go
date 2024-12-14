package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eth-bridging/config"
	"github.com/eth-bridging/internal/consumer"
	"github.com/eth-bridging/internal/producer"
	"github.com/eth-bridging/internal/routers"
	"github.com/eth-bridging/pkg/di"
	ethereum "github.com/eth-bridging/pkg/go-eth"
)

func Run() {
	cfg := config.LoadConfig()
	wg := &sync.WaitGroup{}

	ethClient, err := ethereum.NewEthereumClient(cfg.EthereumRPCURL, cfg.SocketGateAddr, cfg.ContractABI, cfg.TopicHex)
	if err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}

	// Initialize the DI container
	container := di.InitializeContainer(cfg, ethClient, wg)

	// Initialize Router
	router := routers.SetupRouter(container)

	server := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: router,
	}

	// Start the server
	wg.Add(1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run server: %v", err)
		}
		wg.Done()
	}()

	go GracefulShutdown(server, &container.Consumer, &container.Producer)
	wg.Wait()
}

// GracefulShutdown: Handles stopping the server, producer, and consumer gracefully
// to ensure no abrupt server stopping during deployments
func GracefulShutdown(server *http.Server, consumer *consumer.RedisStreamConsumer, producer *producer.RedisProducer) {
	// Create a channel to listen for OS signals (e.g., CTRL+C)
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	// Wait for stop signal
	sig := <-stopSignal
	log.Printf("Received shutdown signal: %+v, initiating graceful shutdown...", sig)

	// Stop the consumer gracefully
	consumer.Stop()

	// Stop the producer gracefully
	producer.Stop()

	// Stop the API server from accepting new requests
	// Allow current requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
