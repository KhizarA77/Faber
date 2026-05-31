// Package docs carries Faber's documentation-as-the-source-of-truth policy.
//
// Faber does not fetch documentation. Under pure delegation its only job is to
// make agents ALWAYS ground external-API code in current official docs using the
// host's own web and codebase tools - never in stale training memory. That
// behavioral contract is the Directive below; docs-first agents embed it in
// their briefs.
package docs

// Directive is the standing instruction Faber attaches to docs-first work. It
// encodes the source-of-truth order: codebase first, then official docs, which
// win any conflict. Kept ASCII-only because it travels over the wire as protocol
// payload.
const Directive = "Ground all external API/library usage in the source of truth, in this order: " +
	"(1) FIRST use your codebase search tools (grep/glob) to find existing usage and follow the " +
	"project's established patterns; (2) THEN use your web search/fetch tools to read the official " +
	"documentation for the libraries involved; (3) treat the official docs as the ABSOLUTE source " +
	"of truth. When prior knowledge or assumptions conflict with the docs, the docs win. Never " +
	"rely on memory for an external API you can verify."
