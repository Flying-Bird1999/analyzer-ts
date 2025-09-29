package api_tracer

// ApiCallSite 代表一个API调用在代码中的具体位置和相关信息。
type ApiCallSite struct {
	// ApiPath 是匹配到的API路径字符串。
	ApiPath string `json:"apiPath"`
	// FilePath 是包含此次API调用的文件的绝对路径。
	FilePath string `json:"filePath"`
	// Raw 是该调用表达式在源代码中的原始文本。
	Raw string `json:"raw"`
}
