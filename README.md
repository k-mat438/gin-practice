# Gin Flea Market API

## 🏗️ プロジェクト構成

このプロジェクトは、**レイヤードアーキテクチャ**（層状アーキテクチャ）を採用したGin Frameworkベースのフリーマーケットアプリケーションです。

### アーキテクチャ概要

```
HTTP Request
    ↓
Controller層 (Presentation Layer)
    ↓
Service層 (Business Logic Layer)
    ↓
Repository層 (Data Access Layer)
    ↓
Database / Memory Storage
```

## 📁 ディレクトリ構成

```
gin-fleamarket/
├── main.go                 # エントリーポイント
├── controllers/           # プレゼンテーション層
│   ├── README.md
│   └── item_controller.go
├── services/              # ビジネスロジック層
│   ├── README.md
│   └── item_services.go
├── repositories/          # データアクセス層
│   ├── README.md
│   └── item_repository.go
├── models/                # データモデル定義
│   ├── README.md
│   └── item.go
├── dto/                   # データ転送オブジェクト
│   ├── README.md
│   └── item_dto.go
├── infra/                 # インフラストラクチャ層
│   ├── README.md
│   ├── db.go
│   └── initializer.go
└── migrations/            # データベースマイグレーション
    └── migration.go
```

## 🎯 各層の責務

### 1. **Controller層** (`controllers/`)
- HTTPリクエスト/レスポンスの処理
- リクエストパラメータの検証
- HTTPステータスコードの設定
- Service層への処理委譲

### 2. **Service層** (`services/`)
- ビジネスロジックの実装
- トランザクション管理
- DTOからModelへの変換
- 複数のRepositoryの協調

### 3. **Repository層** (`repositories/`)
- データの永続化とアクセス
- CRUDオペレーション
- データベース固有の処理
- クエリの最適化

### 4. **Model層** (`models/`)
- データベーステーブルの構造定義
- リレーションシップの定義
- GORMタグによる制約

### 5. **DTO層** (`dto/`)
- APIリクエスト/レスポンスの型定義
- バリデーションルール
- データ転送用の構造体

### 6. **Infrastructure層** (`infra/`)
- データベース接続管理
- 設定の初期化
- 外部サービスとの接続
- ミドルウェアの設定

## 🚀 セットアップ

### 必要な環境
- Go 1.19以上
- MySQL 8.0以上（またはDocker）

### インストール手順

1. リポジトリのクローン
```bash
git clone https://github.com/yourusername/gin-fleamarket.git
cd gin-fleamarket
```

2. 依存関係のインストール
```bash
go mod download
```

3. 環境変数の設定
```bash
cp .env.example .env
# .envファイルを編集して必要な値を設定
```

4. データベースの起動（Docker使用時）
```bash
docker-compose up -d
```

5. マイグレーションの実行
```bash
go run migrations/migration.go
```

6. アプリケーションの起動
```bash
go run main.go
```

### 開発モード（ホットリロード）
```bash
air
```

## 📝 API仕様

### 商品関連エンドポイント

#### 商品一覧取得
```
GET /api/v1/items
```

#### 商品詳細取得
```
GET /api/v1/items/:id
```

#### 商品作成
```
POST /api/v1/items
Content-Type: application/json

{
  "name": "商品名",
  "price": 1000,
  "description": "商品の説明"
}
```

#### 商品更新
```
PUT /api/v1/items/:id
Content-Type: application/json

{
  "name": "更新後の商品名",
  "price": 1500,
  "description": "更新後の説明",
  "sold_out": true
}
```

#### 商品削除
```
DELETE /api/v1/items/:id
```

## 🧪 テスト

### ユニットテストの実行
```bash
go test ./...
```

### カバレッジレポート
```bash
go test -cover ./...
```

### 統合テストの実行
```bash
go test -tags=integration ./...
```

## 🔧 開発ガイドライン

### コーディング規約
- 各層の責務を明確に分離する
- インターフェースを使用して依存性を注入する
- エラーハンドリングを適切に行う
- テストを書く

### コミットメッセージ
```
feat: 新機能の追加
fix: バグ修正
docs: ドキュメントの更新
style: コードスタイルの修正
refactor: リファクタリング
test: テストの追加・修正
chore: ビルドプロセスやツールの変更
```

## 📚 詳細ドキュメント

各層の詳細な説明と実装例については、各ディレクトリのREADMEを参照してください：

- [Controllers層の詳細](./controllers/README.md)
- [Services層の詳細](./services/README.md)
- [Repositories層の詳細](./repositories/README.md)
- [Models層の詳細](./models/README.md)
- [DTO層の詳細](./dto/README.md)
- [Infrastructure層の詳細](./infra/README.md)

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'feat: Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。