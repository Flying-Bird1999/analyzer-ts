# Unconsumed 分析器

## 概述

Unconsumed 分析器是 analyzer-ts 工具中的一个核心分析器插件，专门用于检测 TypeScript 项目中的"死导出"——即那些被导出但在其他任何地方都没有被使用的符号。这些未使用的导出会增加代码包的大小，影响项目的维护性，并且可能表明存在遗留的、不再需要的代码。

## 功能特性

### 🔍 深度检测
- **全面的导出类型支持**：
  - 函数和函数声明（function declarations 和 function expressions）
  - 变量声明（var, const, let）
  - 类声明（class declarations）
  - 接口声明（interface declarations）
  - 类型声明（type declarations）
  - 枚举声明（enum declarations）
  - 默认导出（default exports）
  - 命名导出（named exports）

### 🔗 智能追踪
- **重导出追踪**：能够追踪通过 `export { name } from 'module'` 语法的二次导出
- **JSX 支持**：能够识别 React 组件的导入使用，包括 `<Component />` 的隐式导入
- **别名映射**：支持 `export { OriginalName as NewName }` 的别名导出追踪

### 🛡️ 智能排除
- **自动排除测试文件**：忽略 `*.test.*`、`*.spec.*` 文件
- **排除类型定义**：忽略 `*.d.ts` 文件
- **排除测试工具**：忽略 `__tests__`、`__mocks__` 目录

### 📊 详细报告
- **位置信息**：每个未使用导出的具体行号
- **类型标识**：清楚标识导出的类型（function、const、interface 等）
- **统计信息**：提供扫描文件数、总导出数、未使用数等统计

## 使用方法

### 基本用法
```bash
# 分析项目中的未使用导出
./analyzer-ts analyze unconsumed -i /path/to/project -o /path/to/output

# 结合排除规则
./analyzer-ts analyze unconsumed -i /path/to/project -x "node_modules/**" -x "**/*.test.ts"

# 在 monorepo 项目中使用
./analyzer-ts analyze unconsumed -i /path/to/monorepo -m
```

### 高级用法
```bash
# 结合其他分析器一起使用
./analyzer-ts analyze unconsumed count-any npm-check -i /path/to/project

# 指定输出目录
./analyzer-ts analyze unconsumed -i /path/to/project -o /path/to/output-dir

# 仅分析特定目录
./analyzer-ts analyze unconsumed -i /path/to/project -x "**/node_modules/**"
```

## 输出示例

### 控制台输出（无未使用导出时）
```
✅ 扫描文件 150 个，发现导出 89 个，其中未使用导出 0 个。 没有发现未使用的导出。
```

### 控制台输出（有未使用导出时）
```
⚠️ 扫描文件 150 个，发现导出 89 个，其中未使用导出 5 个。
--------------------------------------------------
  - [function] /src/utils/legacy.ts:42      (formatDate)
  - [const] /src/components/old.tsx:15      (LegacyComponent)
  - [interface] /src/types/deprecated.ts:23  (OldApi)
  - [type] /src/types/deprecated.ts:45      (DeprecatedConfig)
  - [default] /src/utils/helper.ts:89       (default)
--------------------------------------------------
```

### JSON 输出
```json
{
  "findings": [
    {
      "filePath": "/src/utils/legacy.ts",
      "exportName": "formatDate",
      "line": 42,
      "kind": "function"
    },
    {
      "filePath": "/src/components/old.tsx",
      "exportName": "LegacyComponent",
      "line": 15,
      "kind": "const"
    }
  ],
  "stats": {
    "totalFilesScanned": 150,
    "totalExportsFound": 89,
    "unconsumedExportsFound": 5
  }
}
```

## 技术架构

### 工作原理
分析器采用四阶段的算法来识别未使用的导出项：

1. **第一阶段：收集被消费的导出项**
   - 分析所有 `import` 语句
   - 处理 JSX 组件的隐式导入
   - 记录所有被使用的导出项

2. **第二阶段：收集重导出映射关系**
   - 处理 `export { name } from 'module'` 语法
   - 建立别名映射关系

3. **第三阶段：解析重导出关系**
   - 根据映射关系，将被重导出的符号标记为已消费

4. **第四阶段：识别未消费的导出项**
   - 对比所有导出项和已消费导出项的集合
   - 找出差异即为未使用的导出

### 核心组件
```go
// 分析器主体
type Finder struct{}

// 未使用导出项的详细信息
type Finding struct {
    FilePath   string // 文件路径
    ExportName string // 导出名称
    Line       int    // 行号
    Kind       string // 导出类型
}

// 分析结果
type Result struct {
    Findings []Finding    // 未使用导出列表
    Stats   SummaryStats // 统计信息
}
```

## 最佳实践

### 1. 定期清理
```bash
# 在 CI/CD 中集成检查
./analyzer-ts analyze unconsumed -i ./src --exit-on-unconsumed
```

### 2. 分阶段清理
```bash
# 按类型逐步清理
./analyzer-ts analyze unconsumed -i ./src/utils -o utils-unused.json
./analyzer-ts analyze unconsumed -i ./src/components -o components-unused.json
```

### 3. 重构支持
```bash
# 在大型重构前分析依赖关系
./analyzer-ts analyze unconsumed -i ./src -o ./reports/unused-exports-$(date +%Y%m%d).json
```

### 4. 代码审查
```bash
# 在团队代码审查前运行
./analyzer-ts analyze unconsumed -i ./src -o ./reports/pre-review-$(date +%Y%m%d).json
```

## 性能考虑

### 分析速度
- **小型项目**（<100 文件）：通常在 2-3 秒内完成
- **中型项目**（100-1000 文件）：通常在 8-15 秒内完成
- **大型项目**（>1000 文件）：通常在 20-40 秒内完成

### 内存使用
- 分析器需要存储所有导出项和导入项的映射关系
- 内存使用与项目大小成线性关系
- 对于超大型项目，建议分模块进行分析

## 故障排除

### 常见问题

1. **误报问题**
   ```bash
   # 排除特定类型的文件
   ./analyzer-ts analyze unconsumed -i /path/to/project -x "**/legacy/**" -x "**/deprecated/**"
   ```

2. **分析时间过长**
   ```bash
   # 排除不必要的文件
   ./analyzer-ts analyze unconsumed -i /path/to/project -x "node_modules/**" -x "**/*.test.ts" -x "**/*.spec.ts"
   ```

3. **内存不足**
   ```bash
   # 分模块分析
   ./analyzer-ts analyze unconsumed -i /path/to/project/src/utils -o utils-unused.json
   ./analyzer-ts analyze unconsumed -i /path/to/project/src/components -o components-unused.json
   ```

### 理解误报

某些情况可能导致误报，但这是设计上的权衡：

1. **动态导入**：使用 `import('module')` 动态加载的模块
2. **模板字符串**：在模板字符串中使用的模块路径
3. **条件导入**：在某些条件下才会执行的导入
4. **第三方依赖**：被第三方库内部使用的导出

## 扩展和定制

### 添加自定义排除规则
可以通过修改 `isIgnoredFile` 函数来添加自定义的排除逻辑。

### 集成到构建流程
```go
// 在自定义工具中使用
result := &unconsumed.Result{
    // 自定义分析逻辑
}
if len(result.Findings) > 0 {
    log.Fatalf("发现 %d 个未使用的导出", len(result.Findings))
}
```

### 集成到 CI/CD
```yaml
# GitHub Actions 示例
- name: 检查未使用的导出
  run: ./analyzer-ts analyze unconsumed -i ./src -o unused-exports.json
- name: 如果有未使用导出则失败
  if: steps.unconsumed.outputs.unconsumed_count > 0
  run: echo "发现未使用的导出，请清理代码"
```

## 版本历史

- **v1.0.0**: 初始版本，基本的未使用导出检测
- **v1.1.0**: 添加重导出追踪和 JSX 支持
- **v1.2.0**: 优化性能，改进输出格式
- **v1.3.0**: 添加智能排除规则，减少误报

## 相关链接

- [analyzer-ts 项目主页](../../README.md)
- [分析器架构文档](../README.md)
- [Count-Any 分析器](../countAny/README.md)
- [TypeScript 模块系统文档](https://www.typescriptlang.org/docs/handbook/modules.html)

---

💡 **提示**: 此分析器是代码清理和维护的强大工具，建议定期运行以保持代码库的整洁。但在删除导出前，请确保它们确实没有被使用，特别是对于公共 API。