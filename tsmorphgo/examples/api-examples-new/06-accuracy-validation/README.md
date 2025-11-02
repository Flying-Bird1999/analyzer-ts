# 准确性验证模块

## 概述

准确性验证模块是 TSMorphGo API 准确性验证的核心组件，专门用于验证各种 API 在真实项目中的准确性和可靠性。

## 功能特性

- 🎯 **数据驱动验证**: 基于预期vs实际结果对比的验证框架
- 📊 **详细分析**: 提供失败原因分析和性能指标
- 🔍 **多维度验证**: 支持符号和类型 API 的准确性验证
- 💡 **改进建议**: 基于验证结果提供具体的改进建议
- 📈 **趋势分析**: 跟踪 API 准确性的改进趋势

## 目录结构

```
06-accuracy-validation/
├── README.md                      # 本文档
├── symbol-accuracy.go            # 符号 API 准确性验证
├── type-accuracy.go             # 类型 API 准确性验证
└── validation-results/           # 验证结果输出目录
    ├── symbol-accuracy-results.json    # 符号准确性结果
    ├── type-accuracy-results.json     # 类型准确性结果
    ├── type-accuracy-report.json      # 类型准确性统计报告
```

## 快速开始

### 1. 运行符号准确性验证

```bash
# 进入准确性验证目录
cd tsmorphgo/examples/api-examples-new/06-accuracy-validation

# 运行符号准确性验证
go run -tags accuracy-validation symbol-accuracy.go ../../demo-react-app
```

### 2. 运行类型准确性验证

```bash
# 运行类型准确性验证
go run -tags accuracy-validation type-accuracy.go ../../demo-react-app
```

### 3. 通过验证套件运行

```bash
# 通过验证套件运行所有准确性验证
cd ../07-validation-suite
go run -tags validation-suite run-all.go ../../demo-react-app accuracy-validation
```

## 验证模块详解

### 1. 符号准确性验证 (`symbol-accuracy.go`)

**功能**: 验证符号系统 API 的准确性

**验证内容**:
- 符号名称提取准确性
- 符号类型识别准确性
- 符号导出状态检测准确性
- 符号位置信息准确性
- 符号声明类型判断准确性

**测试用例结构**:
```go
type SymbolAccuracyTestCase struct {
    Name        string                 `json:"name"`        // 测试用例名称
    Description string                 `json:"description"` // 测试用例描述
    Input       SymbolAccuracyInput   `json:"input"`       // 输入参数
    Expected    SymbolAccuracyExpected `json:"expected"`    // 期望结果
}

type SymbolAccuracyInput struct {
    FilePath string `json:"filePath"` // 文件路径
    Line     int    `json:"line"`     // 行号
    Char     int    `json:"char"`     // 列号
    Symbol   string `json:"symbol"`   // 期望的符号名称
}

type SymbolAccuracyExpected struct {
    Name        string   `json:"name"`        // 期望的符号名称
    Kind        string   `json:"kind"`        // 期望的符号类型
    IsExported  bool     `json:"isExported"`  // 期望的导出状态
    Line        int      `json:"line"`        // 期望的行号
    Declaration  string   `json:"declaration"` // 期望的声明类型
}
```

**运行示例**:
```bash
go run -tags accuracy-validation symbol-accuracy.go ../../demo-react-app
```

**预期输出**:
```
🎯 符号 API 准确性验证
================================
✅ 加载 3 个测试用例
🔧 创建项目...
✅ 项目创建成功，发现 25 个源文件

🧪 执行准确性验证...
================================

🔍 [1/3] 测试: User interface symbol
   描述: 验证 User 接口的符号信息
   位置: src/types.ts:8:1
   ✅ 通过

🔍 [2/3] 测试: UserRole type alias symbol
   描述: 验证 UserRole 类型别名的符号信息
   位置: src/types.ts:29:1
   ✅ 通过

🔍 [3/3] 测试: UserService class symbol
   描述: 验证 UserService 类的符号信息
   位置: src/services/api.ts:1:1
   ✅ 通过

📊 验证结果汇总
================================
   总测试数: 3
   通过数量: 3
   失败数量: 0
   成功率: 100.0%

🎉 符号 API 准确性验证通过！成功率 100.0%
   API 准确性达到优秀水平
```

### 2. 类型准确性验证 (`type-accuracy.go`)

**功能**: 验证类型检查和转换 API 的准确性

**验证内容**:
- IsXXX 类型检查函数准确性
- AsXXX 类型转换函数准确性
- 类型名称提取准确性
- 类型文本表示准确性
- 类型标志信息准确性

**测试用例结构**:
```go
type TypeAccuracyTestCase struct {
    Name        string              `json:"name"`        // 测试用例名称
    Description string              `json:"description"` // 测试用例描述
    Input       TypeAccuracyInput   `json:"input"`       // 输入参数
    Expected    TypeAccuracyExpected `json:"expected"`    // 期望结果
}

type TypeAccuracyInput struct {
    FilePath      string `json:"filePath"`      // 文件路径
    Line          int    `json:"line"`          // 行号
    Char          int    `json:"char"`          // 列号
    ExpectedKind  string `json:"expectedKind"`  // 期望的节点类型
    TypeCheckType string `json:"typeCheckType"` // 类型检查类型
}

type TypeAccuracyExpected struct {
    IsTypeResult     bool   `json:"isTypeResult"`     // IsXXX 函数的期望结果
    AsTypeResult     bool   `json:"asTypeResult"`     // AsXXX 函数的期望结果
    ExpectedTypeName string `json:"expectedTypeName"` // 期望的类型名称
    ExpectedTypeText string `json:"expectedTypeText"` // 期望的类型文本
}
```

**运行示例**:
```bash
go run -tags accuracy-validation type-accuracy.go ../../demo-react-app
```

**预期输出**:
```
🎯 类型 API 准确性验证
================================
✅ 加载 8 个测试用例
🔧 创建项目...
✅ 项目创建成功，发现 25 个源文件

🧪 执行类型准确性验证...
================================

🔍 [1/8] 测试: Interface IsFunction test
   描述: 验证接口节点的 IsFunction() 函数
   位置: src/types.ts:1:1
   类型检查: IsFunction
   ✅ 通过 (耗时: 125μs)

🔍 [2/8] 测试: TypeAlias IsTypeAlias test
   描述: 验证类型别名节点的 IsTypeAlias() 函数
   位置: src/types.ts:15:1
   类型检查: IsTypeAlias
   ✅ 通过 (耗时: 98μs)

📊 验证结果汇总
================================
   总测试数: 8
   通过数量: 8
   失败数量: 0
   成功率: 100.0%
   总耗时: 2.145ms
   平均耗时: 268μs

⏱️ 性能分析:
------------------------------
   平均执行时间: 268μs
   最小执行时间: 45μs
   最大执行时间: 892μs

   性能分布:
     正常 (<100μs): 3 次
     慢 (100-500μs): 4 次
     很慢 (>500μs): 1 次

🎉 类型 API 准确性验证通过！成功率 100.0%
   API 准确性达到优秀水平
```

## 验证结果分析

### 1. 验证报告

验证完成后会生成详细的 JSON 报告，包含：

- **符号准确性结果** (`symbol-accuracy-results.json`)
  - 每个测试用例的详细结果
  - 期望vs实际结果对比
  - 差异分析和错误信息

- **类型准确性结果** (`type-accuracy-results.json`)
  - IsXXX 和 AsXXX 函数的验证结果
  - 性能指标和执行时间
  - 类型信息提取详情

- **统计报告** (`type-accuracy-report.json`)
  - 整体准确性统计
  - 性能分析报告
  - 改进建议

### 2. 失败原因分析

当验证失败时，系统会自动分析失败原因：

```
🔍 失败原因分析:
------------------------------
   名称错误: 1 次
   类型错误: 2 次
   导出状态错误: 0 次
   行号错误: 0 次
   其他错误: 0 次

   💡 建议：检查符号类型判断逻辑
```

### 3. 性能分析

系统会提供详细的性能分析：

```
⏱️ 性能分析:
------------------------------
   平均执行时间: 268μs
   最小执行时间: 45μs
   最大执行时间: 892μs

   性能分布:
     正常 (<100μs): 3 次
     慢 (100-500μs): 4 次
     很慢 (>500μs): 1 次

   💡 建议：存在性能瓶颈，需要优化慢查询
```

## 自定义测试用例

### 1. 创建新的测试用例

要添加新的准确性测试用例，修改相应的验证文件中的测试用例加载函数：

```go
// 在 symbol-accuracy.go 中添加符号测试用例
func loadSymbolAccuracyTestCases() ([]SymbolAccuracyTestCase, error) {
    return []SymbolAccuracyTestCase{
        {
            Name:        "Custom interface test",
            Description: "验证自定义接口的符号信息",
            Input: SymbolAccuracyInput{
                FilePath: "src/custom.ts",
                Line:     10,
                Char:     1,
                Symbol:   "CustomInterface",
            },
            Expected: SymbolAccuracyExpected{
                Name:       "CustomInterface",
                Kind:       "interface",
                IsExported: true,
                Line:       10,
                Declaration: "InterfaceDeclaration",
            },
        },
        // ... 更多测试用例
    }, nil
}
```

```go
// 在 type-accuracy.go 中添加类型测试用例
func loadTypeAccuracyTestCases() ([]TypeAccuracyTestCase, error) {
    return []TypeAccuracyTestCase{
        {
            Name:        "Custom class IsClass test",
            Description: "验证自定义类节点的 IsClass() 函数",
            Input: TypeAccuracyInput{
                FilePath:      "src/custom.ts",
                Line:          15,
                Char:          1,
                ExpectedKind:  "ClassDeclaration",
                TypeCheckType: "IsClass",
            },
            Expected: TypeAccuracyExpected{
                IsTypeResult:     true,
                AsTypeResult:     true,
                ExpectedTypeName:  "ClassDeclaration",
                ExpectedTypeText: "class",
            },
        },
        // ... 更多测试用例
    }, nil
}
```

### 2. 从 JSON 文件加载测试用例

为了更好的可维护性，可以将测试用例存储在外部 JSON 文件中：

```go
func loadTestCasesFromJSON(filename string, testCases interface{}) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("读取测试用例文件失败: %w", err)
    }

    if err := json.Unmarshal(data, testCases); err != nil {
        return fmt.Errorf("解析测试用例JSON失败: %w", err)
    }

    return nil
}
```

### 3. 测试用例最佳实践

- **覆盖性**: 确保测试用例涵盖所有主要的 API 功能
- **真实性**: 使用真实的 TypeScript 代码作为测试基础
- **边界情况**: 包含边界情况和异常情况的测试
- **可重复性**: 测试用例应该是确定性的，可以重复执行
- **清晰性**: 为每个测试用例提供清晰的名称和描述

## 集成到 CI/CD

### 1. GitHub Actions 集成

```yaml
name: Accuracy Validation
on: [push, pull_request]

jobs:
  accuracy-check:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.19'

    - name: Run Symbol Accuracy Validation
      run: |
        cd tsmorphgo/examples/api-examples-new/06-accuracy-validation
        go run -tags accuracy-validation symbol-accuracy.go ../../demo-react-app

    - name: Run Type Accuracy Validation
      run: |
        cd tsmorphgo/examples/api-examples-new/06-accuracy-validation
        go run -tags accuracy-validation type-accuracy.go ../../demo-react-app

    - name: Upload Results
      uses: actions/upload-artifact@v3
      with:
        name: accuracy-results
        path: validation-results/
```

### 2. 质量门禁

可以在 CI 流程中设置准确性阈值：

```bash
#!/bin/bash

# 运行准确性验证
cd tsmorphgo/examples/api-examples-new/06-accuracy-validation

# 执行验证并提取成功率
success_rate=$(go run -tags accuracy-validation symbol-accuracy.go ../../demo-react-app 2>/dev/null | grep "成功率:" | awk '{print $2}' | tr -d '%')

# 检查是否达到质量门禁
if (( $(echo "$success_rate >= 90.0" | bc -l) )); then
    echo "✅ 准确性验证通过: $success_rate%"
    exit 0
else
    echo "❌ 准确性验证失败: $success_rate% (期望 >= 90.0%)"
    exit 1
fi
```

## 故障排除

### 常见问题

**Q: 验证找不到指定的节点**
```
❌ 未找到指定位置的节点: src/types.ts:8:1
```
A: 确保文件路径正确，行号和列号准确，检查项目是否包含该文件

**Q: 验证结果不匹配期望**
```
❌ 失败
   差异: 名称="ExpectedName", 类型="interface", 行号=8
```
A: 检查测试用例中的期望值是否正确，分析实际值以确定问题所在

**Q: 验证超时**
A: 检查项目大小，考虑增加超时时间或优化验证逻辑

### 调试技巧

1. **启用详细输出**
   ```bash
   # 某些验证支持详细输出模式
   go run -tags accuracy-validation symbol-accuracy.go ../../demo-react-app --verbose
   ```

2. **检查验证结果文件**
   ```bash
   cat validation-results/symbol-accuracy-results.json
   ```

3. **运行单个测试用例**
   修改测试用例加载函数，只保留一个测试用例进行调试

4. **验证项目配置**
   确保项目配置正确，能够找到所有必要的源文件

### 性能优化

1. **批量处理**: 将多个验证请求合并为批量操作
2. **缓存机制**: 对重复的验证结果进行缓存
3. **并行执行**: 利用并发执行多个验证测试
4. **延迟加载**: 只在需要时加载和处理文件

## 版本历史

### v1.0.0 (当前版本)
- 初始版本
- 支持符号 API 准确性验证
- 支持类型 API 准确性验证
- 详细的失败原因分析
- 性能监控和报告

## 贡献指南

### 添加新的验证模块

1. 在 `06-accuracy-validation/` 目录下创建新的验证文件
2. 实现相应的测试用例结构
3. 添加验证逻辑和结果分析
4. 更新文档和 README

### 报告问题

1. 创建包含以下信息的 Issue：
   - 问题描述
   - 重现步骤
   - 期望结果 vs 实际结果
   - 相关的验证报告

## 许可证

本项目采用 MIT 许可证。