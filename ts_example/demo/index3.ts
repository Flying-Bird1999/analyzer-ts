import { Age, Age2 } from "./index4.ts";

export type Student = {
  name: string;
  age: Age2;
  teacher: Teacher;
}

export type Teacher = {
  age: Age
}
