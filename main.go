package main

// ginフレームワークをインポートします。
import (
	"gin-fleamarket/controllers"
	"gin-fleamarket/infra"
	"gin-fleamarket/repositories"
	"gin-fleamarket/services"

	"github.com/gin-gonic/gin"
)

func main() {
	infra.Initialize()
	// ginのデフォルトのルーターを作成します。
	// ルーターは、HTTPリクエストを処理するためのエンドポイントを定義します。
	router := gin.Default()

	// ルートエンドポイントを定義します。
	// ここでは、"/ping"というパスにGETリクエストが来たときに、
	// 無名関数を実行して、JSON形式でレスポンスを返します。
	router.GET("/ping", func(c *gin.Context) {
		// ここでcはjson形式のレスポンスを設定する
		c.JSON(200, gin.H{
			// gin.Hは、map[string]interface{}のエイリアスで、
			// JSONレスポンスを簡単に作成するために使用されます。
			"message": "pong",
		})
	})

	infra.Initialize()
	db := infra.SetupDB()
	// items := []models.Item{
	// 	{ID: 1, Name: "Item1", Price: 1000, Description: "Description1", SoldOut: false},
	// 	{ID: 2, Name: "Item2", Price: 2000, Description: "Description2", SoldOut: true},
	// 	{ID: 3, Name: "Item3", Price: 3000, Description: "Description3", SoldOut: false},
	// 	{ID: 4, Name: "Item4", Price: 4000, Description: "Description4", SoldOut: true},
	// }

	// itemRepository := repositories.NewItemMemoryRepository(items)
	itemRepository := repositories.NewItemRepository(db)

	itemService := services.NewItemService(itemRepository)
	itemController := controllers.NewItemController(itemService)
	router.GET("/items", itemController.FindAll)
	router.GET("/items/:id", itemController.FindById)
	router.POST("/items", itemController.Create)
	router.PUT("/items/:id", itemController.Update)
	router.DELETE("/items/:id", itemController.Delete)

	router.Run("localhost:8080") // 0.0.0.0:8080 でサーバーを立てます。
}
