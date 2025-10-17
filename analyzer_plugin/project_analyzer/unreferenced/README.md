# Find-Unreferenced-Files 分析器

## 概述

Find-Unreferenced-Files 分析器是 analyzer-ts 工具中的专业死代码检测插件，采用图论算法和智能分类技术，用于识别项目中未被任何其他文件引用的"死代码"文件。该分析器通过构建完整的文件依赖图，从入口文件开始执行深度优先搜索，精准识别可以安全删除的孤立文件。

## 功能特性

### 🔍 智能死代码检测
- **图论算法**：采用深度优先搜索(DFS)算法分析文件可达性
- **入口点识别**：支持自定义入口文件或自动识别常见入口模式
- **依赖关系分析**：分析导入、导出、JSX 组件引用等多种关系
- **循环引用处理**：智能处理复杂的循环依赖和 re-export 场景

### 🎯 多层次文件分类
- **真正的未引用文件**：可以安全删除的死代码文件
- **可疑文件**：需要人工确认的重要文件（配置文件、入口文件等）
- **测试文件过滤**：自动忽略测试文件、类型声明、故事文件等
- **配置文件识别**：智能识别构建配置、代码质量配置等文件

### 📊 详细的分析报告
- **统计信息**：提供项目整体文件引用状况的统计数据
- **配置追溯**：记录分析使用的配置参数，确保结果可重现
- **分类展示**：按文件类别清晰展示分析结果
- **优先级排序**：按处理优先级排序文件列表

### 🛠️ 灵活的配置选项
- **自定义入口**：支持指定多个入口文件路径
- **智能入口检测**：自动识别常见的入口文件模式
- **文件排除**：支持 glob 模式排除特定目录
- **精确分析**：可选择包含或不包含入口目录

## 使用方法

### 基本用法
```bash
# 分析项目中的未引用文件
./analyzer-ts analyze find-unreferenced-files -i /path/to/project

# 指定入口文件进行分析
./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.entrypoint=src/index.ts"

# 启用智能入口检测
./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.include-entry-dirs=true"

# 将分析结果保存为 JSON 文件
./analyzer-ts analyze find-unreferenced-files -i /path/to/project -o /path/to/output.json

# 在 monorepo 项目中使用
./analyzer-ts analyze find-unreferenced-files -i /path/to/monorepo -m
```

### 高级用法
```bash
# 指定多个入口文件
./analyzer-ts analyze find-unreferenced-files -i /path/to/project \
  -p "unreferenced.entrypoint=src/index.ts,src/App.tsx"

# 结合文件排除优化分析
./analyzer-ts analyze find-unreferenced-files -i /path/to/project \
  -x "node_modules/**" \
  -x "**/dist/**" \
  -x "**/coverage/**"

# 分析特定目录
./analyzer-ts analyze find-unreferenced-files -i /path/to/project/src/components

# 仅分析生产代码
./analyzer-ts analyze find-unreferenced-files -i /path/to/project \
  -x "**/*.test.ts" \
  -x "**/*.spec.ts" \
  -x "**/test/**" \
  -x "**/__tests__/**"
```

## 输出示例

### 控制台输出（无未引用文件）
```
✅ 扫描文件 156 个，发现 0 个真正未引用文件和 0 个可疑文件。没有发现任何未引用文件。
```

### 控制台输出（发现未引用文件）
```
⚠️ 扫描文件 156 个，发现 5 个真正未引用文件和 3 个可疑文件。

--- 🗑️ 真正未引用的文件 (可以安全删除) ---
  - /src/components/OldButton.tsx
  - /src/utils/deprecated-helper.ts
  - /src/services/legacy-api.ts
  - /src/hooks/use-legacy-effect.ts
  - /src/styles/deprecated-theme.scss

--- 🤔 可疑的未引用文件 (请人工检查) ---
  - /src/config.ts
  - /src/router/index.ts
  - /src/store/index.ts
```

### JSON 输出
```json
{
  "configuration": {
    "inputDir": "/path/to/project",
    "entrypointsSpecified": true,
    "includeEntryDirs": false
  },
  "stats": {
    "totalFiles": 156,
    "referencedFiles": 148,
    "trulyUnreferencedFiles": 5,
    "suspiciousFiles": 3
  },
  "entrypointFiles": [
    "/src/index.ts"
  ],
  "suspiciousFiles": [
    "/src/config.ts",
    "/src/router/index.ts",
    "/src/store/index.ts"
  ],
  "trulyUnreferencedFiles": [
    "/src/components/OldButton.tsx",
    "/src/utils/deprecated-helper.ts",
    "/src/services/legacy-api.ts",
    "/src/hooks/use-legacy-effect.ts",
    "/src/styles/deprecated-theme.scss"
  ]
}
```

## 技术架构

### 工作原理

分析器采用基于图论的深度优先搜索算法，包含四个核心阶段：

1. **引用关系图构建**
   - 遍历所有文件的导入语句
   - 分析所有文件的导出语句
   - 识别 JSX 组件引用关系
   - 建立完整的文件依赖图

2. **入口文件识别**
   - 接受用户指定的入口文件
   - 或自动识别常见入口模式
   - 将入口文件作为搜索起始点

3. **可达性分析**
   - 从入口文件开始执行 DFS
   - 标记所有可达文件
   - 识别不可达的未引用文件

4. **智能文件分类**
   - 应用启发式规则分类未引用文件
   - 区分真正的死代码和重要文件
   - 生成结构化的分析报告

### 核心算法

**深度优先搜索算法**：
```go
func performDFS(entrypointFiles, deps) {
    visited := make(map[string]bool)

    var dfs func(filePath string)
    dfs = func(filePath string) {
        if visited[filePath] {
            return
        }
        visited[filePath] = true

        // 递归访问所有依赖文件
        for each dependency in fileDeps {
            dfs(dependency.filePath)
        }
    }

    // 从所有入口文件开始搜索
    for each entrypoint in entrypointFiles {
        dfs(entrypoint)
    }

    return visited
}
```

### 智能分类规则

**层次 1: 忽略规则**
- 测试文件：`.test.`、`.spec.`、`__tests__`
- 故事文件：`.story.`、`.stories.`
- 类型声明：`.d.ts`

**层次 2: 配置文件识别**
- 构建配置：webpack、vite、rollup、babel
- 代码质量：prettier、eslint、stylelint
- 测试配置：jest、cypress、playwright

**层次 3: 位置和命名分析**
- 非 src 目录的文件标记为可疑
- 入口文件模式标记为可疑
- 核心模块模式标记为可疑

### 核心组件
```go
// 分析器主体
type Finder struct {
    entrypoints      []string // 自定义入口文件
    includeEntryDirs bool     // 是否包含入口目录模式
}

// 分析结果
type FindUnreferencedFilesResult struct {
    Configuration   AnalysisConfiguration   // 分析配置
    Stats           SummaryStats            // 统计数据
    EntrypointFiles []string                // 入口文件列表
    SuspiciousFiles  []string                // 可疑文件
    TrulyUnreferencedFiles []string          // 真正未引用文件
}
```

### 性能优化

- **DFS 算法**：确保在有限时间内完成分析
- **缓存优化**：使用 map 结构实现 O(1) 查找
- **内存效率**：避免不必要的数据复制
- **并发安全**：支持多线程分析场景

## 最佳实践

### 1. 代码清理和优化
```bash
# 定期清理死代码
./analyzer-ts analyze find-unreferenced-files -i ./src -o cleanup-$(date +%Y%m%d).json

# 在发布前检查
./analyzer-ts analyze find-unreferenced-files -i ./src -p "unreferenced.entrypoint=src/index.ts"

# 监控文件引用健康度
./analyzer-ts analyze find-unreferenced-files -i ./ | jq '.stats.referencedFiles / .stats.totalFiles'
```

### 2. 项目重构规划
```bash
# 重构前分析
./analyzer-ts analyze find-unreferenced-files -i ./src -o before-refactor.json

# 重构后验证
./analyzer-ts analyze find-unreferenced-files -i ./src -o after-refactor.json

# 对比重构效果
jq -n '{before: input, after: input}' before-refactor.json after-refactor.json > refactor-comparison.json
```

### 3. 架构维护
```bash
# 识别孤立的模块
./analyzer-ts analyze find-unreferenced-files -i ./src/modules -p "unreferenced.entrypoint=modules/index.ts"

# 分析组件依赖结构
./analyzer-ts analyze find-unreferenced-files -i ./src/components -p "unreferenced.include-entry-dirs=true"

# 检查配置文件完整性
./analyzer-ts analyze find-unreferenced-files -i ./src/config | jq '.suspiciousFiles'
```

### 4. 团队协作
```bash
# 在代码合并前运行检查
./analyzer-ts analyze find-unreferenced-files -i ./src -o pre-merge-$(date +%Y%m%d).json

# 生成项目维护报告
./analyzer-ts analyze find-unreferenced-files -i ./ -o maintenance-report.json

# 监控代码健康度趋势
./analyzer-ts analyze find-unreferenced-files -i ./src -o health-$(date +%Y%m%d).json
```

### 5. CI/CD 集成
```bash
# 设置文件引用健康阈值
if [ $(./analyzer-ts analyze find-unreferenced-files -i ./src -o - | jq '.stats.totalFiles - .stats.referencedFiles') -gt 10 ]; then
    echo "警告：发现大量未引用文件，请检查项目结构"
    exit 1
fi

# 阻止合并包含死代码的 PR
./analyzer-ts analyze find-unreferenced-files -i ./src -o result.json
if [ $(jq '.trulyUnreferencedFiles | length' result.json) -gt 5 ]; then
    echo "错误：项目包含过多死代码，请清理后再合并"
    exit 1
fi
```

## 性能考虑

### 分析速度
- **小型项目**（<50 文件）：通常在 1-3 秒内完成
- **中型项目**（50-200 文件）：通常在 3-8 秒内完成
- **大型项目**（>200 文件）：通常在 8-15 秒内完成

### 内存使用
- 内存使用与项目文件数量和依赖关系复杂度相关
- DFS 算法采用递归实现，需要注意栈深度
- 支持大型项目的分块分析

### 优化建议
- 合理设置入口文件，减少不必要的分析范围
- 使用文件排除功能，跳过测试和构建文件
- 定期分析，避免一次性分析大量变更
- 监控分析性能，及时调整配置参数

## 故障排除

### 常见问题

1. **入口文件配置错误**
   ```bash
   # 错误：入口文件不存在
   ./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.entrypoint=nonexistent.ts"

   # 正确：确保入口文件存在
   ./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.entrypoint=src/index.ts"
   ```

2. **循环引用问题**
   ```bash
   # 如果遇到循环引用导致的分析问题
   # 使用单个明确的入口文件
   ./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.entrypoint=src/main.ts"

   # 或启用智能入口检测
   ./analyzer-ts analyze find-unreferenced-files -i /path/to/project -p "unreferenced.include-entry-dirs=true"
   ```

3. **结果分析问题**
   ```bash
   # 查看详细的分类信息
   ./analyzer-ts analyze find-unreferenced-files -i ./src -o result.json
   cat result.json | jq '.suspiciousFiles'

   # 统计文件引用健康度
   ./analyzer-ts analyze find-unreferenced-files -i ./src -o - | jq '.stats'
   ```

### 理解分析结果

1. **真正未引用文件的处理建议**
   - 首先人工确认文件确实没有被使用
   - 检查是否有动态导入或运行时引用
   - 确认文件没有特殊的元数据或配置作用
   - 建议先移动到备份目录，观察一段时间后再删除

2. **可疑文件的检查建议**
   - 配置文件：确认是否在构建工具中被引用
   - 入口文件：确认是否在打包配置中被使用
   - 核心模块：确认是否有运行时动态引用
   - 类型声明：确认是否被其他项目依赖

3. **误报情况的处理**
   - 如果发现有重要文件被误判为未引用
   - 检查文件是否被动态导入
   - 确认文件是否在配置文件中被引用
   - 考虑调整分析器的分类规则

## 扩展和定制

### 添加自定义分类规则
可以通过修改 `classifyFiles` 函数添加项目特定的分类规则：

```go
// 在 classifyFiles 函数中添加项目特定规则
customImportantPatterns := []string{
    "my-custom-module",
    "company-specific",
}
```

### 集成到 CI/CD 流程
```yaml
# GitHub Actions 示例
- name: 检查未引用文件
  run: ./analyzer-ts analyze find-unreferenced-files -i ./src -o unreferenced.json
- name: 如果死代码过多则警告
  if: steps.unreferenced.outputs.dead_files > 10
  run: echo "警告：项目中包含较多死代码，建议清理"
- name: 生成代码健康报告
  run: |
    echo "## 代码健康报告" >> $GITHUB_STEP_SUMMARY
    echo "### 文件引用统计" >> $GITHUB_STEP_SUMMARY
    cat unreferenced.json >> $GITHUB_STEP_SUMMARY
```

### 集成到构建脚本
```bash
# package.json scripts 示例
{
  "scripts": {
    "find-dead-code": "analyzer-ts analyze find-unreferenced-files -i ./src",
    "prebuild": "npm run find-dead-code",
    "analyze-structure": "analyzer-ts analyze find-unreferenced-files -i ./ -p \"unreferenced.include-entry-dirs=true\"",
    "health-check": "npm run find-dead-code && npm run analyze-structure"
  }
}
```

## 版本历史

- **v1.0.0**: 初始版本，基本的死代码检测功能
- **v1.1.0**: 添加智能文件分类和配置识别
- **v1.2.0**: 改进 DFS 算法，支持循环引用处理
- **v1.3.0**: 增强入口文件识别，支持多种配置模式
- **v1.4.0**: 优化输出格式，提供更详细的分析报告

## 相关链接

- [analyzer-ts 项目主页](../../README.md)
- [分析器架构文档](../README.md)
- [Component-Deps 分析器](../component_deps/README.md)
- [Unconsumed 分析器](../unconsumed/README.md)
- [图论算法简介](https://en.wikipedia.org/wiki/Graph_theory)
- [深度优先搜索算法](https://en.wikipedia.org/wiki/Depth-first_search)

---

💡 **提示**: 未引用文件分析是一个强大的代码清理工具，但在删除文件前务必进行人工确认。某些文件可能被动态导入、配置文件引用或在运行时被使用，建议采用"先移动观察，后安全删除"的策略。