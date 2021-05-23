package model

import "github.com/jinzhu/gorm"

// Address structure
type Address struct {
	gorm.Model `json:"-"`
	PublicKey  string  `gorm:"unique" json:"public_key"`
	AcoountID  uint    `gorm:"type:bigint REFERENCES accounts(id)" json:"account_id"`
	Account    Account `json:"-"`
	Chain      string  `json:"chain"`
}

// NewAddress creates a new address for an account
func NewAddress(publicKey, chain string, accountID uint) *Address {
	return &Address{
		PublicKey: publicKey,
		Chain:     chain,
		AccountID: accountID,
	}
}
