import { School, School2 } from './index2.ts';

interface A {
  code: number;
  message: string;
}

export interface Class extends A {
  name: string;
  age: number;
  school: School;
  school2: School2;
}