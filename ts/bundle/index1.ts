import { School, School2} from "./index2";
import { Class2 } from "./index3";

interface A {
  code: number;
  message: string;
}

export interface Class extends A {
  name: string;
  age: number;
  // 学校
  "school": School;
  school2: School2;
  ["class2"]: Class2;
  pack: Package;
}

export type Package = {
  name: string;
  version: string;
}
