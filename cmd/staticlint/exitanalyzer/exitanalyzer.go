package exitanalyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ErrCheckAnalyzer is a static analysis analyzer that forbids direct calls to os.Exit in the main function of the
// main package.
//
// This analyzer helps to improve code testability and control flow management by ensuring that the main function
// returns an exit code instead of calling os.Exit directly.
//
// Example of forbidden code:
//
//	package main
//	import "os"
//	func main() {
//	    os.Exit(1) // This will be reported
//	}
//
// Example of allowed code:
//
//	package main
//	func main() {
//	    return // Return exit code instead
//	}
var ErrCheckAnalyzer = &analysis.Analyzer{
	Name: "exitanalyzer",
	Doc:  "forbid direct calls to os.Exit in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if pass.Pkg.Name() != "main" {
				return true
			}

			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			fun, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			selectorIdent, ok := fun.X.(*ast.Ident)
			if !ok || selectorIdent.Name != "os" || fun.Sel.Name != "Exit" {
				return true
			}

			parent := findParentFunction(pass, callExpr.Pos())
			if parent == nil || parent.Name.Name != "main" {
				return true
			}

			pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function is forbidden")

			return true
		})
	}
	return nil, nil
}

func findParentFunction(pass *analysis.Pass, pos token.Pos) *ast.FuncDecl {
	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), ".go") {
			continue
		}

		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Pos() <= pos && pos <= fn.End() {
					return fn
				}
			}
		}
	}
	return nil
}
