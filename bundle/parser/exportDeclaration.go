package parser

// 解析导出模块
// - 默认导出 default: export default Bird;
// - 命名导出 named:
// 		- export { School, School2 as NewSchool2 };
// 		- export type { CurrentRes };
//  	- export const name = "bird"
//  	- export function name() {}
