package model

import (
	"fmt"

	"github.com/jinzhu/gorm"

	// import support for postgres for gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Repo structure
type Repo struct {
	Conn *gorm.DB
}

var repo *Repo

// NewRepo starts new connections
func NewRepo(host, port, user, pass, name string) *Repo {
	conn, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name))
	if err != nil {
		panic("failed to connect database")
	}
	repo = &Repo{Conn: conn}
	return repo
}

// GetRepo retuns the existing repo
func GetRepo() *Repo {
	if repo == nil {
		panic("no active database connection")
	}
	return repo
}

// SetRepo updates the local active repo
func SetRepo(rep *Repo) {
	repo = rep
}

// Create a record
func (repo *Repo) Create(model interface{}) error {
	db := repo.Conn.Create(model)
	return db.Error
}

// FindAll record of a table
func (repo *Repo) FindAll(model interface{}, where string) error {
	db := repo.Conn.Find(model)
	return db.Error
}

// FindByID a record of a table
func (repo *Repo) FindByID(model interface{}, id uint) error {
	db := repo.Conn.Where("id = ?", id).Find(model)
	return db.Error
}

// Update a record of a table
func (repo *Repo) Update(model interface{}) error {
	db := repo.Conn.Save(model)
	return db.Error
}

// Delete a record from a table
func (repo *Repo) Delete(model interface{}, id uint) error {
	db := repo.Conn.Where("id = ?", id).Delete(model)
	return db.Error
}
