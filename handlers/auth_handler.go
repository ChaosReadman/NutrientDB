package handlers

import (
	"albion-app/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type AuthHandler struct {
	DB    *sql.DB
	Store *session.Store
}

// ShowRegister は登録画面を表示します
func (h *AuthHandler) ShowRegister(c *fiber.Ctx) error {
	return c.Render("register", fiber.Map{
		"Title": "ユーザー登録",
	})
}

// ShowLogin はログイン画面を表示します
func (h *AuthHandler) ShowLogin(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{
		"Title": "ログイン",
	})
}

// HandleRegister は登録処理を実行します
func (h *AuthHandler) HandleRegister(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.Status(400).SendString("ユーザー名とパスワードを入力してください")
	}

	if err := models.Register(h.DB, username, password); err != nil {
		return c.Status(500).SendString("ユーザー登録に失敗しました: " + err.Error())
	}

	// 登録成功後、そのままログイン状態にする
	sess, _ := h.Store.Get(c)
	sess.Set("username", username)
	sess.Save()

	return c.Redirect("/")
}

// HandleLogin はログイン処理を実行します
func (h *AuthHandler) HandleLogin(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	user, err := models.Authenticate(h.DB, username, password)
	if err != nil {
		return c.Status(401).SendString("ユーザー名またはパスワードが正しくありません")
	}

	// セッションに保存
	sess, _ := h.Store.Get(c)
	sess.Set("username", user.Username)
	if err := sess.Save(); err != nil {
		return c.Status(500).SendString(err.Error())
	}

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
