package ts_bundle

import (
	"fmt"
	"os"
)

// 入口方法
func GenerateBundle(inputAnalyzeFile string, inputAnalyzeType string, outputFile string, projectRoot string) {
	br := NewCollectResult(inputAnalyzeFile, inputAnalyzeType, projectRoot)
	br.collectFileType(inputAnalyzeFile, inputAnalyzeType, "", "")

	bundler := NewTypeBundler()
	bundledContent, _ := bundler.Bundle(br.SourceCodeMap)

	// 输出到文件
	err := os.WriteFile(outputFile, []byte(bundledContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		return
	}

	fmt.Printf("Bundle completed: %s\n", outputFile)
	fmt.Printf("\nName mappings:\n")
	for key, finalName := range bundler.FinalNameMap {
		fmt.Printf("  %s -> %s\n", key, finalName)
	}
}
