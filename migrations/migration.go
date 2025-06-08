// アプリケーションの起動とデータベースのマイグレーションを分離するため

package main

import (
	"gin-fleamarket/infra"
	"gin-fleamarket/models"
)

func main() {
	infra.Initialize()
	db := infra.SetupDB()

	if err := db.AutoMigrate(&models.Item{}); err != nil {
		panic("Failed to migrate database: ")
	}
}
