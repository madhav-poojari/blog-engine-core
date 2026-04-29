package handler

import (
	"net/http"
	"strings"

	"github.com/typewriter-blog/api/internal/storage"
)

const maxUploadSize = 10 << 20 // 10 MB

var allowedContentTypes = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
}

type ImageHandler struct {
	store *storage.S3Client
}

func NewImageHandler(store *storage.S3Client) *ImageHandler {
	return &ImageHandler{store: store}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 10MB)")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'image' field")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// Validate content type
	baseCT := strings.Split(contentType, ";")[0]
	if !allowedContentTypes[baseCT] {
		writeError(w, http.StatusBadRequest, "unsupported image format")
		return
	}

	result, err := h.store.Upload(r.Context(), file, contentType, header.Filename)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	writeJSON(w, http.StatusCreated, result)
}
