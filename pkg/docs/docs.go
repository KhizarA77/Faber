// Package docs implements documentation-as-the-source-of-truth: it resolves a
// library to its canonical documentation, fetches the relevant pages, and caches
// them so agents can inject authoritative excerpts into their briefs.
//
// v0 (M1) defines only the data types and the Service facade. The concrete
// resolver→fetcher→cache implementation lands in M2.
package docs

import (
	"context"
	"time"
)

// Excerpt is a single relevant passage pulled from a documentation page.
type Excerpt struct {
	Heading string
	Text    string
	Anchor  string
}

// Pack is the authoritative documentation gathered for one library.
type Pack struct {
	Library   string
	Version   string
	URL       string
	Excerpts  []Excerpt
	FetchedAt time.Time
}

// Service is the facade agents depend on. PrefetchAll is what makes the
// docs-first policy real — BuildBrief calls it and drops the result straight
// into Brief.DocPacks.
type Service interface {
	// Consult returns authoritative docs for a specific library/version/query.
	Consult(ctx context.Context, lib, version, query string) (Pack, error)

	// PrefetchAll gathers docs for every library in one call.
	PrefetchAll(ctx context.Context, libs []string) ([]Pack, error)
}
