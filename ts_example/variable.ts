export const PI: number = 3.14;

let dynamicValue = 'hello world';

// This is a line comment
const appName = 'Gemini AI';

export const { name, age } = { name: 'bird', age: 20 };

const [first, second] = [1, 2];

let {
  config: { host, port },
  settings: [theme],
} = { config: { host: 'localhost', port: 8080 }, settings: ['dark'] };

const sayHi = () => {
  console.log('sayHi')
}