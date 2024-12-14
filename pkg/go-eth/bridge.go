package ethereum

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/eth-bridging/internal/models"
	"github.com/eth-bridging/internal/producer"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Event data structure
type BridgingEvent struct {
	TxHash     string
	Amount     *big.Int       `json:"amount"`
	Token      common.Address `json:"token"`
	ToChainId  *big.Int       `json:"toChainId"`
	BridgeName [32]byte       `json:"bridgeName"`
	Sender     common.Address `json:"sender"`
	Receiver   common.Address `json:"receiver"`
	Metadata   [32]byte       `json:"metadata"`
}

// Ethereum client wrapper
type EthereumClient struct {
	client  *ethclient.Client
	address common.Address
	topic   common.Hash
	abi     abi.ABI
}

// NewEthereumClient initializes the Ethereum client with parsed ABI interface
func NewEthereumClient(url, contractAddress, contractABI, topicHex string) (*EthereumClient, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}

	address := common.HexToAddress(contractAddress)
	socketTopicHash := common.HexToHash(topicHex)

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	return &EthereumClient{
		client:  client,
		address: address,
		topic:   socketTopicHash,
		abi:     parsedABI,
	}, nil
}

// StartBridgingEventPublisher listens for bridging events from an Ethereum contract
// and publishes them to the Redis stream via the provided streamProducer.
// It subscribes to Ethereum logs, decodes the events, and processes each
// bridging event by publishing it to Redis.
// This way, even if something fails during consuming, the messages can be retried as it's queue based.
func (ec *EthereumClient) StartBridgingEventPublisher(ctx context.Context, ch chan BridgingEvent, streamProducer producer.Producer) error {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{ec.address},
		Topics:    [][]common.Hash{{ec.topic}},
	}

	// Channel to receive results of the streaming filter query
	logs := make(chan types.Log)

	// Subscribe to the logs with the filter query
	sub, err := ec.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to contract events: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			// Error while subscribing, log and return the error
			return fmt.Errorf("error while subscribing to logs: %w", err)
		case vLog := <-logs:
			// Decode the log into a bridging event
			bridgingEvent, err := decodeSocketBridgeEvent(ec.abi, vLog)
			if err != nil || bridgingEvent == nil {
				// If decoding fails, skip the log and continue as other logs might not be failing
				log.Printf("Error decoding event: %+v, log: %+v\n", err, vLog)
				continue
			}

			// Create the BridgeEvent struct
			bridgeEvent := &models.BridgeEvent{
				TransactionHash: bridgingEvent.TxHash,
				FromChain:       bridgingEvent.Sender.Hex(),
				ToChain:         bridgingEvent.Receiver.Hex(),
				Amount:          fmt.Sprint(bridgingEvent.Amount),
				Token:           fmt.Sprint(bridgingEvent.Token),
				Timestamp:       time.Now(),
			}

			// Publish the event to the Redis stream
			if err := streamProducer.PublishEvent(*bridgeEvent); err != nil {
				log.Printf("Error publishing event to Redis, event: %+v error: %+v\n", *bridgeEvent, err)
				continue
			}
		}
	}
}

// decodeSocketBridgeEvent decodes the log data from streaming filter query into a
// BridgingEvent struct. It uses the provided ABI to unpack the log data and
// fills the fields of the BridgingEvent, including the transaction hash.
func decodeSocketBridgeEvent(parsedABI abi.ABI, vLog types.Log) (*BridgingEvent, error) {
	eventData := BridgingEvent{
		TxHash: vLog.TxHash.Hex(),
	}

	err := parsedABI.UnpackIntoInterface(&eventData, "SocketBridge", vLog.Data)
	if err != nil {
		log.Printf("Failed to unpack log data: %v", err)
		return nil, err
	}

	return &eventData, nil
}
