// 匿名 default export 测试
// 这个文件测试 export default 后面是匿名表达式的场景

// 匿名箭头函数作为 default export
export default () => {
  console.log('anonymous function');
};

// 同时导出一个具名的 named export
export const namedExport = () => {
  console.log('named export');
};
