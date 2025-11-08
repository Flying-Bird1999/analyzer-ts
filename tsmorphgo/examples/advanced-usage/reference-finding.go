//go:build reference_finding
// +build reference_finding

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Flying-Bird1999/analyzer-ts/tsmorphgo"
)

func main() {
	fmt.Println("🔗 TSMorphGo 引用查找示例")
	fmt.Println("=" + repeat("=", 50))

	// 创建包含变量引用的演示项目
	project := tsmorphgo.NewProjectFromSources(map[string]string{
		"/src/config/app.ts": `
			// 应用配置
			export const APP_CONFIG = {
				apiUrl: 'https://api.example.com',
				timeout: 5000,
				retryCount: 3,
				debug: process.env.NODE_ENV === 'development'
			};

			// 默认配置
			export const DEFAULT_CONFIG = {
				...APP_CONFIG,
				timeout: 3000
			};

			// 环境配置
			export const ENV_CONFIG = {
				development: {
					...APP_CONFIG,
					debug: true,
					logLevel: 'verbose'
				},
				production: {
					...APP_CONFIG,
					debug: false,
					logLevel: 'error'
				}
			};
		`,
		"/src/services/api.ts": `
			// API服务模块
			import { APP_CONFIG } from '../config/app';

			class ApiService {
				private config = APP_CONFIG;

				// 使用配置的方法
				public async makeRequest(endpoint: string, options?: RequestInit): Promise<Response> {
					const url = this.config.apiUrl + "/" + endpoint;
					const requestOptions: RequestInit = {
						timeout: this.config.timeout,
						...options
					};

					if (this.config.debug) {
						console.log('Making request to:', url);
					}

					const response = await fetch(url, requestOptions);

					// 重试逻辑
					if (!response.ok && this.config.retryCount > 0) {
						return this.retryRequest(endpoint, requestOptions);
					}

					return response;
				}

				// 配置更新方法
				public updateConfig(newConfig: Partial<typeof APP_CONFIG>): void {
					this.config = { ...this.config, ...newConfig };
				}

				// 获取当前配置
				public getConfig(): typeof APP_CONFIG {
					return this.config;
				}

				// 验证配置
				public validateConfig(): boolean {
					return !!(this.config.apiUrl && this.config.timeout);
				}
			}

			// 导出服务实例
			export const apiService = new ApiService();

			// 工具函数
			export const createApiUrl = (path: string): string => {
				return APP_CONFIG.apiUrl + "/" + path;
			};
		`,
		"/src/utils/logger.ts": `
			// 日志工具模块
			import { APP_CONFIG } from '../config/app';

			// 日志级别枚举
			enum LogLevel {
				ERROR = 'error',
				WARN = 'warn',
				INFO = 'info',
				DEBUG = 'debug'
			}

			// 日志配置
			const loggerConfig = {
				level: APP_CONFIG.debug ? LogLevel.DEBUG : LogLevel.INFO,
				timestamp: true,
				colors: true
			};

			// 日志类
			class Logger {
				private config = loggerConfig;

				// 日志方法
				public log(message: string, level: LogLevel = LogLevel.INFO): void {
					if (!this.shouldLog(level)) {
						return;
					}

					const timestamp = this.config.timestamp ?
						"[" + new Date().toISOString() + "] " : "";
					console.log(timestamp + level.toUpperCase() + ": " + message);
				}

				// 使用配置的示例
				public logConfig(): void {
					this.log("当前配置: timeout=" + APP_CONFIG.timeout + ", debug=" + APP_CONFIG.debug, LogLevel.INFO);
				}

				// 验证配置方法
				private validateConfig(): boolean {
					// 验证APP_CONFIG是否可用
					return typeof APP_CONFIG === 'object' && APP_CONFIG !== null;
				}

				// 判断是否应该记录日志
				private shouldLog(level: LogLevel): boolean {
					// 简化的级别比较逻辑
					const levels = [LogLevel.ERROR, LogLevel.WARN, LogLevel.INFO, LogLevel.DEBUG];
					const currentLevelIndex = levels.indexOf(this.config.level);
					const messageLevelIndex = levels.indexOf(level);
					return messageLevelIndex <= currentLevelIndex;
				}

				// 使用全局配置的快捷方法
				public debug(message: string): void {
					if (APP_CONFIG.debug) {
						this.log(message, LogLevel.DEBUG);
					}
				}
			}

			// 导出日志实例
			export const logger = new Logger();
		`,
	})
	defer project.Close()

	// 示例1: 基础引用查找
	fmt.Println("\n🔍 示例1: 基础引用查找")

	configFile := project.GetSourceFile("/src/config/app.ts")
	if configFile == nil {
		log.Fatal("配置文件未找到")
	}

	// 查找APP_CONFIG变量的所有引用
	var appConfigNode *tsmorphgo.Node
	configFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) &&
			strings.TrimSpace(node.GetText()) == "APP_CONFIG" &&
			node.GetParent() != nil && tsmorphgo.IsVariableDeclaration(*node.GetParent()) {
			nodeCopy := node
			appConfigNode = &nodeCopy
		}
	})

	if appConfigNode == nil {
		log.Fatal("未找到APP_CONFIG变量声明")
	}

	fmt.Printf("APP_CONFIG 变量位置: 行 %d\n", appConfigNode.GetStartLineNumber())

	// 查找所有引用
	refs, err := tsmorphgo.FindReferences(*appConfigNode)
	if err != nil {
		log.Printf("查找引用失败: %v", err)
		return
	}

	fmt.Printf("找到 %d 个APP_CONFIG引用:\n", len(refs))
	for i, ref := range refs {
		parent := ref.GetParent()
		context := ""
		if parent != nil {
			// 获取上下文（最多80字符）
			parentText := strings.TrimSpace(parent.GetText())
			if len(parentText) > 80 {
				parentText = parentText[:80] + "..."
			}
			context = parentText
		}

		fmt.Printf("  %d. %s:%d - %s\n",
			i+1, ref.GetSourceFile().GetFilePath(), ref.GetStartLineNumber(), context)
	}

	// 示例2: 带缓存的引用查找
	fmt.Println("\n⚡ 示例2: 带缓存的引用查找性能对比")

	if len(refs) > 0 {
		testRef := refs[0] // 使用第一个引用进行测试

		// 第一次查找（来自LSP服务）
		start := time.Now()
		refs1, fromCache1, err := tsmorphgo.FindReferencesWithCache(*testRef)
		duration1 := time.Since(start)

		if err != nil {
			log.Printf("查找失败: %v", err)
			return
		}

		source1 := "LSP服务"
		if fromCache1 {
			source1 = "缓存"
		}

		fmt.Printf("第一次查找:\n")
		fmt.Printf("  - 耗时: %v\n", duration1)
		fmt.Printf("  - 来源: %s\n", source1)
		fmt.Printf("  - 引用数: %d\n", len(refs1))

		// 第二次查找（应该来自缓存）
		start = time.Now()
		refs2, fromCache2, err := tsmorphgo.FindReferencesWithCache(*testRef)
		duration2 := time.Since(start)

		if err != nil {
			log.Printf("查找失败: %v", err)
			return
		}

		source2 := "LSP服务"
		if fromCache2 {
			source2 = "缓存"
		}

		fmt.Printf("第二次查找:\n")
		fmt.Printf("  - 耗时: %v\n", duration2)
		fmt.Printf("  - 来源: %s\n", source2)
		fmt.Printf("  - 引用数: %d\n", len(refs2))

		// 性能提升计算
		if duration1 > 0 && duration2 > 0 {
			speedup := float64(duration1) / float64(duration2)
			fmt.Printf("  - 性能提升: %.1fx 倍\n", speedup)
		}
	}

	// 示例3: 跳转到定义
	fmt.Println("\n📍 示例3: 跳转到定义")

	// 查找API服务文件中的APP_CONFIG使用
	apiFile := project.GetSourceFile("/src/services/api.ts")
	if apiFile != nil {
		apiFile.ForEachDescendant(func(node tsmorphgo.Node) {
			if tsmorphgo.IsIdentifier(node) &&
				strings.TrimSpace(node.GetText()) == "APP_CONFIG" &&
				(node.GetParent() == nil || !tsmorphgo.IsVariableDeclaration(*node.GetParent())) {
				// 找到了APP_CONFIG的使用，跳转到定义
				defs, err := tsmorphgo.GotoDefinition(node)
				if err != nil {
					log.Printf("跳转到定义失败: %v", err)
					return
				}

				fmt.Printf("引用位置: %s:%d\n",
					node.GetSourceFile().GetFilePath(),
					node.GetStartLineNumber())

				fmt.Printf("跳转到定义:\n")
				for i, def := range defs {
					fmt.Printf("  %d. %s:%d - %s\n",
						i+1, def.GetSourceFile().GetFilePath(),
						def.GetStartLineNumber(),
						func() string {
				text := strings.TrimSpace(def.GetText())
				if len(text) > 50 {
					text = text[:50] + "..."
				}
				return text
			}())
				}
			}
		})
	}

	// 示例4: 引用分析 - 分析变量使用模式
	fmt.Println("\n📊 示例4: 引用分析 - 变量使用模式分析")

	// 分析所有变量的使用情况
	var variableUsages []struct {
		name       string
		file       string
		declLine   int
		usageCount int
		usageFiles []string
	}

	allFiles := project.GetSourceFiles()
	for _, file := range allFiles {
		file.ForEachDescendant(func(node tsmorphgo.Node) {
			// 查找变量声明
			if tsmorphgo.IsVariableDeclaration(node) &&
				node.GetParent() != nil {
				varNameNode, ok := node.GetParent().GetFirstChild()
				if !ok {
					return
				}

				if !tsmorphgo.IsIdentifier(*varNameNode) {
					return
				}

				varName := strings.TrimSpace(varNameNode.GetText())
				if varName == "" {
					return
				}

				// 查找这个变量的引用
				varRefs, err := tsmorphgo.FindReferences(node)
				if err != nil {
					return
				}

				// 统计引用所在的文件
				usageFiles := make(map[string]bool)
				for _, ref := range varRefs {
					usageFiles[ref.GetSourceFile().GetFilePath()] = true
				}

				// 转换为切片
				fileList := make([]string, 0, len(usageFiles))
				for file := range usageFiles {
					fileList = append(fileList, file)
				}

				variableUsages = append(variableUsages, struct {
					name       string
					file       string
					declLine   int
					usageCount int
					usageFiles []string
				}{
					name:       varName,
					file:       file.GetFilePath(),
					declLine:   node.GetStartLineNumber(),
					usageCount: len(varRefs),
					usageFiles: fileList,
				})
			}
		})
	}

	fmt.Printf("变量使用分析结果:\n")
	for _, usage := range variableUsages {
		fmt.Printf("变量: %s\n", usage.name)
		fmt.Printf("  - 声明位置: %s:%d\n", usage.file, usage.declLine)
		fmt.Printf("  - 使用次数: %d\n", usage.usageCount)
		fmt.Printf("  - 使用文件: %d 个\n", len(usage.usageFiles))
		if len(usage.usageFiles) > 1 {
			fmt.Printf("  - 跨文件使用: 是\n")
		}
		fmt.Println()
	}

	// 示例5: 错误处理和降级策略
	fmt.Println("\n🛡️ 示例5: 错误处理和降级策略")

	// 创建一个可能导致错误的场景（查找不存在符号的引用）
	var nonExistentNode *tsmorphgo.Node
	configFile.ForEachDescendant(func(node tsmorphgo.Node) {
		if tsmorphgo.IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "nonExistentVar" {
			nodeCopy := node
			nonExistentNode = &nodeCopy
		}
	})

	if nonExistentNode != nil {
		refs, err := tsmorphgo.FindReferences(*nonExistentNode)
		if err != nil {
			fmt.Printf("预期错误处理: %v\n", err)
			fmt.Println("这种错误是正常的，因为查找的是不存在的变量引用")
		} else {
			fmt.Printf("意外成功找到 %d 个引用\n", len(refs))
		}
	} else {
		fmt.Println("未找到用于错误处理的测试节点")
	}

	fmt.Println("\n✅ 引用查找示例完成!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}