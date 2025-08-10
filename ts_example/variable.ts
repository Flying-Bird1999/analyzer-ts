export const PI: number = 3.14;

let dynamicValue = 'hello world';

// This is a line comment
const appName: string = 'Gemini AI';

export const { name: name2, age } = { name: 'bird', age: 20 };

const [first, second, ...reset] = [1, 2, 3];

let {
  config: { host, port },
  settings: [theme],
} = { config: { host: 'localhost', port: 8080 }, settings: ['dark'] };

const sayHi = () => {
  console.log('sayHi')
}