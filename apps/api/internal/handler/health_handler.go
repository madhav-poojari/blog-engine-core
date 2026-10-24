package handler

import (
	"net/http"

	"github.com/typewriter-blog/api/internal/content"
)

// HealthSource exposes just enough of the sync worker to report readiness.
type HealthSource interface {
	Index() *content.Index
}

type HealthHandler struct {
	src HealthSource
}

func NewHealthHandler(src HealthSource) *HealthHandler {
	return &HealthHandler{src: src}
}

// Check reports the index state so a monitor can alert on a stale or empty
// index, which is how a broken content sync shows up from the outside.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	idx := h.src.Index()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"posts":       idx.Count(true),
		"publicPosts": idx.Count(false),
		"headSha":     idx.HeadSHA,
		"indexedAt":   idx.BuiltAt,
		"quarantined": len(idx.Errors),
	})
}
