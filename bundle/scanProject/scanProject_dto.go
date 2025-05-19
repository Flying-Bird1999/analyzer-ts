package scanProject

// 方便后续扩展字段
type FileItem struct {
	Path string
}

// 方便后续扩展字段
type NpmItem struct {
	Name    string
	Type    string
	Version string
}
