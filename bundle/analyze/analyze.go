package analyze

type AnalyzeResult struct {
	File map[string]FileAnalyzeResult
	Npm  string
}
