# Services層

## 概要
Service層は、アプリケーションの**ビジネスロジック**を実装する層です。Controller層とRepository層の間に位置し、ビジネスルールやデータの加工を担当します。

## 責務
- ビジネスロジックの実装
- トランザクション管理
- 複数のRepositoryの協調
- DTOからModelへの変換
- ビジネスルールの適用
- 複雑な計算処理

## 書くべきもの
- ✅ ビジネスロジック
- ✅ データの加工・変換
- ✅ 複数のRepositoryを使った処理
- ✅ ビジネスルールのバリデーション
- ✅ トランザクション制御
- ❌ HTTPに関する処理
- ❌ 直接的なデータベース操作（ORMの使用）

## 実装例

### 良い例 ✅
```go
type ItemService struct {
    repository repositories.IItemRepository
    userRepo   repositories.IUserRepository  // 複数のRepositoryを使用可能
}

func (s *ItemService) Create(createItemInput dto.CreateItemInput) (*models.Item, error) {
    // ビジネスロジック：価格の計算（Service層の責務）
    finalPrice := s.calculatePrice(createItemInput.Price)
    
    // DTOからModelへの変換（Service層の責務）
    newItem := models.Item{
        Name:        createItemInput.Name,
        Price:       finalPrice,
        Description: createItemInput.Description,
        SoldOut:     false,  // 初期値の設定
    }
    
    // Repository層に永続化を委譲
    return s.repository.Create(newItem)
}

func (s *ItemService) calculatePrice(basePrice uint) uint {
    // ビジネスロジック：税込み価格の計算
    const taxRate = 1.10
    return uint(float64(basePrice) * taxRate)
}

// 複数のRepositoryを使った複雑な処理の例
func (s *ItemService) PurchaseItem(itemId uint, userId uint) error {
    // トランザクション開始（実際の実装では）
    // tx := s.db.Begin()
    
    // 商品の取得
    item, err := s.repository.FindById(itemId)
    if err != nil {
        return err
    }
    
    // ビジネスルール：売り切れチェック
    if item.SoldOut {
        return errors.New("この商品は売り切れです")
    }
    
    // ユーザーの取得
    user, err := s.userRepo.FindById(userId)
    if err != nil {
        return err
    }
    
    // ビジネスルール：残高チェック
    if user.Balance < item.Price {
        return errors.New("残高が不足しています")
    }
    
    // 商品を売り切れに更新
    item.SoldOut = true
    _, err = s.repository.Update(*item)
    if err != nil {
        // tx.Rollback()
        return err
    }
    
    // ユーザーの残高を更新
    user.Balance -= item.Price
    _, err = s.userRepo.Update(*user)
    if err != nil {
        // tx.Rollback()
        return err
    }
    
    // tx.Commit()
    return nil
}
```

### 悪い例 ❌
```go
func (s *ItemService) FindAll(ctx *gin.Context) (*[]models.Item, error) {
    // ❌ HTTPコンテキストを受け取ってはいけない
    userId := ctx.GetHeader("User-ID")
    
    // ❌ 直接SQLを書いてはいけない
    rows, err := s.db.Query("SELECT * FROM items WHERE user_id = ?", userId)
    
    // ❌ HTTPレスポンスを返してはいけない
    ctx.JSON(200, items)
}
```

## パターン例

### 1. 部分更新パターン
```go
func (s *ItemService) Update(itemId uint, updateInput dto.UpdateItemInput) (*models.Item, error) {
    // 既存のデータを取得
    targetItem, err := s.repository.FindById(itemId)
    if err != nil {
        return nil, err
    }
    
    // nilでない値のみ更新（部分更新）
    if updateInput.Name != nil {
        targetItem.Name = *updateInput.Name
    }
    if updateInput.Price != nil {
        // ビジネスロジック：価格更新時は税込み計算
        targetItem.Price = s.calculatePrice(*updateInput.Price)
    }
    if updateInput.Description != nil {
        targetItem.Description = *updateInput.Description
    }
    
    return s.repository.Update(*targetItem)
}
```

### 2. 集計処理パターン
```go
func (s *ItemService) GetStatistics() (*dto.ItemStatistics, error) {
    items, err := s.repository.FindAll()
    if err != nil {
        return nil, err
    }
    
    // ビジネスロジック：統計情報の計算
    var totalPrice uint
    var soldCount int
    
    for _, item := range *items {
        totalPrice += item.Price
        if item.SoldOut {
            soldCount++
        }
    }
    
    avgPrice := totalPrice / uint(len(*items))
    
    return &dto.ItemStatistics{
        TotalItems:   len(*items),
        SoldItems:    soldCount,
        AveragePrice: avgPrice,
    }, nil
}
```

## ディレクトリ構造
```
services/
├── README.md
├── item_services.go        # 商品関連のサービス
├── user_services.go        # ユーザー関連のサービス（例）
├── auth_services.go        # 認証関連のサービス（例）
└── notification_services.go # 通知関連のサービス（例）
```

## テストの書き方
```go
func TestItemService_Create(t *testing.T) {
    // モックリポジトリの準備
    mockRepo := new(MockItemRepository)
    service := NewItemService(mockRepo)
    
    // テストデータ
    input := dto.CreateItemInput{
        Name:        "テスト商品",
        Price:       1000,
        Description: "テスト用の商品です",
    }
    
    // 期待される結果（税込み価格）
    expectedItem := &models.Item{
        ID:          1,
        Name:        "テスト商品",
        Price:       1100,  // 税込み
        Description: "テスト用の商品です",
        SoldOut:     false,
    }
    
    // モックの設定
    mockRepo.On("Create", mock.MatchedBy(func(item models.Item) bool {
        return item.Price == 1100  // 税込み価格になっているか確認
    })).Return(expectedItem, nil)
    
    // 実行
    result, err := service.Create(input)
    
    // 検証
    assert.NoError(t, err)
    assert.Equal(t, expectedItem, result)
    mockRepo.AssertExpectations(t)
}
```

## 注意事項
- Service層はビジネスロジックの中心
- 複雑な処理はServiceで実装し、Controllerは薄く保つ
- トランザクションが必要な場合はService層で管理
- 外部サービスとの連携もService層で行う
- テストしやすい設計を心がける（依存性の注入）