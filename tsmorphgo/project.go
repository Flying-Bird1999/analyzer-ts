package tsmorphgo

import (
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// Project 代表一个完整的 TypeScript 项目的视图，提供了与 ts-morph 类似的 API。
type Project struct {
	parserResult *projectParser.ProjectParserResult
	sourceFiles  map[string]*SourceFile
}

// ProjectConfig 定义了初始化一个新项目所需的配置。
type ProjectConfig struct {
	RootPath         string
	IgnorePatterns   []string
	IsMonorepo       bool
	TargetExtensions []string
}

// NewProject 是创建和初始化一个新项目实例的入口点。
func NewProject(config ProjectConfig) *Project {
	ppConfig := projectParser.NewProjectParserConfig(config.RootPath, config.IgnorePatterns, config.IsMonorepo, config.TargetExtensions)
	ppResult := projectParser.NewProjectParserResult(ppConfig)
	ppResult.ProjectParser()

	p := &Project{
		parserResult: ppResult,
		sourceFiles:  make(map[string]*SourceFile),
	}

	for path, jsResult := range ppResult.Js_Data {
		sf := &SourceFile{
			filePath:      path,
			fileResult:    &jsResult,
			astNode:       jsResult.Ast,
			project:       p,
			nodeResultMap: make(map[*ast.Node]interface{}),
		}
		p.sourceFiles[path] = sf
		sf.buildNodeResultMap()
	}

	return p
}

// NewProjectFromSources 从内存中的源码 map 创建一个新项目。
func NewProjectFromSources(sources map[string]string) *Project {
	ppConfig := projectParser.NewProjectParserConfig("/", nil, false, nil)
	ppResult := projectParser.NewProjectParserResult(ppConfig)
	ppResult.ProjectParserFromMemory(sources)

	p := &Project{
		parserResult: ppResult,
		sourceFiles:  make(map[string]*SourceFile),
	}

	for path, jsResult := range ppResult.Js_Data {
		sf := &SourceFile{
			filePath:      path,
			fileResult:    &jsResult,
			astNode:       jsResult.Ast,
			project:       p,
			nodeResultMap: make(map[*ast.Node]interface{}),
		}
		p.sourceFiles[path] = sf
		sf.buildNodeResultMap()
	}

	return p
}

// GetSourceFile 根据文件路径从项目中获取一个 SourceFile 实例。
func (p *Project) GetSourceFile(path string) *SourceFile {
	return p.sourceFiles[path]
}

// GetSourceFiles 返回项目中的所有源文件
func (p *Project) GetSourceFiles() []*SourceFile {
	files := make([]*SourceFile, 0, len(p.sourceFiles))
	for _, file := range p.sourceFiles {
		files = append(files, file)
	}
	return files
}

// findNodeAt 在指定的源文件中，根据行列号查找最精确匹配的 AST 节点。
func (p *Project) findNodeAt(filePath string, line, char int) *ast.Node {
	sf, ok := p.sourceFiles[filePath]
	if !ok {
		return nil
	}

	lines := strings.Split(sf.fileResult.Raw, "\n")
	if line-1 >= len(lines) {
		return nil
	}
	offset := 0
	for i := 0; i < line-1; i++ {
		offset += len(lines[i]) + 1
	}
	offset += char - 1

	var foundNode *ast.Node
	var smallestSpan int = -1

	var walk func(node *ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		start, end := node.Pos(), node.End()
		if start <= offset && offset < end {
			span := end - start
			if smallestSpan == -1 || span < smallestSpan {
				smallestSpan = span
				foundNode = node
			}
			node.ForEachChild(func(child *ast.Node) bool {
				walk(child)
				return false
			})
		}
	}

	walk(sf.astNode)
	return foundNode
}
