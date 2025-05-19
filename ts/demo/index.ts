import { School, School2 } from './index2.ts';

interface CurrentRes2 {
  code: number;
  message: string;
}

export interface Class extends CurrentRes2 {
  name: string;
  age: number;
  school: School;
  school2: School2;
}