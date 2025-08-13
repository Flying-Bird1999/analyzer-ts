package project_analyzer

import (
	"encoding/json"
	"fmt"
	"main/analyzer/projectParser"
	"os"
	"path/filepath"
)

func AnalyzeProject(rootPath string, outputDir string, alias map[string]string, extensions []string, ignore []string, isMonorepo bool) {
	config := projectParser.NewProjectParserConfig(rootPath, alias, extensions, ignore, isMonorepo)
	ar := projectParser.NewProjectParserResult(config)
	ar.ProjectParser()

	jsonData, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %s\n", err)
		return
	}

	// Write to file
	outputFile := filepath.Join(outputDir, filepath.Base(rootPath)+".json")
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %s\n", err)
		return
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFile)
}
