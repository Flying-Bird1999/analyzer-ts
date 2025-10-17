package dependency

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Flying-Bird1999/analyzer-ts/analyzer/projectParser"
)

func TestFindImplicitAndUsedDependencies(t *testing.T) {
	// 准备路径
	projectRoot, _ := filepath.Abs("/test-project")
	indexPath := filepath.Join(projectRoot, "index.ts")

	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Js_Data: map[string]projectParser.JsFileParserResult{
			indexPath: {
				ImportDeclarations: []projectParser.ImportDeclarationResult{
					// 使用已声明的依赖
					{Source: projectParser.SourceData{Type: "npm", NpmPkg: "used-lib"}},
					// 使用未声明的依赖（隐式依赖）
					{Source: projectParser.SourceData{Type: "npm", NpmPkg: "implicit-lib"}},
				},
			},
		},
	}
	declaredDependencies := map[string]bool{"used-lib": true, "unused-lib": true}

	// 2. 执行函数
	implicitDeps, usedDeps := findImplicitAndUsedDependencies(mockParsingResult, declaredDependencies)

	// 3. 断言结果
	// 检查隐式依赖
	if len(implicitDeps) != 1 {
		t.Fatalf("Expected 1 implicit dependency, but got %d", len(implicitDeps))
	}
	if implicitDeps[0].Name != "implicit-lib" {
		t.Errorf("Expected implicit dependency to be 'implicit-lib', but got %s", implicitDeps[0].Name)
	}

	// 检查已使用依赖
	expectedUsed := map[string]bool{"used-lib": true, "implicit-lib": true}
	if !reflect.DeepEqual(usedDeps, expectedUsed) {
		t.Errorf("Expected used dependencies to be %v, but got %v", expectedUsed, usedDeps)
	}
}

func TestFindUnusedDependencies(t *testing.T) {
	// 准备路径
	projectRoot, _ := filepath.Abs("/test-project")
	pkgJsonPath := filepath.Join(projectRoot, "package.json")

	// 1. 准备测试数据
	mockParsingResult := &projectParser.ProjectParserResult{
		Package_Data: map[string]projectParser.PackageJsonFileParserResult{
			pkgJsonPath: {
				NpmList: map[string]projectParser.NpmItem{
					"used-lib":   {Name: "used-lib"},
					"unused-lib": {Name: "unused-lib"},
					// @types/node 应该被忽略
					"@types/node": {Name: "@types/node"},
				},
			},
		},
	}
	// 模拟只有 used-lib 被使用了
	usedDependencies := map[string]bool{"used-lib": true}

	// 2. 执行函数
	unusedDeps := findUnusedDependencies(mockParsingResult, usedDependencies)

	// 3. 断言结果
	if len(unusedDeps) != 1 {
		t.Fatalf("Expected 1 unused dependency, but got %d", len(unusedDeps))
	}
	if unusedDeps[0].Name != "unused-lib" {
		t.Errorf("Expected unused dependency to be 'unused-lib', but got %s", unusedDeps[0].Name)
	}
}
