package repositories

import (
	"fmt"

	"github.com/eth-bridging/config"
	"github.com/eth-bridging/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BridgeEventRepository interface {
	Save(event *models.BridgeEvent) error
	GetAll(lastID uint, limit int, currency string) ([]models.BridgeEvent, error)
}

type bridgeEventRepositoryImpl struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewBridgeEventRepository(db *gorm.DB, cfg *config.Config) BridgeEventRepository {
	return &bridgeEventRepositoryImpl{db: db, cfg: cfg}
}

func (r *bridgeEventRepositoryImpl) Save(event *models.BridgeEvent) error {
	return r.db.Omit("TxnCurrency").Create(event).Error
}

func (r *bridgeEventRepositoryImpl) GetAll(lastID uint, limit int, currency string) ([]models.BridgeEvent, error) {
	var events []models.BridgeEvent

	amountSQL := r.generateAmountSQL(currency)
	currencySQL := r.generateCurrencySQL(currency)

	// Build the base query
	query := r.db.Select("id",
		"token",
		amountSQL.SQL,
		currencySQL.SQL,
		"from_chain",
		"to_chain",
		"transaction_hash",
		"timestamp",
	).Order("timestamp desc").Limit(limit)

	// If a cursor is provided, use it for keyset pagination
	if lastID != 0 {
		query = query.Where("id < ?", lastID)
	}

	// Execute the query
	err := query.Debug().Find(&events).Error
	return events, err
}

// generateCurrencySQL generates the CASE statements for amount and currency fields
func (r *bridgeEventRepositoryImpl) generateAmountSQL(currency string) clause.Expr {
	currencyDetails := r.cfg.GetCurrencyDetails(currency)

	// This is not prone to sql injection
	// as Factor is not sent by user, instead
	// comes from our own config, so if user says
	// something wrong, it will be set to default
	// values
	amountExpr := gorm.Expr(fmt.Sprintf("(amount::numeric / POWER(10, %d))::text as amount", currencyDetails.Factor))

	return amountExpr
}

// generateCurrencySQL generates the CASE statements for amount and currency fields
func (r *bridgeEventRepositoryImpl) generateCurrencySQL(currency string) clause.Expr {
	currencyDetails := r.cfg.GetCurrencyDetails(currency)

	// This is not prone to sql injection
	// as Currency is not sent by user, instead
	// comes from our own config, so if user says
	// something wrong, it will be set to default
	// values
	currencyExpr := gorm.Expr(fmt.Sprintf("'%s'::text AS txn_currency", currencyDetails.Currency))

	return currencyExpr
}
