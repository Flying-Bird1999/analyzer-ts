/**
 * Symbol 验证测试文件
 * 演示不同作用域的同名变量和同一作用域下的多次引用
 */

// ==================== 全局作用域 ====================
let globalCounter: number = 1;  // 全局变量
const globalConfig = {          // 全局常量对象
  theme: "dark",
  version: "1.0.0"
};

// ==================== 函数作用域 ====================
function outerFunction() {
  // 外层函数的变量
  let counter: number = 10;    // 与全局变量同名，但不同作用域
  const config = {             // 与全局变量同名，但不同作用域
    debug: true
  };

  function innerFunction() {
    // 内层函数的变量
    let counter: number = 100;  // 再次同名，但作用域不同
    const config = {           // 再次同名，但作用域不同
      inner: true
    };

    // 使用不同作用域的变量
    console.log(globalCounter); // 全局变量
    console.log(counter);       // 内层函数的变量
    console.log(config);        // 内层函数的变量
  }

  // 使用外层函数和全局变量
  console.log(globalCounter);   // 全局变量
  console.log(counter);         // 外层函数的变量
  console.log(config);          // 外层函数的变量
}

// ==================== 类作用域 ====================
class SymbolTest {
  private counter: number = 0;   // 类属性，与其他 counter 不同
  public config: object = {};    // 类属性，与其他 config 不同

  constructor() {
    this.counter = 1000;         // 构造函数中的使用
  }

  method() {
    const counter: number = 2000; // 方法内局部变量
    const config = { method: true }; // 方法内局部变量

    // 使用类属性和局部变量
    console.log(this.counter);   // 类属性
    console.log(counter);         // 局部变量
  }
}

// ==================== 模块作用域 ====================
// 导出的变量
export const exportedCounter: number = 100;
export const exportedConfig = { exported: true };

// 未导出的变量
const internalCounter: number = 200;
const internalConfig = { internal: true };

// ==================== 同一作用域多次引用 ====================
function multipleReferences() {
  const sharedVar: string = "shared"; // 在同一作用域被多次使用

  // 多次使用同一个变量
  console.log(sharedVar);     // 第一次使用
  console.log(sharedVar);     // 第二次使用
  console.log(sharedVar);     // 第三次使用

  // 在表达式中多次使用
  const result = sharedVar + sharedVar + sharedVar;

  return result;
}

// ==================== 复杂嵌套结构 ====================
namespace TestNamespace {
  export const namespaceVar: string = "namespace";

  export class NestedClass {
    constructor(private value: string) {}

    getValue(): string {
      return this.value; // this.value 的使用
    }
  }
}

// ==================== 导出 ====================
export {
  outerFunction,
  SymbolTest,
  multipleReferences
};

export default {
  globalCounter,
  globalConfig
};