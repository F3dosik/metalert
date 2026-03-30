// Package linter содержит статический анализатор, который проверяет:
// 1. Использование встроенной функции panic.
// 2. Вызов log.Fatal вне функции main пакета main.
// 3. Вызов os.Exit вне функции main пакета main.
package linter

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer — статический анализатор.
var Analyzer = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "запрещает panic, log.Fatal и os.Exit вне main.main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Интересуют только вызовы функций.
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		pos := pass.Fset.Position(n.Pos())
		for _, f := range pass.Files {
			if pass.Fset.Position(f.Pos()).Filename == pos.Filename {
				if isGeneratedFile(f) {
					return false
				}
				break
			}
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		switch fn := call.Fun.(type) {

		// panic(...)
		case *ast.Ident:
			if fn.Name == "panic" {
				pass.Reportf(call.Pos(), "использование panic запрещено")
			}

		// pkg.Func(...)
		case *ast.SelectorExpr:
			pkgIdent, ok := fn.X.(*ast.Ident)
			if !ok {
				return true
			}

			pkg := pkgIdent.Name
			name := fn.Sel.Name

			isForbidden := (pkg == "log" && name == "Fatal") ||
				(pkg == "os" && name == "Exit")

			if !isForbidden {
				return true
			}

			// Разрешено только внутри функции main пакета main.
			if pass.Pkg.Name() == "main" && insideMainFunc(stack) {
				return true
			}

			if pkg == "log" {
				pass.Reportf(call.Pos(), "вызов log.Fatal запрещён вне функции main пакета main")
			} else {
				pass.Reportf(call.Pos(), "вызов os.Exit запрещён вне функции main пакета main")
			}
		}

		return true
	})

	return nil, nil
}

// insideMainFunc возвращает true, если узел находится
// непосредственно внутри функции с именем main.
func insideMainFunc(stack []ast.Node) bool {
	for i := len(stack) - 1; i >= 0; i-- {
		fd, ok := stack[i].(*ast.FuncDecl)
		if !ok {
			continue
		}
		// Метод (есть receiver) — не main.
		if fd.Recv != nil {
			return false
		}
		return fd.Name.Name == "main"
	}
	return false
}

// isGeneratedFile возвращает true, если файл был сгенерирован
func isGeneratedFile(f *ast.File) bool {
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "DO NOT EDIT") ||
				strings.Contains(c.Text, "Code generated") {
				return true
			}
		}
	}
	return false
}
