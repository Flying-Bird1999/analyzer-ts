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


export type Teacher = {
  age: Age
}
export interface Age {
  age: number;
}


export interface Class extends A {
  name: string;
  age: number;
  school: School;
  school2: School2;
}


interface A {
  code: number;
  message: string;
}


export type School = {
  area: string;
}