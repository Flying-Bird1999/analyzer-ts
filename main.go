package main

import (
	_ "main/analyzer_plugin/project_analyzer/cmd"
	_ "main/analyzer_plugin/ts_bundle/cmd"
	"main/cmd"
)

func main() {
	cmd.Execute()
}
