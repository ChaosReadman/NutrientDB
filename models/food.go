package models

import (
	"database/sql"
)

type Food struct {
	ID   string
	Name string
}

// Search は名前で食品を検索します
func Search(db *sql.DB, query string) ([]Food, error) {
	var rows *sql.Rows
	var err error

	if query != "" {
		rows, err = db.Query("SELECT food_id, name FROM foods WHERE name LIKE ? LIMIT 50", "%"+query+"%")
	} else {
		rows, err = db.Query("SELECT food_id, name FROM foods LIMIT 20")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foods []Food
	for rows.Next() {
		var f Food
		if err := rows.Scan(&f.ID, &f.Name); err != nil {
			return nil, err
		}
		foods = append(foods, f)
	}
	return foods, nil
}

// GetByID はIDから詳細情報を取得します（動的なマップを返します）
func GetByID(db *sql.DB, id string) (map[string]interface{}, error) {
	rows, err := db.Query("SELECT * FROM foods WHERE food_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	foodDetail := make(map[string]interface{})
	for i, colName := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			foodDetail[colName] = string(b)
		} else {
			foodDetail[colName] = val
		}
	}
	return foodDetail, nil
}

// GetUserRecipeIngredients はユーザーがレシピで使用したことのある材料を取得します
func GetUserRecipeIngredients(db *sql.DB, userID int) ([]Food, error) {
	query := `
		SELECT DISTINCT f.food_id, f.name 
		FROM foods f 
		JOIN recipe_ingredients ri ON f.food_id = ri.food_id 
		JOIN recipes r ON ri.recipe_id = r.id 
		WHERE r.user_id = ?
		LIMIT 20`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foods []Food
	for rows.Next() {
		var f Food
		if err := rows.Scan(&f.ID, &f.Name); err != nil {
			return nil, err
		}
		foods = append(foods, f)
	}
	return foods, nil
}

// SearchRecipes はレシピを検索または一覧取得します
func SearchRecipes(db *sql.DB, query string) ([]map[string]interface{}, error) {
	var rows *sql.Rows
	var err error
	if query != "" {
		rows, err = db.Query("SELECT id, title, description FROM recipes WHERE title LIKE ? OR description LIKE ? LIMIT 20", "%"+query+"%", "%"+query+"%")
	} else {
		rows, err = db.Query("SELECT id, title, description FROM recipes ORDER BY created_at DESC LIMIT 10")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []map[string]interface{}
	for rows.Next() {
		var id int
		var title, desc string
		rows.Scan(&id, &title, &desc)
		recipes = append(recipes, map[string]interface{}{"ID": id, "Title": title, "Description": desc})
	}
	return recipes, nil
}

type RecipeIngredientDetail struct {
	FoodID    string
	Name      string
	Quantity  float64
	GroupName string
}

type RecipeStepDetail struct {
	StepNumber  int
	Instruction string
}

type RecipeFull struct {
	ID          int
	UserID      int
	Title       string
	Description string
	Ingredients []RecipeIngredientDetail
	Steps       []RecipeStepDetail
}

// GetRecipeByID はレシピの全情報を取得します
func GetRecipeByID(db *sql.DB, id string) (*RecipeFull, error) {
	var r RecipeFull
	err := db.QueryRow("SELECT id, user_id, title, description FROM recipes WHERE id = ?", id).Scan(&r.ID, &r.UserID, &r.Title, &r.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// 材料の取得 (foodsテーブルと結合して名称を取得)
	rows, err := db.Query(`
		SELECT f.food_id, f.name, ri.quantity, ri.group_name 
		FROM recipe_ingredients ri 
		JOIN foods f ON ri.food_id = f.food_id 
		WHERE ri.recipe_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ing RecipeIngredientDetail
		rows.Scan(&ing.FoodID, &ing.Name, &ing.Quantity, &ing.GroupName)
		r.Ingredients = append(r.Ingredients, ing)
	}

	// 工程の取得
	rows, err = db.Query("SELECT step_number, instruction FROM recipe_steps WHERE recipe_id = ? ORDER BY step_number ASC", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var step RecipeStepDetail
		rows.Scan(&step.StepNumber, &step.Instruction)
		r.Steps = append(r.Steps, step)
	}

	return &r, nil
}
