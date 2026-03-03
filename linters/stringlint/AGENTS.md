# stringlint

golangci-lint module plugin that flags direct string comparisons (`s == ""`, `len(s) == 0`) and recommends `wegostrings.IsEmpty(s)` / `wegostrings.IsNotEmpty(s)` from `github.com/wego/pkg/strings` instead. Suppress false positives with `//nolint:stringlint`.

## Structure

- `analyzer.go` — entry point; filters `*ast.BinaryExpr` nodes for comparison operators
- `patterns.go` — pattern matchers for empty-string and len-based comparisons, diagnostic reporting
- `plugin.go` — golangci-lint module plugin registration
- `cmd/stringlint/` — standalone CLI
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
- `valid.go` asserts zero false positives (negative tests)
- `.go.golden` files verify auto-fix output via `RunWithSuggestedFixes`
- `testdata/` is its own Go module — run `cd testdata && go mod tidy` after changing dependencies

## Key decisions

See [docs/decisions.md](docs/decisions.md) for rationale behind each decision.

- **Only checks `*ast.BinaryExpr`** — targets comparisons, not all string literals
- **Handles pointer dereferences** — `*ptr == ""` suggests `IsEmptyP(ptr)` (P-suffix variants)
- **Import alias `wegostrings`** — avoids conflict with stdlib `strings`; defined as [`pkgAlias`](patterns.go) in `patterns.go`
- **Load mode is `LoadModeTypesInfo`** — needs type info to distinguish `string` from `[]byte`, slices, maps

## Performance

See [docs/decisions.md](docs/decisions.md) for type lookup ordering rationale and anti-patterns.

- **`LoadModeTypesInfo`** — required; no syntax-only way to distinguish `string` from `[]byte`
- **Shared inspector** — `pass.ResultOf[inspect.Analyzer]`
- **Narrow node filter** — `[]ast.Node{(*ast.BinaryExpr)(nil)}` via `inspector.Preorder`
- **Guard order** — operator check, syntactic match, type lookup, then `printer.Fprint` only on violations
