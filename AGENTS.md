# blizzaga — Agent Guide

A fork of charmbracelet/freeze that generates images of code and terminal output, with the srcery theme set as the default.

@../eigengo/CONVENTIONS.md

## Cross-Project Documentation — first-party stack

This project is part of Kyle's first-party stack (`~/dev/projects/*`). Sibling projects' curated
documentation does NOT auto-load into your context — you must go read it.

- Before working with or against a first-party dependency's API or behavior, read that project's
  `~/dev/projects/<dep>/AGENTS.md` and `~/dev/projects/<dep>/docs/` directly.
- First-party dependency: `tree-sitter-baml` supplies the statically linked BAML/FML grammars and
  the canonical BAML highlight query. The flake pins its concrete source; `go.mod` records its module
  identity. Blizzaga owns only renderer-specific query overrides and injection routing.
- Answer how/why questions from those curated docs first; sibling source is for verifying or
  extending what the docs say — and it is first-party, editable at the source when work here
  surfaces a problem there.
