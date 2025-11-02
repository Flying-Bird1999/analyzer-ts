# TSMorphGo 验证套件

## 概述

验证套件是一个完整的 TSMorphGo API 准确性验证框架，用于验证 API 在真实 React + TypeScript 项目中的表现和准确性。

## 功能特性

- 🎯 **多类别验证**: 支持项目、节点、符号、类型、LSP 和准确性验证
- 📊 **详细报告**: 生成 JSON 格式的详细验证报告
- ⏱️ **性能监控**: 跟踪每个验证的执行时间和性能指标
- 🚀 **并行执行**: 支持并发执行多个验证测试
- 📈 **健康评估**: 提供整体 API 健康度评估和改进建议

## 目录结构

```
07-validation-suite/
├── README.md                   # 本文档
├── run-all.go                  # 主验证运行器
├── validation-utils.go         # 验证工具函数
├── json-report.go             # JSON 报告生成器
└── validation-results/         # 验证结果输出目录
    ├── validation-report-*.json      # 主验证报告
    ├── category-*-report-*.json      # 分类报告
    └── summary-report-*.json         # 摘要报告
```

## 快速开始

### 1. 运行完整验证套件

```bash
# 进入验证套件目录
cd tsmorphgo/examples/api-examples-new/07-validation-suite

# 运行完整验证套件
go run -tags validation-suite run-all.go ../../demo-react-app
```

### 2. 运行特定验证类别

```bash
# 只运行项目API验证
go run -tags validation-suite run-all.go ../../demo-react-app project-api

# 只运行节点API验证
go run -tags validation-suite run-all.go ../../demo-react-app node-api

# 只运行符号API验证
go run -tags validation-suite run-all.go ../../demo-react-app symbol-api

# 只运行类型API验证
go run -tags validation-suite run-all.go ../../demo-react-app type-api

# 只运行LSP API验证
go run -tags validation-suite run-all.go ../../demo-react-app lsp-api

# 只运行准确性验证
go run -tags validation-suite run-all.go ../../demo-react-app accuracy-validation
```

### 3. 运行多个验证类别

```bash
# 运行项目、节点和符号验证
go run -tags validation-suite run-all.go ../../demo-react-app project-api node-api symbol-api
```

## 验证类别

### 1. Project API (项目API)
- **文件**: `01-project-api/project-creation.go`
- **功能**: 验证项目创建、配置和基础API功能
- **验证内容**:
  - 基础项目创建
  - 高级项目配置
  - 内存源码项目创建
  - 项目配置验证
  - 项目API方法验证
  - 错误处理验证
  - 性能基准测试

### 2. Node API (节点API)
- **文件**: `02-node-api/node-navigation.go`, `02-node-api/node-properties.go`
- **功能**: 验证AST节点操作和属性API
- **验证内容**:
  - 节点发现和导航
  - 父子节点关系
  - 祖先节点遍历
  - 节点属性验证
  - 条件节点查找
  - 节点深度计算
  - 性能基准测试

### 3. Symbol API (符号API)
- **文件**: `03-symbol-api/symbol-basics.go`, `03-symbol-api/symbol-types.go`
- **功能**: 验证符号系统API
- **验证内容**:
  - 符号发现和提取
  - 符号类型识别
  - 符号导出状态
  - 符号详细信息
  - 符号声明节点
  - 符号字符串表示
  - 符号数量统计

### 4. Type API (类型API)
- **文件**: `04-type-api/type-checking.go`, `04-type-api/type-conversion.go`
- **功能**: 验证类型检查和转换API
- **验证内容**:
  - IsXXX 类型检查函数
  - AsXXX 类型转换函数
  - 类型覆盖度分析
  - 准确性验证
  - 错误处理测试
  - 性能测试

### 5. LSP API (LSP服务API)
- **文件**: `05-lsp-api/lsp-service.go`, `05-lsp-api/quickinfo-advanced.go`
- **功能**: 验证LSP服务集成
- **验证内容**:
  - LSP服务创建
  - 服务生命周期管理
  - QuickInfo查询
  - 并发操作安全
  - 错误处理
  - 性能基准测试

### 6. Accuracy Validation (准确性验证)
- **文件**: `06-accuracy-validation/symbol-accuracy.go`
- **功能**: 数据驱动的准确性验证
- **验证内容**:
  - 预期vs实际结果对比
  - 准确性指标计算
  - 测试用例管理
  - 详细错误分析
  - 统计报告生成

## 验证报告

### 输出文件

验证完成后，会在 `validation-results/` 目录下生成以下文件：

1. **主验证报告** (`validation-report-<timestamp>.json`)
   - 包含完整的验证结果
   - 项目信息和配置
   - 详细的分析和建议

2. **分类报告** (`category-<category>-<timestamp>.json`)
   - 按API类别分类的详细结果
   - 包含每个测试的详细数据

3. **摘要报告** (`summary-report-<timestamp>.json`)
   - 验证结果摘要
   - 健康度评估
   - 改进建议

### 报告结构

```json
{
  "metadata": {
    "reportId": "val-1234567890",
    "generatedAt": "2024-01-01T12:00:00Z",
    "totalTests": 50,
    "testDuration": "30.5s"
  },
  "suite": {
    "name": "TSMorphGo API Validation",
    "tests": [
      {
        "name": "项目API验证",
        "category": "project-api",
        "status": "passed",
        "duration": "2.3s",
        "metrics": {
          "totalItems": 100,
          "accuracyRate": 98.5
        }
      }
    ]
  },
  "analysis": {
    "overallHealth": "excellent",
    "passRate": 95.0,
    "recommendations": [
      {
        "priority": "medium",
        "category": "performance",
        "title": "优化LSP性能",
        "action": "减少QuickInfo查询时间"
      }
    ]
  }
}
```

## 配置选项

### 环境变量

```bash
# 设置输出目录
export VALIDATION_OUTPUT_DIR=./custom-output

# 启用详细日志
export VALIDATION_VERBOSE=true

# 设置超时时间
export VALIDATION_TIMEOUT=60s

# 选择测试类别
export VALIDATION_CATEGORIES=project-api,node-api,symbol-api
```

### 命令行参数

```bash
go run -tags validation-suite run-all.go [项目路径] [验证类别...]

# 示例：详细输出模式
go run -tags validation-suite run-all.go ../demo-react-app --verbose

# 示例：自定义输出目录
go run -tags validation-suite run-all.go ../demo-react-app --output-dir ./results

# 示例：设置超时
go run -tags validation-suite run-all.go ../demo-react-app --timeout 2m
```

## 使用示例

### 验证现有项目

```bash
# 验证React项目
go run -tags validation-suite run-all.go /path/to/react-project

# 验证TypeScript库
go run -tags validation-suite run-all.go /path/to/typescript-library

# 验证Monorepo项目
go run -tags validation-suite run-all.go /path/to/monorepo
```

### 集成到CI/CD

```yaml
# GitHub Actions 示例
name: API Validation
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.19'

    - name: Run API Validation
      run: |
        cd tsmorphgo/examples/api-examples-new/07-validation-suite
        go run -tags validation-suite run-all.go ../../demo-react-app

    - name: Upload Results
      uses: actions/upload-artifact@v3
      with:
        name: validation-results
        path: validation-results/
```

## 最佳实践

### 1. 项目选择
- 选择包含多种TypeScript文件的真实项目
- 项目应该包含接口、类、类型别名等复杂类型
- 避免使用空项目或过于简单的项目

### 2. 验证频率
- 代码变更后：运行完整验证套件
- 日常开发：运行相关类别的验证
- 发布前：运行完整的准确性验证

### 3. 结果分析
- 关注失败率较高的API类别
- 分析性能瓶颈
- 根据建议进行改进

### 4. 报告管理
- 定期归档验证结果
- 跟踪API改进趋势
- 使用报告作为质量度量指标

## 故障排除

### 常见问题

**Q: 验证套件找不到源文件**
A: 确保项目路径正确，且包含 `.ts` 或 `.tsx` 文件

**Q: 验证超时**
A: 增加超时时间或减少验证类别数量

**Q: 报告生成失败**
A: 检查输出目录权限和磁盘空间

**Q: LSP服务验证失败**
A: 确保 TypeScript 和相关依赖已正确安装

### 调试技巧

1. **启用详细日志**
   ```bash
   go run -tags validation-suite run-all.go ../demo-react-app --verbose
   ```

2. **运行单个验证类别**
   ```bash
   go run -tags validation-suite run-all.go ../demo-react-app project-api
   ```

3. **检查验证结果**
   ```bash
   cat validation-results/validation-report-*.json
   ```

## 贡献指南

### 添加新的验证类别

1. 在相应的目录下创建验证文件
2. 实现验证逻辑
3. 在 `run-all.go` 中注册验证函数
4. 更新文档和README

### 修复问题

1. 识别问题所在的验证类别
2. 修复验证逻辑
3. 测试修复结果
4. 提交PR并描述修复内容

## 版本历史

### v1.0.0 (当前版本)
- 初始版本
- 支持6个验证类别
- 生成详细JSON报告
- 并行执行支持

## 许可证

本项目采用 MIT 许可证。