package main

import (
	gostyle "github.com/k1LoW/gostyle/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
)

func main() {
	multichecker.Main(append(gostyle.Analyzers, nilness.Analyzer, shadow.Analyzer, unusedwrite.Analyzer)...)
}
