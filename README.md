# NutrientDB - 食品栄養素データベース

食品成分表のXMLデータをSQLiteに変換し、Go (Fiber) で閲覧・検索するためのWebアプリケーションです。

## 1. 環境構築

```bash
cd NutrientDB/

# Goモジュールの初期化と依存関係のインストール
go mod init albion-app
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/template/html/v2
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
go mod tidy

# ホットリロードツール (Air) のインストール
go install github.com/air-verse/air@latest
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

## 2. データベースの準備

XMLデータからSQLiteデータベース（`data/nutrient.db`）を生成します。

```bash
python3 data/xml_to_sqlite.py
```

## 3. 開発サーバーの起動（ホットリロード有効）

```bash
air
```
