// Package gitrepo wraps the git CLI over the content repo working clone.
//
// The clone is disposable cache, owned entirely by the server: nothing is ever
// edited there by hand, and deleting it is always a valid recovery. That is
// what makes `git reset --hard` unconditionally safe here.
package gitrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrPushRejected means the remote moved while we were working. The caller
// should discard its local commit and start over.
var ErrPushRejected = errors.New("push rejected: remote has moved")

// Config describes the clone and how to reach it.
type Config struct {
	URL        string
	Path       string
	Branch     string
	SSHKeyPath string
	BotName    string
	BotEmail   string
}

// Repo is a working clone. It is not safe for concurrent use; the sync worker
// is the only caller and it is single-threaded by construction.
type Repo struct {
	cfg Config
}

func New(cfg Config) *Repo { return &Repo{cfg: cfg} }

// BotEmail is the identity used for frontmatter commits. History attributes
// these to the bot so they do not count as content updates.
func (r *Repo) BotEmail() string { return r.cfg.BotEmail }

// Path is the clone's location on disk.
func (r *Repo) Path() string { return r.cfg.Path }

// CheckGitAvailable verifies the git binary exists, so a missing dependency
// surfaces at boot rather than during the first sweep.
func CheckGitAvailable() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is required but was not found on PATH: %w", err)
	}
	return nil
}

// git runs a git command inside the clone.
func (r *Repo) git(ctx context.Context, args ...string) (string, error) {
	return r.gitIn(ctx, r.cfg.Path, args...)
}

func (r *Repo) gitIn(ctx context.Context, dir string, args ...string) (string, error) {
	// quotepath=false keeps non-ASCII filenames readable instead of escaped.
	full := append([]string{"-c", "core.quotepath=false"}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	if dir != "" {
		cmd.Dir = dir
	}

	env := append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_NOSYSTEM=1",
	)
	if r.cfg.SSHKeyPath != "" {
		env = append(env, "GIT_SSH_COMMAND=ssh -i "+r.cfg.SSHKeyPath+
			" -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new -o BatchMode=yes")
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("git %s: %w: %s",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// EnsureClone creates the clone if it is missing, and makes sure its settings
// match what indexing expects.
func (r *Repo) EnsureClone(ctx context.Context) error {
	if _, err := os.Stat(filepath.Join(r.cfg.Path, ".git")); err == nil {
		return r.configure(ctx)
	}

	if err := os.MkdirAll(filepath.Dir(r.cfg.Path), 0o755); err != nil {
		return fmt.Errorf("create clone parent: %w", err)
	}
	// A full clone, not shallow: history is needed to seed post dates, and
	// --depth plus repeated fetch/reset is the fragile combination.
	if _, err := r.gitIn(ctx, "", "clone", "--branch", r.cfg.Branch, r.cfg.URL, r.cfg.Path); err != nil {
		return err
	}
	return r.configure(ctx)
}

// configure pins the settings that keep write-back from churning. autocrlf
// must stay off: posts are authored on Windows, and a clone that rewrote line
// endings would show every file as modified on every sweep.
func (r *Repo) configure(ctx context.Context) error {
	settings := [][2]string{
		{"core.autocrlf", "false"},
		{"core.safecrlf", "false"},
		{"user.name", r.cfg.BotName},
		{"user.email", r.cfg.BotEmail},
	}
	for _, kv := range settings {
		if kv[1] == "" {
			continue
		}
		if _, err := r.git(ctx, "config", kv[0], kv[1]); err != nil {
			return err
		}
	}
	return nil
}

// RemoteHead returns the SHA at the tip of the configured branch on the
// remote, without fetching objects.
func (r *Repo) RemoteHead(ctx context.Context) (string, error) {
	out, err := r.git(ctx, "ls-remote", "origin", "refs/heads/"+r.cfg.Branch)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return "", fmt.Errorf("branch %q not found on remote", r.cfg.Branch)
	}
	return fields[0], nil
}

// Head returns the local HEAD SHA.
func (r *Repo) Head(ctx context.Context) (string, error) {
	out, err := r.git(ctx, "rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}

// FetchAndReset brings the clone to the remote tip, discarding anything local.
// Reset rather than pull or merge, so there is never a conflict state to get
// wedged in. It returns the new HEAD.
func (r *Repo) FetchAndReset(ctx context.Context) (string, error) {
	if _, err := r.git(ctx, "fetch", "--prune", "origin", r.cfg.Branch); err != nil {
		return "", err
	}
	if _, err := r.git(ctx, "reset", "--hard", "origin/"+r.cfg.Branch); err != nil {
		return "", err
	}
	return r.Head(ctx)
}

// IsDirty reports whether the working tree differs from HEAD. This is the
// primary guard against committing nothing: ask git rather than trusting the
// normalizer's own opinion about whether it changed anything.
func (r *Repo) IsDirty(ctx context.Context) (bool, error) {
	out, err := r.git(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// CommitAll stages everything and commits as the bot identity. It reports
// false when there was nothing to commit.
func (r *Repo) CommitAll(ctx context.Context, message string) (bool, error) {
	dirty, err := r.IsDirty(ctx)
	if err != nil || !dirty {
		return false, err
	}
	if _, err := r.git(ctx, "add", "-A"); err != nil {
		return false, err
	}
	if _, err := r.git(ctx,
		"-c", "user.name="+r.cfg.BotName,
		"-c", "user.email="+r.cfg.BotEmail,
		"commit", "-m", message,
	); err != nil {
		return false, err
	}
	return true, nil
}

// Push sends the local branch. A rejection is reported as ErrPushRejected so
// the caller can discard and retry rather than forcing.
func (r *Repo) Push(ctx context.Context) error {
	_, err := r.git(ctx, "push", "origin", "HEAD:refs/heads/"+r.cfg.Branch)
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "non-fast-forward") ||
		strings.Contains(msg, "fetch first") ||
		strings.Contains(msg, "rejected") {
		return fmt.Errorf("%w: %v", ErrPushRejected, err)
	}
	return err
}

// LastCommitAuthorEmail returns the author email of local HEAD, used on boot
// to recognise a frontmatter commit that was made but never pushed.
func (r *Repo) LastCommitAuthorEmail(ctx context.Context) (string, error) {
	out, err := r.git(ctx, "log", "-1", "--format=%ae")
	return strings.TrimSpace(out), err
}
