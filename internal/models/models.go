package models

import (
	"time"
)

// Transaction represents a token transaction
type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TxHash      string    `gorm:"uniqueIndex;not null" json:"tx_hash"`
	BlockNumber uint64    `gorm:"index;not null" json:"block_number"`
	LogIndex    uint      `gorm:"not null" json:"log_index"`
	
	// Contract information
	ContractAddress string `gorm:"index;not null" json:"contract_address"`
	TokenName       string `gorm:"index" json:"token_name"`
	
	// Transaction details
	FromAddress string `gorm:"index;not null" json:"from_address"`
	ToAddress   string `gorm:"index;not null" json:"to_address"`
	Amount      string `gorm:"not null" json:"amount"` // Using string to handle big numbers
	
	// Price information (if available)
	PriceUSD    *float64 `json:"price_usd,omitempty"`
	ValueUSD    *float64 `json:"value_usd,omitempty"`
	
	// Timestamps
	BlockTimestamp time.Time `gorm:"index;not null" json:"block_timestamp"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Contract represents a monitored token contract
type Contract struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Address     string `gorm:"uniqueIndex;not null" json:"address"`
	StartBlock  uint64 `gorm:"not null" json:"start_block"`
	LastBlock   uint64 `gorm:"default:0" json:"last_block"`
	ABIFile     string `json:"abi_file,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relations
	Transactions []Transaction `gorm:"foreignKey:ContractAddress;references:Address" json:"transactions,omitempty"`
}

// BlockProgress tracks indexing progress
type BlockProgress struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Contract    string `gorm:"uniqueIndex;not null" json:"contract"`
	LastBlock   uint64 `gorm:"not null" json:"last_block"`
	UpdatedAt   time.Time `json:"updated_at"`
}