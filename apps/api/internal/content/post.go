package content

import "time"

// Visibility values. A post is only ever served anonymously when it is
// explicitly marked public, so private is the safe default everywhere.
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Post is a single content file as the API serves it. Content is rendered to
// HTML once, at index time, so request handling is a map lookup.
type Post struct {
	ID              string    `json:"id"`
	Slug            string    `json:"slug"`
	Title           string    `json:"title"`
	Excerpt         string    `json:"excerpt"`
	ContentHTML     string    `json:"contentHtml"`
	CoverImage      string    `json:"coverImage,omitempty"`
	Tags            []string  `json:"tags"`
	Visibility      string    `json:"visibility"`
	ReadTimeMinutes int       `json:"readTimeMinutes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Path is the file's location in the content repo. Admin-only: it is
	// stripped before a post reaches an anonymous caller.
	Path string `json:"path,omitempty"`
}

// IsPublic reports whether the post may be served without admin auth.
func (p *Post) IsPublic() bool { return p.Visibility == VisibilityPublic }

// PublicCopy returns the post with admin-only fields removed.
func (p *Post) PublicCopy() *Post {
	c := *p
	c.Path = ""
	return &c
}

// FileError records a content file that could not be indexed. These are
// surfaced on the admin sync status endpoint rather than failing the sync:
// one bad file must not take the whole blog down.
type FileError struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}
