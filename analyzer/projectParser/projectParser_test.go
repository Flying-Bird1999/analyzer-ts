package projectParser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/parser"
)

// setupTestProject 创建一个用于测试的临时项目目录结构。
// 这个函数会模拟一个包含根 `tsconfig.json`、子项目 `tsconfig.json`、
// TypeScript 源文件以及 `package.json` 的 monorepo 结构。
// 它返回临时目录的路径和一个用于在测试结束后清理该目录的函数。
func setupTestProject(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "projectParser-test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	// 创建根 tsconfig.json
	tsconfigRoot := `{
		"compilerOptions": {
			"baseUrl": ".",
			"paths": {
				"@/*": ["src/*"]
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfigRoot), 0644); err != nil {
		t.Fatalf("写入根 tsconfig 失败: %v", err)
	}

	// 为 monorepo 测试创建子项目的 tsconfig.json
	subProjectDir := filepath.Join(tmpDir, "packages", "sub")
	if err := os.MkdirAll(subProjectDir, 0755); err != nil {
		t.Fatalf("创建子项目目录失败: %v", err)
	}
	tsconfigSub := `{
		"extends": "../../tsconfig.json",
		"compilerOptions": {
			"paths": {
				"@sub/*": ["./lib/*"]
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(subProjectDir, "tsconfig.json"), []byte(tsconfigSub), 0644); err != nil {
		t.Fatalf("写入子项目 tsconfig 失败: %v", err)
	}

	// 创建一些源文件
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("创建 src 目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.ts"), []byte(`import App from "@/App";`), 0644); err != nil {
		t.Fatalf("写入 main.ts 失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "App.ts"), []byte(`export default "App";`), 0644); err != nil {
		t.Fatalf("写入 App.ts 失败: %v", err)
	}
	subLibDir := filepath.Join(subProjectDir, "lib")
	if err := os.Mkdir(subLibDir, 0755); err != nil {
		t.Fatalf("创建子项目 lib 目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subLibDir, "component.ts"), []byte(`export const MyComponent = "Component";`), 0644); err != nil {
		t.Fatalf("写入 component.ts 失败: %v", err)
	}

	// 创建根 package.json
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name": "root-pkg"}`), 0644); err != nil {
		t.Fatalf("写入根 package.json 失败: %v", err)
	}

	// 创建子项目的 package.json
	if err := os.WriteFile(filepath.Join(subProjectDir, "package.json"), []byte(`{"name": "sub-pkg"}`), 0644); err != nil {
		t.Fatalf("写入子项目 package.json 失败: %v", err)
	}

	return tmpDir, func() {
		os.RemoveAll(tmpDir)
	}
}

// TestNewProjectParserConfig 测试 NewProjectParserConfig 函数的正确性。
// 它验证在 monorepo 和非 monorepo 模式下，配置是否能被正确地创建，
// 特别是路径别名的解析是否符合预期。
func TestNewProjectParserConfig(t *testing.T) {
	rootPath, cleanup := setupTestProject(t)
	defer cleanup()

	// 测试非 monorepo 模式
	config := NewProjectParserConfig(rootPath, nil, false, []string{})
	if config.RootPath != rootPath {
		t.Errorf("预期的 RootPath 是 %s, 得到 %s", rootPath, config.RootPath)
	}
	expectedRootAlias := map[string]string{"@": "src"}
	if !reflect.DeepEqual(config.RootTsConfig.Alias, expectedRootAlias) {
		t.Errorf("预期的 RootTsConfig.Alias 是 %+v, 得到 %+v", expectedRootAlias, config.RootTsConfig.Alias)
	}
	if len(config.PackageTsConfigMaps) != 0 {
		t.Errorf("当 isMonorepo 为 false 时，预期的 PackageTsConfigMaps 为空, 得到 %d 个项目", len(config.PackageTsConfigMaps))
	}

	// 测试 monorepo 模式
	configMono := NewProjectParserConfig(rootPath, nil, true, []string{})
	if len(configMono.PackageTsConfigMaps) == 0 {
		t.Errorf("当 isMonorepo 为 true 时，预期的 PackageTsConfigMaps 不为空")
	}
	subProjectDir := filepath.Join(rootPath, "packages", "sub")
	if _, ok := configMono.PackageTsConfigMaps[subProjectDir]; !ok {
		t.Errorf("预期在 %s 找到子项目的别名", subProjectDir)
	}
	expectedSubAlias := map[string]string{"@": "src", "@sub": "./lib"}
	if !reflect.DeepEqual(configMono.PackageTsConfigMaps[subProjectDir].Alias, expectedSubAlias) {
		t.Errorf("预期的子项目别名是 %+v, 得到 %+v", expectedSubAlias, configMono.PackageTsConfigMaps[subProjectDir].Alias)
	}
}

// TestGetTsConfigForFile 测试 getTsConfigForFile 方法的正确性。
// 它验证该方法是否能为给定路径的文件（无论是位于根目录还是子项目）找到最匹配的路径别名配置。
func TestGetTsConfigForFile(t *testing.T) {
	rootPath, cleanup := setupTestProject(t)
	defer cleanup()

	config := NewProjectParserConfig(rootPath, nil, true, []string{})
	ppr := NewProjectParserResult(config)

	// 测试根目录中的文件
	rootFile := filepath.Join(rootPath, "src", "main.ts")
	alias, dir, baseUrl := ppr.getTsConfigForFile(rootFile)
	expectedRootAlias := map[string]string{"@": "src"}
	if !reflect.DeepEqual(alias, expectedRootAlias) {
		t.Errorf("预期根目录文件的别名是 %+v, 得到 %+v", expectedRootAlias, alias)
	}
	if dir != rootPath {
		t.Errorf("预期根目录文件的目录是 %s, 得到 %s", rootPath, dir)
	}
	if baseUrl != "." {
		t.Errorf("预期根目录文件的 baseUrl 是 '.', 得到 %s", baseUrl)
	}

	// 测试子项目中的文件
	subFile := filepath.Join(rootPath, "packages", "sub", "lib", "component.ts")
	alias, dir, baseUrl = ppr.getTsConfigForFile(subFile)
	expectedSubAlias := map[string]string{"@": "src", "@sub": "./lib"}
	if !reflect.DeepEqual(alias, expectedSubAlias) {
		t.Errorf("预期子项目文件的别名是 %+v, 得到 %+v", expectedSubAlias, alias)
	}
	subProjectDir := filepath.Join(rootPath, "packages", "sub")
	if dir != subProjectDir {
		t.Errorf("预期子项目文件的目录是 %s, 得到 %s", subProjectDir, dir)
	}
	// baseUrl should be inherited from root
	if baseUrl != "." {
		t.Errorf("预期子项目文件的 baseUrl 是 '.', 得到 %s", baseUrl)
	}
}

// TestTransformImportDeclarations 测试 transformImportDeclarations 方法的正确性。
// 它验证导入声明是否能被正确地转换，特别是路径别名是否能被成功解析为绝对文件路径。
func TestTransformImportDeclarations(t *testing.T) {
	rootPath, cleanup := setupTestProject(t)
	defer cleanup()

	config := NewProjectParserConfig(rootPath, nil, false, []string{})
	ppr := NewProjectParserResult(config)

	importerPath := filepath.Join(rootPath, "src", "main.ts")
	decls := []parser.ImportDeclarationResult{
		{
			Source: "@/App",
			ImportModules: []parser.ImportModule{
				{ImportModule: "default", Type: "default", Identifier: "App"},
			},
		},
	}

	transformed := ppr.transformImportDeclarations(importerPath, decls, ppr.Config.RootTsConfig.Alias, ppr.Config.RootPath, ppr.Config.RootTsConfig.BaseUrl)

	if len(transformed) != 1 {
		t.Fatalf("预期转换后有 1 个声明, 得到 %d", len(transformed))
	}

	sourceData := transformed[0].Source
	expectedFilePath := filepath.Join(rootPath, "src", "App.ts")
	if sourceData.Type != "file" {
		t.Errorf("预期的来源类型是 'file', 得到 '%s'", sourceData.Type)
	}
	if sourceData.FilePath != expectedFilePath {
		t.Errorf("预期的解析文件路径是 %s, 得到 %s", expectedFilePath, sourceData.FilePath)
	}
}

// TestProjectParser 测试 ProjectParser 的整体功能。
// 它通过解析一个模拟项目来验证是否所有的 JS/TS 文件和 package.json 文件都被正确地识别和处理。
func TestProjectParser(t *testing.T) {
	rootPath, cleanup := setupTestProject(t)
	defer cleanup()

	config := NewProjectParserConfig(rootPath, nil, true, []string{})
	ppr := NewProjectParserResult(config)
	ppr.ProjectParser()

	// 检查 JS 文件
	if len(ppr.Js_Data) != 3 {
		t.Errorf("预期解析 3 个 JS/TS 文件, 但得到 %d", len(ppr.Js_Data))
	}
	mainTsPath := filepath.Join(rootPath, "src", "main.ts")
	if _, ok := ppr.Js_Data[mainTsPath]; !ok {
		t.Errorf("预期找到 %s 的解析数据", mainTsPath)
	}

	// 检查 package.json 文件
	if len(ppr.Package_Data) != 2 {
		t.Errorf("预期解析 2 个 package.json 文件, 但得到 %d", len(ppr.Package_Data))
	}
	if _, ok := ppr.Package_Data["root"]; !ok {
		t.Errorf("预期找到根 package.json 的解析数据")
	}
	if ppr.Package_Data["root"].Namespace != "root-pkg" {
		t.Errorf("预期根 package 的命名空间是 'root-pkg', 得到 '%s'", ppr.Package_Data["root"].Namespace)
	}
	if _, ok := ppr.Package_Data["sub"]; !ok {
		t.Errorf("预期找到子项目 package.json 的解析数据")
	}
	if ppr.Package_Data["sub"].Namespace != "sub-pkg" {
		t.Errorf("预期子 package 的命名空间是 'sub-pkg', 得到 '%s'", ppr.Package_Data["sub"].Namespace)
	}
}
