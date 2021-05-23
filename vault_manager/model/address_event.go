package model

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

// AddressEvent structure
type AddressEvent struct {
	gorm.Model
	TxID       string
	Type       string
	Symbol     string
	Amount     string
	Info       string
	CostAmount string
	CostSymbol string
	AddressID  sql.NullInt64 `gorm:"type:bigint REFERENCES addresses(id)" json:"address_id"`
	Address    Address       `json:"-"`
}

func NewAddressEvent(addressID uint, eventType, symbol, amount, txid, info, costAmount, costSymbol string) *AddressEvent {
	return &AddressEvent{
		AddressID:  addressID,
		Amount:     amount,
		Symbol:     symbol,
		Info:       info,
		TxID:       txid,
		Type:       eventType,
		CostAmount: costAmount,
		CostSymbol: costSymbol,
	}
}
