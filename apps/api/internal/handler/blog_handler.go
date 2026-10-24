package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/typewriter-blog/api/internal/content"
	"github.com/typewriter-blog/api/internal/middleware"
)

// IndexSource hands out the current content snapshot.
type IndexSource interface {
	Index() *content.Index
}

type BlogHandler struct {
	src IndexSource
}

func NewBlogHandler(src IndexSource) *BlogHandler {
	return &BlogHandler{src: src}
}

// List returns posts newest first, optionally filtered by ?tag=. Anonymous
// callers only ever see the index's public view.
func (h *BlogHandler) List(w http.ResponseWriter, r *http.Request) {
	admin := middleware.IsAdmin(r.Context())
	posts := h.src.Index().List(admin, r.URL.Query().Get("tag"))
	writeJSON(w, http.StatusOK, posts)
}

// GetBySlug returns a single post. A private post is a 404 for anonymous
// callers rather than a 403, so its existence is not disclosed.
func (h *BlogHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	admin := middleware.IsAdmin(r.Context())
	post := h.src.Index().Get(chi.URLParam(r, "slug"), admin)
	if post == nil {
		writeError(w, http.StatusNotFound, "blog not found")
		return
	}
	writeJSON(w, http.StatusOK, post)
}

// Tags returns tag names with post counts, for the listing page's filters.
func (h *BlogHandler) Tags(w http.ResponseWriter, r *http.Request) {
	admin := middleware.IsAdmin(r.Context())
	writeJSON(w, http.StatusOK, h.src.Index().Tags(admin))
}

// WriteNotSupported answers the old create/update/delete routes. Posts are
// written by committing to the content repo; the admin editor returns in a
// later phase as a commit through the GitHub API.
func WriteNotSupported(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusMethodNotAllowed,
		"posts are managed in the content repository; commit a markdown file instead")
}
