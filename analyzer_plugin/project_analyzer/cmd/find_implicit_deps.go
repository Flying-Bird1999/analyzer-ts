package cmd

// go run main.go find-implicit-deps -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/implicit_deps_result -x "node_modules/**" -x "bffApiDoc/**"

// go run main.go find-implicit-deps -i /Users/bird/company/sc1.0/components/nova -o /Users/bird/Desktop/alalyzer/analyzer-ts/analyzer_plugin/project_analyzer/result/implicit_deps_result -m true -x "**/e2e/**" -x "**/dist/**" -x "**/demo/**"

import (
	"encoding/json"
	"fmt"
	"main/analyzer_plugin/project_analyzer"
	"main/cmd"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	findImplicitDepsInputDir   string
	findImplicitDepsOutputDir  string
	findImplicitDepsExclude    []string
	findImplicitDepsIsMonorepo bool
)

var findImplicitDepsCmd = &cobra.Command{
	Use:   "find-implicit-deps",
	Short: "查找项目中的隐式依赖（幽灵依赖）。",
	Long:  `分析一个 TypeScript 项目，并识别出那些在代码中被使用但没有在 package.json 文件中明确声明的依赖。`,
	Run: func(cmd *cobra.Command, args []string) {
		if findImplicitDepsInputDir == "" {
			fmt.Println("需要输入路径。")
			cmd.Help()
			os.Exit(1)
		}
		implicitDependencies := project_analyzer.FindImplicitDependencies(findImplicitDepsInputDir, findImplicitDepsExclude, findImplicitDepsIsMonorepo)

		if findImplicitDepsOutputDir != "" {
			jsonData, err := json.MarshalIndent(implicitDependencies, "", "  ")
			if err != nil {
				fmt.Printf("Error marshalling to JSON: %s\n", err)
				return
			}

			outputFile := filepath.Join(findImplicitDepsOutputDir, filepath.Base(findImplicitDepsInputDir)+"-implicit-deps.json")
			err = os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				fmt.Printf("Error writing JSON to file: %s\n", err)
				return
			}

			fmt.Printf("隐式依赖分析结果已写入文件: %s\n", outputFile)
		} else {
			if len(implicitDependencies) > 0 {
				fmt.Println("发现隐式依赖 (幽灵依赖):")
				for _, dep := range implicitDependencies {
					fmt.Printf("  - %s in %s\n", dep.Name, dep.FilePath)
				}
			} else {
				fmt.Println("未发现隐式依赖。")
			}
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(findImplicitDepsCmd)

	findImplicitDepsCmd.Flags().StringVarP(&findImplicitDepsInputDir, "input", "i", "", "要分析的 TypeScript 项目目录的路径")
	findImplicitDepsCmd.Flags().StringVarP(&findImplicitDepsOutputDir, "output", "o", "", "用于存储 JSON 输出文件的目录路径")
	findImplicitDepsCmd.Flags().StringSliceVarP(&findImplicitDepsExclude, "exclude", "x", []string{}, "要从分析中排除的 Glob 模式")
	findImplicitDepsCmd.Flags().BoolVarP(&findImplicitDepsIsMonorepo, "monorepo", "m", false, "如果要分析的是 monorepo，则设置为 true")

	findImplicitDepsCmd.MarkFlagRequired("input")
}
