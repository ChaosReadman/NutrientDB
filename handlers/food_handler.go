package handlers

import (
	"albion-app/models"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

type FoodHandler struct {
	DB *sql.DB
}

// Index は一覧と検索を表示します
func (h *FoodHandler) Index(c *fiber.Ctx) error {
	query := c.Query("q")

	foods, err := models.Search(h.DB, query)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Render("index", fiber.Map{
		"Title": "食品栄養素データベース",
		"Foods": foods,
		"Query": query,
	})
}

// Detail は詳細を表示します
func (h *FoodHandler) Detail(c *fiber.Ctx) error {
	id := c.Params("id")

	food, err := models.GetByID(h.DB, id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if food == nil {
		return c.Status(404).SendString("食品が見つかりませんでした")
	}

	return c.Render("detail", fiber.Map{
		"Title": "詳細情報",
		"Food":  food,
	})
}
