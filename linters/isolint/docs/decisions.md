# isolint — Design Decisions

## Uppercase only

Only uppercase string literals (`"USD"`, `"SG"`) are flagged. Lowercase (`"usd"`, `"sg"`) and mixed case (`"Usd"`, `"Sg"`) are ignored.

Rationale: uppercase in Go source code signals an intentional ISO reference — it matches the canonical form defined by ISO 4217 (currency) and ISO 3166-1 alpha-2 (country/site). Lowercase is almost always an identifier, parameter name, or English word. Flagging it would produce excessive false positives with minimal benefit.

## Code validation delegates to source packages

[`currency.IsISO4217()`](../codes.go) and [`site.Currency()`](../codes.go) are called directly rather than maintaining hardcoded maps inside the linter.

- Single source of truth — if a new currency or site is added to the domain packages, the linter picks it up automatically.
- `site.Currency()` covers only sites with currency mappings (AQ, AX, etc. without mappings won't be flagged). This is intentional — flagging a site code that has no constant to replace it with would be a false positive.

## Skip mechanisms

### Package skip ([`skipPackages`](../analyzer.go))

`pass.Pkg.Path()` is checked against [`skipPackages`](../analyzer.go). This prevents the linter from flagging ISO literals inside the packages that _define_ the constants (e.g. the `currency` or `site` package itself). Add new entries to `skipPackages` in [`analyzer.go`](../analyzer.go) when introducing new definition packages.

### Import path skip

Import paths like `import "io"` are syntactically `*ast.BasicLit` strings. The linter checks the parent stack (via [`inspector.WithStack`](../analyzer.go)) to skip `*ast.ImportSpec` parents.

### Call argument skip ([`skipMethods`](../analyzer.go))

String arguments to ORM, HTTP, and filter methods (e.g. `db.Select("SG")`, `c.Query("TH")`) are skipped. The linter inspects the parent `*ast.CallExpr` and checks the callee name against [`skipMethods`](../analyzer.go). To skip a new method, add its name there.

### Field-name skip ([`Settings.SkipFields`](../analyzer.go))

Uppercase 2- and 3-letter strings sometimes collide with ISO codes by accident — most notably card scheme abbreviations like `"MC"` (MasterCard, also Monaco) and `"AX"` (American Express, also Åland Islands). When a literal is assigned to a struct field whose name appears in [`Settings.SkipFields`](../analyzer.go), the linter skips it.

Two assignment shapes are recognized syntactically — no type information is needed:

- Single-target `*ast.AssignStmt` whose sole `Lhs` element is a `*ast.SelectorExpr` (covers `a.CardSchemes = pq.StringArray{"MC"}`). Tuple assignments (`len(Lhs) > 1`) fall through to the default flag behavior because the analyzer cannot correlate which RHS literal belongs to which LHS target without type information.
- `*ast.KeyValueExpr` whose `Key` is an `*ast.Ident` (covers `Foo{CardSchemes: pq.StringArray{"MC"}}`)

Bare local variables that happen to share a skip-field name (e.g. `CardSchemes := "MC"`) are intentionally NOT skipped — a local variable is not a struct field, and broadening to plain identifiers would silently expand the linter's blind spot beyond the documented design.

The walk in [`isAssignToSkipField`](../analyzer.go) stops at function boundaries (`*ast.FuncLit`/`*ast.FuncDecl`) so a literal beyond a closure boundary is no longer treated as part of the original assignment.

**Why not type-based?** Matching by the destination field's _type_ (e.g. "skip everything assigned to a `pq.StringArray`") would require `pass.TypesInfo`, which forces escalation to `LoadModeTypesInfo`. That has a project-wide cost (see below) and is also coarser — a real `pq.StringArray` of site codes would also be skipped. Field-name targeting is precise and stays within `LoadModeSyntax`.

**Why not a flat value allowlist (e.g. `allow-values: [MC]`)?** A global allowlist suppresses the same literal everywhere, including in genuine country contexts. Field-name targeting suppresses it only where the field semantics warrant it, leaving the linter useful elsewhere.

## `LoadModeSyntax`

This is the cheapest load mode — the type-checker never runs. The linter only needs string literal values, not type information, so `LoadModeSyntax` is sufficient.

Why this matters: golangci-lint takes the [union of all enabled linters' load modes](https://golangci-lint.run/docs/contributing/architecture/). Claiming `LoadModeTypesInfo` unnecessarily forces type-checking for the entire batch. A [real-world regression in depguard](https://github.com/golangci/golangci-lint/issues/2670) was caused by claiming type info when it wasn't needed.

**Do not change to `LoadModeTypesInfo`** unless you add logic that genuinely requires `pass.TypesInfo`.

## Guard ordering

Guards in the callback are ordered by cost (cheapest first):

1. `lit.Kind` — token comparison (integer)
2. `len(lit.Value)` — length check (integer)
3. [`isImportPath`](../analyzer.go) — stack walk (already loaded)
4. [`isArgToSkipMethod`](../analyzer.go) — stack walk + string comparison
5. [`isAssignToSkipField`](../analyzer.go) — stack walk + string comparison; no-ops when `SkipFields` is empty
6. `strconv.Unquote` — first allocation
7. Code validation — delegates to `currency`/`site` packages
8. `fmt.Sprintf` — only on the reporting path

This ordering ensures the hot path (non-string, wrong-length, import-path literals) exits before any allocation occurs.

## Anti-patterns to avoid when extending

- Claiming `LoadModeTypesInfo` when only syntax is needed — forces type-checking for all linters in the batch
- Constructing `inspector.New(pass.Files)` instead of using `pass.ResultOf[inspect.Analyzer]` — pays double construction cost and loses sharing
- Passing `nil` as the node filter to `inspector.WithStack` — visits every AST node
- Calling `fmt.Sprintf` or `strconv.Unquote` on every visited node — allocates on the hot path
