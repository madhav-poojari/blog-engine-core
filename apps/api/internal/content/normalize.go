package content

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Dates are the git-derived timestamps for one file. First seeds `created`;
// LastNonBot feeds `updated`, which is deliberately never written into the
// file — writing it would change the file's last-commit time on every sweep,
// which would change the value again, forever.
type Dates struct {
	First      time.Time
	LastNonBot time.Time
}

// Options configure derivation.
type Options struct {
	// IgnoredPathTags are directory names that should not become tags, so a
	// filing-system folder like "uncategorised" does not show up as a topic.
	IgnoredPathTags map[string]bool
	// Now is injectable for tests; zero means time.Now.
	Now time.Time
}

func (o Options) now() time.Time {
	if !o.Now.IsZero() {
		return o.Now.UTC()
	}
	return time.Now().UTC()
}

// Normalize computes the effective metadata for a file and records any keys
// missing from its frontmatter block, so a later write-back can materialize
// them. It never overwrites a value the author wrote: that is what makes a
// second normalization pass a no-op, and that no-op is what stops the
// write-back/webhook cycle from looping.
func Normalize(doc *Document, relPath string, dates Dates, opts Options) Meta {
	m := doc.Meta

	if m.Title == "" {
		m.Title = deriveTitle(doc.Body, relPath)
	}
	if m.Slug == "" {
		m.Slug = Slugify(m.Title)
	}
	if m.Visibility == "" {
		m.Visibility = VisibilityPrivate
	}
	if m.Created.IsZero() {
		switch {
		case !dates.First.IsZero():
			m.Created = dates.First
		default:
			m.Created = opts.now()
		}
	}

	// git is authoritative for `updated` whenever it has an answer, because the
	// field is never materialized and so cannot drift into the file.
	switch {
	case !dates.LastNonBot.IsZero():
		m.Updated = dates.LastNonBot
	case m.Updated.IsZero():
		m.Updated = m.Created
	}
	if m.Updated.Before(m.Created) {
		m.Updated = m.Created
	}

	m.Tags = mergeTags(PathTags(relPath, opts.IgnoredPathTags), m.Tags)

	// Record what the file is missing, in the order a fresh block should read.
	doc.Ensure("title", yamlScalar(m.Title))
	doc.Ensure("slug", m.Slug)
	doc.Ensure("visibility", m.Visibility)
	doc.Ensure("created", m.Created.Format("2006-01-02"))
	doc.Ensure("tags", yamlFlowList(m.Tags))

	return m
}

// PathTags turns a file's directory path into tags, one per segment, so the
// repo's folder structure becomes navigation without a separate category
// concept.
func PathTags(relPath string, ignored map[string]bool) []string {
	dir := path.Dir(filepathToSlash(relPath))
	if dir == "." || dir == "/" || dir == "" {
		return nil
	}
	var out []string
	for _, segment := range strings.Split(strings.Trim(dir, "/"), "/") {
		tag := NormalizeTag(segment)
		if tag == "" || ignored[tag] {
			continue
		}
		out = append(out, tag)
	}
	return out
}

// mergeTags keeps path tags first, then any extra explicit tags, deduplicated.
// Order is deterministic so the written `tags:` line is stable across runs.
func mergeTags(pathTags, explicit []string) []string {
	seen := make(map[string]bool, len(pathTags)+len(explicit))
	out := make([]string, 0, len(pathTags)+len(explicit))
	for _, group := range [][]string{pathTags, explicit} {
		for _, t := range group {
			t = NormalizeTag(t)
			if t == "" || seen[t] {
				continue
			}
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

var (
	atxHeading     = regexp.MustCompile(`(?m)^#[ \t]+(.+?)[ \t]*#*[ \t]*$`)
	nonAlphanumRun = regexp.MustCompile(`[^a-z0-9]+`)
)

// deriveTitle prefers the body's first top-level heading, falling back to the
// filename. Many existing posts are plain .txt files whose name is already a
// perfectly good title.
func deriveTitle(body, relPath string) string {
	if m := atxHeading.FindStringSubmatch(body); m != nil {
		if t := strings.TrimSpace(m[1]); t != "" {
			return t
		}
	}
	if t := setextTitle(body); t != "" {
		return t
	}
	return titleFromFilename(relPath)
}

// setextTitle finds a heading underlined with ===, which plain-text notes use
// more often than they use a # prefix.
func setextTitle(body string) string {
	lines := strings.Split(body, "\n")
	for i := 0; i+1 < len(lines) && i < 10; i++ {
		text := strings.TrimSpace(lines[i])
		rule := strings.TrimSpace(lines[i+1])
		if text == "" || len(rule) < 2 {
			continue
		}
		if strings.Trim(rule, "=") == "" {
			return text
		}
	}
	return ""
}

func titleFromFilename(relPath string) string {
	name := path.Base(filepathToSlash(relPath))
	name = strings.TrimSuffix(name, path.Ext(name))
	name = strings.NewReplacer("-", " ", "_", " ").Replace(name)
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "Untitled"
	}
	// Only capitalize when the name carries no case information of its own, so
	// a filename like "TD-Deep-work" keeps its shape.
	if name == strings.ToLower(name) {
		r := []rune(name)
		r[0] = unicode.ToUpper(r[0])
		name = string(r)
	}
	return name
}

// Slugify builds a URL-safe slug. Moved here from the old blog service, which
// generated slugs from titles at write time.
func Slugify(s string) string {
	if slug := slugifyOrEmpty(s); slug != "" {
		return slug
	}
	return "untitled"
}

// slugifyOrEmpty is Slugify without the placeholder, so callers that would
// rather drop an unusable value than invent one can tell the difference.
func slugifyOrEmpty(s string) string {
	lower := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
	return strings.Trim(nonAlphanumRun.ReplaceAllString(lower, "-"), "-")
}

// NormalizeTag lowercases and slug-cases a tag so that "Deep Work" from
// frontmatter and a "deep-work" directory collapse to one tag. A value with
// nothing slug-able in it yields the empty string and is dropped.
func NormalizeTag(s string) string {
	return slugifyOrEmpty(s)
}

// filepathToSlash accepts either separator, since paths are produced by a
// filesystem walk on Windows during development and on Linux in production.
func filepathToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

// yamlFlowList renders tags as an inline sequence: [a, b, c].
func yamlFlowList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	return "[" + strings.Join(items, ", ") + "]"
}

// yamlScalar emits a value that YAML will read back unchanged, quoting only
// when a plain scalar would be ambiguous.
func yamlScalar(s string) string {
	if s == "" {
		return `""`
	}
	if isSafePlain(s) {
		return s
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\x%02x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

var yamlReservedWords = map[string]bool{
	"true": true, "false": true, "yes": true, "no": true,
	"on": true, "off": true, "null": true, "~": true,
}

func isSafePlain(s string) bool {
	if s != strings.TrimSpace(s) {
		return false
	}
	if yamlReservedWords[strings.ToLower(s)] {
		return false
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return false
	}
	if strings.ContainsAny(s, "\n\r\t") {
		return false
	}
	// A colon followed by space, or a space followed by #, would start a new
	// mapping or a comment.
	if strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":") {
		return false
	}
	switch s[0] {
	case '-', '?', ':', ',', '[', ']', '{', '}', '#', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
		return false
	}
	return true
}
