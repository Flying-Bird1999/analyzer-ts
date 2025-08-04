package project_analyzer

import (
	"encoding/json"
	"fmt"
	"main/analyzer/projectParser"
	"os"
	"path/filepath"
)

// FlatImport represents a single flattened import declaration.
type FlatImport struct {
	SourceFile     string `json:"sourceFile"`
	ImportedModule string `json:"importedModule"`
	Identifier     string `json:"identifier"`
	ImportType     string `json:"importType"`
	Source         string `json:"source"`
	SourceType     string `json:"sourceType"`
	Raw            string `json:"raw"`
}

// FlatPackageDependency represents a single flattened package dependency.
type FlatPackageDependency struct {
	PackageJsonPath   string `json:"packageJsonPath"`
	Name              string `json:"name"`
	Version           string `json:"version"`
	NodeModuleVersion string `json:"nodeModuleVersion"`
	Type              string `json:"type"`
	Workspace         string `json:"workspace"`
}

// FlatOutput is the final structure to be marshalled to JSON.
type FlatOutput struct {
	Imports  []FlatImport            `json:"imports"`
	Packages []FlatPackageDependency `json:"packages"`
}

func AnalyzeProject(rootPath string, outputDir string, alias map[string]string, extensions []string, ignore []string, isMonorepo bool) {
	ar := projectParser.NewProjectParserResult(rootPath, alias, extensions, ignore, isMonorepo)

	ar.ProjectParser()

	// Create the flattened data structure
	flatOutput := FlatOutput{
		Imports:  []FlatImport{},
		Packages: []FlatPackageDependency{},
	}

	// Process Js_Data
	for filePath, jsData := range ar.Js_Data {
		// Flatten ImportDeclarations
		for _, importDecl := range jsData.ImportDeclarations {
			for _, module := range importDecl.ImportModules {
				flatImport := FlatImport{
					SourceFile:     filePath,
					ImportedModule: module.ImportModule,
					Identifier:     module.Identifier,
					ImportType:     module.Type,
					Source:         importDecl.Source.FilePath,
					SourceType:     importDecl.Source.Type,
					Raw:            importDecl.Raw,
				}
				flatOutput.Imports = append(flatOutput.Imports, flatImport)
			}
		}
	}

	// Process Package_Data
	for pkgPath, pkgData := range ar.Package_Data {
		for _, npmItem := range pkgData.NpmList {
			flatPkg := FlatPackageDependency{
				PackageJsonPath:   pkgPath,
				Name:              npmItem.Name,
				Version:           npmItem.Version,
				NodeModuleVersion: npmItem.NodeModuleVersion,
				Type:              npmItem.Type,
				Workspace:         pkgData.Workspace,
			}
			flatOutput.Packages = append(flatOutput.Packages, flatPkg)
		}
	}

	// Marshal the data to JSON
	jsonData, err := json.MarshalIndent(flatOutput, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %s\n", err)
		return
	}

	// Write to file
	outputFile := filepath.Join(outputDir, filepath.Base(rootPath)+".json")
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %s\n", err)
		return
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFile)
}
