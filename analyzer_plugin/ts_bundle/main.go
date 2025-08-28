package ts_bundle

// 入口方法
func GenerateBundle(inputAnalyzeFile string, inputAnalyzeType string, projectRoot string) (string, error) {
	br := NewCollectResult(inputAnalyzeFile, inputAnalyzeType, projectRoot)
	if err := br.collectFileType(inputAnalyzeFile, inputAnalyzeType, "", ""); err != nil {
		return "", err
	}
	bundler := NewTypeBundler()
	bundledContent, err := bundler.Bundle(br.SourceCodeMap)
	return bundledContent, err
}