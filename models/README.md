# Models

## 概要
Model層は、データベースのテーブル構造を表現するエンティティを定義します。GORMのタグを使用してデータベースとのマッピングを行います。

## 責務
- データベーステーブルの構造定義
- リレーションシップの定義
- データベース制約の定義
- フックメソッドの実装（必要に応じて）

## 書くべきもの
- ✅ テーブル構造の定義
- ✅ GORMタグによる制約
- ✅ リレーションシップ
- ✅ インデックスの定義
- ✅ フックメソッド（BeforeCreate等）
- ❌ ビジネスロジック
- ❌ バリデーション処理（基本的な制約以外）
- ❌ データアクセスロジック

## 実装例

### 基本的なModel

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// 商品モデル
type Item struct {
    gorm.Model  // ID, CreatedAt, UpdatedAt, DeletedAtを含む
    
    // 基本フィールド
    Name        string `gorm:"not null;size:100;index"`
    Price       uint   `gorm:"not null;check:price >= 0"`
    Description string `gorm:"type:text"`
    SoldOut     bool   `gorm:"not null;default:false;index"`
    
    // 外部キー
    UserId     uint `gorm:"not null;index:idx_user_soldout"`
    CategoryId uint `gorm:"not null;index"`
    
    // リレーション
    User     User     `gorm:"foreignKey:UserId"`
    Category Category `gorm:"foreignKey:CategoryId"`
    Images   []Image  `gorm:"foreignKey:ItemId"`
}

// ユーザーモデル
type User struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // ユニーク制約
    Email    string `gorm:"unique;not null;size:255"`
    Username string `gorm:"unique;not null;size:50"`
    
    // その他のフィールド
    Password    string `gorm:"not null"`
    DisplayName string `gorm:"size:100"`
    Balance     uint   `gorm:"default:0"`
    IsActive    bool   `gorm:"default:true;index"`
    
    // リレーション
    Items    []Item    `gorm:"foreignKey:UserId"`
    Orders   []Order   `gorm:"foreignKey:UserId"`
    Profile  Profile   `gorm:"foreignKey:UserId"`
}

// カテゴリーモデル
type Category struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    
    Name     string `gorm:"unique;not null;size:50"`
    ParentId *uint  `gorm:"index"` // 自己参照（階層構造）
    
    // リレーション
    Parent   *Category  `gorm:"foreignKey:ParentId"`
    Children []Category `gorm:"foreignKey:ParentId"`
    Items    []Item     `gorm:"foreignKey:CategoryId"`
}
```

### 複雑なリレーションシップ

```go
// 多対多の関係
type Order struct {
    gorm.Model
    UserId      uint   `gorm:"not null"`
    TotalAmount uint   `gorm:"not null"`
    Status      string `gorm:"not null;default:'pending'"`
    
    // リレーション
    User  User        `gorm:"foreignKey:UserId"`
    Items []Item      `gorm:"many2many:order_items;"`
}

// 中間テーブルをカスタマイズ
type OrderItem struct {
    OrderId  uint `gorm:"primaryKey"`
    ItemId   uint `gorm:"primaryKey"`
    Quantity uint `gorm:"not null;default:1"`
    Price    uint `gorm:"not null"` // 購入時の価格を保存
    
    // リレーション
    Order Order `gorm:"foreignKey:OrderId"`
    Item  Item  `gorm:"foreignKey:ItemId"`
}

// ポリモーフィック関連
type Comment struct {
    gorm.Model
    Content       string `gorm:"not null;type:text"`
    UserId        uint   `gorm:"not null"`
    
    // ポリモーフィック
    CommentableId   uint   `gorm:"not null"`
    CommentableType string `gorm:"not null;size:50"`
    
    User User `gorm:"foreignKey:UserId"`
}

// 使用例
func (Item) TableName() string {
    return "items"
}

func (i *Item) GetComments() []Comment {
    var comments []Comment
    db.Where("commentable_type = ? AND commentable_id = ?", "items", i.ID).Find(&comments)
    return comments
}
```

### フックメソッドの実装

```go
// 作成前のフック
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // パスワードのハッシュ化
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}

// 更新前のフック
func (i *Item) BeforeUpdate(tx *gorm.DB) error {
    // 売り切れ商品の価格変更を防ぐ
    var oldItem Item
    tx.First(&oldItem, i.ID)
    
    if oldItem.SoldOut && oldItem.Price != i.Price {
        return errors.New("売り切れ商品の価格は変更できません")
    }
    return nil
}

// 削除前のフック（ソフトデリート）
func (u *User) BeforeDelete(tx *gorm.DB) error {
    // 関連データの確認
    var itemCount int64
    tx.Model(&Item{}).Where("user_id = ?", u.ID).Count(&itemCount)
    
    if itemCount > 0 {
        return errors.New("商品が存在するユーザーは削除できません")
    }
    return nil
}

// カスタムメソッド
func (i *Item) IsPurchasable() bool {
    return !i.SoldOut && i.Price > 0
}

func (u *User) GetDisplayName() string {
    if u.DisplayName != "" {
        return u.DisplayName
    }
    return u.Username
}
```

### インデックスとマイグレーション

```go
// 複合インデックスの定義
type Product struct {
    gorm.Model
    Code  string `gorm:"uniqueIndex:idx_code_shop"`
    Shop  string `gorm:"uniqueIndex:idx_code_shop"`
    Price uint   `gorm:"index:idx_price_date"`
    Date  time.Time `gorm:"index:idx_price_date"`
}

// マイグレーション用の関数
func Migrate(db *gorm.DB) error {
    // テーブルの作成
    err := db.AutoMigrate(
        &User{},
        &Item{},
        &Category{},
        &Order{},
        &OrderItem{},
        &Comment{},
    )
    if err != nil {
        return err
    }
    
    // カスタムインデックスの追加
    db.Exec("CREATE INDEX idx_items_price_range ON items(price) WHERE sold_out = false")
    
    // 初期データの投入
    categories := []Category{
        {Name: "電化製品"},
        {Name: "書籍"},
        {Name: "衣類"},
    }
    db.Create(&categories)
    
    return nil
}
```

### バリデーションタグとデータベース制約

```go
type Product struct {
    gorm.Model
    
    // 文字列制約
    SKU         string `gorm:"unique;not null;size:20"`
    Name        string `gorm:"not null;size:200"`
    Description string `gorm:"type:text"`
    
    // 数値制約
    Price    decimal.Decimal `gorm:"type:decimal(10,2);not null"`
    Stock    int            `gorm:"not null;default:0;check:stock >= 0"`
    MinStock int            `gorm:"not null;default:0"`
    
    // 日付制約
    ReleaseDate *time.Time `gorm:"index"`
    ExpiryDate  *time.Time
    
    // ENUM型
    Status string `gorm:"type:enum('active','inactive','discontinued');default:'active'"`
    
    // JSON型
    Attributes datatypes.JSON `gorm:"type:json"`
}

// カスタム型の定義
type Status string

const (
    StatusActive       Status = "active"
    StatusInactive     Status = "inactive"
    StatusDiscontinued Status = "discontinued"
)

func (s *Status) Scan(value interface{}) error {
    *s = Status(value.(string))
    return nil
}

func (s Status) Value() (driver.Value, error) {
    return string(s), nil
}
```

## ディレクトリ構造
```
models/
├── README.md
├── item.go          # 商品モデル
├── user.go          # ユーザーモデル
├── order.go         # 注文モデル
├── category.go      # カテゴリーモデル
├── common.go        # 共通の型定義
└── hooks.go         # フックメソッドをまとめる場合
```

## ベストプラクティス

### 1. 命名規則
```go
// テーブル名は複数形
func (Item) TableName() string {
    return "items"
}

// 外部キーは`モデル名ID`
UserId uint `gorm:"not null"`

// 中間テーブルは`モデル1_モデル2`
// order_items, user_favorites
```

### 2. Null許容の扱い
```go
// Nullを許容する場合はポインタ型
type Profile struct {
    gorm.Model
    UserId      uint
    Bio         *string    // NULL許容
    BirthDate   *time.Time // NULL許容
    PhoneNumber string     // NOT NULL
}
```

### 3. ソフトデリート
```go
// gorm.Modelを使用すると自動的にソフトデリート対応
type Item struct {
    gorm.Model // DeletedAtフィールドが含まれる
}

// ハードデリートが必要な場合
db.Unscoped().Delete(&item)
```

## 注意事項
- Modelはデータ構造の定義に専念
- ビジネスロジックはService層に実装
- 複雑なクエリはRepository層に実装
- フックメソッドは最小限に留める
- パフォーマンスを考慮したインデックス設計
- 適切な制約でデータ整合性を保つ