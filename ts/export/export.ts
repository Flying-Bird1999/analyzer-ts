export const name = 'John';

export function greet(name: string) {
    console.log('Hello, ' + name);
}

export type Name = {
  name: string;
}

const a = '1'
const b = '2'
const c = '3'

export { a, b, c }


export default function greet2(name: string) {
    console.log('Hello,'+ name);
}