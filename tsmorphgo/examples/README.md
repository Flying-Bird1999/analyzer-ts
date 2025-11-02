# TSMorphGo API 验证示例

这个目录提供了 TSMorphGo 库的完整API验证示例，采用现代化的模块化架构设计。

## 🎯 项目目标

- **API验证**: 在真实React+TypeScript项目中验证TSMorphGo的API准确性
- **性能测试**: 评估API在大型项目中的表现和响应时间
- **功能覆盖**: 确保所有核心API功能正常工作
- **文档完善**: 提供完整的API使用指南和最佳实践

## 🏗️ 新架构概览

### ✅ 6大核心API类别

1. **01-project-api** - 项目管理API
   - 项目创建和配置
   - 源文件发现和管理
   - 项目级统计信息

2. **02-node-api** - 节点操作API
   - AST节点导航
   - 节点属性访问
   - 遍历和搜索

3. **03-symbol-api** - 符号系统API
   - 符号获取和分类
   - 符号关系分析
   - 导出状态检查

4. **04-type-api** - 类型检查API
   - 类型节点识别
   - 类型转换操作
   - 类型系统验证

5. **05-lsp-api** - LSP服务API
   - 语言服务集成
   - QuickInfo功能
   - 诊断信息获取

6. **06-accuracy-validation** - 准确性验证
   - API结果对比
   - 准确率统计
   - 性能基准测试

### 🔬 07-validation-suite - 验证套件

完整的数据驱动验证框架，包含：
- 自动化测试执行
- JSON报告生成
- 性能指标收集
- 错误分析和建议

## 📊 验证结果

基于真实React项目的验证结果：

| API类别 | 准确率 | 处理项目数 | 平均响应时间 |
|---------|--------|-----------|------------|
| Project API | 100.0% | 14个源文件 | <1ms |
| Node API | 99.8% | 22,524个节点 | <10ms |
| Symbol API | 98.6% | 22,508个符号 | <54ms |
| Type API | 99.8% | 22,524个类型节点 | <6ms |
| LSP API | 100.0% | 完整LSP服务 | <131µs |

## 📁 目录结构

```
examples/
├── README.md                           # 项目说明（本文档）
├── api-examples-new/                   # 新架构API示例
│   ├── 01-project-api/                 # 项目管理API
│   ├── 02-node-api/                     # 节点操作API
│   ├── 03-symbol-api/                   # 符号系统API
│   ├── 04-type-api/                     # 类型检查API
│   ├── 05-lsp-api/                      # LSP服务API
│   ├── 06-accuracy-validation/          # 准确性验证
│   ├── 07-validation-suite/            # 验证套件
│   └── API_DOCUMENTATION.md             # 完整API文档
├── demo-react-app/                       # 真实React+TS项目
│   ├── src/                            # 标准React项目结构
│   │   ├── components/                 # React组件
│   │   ├── hooks/                       # 自定义Hooks
│   │   ├── services/                    # API服务
│   │   ├── store/                       # 状态管理
│   │   ├── types/                       # 类型定义
│   │   └── forms/                       # 表单组件
│   └── package.json
└── validation-results/                  # 验证结果输出
```

## 🚀 快速开始

### 1. 运行单个API示例

```bash
# 项目管理API
cd api-examples-new/01-project-api
go run project-creation.go ../../demo-react-app

# 节点操作API
cd ../02-node-api
go run node-navigation.go ../../demo-react-app
go run node-properties.go ../../demo-react-app

# 符号系统API
cd ../03-symbol-api
go run symbol-basics.go ../../demo-react-app
go run symbol-types.go ../../demo-react-app
```

### 2. 运行验证套件（推荐）

```bash
# 运行完整的API验证套件
cd api-examples-new/07-validation-suite
go run -tags validation-suite run-all.go validation-utils.go json-report.go ../../demo-react-app
```

### 3. 查看验证结果

```bash
# 查看详细验证报告
cat ../validation-results/validation-report-*.json
```

## 📋 API使用示例

### 项目管理示例
```go
config := tsmorphgo.ProjectConfig{
    RootPath:         "./demo-react-app",
    IgnorePatterns:   []string{"node_modules", "dist"},
    TargetExtensions: []string{".ts", ".tsx"},
}
project := tsmorphgo.NewProject(config)
sourceFiles := project.GetSourceFiles()
fmt.Printf("发现 %d 个源文件\n", len(sourceFiles))
```

### 符号分析示例
```go
for _, sf := range sourceFiles {
    sf.ForEachDescendant(func(node tsmorphgo.Node) {
        if symbol, ok := tsmorphgo.GetSymbol(node); ok {
            fmt.Printf("符号: %s (类型: %d)\n", symbol.GetName(), symbol.GetFlags())
        }
    })
}
```

## 🔧 技术栈

- **TSMorphGo**: 核心 TypeScript 分析库
- **typescript-go**: AST 解析和遍历
- **标准库**: JSON, 时间处理, 文件操作
- **构建标签**: Go build tags for conditional compilation

## 💡 使用建议

1. **新手用户**: 从 `01-project-api` 开始，了解基本概念
2. **深入分析**: 查看 `03-symbol-api` 和 `04-type-api`
3. **生产验证**: 使用 `07-validation-suite` 进行完整验证
4. **性能优化**: 参考 `validation-results/` 中的性能指标

## 📚 扩展资源

- [完整API文档](api-examples-new/API_DOCUMENTATION.md)
- [验证套件说明](api-examples-new/07-validation-suite/README.md)
- [准确性验证指南](api-examples-new/06-accuracy-validation/README.md)

## 🐛 问题反馈

遇到问题请访问：
- [GitHub Issues](https://github.com/Flying-Bird1999/analyzer-ts/issues)
- 项目路径: `/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo`

---

✨ 使用 TSMorphGo 构建强大的TypeScript代码分析工具！