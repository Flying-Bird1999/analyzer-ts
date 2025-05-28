package main

import (
	"fmt"
	"main/bundle"
	"main/bundle/parser"
)

func main() {
	bundle.GenerateBundle()

	// ------------------------------------------------------------

	pr := parser.NewParserResult("/Users/bird/Desktop/alalyzer/analyzer-ts/ts/export/export.ts")
	pr.Traverse()
	exportData := pr.GetResult().ExportDeclarations
	for _, export := range exportData {
		fmt.Printf("export Raw: %s\n", export.Raw)
	}
}
