package db

import (
	"package/platform/structs"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dsn = "root:2580@tcp(localhost:8080)/data?charset=utf8mb4&parseTime=True&loc=Local"
var Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

func ConnectToDatabase() {
	if err != nil {
		panic("failed to connect database")
	}
	Db.AutoMigrate(&structs.Video{}, &structs.User{}, &structs.Creator{})

}
