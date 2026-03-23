package linter_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/F3dosik/metalert/cmd/linter/linter" // замените на реальный module path
)

func TestAnalyzer(t *testing.T) {
	// analysistest.TestData() ищет папку testdata рядом с тестом.
	testdata := analysistest.TestData()

	// Пакет с нарушениями — анализатор должен их найти.
	analysistest.Run(t, testdata, linter.Analyzer, "sample")

	// Пакет без нарушений — анализатор не должен ничего сообщать.
	analysistest.Run(t, testdata, linter.Analyzer, "good")
}
