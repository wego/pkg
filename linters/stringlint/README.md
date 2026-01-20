# stringlint

A Go analyzer that detects direct string comparison patterns and recommends using `github.com/wego/pkg/strings` utility functions.

## Installation

### Automatic way (recommended)

This follows golangci-lint's "Automatic Way" module plugin flow.

Requirements: Go and git.

1. Create `.custom-gcl.yml` in your project:

```yaml
version: v2.8.0
plugins:
  - module: github.com/wego/pkg/linters/stringlint
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
    - stringlint
  settings:
    custom:
      stringlint:
        type: "module"
        description: "Enforces wego/pkg/strings usage"
```

4. Run the resulting custom binary:

```bash
./custom-gcl run ./...
```

### As a standalone tool

```bash
go install github.com/wego/pkg/linters/stringlint/cmd/stringlint@latest
stringlint ./...
```

## What it detects

| Pattern | Suggestion |
|---------|------------|
| `s == ""` | `wegostrings.IsEmpty(s)` |
| `s != ""` | `wegostrings.IsNotEmpty(s)` |
| `len(s) == 0` | `wegostrings.IsEmpty(s)` |
| `len(s) != 0` | `wegostrings.IsNotEmpty(s)` |
| `len(s) > 0` | `wegostrings.IsNotEmpty(s)` |
| `*ptr == ""` | `wegostrings.IsEmptyP(ptr)` |
| `*ptr != ""` | `wegostrings.IsNotEmptyP(ptr)` |

## Import Convention

The linter uses the alias `wegostrings` to avoid conflict with the stdlib `strings` package:

```go
import wegostrings "github.com/wego/pkg/strings"

// The linter suggests:
if wegostrings.IsEmpty(s) { ... }
```

## Auto-fix

The linter provides suggested fixes that can be applied automatically:

```bash
# With golangci-lint
./custom-gcl run --fix ./...

# With standalone tool
stringlint -fix ./...
```

**Note**: Auto-fix replaces the comparison but does not add the import statement. You will need to:
1. Run `goimports` to add missing imports
2. Ensure the import uses the `wegostrings` alias

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
go list -m -json github.com/wego/pkg/linters/stringlint@<commit>
```

Then use the returned `Version` value in `.custom-gcl.yml`:

```yaml
version: v2.8.0
plugins:
  - module: github.com/wego/pkg/linters/stringlint
    version: v0.0.0-20260120hhmmss-abcdef123456
```
