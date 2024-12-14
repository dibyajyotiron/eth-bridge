package ethereum

import (
	"strings"
	"testing"

	"github.com/eth-bridging/internal/models"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisProducer is a mock for the RedisProducer
type MockRedisProducer struct {
	mock.Mock
}

func (m *MockRedisProducer) PublishEvent(event models.BridgeEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockRedisProducer) Stop() {
	m.Called()
}

// Helper function to create a sample Ethereum log
func createSampleLog() types.Log {
	return types.Log{
		TxHash: common.HexToHash("0x1234567890abcdef"),
		Data:   []byte{0x00, 0x01, 0x02}, // Random testing data, no meaning whatsoever
	}
}

func TestDecodeSocketBridgeEvent(t *testing.T) {
	contractABI := `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"SocketBridge","type":"event"}]`
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	assert.NoError(t, err)

	sampleLog := createSampleLog()

	event, err := decodeSocketBridgeEvent(parsedABI, sampleLog)

	assert.Error(t, err)
	assert.Nil(t, event)
}
