package ts_bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

// 入口方法
func GenerateBundle() {
	inputAnalyzeFile, _ := filepath.Abs("ts/bundle/index1.ts")
	inputAnalyzeType := "Class"

	br := NewCollectResult(inputAnalyzeFile, inputAnalyzeType, "")
	br.collectFileType(inputAnalyzeFile, inputAnalyzeType, "", "")

	bundler := NewTypeBundler()
	bundledContent, _ := bundler.Bundle(br.SourceCodeMap)

	// 输出到文件
	outputFile := "./ts/output/result.ts"
	err := os.WriteFile(outputFile, []byte(bundledContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		return
	}

	fmt.Printf("Bundle completed: %s\n", outputFile)
	fmt.Printf("\nName mappings:\n")
	for key, finalName := range bundler.finalNameMap {
		fmt.Printf("  %s -> %s\n", key, finalName)
	}
}
