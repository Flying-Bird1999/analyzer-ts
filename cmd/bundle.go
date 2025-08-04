package cmd

// example:
// go run main.go bundle -i /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index1.ts -t Class -o /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/output/result.ts

import (
	"fmt"
	"main/analyzer_plugin/ts_bundle"
	"os"

	"github.com/spf13/cobra"
)

var inputFile string
var inputType string
var outputFile string
var projectRoot string

var rootCmd = &cobra.Command{
	Use:   "analyzer-ts",
	Short: "A CLI tool for analyzing TypeScript projects",
	Long:  `analyzer-ts is a powerful CLI tool designed to help you analyze and understand your TypeScript projects.`,
}

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Bundle TypeScript type declarations",
	Long:  `Recursively collects all referenced type declarations from a given entry file and bundles them into a single file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if inputFile == "" || inputType == "" {
			fmt.Println("Usage: analyzer-ts bundle --input <entry file> --type <type name>")
			cmd.Help()
			os.Exit(1)
		}
		ts_bundle.GenerateBundle(inputFile, inputType, outputFile, projectRoot)
	},
}

func init() {
	rootCmd.AddCommand(bundleCmd)

	bundleCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Entry file path (required)")
	bundleCmd.Flags().StringVarP(&inputType, "type", "t", "", "Type name to analyze (required)")
	bundleCmd.Flags().StringVarP(&outputFile, "output", "o", "./output.ts", "Output file path")
	bundleCmd.Flags().StringVarP(&projectRoot, "root", "r", "", "Project root path")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
