package projectParser

import (
	"encoding/json"
	"fmt"
	"os"
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

// FlatDeclaration represents a single flattened declaration (interface, type, enum).
type FlatDeclaration struct {
	SourceFile      string      `json:"sourceFile"`
	DeclarationType string      `json:"declarationType"`
	Name            string      `json:"name"`
	Content         interface{} `json:"content"`
}

// FlatPackageDependency represents a single flattened package dependency.
type FlatPackageDependency struct {
	PackageJsonPath string `json:"packageJsonPath"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	Type            string `json:"type"`
	Workspace       string `json:"workspace"`
}

// FlatOutput is the final structure to be marshalled to JSON.
type FlatOutput struct {
	Imports      []FlatImport            `json:"imports"`
	Declarations []FlatDeclaration       `json:"declarations"`
	Packages     []FlatPackageDependency `json:"packages"`
}

func ProjectParser_run() {
	// inputDir := "/Users/zxc/Desktop/shopline-live-sale"
	// ar := NewProjectParserResult(inputDir, nil, nil, false)

	// inputDir := "/Users/zxc/Desktop/message-center/client"
	// inputDir := "/Users/bird/company/sc1.0/mc/message-center/client"
	inputDir := "/Users/bird/company/sc1.0/live/shopline-live-sale"
	ar := NewProjectParserResult(inputDir, nil, nil, []string{"node_modules/**"}, false)

	ar.ProjectParser()

	// Create the flattened data structure
	flatOutput := FlatOutput{
		Imports: []FlatImport{},
		// Declarations: []FlatDeclaration{},
		// Packages:     []FlatPackageDependency{},
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

		// Flatten InterfaceDeclarations
		// for name, decl := range jsData.InterfaceDeclarations {
		// 	flatDecl := FlatDeclaration{
		// 		SourceFile:      filePath,
		// 		DeclarationType: "interface",
		// 		Name:            name,
		// 		Content:         decl,
		// 	}
		// 	flatOutput.Declarations = append(flatOutput.Declarations, flatDecl)
		// }

		// Flatten TypeDeclarations
		// for name, decl := range jsData.TypeDeclarations {
		// 	flatDecl := FlatDeclaration{
		// 		SourceFile:      filePath,
		// 		DeclarationType: "type",
		// 		Name:            name,
		// 		Content:         decl,
		// 	}
		// 	flatOutput.Declarations = append(flatOutput.Declarations, flatDecl)
		// }

		// Flatten EnumDeclarations
		// for name, decl := range jsData.EnumDeclarations {
		// 	flatDecl := FlatDeclaration{
		// 		SourceFile:      filePath,
		// 		DeclarationType: "enum",
		// 		Name:            name,
		// 		Content:         decl,
		// 	}
		// 	flatOutput.Declarations = append(flatOutput.Declarations, flatDecl)
		// }
	}

	// Process Package_Data
	// for pkgPath, pkgData := range ar.Package_Data {
	// 	for _, npmItem := range pkgData.NpmList {
	// 		flatPkg := FlatPackageDependency{
	// 			PackageJsonPath: pkgPath,
	// 			Name:            npmItem.Name,
	// 			Version:         npmItem.Version,
	// 			Type:            npmItem.Type,
	// 			Workspace:       pkgData.Workspace,
	// 		}
	// 		flatOutput.Packages = append(flatOutput.Packages, flatPkg)
	// 	}
	// }

	// Define output file path
	outputFilePath := "./analyzer/projectParser/projectParser_output.json"

	// Marshal the data to JSON
	jsonData, err := json.MarshalIndent(flatOutput, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %s\n", err)
		return
	}

	// Write to file
	err = os.WriteFile(outputFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON to file: %s\n", err)
		return
	}

	fmt.Printf("分析结果已写入文件: %s\n", outputFilePath)
}
