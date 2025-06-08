# Controllers層

## 概要
Controller層は、HTTPリクエストの受け取りとレスポンスの返却を担当する**プレゼンテーション層**です。

## 責務
- HTTPリクエストの受け取りとレスポンスの返却
- リクエストパラメータの検証とパース
- DTOへのバインディング
- HTTPステータスコードの設定
- エラーハンドリング（ユーザー向けのエラーメッセージ）

## 書くべきもの
- ✅ HTTPに関する処理
- ✅ リクエスト/レスポンスの変換
- ✅ 基本的なバリデーション
- ❌ ビジネスロジック
- ❌ データベースアクセス
- ❌ 複雑な計算処理

## 実装例

### 良い例 ✅
```go
func (c *ItemController) FindById(ctx *gin.Context) {
    // URLパラメータの取得と検証（Controller層の責務）
    itemId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }
    
    // Service層に処理を委譲
    item, err := c.service.FindById(uint(itemId))
    if err != nil {
        // エラーの種類に応じたHTTPステータスの設定（Controller層の責務）
        if err.Error() == "Item not found" {
            ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected Error"})
        return
    }
    
    // レスポンスの返却（Controller層の責務）
    ctx.JSON(http.StatusOK, gin.H{"data": item})
}
```

### 悪い例 ❌
```go
func (c *ItemController) Create(ctx *gin.Context) {
    var input dto.CreateItemInput
    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // ❌ ビジネスロジックをControllerに書いてはいけない
    if input.Price > 100000 {
        input.Price = input.Price * 0.9 // 10%割引
    }
    
    // ❌ データベースアクセスを直接行ってはいけない
    db := getDB()
    newItem := models.Item{
        Name:  input.Name,
        Price: input.Price,
    }
    db.Create(&newItem)
    
    ctx.JSON(http.StatusCreated, gin.H{"data": newItem})
}
```

## ディレクトリ構造
```
controllers/
├── README.md
├── item_controller.go      # 商品関連のコントローラー
├── user_controller.go      # ユーザー関連のコントローラー（例）
└── auth_controller.go      # 認証関連のコントローラー（例）
```

## テストの書き方
```go
func TestItemController_FindById(t *testing.T) {
    // モックサービスの準備
    mockService := new(MockItemService)
    controller := NewItemController(mockService)
    
    // テスト用のGinコンテキストを作成
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    // パラメータの設定
    c.Params = []gin.Param{{Key: "id", Value: "1"}}
    
    // モックの期待値設定
    expectedItem := &models.Item{ID: 1, Name: "Test Item"}
    mockService.On("FindById", uint(1)).Return(expectedItem, nil)
    
    // 実行
    controller.FindById(c)
    
    // 検証
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}
```

## 注意事項
- Controllerは薄く保つ（Thin Controller）
- 複雑なロジックはService層に委譲する
- HTTPに関する処理のみに集中する
- エラーメッセージはユーザーフレンドリーに
