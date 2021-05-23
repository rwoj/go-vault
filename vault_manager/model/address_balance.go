package model

import "github.com/jinzhu/gorm"

type AddressBalance struct {
	gorm.Model
	Balance       string
	LockedBalance string
	AddressID     uint `gorm:"type:bigint REFERENCES addresses(id)"`
	Address       Address
	LastEventID   uint `gorm:"type:bigint REFERENCES address_events(id)"`
	LastEvent     AddressEvent
	Chain         string
	Coin          string
}

// NewAddressBalance creates a new address balance for an address
func NewAddressBalance(coin, chain string, addressID, lastEventID uint, balance, lockedBalance string) *AddressBalance {
	return &AddressBalance{
		Coin:          coin,
		Chain:         chain,
		AddressID:     addressID,
		LastEventID:   lastEventID,
		Balance:       balance,
		LockedBalance: lockedBalance,
	}
}
