package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	websync "github.com/typewriter-blog/api/internal/sync"
)

// maxWebhookBody caps what we will read from a push event.
const maxWebhookBody = 1 << 20

// manualSyncTimeout bounds how long an admin request waits for a cycle.
const manualSyncTimeout = 30 * time.Second

// Syncer is the subset of the sync worker the handlers need.
type Syncer interface {
	Trigger(reason string)
	TriggerAndWait(ctx context.Context, reason string, writeBack bool) (websync.Status, error)
	Status() websync.Status
}

type SyncHandler struct {
	syncer        Syncer
	webhookSecret string
	branch        string
}

func NewSyncHandler(syncer Syncer, webhookSecret, branch string) *SyncHandler {
	return &SyncHandler{syncer: syncer, webhookSecret: webhookSecret, branch: branch}
}

type pushPayload struct {
	Ref string `json:"ref"`
}

// Webhook handles a GitHub push event.
//
// It verifies the signature, then answers immediately and syncs in the
// background: GitHub times out at ten seconds, and a git fetch plus a full
// reindex has no business happening inside a request.
func (h *SyncHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if h.webhookSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "webhook is not configured")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "payload too large")
		return
	}

	if !h.validSignature(r.Header.Get("X-Hub-Signature-256"), body) {
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	switch event := r.Header.Get("X-GitHub-Event"); event {
	case "ping":
		writeJSON(w, http.StatusOK, map[string]string{"status": "pong"})
		return
	case "push", "":
		// handled below
	default:
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "ignored", "event": event})
		return
	}

	// Only the configured branch matters; a push to any other is not content.
	var payload pushPayload
	if err := json.Unmarshal(body, &payload); err == nil && payload.Ref != "" {
		if payload.Ref != "refs/heads/"+h.branch {
			writeJSON(w, http.StatusAccepted, map[string]string{"status": "ignored", "ref": payload.Ref})
			return
		}
	}

	h.syncer.Trigger("webhook")
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

// validSignature checks GitHub's HMAC in constant time.
func (h *SyncHandler) validSignature(header string, body []byte) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	want, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), want)
}

// SyncNow re-reads the content repo and returns the result. Admin only.
func (h *SyncHandler) SyncNow(w http.ResponseWriter, r *http.Request) {
	h.runAndRespond(w, r, "manual", false)
}

// NormalizeNow forces a frontmatter write-back pass instead of waiting for the
// daily sweep. Admin only.
func (h *SyncHandler) NormalizeNow(w http.ResponseWriter, r *http.Request) {
	h.runAndRespond(w, r, "manual-normalize", true)
}

func (h *SyncHandler) runAndRespond(w http.ResponseWriter, r *http.Request, reason string, writeBack bool) {
	ctx, cancel := context.WithTimeout(r.Context(), manualSyncTimeout)
	defer cancel()

	status, err := h.syncer.TriggerAndWait(ctx, reason, writeBack)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error":  "sync did not finish in time; it is still running",
			"status": status,
		})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

// Status reports the last sync. Admin only, since it names quarantined files.
func (h *SyncHandler) Status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.syncer.Status())
}
