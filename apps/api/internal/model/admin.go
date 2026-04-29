package model

import "time"

type Admin struct {
	ID           string    `db:"id"            json:"id"`
	Username     string    `db:"username"      json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at"    json:"createdAt"`
}
