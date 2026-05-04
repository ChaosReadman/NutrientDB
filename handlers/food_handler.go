package handlers

import (
	"albion-app/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type FoodHandler struct {
	DB    *sql.DB
	Store *session.Store
}

// Index は一覧と検索を表示します
func (h *FoodHandler) Index(c *fiber.Ctx) error {
	query := c.Query("q")
	sess, _ := h.Store.Get(c)
	user := sess.Get("username")

	foods, err := models.Search(h.DB, query)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("index", fiber.Map{
		"Title": "食品栄養素データベース",
		"User":  user,
		"Foods": foods,
		"Query": query,
	})
}

// Detail は詳細を表示します
func (h *FoodHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")
	sess, _ := h.Store.Get(c)
	user := sess.Get("username")

	food, err := models.GetByID(h.DB, id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if food == nil {
		return c.Status(404).SendString("食品が見つかりませんでした")
	}

	return c.Render("detail", fiber.Map{
		"Title": "詳細情報",
		"User":  user,
		"Food":  food,
	})
}
