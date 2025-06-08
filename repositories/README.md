# Repositories層

## 概要
Repository層は、データの永続化とアクセスを担当する**データアクセス層**です。データベースやメモリストレージとの直接的なやり取りを行います。

## 責務
- データの永続化（Create, Update, Delete）
- データの取得（Read）
- データベース固有の処理
- エラーハンドリング（データアクセスエラー）
- クエリの最適化

## 書くべきもの
- ✅ データアクセスロジック
- ✅ SQL/ORMの操作
- ✅ データベース固有のエラー処理
- ✅ クエリの最適化
- ✅ トランザクション処理（必要に応じて）
- ❌ ビジネスロジック
- ❌ データの加工・計算
- ❌ HTTPに関する処理

## 実装例

### 良い例 ✅

#### インターフェース定義
```go
type IItemRepository interface {
    FindAll() (*[]models.Item, error)
    FindById(itemId uint) (*models.Item, error)
    Create(newItem models.Item) (*models.Item, error)
    Update(updateItem models.Item) (*models.Item, error)
    Delete(itemId uint) error
    
    // 追加のメソッド例
    FindByUserId(userId uint) (*[]models.Item, error)
    FindBySoldOut(soldOut bool) (*[]models.Item, error)
    CountByCategory(category string) (int64, error)
}
```

#### GORM実装
```go
type ItemRepository struct {
    db *gorm.DB
}

func NewItemRepository(db *gorm.DB) IItemRepository {
    return &ItemRepository{db: db}
}

// 基本的なCRUD操作
func (r *ItemRepository) FindAll() (*[]models.Item, error) {
    var items []models.Item
    // プリロードを使った関連データの取得
    result := r.db.Preload("User").Find(&items)
    if result.Error != nil {
        return nil, result.Error
    }
    return &items, nil
}

func (r *ItemRepository) FindById(itemId uint) (*models.Item, error) {
    var item models.Item
    result := r.db.First(&item, itemId)
    if result.Error != nil {
        // エラーの種類に応じた処理（Repository層の責務）
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, errors.New("Item not found")
        }
        return nil, result.Error
    }
    return &item, nil
}

func (r *ItemRepository) Create(newItem models.Item) (*models.Item, error) {
    result := r.db.Create(&newItem)
    if result.Error != nil {
        return nil, result.Error
    }
    return &newItem, nil
}

func (r *ItemRepository) Update(updateItem models.Item) (*models.Item, error) {
    // 更新対象のカラムを明示的に指定
    result := r.db.Model(&updateItem).Updates(map[string]interface{}{
        "name":        updateItem.Name,
        "price":       updateItem.Price,
        "description": updateItem.Description,
        "sold_out":    updateItem.SoldOut,
    })
    if result.Error != nil {
        return nil, result.Error
    }
    if result.RowsAffected == 0 {
        return nil, errors.New("Item not found")
    }
    return &updateItem, nil
}

func (r *ItemRepository) Delete(itemId uint) error {
    result := r.db.Delete(&models.Item{}, itemId)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("Item not found")
    }
    return nil
}

// 複雑なクエリの例
func (r *ItemRepository) FindByPriceRange(minPrice, maxPrice uint) (*[]models.Item, error) {
    var items []models.Item
    result := r.db.Where("price >= ? AND price <= ?", minPrice, maxPrice).
        Order("price ASC").
        Find(&items)
    if result.Error != nil {
        return nil, result.Error
    }
    return &items, nil
}

// ページネーションの例
func (r *ItemRepository) FindWithPagination(page, pageSize int) (*[]models.Item, int64, error) {
    var items []models.Item
    var total int64
    
    // 総件数を取得
    r.db.Model(&models.Item{}).Count(&total)
    
    // ページネーション
    offset := (page - 1) * pageSize
    result := r.db.Offset(offset).Limit(pageSize).Find(&items)
    if result.Error != nil {
        return nil, 0, result.Error
    }
    
    return &items, total, nil
}

// トランザクションを使った例
func (r *ItemRepository) CreateWithTransaction(tx *gorm.DB, newItem models.Item) (*models.Item, error) {
    result := tx.Create(&newItem)
    if result.Error != nil {
        return nil, result.Error
    }
    return &newItem, nil
}
```

### 悪い例 ❌
```go
func (r *ItemRepository) FindAll() (*[]models.Item, error) {
    var items []models.Item
    r.db.Find(&items)
    
    // ❌ ビジネスロジックをRepositoryに書いてはいけない
    for i, item := range items {
        if item.Price > 10000 {
            items[i].Price = item.Price * 0.9  // 割引計算
        }
    }
    
    // ❌ HTTPステータスを返してはいけない
    if len(items) == 0 {
        return nil, fmt.Errorf("404: No items found")
    }
    
    return &items, nil
}
```

## メモリ実装の例
```go
type ItemMemoryRepository struct {
    items []models.Item
    mu    sync.RWMutex  // 並行アクセス対策
}

func NewItemMemoryRepository(items []models.Item) IItemRepository {
    return &ItemMemoryRepository{items: items}
}

func (r *ItemMemoryRepository) FindAll() (*[]models.Item, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    // スライスのコピーを返す（参照の問題を避ける）
    itemsCopy := make([]models.Item, len(r.items))
    copy(itemsCopy, r.items)
    return &itemsCopy, nil
}

func (r *ItemMemoryRepository) Create(newItem models.Item) (*models.Item, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // IDの自動採番
    newItem.ID = uint(len(r.items) + 1)
    r.items = append(r.items, newItem)
    return &newItem, nil
}
```

## パフォーマンス最適化の例

### 1. インデックスの活用
```go
// マイグレーションでインデックスを追加
func (r *ItemRepository) Migrate() error {
    return r.db.AutoMigrate(&models.Item{})
}

// models/item.go でインデックスを定義
type Item struct {
    gorm.Model
    Name        string `gorm:"not null;index"`           // 単一インデックス
    Price       uint   `gorm:"not null;index:idx_price"` // 名前付きインデックス
    UserId      uint   `gorm:"index:idx_user_soldout"`   // 複合インデックス
    SoldOut     bool   `gorm:"index:idx_user_soldout"`
}
```

### 2. N+1問題の回避
```go
// 悪い例：N+1問題が発生
func (r *ItemRepository) FindAllBad() (*[]models.Item, error) {
    var items []models.Item
    r.db.Find(&items)
    // ユーザー情報を個別に取得（N回のクエリ）
    for i := range items {
        r.db.First(&items[i].User, items[i].UserId)
    }
    return &items, nil
}

// 良い例：Preloadを使用
func (r *ItemRepository) FindAllGood() (*[]models.Item, error) {
    var items []models.Item
    // 1回のクエリで関連データも取得
    result := r.db.Preload("User").Find(&items)
    if result.Error != nil {
        return nil, result.Error
    }
    return &items, nil
}
```

## ディレクトリ構造
```
repositories/
├── README.md
├── item_repository.go      # 商品リポジトリ
├── user_repository.go      # ユーザーリポジトリ（例）
├── interfaces.go           # インターフェース定義をまとめる場合
└── mock/
    └── item_repository_mock.go  # テスト用モック
```

## テストの書き方
```go
func TestItemRepository_Create(t *testing.T) {
    // テスト用のDBセットアップ
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    repo := NewItemRepository(db)
    
    // テストデータ
    newItem := models.Item{
        Name:        "テスト商品",
        Price:       1000,
        Description: "テスト",
        SoldOut:     false,
    }
    
    // 実行
    created, err := repo.Create(newItem)
    
    // 検証
    assert.NoError(t, err)
    assert.NotZero(t, created.ID)
    assert.Equal(t, newItem.Name, created.Name)
    
    // DBに実際に保存されたか確認
    var saved models.Item
    db.First(&saved, created.ID)
    assert.Equal(t, created.Name, saved.Name)
}
```

## 注意事項
- Repository層は純粋にデータアクセスのみを扱う
- ビジネスロジックは一切含めない
- エラーハンドリングは適切に行う（特にNotFoundエラー）
- パフォーマンスを意識したクエリを書く
- インターフェースを定義してテスタビリティを確保
- 並行アクセスに注意（特にメモリ実装）
