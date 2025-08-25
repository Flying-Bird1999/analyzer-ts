// TypeScript 类型声明打包工具（TypeBundler），用于将多个 TS 文件中的类型声明（interface/type/class/enum）合并为一个 bundle 文件，并解决跨文件同名类型冲突、类型引用正确性等问题。其核心设计思路如下：

// 分离式处理架构

// 首先收集所有类型声明，检测同名冲突，确定每个类型的最终名称（如有冲突则加后缀）。
// 然后基于名称映射，一次性更新所有类型声明和引用，避免重复替换和错误引用。
// 智能冲突解决

// 第一个遇到的类型保持原名，后续冲突类型基于文件名生成唯一后缀（如 Package_index2）。
// 全局唯一性保证：所有类型最终名称唯一，避免命名污染。
// 上下文感知的引用更新

// 只替换类型引用，不替换类型声明。
// 优先引用本文件的同名类型，否则引用全局唯一版本。
// 精确的字符串替换机制

// 通过正则和上下文判断，区分类型声明和类型引用，避免误替换。
// 数据结构设计

// TypeDeclaration：封装类型声明的所有信息（文件路径、原名、最终名、源码）。
// TypeBundler：管理类型名映射、已用名、原始名等。

package ts_bundle

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// TypeBundler 类型打包器
type TypeBundler struct {
	// 存储原始名称到最终名称的映射
	FinalNameMap map[string]string `json:"finalNameMap"`
	// 存储已使用的名称，避免冲突
	UsedNames map[string]bool `json:"usedNames"`
	// 存储每个文件路径中的原始类型名称
	OriginalNames map[string]string `json:"originalNames"`
}

// TypeDeclaration 类型声明
type TypeDeclaration struct {
	FilePath     string `json:"filePath"`     // 文件路径
	TypeName     string `json:"typeName"`     // 类型名称
	OriginalName string `json:"originalName"` // 原始名称
	SourceCode   string `json:"sourceCode"`   // 源码
	FinalName    string `json:"finalName"`    // 最终确定的名称
}

func NewTypeBundler() *TypeBundler {
	return &TypeBundler{
		FinalNameMap:  make(map[string]string),
		UsedNames:     make(map[string]bool),
		OriginalNames: make(map[string]string),
	}
}

func (b *TypeBundler) Bundle(typeMap map[string]string) (string, error) {
	// 第一步：解析所有类型声明
	declarations := b.parseTypeMap(typeMap)

	// 第二步：检测冲突并生成最终名称
	b.resolveAllNameConflicts(declarations)

	// 第三步：一次性更新所有代码中的类型引用
	b.updateAllTypeReferences(declarations)

	// 第四步：生成最终的 bundle 内容
	return b.generateBundle(declarations), nil
}

// parseTypeMap 解析类型声明map为 TypeDeclaration 列表
func (b *TypeBundler) parseTypeMap(typeMap map[string]string) []*TypeDeclaration {
	var declarations []*TypeDeclaration

	for key, sourceCode := range typeMap {
		// 解析 key: filePath_typeName
		lastUnderscoreIndex := strings.LastIndex(key, "_")
		if lastUnderscoreIndex == -1 {
			continue
		}

		filePath := key[:lastUnderscoreIndex]
		typeName := key[lastUnderscoreIndex+1:]

		decl := &TypeDeclaration{
			FilePath:     filePath,
			TypeName:     typeName,
			OriginalName: typeName,
			SourceCode:   sourceCode,
		}

		declarations = append(declarations, decl)

		// 记录原始名称
		b.OriginalNames[key] = typeName
	}

	return declarations
}

// resolveAllNameConflicts 检测所有类型声明的同名冲突并生成最终名称
func (b *TypeBundler) resolveAllNameConflicts(declarations []*TypeDeclaration) {
	// 按类型名称分组
	typeGroups := make(map[string][]*TypeDeclaration)
	for _, decl := range declarations {
		typeGroups[decl.TypeName] = append(typeGroups[decl.TypeName], decl)
	}

	// 为每个类型组解决冲突
	for typeName, decls := range typeGroups {
		if len(decls) == 1 {
			// 没有冲突，使用原名
			finalName := typeName
			decls[0].FinalName = finalName
			b.FinalNameMap[b.getUniqueKey(decls[0])] = finalName
			b.UsedNames[finalName] = true
		} else {
			// 有冲突，需要重命名
			b.resolveConflictingGroup(typeName, decls)
		}
	}
}

// resolveConflictingGroup 处理同名类型冲突，生成唯一名称
func (b *TypeBundler) resolveConflictingGroup(baseName string, decls []*TypeDeclaration) {
	// 按文件路径排序，确保一致性
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FilePath < decls[j].FilePath
	})

	for i, decl := range decls {
		var finalName string

		if i == 0 && !b.UsedNames[baseName] {
			// 第一个保持原名（如果可能）
			finalName = baseName
		} else {
			// 生成唯一名称
			finalName = b.generateUniqueName(baseName, decl)
		}

		decl.FinalName = finalName
		b.FinalNameMap[b.getUniqueKey(decl)] = finalName
		b.UsedNames[finalName] = true
	}
}

// generateUniqueName 基于文件路径生成唯一后缀，确保类型名唯一
func (b *TypeBundler) generateUniqueName(baseName string, decl *TypeDeclaration) string {
	// 基于文件路径生成后缀
	pathParts := strings.Split(decl.FilePath, "/")
	var suffix string

	if len(pathParts) > 0 {
		fileName := pathParts[len(pathParts)-1]
		// 清理文件名，移除扩展名和特殊字符
		fileName = strings.TrimSuffix(fileName, ".ts")
		fileName = strings.TrimSuffix(fileName, ".d")
		fileName = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(fileName, "_")
		suffix = fileName
	}

	if suffix == "" {
		suffix = "1"
	}

	candidateName := fmt.Sprintf("%s_%s", baseName, suffix)

	// 确保名称唯一
	counter := 1
	finalName := candidateName
	for b.UsedNames[finalName] {
		finalName = fmt.Sprintf("%s_%d", candidateName, counter)
		counter++
	}

	return finalName
}

// getUniqueKey 生成唯一key（文件路径+类型名）
func (b *TypeBundler) getUniqueKey(decl *TypeDeclaration) string {
	return fmt.Sprintf("%s:%s", decl.FilePath, decl.OriginalName)
}

// generateBundle 生成最终 bundle 文件内容（带来源注释）
func (b *TypeBundler) generateBundle(declarations []*TypeDeclaration) string {
	var result strings.Builder

	b.writeDeclarations(&result, declarations)
	return result.String()
}

// writeDeclarations 写入所有类型声明到 bundle
func (b *TypeBundler) writeDeclarations(result *strings.Builder, decls []*TypeDeclaration) {
	if len(decls) == 0 {
		return
	}

	// 按最终名称排序
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FinalName < decls[j].FinalName
	})

	for _, decl := range decls {
		// 1. 清理源码前后的空白，保证格式一致
		trimmedSource := strings.TrimSpace(decl.SourceCode)

		// 2. 写入清理后的代码
		result.WriteString(trimmedSource)

		// 3. 写入两个换行符，确保条目之间有一个空行
		result.WriteString("\n\n")
	}
}

// updateAllTypeReferences 更新所有类型声明中的类型引用
func (b *TypeBundler) updateAllTypeReferences(declarations []*TypeDeclaration) {
	// 为每个声明单独处理类型引用更新
	for _, decl := range declarations {
		decl.SourceCode = b.updateSingleDeclarationReferences(decl, declarations)
	}
}

// updateSingleDeclarationReferences 更新单个类型声明中的类型引用
func (b *TypeBundler) updateSingleDeclarationReferences(currentDecl *TypeDeclaration, allDeclarations []*TypeDeclaration) string {
	updatedCode := currentDecl.SourceCode

	// 第一步：更新当前类型的声明名称
	if currentDecl.FinalName != currentDecl.OriginalName {
		declarationPatterns := []string{
			fmt.Sprintf(`\binterface\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\btype\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\bclass\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\benum\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\bexport\s+interface\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\bexport\s+type\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\bexport\s+class\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
			fmt.Sprintf(`\bexport\s+enum\s+%s\b`, regexp.QuoteMeta(currentDecl.OriginalName)),
		}

		declarationReplacements := []string{
			fmt.Sprintf("interface %s", currentDecl.FinalName),
			fmt.Sprintf("type %s", currentDecl.FinalName),
			fmt.Sprintf("class %s", currentDecl.FinalName),
			fmt.Sprintf("enum %s", currentDecl.FinalName),
			fmt.Sprintf("export interface %s", currentDecl.FinalName),
			fmt.Sprintf("export type %s", currentDecl.FinalName),
			fmt.Sprintf("export class %s", currentDecl.FinalName),
			fmt.Sprintf("export enum %s", currentDecl.FinalName),
		}

		for i, pattern := range declarationPatterns {
			re := regexp.MustCompile(pattern)
			updatedCode = re.ReplaceAllString(updatedCode, declarationReplacements[i])
		}
	}

	// 第二步：更新对其他类型的引用
	for _, otherDecl := range allDeclarations {
		if otherDecl == currentDecl {
			continue
		}

		shouldReplace := b.shouldReplaceReference(currentDecl, otherDecl, allDeclarations)
		if shouldReplace && otherDecl.FinalName != otherDecl.OriginalName {
			// 使用简单的单词边界匹配，然后手动过滤声明语句
			pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(otherDecl.OriginalName))
			re := regexp.MustCompile(pattern)

			// 找到所有匹配位置，然后检查上下文
			updatedCode = b.replaceTypeReferencesOnly(updatedCode, re, otherDecl.OriginalName, otherDecl.FinalName)
		}
	}

	return updatedCode
}

// replaceTypeReferencesOnly 只替换类型引用，不替换声明
func (b *TypeBundler) replaceTypeReferencesOnly(text string, re *regexp.Regexp, oldName, newName string) string {
	// 找到所有匹配的位置
	matches := re.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}

	// 从后往前替换，避免索引偏移问题
	result := text
	for i := len(matches) - 1; i >= 0; i-- {
		start, end := matches[i][0], matches[i][1]

		// 检查这个匹配是否是类型声明（而不是引用）
		if b.isTypeDeclaration(result, start, end, oldName) {
			continue // 跳过声明，不替换
		}

		// 替换引用
		result = result[:start] + newName + result[end:]
	}

	return result
}

// isTypeDeclaration 判断匹配位置是否为类型声明（而不是引用）
func (b *TypeBundler) isTypeDeclaration(text string, start, end int, typeName string) bool {
	// 获取匹配位置前的文本，检查是否是声明关键字
	beforeMatch := ""
	if start > 0 {
		// 取前面最多50个字符来检查上下文
		contextStart := start - 50
		if contextStart < 0 {
			contextStart = 0
		}
		beforeMatch = text[contextStart:start]
	}

	// 检查是否包含声明关键字
	declarationKeywords := []string{
		"interface ", "type ", "class ", "enum ",
		"export interface ", "export type ", "export class ", "export enum ",
	}

	for _, keyword := range declarationKeywords {
		if strings.Contains(beforeMatch, keyword) {
			// 进一步检查是否紧接着类型名
			keywordIndex := strings.LastIndex(beforeMatch, keyword)
			if keywordIndex >= 0 {
				afterKeyword := beforeMatch[keywordIndex+len(keyword):]
				afterKeyword = strings.TrimSpace(afterKeyword)
				if afterKeyword == "" {
					// 关键字后直接是类型名，这是声明
					return true
				}
			}
		}
	}

	return false
}

// shouldReplaceReference 判断是否需要替换类型引用（上下文感知）
func (b *TypeBundler) shouldReplaceReference(currentDecl, referencedDecl *TypeDeclaration, allDeclarations []*TypeDeclaration) bool {
	// 如果被引用的类型没有重命名，不需要替换
	if referencedDecl.FinalName == referencedDecl.OriginalName {
		return false
	}

	// 查找所有同名的类型声明
	sameNameDecls := []*TypeDeclaration{}
	for _, decl := range allDeclarations {
		if decl.OriginalName == referencedDecl.OriginalName {
			sameNameDecls = append(sameNameDecls, decl)
		}
	}

	// 如果只有一个同名类型，不需要特殊处理
	if len(sameNameDecls) <= 1 {
		return false
	}

	// 确定当前文件应该引用哪个版本
	// 规则：优先引用同一文件中的类型，否则引用第一个（保持原名的）版本
	for _, sameNameDecl := range sameNameDecls {
		if sameNameDecl.FilePath == currentDecl.FilePath {
			// 如果当前文件中有同名类型，引用当前文件的版本
			return sameNameDecl == referencedDecl
		}
	}

	// 如果当前文件中没有同名类型，引用保持原名的版本（通常是第一个）
	for _, sameNameDecl := range sameNameDecls {
		if sameNameDecl.FinalName == sameNameDecl.OriginalName {
			return false // 引用原名版本，不需要替换
		}
	}

	// 如果所有版本都被重命名了，引用第一个版本
	sort.Slice(sameNameDecls, func(i, j int) bool {
		return sameNameDecls[i].FilePath < sameNameDecls[j].FilePath
	})

	return sameNameDecls[0] == referencedDecl
}
