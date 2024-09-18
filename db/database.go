package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserStore interface {
	GetUser(id string) (*User, error)
	CreateUser(user *User) (int, error)
	UpdateUser(user *User) (int, error)
	DeleteUser(id int) error
	GetUsers() ([]User, error)
}

type User struct {
	gorm.Model
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Database struct {
	DB *gorm.DB
}

func NewDatabase(dbName string) *Database {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}

	return &Database{DB: db}
}

func (db *Database) GetUser(id string) (*User, error) {
	var user *User
	tx := db.DB.First(&user, id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return user, nil
}

func (db *Database) CreateUser(user *User) (int, error) {
	tx := db.DB.Save(&user)
	if tx.Error != nil {
		return -1, tx.Error
	}

	tx = tx.Find(&user)
	if tx.Error != nil {
		return -1, tx.Error
	}

	return int(user.ID), nil
}

func (db *Database) UpdateUser(user *User) (int, error) {
	tx := db.DB.Save(&user)
	if tx.Error != nil {
		return -1, tx.Error
	}

	tx = tx.Find(&user)
	if tx.Error != nil {
		return -1, tx.Error
	}

	return int(user.ID), nil
}

func (db *Database) DeleteUser(id int) error {
	var user User
	tx := db.DB.Model(User{}).Delete(&user, id)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (db *Database) GetUsers() ([]User, error) {
	var users []User
	tx := db.DB.Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return users, nil
}
