package producer_test

import (
	"bytes"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/eth-bridging/internal/models"
	"github.com/eth-bridging/internal/producer"
	redisCli "github.com/eth-bridging/pkg/redisclient"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublishEvent(t *testing.T) {

	mockClient := new(redisCli.MockRedisClient)

	mockProducer := producer.NewRedisProducer(mockClient, "test-stream", nil)

	event := models.BridgeEvent{
		TransactionHash: "0x1234",
		Token:           "ETH",
		Amount:          "1000",
		FromChain:       "Ethereum",
		ToChain:         "Binance",
		Timestamp:       time.Now(),
	}

	mockClient.On("XAdd", mock.Anything, mock.Anything).Return(&redis.StringCmd{})

	err := mockProducer.PublishEvent(event)

	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestPublishEvent_ErrorOnXAdd(t *testing.T) {

	mockClient := new(redisCli.MockRedisClient)

	mockProducer := producer.NewRedisProducer(mockClient, "test-stream", nil)

	event := models.BridgeEvent{
		TransactionHash: "0x1234",
		Token:           "ETH",
		Amount:          "1000",
		FromChain:       "Ethereum",
		ToChain:         "Binance",
		Timestamp:       time.Now(),
	}

	cmd := &redis.StringCmd{}
	redisErrorTestMessage := "redis error"

	cmd.SetErr(errors.New(redisErrorTestMessage))

	mockClient.On("XAdd", mock.Anything, mock.Anything).Return(cmd)

	err := mockProducer.PublishEvent(event)

	assert.Error(t, err)
	assert.Equal(t, redisErrorTestMessage, err.Error())

	mockClient.AssertExpectations(t)
}

func TestPublishEvent_StopSignal(t *testing.T) {
	mockClient := new(redisCli.MockRedisClient)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	mockProducer := producer.NewRedisProducer(mockClient, "test-stream", wg)

	event := models.BridgeEvent{
		TransactionHash: "0x1234",
		Token:           "ETH",
		Amount:          "1000",
		FromChain:       "Ethereum",
		ToChain:         "Binance",
		Timestamp:       time.Now(),
	}
	var err error

	go func() {
		for {
			mockClient.On("XAdd", mock.Anything, mock.Anything).Return(&redis.StringCmd{})

			err = mockProducer.PublishEvent(event)
			if err != nil {
				t.Errorf("Error publishing event: %v", err)
			}
		}
	}()

	go mockProducer.Stop()

	wg.Wait()

	assert.NoError(t, err)
	// Ideally, we're expecting server should stop, so it should reach here and there should not come any error
	// If test is run on terminal, a log should be visible as well ->
	// Stop signal received. Publisher will not publish any more events.
}

func TestPublishEvent_Success(t *testing.T) {

	mockClient := new(redisCli.MockRedisClient)

	mockProducer := producer.NewRedisProducer(mockClient, "test-stream", nil)

	event := models.BridgeEvent{
		TransactionHash: "0x1234",
		Token:           "ETH",
		Amount:          "1000",
		FromChain:       "Ethereum",
		ToChain:         "Binance",
		Timestamp:       time.Now(),
	}

	mockClient.On("XAdd", mock.Anything, mock.Anything).Return(&redis.StringCmd{})

	err := mockProducer.PublishEvent(event)

	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestPublishEvent_WithEmptyTransactionHash(t *testing.T) {

	mockClient := new(redisCli.MockRedisClient)

	mockProducer := producer.NewRedisProducer(mockClient, "test-stream", nil)

	event := models.BridgeEvent{
		TransactionHash: "",
		Token:           "ETH",
		Amount:          "1000",
		FromChain:       "Ethereum",
		ToChain:         "Binance",
		Timestamp:       time.Now(),
	}

	mockClient.On("XAdd", mock.Anything, mock.Anything).Return(&redis.StringCmd{})

	err := mockProducer.PublishEvent(event)

	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestPublishEvent_WithInvalidData(t *testing.T) {

	mockClient := new(redisCli.MockRedisClient)

	mockProducer := producer.NewRedisProducer(mockClient, "test-stream", nil)

	event := models.BridgeEvent{
		TransactionHash: "0x1234",
		Token:           "ETH",
		Amount:          "-1000",
		FromChain:       "Ethereum",
		ToChain:         "Binance",
		Timestamp:       time.Now(),
	}

	mockClient.On("XAdd", mock.Anything, mock.Anything).Return(&redis.StringCmd{})

	err := mockProducer.PublishEvent(event)

	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}
