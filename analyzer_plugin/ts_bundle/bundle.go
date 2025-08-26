package ts_bundle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// GenericDeclaration 是一个通用的结构体，用于存放任何类型声明（如 interface, type, enum）的关键信息。
// 它是在 `collect.go` 中收集的信息的基础上进行打包处理的中间表示。
type GenericDeclaration struct {
	Name      string            // 原始声明名称
	Raw       string            // 原始声明的源代码文本
	Reference map[string]bool   // 此声明内部引用的其他类型名称集合
	FilePath  string            // 声明所在的文件路径
}

// TypeBundler 是实现 TypeScript 类型打包功能的核心结构体。
// 它持有所需的所有状态，并提供打包方法。
type TypeBundler struct {
	declarations      map[UniqueTypeID]GenericDeclaration // 从收集器传入的所有类型声明
	resolutionMaps    map[string]FileScopeResolutionMap   // 每个文件的作用域解析图
	finalNameMap      map[UniqueTypeID]string             // 存储每个类型最终在打包文件中使用的名称
	usedNames         map[string]bool                     // 用于跟踪已经使用过的名称，以避免冲突
	defaultExportUIDs map[UniqueTypeID]bool             // 记录哪些类型是默认导出，用于处理别名
}

// safeReplace 使用正则表达式安全地替换字符串中的标识符。
// `\b` 边界匹配符确保只替换完整的单词，避免将一个名称作为另一个更长名称的子串进行替换。
// 例如，确保 `safeReplace("type T = T1", "T", "NewT")` 得到 `"type NewT = T1"` 而不是 `"type NewT = NewT1"`。
func safeReplace(source, oldName, newName string) string {
	if oldName == "" || newName == "" || oldName == newName {
		return source
	}
	// 使用 `\b` 来匹配单词边界，确保只替换独立的标识符
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)
	return re.ReplaceAllString(source, newName)
}

// NewTypeBundler 创建一个新的 TypeBundler 实例。
// 它接收 `CollectResult` 作为输入，这是前一阶段（依赖收集）的产物。
func NewTypeBundler(result *CollectResult) *TypeBundler {
	return &TypeBundler{
		declarations:      result.Declarations,
		resolutionMaps:    result.ResolutionMaps,
		finalNameMap:      make(map[UniqueTypeID]string),
		usedNames:         make(map[string]bool),
		defaultExportUIDs: make(map[UniqueTypeID]bool),
	}
}

// Bundle 是执行打包过程的主函数。
// 它按顺序执行名称冲突解决、引用更新和最终代码生成。
func (b *TypeBundler) Bundle(entryFile string, entryType string) (string, error) {
	// 1. 识别出所有的默认导出
	b.findDefaultExports()
	// 2. 解决所有潜在的名称冲突
	b.resolveNameConflicts()
	// 3. 更新所有声明中的类型引用，确保它们指向正确的、重命名后的类型
	b.updateAllTypeReferences()
	// 4. 生成最终的打包输出字符串
	return b.generateBundleOutput(), nil
}

// findDefaultExports 遍历所有文件的解析图，找出被 `export default` 的类型，
// 并将它们的 UniqueTypeID 记录下来。这对于后续处理导入别名很重要。
func (b *TypeBundler) findDefaultExports() {
	for _, resMap := range b.resolutionMaps {
		if uid, ok := resMap["default"]; ok {
			b.defaultExportUIDs[uid] = true
		}
	}
}

// resolveNameConflicts 是打包过程中最关键和复杂的步骤之一。
// 它的目标是为每个类型分配一个在最终打包文件中唯一的名称。
func (b *TypeBundler) resolveNameConflicts() {
	// 策略 1: 初始名称。将每个类型的最终名称初步设置为其原始声明名称。
	for uid, decl := range b.declarations {
		b.finalNameMap[uid] = decl.Name
	}

	// 策略 2: 别名优先。如果一个类型在某个文件中被用别名导入或导出，
	// 那么这个别名通常是用户更希望看到的名称，因此给予它更高的优先级。
	// 但要排除 `import MyType from ...` 这种情况，因为 `MyType` 是本地的，不应作为全局优先名。
	uidToPreferredName := make(map[UniqueTypeID]string)
	for _, resolutionMap := range b.resolutionMaps {
		for nameInFile, uid := range resolutionMap {
			if decl, ok := b.declarations[uid]; ok {
				// 如果文件中的名称与原始名称不同，并且它不是一个默认导入的本地别名，则认为它是一个优先的别名。
				isDefaultImportAlias := b.defaultExportUIDs[uid]
				if nameInFile != decl.Name && !isDefaultImportAlias {
					uidToPreferredName[uid] = nameInFile
				}
			}
		}
	}
	// 应用这些找到的优先名称。
	for uid, preferredName := range uidToPreferredName {
		b.finalNameMap[uid] = preferredName
	}

	// 策略 3: 命名空间。处理 `import * as ns from ...` 的情况。
	// 导入的类型会以 `ns.TypeName` 的形式存在，我们将 `.` 替换为 `_` 来创建一个合法的标识符，
	// 例如 `ns_TypeName`。这具有很高的优先级。
	for _, resolutionMap := range b.resolutionMaps {
		for nameInFile, uid := range resolutionMap {
			if strings.Contains(nameInFile, ".") {
				b.finalNameMap[uid] = strings.ReplaceAll(nameInFile, ".", "_")
			}
		}
	}

	// 策略 4: 解决冲突。在应用了以上策略后，仍然可能存在名称冲突（例如，两个不同文件中的 `type T = {}`）。
	// 我们检测这些冲突，并为冲突的类型生成新的唯一名称。
	nameToUIDs := make(map[string][]UniqueTypeID)
	for uid, name := range b.finalNameMap {
		nameToUIDs[name] = append(nameToUIDs[name], uid)
	}

	for name, uids := range nameToUIDs {
		if len(uids) > 1 { // 如果一个名称对应多个类型，则存在冲突
			// 对UID进行排序以确保重命名行为是确定性的
			sort.Slice(uids, func(i, j int) bool { return uids[i] < uids[j] })
			// 保留第一个类型使用原始名称，为其余的生成新名称
			for i := 1; i < len(uids); i++ {
				b.finalNameMap[uids[i]] = b.generateUniqueName(name, uids[i])
			}
		}
	}
}

// generateUniqueName 为冲突的类型生成一个基于其原始文件名和路径的、可读性强的唯一名称。
// 例如，如果 `Button` 类型在 `components/button.ts` 和 `theme/button.ts` 中都有定义，
// 它们可能会被重命名为 `ButtonFromComponentsButton` 和 `ButtonFromThemeButton`。
func (b *TypeBundler) generateUniqueName(baseName string, uid UniqueTypeID) string {
	// 从 UID 中提取文件路径部分
	path := strings.Split(string(uid), ":")[0]
	pathParts := strings.Split(path, string(filepath.Separator))
	fileName := pathParts[len(pathParts)-1]

	// 清理文件名，移除非法字符，并转换为驼峰式
	suffix := strings.TrimSuffix(fileName, ".ts")
	suffix = strings.TrimSuffix(suffix, ".d")
	suffix = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(suffix, "_")
	suffix = strings.Title(strings.ToLower(suffix)) // e.g., "my-component" -> "MyComponent"

	// 组合新名称并确保其唯一性
	newName := fmt.Sprintf("%sFrom%s", baseName, suffix)
	counter := 1
	finalName := newName
	for b.usedNames[finalName] { // 如果新生成的名称也冲突了，则添加数字后缀
		finalName = fmt.Sprintf("%s_%d", newName, counter)
		counter++
	}
	b.usedNames[finalName] = true // 标记新名称已被使用
	return finalName
}

// updateAllTypeReferences 在所有类型名称都最终确定后，此函数会遍历每一个类型声明的源代码，
// 将其内部对其他类型的引用更新为这些类型最终确定的名称。
func (b *TypeBundler) updateAllTypeReferences() {
	for uid, decl := range b.declarations {
		newSource := decl.Raw
		resolutionMap := b.resolutionMaps[decl.FilePath]

		finalDeclName, ok := b.finalNameMap[uid]
		if !ok {
			continue // 如果该声明未被使用，则跳过
		}

		// 1. 首先，将声明本身的名称更新为其最终名称。
		//    例如 `type MyType = ...` -> `type MyTypeFromComponent = ...`
		newSource = safeReplace(newSource, decl.Name, finalDeclName)

		// 2. 然后，更新此声明内部对其他类型的所有引用。
		//    例如 `type T = { a: OtherType }` -> `type T = { a: OtherTypeFromLib }`
		for refName, refUID := range resolutionMap {
			if finalRefName, ok := b.finalNameMap[refUID]; ok {
				newSource = safeReplace(newSource, refName, finalRefName)
			}
		}

		// 3. 将更新后的源代码存回声明对象中。
		decl.Raw = newSource
		b.declarations[uid] = decl
	}
}

// generateBundleOutput 生成最终的打包文件内容。
// 它将所有处理过的、需要包含在最终产物中的类型声明，按照名称排序后拼接成一个字符串。
func (b *TypeBundler) generateBundleOutput() string {
	var finalDecls []GenericDeclaration
	// 筛选出所有需要被打包的声明
	for uid := range b.declarations {
		if _, ok := b.finalNameMap[uid]; ok {
			finalDecls = append(finalDecls, b.declarations[uid])
		}
	}

	// 对声明进行排序，以确保每次打包的结果都是稳定和一致的。
	// 主要按最终确定的类型名称进行字母序排序。
	sort.Slice(finalDecls, func(i, j int) bool {
		uid_i := UniqueTypeID(fmt.Sprintf("%s:%s", finalDecls[i].FilePath, finalDecls[i].Name))
		uid_j := UniqueTypeID(fmt.Sprintf("%s:%s", finalDecls[j].FilePath, finalDecls[j].Name))

		nameI, okI := b.finalNameMap[uid_i]
		nameJ, okJ := b.finalNameMap[uid_j]

		if okI && okJ {
			return nameI < nameJ
		}
		return uid_i < uid_j // 如果找不到名称，则按UID排序作为备用方案
	})

	// 将所有排序后的声明拼接成一个字符串。
	var result strings.Builder
	for _, decl := range finalDecls {
		trimmedSource := strings.TrimSpace(decl.Raw)
		result.WriteString(trimmedSource)
		result.WriteString("\n\n") // 在每个声明之间添加两个换行符以提高可读性
	}

	return result.String()
}
