package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserStore interface {
	GetUser(id string) (*User, error)
}

// Define a simple User model
type User struct {
	gorm.Model
}

// Database struct to manage the connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase initializes a new SQLite database
func NewDatabase(dbName string) *Database {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Auto migrate the User model
	db.AutoMigrate(&User{})

	return &Database{DB: db}
}

func (db *Database) GetUser(id string) (*User, error) {
	var user *User
	tx := db.DB.Find(user, id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return user, nil
}
