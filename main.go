package main

import (
	//"database/sql"
	//"fmt
	//"net/http"

	"package/router"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dsn = "root:2580@tcp(localhost:8080)/data?charset=utf8mb4&parseTime=True&loc=Local"
var db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

func main() {

	router.Routes()

}

func debug() {

}
