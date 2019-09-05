package utils

import (
	"log"

	"github.com/jinzhu/gorm"
)

// GetConnection return DB Object
func GetConnection() *gorm.DB {
	db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalf("DB connection failed %v", err)
	}
	return db
}
