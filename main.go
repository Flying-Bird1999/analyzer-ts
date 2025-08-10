package main

import (
	"main/analyzer/parser"
	"main/cmd"
)

func main() {
	cmd.Execute()
	parser.Parser_run()
}
