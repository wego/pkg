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

Uses `golang.org/x/tools/go/analysis/analysistest`:
- Files with `// want` annotations assert **expected diagnostics** (positive tests)
- `valid.go` contains code that must **not** trigger diagnostics (negative tests) — any unexpected diagnostic on a line without `// want` fails the test
- `.go.golden` files verify auto-fix output via `RunWithSuggestedFixes`

The `testdata/` directory is its own Go module with a `replace` directive pointing to `../../../strings`. Run `cd testdata && go mod tidy` after changing that dependency.

### Writing good linter tests

`analysistest` enforces a **bidirectional contract**:

1. Every `// want` annotation must match a produced diagnostic — otherwise the test fails (missing expected diagnostic)
2. Every produced diagnostic must match a `// want` annotation — otherwise the test fails (unexpected diagnostic)

Rule 2 is why `valid.go` exists and matters. It contains inputs that resemble flaggable code but should **not** be flagged (e.g. comparisons against non-string types, struct field access, method calls returning strings). If the analyzer accidentally matches any of these, the test fails because there is no `// want` to match. This makes `valid.go` an assertion of **zero false positives**.

When adding new rules or modifying detection logic:
- **Add positive cases** in a dedicated file with `// want` annotations and a `.go.golden` file for the suggested fix
- **Add negative cases** in `valid.go` — include edge cases that look similar to flaggable code but should be skipped
- `.go.golden` files are only needed for files that produce fixes; `valid.go` needs no golden file because it produces no diagnostics

## Key decisions

- **Only checks `*ast.BinaryExpr`** — this linter targets comparisons specifically, not all string literals.
- **Handles pointer dereferences** — `*ptr == ""` suggests `IsEmptyP(ptr)` (P-suffix variants).
- **Import alias `wegostrings`** — avoids conflict with stdlib `strings`. Defined as `pkgAlias` in `patterns.go`.
- **Load mode is `LoadModeTypesInfo`** — needs type information to distinguish `string` from `[]byte`, slices, maps, etc.

## Performance

Two axes determine linter speed: [load mode](https://golangci-lint.run/docs/contributing/architecture/) (how much package data is loaded before analysis) and [AST traversal efficiency](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector) (how many nodes reach your callback). This linter requires `LoadModeTypesInfo` — a genuinely necessary cost — but optimizes everything else within that constraint.

### Why `LoadModeTypesInfo` is required here

Without type information, `len(s) == 0` where `s` is a `[]byte` would be a false positive. The linter calls `pass.TypesInfo.TypeOf(expr)` to confirm operands are `string`-typed before reporting. This is a correct architectural decision — there is no syntax-only way to distinguish `string` from `[]byte`, slices, or interfaces. See the [analysis package docs on Pass.TypesInfo](https://pkg.go.dev/golang.org/x/tools/go/analysis#Pass).

Note: golangci-lint takes the [union of all enabled linters' load modes](https://golangci-lint.run/docs/contributing/architecture/). A linter claiming `LoadModeTypesInfo` forces type-checking for the entire batch. A [real-world regression in depguard](https://github.com/golangci/golangci-lint/issues/2670) was caused by claiming type info unnecessarily. Only declare this mode when you genuinely need `pass.TypesInfo`.

### What this linter does well

- **Shared inspector via `pass.ResultOf[inspect.Analyzer]`** — the [inspect pass](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inspect) constructs the inspector once per package and caches it. All analyzers that declare `Requires: []*analysis.Analyzer{inspect.Analyzer}` share the same instance. Never call `inspector.New(pass.Files)` inside `run()` — it pays the construction cost twice and loses the sharing benefit.
- **Narrow node filter `[]ast.Node{(*ast.BinaryExpr)(nil)}`** — [`inspector.Preorder`](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector#Inspector.Preorder) with a typed filter skips entire subtrees via a pre-computed bitmask. Only `*ast.BinaryExpr` nodes reach the callback. Passing `nil` would visit every node in every file.
- **Early return on non-comparison operators** — the callback checks `op` against the comparison operator set before any type lookups. This is a cheap integer comparison that eliminates arithmetic, assignment, and bitwise binary expressions.
- **Type lookups only on candidates** — `pass.TypesInfo.TypeOf(expr)` (a pointer-keyed map lookup) is called only after syntactic checks pass, not on every `*ast.BinaryExpr`.
- **`printer.Fprint` only on the reporting path** — the `render()` helper in `patterns.go` is the most expensive per-call operation (it walks a sub-AST and pretty-prints it), but it only runs on confirmed violations, not in the hot loop.

### Anti-patterns to avoid when extending

- Claiming `LoadModeTypesInfo` when you only need syntax — forces type-checking for all linters in the batch
- Passing `nil` as the `Preorder` node filter — visits every AST node
- Calling `fmt.Sprintf` or `printer.Fprint` on every visited node — allocates on the hot path
- Constructing `inspector.New(pass.Files)` instead of using `pass.ResultOf[inspect.Analyzer]` — pays double construction cost
- Calling `pass.TypesInfo.TypeOf()` before cheaper syntactic guards — do cheap checks first
