// Package sync owns the content repo: it is the only thing that touches the
// clone, so git state never needs locking beyond this one goroutine.
package sync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/typewriter-blog/api/internal/content"
	"github.com/typewriter-blog/api/internal/gitrepo"
)

const (
	// debounceWindow absorbs a burst of webhooks from a multi-commit push into
	// one cycle. Kept short, because webhook-to-live latency is the point.
	debounceWindow = 2 * time.Second
	// pushAttempts bounds retries when the author pushes while we are pushing.
	pushAttempts = 3
	// writeBackLimitPerHour is a circuit breaker. If frontmatter write-back
	// ever starts looping, this is what stops it filling the history with bot
	// commits; the blog keeps serving either way.
	writeBackLimitPerHour = 5

	commitMessage = "chore(content): normalize frontmatter [skip-normalize]"
)

// Status is a snapshot of the last sync, surfaced to the admin UI.
type Status struct {
	LastSyncAt    time.Time           `json:"lastSyncAt"`
	LastSyncOK    bool                `json:"lastSyncOk"`
	LastError     string              `json:"lastError,omitempty"`
	HeadSHA       string              `json:"headSha"`
	Posts         int                 `json:"posts"`
	PublicPosts   int                 `json:"publicPosts"`
	DurationMS    int64               `json:"durationMs"`
	Quarantined   []content.FileError `json:"quarantined"`
	WriteBackHour int                 `json:"writeBacksThisHour"`
	BreakerOpen   bool                `json:"breakerOpen"`
	Degraded      bool                `json:"degraded"`
}

// Options configure the worker.
type Options struct {
	PollInterval         time.Duration
	DailySweepAt         string // "HH:MM", local time; empty disables
	NormalizeOnEverySync bool
}

type trigger struct {
	reason    string
	writeBack bool
	waiters   []chan Status
}

// Worker keeps the in-memory index in step with the content repo.
type Worker struct {
	repo    *gitrepo.Repo
	builder *content.Builder
	opts    Options

	index    atomic.Pointer[content.Index]
	triggers chan trigger

	mu            sync.Mutex
	status        Status
	lastPushedSHA string
	writeBacks    []time.Time
}

// New builds a worker with an empty index, so the API can serve (an empty
// list) before the first sync completes.
func New(repo *gitrepo.Repo, builder *content.Builder, opts Options) *Worker {
	w := &Worker{
		repo:     repo,
		builder:  builder,
		opts:     opts,
		triggers: make(chan trigger, 1),
	}
	w.index.Store(content.NewEmptyIndex())
	return w
}

// Index returns the current snapshot. Never nil.
func (w *Worker) Index() *content.Index { return w.index.Load() }

// Status returns the last sync result.
func (w *Worker) Status() Status {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := w.status
	s.WriteBackHour = len(w.recentWriteBacks())
	s.BreakerOpen = len(w.recentWriteBacks()) >= writeBackLimitPerHour
	return s
}

// Trigger requests a sync without waiting. If one is already pending the
// request is dropped: pending syncs are interchangeable, which is the whole
// queue-bounding story.
func (w *Worker) Trigger(reason string) {
	select {
	case w.triggers <- trigger{reason: reason}:
	default:
	}
}

// TriggerAndWait runs a sync and returns its result, for the admin endpoints.
func (w *Worker) TriggerAndWait(ctx context.Context, reason string, writeBack bool) (Status, error) {
	done := make(chan Status, 1)
	select {
	case w.triggers <- trigger{reason: reason, writeBack: writeBack, waiters: []chan Status{done}}:
	case <-ctx.Done():
		return w.Status(), ctx.Err()
	}

	select {
	case s := <-done:
		return s, nil
	case <-ctx.Done():
		return w.Status(), ctx.Err()
	}
}

// Run drives the worker until ctx is cancelled. It performs one sync
// immediately so the index is warm, then responds to triggers, the poll
// ticker, and the daily sweep.
func (w *Worker) Run(ctx context.Context) {
	if err := w.recoverLocalState(ctx); err != nil {
		log.Printf("sync: clone not ready: %v", err)
	}
	w.runCycle(ctx, trigger{reason: "startup", writeBack: w.opts.NormalizeOnEverySync})

	poll := newTicker(w.opts.PollInterval)
	defer poll.Stop()

	sweep := time.NewTimer(durationUntil(w.opts.DailySweepAt, time.Now()))
	defer sweep.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case t := <-w.triggers:
			w.runCycle(ctx, w.debounce(ctx, t))

		case <-poll.C:
			if w.remoteMoved(ctx) {
				w.runCycle(ctx, trigger{reason: "poll", writeBack: w.opts.NormalizeOnEverySync})
			}

		case <-sweep.C:
			// The daily sweep is the one scheduled moment that materializes
			// frontmatter into the repo, keeping the hot path read-only.
			w.runCycle(ctx, trigger{reason: "daily-sweep", writeBack: true})
			sweep.Reset(durationUntil(w.opts.DailySweepAt, time.Now()))
		}
	}
}

// debounce waits out the debounce window, folding any further triggers into
// this one cycle. Write-back intent and waiters accumulate, so nobody who
// asked for a result is left hanging.
func (w *Worker) debounce(ctx context.Context, t trigger) trigger {
	timer := time.NewTimer(debounceWindow)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return t
		case extra := <-w.triggers:
			t.writeBack = t.writeBack || extra.writeBack
			t.waiters = append(t.waiters, extra.waiters...)
		case <-timer.C:
			return t
		}
	}
}

func (w *Worker) runCycle(ctx context.Context, t trigger) {
	started := time.Now()
	err := w.cycle(ctx, t.writeBack)

	idx := w.Index()
	w.mu.Lock()
	w.status.LastSyncAt = time.Now().UTC()
	w.status.LastSyncOK = err == nil
	w.status.DurationMS = time.Since(started).Milliseconds()
	w.status.HeadSHA = idx.HeadSHA
	w.status.Posts = idx.Count(true)
	w.status.PublicPosts = idx.Count(false)
	w.status.Quarantined = idx.Errors
	w.status.Degraded = err != nil
	if err != nil {
		w.status.LastError = err.Error()
		log.Printf("sync (%s): %v", t.reason, err)
	} else {
		w.status.LastError = ""
		log.Printf("sync (%s): ok, %d posts (%d public) at %s in %dms",
			t.reason, w.status.Posts, w.status.PublicPosts, shortSHA(idx.HeadSHA), w.status.DurationMS)
	}
	w.mu.Unlock()

	final := w.Status()
	for _, waiter := range t.waiters {
		select {
		case waiter <- final:
		default:
		}
	}
}

// cycle performs one full read, and optionally the write-back pass.
func (w *Worker) cycle(ctx context.Context, writeBack bool) error {
	if err := w.repo.EnsureClone(ctx); err != nil {
		return err
	}

	head, err := w.repo.FetchAndReset(ctx)
	if err != nil {
		return err
	}

	if err := w.rebuild(ctx, head, writeBack); err != nil {
		return err
	}
	return nil
}

// rebuild indexes the tree and, when asked, materializes missing frontmatter.
func (w *Worker) rebuild(ctx context.Context, head string, writeBack bool) error {
	result, err := w.buildAndStore(ctx, head)
	if err != nil {
		return err
	}

	switch {
	case !writeBack || len(result.Writes) == 0:
		return nil
	case w.isSelfPush(head):
		// Our own frontmatter commit came back around; the index is refreshed,
		// and there is nothing further to write.
		return nil
	case w.breakerOpen():
		log.Printf("sync: write-back suppressed, already %d commits in the last hour", writeBackLimitPerHour)
		return nil
	}
	return w.writeBack(ctx, head, result.Writes)
}

// buildAndStore re-reads history, rebuilds the index and swaps it in.
func (w *Worker) buildAndStore(ctx context.Context, head string) (*content.BuildResult, error) {
	history, err := w.repo.LoadHistory(ctx)
	if err != nil {
		return nil, err
	}
	w.builder.History = history

	result, err := w.builder.Build(w.repo.Path())
	if err != nil {
		return nil, err
	}
	result.Index.HeadSHA = head
	w.index.Store(result.Index)
	return result, nil
}

// writeBack applies pending frontmatter, commits and pushes.
//
// If the author pushed while we were working, the local commit is discarded and
// the pass starts over against the new tip. Discarding is always safe: the
// commit holds only derived data, which the same inputs regenerate.
func (w *Worker) writeBack(ctx context.Context, base string, writes []content.PendingWrite) error {
	for attempt := 1; attempt <= pushAttempts; attempt++ {
		if err := applyWrites(w.repo.Path(), writes); err != nil {
			return err
		}

		committed, err := w.repo.CommitAll(ctx, commitMessage)
		if err != nil {
			return err
		}
		if !committed {
			return nil
		}

		// Checking the remote first is cheaper than eating a rejection.
		remote, remoteErr := w.repo.RemoteHead(ctx)
		raced := remoteErr == nil && remote != base

		if !raced {
			pushErr := w.repo.Push(ctx)
			if pushErr == nil {
				sha, _ := w.repo.Head(ctx)
				w.noteWriteBack(sha)
				log.Printf("sync: pushed frontmatter for %d file(s) as %s", len(writes), shortSHA(sha))
				return nil
			}
			if !errors.Is(pushErr, gitrepo.ErrPushRejected) {
				return pushErr
			}
		}

		// Lost the race: drop our commit, take their tip, recompute.
		newHead, err := w.repo.FetchAndReset(ctx)
		if err != nil {
			return err
		}
		result, err := w.buildAndStore(ctx, newHead)
		if err != nil {
			return err
		}
		if len(result.Writes) == 0 {
			return nil
		}
		base, writes = newHead, result.Writes

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff(attempt)):
		}
	}
	return errors.New("gave up pushing frontmatter after repeated races with the remote; the next sync will retry")
}

func applyWrites(root string, writes []content.PendingWrite) error {
	for _, write := range writes {
		target := filepath.Join(root, filepath.FromSlash(write.Path))
		if err := os.WriteFile(target, write.Content, 0o644); err != nil {
			return fmt.Errorf("write frontmatter to %s: %w", write.Path, err)
		}
	}
	return nil
}

// isSelfPush reports whether head is a commit this process just pushed, so a
// webhook for our own frontmatter commit still refreshes the index but does not
// start another write-back.
func (w *Worker) isSelfPush(head string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastPushedSHA != "" && w.lastPushedSHA == head
}

func (w *Worker) breakerOpen() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.recentWriteBacks()) >= writeBackLimitPerHour
}

func (w *Worker) noteWriteBack(sha string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastPushedSHA = sha
	w.writeBacks = append(w.recentWriteBacks(), time.Now())
}

// recentWriteBacks returns write-back times inside the last hour. The caller
// holds the lock.
func (w *Worker) recentWriteBacks() []time.Time {
	cutoff := time.Now().Add(-time.Hour)
	kept := w.writeBacks[:0:0]
	for _, t := range w.writeBacks {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}

// recoverLocalState handles a crash between commit and push. Local git state
// is pure cache, so an unpushed bot commit is discarded rather than resurrected.
func (w *Worker) recoverLocalState(ctx context.Context) error {
	if err := w.repo.EnsureClone(ctx); err != nil {
		return err
	}
	email, err := w.repo.LastCommitAuthorEmail(ctx)
	if err != nil {
		return err
	}
	if !strings.EqualFold(email, w.repo.BotEmail()) {
		return nil
	}
	remote, err := w.repo.RemoteHead(ctx)
	if err != nil {
		return err
	}
	head, err := w.repo.Head(ctx)
	if err != nil {
		return err
	}
	if head != remote {
		log.Printf("sync: discarding unpushed frontmatter commit %s", shortSHA(head))
		_, err = w.repo.FetchAndReset(ctx)
		return err
	}
	return nil
}

func (w *Worker) remoteMoved(ctx context.Context) bool {
	remote, err := w.repo.RemoteHead(ctx)
	if err != nil {
		log.Printf("sync: poll failed: %v", err)
		return false
	}
	return remote != w.Index().HeadSHA
}

func backoff(attempt int) time.Duration {
	base := time.Duration(attempt*attempt) * 2 * time.Second
	return base + time.Duration(rand.Int63n(int64(time.Second)))
}

// newTicker tolerates a non-positive interval by effectively disabling polling.
func newTicker(d time.Duration) *time.Ticker {
	if d <= 0 {
		d = time.Hour
	}
	return time.NewTicker(d)
}

// durationUntil returns how long until the next occurrence of an "HH:MM" local
// time. An unparseable or empty value pushes the sweep a day out.
func durationUntil(hhmm string, now time.Time) time.Duration {
	t, err := time.Parse("15:04", strings.TrimSpace(hhmm))
	if err != nil {
		return 24 * time.Hour
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	if sha == "" {
		return "(none)"
	}
	return sha
}
