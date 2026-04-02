package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/F3dosik/metalert/cmd/linter/linter"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
