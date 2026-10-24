package content

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const wordsPerMinute = 200

// excerptLimit is roughly a card's worth of text.
const excerptLimit = 200

// Renderer converts markdown bodies to HTML. It is safe for concurrent use and
// is built once at startup.
type Renderer struct {
	md goldmark.Markdown
}

// NewRenderer builds the markdown pipeline.
//
// Raw HTML passthrough is deliberately NOT enabled. A post is rendered straight
// into a public page, so turning on html.WithUnsafe would make write access to
// the content repo equivalent to script execution on the site. If raw HTML in
// posts is ever wanted, add a sanitizer rather than flipping that switch.
func NewRenderer() *Renderer {
	return &Renderer{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Footnote,
				extension.Typographer,
				highlighting.NewHighlighting(
					highlighting.WithStyle("github"),
					highlighting.WithFormatOptions(),
				),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
			),
		),
	}
}

// Rendered is the output of turning one body into servable content.
type Rendered struct {
	HTML            string
	Excerpt         string
	ReadTimeMinutes int
}

// Render converts a markdown body. The excerpt falls back to the opening prose
// when the file carries no summary of its own.
func (r *Renderer) Render(body, summary string) (Rendered, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(body), &buf); err != nil {
		return Rendered{}, err
	}

	plain := PlainText(body)
	excerpt := strings.TrimSpace(summary)
	if excerpt == "" {
		excerpt = truncateWords(plain, excerptLimit)
	}

	return Rendered{
		HTML:            buf.String(),
		Excerpt:         excerpt,
		ReadTimeMinutes: readTime(plain),
	}, nil
}

func readTime(plain string) int {
	minutes := len(strings.Fields(plain)) / wordsPerMinute
	if minutes < 1 {
		return 1
	}
	return minutes
}

// PlainText strips the markup that would otherwise show up in an excerpt or
// inflate a word count. It is intentionally approximate: it feeds summaries and
// read-time estimates, not the rendered page.
func PlainText(body string) string {
	var b strings.Builder
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			b.WriteString("\n")
			continue
		case strings.HasPrefix(trimmed, "```"), strings.HasPrefix(trimmed, "~~~"):
			continue
		case strings.Trim(trimmed, "-=*_ ") == "" && len(trimmed) >= 3:
			// A thematic break or a setext underline.
			continue
		}
		trimmed = strings.TrimLeft(trimmed, "#> \t")
		trimmed = strings.TrimPrefix(trimmed, "- ")
		trimmed = strings.TrimPrefix(trimmed, "* ")
		b.WriteString(stripInline(trimmed))
		b.WriteString(" ")
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// stripInline removes emphasis markers, inline code ticks, and the bracket
// scaffolding of links and images, keeping the visible text.
func stripInline(s string) string {
	var b strings.Builder
	var skippingURL bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case skippingURL:
			if c == ')' {
				skippingURL = false
			}
		case c == '(' && i > 0 && s[i-1] == ']':
			skippingURL = true
		case c == '*' || c == '_' || c == '`' || c == '[' || c == ']' || c == '!':
			// drop the marker, keep the text it wraps
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// truncateWords cuts at a word boundary near limit and appends an ellipsis.
func truncateWords(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := strings.LastIndexFunc(s[:limit], unicode.IsSpace)
	if cut <= 0 {
		cut = limit
	}
	return strings.TrimRight(s[:cut], " ,;:.") + "…"
}
