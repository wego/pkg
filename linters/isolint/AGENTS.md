# isolint

golangci-lint module plugin that flags uppercase ISO code string literals (`"USD"`, `"SG"`) and recommends `currency.USD` / `site.SG` package constants instead. Lowercase strings (`"usd"`, `"sg"`) are ignored — only uppercase is considered an intentional ISO reference. Suppress remaining false positives with `//nolint:isolint`.

## Structure

- `analyzer.go` — entry point; walks `*ast.BasicLit` nodes with `inspector.WithStack` for parent context
- `codes.go` — delegates to `currency.IsISO4217()` and `site.Currency()` for code validation (uppercase only)
- `report.go` — diagnostic messages and `SuggestedFix` text edits
- `plugin.go` — golangci-lint module plugin registration
- `cmd/isolint/` — standalone CLI
- `testdata/` — separate Go module with test fixtures and `.golden` files

## Commands

```bash
go test -v ./...    # run all tests
go build ./...      # compile
go vet ./...        # static analysis
revive ./...                            # lint (catches unused params, etc.)
golangci-lint run --enable-only revive  # alternative: revive via golangci-lint
```

## Testing

Uses `golang.org/x/tools/go/analysis/analysistest`. See [docs/testing.md](docs/testing.md) for the full testing guide.

- Files with `// want` annotations assert expected diagnostics (positive tests)
- `valid.go` / `valid_contexts.go` assert zero false positives (negative tests)
- `.go.golden` files verify auto-fix output via `RunWithSuggestedFixes`
- `testdata/` is its own Go module — run `cd testdata && go mod tidy` after changing dependencies

## Key decisions

See [docs/decisions.md](docs/decisions.md) for rationale behind each decision.

- **Uppercase only** — lowercase and mixed case are ignored
- **Code validation delegates to source packages** — `currency.IsISO4217()` and `site.Currency()`, no hardcoded maps
- **Skips definition packages** — [`skipPackages`](analyzer.go) in `analyzer.go`
- **Skips import paths and call arguments** — [`skipMethods`](analyzer.go) in `analyzer.go`
- **Load mode is `LoadModeSyntax`** — only needs string literal values, not type info

## Performance

See [docs/decisions.md](docs/decisions.md) for guard ordering rationale and anti-patterns.

- **`LoadModeSyntax`** — cheapest load mode; type-checker never runs
- **Shared inspector** — `pass.ResultOf[inspect.Analyzer]`
- **Narrow node filter** — `[]ast.Node{(*ast.BasicLit)(nil)}` with `inspector.WithStack`
- **Guard order** — cheapest checks first; allocations (`strconv.Unquote`, `fmt.Sprintf`) only on the reporting path
