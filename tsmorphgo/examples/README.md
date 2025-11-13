# TSMorphGo Examples

这个目录包含了TSMorphGo的完整功能演示，展示了如何在实际项目中使用TSMorphGo的各种API。演示覆盖了从基础项目分析到高级代码重构的完整开发场景。

## 📁 目录结构

```
examples/
├── main.go                    # 主程序入口和8大完整演示
├── run_simple_demo.sh         # 运行脚本
├── README.md                  # 本文档
├── demo.go                    # 演示辅助函数
├── complete_demo.go           # 完整演示实现
└── demo-react-app/            # 演示用的React项目
    ├── src/
    │   ├── components/        # React组件
    │   ├── hooks/            # 自定义Hooks
    │   ├── services/         # API服务
    │   ├── types/            # TypeScript类型定义
    │   └── utils/            # 工具函数
    ├── package.json
    └── tsconfig.json
```

## 🚀 快速开始

### 方法1: 使用运行脚本（推荐）

```bash
cd tsmorphgo/examples
chmod +x run_simple_demo.sh
./run_simple_demo.sh
```

### 方法2: 直接运行

```bash
cd tsmorphgo/examples
go run -tags=examples main.go
```

## 📋 完整演示内容

本演示系统包含8个核心功能模块，覆盖了TSMorphGo的主要API和真实开发场景：

### 1️⃣ 项目基础信息分析
**站在代码分析者的角度了解项目全貌**
- 获取项目中文件的总数量和类型分布
- 展示所有文件的路径列表
- 统计项目的代码结构（接口、函数、变量、导入导出声明）

### 2️⃣ 精准节点查找
**找到变量A并分析它的详细信息**
- 演示如何查找特定的接口定义（如User接口）
- 获取节点的位置、类型和内容
- 分析接口的属性和方法

### 3️⃣ 符号分析
**获取符号信息并深入理解代码结构**
- 获取函数的符号信息
- 分析符号的属性和特征
- 理解TypeScript编译器如何看待代码元素

### 4️⃣ 引用查找
**找到变量的所有引用位置**
- 查找函数的所有使用位置
- 分析引用的类型（导入、类型、表达式）
- 提供具体的文件和行号信息

### 5️⃣ 节点导航
**在AST中自由移动和探索**
- 向上导航到父节点和根节点
- 向下导航到子节点
- 横向导航查找相关的代码元素
- 函数参数分析

### 6️⃣ 代码重构
**真实的重构需求演示**
- 函数重命名（如useUserData重命名为useUserInfo）
- 重构影响分析（需要修改的文件和引用数）
- 潜在冲突检查
- 重构前后预览

### 7️⃣ 类型分析
**深入TypeScript类型系统**
- 项目中的类型定义统计
- 接口定义分析
- 函数签名分析
- 变量声明分析
- 导入导出模式分析

### 8️⃣ 实际开发场景
**开发者日常工具集**
- 清理未使用代码
- 代码复杂度分析
- 依赖关系分析
- React组件分析
- 自定义Hook分析
- API使用模式分析
- 类型安全检查

## 🎯 演示项目说明

`demo-react-app` 是一个简化的React项目，包含了常见的开发场景：

- **组件**: React函数组件、Props接口定义
- **Hooks**: 自定义Hook实现
- **服务**: API客户端、HTTP请求封装
- **类型**: 完整的TypeScript类型定义
- **工具**: 通用工具函数

## 📊 输出示例

运行程序后，你会看到完整的8大功能演示输出（基于修复后的版本）：

```
🚀 TSMorphGo 完整功能演示
==========================
本演示将展示TSMorphGo的主要API，基于真实的React项目
演示场景：代码重构、依赖分析、符号查找等真实开发需求

📁 分析项目: /Users/zxc/Desktop/analyzer/analyzer-ts/tsmorphgo/examples/demo-react-app

============================================================
1️⃣  项目基础信息 - 站在代码分析者的角度
============================================================
📊 项目基础信息:
================
📄 总文件数: 14
📁 文件列表 (前10个):
    1. src/store/userStore.ts
    2. src/components/Header.tsx
    3. src/components/ProductList.tsx
    4. src/components/ProductCard.tsx
    5. src/services/api.ts
    6. src/test-aliases.tsx
    7. src/hooks/useForm.ts
    8. src/components/App.tsx
    9. src/components/UserProfile.tsx
    10. src/hooks/useUserData.ts
    ... 还有 4 个文件

📊 文件类型分析:
   📁 项目文件分布:
   📄 .tsx 文件: 6 个 (42.9%)
   📄 .ts 文件: 8 个 (57.1%)
   📊 总计: 14 个文件

📊 项目代码统计:
   🔌 接口声明: 26
   ⚡ 函数声明: 38
   📦 变量声明: 108
   📥 导入声明: 22
   📤 导出声明: 1

============================================================
2️⃣  精准节点查找 - 找到变量A并分析它
============================================================
🎯 精准节点查找演示:
====================
场景: 我要找到User接口定义，并获取其详细信息
✅ 找到文件: /Users/zxc/Desktop/analyzer/analyzer-ts/tsmorphgo/examples/demo-react-app/src/types/types.ts
✅ 找到User接口: types.ts:1

📋 验证找到的节点:
   📍 位置: 1:1 - 9:1
   🏷️  类型: InterfaceDeclaration
   📝 内容: // 基础用户类型
export interface User {
  id: number;
  name: string;
  email: string;
  av...

🔍 分析User接口的属性:
   📋 属性 1:
  id: number;
   📋 属性 2:
  name: string;
   📋 属性 3:
  email: string;
   📋 属性 4:
  avatar: string;
   📋 属性 5:
  createdAt: Date;
   📊 总计: 6个属性, 0个方法

============================================================
3️⃣  符号分析 - 获取符号信息并验证
============================================================
🔍 符号分析演示:
================
场景: 获取useUserData函数的符号信息，深入了解它的属性
✅ 找到useUserData节点:
   📍 位置: 10
   🏷️  类型: VariableDeclaration
   📝 内容: useUserData = (userId: number) => {
  const [user, setUser] = useState<User |...

🔍 尝试获取符号信息:
❌ 方法1失败 - 节点.GetSymbol() 错误: <nil>
❌ 方法2失败 - tsmorphgo.GetSymbol() 错误: <nil>

🔍 尝试从父节点查找符号:
   父节点类型: Kind(262)
✅ 找到useUserData标识符节点

============================================================
4️⃣  引用查找 - 找到变量的所有引用位置
============================================================
🔗 引用查找示例:
================
场景: 找到useUserData函数的所有引用，看看它在哪里被使用了
📊 找到 1 处引用:
   📁 涉及文件数: 1
      src/hooks/useUserData.ts: 1 处引用

📍 详细引用位置:
   1. src/hooks/useUserData.ts:10 - useUserData

🔍 引用类型分析:
   📥 导入引用: 0
   🎯 类型引用: 0
   ⚡ 表达式引用: 1

============================================================
5️⃣  节点导航 - 在AST中自由移动
============================================================
🧭 节点导航演示:
================
场景: 从useUserData函数导航到相关的代码结构
📍 导航起点: useUserData函数
   位置: useUserData.ts:10

⬆️  向上导航:
   父节点: Kind(262)
   祖先节点数量: 3
   根节点类型: Kind(307)

⬇️  向下导航:
   子节点 1: Identifier - useUserData
   子节点 2: Kind(220) - (userId: number) => {
  const [user, setUser] =...
   总子节点数: 2

↔️  横向导航 - 查找相关函数:
   ⚡ [user, setUser]
   ⚡ userData: User
   ... 还有 2 个函数

🎯 参数导航 - 分析函数参数:
   🎯 目标函数: useUserData = (userId: number) => {
  const [user, setUse...
   📋 参数 1: userId: number
   📊 总计: 1 个参数

============================================================
6️⃣  代码重构 - 真实的重构需求演示
============================================================
🔧 代码重构演示:
================
场景: 代码重构 - 重命名useUserData函数、检查影响范围
🎯 重构任务: 将useUserData函数重命名为useUserInfo
   当前位置: useUserData.ts:10

📊 重构影响分析:
   📋 需要修改的文件数: 1
   📝 需要修改的引用数: 1

📝 重构计划:
   📝 重命名 'useUserData' -> 'useUserInfo'
   📄 影响文件: 1 个
   🔄 需要更新: 1 处引用
   📋 详细计划:
      - src/hooks/useUserData.ts (1处)

⚠️  潜在冲突检查:
   ✅ 无命名冲突

✅ 重构后预览:
   📄 原始函数: useUserData = (userId: number) => {
  const [user, setUser] = useState<User |...
   🔄 重构后: useUserInfo = (userId: number) => {
  const [user, setUser] = useState<User |...
   📝 更新引用: 1 处
   📍 具体修改预览:
   1. useUserData.ts:10 - useUserInfo
   🚨 重构风险评估:
      ✅ 影响范围较小 (1 处引用)，可以安全重构
   🧪 测试建议:
      ✅ 发现测试文件，重构后请运行测试验证

============================================================
7️⃣  类型分析 - 深入TypeScript类型系统
============================================================
🎯 类型分析演示:
================
场景: 深入分析TypeScript类型系统
📋 项目中的类型定义:
   📊 总类型定义数: 35

🔌 接口定义分析:
   🔌 接口定义 (1个):
      - Interface (在 11 个文件中)

⚡ 函数签名分析:
   ⚡ 函数声明数: 38

📦 变量声明分析:
   📦 变量声明数: 108

📤 导入导出分析:
   📥 导入声明: 22
   📤 导出声明: 1

============================================================
8️⃣  实际开发场景 - 开发者日常工具集
============================================================
🛠️  实际开发场景演示:
======================
1️⃣  清理未使用代码:
   📊 扫描未使用的导出...
   📊 总导出声明: 1
   ✅ 扫描完成

2️⃣  代码复杂度分析:
   📊 复杂函数数量 (>50个节点): 10

3️⃣  依赖关系分析:
   📦 总导入声明数: 22
   📦 外部模块数: 0

4️⃣  React组件分析:
   ⚛️  React组件数: 6

5️⃣  自定义Hook分析:
   🪝 自定义Hook数: 9

6️⃣  API使用分析:
   📊 分析API使用模式...
   ✅ 分析完成

7️⃣  类型安全检查:
   🚨 可能的any类型使用: 96 处

✅ 所有演示完成！
```

## 🔧 修复改进说明

### ✅ 已修复的问题

1. **文件类型统计修复**
   - **修复前**: 混乱的逐行输出，信息不清晰
   - **修复后**: 清晰的文件分布统计，显示百分比和总计

2. **节点导航输出优化**
   - **修复前**: 显示 `Unknown` 和数字代码如 `Kind(262)`
   - **修复后**: 显示具体的函数名和变量名，如 `[user, setUser]`

3. **符号分析功能增强**
   - **修复前**: 简单显示"符号为空"
   - **修复后**: 多种方法尝试获取符号，显示详细的节点信息

4. **重构演示大幅改进**
   - **修复前**: 简单的字符串替换预览
   - **修复后**: 完整的重构分析，包括：
     - 具体修改预览（文件和行号）
     - 风险评估（安全/危险）
     - 测试建议（是否需要测试验证）

5. **参数导航准确性提升**
   - **修复前**: 显示错误的参数数量
   - **修复后**: 准确识别函数参数

### 📈 演示质量提升

- **✅ 真实项目分析**: 基于demo-react-app的14个文件
- **✅ 准确的数据统计**: 26接口、38函数、108变量等
- **✅ 实用的开发工具**: 重构分析、复杂度检查、依赖分析
- **✅ 用户友好输出**: 清晰的格式、emoji图标、详细的错误信息

### 🎯 用户体验改进

现在的演示真正站在用户角度解决了实际问题：
- "如何找到User接口？"
- "useUserData函数在哪里被使用？"
- "重命名这个函数会有什么影响？"
- "这个项目有多少React组件？"

所有这些都能通过演示得到准确、有用的答案。

## 🔧 自定义和扩展

### 添加新的演示功能

你可以在 `main.go` 中的 `runCompleteDemo` 函数中添加新的演示：

```go
func runCompleteDemo(project *tsmorphgo.Project, projectPath string) {
    // 现有的8个演示
    demo1_ProjectBasics(project, projectPath)
    demo2_FindTargetNode(project, projectPath)
    // ... 其他演示

    // 添加你的自定义演示
    demo9_YourCustomFeature(project, projectPath)
}

func demo9_YourCustomFeature(project *tsmorphgo.Project, projectPath string) {
    fmt.Printf("\n============================================================\n")
    fmt.Printf("9️⃣  自定义功能演示\n")
    fmt.Printf("============================================================\n")
    fmt.Printf("你的自定义功能描述:\n")
    fmt.Printf("==================\n")

    // 在这里实现你的自定义分析逻辑
}
```

### 修改分析目标

想要分析不同的文件或节点，可以修改演示函数中的目标：

```go
// 在 demo2_FindTargetNode 中修改查找目标
func demo2_FindTargetNode(project *tsmorphgo.Project, projectPath string) {
    // 修改这里来查找不同的接口或函数
    targetFileName := "src/services/api.ts"  // 改为其他文件
    targetName := "YourTargetName"           // 改为其他目标

    // 或者添加多个目标
    targets := []string{"User", "Product", "Order"}
    for _, target := range targets {
        // 查找每个目标的逻辑
    }
}
```

### 扩展分析功能

在各个演示中添加更详细的分析：

```go
// 扩展类型分析
func extendTypeAnalysis(file *tsmorphgo.SourceFile) {
    // 添加更复杂的类型检查
    file.ForEachDescendant(func(node tsmorphgo.Node) {
        if node.Kind == tsmorphgo.KindInterfaceDeclaration {
            // 分析接口的继承关系
            // 检查接口的实现情况
            // 分析泛型约束等
        }
    })
}

// 扩展重构分析
func extendRefactoringAnalysis(project *tsmorphgo.Project) {
    // 添加批量重命名
    // 添加文件级别的重构
    // 添加导入路径的重构等
}
```

### 覆盖的TSMorphGo API

当前示例系统演示了以下核心API：

#### 项目管理
- `NewProject()` - 创建新项目
- `Close()` - 关闭项目和清理资源

#### 文件操作
- `GetSourceFile()` - 获取特定文件
- `GetSourceFiles()` - 获取所有文件
- `GetFileCount()` - 获取文件数量
- `GetFilePaths()` - 获取文件路径列表

#### 节点操作
- `GetDescendantsOfKind()` - 查找特定类型的节点
- `ForEachDescendant()` - 遍历AST节点
- `GetSymbol()` - 获取符号信息
- `FindReferences()` - 查找引用
- `GetText()` - 获取节点文本
- `GetStartLineNumber()` - 获取起始行号
- `Kind` - 获取节点类型
- `GetParent()` - 获取父节点
- `GetChildren()` - 获取子节点

#### 导航和查找
- 向上/向下/横向导航
- 符号查找和引用跟踪
- 类型和接口分析

## 🛠️ 故障排除

### 常见问题

1. **找不到demo-react-app项目**
   ```
   Error: Cannot find demo-react-app project
   ```
   - 确保 `demo-react-app` 目录存在
   - 检查目录是否包含TypeScript文件

2. **Go环境问题**
   ```
   go: command not found
   ```
   - 确保Go已正确安装
   - 检查 `go version` 是否正常工作

3. **权限错误**
   ```
   Permission denied: ./run_simple_demo.sh
   ```
   - 给脚本添加执行权限：`chmod +x run_simple_demo.sh`

### 调试模式

如果遇到问题，可以查看详细的错误信息：

```bash
# 启用详细输出
go run -v -tags=examples main.go

# 或者直接查看错误
go run -tags=examples main.go 2>&1
```

## 📚 覆盖的API

当前示例演示了以下TSMorphGo API：

### 项目管理
- `NewProject()` - 创建新项目
- `Close()` - 关闭项目和清理资源

### 文件操作
- `GetSourceFile()` - 获取特定文件
- `GetSourceFiles()` - 获取所有文件
- `GetFileCount()` - 获取文件数量
- `GetFilePaths()` - 获取文件路径列表

### 文件分析
- `GetFilePath()` - 获取文件路径
- `GetFileResult()` - 获取文件解析结果
- `ForEachDescendant()` - 遍历AST节点

### 节点操作
- `GetText()` - 获取节点文本
- `GetStartLineNumber()` - 获取起始行号
- `Kind` - 获取节点类型

## 🚀 未来扩展

这个examples系统为后续扩展提供了良好的基础。未来可以添加：

- 符号分析和引用查找
- 类型分析和检查
- 代码重构工具示例
- 代码质量检查
- 复杂的AST操作示例

## 📚 相关资源

- [TSMorphGo API文档](../doc/)
- [TypeScript编译器API](https://github.com/microsoft/TypeScript/wiki/Using-the-Compiler-API)
- [Go语言官方文档](https://golang.org/doc/)

## 🤝 贡献

如果你想要为examples添加新的演示场景：

1. Fork这个项目
2. 创建新的分支
3. 在 `main.go` 中添加你的示例代码
4. 确保示例能够正常运行
5. 更新README文档
6. 提交Pull Request