package symbol_analysis

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// =============================================================================
// 辅助函数
// =============================================================================

// getLineNumberFromAST 从 AST 节点安全地获取行号。
func getLineNumberFromAST(node *ast.Node) int {
	if node == nil {
		return 0
	}
	// 使用源文件中的行号
	// 这是一个简化的方法；在生产环境中，你需要从位置计算行号
	return 0 // 占位符 - 将从源文件计算
}

// getKindNameFromAST 从 AST 节点安全地获取种类名称。
func getKindNameFromAST(node *ast.Node) string {
	if node == nil {
		return ""
	}
	return node.Kind.String()
}

// getLineNumber 从 tsmorphgo 节点安全地获取行号。
func getLineNumber(node *tsmorphgo.Node) int {
	if node == nil || !node.IsValid() {
		return 0
	}
	return node.GetStartLineNumber()
}

// getKindName 从 tsmorphgo 节点安全地获取种类名称。
func getKindName(node *tsmorphgo.Node) string {
	if node == nil || !node.IsValid() {
		return ""
	}
	return node.GetKindName()
}

// isAlpha 检查字符是否为字母。
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isAlphaNum 检查字符是否为字母数字。
func isAlphaNum(c rune) bool {
	if c >= 'a' && c <= 'z' {
		return true
	}
	if c >= 'A' && c <= 'Z' {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	return false
}

// =============================================================================
// 公共 API
// =============================================================================

// AnalyzeChangedLines 分析变更行并返回每个文件的受影响符号。
// 这是符号分析的主要入口点。
// 对于无法进行符号分析的文件（如二进制文件、CSS等），也会返回基本信息。
func (a *Analyzer) AnalyzeChangedLines(
	lineSets ChangedLineSetOfFiles,
) map[string]*FileAnalysisResult {
	results := make(map[string]*FileAnalysisResult)

	for filePath, changedLines := range lineSets {
		result, err := a.AnalyzeFile(filePath, changedLines)
		if err != nil {
			// 无法进行符号分析，返回非符号文件的基本信息
			results[filePath] = a.createNonSymbolFileResult(filePath, changedLines)
		} else {
			results[filePath] = result
		}
	}

	return results
}

// createNonSymbolFileResult 为非符号文件创建基本的分析结果。
func (a *Analyzer) createNonSymbolFileResult(filePath string, changedLines map[int]bool) *FileAnalysisResult {
	// 转换变更行为切片
	changedLineList := make([]int, 0, len(changedLines))
	for line := range changedLines {
		changedLineList = append(changedLineList, line)
	}

	return &FileAnalysisResult{
		FilePath:        filePath,
		FileType:        determineFileType(filePath),
		AffectedSymbols: []SymbolChange{},
		FileExports:     []ExportInfo{},
		ChangedLines:    changedLineList,
		IsSymbolFile:    false,
	}
}

// determineFileType 根据文件扩展名确定文件类型。
func determineFileType(filePath string) FileType {
	// 获取文件扩展名
	ext := filepath.Ext(filePath)
	if ext == "" {
		return FileTypeUnknown
	}

	// 转换为小写
	ext = strings.ToLower(ext)

	// 根据扩展名判断文件类型
	switch ext {
	case ".ts", ".tsx", ".mts":
		return FileTypeTypeScript
	case ".js", ".jsx", ".mjs":
		return FileTypeJavaScript
	case ".css", ".scss", ".sass", ".less":
		return FileTypeStyle
	case ".html", ".htm", ".xml", "svg":
		return FileTypeMarkup
	case ".json", ".yaml", ".yml", "toml":
		return FileTypeData
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".webp",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".mp3", ".mp4", ".wav", ".ogg", ".webm",
		".zip", ".tar", ".gz", ".rar":
		return FileTypeBinary
	default:
		return FileTypeUnknown
	}
}

// AnalyzeFile 分析单个文件并返回受影响的符号。
func (a *Analyzer) AnalyzeFile(
	filePath string,
	changedLines map[int]bool,
) (*FileAnalysisResult, error) {
	sourceFile := a.project.GetSourceFile(filePath)
	if sourceFile == nil {
		return nil, fmt.Errorf("文件未找到: %s", filePath)
	}

	// 转换变更行为切片
	changedLineList := make([]int, 0, len(changedLines))
	for line := range changedLines {
		changedLineList = append(changedLineList, line)
	}

	result := &FileAnalysisResult{
		FilePath:        filePath,
		FileType:        determineFileType(filePath),
		AffectedSymbols: make([]SymbolChange, 0),
		FileExports:     make([]ExportInfo, 0),
		ChangedLines:    changedLineList,
		IsSymbolFile:    true,
	}

	// 步骤 1: 从文件结果中提取导出信息
	a.extractExportsFromFileResult(sourceFile, result)

	// 步骤 2: 通过遍历 AST 查找受影响的符号
	a.extractAffectedSymbols(sourceFile, changedLines, result)

	// 步骤 3: 为受影响的符号填充导出信息
	a.fillExportInfo(result)

	return result, nil
}

// =============================================================================
// 导出提取
// =============================================================================

// extractExportsFromFileResult 从文件的解析结果中提取导出信息。
func (a *Analyzer) extractExportsFromFileResult(
	sourceFile *tsmorphgo.SourceFile,
	result *FileAnalysisResult,
) {
	fileResult := sourceFile.GetFileResult()
	if fileResult == nil {
		return
	}

	// 从 ExportDeclarations 提取（例如，export { A, B }）
	for _, exportDecl := range fileResult.ExportDeclarations {
		for _, module := range exportDecl.ExportModules {
			exportType := ExportTypeNamed
			if module.Type == "default" {
				exportType = ExportTypeDefault
			} else if module.Type == "namespace" {
				exportType = ExportTypeNamespace
			}

			// 如果可能，从原始源代码计算行号
			lineNum := 0
			if exportDecl.Node != nil {
				lineNum = a.calculateLineNumber(sourceFile, exportDecl.Node)
			}

			result.FileExports = append(result.FileExports, ExportInfo{
				Name:       module.Identifier,
				ExportType: exportType,
				DeclLine:   lineNum,
				DeclNode:   "ExportDeclaration",
			})
		}
	}

	// 从 ExportAssignments 提取（例如，export default class Button）
	for _, exportAssign := range fileResult.ExportAssignments {
		// 如果可能，从原始源代码计算行号
		lineNum := 0
		if exportAssign.Node != nil {
			lineNum = a.calculateLineNumber(sourceFile, exportAssign.Node)
		}

		result.FileExports = append(result.FileExports, ExportInfo{
			Name:       a.extractDefaultExportNameFromAssign(exportAssign),
			ExportType: ExportTypeDefault,
			DeclLine:   lineNum,
			DeclNode:   "ExportAssignment",
		})
	}

	// 从带有内联导出的 VariableDeclarations 提取（例如，export const A = 1）
	for _, varDecl := range fileResult.VariableDeclarations {
		if varDecl.Exported {
			// 从声明器中提取符号名称
			for _, declarator := range varDecl.Declarators {
				if declarator.Identifier != "" {
					lineNum := 0
					if varDecl.Node != nil {
						lineNum = a.calculateLineNumber(sourceFile, varDecl.Node)
					}

					result.FileExports = append(result.FileExports, ExportInfo{
						Name:       declarator.Identifier,
						ExportType: ExportTypeNamed,
						DeclLine:   lineNum,
						DeclNode:   "VariableDeclaration",
					})
				}
			}
		}
	}

	// 从带有内联导出的 FunctionDeclarations 提取（例如，export function A() {}）
	for _, fnDecl := range fileResult.FunctionDeclarations {
		if fnDecl.Exported {
			lineNum := 0
			if fnDecl.Node != nil {
				lineNum = a.calculateLineNumber(sourceFile, fnDecl.Node)
			}

			result.FileExports = append(result.FileExports, ExportInfo{
				Name:       fnDecl.Identifier,
				ExportType: ExportTypeNamed,
				DeclLine:   lineNum,
				DeclNode:   "FunctionDeclaration",
			})
		}
	}

	// 从带有内联导出的 InterfaceDeclarations 提取（例如，export interface A {}）
	for _, ifaceDecl := range fileResult.InterfaceDeclarations {
		if ifaceDecl.Exported {
			lineNum := 0
			if ifaceDecl.Node != nil {
				lineNum = a.calculateLineNumber(sourceFile, ifaceDecl.Node)
			}

			result.FileExports = append(result.FileExports, ExportInfo{
				Name:       ifaceDecl.Identifier,
				ExportType: ExportTypeNamed,
				DeclLine:   lineNum,
				DeclNode:   "InterfaceDeclaration",
			})
		}
	}

	// 从带有内联导出的 TypeDeclarations 提取（例如，export type A = ...）
	for _, typeDecl := range fileResult.TypeDeclarations {
		if typeDecl.Exported {
			lineNum := 0
			if typeDecl.Node != nil {
				lineNum = a.calculateLineNumber(sourceFile, typeDecl.Node)
			}

			result.FileExports = append(result.FileExports, ExportInfo{
				Name:       typeDecl.Identifier,
				ExportType: ExportTypeNamed,
				DeclLine:   lineNum,
				DeclNode:   "TypeDeclaration",
			})
		}
	}

	// 从带有内联导出的 EnumDeclarations 提取（例如，export enum A {...}）
	for _, enumDecl := range fileResult.EnumDeclarations {
		if enumDecl.Exported {
			lineNum := 0
			if enumDecl.Node != nil {
				lineNum = a.calculateLineNumber(sourceFile, enumDecl.Node)
			}

			result.FileExports = append(result.FileExports, ExportInfo{
				Name:       enumDecl.Identifier,
				ExportType: ExportTypeNamed,
				DeclLine:   lineNum,
				DeclNode:   "EnumDeclaration",
			})
		}
	}

	// 通过遍历 AST 提取导出的类（parser 没有 ClassDeclarationResult）
	a.extractExportedClasses(sourceFile, result)
}

// extractExportedClasses 通过遍历 AST 提取导出的类声明。
// 这是必要的，因为 parser 没有专门的 ClassDeclarationResult。
// 使用 AST 节点关系和修饰符而不是字符串匹配来进行准确分析。
func (a *Analyzer) extractExportedClasses(sourceFile *tsmorphgo.SourceFile, result *FileAnalysisResult) {
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if !node.IsClassDeclaration() {
			return
		}

		// 检查类是否有 export 修饰符（核心 AST-based 检查）
		hasExportKeyword := false
		isDefaultExport := false

		// 方法1: 检查子节点中的 ExportKeyword（最准确）
		node.ForEachChild(func(child tsmorphgo.Node) bool {
			if child.Kind == ast.KindExportKeyword {
				hasExportKeyword = true
				// 继续检查是否有 DefaultKeyword
				return false // 继续检查 default 关键字
			}
			if child.Kind == ast.KindDefaultKeyword {
				isDefaultExport = true
				hasExportKeyword = true
			}
			return false
		})

		// 方法2: 检查父节点关系（处理 export { Class } 的情况）
		if !hasExportKeyword {
			parent := node.GetParent()
			if parent.IsValid() && parent.IsExportDeclaration() {
				hasExportKeyword = true
			}
		}

		// 方法3: 检查 FileResult 中的 ExportAssignments
		// 处理 "class A {} export default A" 这种独立导出的情况
		if !hasExportKeyword {
			fileResult := sourceFile.GetFileResult()
			className := a.extractClassNameFromNode(node)
			for _, exportAssign := range fileResult.ExportAssignments {
				if className != "" && strings.Contains(exportAssign.Raw, className) {
					hasExportKeyword = true
					if strings.Contains(exportAssign.Raw, "export default") {
						isDefaultExport = true
					}
					break
				}
			}
		}

		if !hasExportKeyword {
			return
		}

		// 确定导出类型
		exportType := ExportTypeNamed
		if isDefaultExport {
			exportType = ExportTypeDefault
		}

		// 提取类名：使用 AST 节点遍历而非字符串解析
		className := a.extractClassNameFromNode(node)
		if className == "" {
			className = "default"
		}

		lineNum := node.GetStartLineNumber()

		result.FileExports = append(result.FileExports, ExportInfo{
			Name:       className,
			ExportType: exportType,
			DeclLine:   lineNum,
			DeclNode:   "ClassDeclaration",
		})
	})
}

// extractClassNameFromNode 从 ClassDeclaration AST 节点提取类名。
// 使用 AST 遍历查找 Identifier 节点，而不是字符串解析。
func (a *Analyzer) extractClassNameFromNode(node tsmorphgo.Node) string {
	var className string

	// 遍历子节点查找 Identifier 节点
	node.ForEachChild(func(child tsmorphgo.Node) bool {
		if child.IsIdentifier() {
			className = child.GetText()
			return true // 找到后停止
		}
		return false
	})

	return className
}

// extractSymbolNameFromExport 从导出声明中提取符号名称。
// 例如，从 "export const A = 1" 中提取 "A"。
func (a *Analyzer) extractSymbolNameFromExport(raw string) string {
	// 移除 "export " 前缀
	afterExport := strings.TrimPrefix(raw, "export ")
	afterExport = strings.TrimSpace(afterExport)

	// 处理不同的模式
	// 1. "const A = 1" -> 提取 "A"
	// 2. "function A()" -> 提取 "A"
	// 3. "class A" -> 提取 "A"

	// 查找第一个标识符
	parts := strings.Fields(afterExport)
	if len(parts) > 0 {
		firstPart := parts[0]
		// 移除常见关键字
		if firstPart == "const" || firstPart == "let" || firstPart == "var" {
			if len(parts) > 1 {
				return parts[1] // 下一部分应该是名称
			}
		} else if firstPart == "function" {
			if len(parts) > 1 {
				// 提取函数名称（可能有括号）
				funcName := parts[1]
				// 如果有括号则移除
				funcName = strings.TrimSuffix(funcName, "(")
				funcName = strings.TrimSpace(funcName)
				if funcName != "" {
					return funcName
				}
			}
		} else if firstPart == "class" {
			if len(parts) > 1 {
				return parts[1]
			}
		} else if firstPart == "interface" || firstPart == "type" || firstPart == "enum" {
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}

	// 后备方案：尝试从原始文本中提取标识符
	for i, c := range afterExport {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			// 找到标识符的开始
			end := len(afterExport)
			for j := i + 1; j < end; j++ {
				c := afterExport[j]
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
					end = j
					break
				}
			}
			return afterExport[i:end]
		}
	}

	return ""
}

// calculateLineNumber 计算 AST 节点的行号。
func (a *Analyzer) calculateLineNumber(sourceFile *tsmorphgo.SourceFile, node *ast.Node) int {
	if node == nil {
		return 0
	}

	// 获取源代码
	sourceCode := sourceFile.GetFileResult().Raw
	if sourceCode == "" {
		return 0
	}

	pos := node.Pos()
	if pos <= 0 {
		return 0
	}

	// 从位置计算行号
	line := 1
	for i, c := range sourceCode {
		if i >= int(pos) {
			break
		}
		if c == '\n' {
			line++
		}
	}

	return line
}

// extractDefaultExportNameFromAssign 使用 AST 从默认导出赋值中提取名称。
// 例如，从 "export default MyClass" 中提取 "MyClass"。
// 对于匿名表达式（如 export default () => {}），返回 "default"。
func (a *Analyzer) extractDefaultExportNameFromAssign(exportAssign parser.ExportAssignmentResult) string {
	// 使用 parser 提取的 Expression 字段（这是 parser 通过 AST 进行的提取）
	if exportAssign.Expression != "" {
		expr := strings.TrimSpace(exportAssign.Expression)
		// 检查是否是有效的标识符（例如：MyClass, helper 等）
		// 如果是表达式（如：() => {}, {} 等），则返回 "default"
		if a.isValidIdentifier(expr) {
			return expr
		}
	}

	// 后备方案：为匿名导出返回 "default"
	return "default"
}

// isValidIdentifier 检查字符串是否是有效的标识符
func (a *Analyzer) isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	// 简单检查：标识符应该以字母或下划线开头，只包含字母数字和下划线
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}

// =============================================================================
// 符号提取
// =============================================================================

// extractAffectedSymbols 通过遍历 AST 提取受影响的符号。
func (a *Analyzer) extractAffectedSymbols(
	sourceFile *tsmorphgo.SourceFile,
	changedLines map[int]bool,
	result *FileAnalysisResult,
) {
	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		// 检查节点是否受变更影响
		if !a.isAffectedByChanges(node, changedLines) {
			return
		}

		// 只处理我们关心的声明节点
		if !a.isDeclarationNode(node) {
			return
		}

		// 检查我们是否应该包含此符号（传递 sourceFile 用于导出检查）
		if !a.shouldIncludeSymbol(sourceFile, node) {
			return
		}

		symbol := a.extractSymbolChange(node, changedLines)
		if symbol != nil {
			result.AffectedSymbols = append(result.AffectedSymbols, *symbol)
		}
	})
}

// =============================================================================
// 内部方法
// =============================================================================

// isAffectedByChanges 检查节点是否受给定行变更的影响。
func (a *Analyzer) isAffectedByChanges(
	node tsmorphgo.Node,
	changedLines map[int]bool,
) bool {
	startLine := node.GetStartLineNumber()
	endLine := node.GetEndLineNumber()

	for line := startLine; line <= endLine; line++ {
		if changedLines[line] {
			return true
		}
	}
	return false
}

// fillExportInfo 为受影响的符号填充导出信息。
// 它构建导出索引，然后将受影响的符号与之匹配。
func (a *Analyzer) fillExportInfo(result *FileAnalysisResult) {
	// 构建导出索引以供快速查找
	exportIndex := make(map[string]ExportInfo)
	defaultExports := make(map[string]bool) // 支持多个默认导出

	for _, export := range result.FileExports {
		if export.ExportType == ExportTypeDefault {
			defaultExports[export.Name] = true
		} else {
			exportIndex[export.Name] = export
		}
	}

	// 为每个受影响的符号填充导出信息
	for i := range result.AffectedSymbols {
		symbol := &result.AffectedSymbols[i]

		// 检查命名导出
		if export, ok := exportIndex[symbol.Name]; ok {
			symbol.IsExported = true
			symbol.ExportType = export.ExportType
			continue
		}

		// 检查默认导出
		// 支持类、函数、变量等所有类型的默认导出
		if defaultExports[symbol.Name] {
			symbol.IsExported = true
			symbol.ExportType = ExportTypeDefault
			continue
		}

		// 检查节点本身是否有导出修饰符
		sourceFile := a.project.GetSourceFile(symbol.FilePath)
		if sourceFile != nil {
			node := a.findSymbolNode(sourceFile, symbol)
			if node.IsValid() && a.hasExportModifier(sourceFile, node) {
				symbol.IsExported = true
				symbol.ExportType = a.getExportTypeFromNode(node)
			}
		}
	}
}

// findSymbolNode 查找对应于符号的 AST 节点。
func (a *Analyzer) findSymbolNode(
	sourceFile *tsmorphgo.SourceFile,
	symbol *SymbolChange,
) tsmorphgo.Node {
	var foundNode tsmorphgo.Node

	sourceFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if !foundNode.IsValid() {
			if node.GetStartLineNumber() == symbol.StartLine &&
				node.GetEndLineNumber() == symbol.EndLine {
				// 附加检查：验证符号名称匹配
				if a.extractSymbolName(node) == symbol.Name {
					foundNode = node
				}
			}
		}
	})

	return foundNode
}

// hasExportModifier 通过检查文件结果来检查节点是否有导出修饰符。
func (a *Analyzer) hasExportModifier(sourceFile *tsmorphgo.SourceFile, node tsmorphgo.Node) bool {
	fileResult := sourceFile.GetFileResult()
	if fileResult == nil {
		return false
	}

	nodeLine := node.GetStartLineNumber()

	// 检查 ExportDeclarations
	for _, exportDecl := range fileResult.ExportDeclarations {
		if exportDecl.Node != nil {
			exportLine := a.calculateLineNumber(sourceFile, exportDecl.Node)
			if exportLine == nodeLine || exportLine+1 == nodeLine {
				return true
			}
		}
	}

	// 检查 ExportAssignments
	for _, exportAssign := range fileResult.ExportAssignments {
		if exportAssign.Node != nil {
			exportLine := a.calculateLineNumber(sourceFile, exportAssign.Node)
			if exportLine == nodeLine || exportLine+1 == nodeLine {
				return true
			}
		}
	}

	return false
}

// getExportTypeFromNode 使用 AST 分析从节点确定导出类型。
func (a *Analyzer) getExportTypeFromNode(node tsmorphgo.Node) ExportType {
	// 检查父节点是否为 ExportAssignment（表示默认导出）
	parent := node.GetParent()
	if parent.IsValid() {
		if parent.Kind == ast.KindExportAssignment {
			return ExportTypeDefault
		}
		if parent.IsExportDeclaration() {
			return ExportTypeNamed
		}
	}

	// 检查子节点中的 DefaultKeyword（用于内联 "export default" 语法）
	hasDefaultKeyword := false
	hasExportKeyword := false
	node.ForEachChild(func(child tsmorphgo.Node) bool {
		if child.Kind == ast.KindDefaultKeyword {
			hasDefaultKeyword = true
		}
		if child.Kind == ast.KindExportKeyword {
			hasExportKeyword = true
		}
		return false // 继续检查
	})

	if hasDefaultKeyword && hasExportKeyword {
		return ExportTypeDefault
	}
	if hasExportKeyword {
		return ExportTypeNamed
	}

	return ExportTypeNone
}

// =============================================================================
// 工具方法
// =============================================================================

// GetProject 返回与此分析器关联的项目。
func (a *Analyzer) GetProject() *tsmorphgo.Project {
	return a.project
}

// GetOptions 返回分析选项。
func (a *Analyzer) GetOptions() AnalysisOptions {
	return a.options
}

// ValidateFile 检查文件是否存在于项目中。
func (a *Analyzer) ValidateFile(filePath string) error {
	sourceFile := a.project.GetSourceFile(filePath)
	if sourceFile == nil {
		return fmt.Errorf("文件未找到: %s", filePath)
	}
	return nil
}
