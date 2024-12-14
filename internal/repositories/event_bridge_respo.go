package repositories

import (
	"github.com/eth-bridging/internal/models"

	"gorm.io/gorm"
)

type BridgeEventRepository interface {
	Save(event *models.BridgeEvent) error
	GetAll(lastID uint, limit int) ([]models.BridgeEvent, error)
}

type bridgeEventRepositoryImpl struct {
	db *gorm.DB
}

func NewBridgeEventRepository(db *gorm.DB) BridgeEventRepository {
	return &bridgeEventRepositoryImpl{db: db}
}

func (r *bridgeEventRepositoryImpl) Save(event *models.BridgeEvent) error {
	return r.db.Create(event).Error
}

func (r *bridgeEventRepositoryImpl) GetAll(lastID uint, limit int) ([]models.BridgeEvent, error) {
	var events []models.BridgeEvent

	// Build the base query
	query := r.db.Order("timestamp desc").Limit(limit)

	// If a cursor is provided, use it for keyset pagination
	if lastID != 0 {
		query = query.Where("id < ?", lastID)
	}

	// Execute the query
	err := query.Find(&events).Error
	return events, err
}
