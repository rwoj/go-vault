package model

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/nu7hatch/gouuid"
)

// Account structure
type Account struct {
	gorm.Model
	AccountID string `gorm:"unique; not null" json:"account_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

// NewAccount creates a new account for a merchant
func NewAccount(name string, email string) *Account {
	apiKey, _ := uuid.NewV4()
	return &Account{
		Name:      name,
		AccountID: apiKey.String(),
		Email:     email,
	}
}
