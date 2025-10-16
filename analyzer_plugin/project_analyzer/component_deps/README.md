# 组件依赖分析器 (component-deps)

## 概述

`component-deps` 是一个用于 `analyzer-ts` 的分析器插件，专门用于分析现代前端项目中，其**公共组件（Public Components）**之间的依赖关系。

它通过分析一个或多个指定的**入口文件（Entry Point）**来确定哪些组件是组件库对外暴露的“公共API”，并以此为基础，构建一个以 `package` 分组的、清晰、准确、高层次的依赖图谱，能够完美处理 Monorepo 场景下的跨包依赖。

## 工作原理

分析器采用以“入口文件”为核心的策略，通过多阶段的分析来确保结果的准确性：

1.  **发现入口**：用户通过 `-p "component-deps.entryPoint=..."` 参数提供一个或多个入口文件的 Glob 模式（例如 `packages/*/src/index.ts`）。分析器会首先查找所有匹配该模式的文件，作为分析的起点。

2.  **识别纯类型**：在分析前，分析器会预扫描整个项目，通过递归回溯 `export` 链，精准地找出所有通过 `interface` 和 `type` 定义的“纯类型”符号，并将它们加入一个“过滤清单”。

3.  **解析“组件清单”**：分析器会遍历所有找到的入口文件，解析其中的 `export` 语句。一个符号只有在同时满足 **(a) 遵循帕斯卡命名法（PascalCase）** 并且 **(b) 不在“纯类型过滤清单”中** 这两个条件时，才会被认定为是一个“公共组件”。

4.  **定位源文件与归属包**：对于清单中的每一个公共组件，分析器会追溯其真实的源代码文件路径，并根据项目的 `package.json` 文件，确定该组件所属的 NPM 包名（`packageName`）。

5.  **建立文件归属**：分析器将项目中的每一个文件，根据目录结构，映射到其所属的“公共组件”上。

6.  **构建依赖图谱**：最后，分析器遍历所有文件。如果发现属于 A 组件的一个文件，导入了属于 B 组件的一个文件，程序就会记录下一条“A 依赖 B”的关系，最终形成一个完整的、按包名分组的依赖图谱。

## 使用方法

通过 `analyzer-ts` 的 `analyze` 命令来调用本分析器。

```bash
./analyzer-ts analyze component-deps -i /path/to/your-project -p "component-deps.entryPoint=<path-to-entry-file>"
```

### 示例：分析整个 Monorepo

通过使用 Glob 模式，可以一次性分析一个 Monorepo 中的所有包。

```bash
./analyzer-ts analyze component-deps \
  -i /path/to/monorepo \
  -m \
  -p "component-deps.entryPoint=packages/*/src/index.ts"
```

## 参数说明

*   `component-deps.entryPoint` (**必需**): 
    一个指向入口文件的路径，**支持 Glob 模式**。分析器将依赖此参数来发现所有需要分析的公共组件。

## 输出示例

最终的报告将以 `packageName` 作为第一层级的键，清晰地展示每个包内的组件及其依赖关系。

```json
{
  "packages": {
    "@sl/sc-product": {
      "ProductSetPicker": {
        "sourcePath": ".../Product/src/ProductSetPicker/index.tsx",
        "dependencies": [
          "AddProductSet"
        ]
      },
      "AddProductSet": {
        "sourcePath": ".../Product/src/AddProductSet/index.tsx",
        "dependencies": []
      }
    },
    "@sl/sc-base": {
      "AsyncButton": {
        "sourcePath": ".../Base/src/AsyncButton/index.tsx",
        "dependencies": []
      },
      "CustomerGroupPicker": {
        "sourcePath": ".../Base/src/CustomerGroupPicker/index.tsx",
        "dependencies": [
          "NovaTree"
        ]
      }
    }
  }
}
```
