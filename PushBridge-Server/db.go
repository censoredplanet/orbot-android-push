package main

import (
	"github.com/censoredplanet/orbot-android-push/PushBridge-server/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

// FCMDB stores the users, their FCM identifiers, pubkeys, and their subscriptions
type FCMDB struct {
	db *gorm.DB
}

func NewFCMDB(path string) (*FCMDB, error) {
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = database.AutoMigrate(&models.User{}, &models.Country{})
	if err != nil {
		return nil, err
	}

	fcmdb := &FCMDB{
		db: database,
	}
	return fcmdb, nil
}

func (DB *FCMDB) Close() {
	sqlDB, err := DB.db.DB()
	if err != nil {
		log.Fatalf("Error getting DB while closing: %v\n", err)
	}

	err = sqlDB.Close()
	if err != nil {
		log.Fatalf("Error closing DB: %v\n", err)
	}
}
