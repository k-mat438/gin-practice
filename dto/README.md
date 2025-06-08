# DTO (Data Transfer Object)

## 概要
DTOは、クライアントとサーバー間でデータを転送するためのオブジェクトです。APIのリクエスト/レスポンスの構造を定義し、バリデーションルールを含みます。

## 責務
- APIリクエストの構造定義
- APIレスポンスの構造定義
- 入力値のバリデーションルール定義
- JSONタグによるフィールドマッピング
- 必須/任意フィールドの明確化

## 書くべきもの
- ✅ リクエスト/レスポンスの構造体
- ✅ バリデーションタグ
- ✅ JSONタグ
- ✅ カスタムバリデーション（必要に応じて）
- ❌ ビジネスロジック
- ❌ データベース操作
- ❌ 複雑な処理

## 実装例

### 基本的なDTO

#### リクエストDTO
```go
// 作成用DTO（必須フィールド）
type CreateItemInput struct {
    Name        string `json:"name" binding:"required,min=2,max=100"`
    Price       uint   `json:"price" binding:"required,min=1,max=999999"`
    Description string `json:"description" binding:"max=500"`
    CategoryId  uint   `json:"category_id" binding:"required"`
}

// 更新用DTO（部分更新対応）
type UpdateItemInput struct {
    Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
    Price       *uint   `json:"price" binding:"omitempty,min=1,max=999999"`
    Description *string `json:"description" binding:"omitempty,max=500"`
    SoldOut     *bool   `json:"sold_out"`
}

// 検索用DTO
type SearchItemInput struct {
    Keyword     string `form:"keyword" binding:"max=100"`
    MinPrice    uint   `form:"min_price"`
    MaxPrice    uint   `form:"max_price" binding:"gtefield=MinPrice"`
    CategoryId  uint   `form:"category_id"`
    Page        int    `form:"page" binding:"min=1"`
    PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}
```

#### レスポンスDTO
```go
// 単一アイテムレスポンス
type ItemResponse struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Price       uint      `json:"price"`
    Description string    `json:"description"`
    SoldOut     bool      `json:"sold_out"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // 関連データ
    User     UserSummary     `json:"user"`
    Category CategorySummary `json:"category"`
}

// リスト用レスポンス
type ItemListResponse struct {
    Items      []ItemResponse `json:"items"`
    Pagination PaginationInfo `json:"pagination"`
}

// ページネーション情報
type PaginationInfo struct {
    CurrentPage int   `json:"current_page"`
    PageSize    int   `json:"page_size"`
    TotalItems  int64 `json:"total_items"`
    TotalPages  int   `json:"total_pages"`
}

// 関連データの要約
type UserSummary struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

type CategorySummary struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}
```

### カスタムバリデーション

```go
// カスタムバリデーション関数
func init() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("price_range", validatePriceRange)
        v.RegisterValidation("future_date", validateFutureDate)
    }
}

// 価格範囲のバリデーション
func validatePriceRange(fl validator.FieldLevel) bool {
    price := fl.Field().Uint()
    // ビジネスルール：価格は100円〜1,000,000円
    return price >= 100 && price <= 1000000
}

// 未来日付のバリデーション
func validateFutureDate(fl validator.FieldLevel) bool {
    date, ok := fl.Field().Interface().(time.Time)
    if !ok {
        return false
    }
    return date.After(time.Now())
}

// カスタムバリデーションを使用
type AuctionItemInput struct {
    Name      string    `json:"name" binding:"required"`
    StartPrice uint     `json:"start_price" binding:"required,price_range"`
    EndDate   time.Time `json:"end_date" binding:"required,future_date"`
}
```

### エラーレスポンスDTO

```go
// エラーレスポンス
type ErrorResponse struct {
    Error   string              `json:"error"`
    Details []ValidationError   `json:"details,omitempty"`
}

// バリデーションエラーの詳細
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// 使用例
func handleValidationErrors(err error) ErrorResponse {
    var details []ValidationError
    
    if validationErrs, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrs {
            details = append(details, ValidationError{
                Field:   e.Field(),
                Message: getErrorMessage(e),
            })
        }
    }
    
    return ErrorResponse{
        Error:   "Validation failed",
        Details: details,
    }
}
```

### 複雑なDTOの例

```go
// ネストした構造
type CreateOrderInput struct {
    Items []OrderItemInput `json:"items" binding:"required,min=1,dive"`
    
    // 配送情報
    Shipping ShippingInput `json:"shipping" binding:"required"`
    
    // 支払い情報
    Payment PaymentInput `json:"payment" binding:"required"`
}

type OrderItemInput struct {
    ItemId   uint `json:"item_id" binding:"required"`
    Quantity uint `json:"quantity" binding:"required,min=1,max=10"`
}

type ShippingInput struct {
    Name       string `json:"name" binding:"required"`
    PostalCode string `json:"postal_code" binding:"required,len=7"`
    Address    string `json:"address" binding:"required"`
    Phone      string `json:"phone" binding:"required,e164"`
}

type PaymentInput struct {
    Method string `json:"method" binding:"required,oneof=credit_card bank_transfer"`
    
    // クレジットカード情報（methodがcredit_cardの場合のみ）
    CardNumber string `json:"card_number" binding:"required_if=Method credit_card"`
    ExpiryDate string `json:"expiry_date" binding:"required_if=Method credit_card"`
}
```

## バリデーションタグ一覧

```go
// よく使うバリデーションタグ
type ExampleDTO struct {
    // 必須
    Required string `binding:"required"`
    
    // 文字列長
    MinLength  string `binding:"min=2"`
    MaxLength  string `binding:"max=100"`
    FixedLength string `binding:"len=10"`
    
    // 数値範囲
    MinValue uint `binding:"min=1"`
    MaxValue uint `binding:"max=999"`
    
    // 正規表現
    AlphaNum string `binding:"alphanum"`
    Email    string `binding:"email"`
    URL      string `binding:"url"`
    
    // 列挙値
    Status string `binding:"oneof=active inactive pending"`
    
    // 条件付き必須
    ConditionalField string `binding:"required_if=Status active"`
    
    // フィールド間の比較
    StartDate time.Time `binding:"required"`
    EndDate   time.Time `binding:"required,gtefield=StartDate"`
    
    // 配列/スライス
    Tags []string `binding:"min=1,max=5,dive,min=2,max=20"`
}
```

## ディレクトリ構造
```
dto/
├── README.md
├── item_dto.go         # 商品関連のDTO
├── user_dto.go         # ユーザー関連のDTO
├── auth_dto.go         # 認証関連のDTO
├── common_dto.go       # 共通DTO（ページネーション等）
└── validation.go       # カスタムバリデーション
```

## ModelとDTOの変換

```go
// Service層での変換例
func (s *ItemService) toItemResponse(item *models.Item) *dto.ItemResponse {
    return &dto.ItemResponse{
        ID:          item.ID,
        Name:        item.Name,
        Price:       item.Price,
        Description: item.Description,
        SoldOut:     item.SoldOut,
        CreatedAt:   item.CreatedAt,
        UpdatedAt:   item.UpdatedAt,
        User: dto.UserSummary{
            ID:   item.User.ID,
            Name: item.User.Name,
        },
    }
}

// CreateItemInputからModelへの変換
func (s *ItemService) toItemModel(input dto.CreateItemInput) models.Item {
    return models.Item{
        Name:        input.Name,
        Price:       input.Price,
        Description: input.Description,
        CategoryId:  input.CategoryId,
        SoldOut:     false, // デフォルト値
    }
}
```

## 注意事項
- DTOはデータ転送専用（ロジックを含めない）
- 適切なバリデーションタグを設定する
- nilを許容する場合はポインタ型を使用
- レスポンスDTOは必要な情報のみを含める（セキュリティ）
- 複雑すぎるDTOは避ける（必要に応じて分割）
- エラーメッセージは分かりやすく