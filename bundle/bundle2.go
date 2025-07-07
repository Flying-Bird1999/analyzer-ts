package bundle

// 核心设计思路
// 这个方案的设计围绕着一个关键挑战：如何在保持类型引用正确性的前提下，解决来自不同文件的同名类型冲突。

// 1. 分离式处理架构
// 方案采用了分离式处理的核心思想，将名称解析和内容更新完全分开：

// 第一阶段：收集所有类型声明，检测冲突，确定最终名称
// 第二阶段：基于确定的名称映射，一次性更新所有代码内容
// 这种设计避免了增量式替换可能导致的重复替换问题（如 Package_index2_index2_index2）。

// 2. 智能冲突解决策略
// 借鉴 TypeScript-Go 的名称生成器设计，方案实现了智能的冲突解决：

// 优先级规则：第一个遇到的类型保持原名（如果没有其他冲突）
// 文件路径后缀：基于文件路径生成唯一标识符（如 Package_index2）
// 全局唯一性：确保生成的名称在整个 bundle 中唯一
// 3. 上下文感知的引用更新
// 方案的关键创新在于上下文感知的引用替换：

// func (b *TypeBundler) shouldReplaceReference(currentDecl, referencedDecl *TypeDeclaration, allDeclarations []*TypeDeclaration) bool
// 这个函数实现了智能的引用解析逻辑：

// 同文件优先：如果当前文件中有同名类型，优先引用本文件的版本
// 原名保持：如果没有本地同名类型，引用保持原名的版本
// 避免错误替换：防止第一个文件中的 Package 引用被错误替换为 Package_index2
// 4. 精确的字符串替换机制
// 由于 Go 的 regexp 包不支持负向前瞻，方案采用了两步替换策略：

// 声明替换：只替换类型声明语句（如 interface Package → interface Package_index2）
// 引用替换：通过上下文检查，只替换类型引用，避免重复替换声明
// func (b *TypeBundler) isTypeDeclaration(text string, start, end int, typeName string) bool
// 这个函数通过检查匹配位置前的关键字来判断是否为类型声明。

// 5. 数据结构设计
// 方案使用了清晰的数据结构来管理复杂的映射关系：

// finalNameMap：存储原始名称到最终名称的映射
// usedNames：跟踪已使用的名称，避免冲突
// TypeDeclaration：封装类型声明的所有信息，包括原始名称和最终名称
// 设计优势
// 可预测性：分离式处理确保了结果的一致性和可预测性
// 扩展性：基于文件路径的命名策略可以处理任意数量的同名冲突
// 正确性：上下文感知的引用解析确保了类型引用的语义正确性
// 性能：一次性替换避免了多次遍历和重复处理
// 这个方案的设计充分考虑了 TypeScript 类型系统的复杂性，通过借鉴 TypeScript-Go 编译器的成熟设计模式，实现了一个既健壮又高效的类型依赖打包解决方案。

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type TypeBundler struct {
	// 存储原始名称到最终名称的映射
	finalNameMap map[string]string
	// 存储已使用的名称，避免冲突
	usedNames map[string]bool
	// 存储每个文件路径中的原始类型名称
	originalNames map[string]string
}

type TypeDeclaration struct {
	FilePath     string
	TypeName     string
	OriginalName string
	SourceCode   string
	Kind         string
	FinalName    string // 最终确定的名称
}

func NewTypeBundler() *TypeBundler {
	return &TypeBundler{
		finalNameMap:  make(map[string]string),
		usedNames:     make(map[string]bool),
		originalNames: make(map[string]string),
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
			Kind:         b.detectTypeKind(sourceCode),
		}

		declarations = append(declarations, decl)

		// 记录原始名称
		b.originalNames[key] = typeName
	}

	return declarations
}

func (b *TypeBundler) detectTypeKind(sourceCode string) string {
	trimmed := strings.TrimSpace(sourceCode)

	patterns := map[string]string{
		`^\s*export\s+interface\s+`: "interface",
		`^\s*interface\s+`:          "interface",
		`^\s*export\s+type\s+`:      "type",
		`^\s*type\s+`:               "type",
		`^\s*export\s+class\s+`:     "class",
		`^\s*class\s+`:              "class",
		`^\s*export\s+enum\s+`:      "enum",
		`^\s*enum\s+`:               "enum",
	}

	for pattern, kind := range patterns {
		if matched, _ := regexp.MatchString(pattern, trimmed); matched {
			return kind
		}
	}

	return "unknown"
}

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
			b.finalNameMap[b.getUniqueKey(decls[0])] = finalName
			b.usedNames[finalName] = true
		} else {
			// 有冲突，需要重命名
			b.resolveConflictingGroup(typeName, decls)
		}
	}
}

func (b *TypeBundler) resolveConflictingGroup(baseName string, decls []*TypeDeclaration) {
	// 按文件路径排序，确保一致性
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FilePath < decls[j].FilePath
	})

	for i, decl := range decls {
		var finalName string

		if i == 0 && !b.usedNames[baseName] {
			// 第一个保持原名（如果可能）
			finalName = baseName
		} else {
			// 生成唯一名称
			finalName = b.generateUniqueName(baseName, decl)
		}

		decl.FinalName = finalName
		b.finalNameMap[b.getUniqueKey(decl)] = finalName
		b.usedNames[finalName] = true
	}
}

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
	for b.usedNames[finalName] {
		finalName = fmt.Sprintf("%s_%d", candidateName, counter)
		counter++
	}

	return finalName
}

func (b *TypeBundler) getUniqueKey(decl *TypeDeclaration) string {
	return fmt.Sprintf("%s:%s", decl.FilePath, decl.OriginalName)
}

func (b *TypeBundler) updateSingleDeclaration(sourceCode, originalName, finalName string, replacementMap map[string]string) string {
	updatedCode := sourceCode

	// 第一步：更新当前类型的声明
	if finalName != originalName {
		declarationPatterns := []string{
			fmt.Sprintf(`\binterface\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\btype\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\bclass\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\benum\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\bexport\s+interface\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\bexport\s+type\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\bexport\s+class\s+%s\b`, regexp.QuoteMeta(originalName)),
			fmt.Sprintf(`\bexport\s+enum\s+%s\b`, regexp.QuoteMeta(originalName)),
		}

		declarationReplacements := []string{
			fmt.Sprintf("interface %s", finalName),
			fmt.Sprintf("type %s", finalName),
			fmt.Sprintf("class %s", finalName),
			fmt.Sprintf("enum %s", finalName),
			fmt.Sprintf("export interface %s", finalName),
			fmt.Sprintf("export type %s", finalName),
			fmt.Sprintf("export class %s", finalName),
			fmt.Sprintf("export enum %s", finalName),
		}

		for i, pattern := range declarationPatterns {
			re := regexp.MustCompile(pattern)
			updatedCode = re.ReplaceAllString(updatedCode, declarationReplacements[i])
		}
	}

	// 第二步：更新其他类型的引用（排除当前类型）
	for originalRef, finalRef := range replacementMap {
		if originalRef != originalName {
			// 使用单词边界确保精确匹配
			pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(originalRef))
			re := regexp.MustCompile(pattern)
			updatedCode = re.ReplaceAllString(updatedCode, finalRef)
		}
	}

	return updatedCode
}

func (b *TypeBundler) generateBundle(declarations []*TypeDeclaration) string {
	var result strings.Builder

	// 文件头
	result.WriteString("// Auto-generated bundle file\n")
	result.WriteString("// This file contains bundled type declarations\n\n")

	// 按类型分组
	var interfaces, types, classes, enums, unknowns []*TypeDeclaration

	for _, decl := range declarations {
		switch decl.Kind {
		case "interface":
			interfaces = append(interfaces, decl)
		case "type":
			types = append(types, decl)
		case "class":
			classes = append(classes, decl)
		case "enum":
			enums = append(enums, decl)
		default:
			unknowns = append(unknowns, decl)
		}
	}

	// 按顺序输出
	b.writeDeclarations(&result, "// Enums\n", enums)
	b.writeDeclarations(&result, "// Interfaces\n", interfaces)
	b.writeDeclarations(&result, "// Type Aliases\n", types)
	b.writeDeclarations(&result, "// Classes\n", classes)
	b.writeDeclarations(&result, "// Other Declarations\n", unknowns)

	return result.String()
}

func (b *TypeBundler) writeDeclarations(result *strings.Builder, header string, decls []*TypeDeclaration) {
	if len(decls) == 0 {
		return
	}

	result.WriteString(header)

	// 按最终名称排序
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FinalName < decls[j].FinalName
	})

	for _, decl := range decls {
		// 添加来源注释
		result.WriteString(fmt.Sprintf("// From: %s (original: %s)\n", decl.FilePath, decl.OriginalName))

		// 输出更新后的代码
		result.WriteString(decl.SourceCode)

		if !strings.HasSuffix(decl.SourceCode, "\n") {
			result.WriteString("\n")
		}
		result.WriteString("\n")
	}
}

func (b *TypeBundler) updateAllTypeReferences(declarations []*TypeDeclaration) {
	// 为每个声明单独处理类型引用更新
	for _, decl := range declarations {
		decl.SourceCode = b.updateSingleDeclarationReferences(decl, declarations)
	}
}

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

// 辅助函数：只替换类型引用，不替换声明
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

// 检查匹配位置是否是类型声明
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

// 使用示例
func Bundle2(typeMap map[string]string) {
	bundler := NewTypeBundler()
	bundledContent, err := bundler.Bundle(typeMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bundle error: %v\n", err)
		return
	}

	// 输出到文件
	outputFile := "./ts/output/result.ts"
	err = os.WriteFile(outputFile, []byte(bundledContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		return
	}

	fmt.Printf("Bundle completed: %s\n", outputFile)
	fmt.Printf("\nName mappings:\n")
	for key, finalName := range bundler.finalNameMap {
		fmt.Printf("  %s -> %s\n", key, finalName)
	}
}
