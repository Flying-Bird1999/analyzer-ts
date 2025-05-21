

export type School2 = {
  area2: string;
  stu: Student;
}


export type Student = {
  name: string;
  age: Age2;
  teacher: Teacher;
}


export interface Age2 {
  age: number;
}


export interface CusNum  {
  number: number;
}


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
}


export type Teacher = {
  age: Age
}
export interface Age {
  age: number;
}


export interface Class2 {
  "name_str": string;
  age: number;
  2: CusNum;
}


export type School = {
  area: string;
}
