# isolint

A Go analyzer that detects raw ISO code string literals and recommends using `github.com/wego/pkg/currency` and `github.com/wego/pkg/iso/site` package constants instead.

## Installation

### Automatic way (recommended)

This follows golangci-lint's "Automatic Way" module plugin flow.

Requirements: Go and git.

1. Create `.custom-gcl.yml` in your project:

```yaml
version: v2.8.0
plugins:
  - module: github.com/wego/pkg/linters/isolint
    version: v0.1.0
```

2. Build custom golangci-lint:

```bash
golangci-lint custom
```

3. Configure the plugin in `.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - isolint
  settings:
    custom:
      isolint:
        type: "module"
        description: "Enforces currency/site package constant usage"
```

4. Run the resulting custom binary:

```bash
./custom-gcl run ./...
```

### As a standalone tool

```bash
go install github.com/wego/pkg/linters/isolint/cmd/isolint@latest
isolint ./...
```

## What it detects

### Currency codes (ISO 4217)

| Pattern | Suggestion |
|---------|------------|
| `"USD"` | `currency.USD` |
| `"EUR"` | `currency.EUR` |
| `"SGD"` | `currency.SGD` |
| ... | All 182 ISO 4217 codes |

### Site codes (ISO 3166-1 alpha-2)

| Pattern | Suggestion |
|---------|------------|
| `"SG"` | `site.SG` |
| `"US"` | `site.US` |
| `"JP"` | `site.JP` |
| ... | All 249 ISO 3166-1 alpha-2 codes |

### Detected positions

The linter catches raw ISO code strings in **all** Go expression positions:

- Comparisons: `code == "USD"`, `"SG" != code`
- Assignments: `x := "USD"`, `var x = "SG"`
- Constant/var declarations: `const c = "USD"`
- Switch/case: `case "USD":`, `case "SG":`
- Map keys/values: `map[string]int{"USD": 1}`, `m["SG"]`
- Function arguments: `foo("USD")`, `fmt.Println("SG")`
- Struct fields: `Config{Currency: "USD"}`
- Return values: `return "USD"`
- Slice/array literals: `[]string{"USD", "SGD"}`

### What it skips

- Files in `github.com/wego/pkg/currency` (defines the currency constants)
- Files in `github.com/wego/pkg/iso/site` (defines the site constants)
- Non-ISO strings like `"hello"`, `""`, `"test"`
- Lowercase strings like `"usd"`, `"sg"`
- Code already using constants: `currency.USD`, `site.SG`

## Import convention

```go
import (
    "github.com/wego/pkg/currency"
    "github.com/wego/pkg/iso/site"
)

// The linter suggests:
if code == currency.USD { ... }
if siteCode == site.SG { ... }
```

## Auto-fix

The linter provides suggested fixes that can be applied automatically:

```bash
# With golangci-lint
./custom-gcl run --fix ./...

# With standalone tool
isolint -fix ./...
```

**Note**: Auto-fix replaces the string literal but does not add the import statement. You will need to:
1. Run `goimports` to add missing imports
2. Verify the correct package is imported

## Development

### Local tests

The testdata directory is a standalone module. Run tests from the module root:

```bash
go test -v ./...
```

### Using a commit before tagging

If you need to consume an untagged commit from another repo, use a Go pseudo-version
instead of a raw SHA.

```bash
go list -m -json github.com/wego/pkg/linters/isolint@<commit>
```

Then use the returned `Version` value in `.custom-gcl.yml`:

```yaml
version: v2.8.0
plugins:
  - module: github.com/wego/pkg/linters/isolint
    version: v0.0.0-20260120hhmmss-abcdef123456
```
