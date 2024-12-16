package services_test

import (
	"context"
	"testing"

	"github.com/eth-bridging/internal/models"
	"github.com/eth-bridging/internal/producer"
	"github.com/eth-bridging/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBridgeEventRepository struct {
	mock.Mock
}

func (m *MockBridgeEventRepository) Save(event *models.BridgeEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockBridgeEventRepository) GetAll(lastID uint, limit int, currency string) ([]models.BridgeEvent, error) {
	args := m.Called(lastID, limit, currency)
	return args.Get(0).([]models.BridgeEvent), args.Error(1)
}

type MockEthereumClient struct {
	mock.Mock
}

func (m *MockEthereumClient) StartBridgingEventPublisher(context.Context, producer.Producer) error {
	m.Called()
	return nil
}

type MockRedisProducer struct {
	mock.Mock
}

func (m *MockRedisProducer) PublishEvent(event models.BridgeEvent) error {
	m.Called(event)
	return nil
}

func (m *MockRedisProducer) Stop() {
	// nothing to do here, as there is no channel to stop that needs mocking
	m.Called()
}

func TestSaveEvent(t *testing.T) {
	mockRepo := new(MockBridgeEventRepository)
	mockRepo.On("Save", mock.Anything).Return(nil)
	service := services.NewBridgeEventService(mockRepo, nil)

	err := service.SaveEvent(&models.BridgeEvent{})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetAllEvents(t *testing.T) {
	mockRepo := new(MockBridgeEventRepository)
	mockRepo.On("GetAll", uint(0), 10, "ETH").Return([]models.BridgeEvent{}, nil)
	service := services.NewBridgeEventService(mockRepo, nil) // Pass nil for EthereumClient as it's not needed here

	events, err := service.GetAllEvents(0, 10, "ETH")

	assert.NoError(t, err)
	assert.NotNil(t, events)
	mockRepo.AssertExpectations(t)
}

func TestProcessIncomingBridgeEvents(t *testing.T) {
	mockClient := new(MockEthereumClient)
	mockProducer := new(MockRedisProducer)

	mockClient.On("StartBridgingEventPublisher", mock.Anything, mock.Anything).Return(nil)

	service := services.NewBridgeEventService(nil, mockClient)

	service.ProcessIncomingBridgeEvents(mockProducer)

	mockClient.AssertExpectations(t)
}
