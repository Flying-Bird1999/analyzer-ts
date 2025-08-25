package main

import (
	_ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/project_analyzer/cmd"
	_ "github.com/Flying-Bird1999/analyzer-ts/analyzer_plugin/ts_bundle/cmd"
	"github.com/Flying-Bird1999/analyzer-ts/cmd"
)

func main() {
	cmd.Execute()
}
