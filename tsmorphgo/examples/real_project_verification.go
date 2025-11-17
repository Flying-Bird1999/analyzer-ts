//go:build examples

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🎯 TSMorphGo 真实项目完整验证示例")
	fmt.Println("==============================")
	fmt.Println("验证项目: /Users/bird/company/sc1.0/mc/message-center/client")
	fmt.Println()

	// 项目路径
	projectPath := "/Users/bird/company/sc1.0/mc/message-center/client"

	// 创建项目
	project := tsmorphgo.NewProject(tsmorphgo.ProjectConfig{
		RootPath:    projectPath,
		UseTsConfig: true,
	})

	if project == nil {
		log.Fatal("❌ 项目创建失败")
	}

	// 获取项目信息
	sourceFiles := project.GetSourceFiles()
	fmt.Printf("✅ 项目初始化成功，扫描到 %d 个源文件\n", len(sourceFiles))

	// ============================================================================
	// 诉求1: 类型引用查找
	// ============================================================================

	// 1.1 查找 DetailDataType 的引用
	fmt.Println()
	fmt.Println("🔍 诉求1.1: 查找 DetailDataType 类型引用")
	fmt.Println("-------------------------------------")

	detailDataTypeFile := project.GetSourceFile(filepath.Join(projectPath, "src/feature/Broadcast/views/BroadcastEditor/constant/index.ts"))
	if detailDataTypeFile == nil {
		fmt.Println("❌ 未找到 constant/index.ts 文件")
	} else {
		fmt.Printf("✅ 找到文件: %s\n", detailDataTypeFile.GetFilePath())

		// 查找第188行的 DetailDataType
		var detailDataTypeNode tsmorphgo.Node
		var detailDataTypeFound bool

		detailDataTypeFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 188 && node.IsIdentifier() && node.GetText() == "DetailDataType" {
				// 确保是 TypeAliasDeclaration 中的标识符
				if parent := node.GetParent(); parent != nil && parent.IsKind(tsmorphgo.KindTypeAliasDeclaration) {
					detailDataTypeNode = node
					detailDataTypeFound = true
					fmt.Printf("✅ 找到 DetailDataType 类型定义: 第%d行\n", node.GetStartLineNumber())
				}
			}
		})

		if detailDataTypeFound {
			// 查找引用
			if refs, err := detailDataTypeNode.FindReferences(); err != nil {
				fmt.Printf("❌ DetailDataType 引用查找失败: %v\n", err)
			} else {
				fmt.Printf("✅ 找到 DetailDataType 的 %d 个引用:\n", len(refs))
				for i, ref := range refs {
					refFile := ref.GetSourceFile()
					if refFile != nil {
						line := ref.GetStartLineNumber()
						col := ref.GetStartColumnNumber()
						text := ref.GetText()
						filePath := refFile.GetFilePath()

						if line == 188 {
							fmt.Printf("  %d. 【类型定义】\n", i+1)
						} else {
							fmt.Printf("  %d. 【类型使用】\n", i+1)
						}
						fmt.Printf("     文件路径: %s\n", filePath)
						fmt.Printf("     位置: 第%d行，第%d列\n", line, col)
						fmt.Printf("     内容: %s\n\n", text)
					}
				}
			}
		} else {
			fmt.Println("❌ 未找到 DetailDataType 类型定义")
		}
	}

	// 1.2 查找 ContentType 的定义和引用
	fmt.Println()
	fmt.Println("🔍 诉求1.2: 查找 ContentType 类型定义和引用")
	fmt.Println("---------------------------------------")

	// 重新获取这个文件以避免作用域问题
	constantFile := project.GetSourceFile(filepath.Join(projectPath, "src/feature/Broadcast/views/BroadcastEditor/constant/index.ts"))
	if constantFile == nil {
		fmt.Println("❌ 未找到 constant/index.ts 文件")
	} else {
		fmt.Printf("✅ 找到文件: %s\n", constantFile.GetFilePath())

		// 查找第112行的 ContentType 定义
		var contentTypeDefNode tsmorphgo.Node
		var contentTypeDefFound bool

		constantFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 112 && node.IsIdentifier() && node.GetText() == "ContentType" {
				// 确保是 TypeAliasDeclaration 中的标识符
				if parent := node.GetParent(); parent != nil && parent.IsKind(tsmorphgo.KindTypeAliasDeclaration) {
					contentTypeDefNode = node
					contentTypeDefFound = true
					fmt.Printf("✅ 找到 ContentType 类型定义: 第%d行\n", node.GetStartLineNumber())
				}
			}
		})

		if contentTypeDefFound {
			// 查找引用
			if refs, err := contentTypeDefNode.FindReferences(); err != nil {
				fmt.Printf("❌ ContentType 引用查找失败: %v\n", err)
			} else {
				fmt.Printf("✅ 找到 ContentType 的 %d 个引用:\n", len(refs))

				// 特别检查第237行的使用
				foundLine237 := false
				for i, ref := range refs {
					refFile := ref.GetSourceFile()
					if refFile != nil {
						line := ref.GetStartLineNumber()
						col := ref.GetStartColumnNumber()
						text := ref.GetText()
						filePath := refFile.GetFilePath()

						if line == 112 {
							fmt.Printf("  %d. 【类型定义】\n", i+1)
						} else {
							fmt.Printf("  %d. 【类型使用】\n", i+1)
						}
						fmt.Printf("     文件路径: %s\n", filePath)
						fmt.Printf("     位置: 第%d行，第%d列\n", line, col)
						fmt.Printf("     内容: %s\n\n", text)

						// 检查是否是第237行
						if line == 237 {
							foundLine237 = true
							fmt.Printf("✅ 第237行引用确认: 这是 ContentType 的使用，位于 %s:%d:%d\n", filepath.Base(filePath), line, col)

							// 分析第237行的上下文
							parent := ref.GetParent()
							if parent != nil {
								fmt.Printf("📍 第237行上下文分析:\n")
								fmt.Printf("   父节点类型: %s\n", parent.GetKind().String())
								fmt.Printf("   父节点内容: %s\n", parent.GetText())

								// 检查是否是数组类型
								if strings.Contains(parent.GetText(), "[]") {
									fmt.Printf("✅ 确认这是 ContentType[] 数组类型的使用\n")
								}
							}
						}
					}
				}

				if !foundLine237 {
					fmt.Println("⚠️  未在第237行找到ContentType引用，可能是行号有误")
				}
			}
		} else {
			fmt.Println("❌ 未找到 ContentType 类型定义")
		}
	}

	// ============================================================================
	// 诉求2: 导入语句和函数调用分析 - 使用AsXXX API获取struct数据
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 诉求2: 导入语句和函数调用高级分析")
	fmt.Println("---------------------------------")

	editorIndexFile := project.GetSourceFile(filepath.Join(projectPath, "src/feature/Broadcast/views/BroadcastEditor/index.tsx"))
	if editorIndexFile == nil {
		fmt.Println("❌ 未找到 BroadcastEditor/index.tsx 文件")
	} else {
		fmt.Printf("✅ 找到文件: %s\n", editorIndexFile.GetFilePath())

		// 2.1 查找FbBroadcastEditor导入语句并使用AsImportDeclaration
		fmt.Println("\n📦 2.1 导入语句高级分析 (AsImportDeclaration)")
		fmt.Println("--------------------------------------------------")

		var targetImportNode *tsmorphgo.Node
		editorIndexFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsImportDeclaration() && targetImportNode == nil {
				// 查找包含 FbBroadcastEditor 的导入语句
				if strings.Contains(node.GetText(), "FbBroadcastEditor") {
					n := node
					targetImportNode = &n
					fmt.Printf("✅ 找到目标导入语句: %s\n", strings.TrimSpace(node.GetText()))
				}
			}
		})

		if targetImportNode != nil {
			fmt.Println("🔧 使用 AsImportDeclaration 获取struct数据:")

			// 使用AsImportDeclaration进行类型收窄
			importDecl, success := targetImportNode.AsImportDeclaration()
			if !success {
				fmt.Println("❌ 类型收窄失败")
			} else {
				fmt.Println("✅ 成功收窄为 ImportDeclaration struct")
				fmt.Printf("📋 Struct类型: %T\n", importDecl)

				// 直接访问struct内部字段
				fmt.Println("\n📋 ImportDeclaration Struct 内部数据:")

				// 检查Node字段
				if importDecl.Node != nil {
					fmt.Printf("✅ Node字段存在: %T\n", importDecl.Node)
					fmt.Printf("📍 Node类型(Kind): %v\n", importDecl.Node.Kind)
				} else {
					fmt.Println("❌ Node字段为空")
				}

				// 使用GetParserData获取更多详细信息
				fmt.Println("\n🔧 GetParserData 详细分析:")
				if parserData, ok := targetImportNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)

					// 尝试类型断言，看看具体有什么字段
					switch data := parserData.(type) {
					case map[string]interface{}:
						fmt.Println("📋 数据是map类型:")
						for key, value := range data {
							fmt.Printf("  %s: %v (%T)\n", key, value, value)
						}
					default:
						fmt.Printf("📋 其他类型数据: %+v\n", data)
					}
				}

				// 分析导入语句的结构
				fmt.Println("\n🔍 导入语句结构分析:")
				importStatementText := strings.TrimSpace(targetImportNode.GetText())
				fmt.Printf("📝 完整导入语句: %s\n", importStatementText)

				// 手动解析导入内容
				if strings.Contains(importStatementText, "import") && strings.Contains(importStatementText, "from") {
					fmt.Println("✅ 标准ES6导入格式")

					// 提取from部分
					fromIndex := strings.LastIndex(importStatementText, "from")
					if fromIndex != -1 {
						modulePath := importStatementText[fromIndex+4:]
						modulePath = strings.TrimSpace(modulePath)
						modulePath = strings.Trim(modulePath, `"'`)
						fmt.Printf("🔗 导入模块路径: %s\n", modulePath)
					}

					// 提取导入内容部分
					importPart := strings.TrimSpace(importStatementText[:fromIndex])
					importPart = strings.TrimPrefix(importPart, "import")
					importPart = strings.TrimSpace(importPart)
					fmt.Printf("📦 导入内容: %s\n", importPart)

					// 分析是否有默认导出和命名导出
					if strings.Contains(importPart, "{") {
						fmt.Println("✅ 包含命名导出")

						// 提取默认导出
						defaultExport := strings.TrimSpace(strings.Split(importPart, "{")[0])
						if defaultExport != "" {
							fmt.Printf("🏷️  默认导出: %s\n", defaultExport)
						}

						// 提取命名导出
						namedPart := strings.TrimSpace(strings.Split(importPart, "{")[1])
						namedPart = strings.TrimSuffix(namedPart, "}")
						namedPart = strings.TrimSpace(namedPart)
						fmt.Printf("🏷️  命名导出: %s\n", namedPart)
					} else {
						fmt.Printf("🏷️  仅默认导出: %s\n", importPart)
					}
				}
			}
		} else {
			fmt.Println("❌ 未找到包含FbBroadcastEditor的导入语句")
		}

		// 2.2 查找notification.error调用并使用AsCallExpression
		fmt.Println("\n📞 2.2 函数调用高级分析 (AsCallExpression)")
		fmt.Println("-----------------------------------------")

		var notificationErrorCall *tsmorphgo.Node
		editorIndexFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsCallExpression() && notificationErrorCall == nil {
				// 检查是否是notification.error调用
				if strings.Contains(node.GetText(), "notification.error") {
					n := node
					notificationErrorCall = &n
					fmt.Printf("✅ 找到notification.error调用\n")
				}
			}
		})

		if notificationErrorCall != nil {
			fmt.Println("🔧 使用 AsCallExpression 获取struct数据:")

			// 使用AsCallExpression进行类型收窄
			callExpr, success := notificationErrorCall.AsCallExpression()
			if !success {
				fmt.Println("❌ 类型收窄失败")
			} else {
				fmt.Println("✅ 成功收窄为 CallExpression struct")
				fmt.Printf("📋 Struct类型: %T\n", callExpr)

				// 使用CallExpression的专有API
				fmt.Println("\n📞 CallExpression Struct 内部数据:")

				// 获取函数表达式
				if expression := callExpr.GetExpression(); expression != nil {
					fmt.Printf("✅ 函数表达式: %s\n", expression.GetText())
					fmt.Printf("📍 函数表达式类型: %s\n", expression.GetKind().String())

					// 如果是属性访问表达式，进一步分析
					if expression.IsPropertyAccessExpression() {
						fmt.Println("🔗 属性访问表达式详情:")

						objectName := ""
						propertyName := ""

						expression.ForEachChild(func(child tsmorphgo.Node) bool {
							if child.IsIdentifier() {
								if objectName == "" {
									objectName = child.GetText()
								} else if propertyName == "" {
									propertyName = child.GetText()
								}
							}
							return false
						})

						fmt.Printf("🏷️  对象名: %s\n", objectName)
						fmt.Printf("🏷️  属性名: %s\n", propertyName)

						// 获取符号信息
						if symbol, err := expression.GetSymbol(); err == nil && symbol != nil {
							fmt.Printf("🏷️  符号名称: %s\n", symbol.GetName())
							fmt.Printf("🆔 符号ID: %d\n", symbol.GetId())
						}
					}
				}

				// 获取参数列表
				if arguments := callExpr.GetArguments(); arguments != nil {
					fmt.Printf("📋 参数数量: %d\n", len(arguments))

					for i, arg := range arguments {
						fmt.Printf("\n  参数 %d:\n", i+1)
						fmt.Printf("    内容: %s\n", arg.GetText())
						fmt.Printf("    类型: %s\n", arg.GetKind().String())

						// 如果是对象字面量，详细分析
						if arg.IsObjectLiteralExpression() {
							fmt.Println("    📦 对象字面量分析:")

							arg.ForEachChild(func(child tsmorphgo.Node) bool {
								if child.IsKind(tsmorphgo.KindPropertyAssignment) {
									fmt.Printf("      属性: %s\n", child.GetText())

									// 分析属性赋值
									propName := ""
									propValue := ""

									child.ForEachChild(func(propChild tsmorphgo.Node) bool {
										if propChild.IsIdentifier() && propName == "" {
											propName = propChild.GetText()
										} else if (propChild.IsKind(tsmorphgo.KindStringLiteral) ||
											propChild.IsKind(tsmorphgo.KindNumericLiteral) ||
											propChild.IsIdentifier()) && propValue == "" {
											propValue = propChild.GetText()
										}
										return false
									})

									if propName != "" {
										fmt.Printf("        属性名: %s\n", propName)
									}
									if propValue != "" {
										fmt.Printf("        属性值: %s\n", propValue)
									}
								}
								return false
							})
						}

						// 获取参数符号信息
						if symbol, err := arg.GetSymbol(); err == nil && symbol != nil {
							fmt.Printf("    🏷️  参数符号: %s\n", symbol.GetName())
						}
					}
				}

				// 使用GetParserData获取底层解析数据
				fmt.Println("\n🔧 GetParserData 底层数据:")
				if parserData, ok := notificationErrorCall.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)

					// 尝试类型断言访问具体字段
					switch data := parserData.(type) {
					case map[string]interface{}:
						fmt.Println("📋 CallExpression数据字段:")
						for key, value := range data {
							if fmt.Sprintf("%v", value) == "[map[]]" {
								fmt.Printf("  %s: [复杂数据结构]\n", key)
							} else {
								fmt.Printf("  %s: %v (%T)\n", key, value, value)
							}
						}
					default:
						dataStr := fmt.Sprintf("%v", data)
						if len(dataStr) > 200 {
							dataStr = dataStr[:200] + "..."
						}
						fmt.Printf("📋 CallExpression数据: %s\n", dataStr)
					}
				}

				// 获取位置信息
				fmt.Println("\n📍 调用位置信息:")
				fmt.Printf("🎯 起始位置: %d (第%d行，第%d列)\n",
					notificationErrorCall.GetStart(),
					notificationErrorCall.GetStartLineNumber(),
					notificationErrorCall.GetStartColumnNumber())
			}
		} else {
			fmt.Println("❌ 未找到notification.error调用")
		}
	}

	// ============================================================================
	// 诉求3: 函数和变量分析
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 诉求3: 函数和变量分析")
	fmt.Println("----------------------")

	utilsFile := project.GetSourceFile(filepath.Join(projectPath, "src/feature/Broadcast/utils/index.ts"))
	if utilsFile == nil {
		fmt.Println("❌ 未找到 Broadcast/utils/index.ts 文件")
	} else {
		fmt.Printf("✅ 找到文件: %s\n", utilsFile.GetFilePath())

		// 3.1 查找 downloadFile 函数 (扩大搜索范围)
		fmt.Println("\n📥 3.1 downloadFile 函数分析")
		fmt.Println("---------------------------")

		var downloadFileNode *tsmorphgo.Node
		var downloadFileFound bool

		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			// 查找 downloadFile 标识符，行号在45-55之间
			if node.IsIdentifier() && node.GetText() == "downloadFile" &&
			   node.GetStartLineNumber() >= 45 && node.GetStartLineNumber() <= 55 {

				// 检查是否是变量声明中的标识符
				parent := node.GetParent()
				if parent != nil && (parent.IsVariableDeclaration()) {
					n := node
					downloadFileNode = &n
					downloadFileFound = true
					fmt.Printf("✅ 找到 downloadFile 函数: 第%d行\n", node.GetStartLineNumber())
				}
			}
		})

		if downloadFileFound {
			fmt.Println("🔧 downloadFile 节点分析:")

			// 基本信息
			fmt.Printf("🏷️  节点类型: %s\n", downloadFileNode.GetKind().String())
			fmt.Printf("📍 位置: 第%d行，第%d列\n",
				downloadFileNode.GetStartLineNumber(),
				downloadFileNode.GetStartColumnNumber())

			// 查找父节点
			parent := downloadFileNode.GetParent()
			if parent != nil {
				fmt.Printf("👨‍👦 父节点类型: %s\n", parent.GetKind().String())

				// 如果是变量声明，获取函数信息
				if parent.IsVariableDeclaration() {
					varDecl, success := parent.AsVariableDeclaration()
					if success {
						fmt.Printf("📝 变量名: %s\n", varDecl.GetName())

						// 分析参数和初始值
						if varDecl.HasInitializer() {
							initializer := varDecl.GetInitializer()
							if initializer != nil {
								fmt.Printf("📋 函数表达式: %s\n", initializer.GetText())

								// 分析函数参数
								paramCount := 0
								initializer.ForEachChild(func(child tsmorphgo.Node) bool {
									if child.IsKind(tsmorphgo.KindParameter) {
										paramCount++
										fmt.Printf("📋 参数 %d: %s\n", paramCount, child.GetText())
									}
									return false
								})
							}
						}
					}
				}
			}

			// 查找函数体内容
			if parent != nil && parent.IsVariableDeclaration() {
				if varDecl, success := parent.AsVariableDeclaration(); success && varDecl.HasInitializer() {
					initializer := varDecl.GetInitializer()
					if initializer != nil {
						fmt.Println("📄 函数体内容分析:")

						// 查找函数体中的关键节点
						funcBodyElements := 0
						initializer.ForEachDescendant(func(descendant tsmorphgo.Node) {
							if descendant.IsKind(tsmorphgo.KindIfStatement) ||
							   descendant.IsKind(tsmorphgo.KindReturnStatement) {
								funcBodyElements++
								if funcBodyElements <= 3 { // 只显示前3个元素
									fmt.Printf("   %d. %s: %s\n", funcBodyElements, descendant.GetKind().String(), descendant.GetText())
								}
							}
						})
						fmt.Printf("📊 函数体包含 %d 个语句\n", funcBodyElements)
					}
				}
			}

		} else {
			fmt.Println("❌ 未找到 downloadFile 函数")
		}

		// 3.2 查找 isContentsSuccess 变量 (第183行)
		fmt.Println("\n✅ 3.2 isContentsSuccess 变量分析 (第183行)")
		fmt.Println("--------------------------------------")

		var isContentsSuccessNode *tsmorphgo.Node
		var isContentsSuccessFound bool

		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.GetStartLineNumber() == 183 && node.IsIdentifier() && node.GetText() == "isContentsSuccess" {
				// 检查是否是变量声明中的标识符
				if parent := node.GetParent(); parent != nil && parent.IsVariableDeclaration() {
					n := node
					isContentsSuccessNode = &n
					isContentsSuccessFound = true
					fmt.Printf("✅ 找到 isContentsSuccess 变量: 第%d行\n", node.GetStartLineNumber())
				}
			}
		})

		if isContentsSuccessFound {
			// 获取变量声明节点
			parent := isContentsSuccessNode.GetParent()
			if parent != nil && parent.IsVariableDeclaration() {
				fmt.Println("🔧 isContentsSuccess 变量声明分析:")

				// 获取初始值（右边部分，应该是个函数）
				if varDecl, success := parent.AsVariableDeclaration(); success && varDecl.HasInitializer() {
					initializer := varDecl.GetInitializer()
					if initializer != nil {
						fmt.Printf("📝 初始值类型: %s\n", initializer.GetKind().String())
						fmt.Printf("📝 初始值文本: %s\n", initializer.GetText())

						// 遍历函数节点，找出其中的某些节点
						fmt.Println("🔍 函数内部节点分析:")

						// 查找函数参数
						initializer.ForEachChild(func(child tsmorphgo.Node) bool {
							if child.IsKind(tsmorphgo.KindParameter) {
								fmt.Printf("📋 函数参数: %s\n", child.GetText())
							}
							return false
						})

						// 查找函数体中的关键节点
						funcBodyFound := false
						initializer.ForEachDescendant(func(descendant tsmorphgo.Node) {
							// 查找 return 语句
							if descendant.IsKind(tsmorphgo.KindReturnStatement) {
								fmt.Printf("🔄 Return语句: %s\n", descendant.GetText())
								funcBodyFound = true
							}

							// 查找条件表达式，使用类型收窄进行详细分析
							if descendant.IsKind(tsmorphgo.KindBinaryExpression) {
								text := descendant.GetText()
								if len(text) < 100 { // 避免过长的表达式
									fmt.Printf("🔀 二元表达式: %s\n", text)
								}

								// 使用AsBinaryExpression进行类型收窄
								fmt.Println("🔧 二元表达式类型收窄分析:")
								if binaryExpr, success := descendant.AsBinaryExpression(); success {
									fmt.Println("✅ 成功收窄为 BinaryExpression struct")
									fmt.Printf("📋 Struct类型: %T\n", binaryExpr)

									// 使用BinaryExpression的专有API
									fmt.Println("\n📊 BinaryExpression 专有API信息:")

									// 获取左操作数
									if left := binaryExpr.GetLeft(); left != nil {
										fmt.Printf("⬅️ 左操作数: %s\n", left.GetText())
										fmt.Printf("📍 左操作数类型: %s\n", left.GetKind().String())

										// 如果左操作数也是二元表达式，递归分析
										if left.IsKind(tsmorphgo.KindBinaryExpression) {
											fmt.Printf("🔁 左操作数也是二元表达式，递归分析:\n")
											if leftBinary, success := left.AsBinaryExpression(); success {
												if leftExpr := leftBinary.GetLeft(); leftExpr != nil {
													fmt.Printf("   ⬅️ 左操作数的左操作数: %s\n", leftExpr.GetText())
												}
												if operator := leftBinary.GetOperatorToken(); operator != nil {
													fmt.Printf("   ➕ 左操作数的操作符: %s\n", operator.GetText())
												}
												if rightExpr := leftBinary.GetRight(); rightExpr != nil {
													fmt.Printf("   ➡️ 左操作数的右操作数: %s\n", rightExpr.GetText())
												}
											}
										}
									}

									// 获取操作符
									if operator := binaryExpr.GetOperatorToken(); operator != nil {
										fmt.Printf("➕ 操作符: %s\n", operator.GetText())
										fmt.Printf("📍 操作符位置: 第%d行，第%d列\n",
											operator.GetStartLineNumber(),
											operator.GetStartColumnNumber())
									}

									// 获取右操作数
									if right := binaryExpr.GetRight(); right != nil {
										fmt.Printf("➡️ 右操作数: %s\n", right.GetText())
										fmt.Printf("📍 右操作数类型: %s\n", right.GetKind().String())

										// 如果右操作数是函数调用，进一步分析
										if right.IsCallExpression() {
											fmt.Printf("🔗 右操作数是函数调用，进一步分析:\n")
											if callExpr, success := right.AsCallExpression(); success {
												if expr := callExpr.GetExpression(); expr != nil {
													fmt.Printf("   📞 函数名: %s\n", expr.GetText())
												}
												if args := callExpr.GetArguments(); args != nil {
													fmt.Printf("   📋 函数参数数量: %d\n", len(args))
												}
											}
										}
									}

									// 使用GetParserData获取底层数据
									fmt.Println("\n🔧 GetParserData 底层数据:")
									if parserData, ok := descendant.GetParserData(); ok {
										fmt.Printf("✅ Parser数据类型: %T\n", parserData)

										// 尝试类型断言访问具体字段
										switch data := parserData.(type) {
										case map[string]interface{}:
											fmt.Println("📋 BinaryExpression数据字段:")
											for key, value := range data {
												if fmt.Sprintf("%v", value) == "[map[]]" {
													fmt.Printf("  %s: [复杂数据结构]\n", key)
												} else {
													dataStr := fmt.Sprintf("%v", value)
													if len(dataStr) > 100 {
														dataStr = dataStr[:100] + "..."
													}
													fmt.Printf("  %s: %s (%T)\n", key, dataStr, value)
												}
											}
										default:
											dataStr := fmt.Sprintf("%v", data)
											if len(dataStr) > 200 {
												dataStr = dataStr[:200] + "..."
											}
											fmt.Printf("📋 BinaryExpression数据: %s\n", dataStr)
										}
									}

									// 表达式位置信息
									fmt.Println("\n📍 表达式位置信息:")
									fmt.Printf("🎯 起始位置: %d (第%d行，第%d列)\n",
										descendant.GetStart(),
										descendant.GetStartLineNumber(),
										descendant.GetStartColumnNumber())
									fmt.Printf("🎯 结束位置: %d (第%d行，第%d列)\n",
										descendant.GetEnd(),
										descendant.GetEndLineNumber(),
										descendant.GetEndColumnNumber())
								} else {
									fmt.Println("❌ 二元表达式类型收窄失败")
								}
								fmt.Println() // 添加空行分隔
							}

							// 查找标识符
							if descendant.IsIdentifier() {
								// 只显示一些关键标识符
								identText := descendant.GetText()
								if len(identText) <= 20 {
									fmt.Printf("🏷️  标识符: %s (行:%d)\n", identText, descendant.GetStartLineNumber())
								}
							}
						})

						if !funcBodyFound {
							// 如果没有找到函数体，可能不是箭头函数，尝试其他分析
							fmt.Println("ℹ️  未找到标准的函数体结构，尝试其他分析:")

							// 查找所有子节点类型
							childTypes := make(map[string]int)
							initializer.ForEachChild(func(child tsmorphgo.Node) bool {
								kind := child.GetKind().String()
								childTypes[kind]++
								return false
							})

							fmt.Printf("📊 子节点类型统计: %+v\n", childTypes)
						}
					}
				}
			}
		} else {
			fmt.Println("❌ 未找到 isContentsSuccess 变量")
		}
	}

	// ============================================================================
	// 诉求4: 更多API验证 - 验证还未使用的AsXXX API
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 诉求4: 更多API验证 - 未使用的AsXXX API")
	fmt.Println("--------------------------------------")

	// 使用包含丰富API的文件进行验证
	shopperInterfaceFile := project.GetSourceFile(filepath.Join(projectPath, "src/shopper/interface/index.ts"))

	// 4.1 验证 AsInterfaceDeclaration - 接口声明 (使用shopper接口文件)
	fmt.Println("\n🔗 4.1 接口声明验证 (AsInterfaceDeclaration)")
	fmt.Println("-----------------------------------------")

	if shopperInterfaceFile != nil {
		fmt.Printf("✅ 使用文件: %s\n", shopperInterfaceFile.GetFilePath())

		var interfaceDeclNode *tsmorphgo.Node
		shopperInterfaceFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsKind(tsmorphgo.KindInterfaceDeclaration) && interfaceDeclNode == nil {
				n := node
				interfaceDeclNode = &n
				fmt.Printf("✅ 找到接口声明: %s (第%d行)\n", node.GetText(), node.GetStartLineNumber())
			}
		})

		if interfaceDeclNode != nil {
			if interfaceDecl, success := interfaceDeclNode.AsInterfaceDeclaration(); success {
				fmt.Println("✅ 成功收窄为 InterfaceDeclaration struct")
				fmt.Printf("📋 Struct类型: %T\n", interfaceDecl)

				// 访问接口声明的专有属性
				fmt.Printf("🔧 接口信息:\n")
				fmt.Printf("   节点类型: %s\n", interfaceDeclNode.GetKind().String())
				fmt.Printf("   位置: 第%d行\n", interfaceDeclNode.GetStartLineNumber())

				// 使用GetParserData
				if parserData, ok := interfaceDeclNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)
				}
			} else {
				fmt.Println("❌ InterfaceDeclaration 类型收窄失败")
			}
		} else {
			fmt.Println("❌ 未找到接口声明")
		}

		// 4.2 验证 AsExportDeclaration - 导出声明
		fmt.Println("\n📤 4.2 导出声明验证 (AsExportDeclaration)")
		fmt.Println("---------------------------------------")

		var exportDeclNode *tsmorphgo.Node
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsKind(tsmorphgo.KindExportDeclaration) && exportDeclNode == nil {
				n := node
				exportDeclNode = &n
				fmt.Printf("✅ 找到导出声明: %s (第%d行)\n", node.GetText(), node.GetStartLineNumber())
			}
		})

		if exportDeclNode != nil {
			if exportDecl, success := exportDeclNode.AsExportDeclaration(); success {
				fmt.Println("✅ 成功收窄为 ExportDeclaration struct")
				fmt.Printf("📋 Struct类型: %T\n", exportDecl)

				// 使用GetParserData
				if parserData, ok := exportDeclNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)
				}
			} else {
				fmt.Println("❌ ExportDeclaration 类型收窄失败")
			}
		} else {
			fmt.Println("❌ 未找到导出声明")
		}

		// 4.3 验证 AsPropertyAccessExpression - 属性访问表达式
		fmt.Println("\n🔗 4.3 属性访问表达式验证 (AsPropertyAccessExpression)")
		fmt.Println("--------------------------------------------------")

		var propAccessNode *tsmorphgo.Node
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsPropertyAccessExpression() && propAccessNode == nil {
				// 选择一些有代表性的属性访问
				text := node.GetText()
				if len(text) > 5 && len(text) < 50 { // 避免太长或太短的表达式
					n := node
					propAccessNode = &n
					fmt.Printf("✅ 找到属性访问表达式: %s (第%d行)\n", text, node.GetStartLineNumber())
				}
			}
		})

		if propAccessNode != nil {
			if propAccess, success := propAccessNode.AsPropertyAccessExpression(); success {
				fmt.Println("✅ 成功收窄为 PropertyAccessExpression struct")
				fmt.Printf("📋 Struct类型: %T\n", propAccess)

				// 使用PropertyAccessExpression的专有API
				fmt.Println("🔧 PropertyAccessExpression 专有API信息:")

				// 获取表达式对象
				if expression := propAccess.GetExpression(); expression != nil {
					fmt.Printf("🏷️ 表达式对象: %s\n", expression.GetText())
					fmt.Printf("📍 对象类型: %s\n", expression.GetKind().String())
				}

				// 获取属性名
				if name := propAccess.GetName(); name != "" {
					fmt.Printf("🏷️ 属性名: %s\n", name)
				} else {
					// 备用方法：从最后一个子节点获取
					children := propAccess.GetChildren()
					if len(children) > 0 {
						lastChild := children[len(children)-1]
						if lastChild.IsIdentifier() {
							fmt.Printf("🏷️ 属性名(从子节点): %s\n", lastChild.GetText())
						}
					}
				}

				// 使用GetParserData
				if parserData, ok := propAccessNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)
				}
			} else {
				fmt.Println("❌ PropertyAccessExpression 类型收窄失败")
			}
		} else {
			fmt.Println("❌ 未找到合适的属性访问表达式")
		}

		// 4.4 验证 AsImportSpecifier - 导入规范器
		fmt.Println("\n📦 4.4 导入规范器验证 (AsImportSpecifier)")
		fmt.Println("---------------------------------------")

		var importSpecifierNode *tsmorphgo.Node
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsKind(tsmorphgo.KindImportSpecifier) && importSpecifierNode == nil {
				n := node
				importSpecifierNode = &n
				fmt.Printf("✅ 找到导入规范器: %s (第%d行)\n", node.GetText(), node.GetStartLineNumber())
			}
		})

		if importSpecifierNode != nil {
			if importSpecifier, success := importSpecifierNode.AsImportSpecifier(); success {
				fmt.Println("✅ 成功收窄为 ImportSpecifier struct")
				fmt.Printf("📋 Struct类型: %T\n", importSpecifier)

				// 使用ImportSpecifier的专有API
				fmt.Println("🔧 ImportSpecifier 专有API信息:")

				// 获取导入的名称
				originalName := importSpecifier.GetOriginalName()
				if originalName != "" {
					fmt.Printf("🏷️ 原始名称: %s\n", originalName)
				}

				// 检查是否有别名
				if aliasNode := importSpecifier.GetAliasNode(); aliasNode != nil {
					fmt.Printf("🏷️ 别名: %s\n", aliasNode.GetText())
				} else {
					fmt.Printf("🏷️ 无别名\n")
				}

				// 使用GetParserData
				if parserData, ok := importSpecifierNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)
				}
			} else {
				fmt.Println("❌ ImportSpecifier 类型收窄失败")
			}
		} else {
			fmt.Println("❌ 未找到导入规范器")
		}
	}

	// ============================================================================
	// 诉求5: 更多API验证 - 函数声明和类型别名
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 诉求5: 函数声明和类型别名API验证")
	fmt.Println("----------------------------------")

	if utilsFile != nil {
		fmt.Printf("✅ 使用文件: %s\n", utilsFile.GetFilePath())

		// 5.1 验证 AsFunctionDeclaration - 函数声明
		fmt.Println("\n📝 5.1 函数声明验证 (AsFunctionDeclaration)")
		fmt.Println("-----------------------------------------")

		var funcDeclNode *tsmorphgo.Node
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsFunctionDeclaration() && funcDeclNode == nil {
				n := node
				funcDeclNode = &n
				fmt.Printf("✅ 找到函数声明: %s (第%d行)\n", node.GetText(), node.GetStartLineNumber())
			}
		})

		if funcDeclNode != nil {
			if funcDecl, success := funcDeclNode.AsFunctionDeclaration(); success {
				fmt.Println("✅ 成功收窄为 FunctionDeclaration struct")
				fmt.Printf("📋 Struct类型: %T\n", funcDecl)

				// 使用FunctionDeclaration的基础信息
				fmt.Println("🔧 FunctionDeclaration 基本信息:")

				// 获取函数名
				if name := funcDecl.GetName(); name != "" {
					fmt.Printf("🏷️ 函数名: %s\n", name)
				}

				// 手动分析函数参数
				fmt.Printf("📋 函数参数分析:\n")
				funcDeclNode.ForEachChild(func(child tsmorphgo.Node) bool {
					if child.IsKind(tsmorphgo.KindParameter) {
						fmt.Printf("  参数: %s\n", child.GetText())
					}
					return false
				})

				// 检查是否包含async关键字 (通过文本分析)
				funcText := funcDeclNode.GetText()
				if strings.Contains(funcText, "async") {
					fmt.Printf("🔄 可能是异步函数\n")
				}
				if strings.Contains(funcText, "*") {
					fmt.Printf("🔄 可能是生成器函数\n")
				}

				// 使用GetParserData
				if parserData, ok := funcDeclNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)
				}
			} else {
				fmt.Println("❌ FunctionDeclaration 类型收窄失败")
			}
		} else {
			fmt.Println("❌ 未找到函数声明")
		}

		// 5.2 验证 AsTypeAliasDeclaration - 类型别名
		fmt.Println("\n🏷️ 5.2 类型别名验证 (AsTypeAliasDeclaration)")
		fmt.Println("-------------------------------------------")

		var typeAliasDeclNode *tsmorphgo.Node
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if node.IsKind(tsmorphgo.KindTypeAliasDeclaration) && typeAliasDeclNode == nil {
				n := node
				typeAliasDeclNode = &n
				fmt.Printf("✅ 找到类型别名: %s (第%d行)\n", node.GetText(), node.GetStartLineNumber())
			}
		})

		if typeAliasDeclNode != nil {
			if typeAliasDecl, success := typeAliasDeclNode.AsTypeAliasDeclaration(); success {
				fmt.Println("✅ 成功收窄为 TypeAliasDeclaration struct")
				fmt.Printf("📋 Struct类型: %T\n", typeAliasDecl)

				// 获取类型别名信息
				fmt.Printf("🔧 类型别名信息:\n")
				fmt.Printf("   节点类型: %s\n", typeAliasDeclNode.GetKind().String())
				fmt.Printf("   位置: 第%d行\n", typeAliasDeclNode.GetStartLineNumber())

				// 使用GetParserData
				if parserData, ok := typeAliasDeclNode.GetParserData(); ok {
					fmt.Printf("✅ Parser数据类型: %T\n", parserData)
				}
			} else {
				fmt.Println("❌ TypeAliasDeclaration 类型收窄失败")
			}
		} else {
			fmt.Println("❌ 未找到类型别名")
		}

		// 5.3 验证其他重要节点类型
		fmt.Println("\n🔧 5.3 其他重要节点类型验证")
		fmt.Println("---------------------------")

		// 统计各种节点类型
		nodeTypeCount := make(map[tsmorphgo.SyntaxKind]int)
		utilsFile.ForEachDescendant(func(node tsmorphgo.Node) {
			nodeTypeCount[node.GetKind()]++
		})

		fmt.Println("📊 节点类型统计:")
		sortedTypes := make([]tsmorphgo.SyntaxKind, 0, len(nodeTypeCount))
		for kind := range nodeTypeCount {
			sortedTypes = append(sortedTypes, kind)
		}

		// 按数量排序，显示前10种
		for i := 0; i < 10 && i < len(sortedTypes); i++ {
			for j := i + 1; j < len(sortedTypes); j++ {
				if nodeTypeCount[sortedTypes[j]] > nodeTypeCount[sortedTypes[i]] {
					sortedTypes[i], sortedTypes[j] = sortedTypes[j], sortedTypes[i]
				}
			}
		}

		for i := 0; i < 10 && i < len(sortedTypes); i++ {
			kind := sortedTypes[i]
			fmt.Printf("  %d. %s: %d 个\n", i+1, kind.String(), nodeTypeCount[kind])
		}
	}

	// ============================================================================
	// 诉求6: Project 和 SourceFile 维度的 API 输出
	// ============================================================================

	fmt.Println()
	fmt.Println("🔍 诉求6: Project 和 SourceFile 维度的 API 输出")
	fmt.Println("-------------------------------------------")

	// 4.1 Project 维度的 API 输出
	fmt.Println()
	fmt.Println("🏗️  Project 维度 API 输出:")
	fmt.Println("------------------------")

	fmt.Printf("📂 项目根路径: %s\n", projectPath)

	// 获取 TypeScript 配置
	tsConfig := project.GetTsConfig()
	if tsConfig != nil {
		fmt.Printf("✅ TypeScript 配置加载成功\n")
		if tsConfig.CompilerOptions != nil {
			fmt.Printf("📋 编译器选项数量: %d\n", len(tsConfig.CompilerOptions))

			// 显示一些重要的编译选项
			if target, ok := tsConfig.CompilerOptions["target"]; ok {
				fmt.Printf("🎯 Target: %v\n", target)
			}
			if module, ok := tsConfig.CompilerOptions["module"]; ok {
				fmt.Printf("📦 Module: %v\n", module)
			}
			if jsx, ok := tsConfig.CompilerOptions["jsx"]; ok {
				fmt.Printf("⚛️  JSX: %v\n", jsx)
			}
		}
	} else {
		fmt.Println("⚠️  未找到 TypeScript 配置")
	}

	// 统计文件类型分布
	fileTypes := make(map[string]int)
	tsFiles := 0
	tsxFiles := 0
	jsFiles := 0
	jsxFiles := 0
	otherFiles := 0

	for _, file := range sourceFiles {
		ext := filepath.Ext(file.GetFilePath())
		switch ext {
		case ".ts":
			tsFiles++
		case ".tsx":
			tsxFiles++
		case ".js":
			jsFiles++
		case ".jsx":
			jsxFiles++
		default:
			otherFiles++
		}
		fileTypes[ext]++
	}

	fmt.Printf("📊 文件类型统计:\n")
	fmt.Printf("   TypeScript (.ts): %d\n", tsFiles)
	fmt.Printf("   TypeScript (.tsx): %d\n", tsxFiles)
	fmt.Printf("   JavaScript (.js): %d\n", jsFiles)
	fmt.Printf("   JavaScript (.jsx): %d\n", jsxFiles)
	fmt.Printf("   其他文件: %d\n", otherFiles)

	// 4.2 SourceFile 维度的 API 输出
	fmt.Println()
	fmt.Println("📄 SourceFile 维度 API 输出:")
	fmt.Println("----------------------------")

	// 选择几个代表性文件进行分析
	sampleFiles := []string{
		"src/feature/Broadcast/views/BroadcastEditor/constant/index.ts",
		"src/feature/Broadcast/views/BroadcastEditor/index.tsx",
		"src/feature/Broadcast/utils/index.ts",
	}

	for _, relativePath := range sampleFiles {
		file := project.GetSourceFile(filepath.Join(projectPath, relativePath))
		if file == nil {
			continue
		}

		fmt.Printf("\n📁 文件: %s\n", relativePath)

		// 基础信息
		fileResult := file.GetFileResult()
		if fileResult != nil {
			fmt.Printf("   - 文件大小: %d 字符\n", len(fileResult.Raw))
			fmt.Printf("   - 导入声明: %d 个\n", len(fileResult.ImportDeclarations))
			fmt.Printf("   - 导出声明: %d 个\n", len(fileResult.ExportDeclarations))
			fmt.Printf("   - 接口声明: %d 个\n", len(fileResult.InterfaceDeclarations))
			fmt.Printf("   - 函数声明: %d 个\n", len(fileResult.FunctionDeclarations))
			fmt.Printf("   - 变量声明: %d 个\n", len(fileResult.VariableDeclarations))
			fmt.Printf("   - 类型别名声明: %d 个\n", len(fileResult.TypeDeclarations))
			fmt.Printf("   - 调用表达式: %d 个\n", len(fileResult.CallExpressions))
		}

		// AST 节点统计
		nodeTypes := make(map[string]int)
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			kind := node.GetKind().String()
			nodeTypes[kind]++
		})

		// 显示最常见的节点类型
		fmt.Printf("   - 节点类型最多的前5种:\n")
		count := 0
		for kind, num := range nodeTypes {
			if count >= 5 {
				break
			}
			fmt.Printf("     %d. %s: %d\n", count+1, kind, num)
			count++
		}
	}

	// 清理资源
	defer project.Close()

	fmt.Println()
	fmt.Println("🎉 真实项目完整验证示例完成！")
	fmt.Println()
	fmt.Println("✅ 验证总结:")
	fmt.Println("   - 类型引用查找 (DetailDataType, ContentType): 完成")
	fmt.Println("   - 导入语句高级分析 (AsImportDeclaration): 完成")
	fmt.Println("   - 函数调用高级分析 (AsCallExpression): 完成")
	fmt.Println("   - 函数和变量节点分析 (downloadFile, isContentsSuccess): 完成")
	fmt.Println("   - Project 和 SourceFile API 输出: 完成")
	fmt.Println("   - GetParserData 底层数据访问: 完成")
}