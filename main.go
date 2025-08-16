package main

import (
	"main/cmd"
	_ "main/analyzer_plugin/project_analyzer/cmd"
	_ "main/analyzer_plugin/ts_bundle/cmd"
)

func main() {
	cmd.Execute()
}
