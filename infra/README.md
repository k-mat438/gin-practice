# Infrastructure層

## 概要
Infrastructure層は、アプリケーションの基盤となる設定や外部サービスとの接続を管理します。データベース接続、設定の初期化、外部APIクライアントなどを含みます。

## 責務
- データベース接続の管理
- 設定ファイルの読み込み
- 環境変数の管理
- 外部サービスとの接続
- ミドルウェアの設定
- ロギングの設定

## 書くべきもの
- ✅ データベース接続設定
- ✅ 設定の初期化処理
- ✅ 外部APIクライアント
- ✅ キャッシュの設定
- ✅ メッセージキューの接続
- ❌ ビジネスロジック
- ❌ アプリケーション固有の処理

## 実装例

### データベース接続 (`db.go`)

```go
package infra

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

// データベース接続の初期化
func InitDB() (*gorm.DB, error) {
    // 環境変数から接続情報を取得
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        getEnv("DB_USER", "root"),
        getEnv("DB_PASSWORD", "password"),
        getEnv("DB_HOST", "localhost"),
        getEnv("DB_PORT", "3306"),
        getEnv("DB_NAME", "fleamarket"),
    )
    
    // ログレベルの設定
    logLevel := logger.Silent
    if getEnv("APP_ENV", "development") == "development" {
        logLevel = logger.Info
    }
    
    // カスタムロガーの設定
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logLevel,
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    )
    
    // データベース接続
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: newLogger,
        NowFunc: func() time.Time {
            return time.Now().Local()
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }
    
    // コネクションプールの設定
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get database instance: %w", err)
    }
    
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    DB = db
    return db, nil
}

// トランザクション処理のヘルパー
func Transaction(fn func(*gorm.DB) error) error {
    return DB.Transaction(fn)
}

// テスト用のインメモリDB
func InitTestDB() (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, err
    }
    
    // テスト用のマイグレーション
    err = db.AutoMigrate(
        &models.User{},
        &models.Item{},
        &models.Category{},
    )
    if err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### 設定管理 (`config.go`)

```go
package infra

import (
    "github.com/spf13/viper"
)

type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    AWS      AWSConfig
}

type AppConfig struct {
    Name        string
    Port        string
    Environment string
    Debug       bool
    LogLevel    string
}

type DatabaseConfig struct {
    Host            string
    Port            string
    User            string
    Password        string
    Name            string
    MaxIdleConns    int
    MaxOpenConns    int
    ConnMaxLifetime int
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type JWTConfig struct {
    Secret          string
    ExpirationHours int
}

type AWSConfig struct {
    Region          string
    AccessKeyID     string
    SecretAccessKey string
    S3Bucket        string
}

var AppConfig *Config

// 設定の初期化
func InitConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    
    // 環境変数の自動読み込み
    viper.AutomaticEnv()
    
    // デフォルト値の設定
    setDefaults()
    
    // 設定ファイルの読み込み
    if err := viper.ReadInConfig(); err != nil {
        // 設定ファイルが見つからない場合は環境変数のみを使用
        log.Printf("Config file not found, using environment variables: %v", err)
    }
    
    // 構造体にバインド
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    AppConfig = &config
    return &config, nil
}

func setDefaults() {
    // アプリケーション設定
    viper.SetDefault("app.name", "Flea Market API")
    viper.SetDefault("app.port", "8080")
    viper.SetDefault("app.environment", "development")
    viper.SetDefault("app.debug", true)
    
    // データベース設定
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", "3306")
    viper.SetDefault("database.maxIdleConns", 10)
    viper.SetDefault("database.maxOpenConns", 100)
    viper.SetDefault("database.connMaxLifetime", 3600)
    
    // Redis設定
    viper.SetDefault("redis.host", "localhost")
    viper.SetDefault("redis.port", "6379")
    viper.SetDefault("redis.db", 0)
}
```

### 初期化処理 (`initializer.go`)

```go
package infra

import (
    "fmt"
    "gin-fleamarket/models"
    "gin-fleamarket/repositories"
    "gin-fleamarket/services"
    "gin-fleamarket/controllers"
    
    "github.com/gin-gonic/gin"
)

type Container struct {
    // Repositories
    ItemRepository repositories.IItemRepository
    UserRepository repositories.IUserRepository
    
    // Services
    ItemService services.IItemService
    UserService services.IUserService
    
    // Controllers
    ItemController controllers.IItemController
    UserController controllers.IUserController
}

// 依存性注入コンテナの初期化
func NewContainer(db *gorm.DB) *Container {
    // Repositories
    itemRepo := repositories.NewItemRepository(db)
    userRepo := repositories.NewUserRepository(db)
    
    // Services
    itemService := services.NewItemService(itemRepo)
    userService := services.NewUserService(userRepo)
    
    // Controllers
    itemController := controllers.NewItemController(itemService)
    userController := controllers.NewUserController(userService)
    
    return &Container{
        ItemRepository: itemRepo,
        UserRepository: userRepo,
        ItemService:    itemService,
        UserService:    userService,
        ItemController: itemController,
        UserController: userController,
    }
}

// ルーターの設定
func SetupRouter(container *Container) *gin.Engine {
    router := gin.Default()
    
    // ミドルウェアの設定
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    router.Use(CORSMiddleware())
    
    // APIグループ
    api := router.Group("/api/v1")
    {
        // 商品関連のルート
        items := api.Group("/items")
        {
            items.GET("", container.ItemController.FindAll)
            items.GET("/:id", container.ItemController.FindById)
            items.POST("", container.ItemController.Create)
            items.PUT("/:id", container.ItemController.Update)
            items.DELETE("/:id", container.ItemController.Delete)
        }
        
        // ユーザー関連のルート
        users := api.Group("/users")
        {
            users.GET("/:id", container.UserController.FindById)
            users.POST("/register", container.UserController.Register)
            users.POST("/login", container.UserController.Login)
        }
    }
    
    return router
}

// CORSミドルウェア
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### 外部サービスクライアント (`external/`)

```go
// external/s3_client.go
package external

import (
    "bytes"
    "fmt"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

type S3Client struct {
    client *s3.S3
    bucket string
}

func NewS3Client(region, bucket string) (*S3Client, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create AWS session: %w", err)
    }
    
    return &S3Client{
        client: s3.New(sess),
        bucket: bucket,
    }, nil
}

func (c *S3Client) Upload(key string, data []byte, contentType string) error {
    _, err := c.client.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(c.bucket),
        Key:         aws.String(key),
        Body:        bytes.NewReader(data),
        ContentType: aws.String(contentType),
    })
    
    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }
    
    return nil
}

// external/redis_client.go
package external

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type RedisClient struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisClient(addr, password string, db int) *RedisClient {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    
    return &RedisClient{
        client: client,
        ctx:    context.Background(),
    }
}

func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return r.client.Set(r.ctx, key, data, expiration).Err()
}

func (r *RedisClient) Get(key string, dest interface{}) error {
    data, err := r.client.Get(r.ctx, key).Result()
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(data), dest)
}
```

## ディレクトリ構造
```
infra/
├── README.md
├── db.go              # データベース接続
├── config.go          # 設定管理
├── initializer.go     # 初期化処理
├── middleware.go      # ミドルウェア
├── logger.go          # ロギング設定
└── external/          # 外部サービスクライアント
    ├── s3_client.go
    ├── redis_client.go
    └── mail_client.go
```

## 環境変数の例
```bash
# .env.example
APP_NAME=FleaMarketAPI
APP_PORT=8080
APP_ENV=development

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=fleamarket

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24

AWS_REGION=ap-northeast-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET=fleamarket-images
```

## 注意事項
- 環境固有の設定は環境変数で管理
- シークレット情報はコードに直接書かない
- 接続エラーは適切にハンドリング
- リソースのクリーンアップを忘れない
- テスト環境用の設定を用意する
- 設定の変更が容易になるよう設計する