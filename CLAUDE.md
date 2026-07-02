# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Blizzaga generates PNG/SVG/WebP images of code and terminal output. It is a fork of
[charmbracelet/freeze](https://github.com/charmbracelet/freeze) whose distinguishing change is the
`srcery` theme, set as the default. Module: `github.com/starbaser/blizzaga` (Go, single `main`
package at the root plus `font/`, `input/`, `svg/`, `render/` subpackages).

## Commands

```sh
make test                 # go test ./...  (builds ./test/blizzaga-test, runs golden comparison, removes binary)
go test ./...             # same as above
go test -run TestCut      # single test
go test -run 'TestBlizzagaConfigurations/shadow'   # single golden subtest (subtest name == the `output` field)
go test -update           # regenerate SVG golden files in test/golden/svg/ from current output
go test -png              # also emit PNGs to test/output/png/ (requires rsvg-convert or falls back to resvg)
make golden               # cp -r test/output/* test/golden  (alternative golden refresh)

go run . file.go -o out.svg        # run locally
golangci-lint run                  # lint (config .golangci.yml; note: tests are excluded from linting)

nix run '.#'              # run via flake
nix build                # builds default.nix (buildGoModule)
```

**Nix `vendorHash` gotcha**: `default.nix` pins `vendorHash`. Any change to `go.mod`/`go.sum`
invalidates it — the Nix build will fail with a hash mismatch until you update the `vendorHash`
value to the one Nix reports.

**PNG dependency**: PNG output prefers the external `rsvg-convert` (librsvg) binary if present on
`PATH`, otherwise falls back to the embedded `resvg-go` (WASM) renderer in `png.go`. `rsvg-convert`
is not guaranteed to be installed in the dev environment.

## Rendering pipeline (the core architecture)

The whole program is `main.go`'s `main()`. Understanding it requires seeing that **SVG is produced
by chroma, then mutated as an XML DOM** (via `github.com/beevik/etree`) before being written or
rasterized. The `svg/` package is just DOM-manipulation helpers (`Move`, `SetDimensions`,
`AddShadow`, `AddClipPath`, `AddCornerRadius`, `NewWindowControls`, `AddOutline`).

```
input (file | stdin pipe | --execute via PTY)
        │
        ▼
  isAnsi?  ── strippedInput != input, or --language ansi
   │                                  │
   │ code path                        │ ANSI path
   ▼                                  ▼
chroma lexer.Tokenise          chroma Literator (stripped text,
   → SVG formatter               text-only, just to SIZE the SVG)
        │                                  │
        └────────────┬─────────────────────┘
                     ▼
        etree parse SVG → mutate DOM in main.go:
          compute width/height from font size & line height,
          apply margin/padding, reposition every <text> line,
          add window controls / shadow / border / clip path / line numbers
                     │
        ANSI path only: ansi.NewParser drives the `dispatcher`
        (ansi.go), which walks SGR escape codes and INJECTS
        <tspan> (fg/style) and <rect> (bg) elements into the DOM
                     │
                     ▼
        .png → libsvgConvert (rsvg-convert) | resvgConvert (resvg-go)
        else → write SVG to file / stdout
```

Key non-obvious points:

- **Two-pass kong parsing for config layering** (`main.go`): kong parses argv once to discover
  `--config`, opens the JSON config (embedded `base`/`full`, `user` → `$XDG_CONFIG/blizzaga/user.json`,
  or an arbitrary path), then re-parses argv with that JSON wired in as a `kong.Resolvers` source so
  **CLI flags override config-file values**.
- **ANSI rendering is bespoke** (`ansi.go`): chroma only lays out empty, correctly-sized text lines
  for ANSI input; the `dispatcher` then re-parses the raw bytes and emits colored `<tspan>`s and
  background `<rect>`s. SGR codes 30–37/90–97 map to the srcery palette (`ansiPalette`); 256-color
  and truecolor (`38;5`, `38;2`, `48;…`) are handled via `palette`.
- **PNG auto-scale**: when both height and width are auto and output ends in `.png`, everything is
  rendered at `scale = 4` for resolution, then margins/padding/font math multiply through.
- **Font metrics come from the font asset**: `render.FontMetrics`/`ColAdvance` parse the embedded
  (or `--font.file`) font with `sfnt` — advance, line pitch, and cell aspect are design-table
  facts, never calibrated constants. Terminal width in auto mode is derived from `lipgloss.Width`
  of the longest line times the per-column advance.
- Fonts are **base64-embedded directly into the SVG** (`font.go`, `font/font.go`): Iosevka Custom
  (terminal-oriented subsets of the iosevka-eigenmage build; ligature + `NL` variants, 0.5em
  advance / 1.25em line / 2.5 cell aspect) is embedded by default; `--font.file` embeds a
  TTF/WOFF/WOFF2 instead.

## Themes (the fork's purpose)

`style.go` defines srcery: the `srcery*` color constants, the `srcery` chroma style (registered via
`styles.Register`), the `charm` style (upstream default, still available), and
`defaultStyle = srceryStyle`. The srcery color constants are also consumed by `ansi.go`'s
`ansiPalette`. Any theme registered with chroma is selectable via `--theme`; unknown themes fall
back to `defaultStyle`.

## Testing model

`blizzaga_test.go` is golden-file based. `TestMain` builds the binary to `./test/blizzaga-test`,
then `TestBlizzagaConfigurations` runs the CLI for each case (input + flags), writes SVG to
`test/output/svg/`, and diffs against `test/golden/svg/` (newline-normalized). The subtest name is
the case's `output` field. When intentionally changing rendering, regenerate with `go test -update`
and review the SVG diff. Golden PNGs are Git-LFS tracked (see `.gitattributes`); `*.svg`/`*.png` are
gitignored except under `test/golden/**` and `examples/`.

## Conventions

- Library/presentation split is loosely followed: `svg/`, `input/`, `font/` are pure helpers;
  `main.go` and `error.go` own stdout/exit. Fatal errors go through `printErrorFatal` (styled
  lipgloss "ERROR" header + `os.Exit(1)`), never bare panics.
- Config struct field tags drive *both* the CLI (kong `help`/`short`/`group`/`placeholder`) and the
  JSON config schema. Add a new option by adding a field with both tag families; `--flag` names map
  directly to JSON keys.
- `golangci.yml` enables `gosec`, `revive`, `bodyclose`, etc. with `tests: false`. Existing
  `//nolint:` directives are deliberate (e.g. `wrapcheck` on thin pass-through error returns).
