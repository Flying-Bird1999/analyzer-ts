package query

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFindReferences_ConfiguredProjectBug 旨在复现 `typescript-go` 在处理包含 tsconfig.json 的项目时，
// `FindReferences` 无法正常工作的已知 bug。
// 预期：此测试目前会失败。
func TestFindReferences_ConfiguredProjectBug(t *testing.T) {
	// 1. 创建一个包含 tsconfig.json 和路径别名的多文件内存项目
	sources := map[string]any{
		"/tsconfig.json": `{
			"compilerOptions": {
				"baseUrl": ".",
				"paths": {
					"@/*": ["src/*"]
				}
			}
		}`,
		"/src/utils.ts": `export const myVar = 123;`,
		"/src/index.ts": `
			import { myVar } from '@/utils';
			console.log(myVar);
		`,
	}

	// 2. 使用 NewServiceForTest (一个修改版的 NewService，用于接受内存 map)
	// 注意：这需要我们先对 NewService 进行重构，或者创建一个测试专用的版本。
	// 这里我们先假设 NewServiceForTest 已经可用。
	service, err := NewServiceForTest(sources)
	assert.NoError(t, err)
	defer service.Close()

	// 3. 在 /src/index.ts 的第 3 行第 14 个字符处（第二个 myVar）查找引用
	// 我们期望找到 2 个引用：定义处和使用处。
	response, err := service.FindReferences(context.Background(), "/src/index.ts", 3, 14)

	// 打印结果以供人工检查
	t.Logf("Error: %v", err)
	t.Logf("Response: %+v", response)
	if response.Locations != nil {
		t.Logf("Found %d locations.", len(*response.Locations))
		for _, loc := range *response.Locations {
			t.Logf("  - %s:%d:%d", loc.Uri, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
		}
	}

	assert.NoError(t, err, "FindReferences 调用不应报错")
}
