package content

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrMalformed marks a file whose frontmatter block is present but unusable.
// Such a file is quarantined rather than rewritten: its author's intent is
// ambiguous, and guessing at it would destroy information.
var ErrMalformed = errors.New("malformed frontmatter")

const delimiter = "---"

// utf8BOM is the byte-order mark some Windows editors prepend.
const utf8BOM = "\xef\xbb\xbf"

// Meta is the metadata a content file carries, however it was obtained: read
// from the frontmatter block, or derived by the normalizer.
type Meta struct {
	Title      string
	Slug       string
	Visibility string
	Created    time.Time
	Updated    time.Time
	Tags       []string
	Summary    string
	Cover      string
}

// Document is a parsed content file.
//
// The frontmatter block is held as raw text rather than as a decoded structure,
// because write-back must not disturb bytes the author wrote. Re-serializing a
// decoded block would reorder keys and rewrite quoting and date formats, and
// every such change is a diff, and every diff is another commit — which is the
// loop this whole design is built to avoid. Missing keys are appended instead.
type Document struct {
	// HasBlock reports whether the file arrived with a frontmatter block.
	HasBlock bool
	// Meta holds the values actually present in the block.
	Meta Meta
	// Body is the markdown following the block, with LF line endings.
	Body string

	rawBlock string          // block contents, no delimiters, no trailing newline
	present  map[string]bool // canonical keys found in the block
	appended []string        // lines added by Ensure, in call order
}

// aliases maps accepted spellings onto canonical keys.
var aliases = map[string]string{
	"last-updated": "updated",
	"last_updated": "updated",
	"lastupdated":  "updated",
	"description":  "summary",
	"cover-image":  "cover",
	"coverimage":   "cover",
}

func canonicalKey(k string) string {
	k = strings.ToLower(strings.TrimSpace(k))
	if c, ok := aliases[k]; ok {
		return c
	}
	return k
}

// Parse reads a content file. A file with no frontmatter block is not an error:
// it parses as a Document with HasBlock false and an empty Meta, which the
// normalizer then fills in.
func Parse(raw []byte) (*Document, error) {
	text := normalizeNewlines(raw)

	doc := &Document{present: map[string]bool{}}

	rest, block, found, err := splitBlock(text)
	if err != nil {
		return nil, err
	}
	if !found {
		doc.Body = strings.TrimLeft(text, "\n")
		return doc, nil
	}

	doc.HasBlock = true
	doc.rawBlock = block
	doc.Body = strings.TrimLeft(rest, "\n")

	if strings.TrimSpace(block) != "" {
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(block), &node); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
		}
		if err := doc.decode(&node); err != nil {
			return nil, err
		}
	}
	return doc, nil
}

// splitBlock separates a leading delimited block from the body. A file that
// opens a block without closing it is malformed; anything else that does not
// start with a delimiter simply has no block.
func splitBlock(text string) (body, block string, found bool, err error) {
	if !strings.HasPrefix(text, delimiter) {
		return "", "", false, nil
	}
	// The opening delimiter must be alone on its line, so a body that starts
	// with a thematic break or a run of dashes is not mistaken for a block.
	first, remainder, ok := strings.Cut(text, "\n")
	if !ok || strings.TrimRight(first, " \t") != delimiter {
		return "", "", false, nil
	}

	for rem := remainder; ; {
		line, next, more := strings.Cut(rem, "\n")
		trimmed := strings.TrimRight(line, " \t")
		if trimmed == delimiter || trimmed == "..." {
			consumed := len(remainder) - len(rem)
			return next, strings.TrimSuffix(remainder[:consumed], "\n"), true, nil
		}
		if !more {
			return "", "", false, fmt.Errorf("%w: opening %q has no closing delimiter", ErrMalformed, delimiter)
		}
		rem = next
	}
}

func (d *Document) decode(node *yaml.Node) error {
	root := node
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return nil
		}
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: expected a key/value block, got %s", ErrMalformed, kindName(root.Kind))
	}

	for i := 0; i+1 < len(root.Content); i += 2 {
		key := canonicalKey(root.Content[i].Value)
		val := root.Content[i+1]
		d.present[key] = true

		switch key {
		case "title":
			d.Meta.Title = strings.TrimSpace(val.Value)
		case "slug":
			d.Meta.Slug = strings.TrimSpace(val.Value)
		case "summary":
			d.Meta.Summary = strings.TrimSpace(val.Value)
		case "cover":
			d.Meta.Cover = strings.TrimSpace(val.Value)
		case "visibility":
			d.Meta.Visibility = strings.ToLower(strings.TrimSpace(val.Value))
		case "created":
			t, err := parseDate(val.Value)
			if err != nil {
				return fmt.Errorf("%w: created: %v", ErrMalformed, err)
			}
			d.Meta.Created = t
		case "updated":
			t, err := parseDate(val.Value)
			if err != nil {
				return fmt.Errorf("%w: updated: %v", ErrMalformed, err)
			}
			d.Meta.Updated = t
		case "tags":
			tags, err := decodeTags(val)
			if err != nil {
				return err
			}
			d.Meta.Tags = tags
		}
	}

	if v := d.Meta.Visibility; v != "" && v != VisibilityPublic && v != VisibilityPrivate {
		return fmt.Errorf("%w: visibility %q must be %q or %q", ErrMalformed, v, VisibilityPublic, VisibilityPrivate)
	}
	return nil
}

func decodeTags(val *yaml.Node) ([]string, error) {
	switch val.Kind {
	case yaml.SequenceNode:
		out := make([]string, 0, len(val.Content))
		for _, item := range val.Content {
			if t := NormalizeTag(item.Value); t != "" {
				out = append(out, t)
			}
		}
		return out, nil
	case yaml.ScalarNode:
		// Accept a comma-separated scalar as well as a real sequence.
		var out []string
		for _, part := range strings.Split(val.Value, ",") {
			if t := NormalizeTag(part); t != "" {
				out = append(out, t)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%w: tags must be a list", ErrMalformed)
	}
}

var dateLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"2006/01/02",
	"02-01-2006",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date %q", s)
}

// Has reports whether the frontmatter block already carries a key.
func (d *Document) Has(key string) bool { return d.present[canonicalKey(key)] }

// Ensure records a key to be written into the frontmatter block, unless the
// block already has it. Calling it is always safe: a key the author supplied is
// never overwritten, which is what keeps repeated normalization a no-op.
func (d *Document) Ensure(key, value string) {
	key = canonicalKey(key)
	if d.present[key] || value == "" {
		return
	}
	d.present[key] = true
	d.appended = append(d.appended, key+": "+value)
}

// Changed reports whether rendering this document would alter the file.
func (d *Document) Changed() bool { return len(d.appended) > 0 || !d.HasBlock }

// Bytes renders the document back to file contents with LF endings. Existing
// block lines are emitted verbatim; added keys follow them.
func (d *Document) Bytes() []byte {
	if !d.HasBlock && len(d.appended) == 0 {
		return []byte(d.Body)
	}

	var b strings.Builder
	b.WriteString(delimiter)
	b.WriteString("\n")
	if d.rawBlock != "" {
		b.WriteString(d.rawBlock)
		b.WriteString("\n")
	}
	for _, line := range d.appended {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(delimiter)
	b.WriteString("\n")
	if d.Body != "" {
		b.WriteString("\n")
		b.WriteString(d.Body)
	}
	return []byte(b.String())
}

// normalizeNewlines strips a UTF-8 BOM and converts CRLF to LF. Posts are
// authored on Windows, so without this every sweep would rewrite every file
// and commit the churn.
func normalizeNewlines(raw []byte) string {
	s := strings.TrimPrefix(string(raw), utf8BOM)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func kindName(k yaml.Kind) string {
	switch k {
	case yaml.SequenceNode:
		return "a list"
	case yaml.ScalarNode:
		return "a single value"
	case yaml.AliasNode:
		return "an alias"
	default:
		return "something unexpected"
	}
}
