# stringlint — Testing Guide

## Framework

Tests use [`golang.org/x/tools/go/analysis/analysistest`](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest), the standard testing harness for Go analysis passes.

## The bidirectional contract

`analysistest` enforces a **bidirectional contract** between test fixtures and analyzer output:

1. **Every `// want` annotation must match a produced diagnostic** — otherwise the test fails ("missing expected diagnostic").
2. **Every produced diagnostic must match a `// want` annotation** — otherwise the test fails ("unexpected diagnostic").

Rule 2 is the crucial one. It means any `.go` file in `testdata/` that does _not_ have a `// want` comment is implicitly asserting **zero diagnostics**. If the analyzer accidentally flags something in that file, the test fails.

## Test file roles

| File                                                                      | Role                                                                                 |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| [`testdata/example/empty_string.go`](../testdata/example/empty_string.go) | Positive test — `s == ""` / `s != ""` patterns with `// want` annotations            |
| [`testdata/example/len_check.go`](../testdata/example/len_check.go)       | Positive test — `len(s) == 0` / `len(s) != 0` patterns with `// want` annotations    |
| [`testdata/example/pointer.go`](../testdata/example/pointer.go)           | Positive test — `*ptr == ""` pointer dereference patterns with `// want` annotations |
| [`testdata/example/valid.go`](../testdata/example/valid.go)               | Negative test — code that resembles flaggable patterns but must NOT be flagged       |
| `testdata/example/*.go.golden`                                            | Expected auto-fix output for positive test files                                     |

### Why `valid.go` matters

`valid.go` contains inputs that look like they could be flagged but should not be:

- Comparisons against non-string types (`[]byte`, `int`, etc.)
- Struct field access patterns
- Method calls returning strings
- Non-comparison binary expressions

If the analyzer accidentally matches any of these, the test fails because there is no `// want` to absorb the diagnostic. This makes `valid.go` an **assertion of zero false positives**.

## Golden files

`.go.golden` files verify auto-fix output via [`analysistest.RunWithSuggestedFixes`](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest#RunWithSuggestedFixes). Each golden file corresponds to a positive test file and shows what the file should look like after all suggested fixes are applied.

Golden files are only needed for files that produce fixes. `valid.go` needs no golden file because it produces no diagnostics.

## How to add tests

### Adding a positive test case (new detection)

1. Create a new `.go` file in `testdata/example/` (or add to an existing positive test file).
2. Add `// want` annotations on lines that should produce diagnostics:
   ```go
   var empty = s == "" // want `use wegostrings\.IsEmpty`
   ```
3. Create a matching `.go.golden` file showing the expected fix:
   ```go
   var empty = wegostrings.IsEmpty(s)
   ```
4. Run `go test ./...` — the test should pass with the new annotations.

### Adding a negative test case (false positive guard)

1. Add the case to `valid.go`.
2. Do **not** add any `// want` annotation.
3. Run `go test ./...` — if the analyzer incorrectly flags your case, the test will fail with "unexpected diagnostic".

## The `testdata/` module

The `testdata/` directory is its own Go module with a `replace` directive pointing to `../../../strings`. After changing that dependency, run:

```bash
cd testdata && go mod tidy
```
