// Package dependency 实现了检查项目 NPM 依赖健康状况的核心业务逻辑。
//
// 这个包提供了 NPM 依赖分析的核心功能，包括：
// - 隐式依赖（幽灵依赖）检测
// - 未使用依赖识别
// - 依赖版本过时检查
// - 智能过滤和忽略规则
//
// 设计理念：
// 通过静态分析和网络请求，帮助开发者维护健康的依赖关系，
// 减少包体积，提高安全性，避免潜在的问题。
package dependency

// =============================================================================
// 内置模块和忽略规则定义
// =============================================================================

// nodeBuiltInModules 是一个包含了所有 Node.js 内置模块的集合。
//
// 作用说明：
// 在检查隐式依赖时，我们会忽略这些模块，因为它们是 Node.js 运行环境提供的，
// 不需要在 package.json 中声明。将这些模块排除可以避免误报。
//
// 包含的模块类型：
// - 核心模块：fs, path, http, https 等
// - 工具模块：util, events, stream 等
// - 加密模块：crypto, tls 等
// - 进程模块：child_process, cluster 等
var nodeBuiltInModules = map[string]bool{
	"assert": true, "async_hooks": true, "buffer": true, "child_process": true, "cluster": true, "console": true,
	"constants": true, "crypto": true, "dgram": true, "dns": true, "domain": true, "events": true, "fs": true,
	"http": true, "http2": true, "https": true, "inspector": true, "module": true, "net": true, "os": true,
	"path": true, "perf_hooks": true, "process": true, "punycode": true, "querystring": true, "readline": true,
	"repl": true, "stream": true, "string_decoder": true, "timers": true, "tls": true, "trace_events": true,
	"tty": true, "url": true, "util": true, "v8": true, "vm": true, "zlib": true,
}

// devDependencyIgnoreList 是一个包含了常见开发工具和构建相关依赖的集合。
//
// 作用说明：
// 这些工具通常在源码中没有直接的 import/require 语句（而是通过命令行或配置文件使用），
// 或者它们只在开发阶段需要，不应被视为"未使用"的依赖。
// 将它们加入忽略列表，可以避免在"未使用依赖"检查中被错误地报告，从而提高报告的信噪比。
//
// 分类说明：
//
// 1. 代码质量工具：
//    - ESLint、Prettier、StyleLint 等代码检查和格式化工具
//    - Husky、lint-staged 等 Git hooks 工具
//    - CommitLint 等提交信息检查工具
//
// 2. 构建工具：
//    - Webpack、Vite、Rollup 等打包工具
//    - Babel、TypeScript 等编译工具
//
// 3. 测试工具：
//    - Jest、Vitest 等测试框架
//    - TS-Node、Nodemon 等开发环境工具
//
// 4. 框架特定工具：
//    - React Scripts、Create React App 相关工具
//    - Vite 插件和配置
//
// 5. Lint/格式化工具的插件和配置：
//    - ESLint 插件和配置包
//    - TypeScript ESLint 工具
//
// 这个列表会根据社区最佳实践持续更新，以减少误报率。
var devDependencyIgnoreList = map[string]bool{
	// 基础开发工具
	"eslint": true, "prettier": true, "husky": true, "lint-staged": true, "typescript": true,
	"vite": true, "vitest": true, "jest": true, "webpack": true, "stylelint": true, "commitlint": true,
	"@commitlint/cli": true, "@commitlint/config-conventional": true, "webpack-cli": true, "webpack-dev-server": true,

	// 编译和构建工具
	"rollup": true, "@babel/core": true, "@babel/cli": true, "babel-loader": true, "ts-node": true, "nodemon": true,

	// 常用框架/插件相关的开发工具
	"react-scripts": true, "customize-cra": true, "react-app-rewired": true,
	"@vitejs/plugin-react": true, "vite-tsconfig-paths": true,

	// Lint/格式化工具的插件和配置
	"eslint-plugin-react": true, "eslint-plugin-import": true, "eslint-config-airbnb": true, "eslint-config-prettier": true,
	"@typescript-eslint/eslint-plugin": true, "@typescript-eslint/parser": true,
	"stylelint-config-standard": true, "postcss": true,
}
