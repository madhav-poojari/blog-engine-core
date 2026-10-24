package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// maxFileSize guards against a large file committed into the content repo. A
// post is prose; anything much bigger is a mistake.
const maxFileSize = 256 << 10

// indexedExtensions are the files treated as posts. Everything else, including
// extensionless files, is ignored rather than guessed at.
var indexedExtensions = map[string]bool{
	".md":       true,
	".markdown": true,
	".txt":      true,
}

// History supplies git-derived dates for a path.
type History interface {
	Dates(relPath string) Dates
}

// noHistory is used when git history is unavailable, so indexing still works
// against a plain directory (and in tests).
type noHistory struct{}

func (noHistory) Dates(string) Dates { return Dates{} }

// TagCount is one tag and how many posts carry it.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// Index is an immutable snapshot of the content repo. It is built off to the
// side and swapped in whole, so readers never observe a half-built index.
//
// Visibility is resolved once, here, by keeping two separate maps. A handler
// holding the public map cannot leak a private post even if it forgets to
// filter, which is the one mistake in this system that would actually matter.
type Index struct {
	public    map[string]*Post
	all       map[string]*Post
	publicAsc []*Post
	allAsc    []*Post
	publicTag []TagCount
	allTag    []TagCount

	Errors  []FileError
	HeadSHA string
	BuiltAt time.Time
}

// NewEmptyIndex returns an index with no posts, used before the first
// successful sync so the API serves an empty list instead of failing.
func NewEmptyIndex() *Index {
	return &Index{
		public:  map[string]*Post{},
		all:     map[string]*Post{},
		BuiltAt: time.Now().UTC(),
	}
}

// Get returns a post by slug, or nil. Anonymous callers can only ever reach
// the public map.
func (i *Index) Get(slug string, admin bool) *Post {
	if admin {
		return i.all[slug]
	}
	if p := i.public[slug]; p != nil {
		return p.PublicCopy()
	}
	return nil
}

// List returns posts newest first, optionally filtered to a single tag.
func (i *Index) List(admin bool, tag string) []*Post {
	source := i.publicAsc
	if admin {
		source = i.allAsc
	}

	tag = NormalizeTag(tag)
	out := make([]*Post, 0, len(source))
	for _, p := range source {
		if tag != "" && !hasTag(p.Tags, tag) {
			continue
		}
		if admin {
			out = append(out, p)
		} else {
			out = append(out, p.PublicCopy())
		}
	}
	return out
}

// Tags returns tag counts, newest-first irrelevant: ordered by count then name.
func (i *Index) Tags(admin bool) []TagCount {
	if admin {
		return i.allTag
	}
	return i.publicTag
}

// Count reports how many posts are indexed.
func (i *Index) Count(admin bool) int {
	if admin {
		return len(i.all)
	}
	return len(i.public)
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

// PendingWrite is a file whose frontmatter is incomplete, along with the
// contents that would complete it. The sync worker applies these only during a
// write-back pass, never on an ordinary read sync.
type PendingWrite struct {
	Path    string
	Content []byte
}

// BuildResult carries the new index plus the write-back work it implies.
type BuildResult struct {
	Index  *Index
	Writes []PendingWrite
}

// Builder walks a content directory and produces an Index.
type Builder struct {
	Renderer *Renderer
	Options  Options
	History  History
}

// Build indexes every content file under root.
//
// A file that cannot be indexed is quarantined and reported, not fatal: one
// unparseable file must not take down the whole blog.
func (b *Builder) Build(root string) (*BuildResult, error) {
	renderer := b.Renderer
	if renderer == nil {
		renderer = NewRenderer()
	}
	history := b.History
	if history == nil {
		history = noHistory{}
	}

	paths, err := collectPaths(root)
	if err != nil {
		return nil, err
	}

	idx := NewEmptyIndex()
	result := &BuildResult{Index: idx}
	// Paths are sorted, so a slug collision always resolves the same way
	// instead of flapping between builds.
	sort.Strings(paths)

	for _, rel := range paths {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			idx.addError(rel, err.Error())
			continue
		}
		if len(raw) > maxFileSize {
			idx.addError(rel, fmt.Sprintf("file is %d bytes, over the %d byte limit", len(raw), maxFileSize))
			continue
		}
		if !utf8.Valid(raw) {
			idx.addError(rel, "not valid UTF-8")
			continue
		}

		doc, err := Parse(raw)
		if err != nil {
			idx.addError(rel, err.Error())
			continue
		}

		meta := Normalize(doc, rel, history.Dates(rel), b.Options)

		if existing, clash := idx.all[meta.Slug]; clash {
			idx.addError(rel, fmt.Sprintf("slug %q already used by %s", meta.Slug, existing.Path))
			continue
		}

		rendered, err := renderer.Render(doc.Body, meta.Summary)
		if err != nil {
			idx.addError(rel, "render: "+err.Error())
			continue
		}

		post := &Post{
			ID:              meta.Slug,
			Slug:            meta.Slug,
			Title:           meta.Title,
			Excerpt:         rendered.Excerpt,
			ContentHTML:     rendered.HTML,
			CoverImage:      meta.Cover,
			Tags:            meta.Tags,
			Visibility:      meta.Visibility,
			ReadTimeMinutes: rendered.ReadTimeMinutes,
			CreatedAt:       meta.Created,
			UpdatedAt:       meta.Updated,
			Path:            rel,
		}
		idx.add(post)

		if doc.Changed() {
			result.Writes = append(result.Writes, PendingWrite{Path: rel, Content: doc.Bytes()})
		}
	}

	idx.finalize()
	return result, nil
}

func (i *Index) add(p *Post) {
	i.all[p.Slug] = p
	i.allAsc = append(i.allAsc, p)
	if p.IsPublic() {
		i.public[p.Slug] = p
		i.publicAsc = append(i.publicAsc, p)
	}
}

func (i *Index) addError(path, reason string) {
	i.Errors = append(i.Errors, FileError{Path: path, Reason: reason})
}

func (i *Index) finalize() {
	sortPosts(i.allAsc)
	sortPosts(i.publicAsc)
	i.allTag = countTags(i.allAsc)
	i.publicTag = countTags(i.publicAsc)
	i.BuiltAt = time.Now().UTC()
	sort.Slice(i.Errors, func(a, b int) bool { return i.Errors[a].Path < i.Errors[b].Path })
}

// sortPosts orders newest first, with the slug as a stable tie-break so two
// posts committed in the same minute do not swap places between builds.
func sortPosts(posts []*Post) {
	sort.Slice(posts, func(a, b int) bool {
		if !posts[a].CreatedAt.Equal(posts[b].CreatedAt) {
			return posts[a].CreatedAt.After(posts[b].CreatedAt)
		}
		return posts[a].Slug < posts[b].Slug
	})
}

func countTags(posts []*Post) []TagCount {
	counts := map[string]int{}
	for _, p := range posts {
		for _, t := range p.Tags {
			counts[t]++
		}
	}
	out := make([]TagCount, 0, len(counts))
	for tag, n := range counts {
		out = append(out, TagCount{Tag: tag, Count: n})
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Count != out[b].Count {
			return out[a].Count > out[b].Count
		}
		return out[a].Tag < out[b].Tag
	})
	return out
}

// collectPaths returns indexable files as slash-separated paths relative to
// root.
func collectPaths(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			// Skip the git directory and any dotted or conventional metadata
			// folder, so repo plumbing never shows up as a post or a tag.
			if p != root && (strings.HasPrefix(name, ".") || name == "node_modules") {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}
		if !indexedExtensions[strings.ToLower(filepath.Ext(name))] {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}
