# Wiki Log

Append-only timeline of wiki maintenance activity.

## [2026-05-12] update | Document .goreleaser.yml file location convention

Added explanation to `09_release_docs.md` for why `.goreleaser.yml` lives at the repo
root rather than `.github/`: GoReleaser is a tool-level config, not a GitHub-specific
feature. The root is its conventional default location.

## [2026-05-12] annotate | Add source links to utility docs

Added `[pkg/<name>/](../pkg/<name>/)` links to every utility header in phase
pages (01, 03, 04, 06, 07). Also linked infrastructure packages in phases 00
and 05. All 55 utility packages now have clickable source links from their
wiki documentation.

## [2026-05-12] ingest | Migrate plan/ directory into wiki/

Moved all 19 plan/*.md phase and reference files into wiki/. Replaced the stub
wiki/phases.md with the full roadmap from plan_updated.md. Updated wiki/index.md
to link all migrated pages. Fixed cross-page wiki/ path references (7 instances)
to use bare wiki-relative links. Removed plan/ directory.
