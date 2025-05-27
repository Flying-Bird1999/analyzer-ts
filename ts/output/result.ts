

export interface Age2 {
  age: number;
}


interface Class11 {
  age: number
  school: {
    name: string
  }
}



interface Class10 {
  name: string
}


interface A {
  code: number;
  message: string;
}


export type School = {
  area: string;
}


export type Student = {
  name: string;
  age: Age2;
  teacher: Teacher;
}
export interface Age {
  age: number;
}


export interface Class2 {
  "name_str": string;
  age: number;
  2: CusNum;
}


export interface CusNum  {
  number: number;
}


export interface Class8 extends Omit<Class12, 'age'>, Pick<Class11, 'age'>, Class10 {
  name: string;
}


interface Class12 {
  age: number
  school: {
    name: string
  }
}


export interface Class extends A {
  name: string;
  age: number;
  "school": School;
  school2: School2;
  ["class2"]: Class2;
  class8: Class8;
}


export type School2 = {
  area2: string;
  stu: Student;
}


export type Teacher = {
  age: Age
}
