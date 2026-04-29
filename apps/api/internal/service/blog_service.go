package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/typewriter-blog/api/internal/model"
	"github.com/typewriter-blog/api/internal/repository"
)

type BlogService struct {
	repo *repository.BlogRepository
}

func NewBlogService(repo *repository.BlogRepository) *BlogService {
	return &BlogService{repo: repo}
}

func (s *BlogService) List(ctx context.Context, includePrivate bool) ([]model.Blog, error) {
	return s.repo.List(ctx, includePrivate)
}

func (s *BlogService) GetBySlug(ctx context.Context, slug string) (*model.Blog, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *BlogService) Create(ctx context.Context, input model.BlogCreateInput) (*model.Blog, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.Visibility == "" {
		input.Visibility = "public"
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}

	slug, err := s.generateUniqueSlug(ctx, input.Title)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	blog := &model.Blog{
		ID:              uuid.New().String(),
		Slug:            slug,
		Title:           input.Title,
		Excerpt:         input.Excerpt,
		Content:         input.Content,
		CoverImage:      input.CoverImage,
		Tags:            model.StringSlice(input.Tags),
		Visibility:      input.Visibility,
		ReadTimeMinutes: estimateReadTime(input.Content),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, blog); err != nil {
		return nil, err
	}
	return blog, nil
}

func (s *BlogService) Update(ctx context.Context, id string, input model.BlogUpdateInput) (*model.Blog, error) {
	blog, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if blog == nil {
		return nil, fmt.Errorf("blog not found")
	}

	if input.Title != nil {
		blog.Title = *input.Title
		slug, err := s.generateUniqueSlug(ctx, *input.Title)
		if err != nil {
			return nil, err
		}
		blog.Slug = slug
	}
	if input.Excerpt != nil {
		blog.Excerpt = *input.Excerpt
	}
	if input.Content != nil {
		blog.Content = *input.Content
		blog.ReadTimeMinutes = estimateReadTime(*input.Content)
	}
	if input.CoverImage != nil {
		blog.CoverImage = input.CoverImage
	}
	if input.Tags != nil {
		blog.Tags = model.StringSlice(*input.Tags)
	}
	if input.Visibility != nil {
		blog.Visibility = *input.Visibility
	}

	blog.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, blog); err != nil {
		return nil, err
	}
	return blog, nil
}

func (s *BlogService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// generateUniqueSlug creates a URL-friendly slug from a title, appending a
// numeric suffix if the slug already exists in the database.
func (s *BlogService) generateUniqueSlug(ctx context.Context, title string) (string, error) {
	base := slugify(title)
	slug := base

	for i := 1; ; i++ {
		exists, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return "", fmt.Errorf("check slug: %w", err)
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	lower := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
	slug := nonAlphanumeric.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "untitled"
	}
	return slug
}

// estimateReadTime calculates approximate read time from JSON content.
// Average reading speed: ~200 words per minute.
func estimateReadTime(content json.RawMessage) int {
	text := extractText(content)
	words := len(strings.Fields(text))
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

func extractText(data json.RawMessage) string {
	var node map[string]json.RawMessage
	if err := json.Unmarshal(data, &node); err != nil {
		return string(data)
	}

	var sb strings.Builder

	if text, ok := node["text"]; ok {
		var t string
		if json.Unmarshal(text, &t) == nil {
			sb.WriteString(t)
			sb.WriteString(" ")
		}
	}

	if content, ok := node["content"]; ok {
		var children []json.RawMessage
		if json.Unmarshal(content, &children) == nil {
			for _, child := range children {
				sb.WriteString(extractText(child))
			}
		}
	}

	return sb.String()
}
