package scanProject

// 方便后续扩展字段
type FileItem struct {
	Path string
}

type ProjectNpmList = map[string]NpmPackage // key为 workspace名称，如果不是 monorepo 项目，则为 "root"

type NpmPackage struct {
	Workspace string             // 如果是 monorepo 项目，则表示所在的 workspace, 最外层或非monorepo项目否则为 "root"
	Path      string             // package.json 的路径
	Namespace string             // 包名的命名空间，例如 @sl/sc-product
	Version   string             // 包的版本号
	NpmList   map[string]NpmItem // npm列表，key为包名
}

type NpmItem struct {
	Name    string // 包名
	Type    string // 包类型: "devDependencies"、“peerDependencies”、“dependencies”
	Version string // 包版本号
}
