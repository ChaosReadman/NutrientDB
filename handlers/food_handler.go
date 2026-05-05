package handlers

import (
	"RecipeApp/models"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type FoodHandler struct {
	DB    *sql.DB
	Store *session.Store
}

// Ingredient は材料リストのアイテム構造体です
type Ingredient struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	GroupName string  `json:"group_name"`
}

// Index は一覧と検索を表示します
func (h *FoodHandler) Index(c *fiber.Ctx) error {
	query := c.Query("q")
	recipeQuery := c.Query("rq")
	sess, _ := h.Store.Get(c)
	user := sess.Get("username")

	// レシピ検索結果と主要レシピ
	recipes, err := models.SearchRecipes(h.DB, recipeQuery)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// 【修正】検索クエリがある場合のみ食品を検索する（これで初期画面の MyIngredients が表示される）
	var foods []models.Food
	if query != "" {
		foods, err = models.Search(h.DB, query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
	}

	return c.Render("index", fiber.Map{
		"Title":         "食品栄養素データベース",
		"User":          user,
		"Foods":         foods,
		"Query":         query,
		"RecipeQuery":   recipeQuery,
		"Ingredients":   h.getIngredientsFromSession(c),
		"MyIngredients": h.getMyIngredients(c),
		"Recipes":       recipes,
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
		"Title":         "詳細情報",
		"User":          user,
		"Food":          food,
		"Ingredients":   h.getIngredientsFromSession(c),
		"MyIngredients": h.getMyIngredients(c),
	})
}

// AddIngredient は材料リストにアイテムを追加します
func (h *FoodHandler) AddIngredient(c *fiber.Ctx) error {
	id := c.FormValue("id")
	name := c.FormValue("name")

	ingredients := h.getIngredientsFromSession(c)

	// 重複チェック（任意）
	for _, item := range ingredients {
		if item.ID == id {
			return c.Redirect("/")
		}
	}

	ingredients = append(ingredients, Ingredient{ID: id, Name: name})
	sess, _ := h.Store.Get(c)
	data, _ := json.Marshal(ingredients)
	sess.Set("ingredients", string(data))
	sess.Save()

	return c.Redirect("/")
}

// SearchJSON は食品を検索し JSON 形式で返します
func (h *FoodHandler) SearchJSON(c *fiber.Ctx) error {
	query := c.Query("q")
	foods, err := models.Search(h.DB, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(foods)
}

// RemoveIngredient は材料リストからアイテムを削除します
func (h *FoodHandler) RemoveIngredient(c *fiber.Ctx) error {
	id := c.Params("id")

	ingredients := h.getIngredientsFromSession(c)

	newIngredients := []Ingredient{}
	for _, item := range ingredients {
		if item.ID != id {
			newIngredients = append(newIngredients, item)
		}
	}

	data, _ := json.Marshal(newIngredients)
	sess, _ := h.Store.Get(c)
	if len(newIngredients) == 0 {
		sess.Delete("ingredients")
	} else {
		sess.Set("ingredients", string(data))
	}
	_ = sess.Save()

	return c.Redirect("/")
}

// NewRecipe はレシピ作成画面を表示します
func (h *FoodHandler) NewRecipe(c *fiber.Ctx) error {
	sess, _ := h.Store.Get(c)
	user := sess.Get("username")

	ingredients := h.getIngredientsFromSession(c)
	if len(ingredients) == 0 {
		return c.Redirect("/")
	}

	return c.Render("recipe_new", fiber.Map{
		"Title":         "レシピ作成",
		"User":          user,
		"Ingredients":   ingredients,
		"MyIngredients": h.getMyIngredients(c),
		"IsRecipePage":  true,
	})
}

// CreateRecipe はレシピをデータベースに保存します
func (h *FoodHandler) CreateRecipe(c *fiber.Ctx) error {
	sess, _ := h.Store.Get(c)
	userID := sess.Get("user_id").(int)

	title := c.FormValue("title")
	description := c.FormValue("description")

	tx, err := h.DB.Begin()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	res, err := tx.Exec("INSERT INTO recipes (user_id, title, description) VALUES (?, ?, ?)", userID, title, description)
	if err != nil {
		tx.Rollback()
		return c.Status(500).SendString(err.Error())
	}
	recipeID, _ := res.LastInsertId()

	// 材料の保存 (フォームから送信された ingredient_ids を元に処理)
	ingIDs := c.Request().PostArgs().PeekMulti("ingredient_ids")
	for _, idByte := range ingIDs {
		id := string(idByte)
		qty := c.FormValue("qty_" + id)
		grp := c.FormValue("grp_" + id)

		if _, err := tx.Exec("INSERT INTO recipe_ingredients (recipe_id, food_id, quantity, group_name) VALUES (?, ?, ?, ?)", recipeID, id, qty, grp); err != nil {
			tx.Rollback()
			return c.Status(500).SendString("材料の保存に失敗しました: " + err.Error())
		}
	}

	// 工程の保存 (複数値の取得)
	steps := c.Request().PostArgs().PeekMulti("steps")
	for i, stepByte := range steps {
		if _, err := tx.Exec("INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES (?, ?, ?)", recipeID, i+1, string(stepByte)); err != nil {
			tx.Rollback()
			return c.Status(500).SendString("工程の保存に失敗しました: " + err.Error())
		}
	}

	tx.Commit()

	// セッションの材料リストをクリア
	sess.Delete("ingredients")
	sess.Save()

	return c.Redirect("/")
}

// RecipeDetail はレシピの詳細画面を表示します
func (h *FoodHandler) RecipeDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	sess, _ := h.Store.Get(c)
	user := sess.Get("username")
	ingredients := h.getIngredientsFromSession(c)

	recipe, err := models.GetRecipeByID(h.DB, id)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if recipe == nil {
		return c.Status(404).SendString("レシピが見つかりませんでした")
	}

	isOwner := false
	if userID := sess.Get("user_id"); userID != nil && userID.(int) == recipe.UserID {
		isOwner = true
	}

	return c.Render("recipe_detail", fiber.Map{
		"Title":                recipe.Title,
		"User":                 user,
		"Recipe":               recipe,
		"Ingredients":          ingredients,
		"IsOwner":              isOwner,
		"HideIngredientDrawer": true,
	})
}

// EditRecipe はレシピの編集画面を表示します
func (h *FoodHandler) EditRecipe(c *fiber.Ctx) error {
	id := c.Params("id")
	sess, _ := h.Store.Get(c)
	user := sess.Get("username")
	userID := sess.Get("user_id").(int)

	recipe, err := models.GetRecipeByID(h.DB, id)
	if err != nil || recipe == nil {
		return c.Status(404).SendString("レシピが見つかりません")
	}

	if recipe.UserID != userID {
		return c.Status(403).SendString("編集権限がありません")
	}

	// 編集時はレシピの材料をセッションの「材料リスト」に同期する
	var ingredients []Ingredient
	for _, ing := range recipe.Ingredients {
		ingredients = append(ingredients, Ingredient{
			ID:        ing.FoodID,
			Name:      ing.Name,
			Quantity:  ing.Quantity,
			GroupName: ing.GroupName,
		})
	}
	data, _ := json.Marshal(ingredients)
	sess.Set("ingredients", string(data))
	_ = sess.Save()

	return c.Render("recipe_edit", fiber.Map{
		"Title":         "レシピ編集",
		"User":          user,
		"Recipe":        recipe,
		"Ingredients":   ingredients,
		"MyIngredients": h.getMyIngredients(c),
		"IsRecipePage":  true,
	})
}

// UpdateRecipe はレシピを更新します
func (h *FoodHandler) UpdateRecipe(c *fiber.Ctx) error {
	id := c.Params("id")
	sess, _ := h.Store.Get(c)
	userID := sess.Get("user_id").(int)

	recipe, err := models.GetRecipeByID(h.DB, id)
	if err != nil || recipe == nil {
		return c.Status(404).SendString("レシピが見つかりません")
	}

	if recipe.UserID != userID {
		return c.Status(403).SendString("編集権限がありません")
	}

	tx, err := h.DB.Begin()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// 基本情報の更新
	_, err = tx.Exec("UPDATE recipes SET title = ?, description = ? WHERE id = ?", c.FormValue("title"), c.FormValue("description"), id)
	if err != nil {
		tx.Rollback()
		return c.Status(500).SendString(err.Error())
	}

	// 材料と工程は一度削除して再登録するのが最も確実
	if _, err := tx.Exec("DELETE FROM recipe_ingredients WHERE recipe_id = ?", id); err != nil {
		tx.Rollback()
		return c.Status(500).SendString(err.Error())
	}
	if _, err := tx.Exec("DELETE FROM recipe_steps WHERE recipe_id = ?", id); err != nil {
		tx.Rollback()
		return c.Status(500).SendString(err.Error())
	}

	// 材料の再保存
	ingIDs := c.Request().PostArgs().PeekMulti("ingredient_ids")
	for _, idByte := range ingIDs {
		ingID := string(idByte)
		qty := c.FormValue("qty_" + ingID)
		grp := c.FormValue("grp_" + ingID)
		if _, err := tx.Exec("INSERT INTO recipe_ingredients (recipe_id, food_id, quantity, group_name) VALUES (?, ?, ?, ?)", id, ingID, qty, grp); err != nil {
			tx.Rollback()
			return c.Status(500).SendString("材料の更新に失敗しました")
		}
	}

	// 工程の再保存
	steps := c.Request().PostArgs().PeekMulti("steps")
	for i, stepByte := range steps {
		if _, err := tx.Exec("INSERT INTO recipe_steps (recipe_id, step_number, instruction) VALUES (?, ?, ?)", id, i+1, string(stepByte)); err != nil {
			tx.Rollback()
			return c.Status(500).SendString("工程の更新に失敗しました")
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Redirect("/recipe/" + id)
}

// getIngredientsFromSession はセッションから材料リストを取得するヘルパーメソッドです
func (h *FoodHandler) getIngredientsFromSession(c *fiber.Ctx) []Ingredient {
	sess, _ := h.Store.Get(c)
	var ingredients []Ingredient
	if raw := sess.Get("ingredients"); raw != nil {
		json.Unmarshal([]byte(raw.(string)), &ingredients)
	}

	// 材料リストのソート: グループなしを最優先し、次にグループ名、最後に名称でソート
	sort.Slice(ingredients, func(i, j int) bool {
		emptyI := ingredients[i].GroupName == ""
		emptyJ := ingredients[j].GroupName == ""
		if emptyI != emptyJ {
			return emptyI // iが空ならtrueを返し、jより前にくる
		}
		// グループ名で比較
		if ingredients[i].GroupName != ingredients[j].GroupName {
			return ingredients[i].GroupName < ingredients[j].GroupName
		}
		// 名称で比較
		return ingredients[i].Name < ingredients[j].Name
	})

	return ingredients
}

// getMyIngredients はユーザーの履歴またはデフォルトの材料を取得します
func (h *FoodHandler) getMyIngredients(c *fiber.Ctx) []models.Food {
	sess, _ := h.Store.Get(c)
	userID := sess.Get("user_id")

	var myIngredients []models.Food
	if userID != nil {
		myIngredients, _ = models.GetUserRecipeIngredients(h.DB, userID.(int))
	}

	if len(myIngredients) == 0 {
		data, err := os.ReadFile("./data/default_ingredients.json")
		if err == nil {
			json.Unmarshal(data, &myIngredients)
		} else {
			log.Println("Warning: Could not load default_ingredients.json:", err)
		}
	}
	return myIngredients
}
