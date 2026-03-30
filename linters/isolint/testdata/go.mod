module example

go 1.24

require (
	github.com/wego/pkg/currency v0.1.0
	github.com/wego/pkg/iso/site v0.1.0
)

require (
	github.com/bojanz/currency v1.3.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
)

replace (
	github.com/wego/pkg/currency => ../../../currency
	github.com/wego/pkg/iso/site => ../../../iso/site
)
