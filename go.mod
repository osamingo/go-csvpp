module github.com/osamingo/go-csvpp

go 1.25.6

require (
	github.com/goccy/go-yaml v1.19.2
	github.com/google/go-cmp v0.7.0
	github.com/k1LoW/gostyle v0.25.2
	golang.org/x/tools v0.40.0
)

require (
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/gostaticanalysis/comment v1.5.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)

tool (
	github.com/osamingo/go-csvpp/cmd/csvppvet
	mvdan.cc/gofumpt
)
