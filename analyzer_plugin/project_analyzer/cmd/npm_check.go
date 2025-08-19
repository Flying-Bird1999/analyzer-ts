package cmd

// go run main.go npm-check -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/npm_check -x "node_modules/**" -x "bffApiDoc/**"

import (
	"encoding/json"
	"fmt"
	"main/analyzer_plugin/project_analyzer"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	npmCheckInputDir   string
	npmCheckOutputDir  string
	npmCheckExclude    []string
	npmCheckIsMonorepo bool
)

func NewNpmCheckCmd() *cobra.Command {
	npmCheckCmd := &cobra.Command{
		Use:   "npm-check",
		Short: "检查项目的NPM依赖，类似npm-check。",
		Long:  `分析一个TypeScript项目，识别隐式依赖（幽灵依赖）、未使用的依赖和过期的依赖。`,
		Run: func(cmd *cobra.Command, args []string) {
			if npmCheckInputDir == "" {
				fmt.Println("需要输入路径。")
				cmd.Help()
				os.Exit(1)
			}

			depCheckResult := project_analyzer.CheckDependencies(npmCheckInputDir, npmCheckExclude, npmCheckIsMonorepo)

			if npmCheckOutputDir != "" {
				// 自动创建输出目录（如果不存在）
				if err := os.MkdirAll(npmCheckOutputDir, os.ModePerm); err != nil {
					fmt.Printf("Error creating output directory: %s\n", err)
					return
				}

				jsonData, err := json.MarshalIndent(depCheckResult, "", "  ")
				if err != nil {
					fmt.Printf("Error marshalling to JSON: %s\n", err)
					return
				}

				outputFile := filepath.Join(npmCheckOutputDir, filepath.Base(npmCheckInputDir)+".json")
				err = os.WriteFile(outputFile, jsonData, 0644)
				if err != nil {
					fmt.Printf("Error writing JSON to file: %s\n", err)
					return
				}

				fmt.Printf("NPM依赖检查结果已写入文件: %s\n", outputFile)
			} else {
				// Print Implicit Dependencies
				if len(depCheckResult.ImplicitDependencies) > 0 {
					fmt.Println("发现隐式依赖 (幽灵依赖):")
					for _, dep := range depCheckResult.ImplicitDependencies {
						fmt.Printf("  - %s (in %s)\n", dep.Name, dep.FilePath)
					}
				} else {
					fmt.Println("✅ 未发现隐式依赖。")
				}

				fmt.Println() // Add a separator line

				// Print Unused Dependencies
				if len(depCheckResult.UnusedDependencies) > 0 {
					fmt.Println("发现未使用依赖:")
					for _, dep := range depCheckResult.UnusedDependencies {
						fmt.Printf("  - %s (%s) (in %s)\n", dep.Name, dep.Version, dep.PackageJsonPath)
					}
				} else {
					fmt.Println("✅ 未发现未使用依赖。")
				}

				fmt.Println() // Add a separator line

				// Print Outdated Dependencies
				if len(depCheckResult.OutdatedDependencies) > 0 {
					fmt.Println("发现过期依赖:")
					for _, dep := range depCheckResult.OutdatedDependencies {
						fmt.Printf("  - %s (current: %s, latest: %s) (in %s)\n", dep.Name, dep.CurrentVersion, dep.LatestVersion, dep.PackageJsonPath)
					}
				} else {
					fmt.Println("✅ 所有依赖都是最新的。")
				}
			}
		},
	}

	npmCheckCmd.Flags().StringVarP(&npmCheckInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径")
	npmCheckCmd.Flags().StringVarP(&npmCheckOutputDir, "output", "o", "", "用于存储 JSON 输出文件的目录路径")
	npmCheckCmd.Flags().StringSliceVarP(&npmCheckExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式")
	npmCheckCmd.Flags().BoolVarP(&npmCheckIsMonorepo, "monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")

	npmCheckCmd.MarkFlagRequired("input")

	return npmCheckCmd
}
