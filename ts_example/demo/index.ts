import { CusNum, School, School2 } from './index2.ts';

interface A {
  code: number;
  message: string;
}

export interface Class extends A {
  name: string;
  age: number;
  "school": School;
  school2: School2;
  ["class2"]: Class2;
  class8: Class8;
}

export interface Class2 {
  "name_str": string;
  age: number;
  2: CusNum;
}

export type Class3 = Omit<Class2, 'age'>
export type Class4 = Pick<Class2, 'age'>
export type Class5 = Partial<Omit<Class2, 'age'>>

export type Class6 = Class2 & Class3 & Class4 & Class5
export interface Class7 extends Class3 {
  name: string
}


interface Class10 {
  name: string
}

interface Class11 {
  age: number
  school: {
    name: string
  }
}

interface Class12 {
  age: number
  school: {
    name: string
  }
}

export interface Class8 extends Omit<Class12, 'age'>, Pick<Class11, 'age'>, Class10 {
  name: string;
}


export interface Class9 extends Class11, Class12 {
  hi: () => void;
}
