package db

import (
	"log"

	"package/structs"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dsn = "root:2580@tcp(localhost:8080)/data?charset=utf8mb4&parseTime=True&loc=Local"
var Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

func db() {
	if err != nil {
		panic("failed to connect database")
	}
	Db.AutoMigrate(&structs.Video{}, &structs.User{})

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}
}
