import { add } from './math.ts'

/**
 * 这是一个示例函数
 * @param name 要问候的名字
 * @returns 问候信息
 */
function greet(name: string, num: number): string {
  return "Hello, " + name + num;
}

interface Person {
  name: string;
}

const message = greet("World");
console.log(message);  