package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() {
	dsn := "host=localhost user=lath_xdomea password=temporary dbname=lath_xdomea port=5432 sslmode=disable TimeZone=Europe/Berlin"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	db = database
}
