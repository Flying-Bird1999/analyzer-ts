package scanProject

// 方便后续扩展字段
type FileItem struct {
	Path string
}

// 方便后续扩展字段
type NpmItem struct {
	Workspace string // 如果是 monorepo 项目，则表示所在的 workspace, 否则为空
	Name      string
	Type      string
	Version   string
}
