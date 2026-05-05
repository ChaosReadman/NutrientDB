package handlers

import (
	"RecipeApp/models"
	"context"
	"database/sql"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	DB          *sql.DB
	Store       *session.Store
	OAuthConfig *oauth2.Config
}

// ShowLogin はログイン画面を表示します
func (h *AuthHandler) ShowLogin(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"Title": "ログイン",
	})
}

// Login リダイレクト処理
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	url := h.OAuthConfig.AuthCodeURL("state")
	return c.Redirect(url)
}

// Callback OAuthからの戻り先
func (h *AuthHandler) Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	token, err := h.OAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(500).SendString("トークンの取得に失敗しました")
	}

	// Googleからユーザー情報を取得
	client := h.OAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(500).SendString("ユーザー情報の取得に失敗しました")
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		return c.Status(resp.StatusCode).SendString("Google APIからのデータ取得に失敗しました (Status: " + resp.Status + ")")
	}

	var profile struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&profile)

	// DBでユーザーを特定または作成
	user, err := models.FindOrCreate(h.DB, profile.Email, profile.Name, "google", profile.ID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// セッションに名前をセット
	sess, _ := h.Store.Get(c)
	sess.Set("user_id", user.ID)
	sess.Set("username", user.Name)
	sess.Save()

	return c.Redirect("/")
}

// Logout はログアウト処理を実行します
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sess, err := h.Store.Get(c)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	if err := sess.Destroy(); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Redirect("/")
}
