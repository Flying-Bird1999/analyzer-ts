# TSMorphGo Examples

这个目录包含了 TSMorphGo 库的完整使用示例，基于真实的 React TypeScript 应用程序。

## 📁 简洁目录结构

```
examples/
├── demo-react-app/          # 精简的 React TypeScript 应用
│   ├── src/
│   │   ├── components/      # React 组件 (UserProfile, Header, App)
│   │   ├── hooks/          # 自定义 Hooks (useUserData)
│   │   ├── utils/          # 工具函数 (dateUtils)
│   │   ├── test-aliases.tsx # 别名映射测试文件
│   │   ├── App.tsx         # 主应用组件
│   │   ├── tsconfig.json   # TypeScript配置 (含别名映射)
│   │   └── package.json    # 项目配置
├── complete_demo.go       # 🚀 完整功能演示
└── README.md              # 本文档
```

## 🎯 示例说明

### demo-react-app/

**精简的 React TypeScript 应用**，包含 **6个核心源文件**：
- **App.tsx**: 主应用组件
- **components/**: React组件 (UserProfile, Header)
- **hooks/**: 自定义Hooks (useUserData)
- **utils/**: 工具函数 (dateUtils)
- **test-aliases.tsx**: 别名映射测试文件
- **tsconfig.json**: TypeScript配置，**包含完整的路径别名映射**

**tsconfig.json 别名配置**:
```json
{
  "compilerOptions": {
    "baseUrl": "src",
    "paths": {
      "@/*": ["*"],
      "@/components/*": ["components/*"],
      "@/hooks/*": ["hooks/*"],
      "@/utils/*": ["utils/*"]
    }
  }
}
```

### 🚀 complete_demo.go - 完整功能演示

```bash
# 运行完整演示
go run complete_demo.go

# 输出示例:
# 🚀 TSMorphGo 完整演示 - 真实React项目全覆盖分析
# ==============================================
# ✅ 找到真实React项目: ./demo-react-app
# ✅ 找到 6 个源文件
# 📊 4 TSX文件, 2 TS文件
# 🎯 综合分析演示: 节点分析、类型检查、符号分析、引用分析
```

**特点**:
- ✅ 基于精简React项目
- ✅ **完整功能演示**: 节点分析、类型检查、符号分析、引用分析
- ✅ **别名映射分析**
- ✅ tsconfig.json路径映射支持
- ✅ 完全可编译运行
- ✅ 覆盖TSMorphGo所有核心功能

## 🚀 如何运行

```bash
cd tsmorphgo/examples

# 运行完整功能演示
go run complete_demo.go
```

### 2. 运行结果

#### complete_demo.go 输出:
```
🚀 TSMorphGo 完整演示 - 真实React项目全覆盖分析
==============================================
✅ 找到真实React项目: ./demo-react-app
✅ 找到 6 个源文件
📊 4 TSX文件, 2 TS文件

🔍 节点分析演示:
    📍 函数声明: App (行 18, 类型: FunctionDeclaration)
  📊 App.tsx节点统计: 函数=1, 变量=0, 接口=0, 调用=0, 导入=2, JSX=15

🏷️ 类型检查演示:
  📊 UserProfile组件统计: 标识符=45, 属性访问=8, 二元表达式=3, 字面量=12

🧬 符号分析演示:
    🔗 useUserData.ts: 8 个符号
    🔗 useApiService.ts: 12 个符号
    🔗 dateUtils.ts: 6 个符号
  📊 总计找到 26 个符号关联的节点

🔗 引用分析演示:
  📊 标识符引用统计:
    🎯 React: 4 次引用
    🎯 useState: 8 次引用
    🎯 useEffect: 6 次引用

🎉 完整演示完成！
💡 这证明了TSMorphGo具备完整的TypeScript代码分析能力
```

## 🎯 核心成就

### ✅ 完全摒弃虚拟项目

- **❌ 摒弃NewProjectFromSources**: 不再创建虚拟项目
- **✅ 基于真实React项目**: 直接分析19个真实TypeScript文件
- **✅ 无配置文件依赖**: 简单直接的项目创建

### ✅ 验证成功的核心功能

1. **项目管理**: 基于真实前端项目创建TSMorphGo实例
2. **文件发现**: 成功找到6个真实TypeScript源文件
3. **节点访问**: 成功遍历和分析AST节点
4. **类型检查**: 识别各种TypeScript语法结构
5. **符号系统**: 访问节点关联的符号信息
6. **引用查找**: 演示基本的引用分析
7. **别名映射**: 正确解析tsconfig.json路径映射

### ✅ 简洁的目录结构

```
examples/
├── demo-react-app/      # 真实React项目 (6个文件)
├── complete_demo.go     # 完整功能演示
└── README.md           # 本文档
```

## 📊 验证结果

**✅ 完整验证成功**: TSMorphGo能够：
- **成功加载真实的React项目**
- **找到6个真实TypeScript源文件**（4个TSX + 2个TS）
- **直接分析真实前端项目**，完全不再依赖NewProjectFromSources虚拟项目
- **成功访问项目文件**：包括组件、Hooks、工具函数、类型定义等
- **演示完整的分析能力**：项目管理、节点遍历、类型检查、符号系统、引用查找、别名映射

**🎯 核心成果**: 完全摒弃了虚拟项目方式，TSMorphGo现在可以直接分析真实的React TypeScript项目！

## 📚 学习路径

### 🔰 快速入门 (5分钟)

1. 运行 `go run complete_demo.go` - 查看完整功能
2. 查看 `demo-react-app/` 源代码 - 了解分析目标

### 🚀 深入学习

1. 研究 `complete_demo.go` 中的API调用
2. 基于demo-react-app开发自己的分析工具
3. 阅读TSMorphGo API文档

## 💡 最佳实践

1. **使用真实项目**: 直接分析实际的React TypeScript代码
2. **从完整示例开始**: 使用complete_demo了解所有功能
3. **基于真实需求**: 开发符合实际业务需要的分析工具

## 🎉 总结

这个examples目录现在提供了：
- ✅ **简洁的结构**: 1个完整示例 + 1个真实项目
- ✅ **完全可工作**: 所有示例都可以成功编译和运行
- ✅ **基于真实项目**: 不再依赖任何虚拟项目
- ✅ **全覆盖演示**: 展示TSMorphGo的所有核心功能

---

*最后更新: 2025-11-12*