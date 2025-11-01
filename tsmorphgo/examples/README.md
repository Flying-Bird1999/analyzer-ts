# TSMorphGo API 验证示例

这个目录提供了 TSMorphGo 库的实际使用示例，每个 Go 文件都验证了特定的 API 功能。

## ✅ 已验证的核心功能

以下是经过测试验证的核心 API 功能：

### ✅ 完全正常工作的示例 (01-10)

1. **01-basic-analysis.go** - 基础分析功能
   - 项目创建和配置
   - TypeScript 文件发现和解析
   - 基础统计（接口、类型别名、函数、类、变量）
   - AST 遍历和节点访问

2. **02-symbol-analysis.go** - 符号分析功能
   - 符号收集和分类
   - 符号类型识别（interface, class, variable, type 等）
   - 导出状态检查
   - 符号引用计数

3. **03-interface-scan.go** - 接口扫描功能
   - 接口定义扫描
   - 类型别名扫描
   - JSON 格式输出
   - 项目结构分析

4. **04-dependency-check.go** - 依赖检查功能
   - Import/Export 语句解析
   - 第三方依赖识别
   - 模块分类统计
   - 依赖关系分析

5. **05-node-navigation.go** - 节点导航功能
	- 父子节点遍历
	- 祖先节点查找
	- 深度和位置分析
	- QuickInfo 功能测试

6. **06-expression-analysis.go** - 表达式分析功能
	- 调用表达式解析
	- 属性访问分析
	- 二元表达式解析
	- 模式识别和统计

7. **07-type-checking.go** - 类型检查功能
   - 节点类型识别
   - 声明类型转换
   - 类型检查函数使用
   - AsXXX 转换函数测试

8. **08-lsp-service.go** - LSP 服务集成功能
   - LSP 服务创建和管理
   - QuickInfo 功能（类型提示）
   - 原生 QuickInfo 功能（通过 TypeScript 语言服务）
   - 引用查找功能
   - 符号获取功能
   - 上下文管理和资源清理

9. **09-advanced-symbols.go** - 高级符号分析功能
   - 符号层次结构
   - 符号关系分析
   - 引用关系深度分析
   - 模块和复杂度分析

10. **10-quickinfo-test-working.go** - QuickInfo 底层能力验证
    - QuickInfo 功能框架验证
    - 原生 QuickInfo 对比测试
    - 属性级别 QuickInfo 分析
    - 复杂类型引用追踪
    - JSDoc 注释处理框架
    - 显示部件类型分析
    - API 衍生能力验证

## 📊 当前 API 覆盖分析

基于已验证的示例（01-10），TSMorphGo 目前完全支持的核心功能包括：

### ✅ 项目管理
- `tsmorphgo.ProjectConfig` - 项目配置
- `tsmorphgo.NewProject()` - 项目创建
- `project.GetSourceFiles()` - 获取源文件列表

### ✅ 基础 AST 操作
- `node.Kind` - 节点类型识别
- `node.GetText()` - 获取节点文本
- `node.GetStartLineNumber()` - 获取行号
- `node.GetSourceFile()` - 获取源文件
- `sf.ForEachDescendant()` - 遍历节点

### ✅ 符号分析
- `tsmorphgo.GetSymbol()` - 获取符号
- `symbol.GetName()` - 获取符号名称
- `symbol.IsFunction()` - 函数类型检查
- `symbol.IsClass()` - 类类型检查
- `symbol.IsInterface()` - 接口类型检查
- `symbol.IsVariable()` - 变量类型检查
- `symbol.IsTypeAlias()` - 类型别名检查
- `symbol.IsExported()` - 导出状态检查
- `symbol.FindReferences()` - 查找引用

### ✅ Import 解析
- `ast.KindImportDeclaration` - Import 语句识别
- `node.AsImportDeclaration()` - Import 节点转换
- `importDecl.ModuleSpecifier()` - 模块说明符获取
- `importDecl.ImportClause()` - Import 子句获取
- 依赖类型分类（local, third-party, scoped）
- 第三方依赖识别

### ✅ 工具函数
- `tsmorphgo.GetVariableName()` - 获取变量名

### ✅ 表达式分析 (06-expression-analysis.go)
- `ast.KindCallExpression` - 调用表达式识别
- `ast.KindPropertyAccessExpression` - 属性访问表达式
- `ast.KindBinaryExpression` - 二元表达式识别
- `ast.KindObjectLiteralExpression` - 对象字面量识别
- `ast.KindIdentifier` - 标识符识别
- 表达式统计和分类

### ✅ LSP 服务集成 (08-lsp-service.go)
- `lsp.NewService()` - 创建 LSP 服务
- `service.FindReferences()` - 查找引用
- `service.GetQuickInfoAtPosition()` - 获取 QuickInfo（类型提示）
- `service.GetNativeQuickInfoAtPosition()` - 获取原生 QuickInfo
- `quickInfo.DisplayParts` - 结构化显示信息
- `quickInfo.Documentation` - 文档信息
- `lsp.NewServiceForTest()` - 创建测试用 LSP 服务
- `quickInfo.Range` - 文本范围

### ✅ QuickInfo 底层能力验证 (10-quickinfo-test-working.go)
- QuickInfo 功能框架验证
- 原生 QuickInfo 对比测试
- 属性级别 QuickInfo 分析
- 复杂类型引用追踪
- JSDoc 注释处理框架
- 显示部件类型分析
- API 衍生能力验证

### 🔧 需要完善的高级 API
- 详细的类型系统访问（属性签名、类型节点等）
- Import/Export 语句解析
- 节点关系导航（父子关系、祖先链等）
- 复杂的符号关系分析
- JSDoc 注释解析（@apiFieldsDepth, @defaultValue, @internal等）
- API 字段收集和深度过滤算法

## 📁 目录结构

```
examples/
├── demo-react-app/          # 真实的 React TypeScript 测试项目
├── api-examples/            # API 验证示例集合
│   ├── 01-basic-analysis.go     # 基础分析功能
│   ├── 02-symbol-analysis.go    # 符号分析功能
│   ├── 03-interface-scan.go     # 接口扫描功能
│   ├── 04-dependency-check.go   # 依赖检查功能
│   ├── 05-node-navigation.go    # 节点导航功能
│   ├── 06-expression-analysis.go # 表达式分析功能
│   ├── 07-type-checking.go     # 类型检查功能
│   ├── 08-lsp-service.go        # LSP服务功能
│   ├── 09-advanced-symbols.go   # 高级符号分析
│   ├── 10-quickinfo-test-working.go # QuickInfo底层能力验证
└── test.sh                   # 快速测试脚本
```

## 🚀 快速开始

### 运行单个示例

```bash
# 基础分析
cd api-examples
go run 01-basic-analysis.go ../demo-react-app

# 符号分析
go run 02-symbol-analysis.go ../demo-react-app

# 接口扫描
go run 03-interface-scan.go ../demo-react-app

# 依赖检查
go run 04-dependency-check.go ../demo-react-app

# 节点导航
go run 05-node-navigation.go ../demo-react-app

# 表达式分析
go run 06-expression-analysis.go ../demo-react-app

# 类型检查
go run 07-type-checking.go ../demo-react-app

# LSP服务测试
go run 08-lsp-service.go ../demo-react-app

# 高级符号分析
go run 09-advanced-symbols.go ../demo-react-app

# QuickInfo 底层能力验证
go run 10-quickinfo-test-working.go
```

### 运行所有测试

```bash
# 一键运行所有示例
chmod +x test.sh
./test.sh
```

## 📋 示例说明

### 01-basic-analysis.go
**目标**: 验证基础分析功能
- 项目创建和配置
- AST 节点遍历
- 基本统计信息收集

**输出示例**:
```
🔍 基础分析示例 - 项目解析和 AST 遍历
✅ 发现 5 个 TypeScript 文件

📊 项目统计摘要:
  📋 接口数量: 12
  🏷️  类型别名: 3
  ⚡ 函数数量: 5
  🏗️  类数量: 2
  📦 变量数量: 15
```

### 02-symbol-analysis.go
**目标**: 验证符号分析功能
- 符号定义查找
- 引用关系分析
- 导出状态检查

### 03-interface-scan.go
**目标**: 验证接口扫描功能
- 接口信息收集
- 字段详细分析
- JSON 输出格式

**输出**: `interfaces.json`

### 04-dependency-check.go
**目标**: 验证依赖检查功能
- Import/Export 分析
- 第三方依赖识别
- 模块分类统计

### 05-node-navigation.go
**目标**: 验证节点导航功能
- 父子节点遍历
- 祖先节点查找
- 深度和位置分析
- QuickInfo 功能测试

### 06-expression-analysis.go
**目标**: 验证表达式分析功能
- 调用表达式解析
- 属性访问分析
- 二元表达式解析
- 模式识别和统计

### 07-type-checking.go
**目标**: 验证类型检查功能
- 节点类型识别
- 声明类型转换
- 类型检查函数使用
- AsXXX 转换函数测试

### 08-lsp-service.go
**目标**: 验证 LSP 服务集成功能
- LSP 服务创建和管理
- QuickInfo 功能（类型提示）
- 原生 QuickInfo 功能（通过 TypeScript 语言服务）
- 引用查找功能
- 符号获取功能
- 上下文管理和资源清理

### 09-advanced-symbols.go
**目标**: 验证高级符号分析功能
- 符号层次结构
- 符号关系分析
- 引用关系深度分析
- 模块和复杂度分析

### 10-quickinfo-test-working.go
**目标**: 验证 QuickInfo 底层能力，为构建高级 API 分析功能做准备
- QuickInfo 功能框架验证
- 原生 QuickInfo 对比测试
- 属性级别 QuickInfo 分析
- 复杂类型引用追踪
- JSDoc 注释处理框架
- 显示部件类型分析
- API 衍生能力验证

## 🔧 技术栈

- **TSMorphGo**: 核心 TypeScript 分析库
- **typescript-go**: AST 解析和遍历
- **标准库**: JSON, 文件操作, 模板系统

## 💡 扩展建议

1. **添加新示例**: 在 `api-examples/` 目录创建新文件
2. **组合功能**: 结合多个示例创建复杂工具
3. **集成 CI/CD**: 将分析步骤加入构建流程
4. **Web 界面**: 基于 API 分析结果构建可视化界面

## 🐛 问题反馈

遇到问题请访问：
- [GitHub Issues](https://github.com/Flying-Bird1999/analyzer-ts/issues)
- 项目路径: `/Users/bird/Desktop/alalyzer/analyzer-ts/tsmorphgo`

---

✨ 使用 TSMorphGo 构建你的代码分析工具！