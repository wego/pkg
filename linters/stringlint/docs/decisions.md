# stringlint — Design Decisions

## `BinaryExpr`-only targeting

The linter filters exclusively for [`*ast.BinaryExpr`](../analyzer.go) nodes. It targets comparisons (`==`, `!=`) against empty strings and `len()` checks, not all string usage.

Rationale: the goal is to replace direct string-emptiness checks with `wegostrings.IsEmpty` / `wegostrings.IsNotEmpty`. These patterns always appear as binary expressions. Checking other node types would add complexity without catching additional violations.

## `LoadModeTypesInfo`

This linter requires [`LoadModeTypesInfo`](https://pkg.go.dev/golang.org/x/tools/go/analysis#Pass) — the type-checker must run.

### Why it's necessary

Without type information, `len(s) == 0` where `s` is a `[]byte` would be a false positive. The linter calls [`pass.TypesInfo.TypeOf(expr)`](../patterns.go) to confirm operands are `string`-typed before reporting. There is no syntax-only way to distinguish `string` from `[]byte`, slices, or interfaces.

### Why it matters

golangci-lint takes the [union of all enabled linters' load modes](https://golangci-lint.run/docs/contributing/architecture/). A linter claiming `LoadModeTypesInfo` forces type-checking for the entire batch. A [real-world regression in depguard](https://github.com/golangci/golangci-lint/issues/2670) was caused by claiming type info unnecessarily. This linter's claim is justified — the alternative is rampant false positives.

## Pointer dereference handling

`*ptr == ""` suggests `wegostrings.IsEmptyP(ptr)` (P-suffix variants). The linter recognizes `*ast.StarExpr` as a unary dereference on the comparison operand and maps it to the pointer-accepting helpers.

This avoids forcing callers to write `wegostrings.IsEmpty(*ptr)` which would panic on nil pointers, while `IsEmptyP` handles nil safely.

## Import alias `wegostrings`

The suggested import uses alias `wegostrings` for `github.com/wego/pkg/strings` to avoid conflict with the stdlib `strings` package. Defined as [`pkgAlias`](../patterns.go) in [`patterns.go`](../patterns.go).

## Type lookup ordering

Guards in the callback are ordered by cost (cheapest first):

1. **Operator check** — `op` against the comparison operator set (integer comparison). Eliminates arithmetic, assignment, and bitwise binary expressions immediately.
2. **Syntactic pattern match** — checks whether the expression matches `x == ""`, `x != ""`, `len(x) == 0`, or `len(x) != 0` structurally.
3. **Type lookup** — [`pass.TypesInfo.TypeOf(expr)`](../patterns.go) (pointer-keyed map lookup) confirms the operand is `string`-typed. Only called after syntactic checks pass.
4. **`printer.Fprint`** — the [`render()`](../patterns.go) helper walks a sub-AST and pretty-prints it. Only runs on confirmed violations, not in the hot loop.

This ordering ensures the common case (non-comparison `BinaryExpr`) exits before any type lookup or allocation.

## Anti-patterns to avoid when extending

- Claiming `LoadModeTypesInfo` when you only need syntax — forces type-checking for all linters in the batch
- Passing `nil` as the `Preorder` node filter — visits every AST node
- Calling `fmt.Sprintf` or `printer.Fprint` on every visited node — allocates on the hot path
- Constructing `inspector.New(pass.Files)` instead of using `pass.ResultOf[inspect.Analyzer]` — pays double construction cost and loses sharing
- Calling `pass.TypesInfo.TypeOf()` before cheaper syntactic guards — do cheap checks first
