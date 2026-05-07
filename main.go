package main

import (
	"database/sql"
	"log"
	"os"

	"RecipeApp/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3" // go get github.com/mattn/go-sqlite3 が必要
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found. Using system environment variables.")
	}

	// テンプレートエンジンの設定
	// Djangoのextendsに近い仕組みを「Layout」として指定できる
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layout", // views/layout.html をデフォルトで使用
	})

	// OAuth設定 (環境変数から読み込むのが一般的)
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/fitness.activity.read",   // 運動量（消費カロリー）取得用
			"https://www.googleapis.com/auth/fitness.body.read",       // 体重・基礎代謝計算用
			"https://www.googleapis.com/auth/fitness.nutrition.read",  // 栄養素の読み取り
			"https://www.googleapis.com/auth/fitness.nutrition.write", // 栄養素の書き込み
		},
		Endpoint: google.Endpoint,
	}

	// 環境変数が正しく設定されているかチェック
	if conf.ClientID == "" || conf.ClientSecret == "" || conf.RedirectURL == "" {
		log.Fatal("Missing required environment variables. Please check your .env file (refer to .env.example).")
	}

	// セッションストアの初期化
	store := session.New()

	// 静的ファイル (CSS, JS, 画像) の配信設定
	app.Static("/static", "./public")

	// データベース接続
	db, err := sql.Open("sqlite3", "./data/nutrient.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// テーブルの自動生成（簡易マイグレーション）
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS daily_health_data (
		user_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		steps INTEGER DEFAULT 0,
		burned_calories INTEGER DEFAULT 0,
		PRIMARY KEY (user_id, date)
	);`)
	if err != nil {
		log.Fatal("Failed to initialize table:", err)
	}

	// ハンドラの初期化
	foodHandler := &handlers.FoodHandler{DB: db, Store: store, OAuthConfig: conf}
	authHandler := &handlers.AuthHandler{DB: db, Store: store, OAuthConfig: conf}

	// 認証チェック用ミドルウェア
	authRequired := func(c *fiber.Ctx) error {
		sess, _ := store.Get(c)
		if sess.Get("user_id") == nil {
			// ログインしていなければログイン画面へ
			return c.Redirect("/login")
		}
		return c.Next()
	}

	// ルーティング
	app.Get("/", foodHandler.Index)
	// 詳細ページはログイン必須にする例
	app.Get("/food/:id", authRequired, foodHandler.Detail)
	app.Get("/login", authHandler.ShowLogin)
	app.Get("/auth/login", authHandler.Login)
	app.Get("/auth/callback", authHandler.Callback)
	app.Get("/logout", authHandler.Logout)

	// 材料リスト操作
	app.Post("/ingredients/add", foodHandler.AddIngredient)
	app.Get("/api/foods/search", foodHandler.SearchJSON)
	app.Post("/ingredients/remove/:id", foodHandler.RemoveIngredient)

	// レシピ操作
	app.Get("/recipe/new", authRequired, foodHandler.NewRecipe)
	app.Post("/recipe/create", authRequired, foodHandler.CreateRecipe)
	app.Get("/recipe/:id", authRequired, foodHandler.RecipeDetail)
	app.Get("/recipe/:id/edit", authRequired, foodHandler.EditRecipe)
	app.Post("/recipe/:id/update", authRequired, foodHandler.UpdateRecipe)

	// カレンダー & 健康データ
	app.Get("/calendar", authRequired, foodHandler.CalendarIndex)
	app.Post("/calendar/add", authRequired, foodHandler.AddToCalendar)
	app.Post("/api/health/disconnect", authRequired, foodHandler.DisconnectHealthData)
	app.Get("/api/recipes/search", authRequired, foodHandler.SearchRecipesJSON)
	app.Get("/api/health/sync", authRequired, foodHandler.SyncHealthData)

	// サーバー起動
	log.Fatal(app.Listen(":3000"))
}
