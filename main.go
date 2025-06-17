package main

import "main/bundle"

func main() {
	// 多包npm包扫描调试
	// var sp = scanProject.NewProjectResult("/Users/zxc/Desktop/nova", []string{"compare/**"}, true)
	// sp.ScanProject()
	// npmList := sp.GetNpmList()
	// for k, v := range npmList {
	// 	fmt.Printf("key: %s, workspace: %s, path: %s, namespace: %s, version: %s\n", k, v.Workspace, v.Path, v.Namespace, v.Version)
	// }

	// 对项目依赖分析，生成分析结果，bundle/analyze/analyze_output.txt
	// analyze.Analyze()

	// 生成ts bundle结果：ts/output/result.ts
	bundle.GenerateBundle()
}
