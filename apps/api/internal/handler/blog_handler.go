package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/typewriter-blog/api/internal/middleware"
	"github.com/typewriter-blog/api/internal/model"
	"github.com/typewriter-blog/api/internal/service"
)

type BlogHandler struct {
	svc *service.BlogService
}

func NewBlogHandler(svc *service.BlogService) *BlogHandler {
	return &BlogHandler{svc: svc}
}

// List returns all blogs. If the request has admin auth, private blogs are included.
func (h *BlogHandler) List(w http.ResponseWriter, r *http.Request) {
	includePrivate := middleware.IsAdmin(r.Context())
	blogs, err := h.svc.List(r.Context(), includePrivate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list blogs")
		return
	}
	writeJSON(w, http.StatusOK, blogs)
}

// GetBySlug returns a single blog post by its slug.
func (h *BlogHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	blog, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get blog")
		return
	}
	if blog == nil {
		writeError(w, http.StatusNotFound, "blog not found")
		return
	}
	// Non-admin users can't view private posts
	if blog.Visibility == "private" && !middleware.IsAdmin(r.Context()) {
		writeError(w, http.StatusNotFound, "blog not found")
		return
	}
	writeJSON(w, http.StatusOK, blog)
}

// Create creates a new blog post. Requires admin auth.
func (h *BlogHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.BlogCreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if input.Visibility != "public" && input.Visibility != "private" {
		writeError(w, http.StatusBadRequest, "visibility must be 'public' or 'private'")
		return
	}

	blog, err := h.svc.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create blog")
		return
	}
	writeJSON(w, http.StatusCreated, blog)
}

// Update updates an existing blog post. Requires admin auth.
func (h *BlogHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input model.BlogUpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Visibility != nil && *input.Visibility != "public" && *input.Visibility != "private" {
		writeError(w, http.StatusBadRequest, "visibility must be 'public' or 'private'")
		return
	}

	blog, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update blog")
		return
	}
	writeJSON(w, http.StatusOK, blog)
}

// Delete deletes a blog post. Requires admin auth.
func (h *BlogHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete blog")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
