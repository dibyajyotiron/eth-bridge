package models

import "time"

type BridgeEvent struct {
	ID              int    `gorm:"primaryKey"`
	Token           string `gorm:"size:100"`
	Amount          string
	FromChain       string `gorm:"size:50"`
	ToChain         string `gorm:"size:50"`
	Timestamp       time.Time
	TransactionHash string
}
