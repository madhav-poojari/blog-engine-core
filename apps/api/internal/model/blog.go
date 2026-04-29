package model

import (
	"encoding/json"
	"time"
)

type Blog struct {
	ID              string          `db:"id"               json:"id"`
	Slug            string          `db:"slug"             json:"slug"`
	Title           string          `db:"title"            json:"title"`
	Excerpt         string          `db:"excerpt"          json:"excerpt"`
	Content         json.RawMessage `db:"content"          json:"content"`
	CoverImage      *string         `db:"cover_image"      json:"coverImage,omitempty"`
	Tags            StringSlice     `db:"tags"             json:"tags"`
	Visibility      string          `db:"visibility"       json:"visibility"`
	ReadTimeMinutes int             `db:"read_time_minutes" json:"readTimeMinutes"`
	CreatedAt       time.Time       `db:"created_at"       json:"createdAt"`
	UpdatedAt       time.Time       `db:"updated_at"       json:"updatedAt"`
}

type BlogCreateInput struct {
	Title      string          `json:"title"      validate:"required"`
	Excerpt    string          `json:"excerpt"`
	Content    json.RawMessage `json:"content"    validate:"required"`
	CoverImage *string         `json:"coverImage"`
	Tags       []string        `json:"tags"`
	Visibility string          `json:"visibility" validate:"required,oneof=public private"`
}

type BlogUpdateInput struct {
	Title      *string          `json:"title"`
	Excerpt    *string          `json:"excerpt"`
	Content    *json.RawMessage `json:"content"`
	CoverImage *string          `json:"coverImage"`
	Tags       *[]string        `json:"tags"`
	Visibility *string          `json:"visibility" validate:"omitempty,oneof=public private"`
}
