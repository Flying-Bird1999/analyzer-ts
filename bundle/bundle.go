package bundle

import (
	"main/bundle/parser"
)

func GenerateBundle() {
	filePath := "/Users/zxc/Desktop/analyzer-ts/ts/example.ts"
	parser.Traverse(filePath)
}
