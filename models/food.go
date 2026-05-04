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
