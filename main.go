package main

import "main/bundle/analyze"

func main() {
	// bundle.GenerateBundle()

	// var sp = scanProject.NewProjectResult("/Users/zxc/Desktop/nova", []string{"compare/**"}, true)
	// sp.ScanProject()
	// npmList := sp.GetNpmList()
	// for k, v := range npmList {
	// 	fmt.Printf("key: %s, workspace: %s, path: %s, namespace: %s, version: %s\n", k, v.Workspace, v.Path, v.Namespace, v.Version)
	// }

	analyze.Analyze()
}
