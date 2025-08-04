package cmd

// example:
// go run main.go analyze -i /Users/bird/company/sc1.0/live/shopline-live-sale -o /Users/bird/Desktop/alalyzer/analyzer-ts -x "node_modules/**" -x "bffApiDoc/**"
// go run main.go analyze -i /Users/bird/Desktop/components/shopline-admin-components -o /Users/bird/Desktop/alalyzer/analyzer-ts -x "examples/**" -x "tests/**"

import (
	"fmt"
	"main/analyzer_plugin/project_analyzer"
	"os"

	"github.com/spf13/cobra"
)

var analyzeInputDir string
var analyzeOutputDir string
var analyzeAlias map[string]string
var analyzeExtensions []string
var analyzeIgnore []string
var analyzeIsMonorepo bool

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a TypeScript project",
	Long:  `Analyzes a TypeScript project, flattens the data, and outputs it as a JSON file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if analyzeInputDir == "" {
			fmt.Println("Usage: analyzer-ts analyze --input <directory>")
			cmd.Help()
			os.Exit(1)
		}
		project_analyzer.AnalyzeProject(analyzeInputDir, analyzeOutputDir, analyzeAlias, analyzeExtensions, analyzeIgnore, analyzeIsMonorepo)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&analyzeInputDir, "input", "i", "", "Input directory path (required)")
	analyzeCmd.Flags().StringVarP(&analyzeOutputDir, "output", "o", ".", "Output directory path")
	analyzeCmd.Flags().StringToStringVarP(&analyzeAlias, "alias", "a", nil, "Path aliases (e.g., --alias @/components=src/components)")
	analyzeCmd.Flags().StringSliceVarP(&analyzeExtensions, "extensions", "e", []string{".ts", ".tsx", ".js", ".jsx"}, "File extensions to include")
	analyzeCmd.Flags().StringSliceVarP(&analyzeIgnore, "ignore", "x", []string{}, "Gglob patterns to ignore")
	analyzeCmd.Flags().BoolVarP(&analyzeIsMonorepo, "monorepo", "m", false, "Whether the project is a monorepo")
}
