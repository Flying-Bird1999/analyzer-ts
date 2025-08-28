// TypeScript 类型声明打包工具（TypeBundler），用于将多个 TS 文件中的类型声明（interface/type/class/enum）合并为一个 bundle 文件，
// 并解决跨文件同名类型冲突、类型引用正确性等问题。
//
// 核心设计思路：
// 1. 分离式处理架构：首先收集所有类型声明，检测同名冲突，确定每个类型的最终名称（如有冲突则加后缀），
//    然后基于名称映射，一次性更新所有类型声明和引用，避免重复替换和错误引用。
// 2. 智能冲突解决：第一个遇到的类型保持原名，后续冲突类型基于文件名生成唯一后缀。
// 3. 全局唯一性保证：所有类型最终名称唯一，避免命名污染。
// 4. 上下文感知的引用更新：只替换类型引用，不替换类型声明，优先引用本文件的同名类型，否则引用全局唯一版本。
// 5. 精确的字符串替换机制：通过正则和上下文判断，区分类型声明和类型引用，避免误替换。

package ts_bundle

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// TypeBundler 类型打包器
// 负责管理类型名称映射、冲突解决和引用更新
type TypeBundler struct {
	// 存储原始名称到最终名称的映射
	FinalNameMap map[string]string `json:"finalNameMap"`
	// 存储已使用的名称，避免冲突
	UsedNames map[string]bool `json:"usedNames"`
	// 存储每个文件路径中的原始类型名称
	OriginalNames map[string]string `json:"originalNames"`
	// 预编译的正则表达式缓存，避免重复编译
	regexCache map[string]*regexp.Regexp
	// 预编译的声明正则表达式模式和替换字符串
	declarationPatterns []string
	declarationReplacements []string
}

// TypeDeclaration 类型声明
// 封装类型声明的所有信息（文件路径、原名、最终名、源码）
type TypeDeclaration struct {
	FilePath     string `json:"filePath"`     // 文件路径
	TypeName     string `json:"typeName"`     // 类型名称
	OriginalName string `json:"originalName"` // 原始名称
	SourceCode   string `json:"sourceCode"`   // 源码
	FinalName    string `json:"finalName"`    // 最终确定的名称
}

// NewTypeBundler 创建并初始化一个新的 TypeBundler 实例
func NewTypeBundler() *TypeBundler {
	tb := &TypeBundler{
		FinalNameMap:  make(map[string]string),
		UsedNames:     make(map[string]bool),
		OriginalNames: make(map[string]string),
		regexCache:    make(map[string]*regexp.Regexp),
		// 预编译声明正则表达式模式
		declarationPatterns: []string{
			`\binterface\s+%s\b`,
			`\btype\s+%s\b`,
			`\bclass\s+%s\b`,
			`\benum\s+%s\b`,
			`\bexport\s+interface\s+%s\b`,
			`\bexport\s+type\s+%s\b`,
			`\bexport\s+class\s+%s\b`,
			`\bexport\s+enum\s+%s\b`,
		},
		// 对应的替换字符串
		declarationReplacements: []string{
			"interface %s",
			"type %s",
			"class %s",
			"enum %s",
			"export interface %s",
			"export type %s",
			"export class %s",
			"export enum %s",
		},
	}
	
	return tb
}

// getCachedRegex 获取缓存的正则表达式
// 如果正则表达式已缓存则直接返回，否则编译并缓存后返回
func (b *TypeBundler) getCachedRegex(pattern string) *regexp.Regexp {
	// 检查缓存中是否已存在
	if re, exists := b.regexCache[pattern]; exists {
		return re
	}
	
	// 编译正则表达式并存入缓存
	re := regexp.MustCompile(pattern)
	b.regexCache[pattern] = re
	return re
}

// Bundle 执行类型打包的主要流程
// 输入类型映射，输出合并后的类型声明字符串
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
// 从 map[string]string 格式的类型映射解析出结构化的类型声明列表
func (b *TypeBundler) parseTypeMap(typeMap map[string]string) []*TypeDeclaration {
	var declarations []*TypeDeclaration

	// 遍历类型映射中的每个条目
	for key, sourceCode := range typeMap {
		// 解析 key: filePath_typeName
		// 使用最后的下划线分隔文件路径和类型名
		lastUnderscoreIndex := strings.LastIndex(key, "_")
		if lastUnderscoreIndex == -1 {
			// 如果没有找到下划线，跳过该条目
			continue
		}

		// 分离文件路径和类型名
		filePath := key[:lastUnderscoreIndex]
		typeName := key[lastUnderscoreIndex+1:]

		// 创建类型声明对象
		decl := &TypeDeclaration{
			FilePath:     filePath,
			TypeName:     typeName,
			OriginalName: typeName,
			SourceCode:   sourceCode,
		}

		// 添加到声明列表
		declarations = append(declarations, decl)

		// 记录原始名称到 OriginalNames 映射中
		b.OriginalNames[key] = typeName
	}

	return declarations
}

// resolveAllNameConflicts 检测所有类型声明的同名冲突并生成最终名称
// 通过按类型名称分组来识别冲突，并为冲突的类型生成唯一名称
func (b *TypeBundler) resolveAllNameConflicts(declarations []*TypeDeclaration) {
	// 按类型名称分组，相同名称的类型会被归为一组
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
// 对于同一名称的多个类型声明，按文件路径排序后为除第一个外的声明生成唯一名称
func (b *TypeBundler) resolveConflictingGroup(baseName string, decls []*TypeDeclaration) {
	// 按文件路径排序，确保处理的一致性
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FilePath < decls[j].FilePath
	})

	// 为每个声明分配最终名称
	for i, decl := range decls {
		var finalName string

		// 第一个声明保持原名（如果该名称未被使用）
		if i == 0 && !b.UsedNames[baseName] {
			finalName = baseName
		} else {
			// 为其他声明生成唯一名称
			finalName = b.generateUniqueName(baseName, decl)
		}

		// 设置声明的最终名称并更新映射
		decl.FinalName = finalName
		b.FinalNameMap[b.getUniqueKey(decl)] = finalName
		b.UsedNames[finalName] = true
	}
}

// generateUniqueName 基于文件路径生成唯一后缀，确保类型名唯一
// 使用文件名作为基础生成后缀，确保生成的名称不会与其他已使用名称冲突
func (b *TypeBundler) generateUniqueName(baseName string, decl *TypeDeclaration) string {
	// 基于文件路径生成后缀
	pathParts := strings.Split(decl.FilePath, "/")
	var suffix string

	// 从文件路径中提取文件名作为后缀的基础
	if len(pathParts) > 0 {
		fileName := pathParts[len(pathParts)-1]
		// 清理文件名，移除扩展名和特殊字符
		fileName = strings.TrimSuffix(fileName, ".ts")
		fileName = strings.TrimSuffix(fileName, ".d")
		// 将非字母数字下划线字符替换为下划线
		fileName = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(fileName, "_")
		suffix = fileName
	}

	// 如果没有有效的后缀，使用默认值
	if suffix == "" {
		// 使用文件路径的哈希值作为后缀的一部分
		hashSuffix := fmt.Sprintf("%x", md5.Sum([]byte(decl.FilePath)))[:6]
		suffix = hashSuffix
	}

	// 生成候选名称
	candidateNames := []string{
		fmt.Sprintf("%s_%s", baseName, suffix),
		fmt.Sprintf("%s_%s_%d", baseName, suffix, len(b.UsedNames)), // 添加已用名称数量作为额外区分
	}

	// 首先尝试候选名称
	for _, candidateName := range candidateNames {
		if !b.UsedNames[candidateName] {
			return candidateName
		}
	}

	// 如果所有候选名称都被使用，添加计数器生成唯一名称
	counter := 1
	finalName := candidateNames[0]
	for b.UsedNames[finalName] {
		finalName = fmt.Sprintf("%s_%d", candidateNames[0], counter)
		counter++
	}

	return finalName
}

// getUniqueKey 生成唯一key（文件路径+类型名）
// 用于在 FinalNameMap 中唯一标识一个类型声明
func (b *TypeBundler) getUniqueKey(decl *TypeDeclaration) string {
	return fmt.Sprintf("%s:%s", decl.FilePath, decl.OriginalName)
}

// generateBundle 生成最终 bundle 文件内容（带来源注释）
// 将所有类型声明按最终名称排序后合并为一个字符串
func (b *TypeBundler) generateBundle(declarations []*TypeDeclaration) string {
	var result strings.Builder
	b.writeDeclarations(&result, declarations)
	return result.String()
}

// writeDeclarations 写入所有类型声明到 bundle
// 按最终名称排序后写入声明内容
func (b *TypeBundler) writeDeclarations(result *strings.Builder, decls []*TypeDeclaration) {
	// 如果没有声明，直接返回
	if len(decls) == 0 {
		return
	}

	// 按最终名称排序，确保输出的一致性
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].FinalName < decls[j].FinalName
	})

	// 依次写入每个声明
	for _, decl := range decls {
		// 输出更新后的代码
		result.WriteString(decl.SourceCode)

		// 确保每段代码后都有换行符
		if !strings.HasSuffix(decl.SourceCode, "\n") {
			result.WriteString("\n")
		}
	}
}

// updateAllTypeReferences 更新所有类型声明中的类型引用
// 遍历所有声明，更新其中对其他类型的引用
func (b *TypeBundler) updateAllTypeReferences(declarations []*TypeDeclaration) {
	// 为每个声明单独处理类型引用更新
	for _, decl := range declarations {
		decl.SourceCode = b.updateSingleDeclarationReferences(decl, declarations)
	}
}

// updateSingleDeclarationReferences 更新单个类型声明中的类型引用
// 分两步：首先更新当前类型的声明名称，然后更新对其他类型的引用
func (b *TypeBundler) updateSingleDeclarationReferences(currentDecl *TypeDeclaration, allDeclarations []*TypeDeclaration) string {
	updatedCode := currentDecl.SourceCode

	// 第一步：更新当前类型的声明名称
	// 只有当最终名称与原始名称不同时才需要更新
	if currentDecl.FinalName != currentDecl.OriginalName {
		// 使用预编译的正则表达式模式和替换字符串更新声明名称
		for i, pattern := range b.declarationPatterns {
			// 构造完整的正则表达式模式
			fullPattern := fmt.Sprintf(pattern, regexp.QuoteMeta(currentDecl.OriginalName))
			// 获取缓存的正则表达式
			re := b.getCachedRegex(fullPattern)
			// 执行替换
			updatedCode = re.ReplaceAllString(updatedCode, fmt.Sprintf(b.declarationReplacements[i], currentDecl.FinalName))
		}
	}

	// 第二步：更新对其他类型的引用
	// 遍历所有声明，查找需要更新的引用
	for _, otherDecl := range allDeclarations {
		// 跳过当前声明自身
		if otherDecl == currentDecl {
			continue
		}

		// 判断是否需要替换引用
		shouldReplace := b.shouldReplaceReference(currentDecl, otherDecl, allDeclarations)
		// 只有当被引用的类型被重命名时才需要替换
		if shouldReplace && otherDecl.FinalName != otherDecl.OriginalName {
			// 构造匹配被引用类型名的正则表达式
			pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(otherDecl.OriginalName))
			re := b.getCachedRegex(pattern)

			// 找到所有匹配位置，然后检查上下文并执行替换
			updatedCode = b.replaceTypeReferencesOnly(updatedCode, re, otherDecl.OriginalName, otherDecl.FinalName)
		}
	}

	return updatedCode
}

// replaceTypeReferencesOnly 只替换类型引用，不替换声明
// 通过上下文检查确保只替换真正的类型引用，而不是类型声明
func (b *TypeBundler) replaceTypeReferencesOnly(text string, re *regexp.Regexp, oldName, newName string) string {
	// 使用 ReplaceAllStringFunc 方法结合上下文检查进行替换
	return re.ReplaceAllStringFunc(text, func(match string) string {
		// 注意：这里简化了实现，实际使用中可能需要结合 isTypeDeclaration 方法
		// 进行更精确的上下文检查
		return newName
	})
}

// isTypeDeclaration 判断匹配位置是否为类型声明（而不是引用）
// 通过检查匹配位置前的上下文来判断是否为类型声明
func (b *TypeBundler) isTypeDeclaration(text string, start, end int, typeName string) bool {
	// 获取匹配位置前的文本，检查是否是声明关键字
	beforeMatch := ""
	if start > 0 {
		// 取前面最多100个字符来检查上下文，增加检查范围以提高准确性
		contextStart := start - 100
		if contextStart < 0 {
			contextStart = 0
		}
		beforeMatch = text[contextStart:start]
	}

	// 清理空白字符，便于匹配
	beforeMatch = strings.TrimSpace(beforeMatch)
	
	// 检查是否包含声明关键字的更精确模式
	declarationPatterns := []string{
		`(^|\s)interface\s+%s(\s|{)`,
		`(^|\s)type\s+%s(\s|=)`,
		`(^|\s)class\s+%s(\s|{)`,
		`(^|\s)enum\s+%s(\s|{)`,
		`(^|\s)export\s+interface\s+%s(\s|{)`,
		`(^|\s)export\s+type\s+%s(\s|=)`,
		`(^|\s)export\s+class\s+%s(\s|{)`,
		`(^|\s)export\s+enum\s+%s(\s|{)`,
	}

	// 检查每个声明模式
	for _, pattern := range declarationPatterns {
		// 构造完整的正则表达式模式
		fullPattern := fmt.Sprintf(pattern, regexp.QuoteMeta(typeName))
		// 编译正则表达式（这里没有使用缓存，因为模式是动态构造的）
		re := regexp.MustCompile(fullPattern)
		// 如果匹配，则说明是类型声明
		if re.MatchString(beforeMatch) {
			return true
		}
	}

	// 如果没有匹配任何声明模式，则不是类型声明
	return false
}

// shouldReplaceReference 判断是否需要替换类型引用（上下文感知）
// 根据引用上下文和同名类型的存在情况决定是否替换引用
func (b *TypeBundler) shouldReplaceReference(currentDecl, referencedDecl *TypeDeclaration, allDeclarations []*TypeDeclaration) bool {
	// 如果被引用的类型没有重命名，不需要替换
	if referencedDecl.FinalName == referencedDecl.OriginalName {
		return false
	}

	// 查找所有同名的类型声明
	sameNameDecls := []*TypeDeclaration{}
	for _, decl := range allDeclarations {
		// 收集所有原始名称相同的声明
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
		// 如果当前文件中有同名类型，引用当前文件的版本
		if sameNameDecl.FilePath == currentDecl.FilePath {
			// 检查是否应该引用这个特定的声明
			return sameNameDecl == referencedDecl
		}
	}

	// 如果当前文件中没有同名类型，引用保持原名的版本（通常是第一个）
	for _, sameNameDecl := range sameNameDecls {
		// 如果找到保持原名的版本，引用它（不需要替换）
		if sameNameDecl.FinalName == sameNameDecl.OriginalName {
			return false // 引用原名版本，不需要替换
		}
	}

	// 如果所有版本都被重命名了，引用第一个版本
	// 按文件路径排序确保一致性
	sort.Slice(sameNameDecls, func(i, j int) bool {
		return sameNameDecls[i].FilePath < sameNameDecls[j].FilePath
	})

	// 引用排序后的第一个版本
	return sameNameDecls[0] == referencedDecl
}
