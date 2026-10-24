package gitrepo

import (
	"context"
	"strings"
	"time"

	"github.com/typewriter-blog/api/internal/content"
)

// commitMarker starts a commit record in the log output. It is a control
// character so it cannot collide with a path.
const commitMarker = "\x01"

// History holds first and last touch times for every path in the repo, built
// from a single git log traversal rather than one spawn per file.
type History struct {
	dates map[string]content.Dates
}

// Dates implements content.History.
func (h *History) Dates(relPath string) content.Dates {
	if h == nil {
		return content.Dates{}
	}
	return h.dates[relPath]
}

// LoadHistory walks the whole commit history once and records, per path, the
// first time it was touched and the last time it was touched by someone other
// than the bot.
//
// Bot commits are excluded from `updated` deliberately: the bot's only job is
// writing frontmatter, so counting its commits would make every post's update
// time read as the date of the last normalize sweep.
//
// Rename detection is intentionally not used. It cannot be batched across a
// whole tree, and it stops mattering once `created` is materialized into each
// file: from then on the file's own value wins and git is never consulted for
// it again.
func (r *Repo) LoadHistory(ctx context.Context) (*History, error) {
	// Author date, not committer date: a rebase rewrites the latter.
	out, err := r.git(ctx, "log", "--reverse", "--date-order",
		"--format="+commitMarker+"%aI\t%ae", "--name-only")
	if err != nil {
		return nil, err
	}

	h := &History{dates: map[string]content.Dates{}}
	var (
		when  time.Time
		isBot bool
		ok    bool
	)

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")

		if rec, found := strings.CutPrefix(line, commitMarker); found {
			when, isBot, ok = parseCommitHeader(rec, r.cfg.BotEmail)
			continue
		}
		if !ok || line == "" {
			continue
		}

		path := line
		d := h.dates[path]
		if d.First.IsZero() {
			d.First = when
		}
		if !isBot {
			d.LastNonBot = when
		}
		h.dates[path] = d
	}

	return h, nil
}

func parseCommitHeader(rec, botEmail string) (when time.Time, isBot, ok bool) {
	dateStr, email, _ := strings.Cut(rec, "\t")
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(dateStr))
	if err != nil {
		return time.Time{}, false, false
	}
	email = strings.TrimSpace(email)
	return t.UTC(), botEmail != "" && strings.EqualFold(email, botEmail), true
}
