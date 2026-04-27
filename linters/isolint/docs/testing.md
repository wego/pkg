# isolint — Testing Guide

## Framework

Tests use [`golang.org/x/tools/go/analysis/analysistest`](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest), the standard testing harness for Go analysis passes.

## The bidirectional contract

`analysistest` enforces a **bidirectional contract** between test fixtures and analyzer output:

1. **Every `// want` annotation must match a produced diagnostic** — otherwise the test fails ("missing expected diagnostic").
2. **Every produced diagnostic must match a `// want` annotation** — otherwise the test fails ("unexpected diagnostic").

Rule 2 is the crucial one. It means any `.go` file in `testdata/` that does _not_ have a `// want` comment is implicitly asserting **zero diagnostics**. If the analyzer accidentally flags something in that file, the test fails.

## Test file roles

| File                                                                              | Role                                                                                          |
| --------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| [`testdata/example/currency_literal.go`](../testdata/example/currency_literal.go) | Positive test — currency literals with `// want` annotations                                  |
| [`testdata/example/site_literal.go`](../testdata/example/site_literal.go)         | Positive test — site literals with `// want` annotations                                      |
| [`testdata/example/valid.go`](../testdata/example/valid.go)                       | Negative test — code that resembles flaggable patterns but must NOT be flagged                |
| [`testdata/example/valid_contexts.go`](../testdata/example/valid_contexts.go)     | Negative test — call-expression contexts (ORM, HTTP, filter methods) that must NOT be flagged |
| `testdata/example/*.go.golden`                                                    | Expected auto-fix output for positive test files                                              |
| [`testdata/skipfields/skip_fields.go`](../testdata/skipfields/skip_fields.go)     | Negative test — assignments to a configured skip field must NOT be flagged                    |
| [`testdata/skipfields/still_flags.go`](../testdata/skipfields/still_flags.go)     | Positive test — same package, same value, but assigned to a different field is still flagged  |

The `./skipfields` package runs under a dedicated `TestAnalyzerWithSkipFields` driver that constructs an analyzer with `Settings{SkipFields: ["CardSchemes"]}` via [`isolint.NewAnalyzer`](../analyzer.go). Because each `analystest.Run` call constructs its own analyzer, configured fixtures are isolated from `./example`'s zero-config assertions.

### Why `valid.go` and `valid_contexts.go` matter

These files contain inputs that look like they could be flagged but should not be:

- `currency.USD` — already using a constant
- `"usd"` — lowercase (not an ISO reference)
- `"Usd"` — mixed case
- `"XYZ"` — not a valid ISO code
- `db.Select("SG")` — call argument to a skipped method

If the analyzer accidentally matches any of these, the test fails because there is no `// want` to absorb the diagnostic. This makes these files **assertions of zero false positives**.

## Golden files

`.go.golden` files verify auto-fix output via [`analysistest.RunWithSuggestedFixes`](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest#RunWithSuggestedFixes). Each golden file corresponds to a positive test file and shows what the file should look like after all suggested fixes are applied.

Golden files are only needed for files that produce fixes. `valid.go` and `valid_contexts.go` need no golden file because they produce no diagnostics.

## How to add tests

### Adding a positive test case (new detection)

1. Create a new `.go` file in `testdata/example/` (or add to an existing positive test file).
2. Add `// want` annotations on lines that should produce diagnostics:
   ```go
   var x = "USD" // want `"USD" is a known ISO 4217 currency code`
   ```
3. Create a matching `.go.golden` file showing the expected fix:
   ```go
   var x = currency.USD
   ```
4. Run `go test ./...` — the test should pass with the new annotations.

### Adding a negative test case (false positive guard)

1. Add the case to `valid.go` (general) or `valid_contexts.go` (call contexts).
2. Do **not** add any `// want` annotation.
3. Run `go test ./...` — if the analyzer incorrectly flags your case, the test will fail with "unexpected diagnostic".

## The `testdata/` module

The `testdata/` directory is its own Go module with `replace` directives pointing to `../../currency` and `../../../iso/site`. After changing those dependencies, run:

```bash
cd testdata && go mod tidy
```
