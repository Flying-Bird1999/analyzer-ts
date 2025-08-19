package project_analyzer

// nodeBuiltInModules 是一个包含了所有 Node.js 内置模块的集合。
// 在检查隐式依赖时，我们会忽略这些模块，因为它们是运行环境提供的，不需要在 package.json 中声明。
var nodeBuiltInModules = map[string]bool{
	"assert": true, "async_hooks": true, "buffer": true, "child_process": true, "cluster": true, "console": true,
	"constants": true, "crypto": true, "dgram": true, "dns": true, "domain": true, "events": true, "fs": true,
	"http": true, "http2": true, "https": true, "inspector": true, "module": true, "net": true, "os": true,
	"path": true, "perf_hooks": true, "process": true, "punycode": true, "querystring": true, "readline": true,
	"repl": true, "stream": true, "string_decoder": true, "timers": true, "tls": true, "trace_events": true,
	"tty": true, "url": true, "util": true, "v8": true, "vm": true, "zlib": true,
}

// devDependencyIgnoreList 是一个包含了常见开发工具的集合。
// 这些工具通常在源码中没有直接的 import/require 语句，而是通过命令行或配置文件使用。
// 将它们加入忽略列表，可以避免在“未使用依赖”检查中被错误地报告，从而提高报告的信噪比。
var devDependencyIgnoreList = map[string]bool{
	"eslint": true, "prettier": true, "husky": true, "lint-staged": true, "typescript": true,
	"vite": true, "vitest": true, "jest": true, "webpack": true, "stylelint": true, "commitlint": true,
	"@commitlint/cli": true, "@commitlint/config-conventional": true, "webpack-cli": true, "webpack-dev-server": true,
	"rollup": true, "@babel/core": true, "@babel/cli": true, "babel-loader": true, "ts-node": true, "nodemon": true,
	// 常用框架/插件相关的开发工具
	"react-scripts": true, "customize-cra": true, "react-app-rewired": true,
	"@vitejs/plugin-react": true, "vite-tsconfig-paths": true,
	// Lint/格式化工具的插件和配置
	"eslint-plugin-react": true, "eslint-plugin-import": true, "eslint-config-airbnb": true, "eslint-config-prettier": true,
	"@typescript-eslint/eslint-plugin": true, "@typescript-eslint/parser": true,
	"stylelint-config-standard": true, "postcss": true,
}
