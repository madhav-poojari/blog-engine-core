package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/typewriter-blog/api/internal/content"
)

// check indexes a directory and reports what the API would serve from it.
//
// It touches no git remote, no network and no config, so it is the fastest way
// to see how a real content repo will be interpreted — in particular which
// posts would be public, which files are quarantined, and which still need
// frontmatter.
func check(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: server check <path-to-content-dir>")
	}
	root := args[0]

	builder := &content.Builder{
		Renderer: content.NewRenderer(),
		Options:  content.Options{IgnoredPathTags: map[string]bool{"uncategorised": true}},
	}
	result, err := builder.Build(root)
	if err != nil {
		return err
	}
	idx := result.Index

	fmt.Printf("Indexed %s\n", root)
	fmt.Printf("  %d posts (%d public, %d private)\n",
		idx.Count(true), idx.Count(false), idx.Count(true)-idx.Count(false))
	fmt.Printf("  %d quarantined, %d awaiting frontmatter\n\n", len(idx.Errors), len(result.Writes))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VISIBILITY\tSLUG\tTITLE\tTAGS\tPATH")
	for _, p := range idx.List(true, "") {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n", p.Visibility, p.Slug, p.Title, p.Tags, p.Path)
	}
	w.Flush()

	if len(idx.Errors) > 0 {
		fmt.Printf("\nQuarantined (excluded from the site, never auto-rewritten):\n")
		for _, e := range idx.Errors {
			fmt.Printf("  %s\n    %s\n", e.Path, e.Reason)
		}
	}

	if len(result.Writes) > 0 {
		fmt.Printf("\nFrontmatter the next normalize pass would add:\n")
		for _, pw := range result.Writes {
			fmt.Printf("  %s\n", pw.Path)
		}
		fmt.Printf("\n(run `server check` again after a normalize to confirm this list is empty)\n")
	}

	return nil
}
