package bundle

import (
	"fmt"
	"main/bundle/parser"
	"path/filepath"
)

func GenerateBundle() {
	filePath, err := filepath.Abs("./ts/example.ts")
	if err != nil {
		fmt.Printf("读取目录失败")
	}
	parser.Traverse(filePath)
}
