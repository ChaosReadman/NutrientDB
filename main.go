package main

import (
	"database/sql"
	"log"

	"albion-app/handlers" // モジュール名に合わせて適宜変更してください

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	_ "github.com/mattn/go-sqlite3" // go get github.com/mattn/go-sqlite3 が必要
)

func main() {
	// テンプレートエンジンの設定
	// Djangoのextendsに近い仕組みを「Layout」として指定できる
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main", // デフォルトのレイアウトファイルを指定
	})

	// セッションストアの初期化
	store := session.New()

	// 静的ファイル (CSS, JS, 画像) の配信設定
	app.Static("/static", "./public")

	// データベース接続 (読み取り専用で開くのが安全です)
	db, err := sql.Open("sqlite3", "./data/nutrient.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// ハンドラの初期化
	foodHandler := &handlers.FoodHandler{DB: db, Store: store}
	authHandler := &handlers.AuthHandler{DB: db, Store: store}

	// ルーティング
	app.Get("/", foodHandler.Index)
	app.Get("/food/:id", foodHandler.Detail)
	app.Get("/register", authHandler.ShowRegister)
	app.Post("/register", authHandler.HandleRegister)
	app.Get("/login", authHandler.ShowLogin)
	app.Post("/login", authHandler.HandleLogin)
	app.Get("/logout", authHandler.Logout)

	// サーバー起動
	log.Fatal(app.Listen(":3000"))
}
