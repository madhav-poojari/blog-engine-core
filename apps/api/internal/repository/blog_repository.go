package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/typewriter-blog/api/internal/model"
)

type BlogRepository struct {
	db *sqlx.DB
}

func NewBlogRepository(db *sqlx.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

func (r *BlogRepository) List(ctx context.Context, includePrivate bool) ([]model.Blog, error) {
	var blogs []model.Blog
	query := `SELECT * FROM blogs`
	if !includePrivate {
		query += ` WHERE visibility = 'public'`
	}
	query += ` ORDER BY created_at DESC`

	if err := r.db.SelectContext(ctx, &blogs, query); err != nil {
		return nil, fmt.Errorf("list blogs: %w", err)
	}
	if blogs == nil {
		blogs = []model.Blog{}
	}
	return blogs, nil
}

func (r *BlogRepository) GetBySlug(ctx context.Context, slug string) (*model.Blog, error) {
	var blog model.Blog
	err := r.db.GetContext(ctx, &blog, `SELECT * FROM blogs WHERE slug = $1`, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get blog by slug: %w", err)
	}
	return &blog, nil
}

func (r *BlogRepository) GetByID(ctx context.Context, id string) (*model.Blog, error) {
	var blog model.Blog
	err := r.db.GetContext(ctx, &blog, `SELECT * FROM blogs WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get blog by id: %w", err)
	}
	return &blog, nil
}

func (r *BlogRepository) Create(ctx context.Context, blog *model.Blog) error {
	query := `
		INSERT INTO blogs (id, slug, title, excerpt, content, cover_image, tags, visibility, read_time_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		blog.ID, blog.Slug, blog.Title, blog.Excerpt, blog.Content,
		blog.CoverImage, blog.Tags, blog.Visibility, blog.ReadTimeMinutes,
		blog.CreatedAt, blog.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create blog: %w", err)
	}
	return nil
}

func (r *BlogRepository) Update(ctx context.Context, blog *model.Blog) error {
	query := `
		UPDATE blogs
		SET title = $2, slug = $3, excerpt = $4, content = $5, cover_image = $6,
		    tags = $7, visibility = $8, read_time_minutes = $9, updated_at = $10
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query,
		blog.ID, blog.Title, blog.Slug, blog.Excerpt, blog.Content,
		blog.CoverImage, blog.Tags, blog.Visibility, blog.ReadTimeMinutes,
		blog.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update blog: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("blog not found")
	}
	return nil
}

func (r *BlogRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM blogs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete blog: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("blog not found")
	}
	return nil
}

func (r *BlogRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM blogs WHERE slug = $1)`, slug)
	return exists, err
}
