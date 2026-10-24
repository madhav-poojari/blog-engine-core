package content

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)

func testOptions() Options {
	return Options{
		IgnoredPathTags: map[string]bool{"uncategorised": true},
		Now:             fixedNow,
	}
}

// writeTree materializes a map of relative paths to contents.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func build(t *testing.T, root string) *BuildResult {
	t.Helper()
	b := &Builder{Options: testOptions()}
	res, err := b.Build(root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return res
}

// applyWrites performs the write-back the sweep would perform.
func applyWrites(t *testing.T, root string, writes []PendingWrite) {
	t.Helper()
	for _, w := range writes {
		full := filepath.Join(root, filepath.FromSlash(w.Path))
		if err := os.WriteFile(full, w.Content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestNormalizeIsAFixpoint is the load-bearing test for the whole design: the
// server writes frontmatter back to the repo, which triggers a webhook, which
// re-reads the repo. If a second pass produced further changes, that cycle
// would commit forever.
func TestNormalizeIsAFixpoint(t *testing.T) {
	root := writeTree(t, map[string]string{
		"personal/books/atomic-habits.md":     "# Atomic Habits\n\nThe lesson is small margins.\n",
		"personal/fitness/badminton.txt":      "Footwork beats power.\n",
		"uncategorised/Learning.txt":          "Spaced repetition works.\n",
		"career/software/TD-Deep-work.txt":    "Blocks of four hours.\n",
		"career/teaching/with-frontmatter.md": "---\ntitle: Already Titled\nvisibility: public\ncreated: 2024-01-02\nslug: already-titled\ntags: [teaching]\n---\n\nBody text.\n",
	})

	first := build(t, root)
	if len(first.Writes) == 0 {
		t.Fatal("expected the first pass to propose frontmatter for files that lack it")
	}
	applyWrites(t, root, first.Writes)

	second := build(t, root)
	if len(second.Writes) != 0 {
		for _, w := range second.Writes {
			t.Errorf("second pass still wants to rewrite %s:\n%s", w.Path, w.Content)
		}
		t.Fatalf("normalize is not a fixpoint: %d files would be rewritten again", len(second.Writes))
	}

	// And a third pass, to catch anything that alternates between two states.
	applyWrites(t, root, second.Writes)
	if third := build(t, root); len(third.Writes) != 0 {
		t.Fatalf("third pass wants %d rewrites", len(third.Writes))
	}
}

func TestExistingFrontmatterBytesArePreserved(t *testing.T) {
	// Deliberately unusual formatting: odd key order, extra spacing, a comment,
	// and a block sequence. None of it should be reformatted.
	original := strings.Join([]string{
		"---",
		"# notes to self",
		"visibility: public",
		"title: Already Titled",
		"tags:",
		"  - teaching",
		"  - Deep Work",
		"created: 2024-01-02",
		"---",
		"",
		"Body text.",
		"",
	}, "\n")

	root := writeTree(t, map[string]string{"career/teaching/post.md": original})
	res := build(t, root)

	if len(res.Writes) != 1 {
		t.Fatalf("expected exactly one write (the missing slug), got %d", len(res.Writes))
	}
	got := string(res.Writes[0].Content)

	for _, line := range []string{"# notes to self", "  - Deep Work", "visibility: public"} {
		if !strings.Contains(got, line) {
			t.Errorf("rewrite dropped or reformatted %q\nfull output:\n%s", line, got)
		}
	}
	if !strings.Contains(got, "slug: already-titled") {
		t.Errorf("expected the missing slug to be appended, got:\n%s", got)
	}
	// The author's key order must survive: title still comes after visibility.
	if strings.Index(got, "visibility: public") > strings.Index(got, "title: Already Titled") {
		t.Error("key order was changed")
	}
}

func TestMissingFrontmatterDefaultsToPrivate(t *testing.T) {
	root := writeTree(t, map[string]string{"personal/journals/raw.txt": "No metadata here.\n"})
	res := build(t, root)

	posts := res.Index.List(true, "")
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Visibility != VisibilityPrivate {
		t.Errorf("visibility = %q, want %q", posts[0].Visibility, VisibilityPrivate)
	}
	if n := res.Index.Count(false); n != 0 {
		t.Errorf("a post with no frontmatter must not be public, but %d are", n)
	}
}

func TestPathSegmentsBecomeTags(t *testing.T) {
	root := writeTree(t, map[string]string{
		"personal/fitness/badminton/footwork.md": "# Footwork\n\nSplit step.\n",
		"uncategorised/loose.txt":                "Loose note.\n",
	})
	idx := build(t, root).Index

	post := idx.Get("footwork", true)
	if post == nil {
		t.Fatal("footwork not indexed")
	}
	want := []string{"personal", "fitness", "badminton"}
	if strings.Join(post.Tags, ",") != strings.Join(want, ",") {
		t.Errorf("tags = %v, want %v", post.Tags, want)
	}

	loose := idx.Get("loose", true)
	if loose == nil {
		t.Fatal("loose not indexed")
	}
	if len(loose.Tags) != 0 {
		t.Errorf("uncategorised should not become a tag, got %v", loose.Tags)
	}
}

func TestExplicitTagsMergeWithPathTags(t *testing.T) {
	root := writeTree(t, map[string]string{
		"personal/books/x.md": "---\ntitle: X\ntags: [Deep Work, books]\n---\n\nBody.\n",
	})
	post := build(t, root).Index.Get("x", true)
	if post == nil {
		t.Fatal("not indexed")
	}
	// Path tags first, explicit extras after, with the duplicate collapsed.
	want := "personal,books,deep-work"
	if got := strings.Join(post.Tags, ","); got != want {
		t.Errorf("tags = %q, want %q", got, want)
	}
}

func TestMalformedFrontmatterIsQuarantined(t *testing.T) {
	root := writeTree(t, map[string]string{
		"a/unclosed.md":    "---\ntitle: Nope\n\nBody with no closing delimiter.\n",
		"a/badvis.md":      "---\ntitle: Bad\nvisibility: sortof\n---\n\nBody.\n",
		"a/notamapping.md": "---\n- just\n- a list\n---\n\nBody.\n",
		"a/good.md":        "# Good\n\nFine.\n",
	})
	res := build(t, root)

	if len(res.Index.Errors) != 3 {
		t.Fatalf("expected 3 quarantined files, got %d: %+v", len(res.Index.Errors), res.Index.Errors)
	}
	if res.Index.Count(true) != 1 {
		t.Errorf("the good file should still be indexed; count = %d", res.Index.Count(true))
	}
	// A quarantined file must never be rewritten: its intent is ambiguous.
	for _, w := range res.Writes {
		if strings.HasPrefix(w.Path, "a/unclosed") || strings.HasPrefix(w.Path, "a/badvis") {
			t.Errorf("quarantined file %s was queued for rewrite", w.Path)
		}
	}
}

func TestSlugCollisionResolvesDeterministically(t *testing.T) {
	files := map[string]string{
		"a/dupe.md": "---\ntitle: Same Title\nslug: same-title\n---\n\nFirst.\n",
		"b/dupe.md": "---\ntitle: Same Title\nslug: same-title\n---\n\nSecond.\n",
	}
	for i := 0; i < 3; i++ {
		res := build(t, writeTree(t, files))
		if res.Index.Count(true) != 1 {
			t.Fatalf("expected 1 indexed post, got %d", res.Index.Count(true))
		}
		post := res.Index.Get("same-title", true)
		if post == nil || post.Path != "a/dupe.md" {
			t.Fatalf("lexicographically first path should win, got %+v", post)
		}
		if len(res.Index.Errors) != 1 || res.Index.Errors[0].Path != "b/dupe.md" {
			t.Fatalf("the losing file should be reported, got %+v", res.Index.Errors)
		}
	}
}

func TestCRLFAndBOMDoNotProduceChurn(t *testing.T) {
	body := "---\r\ntitle: Windows\r\nslug: windows\r\nvisibility: public\r\ncreated: 2024-03-04\r\ntags: [a]\r\n---\r\n\r\nBody line.\r\n"
	root := writeTree(t, map[string]string{"a/win.md": "\xef\xbb\xbf" + body})

	res := build(t, root)
	if len(res.Writes) != 0 {
		t.Fatalf("a complete CRLF/BOM file should need no rewrite, got:\n%s", res.Writes[0].Content)
	}
	post := res.Index.Get("windows", true)
	if post == nil {
		t.Fatal("not indexed")
	}
	if strings.Contains(post.ContentHTML, "\r") {
		t.Error("rendered HTML still carries CR characters")
	}
}

func TestTitleDerivation(t *testing.T) {
	cases := []struct{ path, body, want string }{
		{"a/heading.md", "# From Heading\n\nText.\n", "From Heading"},
		{"a/setext.md", "From Setext\n===========\n\nText.\n", "From Setext"},
		{"a/How to Make a Habit Irresistible.txt", "Text only.\n", "How to Make a Habit Irresistible"},
		{"a/imposter-syndrome.txt", "Text only.\n", "Imposter syndrome"},
		{"a/TD-Deep-work.txt", "Text only.\n", "TD Deep work"},
	}
	for _, tc := range cases {
		root := writeTree(t, map[string]string{tc.path: tc.body})
		posts := build(t, root).Index.List(true, "")
		if len(posts) != 1 {
			t.Fatalf("%s: expected 1 post, got %d", tc.path, len(posts))
		}
		if posts[0].Title != tc.want {
			t.Errorf("%s: title = %q, want %q", tc.path, posts[0].Title, tc.want)
		}
	}
}

func TestPublicAndAdminViewsAreSeparate(t *testing.T) {
	root := writeTree(t, map[string]string{
		"a/pub.md":  "---\ntitle: Public One\nslug: pub\nvisibility: public\ncreated: 2024-01-01\ntags: [a]\n---\n\nBody.\n",
		"a/priv.md": "---\ntitle: Private One\nslug: priv\nvisibility: private\ncreated: 2024-01-02\ntags: [a]\n---\n\nBody.\n",
	})
	idx := build(t, root).Index

	if idx.Count(false) != 1 || idx.Count(true) != 2 {
		t.Fatalf("counts: public=%d all=%d, want 1 and 2", idx.Count(false), idx.Count(true))
	}
	if idx.Get("priv", false) != nil {
		t.Error("private post reachable anonymously")
	}
	if idx.Get("priv", true) == nil {
		t.Error("private post not reachable by admin")
	}
	// Tag counts must not leak the private post either.
	for _, tc := range idx.Tags(false) {
		if tc.Tag == "a" && tc.Count != 1 {
			t.Errorf("public tag count for %q = %d, want 1", tc.Tag, tc.Count)
		}
	}
	// The repo path is admin-only.
	if p := idx.Get("pub", false); p == nil || p.Path != "" {
		t.Errorf("anonymous post should not expose its repo path, got %+v", p)
	}
	if p := idx.Get("pub", true); p == nil || p.Path == "" {
		t.Error("admin post should expose its repo path")
	}
}

func TestTagFilter(t *testing.T) {
	root := writeTree(t, map[string]string{
		"personal/books/a.md":   "---\ntitle: A\nslug: a\nvisibility: public\ncreated: 2024-01-01\n---\n\nBody.\n",
		"personal/fitness/b.md": "---\ntitle: B\nslug: b\nvisibility: public\ncreated: 2024-01-02\n---\n\nBody.\n",
	})
	idx := build(t, root).Index

	if got := idx.List(false, "books"); len(got) != 1 || got[0].Slug != "a" {
		t.Errorf("tag filter books returned %d posts", len(got))
	}
	if got := idx.List(false, "Personal"); len(got) != 2 {
		t.Errorf("tag filter should be case-insensitive, got %d posts", len(got))
	}
	if got := idx.List(false, "nope"); len(got) != 0 {
		t.Errorf("unknown tag returned %d posts", len(got))
	}
}

func TestListIsNewestFirst(t *testing.T) {
	root := writeTree(t, map[string]string{
		"a/old.md": "---\ntitle: Old\nslug: old\nvisibility: public\ncreated: 2020-01-01\n---\n\nBody.\n",
		"a/new.md": "---\ntitle: New\nslug: new\nvisibility: public\ncreated: 2026-01-01\n---\n\nBody.\n",
	})
	posts := build(t, root).Index.List(false, "")
	if len(posts) != 2 || posts[0].Slug != "new" {
		t.Errorf("expected newest first, got %v", []string{posts[0].Slug, posts[1].Slug})
	}
}

func TestOnlyKnownExtensionsAreIndexed(t *testing.T) {
	root := writeTree(t, map[string]string{
		"a/post.md":    "# Post\n\nBody.\n",
		"a/notes.txt":  "Notes.\n",
		"a/hi":         "extensionless, should be ignored\n",
		"a/data.json":  `{"not":"a post"}`,
		"a/.hidden.md": "# Hidden\n\nBody.\n",
	})
	idx := build(t, root).Index
	if idx.Count(true) != 2 {
		paths := []string{}
		for _, p := range idx.List(true, "") {
			paths = append(paths, p.Path)
		}
		t.Errorf("expected 2 indexed files, got %d: %v", idx.Count(true), paths)
	}
}

func TestDatesComeFromHistoryWhenFrontmatterIsSilent(t *testing.T) {
	created := time.Date(2023, 5, 1, 9, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 7, 9, 18, 30, 0, 0, time.UTC)

	root := writeTree(t, map[string]string{"a/post.md": "# Post\n\nBody.\n"})
	b := &Builder{Options: testOptions(), History: stubHistory{"a/post.md": Dates{First: created, LastNonBot: updated}}}
	res, err := b.Build(root)
	if err != nil {
		t.Fatal(err)
	}

	post := res.Index.Get("post", true)
	if !post.CreatedAt.Equal(created) {
		t.Errorf("createdAt = %v, want %v", post.CreatedAt, created)
	}
	if !post.UpdatedAt.Equal(updated) {
		t.Errorf("updatedAt = %v, want %v", post.UpdatedAt, updated)
	}
	// created is materialized into the file; updated must never be, or the
	// write-back commit would change it and loop.
	written := string(res.Writes[0].Content)
	if !strings.Contains(written, "created: 2023-05-01") {
		t.Errorf("created not written back:\n%s", written)
	}
	if strings.Contains(written, "updated:") {
		t.Errorf("updated must not be written into the file:\n%s", written)
	}
}

func TestFrontmatterCreatedWinsOverHistory(t *testing.T) {
	root := writeTree(t, map[string]string{
		"a/post.md": "---\ntitle: Post\nslug: post\ncreated: 2019-02-03\n---\n\nBody.\n",
	})
	b := &Builder{
		Options: testOptions(),
		History: stubHistory{"a/post.md": Dates{First: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}},
	}
	res, err := b.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	post := res.Index.Get("post", true)
	if post.CreatedAt.Year() != 2019 {
		t.Errorf("frontmatter created should win, got %v", post.CreatedAt)
	}
}

func TestUpdatedAliasesAreAccepted(t *testing.T) {
	doc, err := Parse([]byte("---\ntitle: T\nlast-updated: 2024-06-01\n---\n\nBody.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Meta.Updated.IsZero() {
		t.Error("last-updated was not read as updated")
	}
	if !doc.Has("updated") {
		t.Error("alias should mark the canonical key as present")
	}
}

func TestParseRejectsUnclosedBlock(t *testing.T) {
	_, err := Parse([]byte("---\ntitle: T\n\nbody\n"))
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("err = %v, want ErrMalformed", err)
	}
}

func TestBodyStartingWithThematicBreakIsNotFrontmatter(t *testing.T) {
	doc, err := Parse([]byte("--- a dash line, not a block\n\nBody.\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.HasBlock {
		t.Error("a dash line with trailing text should not parse as frontmatter")
	}
}

func TestReadTimeAndExcerpt(t *testing.T) {
	body := "# Title\n\n" + strings.Repeat("word ", 450)
	r := NewRenderer()
	out, err := r.Render(body, "")
	if err != nil {
		t.Fatal(err)
	}
	if out.ReadTimeMinutes != 2 {
		t.Errorf("readTime = %d, want 2", out.ReadTimeMinutes)
	}
	if strings.Contains(out.Excerpt, "#") {
		t.Errorf("excerpt should not contain markdown markup: %q", out.Excerpt)
	}

	withSummary, err := r.Render(body, "A hand-written summary.")
	if err != nil {
		t.Fatal(err)
	}
	if withSummary.Excerpt != "A hand-written summary." {
		t.Errorf("summary should win over a generated excerpt, got %q", withSummary.Excerpt)
	}
}

func TestRawHTMLIsNotPassedThrough(t *testing.T) {
	out, err := NewRenderer().Render("Hello <script>alert(1)</script> there.\n", "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.HTML, "<script>") {
		t.Errorf("raw HTML must not reach the page: %s", out.HTML)
	}
}

type stubHistory map[string]Dates

func (s stubHistory) Dates(rel string) Dates { return s[rel] }
