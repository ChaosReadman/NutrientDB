# NutrientDB - 食品栄養素データベース

食品成分表のXMLデータをSQLiteに変換し、Go (Fiber) で閲覧・検索するためのWebアプリケーションです。

## 1. 環境構築

```bash
cd NutrientDB/

# Goモジュールの初期化と依存関係のインストール
go mod init RecipeApp
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/template/html/v2
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv
go get golang.org/x/oauth2
go get golang.org/x/oauth2/google
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

実行には環境変数の設定が必要です。

プロジェクトルートに `.env` ファイルを作成してください：
※ 値はチームの共有パスワードマネージャーを参照してください。

```env
GOOGLE_CLIENT_ID="あなたのクライアントID"
GOOGLE_CLIENT_SECRET="あなたのクライアントシークレット"
GOOGLE_REDIRECT_URL="http://localhost:3000/auth/callback"
```

その後、以下のコマンドで起動します：

```bash
air
```
