package projectParser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestFormatAlias 测试 FormatAlias 函数。
// 这个测试确保路径别名中的 `/*` 和 `*` 后缀能被正确地移除。
func TestFormatAlias(t *testing.T) {
	alias := map[string]string{
		"@/*":     "src/*",
		"@lib/*":  "src/lib/*",
		"@/utils": "src/utils",
	}
	expected := map[string]string{
		"@":       "src",
		"@/utils": "src/utils",
		"@lib":    "src/lib",
	}
	formatted := FormatAlias(alias)
	if !reflect.DeepEqual(formatted, expected) {
		t.Errorf("预期的格式化别名是 %+v, 得到 %+v", expected, formatted)
	}
}

// TestMatchImportSource 测试 MatchImportSource 函数。
// 这个测试覆盖了三种主要的导入路径解析场景：
// 1. 路径别名 (`@/App`)
// 2. 相对路径 (`./utils`)
// 3. NPM 包 (`react`, `@angular/core`)
func TestMatchImportSource(t *testing.T) {
	tmpDir, cleanup := setupTestProject(t)
	defer cleanup()

	importerPath := filepath.Join(tmpDir, "src", "main.ts")
	basePath := tmpDir
	alias := map[string]string{"@": "src"}
	extensions := []string{".ts", ".tsx", ".d.ts"}

	// 测试路径别名匹配
	sourceData := MatchImportSource(importerPath, "@/App", basePath, alias, extensions)
	expectedPath := filepath.Join(tmpDir, "src", "App.ts")
	if sourceData.Type != "file" || sourceData.FilePath != expectedPath {
		t.Errorf("预期的别名匹配应解析为 %s, 得到类型 %s 和路径 %s", expectedPath, sourceData.Type, sourceData.FilePath)
	}

	// 测试相对路径匹配
	// 创建一个 utils 文件用于相对路径导入测试
	if err := os.WriteFile(filepath.Join(tmpDir, "src/utils.ts"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	sourceData = MatchImportSource(importerPath, "./utils", basePath, alias, extensions)
	expectedPath = filepath.Join(tmpDir, "src", "utils.ts")
	if sourceData.Type != "file" || sourceData.FilePath != expectedPath {
		t.Errorf("预期的相对路径匹配应解析为 %s, 得到类型 %s 和路径 %s", expectedPath, sourceData.Type, sourceData.FilePath)
	}

	// 测试 NPM 包匹配
	sourceData = MatchImportSource(importerPath, "react", basePath, alias, extensions)
	if sourceData.Type != "npm" || sourceData.NpmPkg != "react" {
		t.Errorf("预期的 npm 匹配应解析为 'react', 得到类型 %s 和包 %s", sourceData.Type, sourceData.NpmPkg)
	}

	// 测试带作用域的 NPM 包
	sourceData = MatchImportSource(importerPath, "@angular/core", basePath, alias, extensions)
	if sourceData.Type != "npm" || sourceData.NpmPkg != "@angular/core" {
		t.Errorf("预期的带作用域的 npm 匹配应解析为 '@angular/core', 得到类型 %s 和包 %s", sourceData.Type, sourceData.NpmPkg)
	}

	// 测试 .d.ts 文件的别名解析
	dtsPath := filepath.Join(tmpDir, "src", "feature", "LiveRoom", "components", "MainRight", "components", "Comments")
	if err := os.MkdirAll(dtsPath, 0755); err != nil {
		t.Fatal(err)
	}
	dtsFile := filepath.Join(dtsPath, "type.d.ts")
	if err := os.WriteFile(dtsFile, []byte("export type MyType = string;"), 0644); err != nil {
		t.Fatal(err)
	}
	sourceData = MatchImportSource(importerPath, "@/feature/LiveRoom/components/MainRight/components/Comments/type", basePath, alias, extensions)
	if sourceData.Type != "file" || sourceData.FilePath != dtsFile {
		t.Errorf("预期的 .d.ts 别名匹配应解析为 %s, 得到类型 %s 和路径 %s", dtsFile, sourceData.Type, sourceData.FilePath)
	}
}

// TestExtractNpmPackageName 测试 extractNpmPackageName 函数。
// 它验证函数是否能从不同的导入路径格式中正确地提取出 NPM 包的名称。
func TestExtractNpmPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"react", "react"},
		{"react/jsx-runtime", "react"},
		{"@abc/core", "@abc/core"},
		{"@abc/core/testing", "@abc/core"},
		{"@yy/sl-admin-components/es/SLAntd/components/date-picker/generatePicker", "@yy/sl-admin-components"},
		{"@sl/sc-components/src/aaa", "@sl/sc-components"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := extractNpmPackageName(tt.input)
			if actual != tt.expected {
				t.Errorf("预期的包名是 '%s', 但得到 '%s'", tt.expected, actual)
			}
		})
	}
}

// TestGetPackageJson 测试 GetPackageJson 函数。
// 这个测试验证函数是否能正确地解析一个 `package.json` 文件，
// 包括提取名称、版本以及不同类型的依赖，并获取其在 node_modules 中的实际版本。
func TestGetPackageJson(t *testing.T) {
	tmpDir, cleanup := setupTestProject(t)
	defer cleanup()

	// 创建一个带依赖的 package.json
	pkgJsonContent := `{
        "name": "test-pkg",
        "version": "1.0.0",
        "dependencies": {
            "react": "^18.2.0"
        },
        "devDependencies": {
            "typescript": "^4.7.4"
        }
    }`
	pkgJsonPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgJsonPath, []byte(pkgJsonContent), 0644); err != nil {
		t.Fatalf("写入 package.json 失败: %v", err)
	}

	// 创建一个模拟的 node_modules 结构
	nodeModulesDir := filepath.Join(tmpDir, "node_modules", "react")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("创建 node_modules 目录失败: %v", err)
	}
	reactPkgJsonContent := `{"name": "react", "version": "18.2.0"}`
	if err := os.WriteFile(filepath.Join(nodeModulesDir, "package.json"), []byte(reactPkgJsonContent), 0644); err != nil {
		t.Fatalf("写入 react package.json 失败: %v", err)
	}

	pkgInfo, err := GetPackageJson(pkgJsonPath)
	if err != nil {
		t.Fatalf("GetPackageJson 执行失败: %v", err)
	}

	if pkgInfo.Name != "test-pkg" || pkgInfo.Version != "1.0.0" {
		t.Errorf("预期的名称是 'test-pkg' 且版本是 '1.0.0', 得到 '%s' 和 '%s'", pkgInfo.Name, pkgInfo.Version)
	}

	if len(pkgInfo.NpmList) != 2 {
		t.Errorf("预期有 2 个 npm 依赖, 得到 %d", len(pkgInfo.NpmList))
	}

	if reactDep, ok := pkgInfo.NpmList["react"]; !ok || reactDep.Version != "^18.2.0" || reactDep.NodeModuleVersion != "18.2.0" {
		t.Errorf("React 依赖信息不正确: %+v", reactDep)
	}
}
