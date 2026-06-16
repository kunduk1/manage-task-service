// Package model — персистентные модели уровня репозитория для домена user (MySQL).
package model

import "time"

// User — строка таблицы users (модель 3-го уровня, с тегами хранилища).
type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	Name         string    `db:"name"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
