package scanProject

// FileItem 文件信息
type FileItem struct {
	FileName string `json:"fileName"` // 文件名
	Size     int64  `json:"size"`     // 大小
	Ext      string `json:"ext"`      // 后缀
}