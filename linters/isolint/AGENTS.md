# isolint

golangci-lint module plugin that flags raw ISO code string literals (`"USD"`, `"SG"`) and recommends `currency.USD` / `site.SG` package constants instead. Suppress false positives with `//nolint:isolint`.

## Structure

- `analyzer.go` — entry point; walks `*ast.BasicLit` nodes
- `codes.go` — delegates to `currency.IsISO4217()` and `site.Currency()` for code validation
- `report.go` — diagnostic messages and `SuggestedFix` text edits
- `plugin.go` — golangci-lint module plugin registration
- `cmd/isolint/` — standalone CLI
- `testdata/` — separate Go module with test fixtures and `.golden` files

## Commands

```bash
go test -v ./...    # run all tests
go build ./...      # compile
go vet ./...        # static analysis
```

## Testing

Uses `golang.org/x/tools/go/analysis/analysistest`:
- Files with `// want` annotations assert **expected diagnostics** (positive tests)
- `valid.go` contains code that must **not** trigger diagnostics (negative tests) — any unexpected diagnostic on a line without `// want` fails the test
- `.go.golden` files verify auto-fix output via `RunWithSuggestedFixes`

The `testdata/` directory is its own Go module with `replace` directives pointing to `../../currency` and `../../../iso/site`. Run `cd testdata && go mod tidy` after changing those dependencies.

### Writing good linter tests

`analysistest` enforces a **bidirectional contract**:

1. Every `// want` annotation must match a produced diagnostic — otherwise the test fails (missing expected diagnostic)
2. Every produced diagnostic must match a `// want` annotation — otherwise the test fails (unexpected diagnostic)

Rule 2 is why `valid.go` exists and matters. It contains inputs that resemble flaggable code but should **not** be flagged (e.g. `currency.USD` already using a constant, mixed-case `"Usd"`, non-ISO `"XYZ"`). If the analyzer accidentally matches any of these, the test fails because there is no `// want` to match. This makes `valid.go` an assertion of **zero false positives**.

When adding new rules or modifying detection logic:
- **Add positive cases** in a dedicated file with `// want` annotations and a `.go.golden` file for the suggested fix
- **Add negative cases** in `valid.go` — include edge cases that look similar to flaggable code but should be skipped
- `.go.golden` files are only needed for files that produce fixes; `valid.go` needs no golden file because it produces no diagnostics

## Key decisions

- **Code validation delegates to source packages** — `currency.IsISO4217()` and `site.Currency()` are used directly. No hardcoded maps to maintain. Note: `currency.IsISO4217` is case-insensitive so `codes.go` adds an uppercase guard. `site.Currency()` covers sites with currency mappings only (AQ, AX, etc. without mappings won't be flagged).
- **Skips definition packages** — `pass.Pkg.Path()` is checked against `skipPackages` in `analyzer.go`. Add new entries there if more packages should be excluded.
- **Load mode is `LoadModeSyntax`** — only needs string literal values, not type info.

## Performance

This linter is as fast as a golangci-lint analyzer can be. Two axes determine linter speed: [load mode](https://golangci-lint.run/docs/contributing/architecture/) (how much package data is loaded before analysis) and [AST traversal efficiency](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector) (how many nodes reach your callback). isolint optimizes both.

### What this linter does well

- **`LoadModeSyntax`** — the cheapest load mode. The type-checker never runs for this linter. This matters because golangci-lint takes the [union of all enabled linters' load modes](https://golangci-lint.run/docs/contributing/architecture/); a single linter claiming `LoadModeTypesInfo` forces type-checking for the entire batch. A [real-world regression in depguard](https://github.com/golangci/golangci-lint/issues/2670) was caused by exactly this mistake.
- **Shared inspector via `pass.ResultOf[inspect.Analyzer]`** — the [inspect pass](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inspect) constructs the inspector once per package and caches it. All analyzers that declare `Requires: []*analysis.Analyzer{inspect.Analyzer}` share the same instance. Never call `inspector.New(pass.Files)` inside `run()` — it pays the construction cost twice and loses the sharing benefit.
- **Narrow node filter `[]ast.Node{(*ast.BasicLit)(nil)}`** — [`inspector.Preorder`](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector#Inspector.Preorder) with a typed filter skips entire subtrees via a pre-computed bitmask. Only `*ast.BasicLit` nodes reach the callback. Passing `nil` would visit every node in every file.
- **Cheap guards before expensive operations** — the callback checks `lit.Kind != token.STRING` (integer comparison) and `len(lit.Value) != 4 && != 5` (byte-length check on interned source text) before calling `strconv.Unquote`, which allocates. This eliminates the vast majority of string literals at near-zero cost.
- **Package-path skip at the top of `run()`** — `skipPackages` check on `pass.Pkg.Path()` avoids traversing definition packages entirely, before the inspector is even touched.
- **Allocations only on the reporting path** — `fmt.Sprintf` and `SuggestedFix` slice construction happen only when a violation is confirmed, not in the hot loop.

### Anti-patterns to avoid when extending

- Claiming `LoadModeTypesInfo` when you only need syntax — forces type-checking for all linters in the batch
- Passing `nil` as the `Preorder` node filter — visits every AST node
- Calling `fmt.Sprintf` or `printer.Fprint` on every visited node — allocates on the hot path
- Constructing `inspector.New(pass.Files)` instead of using `pass.ResultOf[inspect.Analyzer]` — pays double construction cost
