package dao

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	// host := os.Getenv("HOST")
	// port := os.Getenv("PORT")
	// dbname := os.Getenv("DBNAME")
	// user := os.Getenv("USER")
	// password := os.Getenv("PASSWORD")
	// dsn := "host=" + host + " port=" + port + " dbname=" + dbname + " user=" + user + " password=" + password

	dsn := os.Getenv("DSN")

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func GetDB() *gorm.DB {
	return db
}
