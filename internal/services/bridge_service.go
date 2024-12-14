package services

import (
	"context"
	"log"

	"github.com/eth-bridging/internal/models"
	"github.com/eth-bridging/internal/producer"
	"github.com/eth-bridging/internal/repositories"
	ethereum "github.com/eth-bridging/pkg/go-eth"
)

type BridgeEventService interface {
	// SaveEvent saves provided event to db
	SaveEvent(event *models.BridgeEvent) error
	// GetAllEvents fetches all events in paginated manner using lastID and limit
	GetAllEvents(lastID uint, limit int) ([]models.BridgeEvent, error)
	ProcessIncomingBridgeEvents(eventChannel chan ethereum.BridgingEvent, streamProducer *producer.RedisProducer)
}

type bridgeEventService struct {
	repo      repositories.BridgeEventRepository
	ethClient *ethereum.EthereumClient
}

// ProcessIncomingBridgeEvents listens for bridging events and saves them to the database
//
//	It is a blocking method, so ideally is should be called with `go` keyword
func (s *bridgeEventService) ProcessIncomingBridgeEvents(eventChannel chan ethereum.BridgingEvent, streamProducer *producer.RedisProducer) {
	// Start listening to events
	if err := s.ethClient.StartBridgingEventPublisher(context.Background(), eventChannel, streamProducer); err != nil {
		log.Fatalf("Error listening to events: %v", err)
	}
}

func NewBridgeEventService(repo repositories.BridgeEventRepository, ethClient *ethereum.EthereumClient) BridgeEventService {
	return &bridgeEventService{
		repo:      repo,
		ethClient: ethClient,
	}
}

func (s *bridgeEventService) SaveEvent(event *models.BridgeEvent) error {
	return s.repo.Save(event)
}

func (s *bridgeEventService) GetAllEvents(lastID uint, limit int) ([]models.BridgeEvent, error) {
	return s.repo.GetAll(lastID, limit)
}
