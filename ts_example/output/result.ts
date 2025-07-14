// Auto-generated bundle file
// This file contains bundled type declarations

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index1.ts (original: A)

interface A {
  code: number;
  message: string;
}

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index1.ts (original: Class)

export interface Class extends A {
  name: string;
  age: number;
  // 学校
  school: School;
  school2: School2;
  ['class2']: Class2;
  pack: Package;
}

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index3.ts (original: Class2)
export interface Class2 {
  "name_str": string;
  age: number;
  2: ClassNum;
}

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index3.ts (original: ClassNum)

interface ClassNum {
  value: number;  
  unit: string;
}

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index1.ts (original: Package)

export type Package = {
  name: string;
  version: string;
};

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index2.ts (original: Package)

export type Package_index2 = {
  name: string;
  version: string;
}

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index2.ts (original: School)

export interface School {
  name: string;
  address: string;
}

// From: /Users/bird/Desktop/alalyzer/analyzer-ts/ts_example/bundle/index2.ts (original: School2)

export interface School2 {
  name: string;
  location: string;
  class222: Class2
  pack: Package_index2;
}

