# analyzer-ts

`analyzer-ts` 是一个使用 Go 语言编写的高性能命令行工具，旨在深度分析和解析 TypeScript 项目。它提供了一系列强大的功能，可以帮助开发者理解代码结构、优化代码质量、清理无用代码，并对依赖项进行管理。

## 架构

`analyzer-ts` 采用插件式架构，其核心思想是将**项目解析**与**代码分析**分离，以实现高性能和高扩展性。

1. **核心解析器 (`analyzer/`)**: 这是工具的基础，负责对整个 TypeScript 项目进行一次性的深度解析，构建所有文件的抽象语法树 (AST) 并提取关键信息。
2. **分析器插件 (`analyzer_plugin/`)**: 包含一系列独立的分析器模块。每个模块都利用核心解析器提供的 AST 数据来执行特定的分析任务（如依赖检查、调用链分析等），而无需重复解析代码。

这种设计使得添加新功能变得非常高效。关于架构的详细说明以及如何开发新的分析器插件，请参阅 **[分析器架构详解](./analyzer_plugin/project_analyzer/README.md)**。

## 安装

我们提供两种安装方式：`go install`（推荐）或从源码构建。

### 方式一：全局安装 (推荐)

这是最简单的安装方式。请确保您已安装 Go (1.24 或更高版本)，然后在终端中运行以下命令：

```bash
go install github.com/Flying-Bird1999/analyzer-ts@latest
```

此命令会自动下载、编译并安装 `analyzer-ts` 到您的 Go 环境中。安装成功后，您可以在系统的任何路径下直接使用 `analyzer-ts` 命令。

### 方式二：从源码构建

如果您想自行编译或修改代码，可以按以下步骤操作：

1. **确保您已经安装了 Go (1.18 或更高版本)。**
2. **克隆此仓库:**

   ```bash
   git clone https://github.com/Flying-Bird1999/analyzer-ts.git
   ```
3. **进入项目目录:**

   ```bash
   cd analyzer-ts
   ```
4. **构建项目:**

   ```bash
   go build -o analyzer-ts
   ```

   此命令会编译源代码，并在项目根目录下生成一个名为 `analyzer-ts` 的可执行文件。

## 命令用法

`analyzer-ts` 的所有功能都通过子命令提供。主要有以下三个命令：

1. `analyze`: 执行一个或多个代码分析任务。
2. `store-db`: 分析项目并将结果存入数据库。
3. `bundle`: 打包 TypeScript 类型声明。

---

### 1. `analyze`

这是最核心的命令，它取代了旧版本中多个独立的分析命令。`analyze` 命令首先会完整解析整个项目，然后根据您提供的参数运行一个或多个分析器。

**用法:**

```bash
./analyzer-ts analyze [分析器名称...] [选项]
```

**分析器 (Analyzers)**

您可以在命令后面跟上一个或多个要执行的分析器的名称。可用的分析器包括：

| 分析器名称                  | 功能描述                                                               |
| --------------------------- | ---------------------------------------------------------------------- |
| `npm-check`               | 检查NPM依赖，识别隐式依赖、未使用和过期依赖。                          |
| `count-any`               | 统计项目中所有 `any` 类型的使用情况。                                |
| `unconsumed`              | 查找项目中所有已导出但从未在别处被导入的符号。                         |
| `find-unreferenced-files` | 在项目中查找所有从未被任何其他文件导入或引用的“孤岛”文件。           |
| `find-callers`            | 查找一个或多个指定文件的所有上游调用方。需要使用 `-p` 参数指定文件。 |
| `structure-simple`        | 输出一个简化的项目整体结构报告，包含了关键的节点信息。                 |
| ...                         | ...                                                                    |

**注意**: 如果不提供任何分析器名称，该命令将只执行项目解析，并将完整的 AST 数据输出到 JSON 文件中。

**选项 (Flags)**

| 标志           | 简写   | 描述                                                                |
| -------------- | ------ | ------------------------------------------------------------------- |
| `--input`    | `-i` | **(必需)** 要分析的 TypeScript 项目的根目录路径。             |
| `--output`   | `-o` | 用于存储生成的 JSON 结果文件的目录路径。 (默认为当前目录)           |
| `--exclude`  | `-x` | 要从分析中排除的文件或目录的 Glob 模式。可多次使用。                |
| `--monorepo` | `-m` | 如果要分析的是一个 monorepo 项目，请设置为 `true`。               |
| `--param`    | `-p` | 为特定分析器传递参数。格式为 `分析器名称.参数名=值`。可多次使用。 |

**示例:**

```bash
# 示例 1: 在一个项目中同时运行 'npm-check' 和 'unconsumed' 分析器
./analyzer-ts analyze npm-check unconsumed -i /path/to/my-project -o /path/to/output

# 示例 2: 查找 'src/api/user.ts' 的调用者
# 注意 'find-callers' 分析器需要一个 'file' 参数
./analyzer-ts analyze find-callers -i /path/to/my-project -p "find-callers.file=src/api/user.ts"

# 示例 3: 运行所有分析器，并排除 node_modules
./analyzer-ts analyze npm-check count-any unconsumed find-unreferenced-files \
  -i /path/to/my-project \
  -o /path/to/output \
  -x "node_modules/**" \
  -x "**/*.test.ts"
```

---

### 2. `store-db`

分析 TypeScript 项目并将完整的分析结果（包括文件、依赖、代码节点等）存储在一个 SQLite 数据库文件中，便于进行复杂的离线查询和历史数据分析。

**用法:**

```bash
./analyzer-ts store-db [选项]
```

**选项 (Flags)**

| 标志           | 简写   | 描述                                                          |
| -------------- | ------ | ------------------------------------------------------------- |
| `--input`    | `-i` | **(必需)** 要分析的 TypeScript 项目的根目录路径。       |
| `--output`   | `-o` | **(必需)** 用于存储生成的 SQLite 数据库文件的目录路径。 |
| `--exclude`  | `-x` | 要从分析中排除的文件或目录的 Glob 模式。可多次使用。          |
| `--monorepo` | `-m` | 如果要分析的是一个 monorepo 项目，请设置为 `true`。         |

**示例:**

```bash
# 分析项目并将结果存入数据库
./analyzer-ts store-db -i /path/to/my-project -o /path/to/output/db
```

---

### 3. `bundle`

从给定的入口文件递归收集所有引用的类型声明，并将它们打包到一个单独的 `.d.ts` 文件中。

**用法:**

```bash
./analyzer-ts bundle [选项]
```

**选项 (Flags)**

| 标志         | 简写   | 描述                                      |
| ------------ | ------ | ----------------------------------------- |
| `--input`  | `-i` | **(必需)** 入口文件的路径。         |
| `--type`   | `-t` | **(必需)** 要分析的根类型名称。     |
| `--output` | `-o` | 输出文件的路径。 (默认为 `./output.ts`) |
| `--root`   | `-r` | 项目的根路径。                            |

**示例:**

```bash
# 将 MyClass 类型及其所有依赖的类型打包
./analyzer-ts bundle -i ./src/index.ts -t MyClass -o ./dist/types.d.ts
```

## 项目结构

- **`main.go`**: 项目的入口点，负责执行根命令。
- **`cmd/`**: 包含使用 `cobra` 库定义的根命令行界面。
- **`analyzer/`**: 包含核心的 TypeScript 解析器 (`parser`) 和项目级分析引擎 (`projectParser`)。
- **`analyzer_plugin/`**: 包含所有可插拔的分析器和功能模块。
  - **`project_analyzer/`**: 核心分析器插件，定义了分析器的标准接口和架构。所有分析命令（如 `npm-check`, `count-any` 等）都在此实现。
  - **`ts_bundle/`**: `bundle` 命令的实现。
