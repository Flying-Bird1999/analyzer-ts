import { CusNum, School, School2 } from './index2.ts';

interface A {
  code: number;
  message: string;
}

export interface Class2 {
  "name_str": string;
  age: number;
  2: CusNum;
}

export interface Class extends A {
  name: string;
  age: number;
  "school": School;
  school2: School2;
  ["class2"]: Class2;
}