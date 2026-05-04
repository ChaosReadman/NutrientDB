package models

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
}

// Register は新しいユーザーを登録します
func Register(db *sql.DB, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, string(hash))
	return err
}

// Authenticate はユーザーを認証します
func Authenticate(db *sql.DB, username, password string) (*User, error) {
	var u User
	var hash string
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", username).Scan(&u.ID, &u.Username, &hash)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return nil, err // パスワード不一致
	}

	return &u, nil
}
