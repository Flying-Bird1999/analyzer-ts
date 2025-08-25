package ts_bundle

// 入口方法
func GenerateBundle(inputAnalyzeFile string, inputAnalyzeType string, outputFile string, projectRoot string) string {
	br := NewCollectResult(inputAnalyzeFile, inputAnalyzeType, projectRoot)
	br.collectFileType(inputAnalyzeFile, inputAnalyzeType, "", "")

	bundler := NewTypeBundler()
	bundledContent, _ := bundler.Bundle(br.SourceCodeMap)

	return bundledContent
}
