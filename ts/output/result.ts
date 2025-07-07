// Auto-generated bundle file
// This file contains bundled type declarations

// Interfaces
// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index1.ts (original: A)


interface A {
  code: number;
  message: string;
}

// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index1.ts (original: Class)


export interface Class extends A {
  name: string;
  age: number;
  // 学校
  "school": School;
  school2: School2;
  ["class2"]: Class2;
  pack: Package;
}

// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index3.ts (original: Class2)

export interface Class2 {
  "name_str": string;
  age: number;
  2: ClassNum;
}

// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index3.ts (original: ClassNum)


interface ClassNum {
  value: number;  
  unit: string;
}

// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index2.ts (original: School)


export interface School {
  name: string;
  address: string;
}

// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index2.ts (original: School2)


export interface School2 {
  name: string;
  location: string;
  class222: Class2
  pack: Package_index2;
}

// Type Aliases
// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index1.ts (original: Package)


export type Package = {
  name: string;
  version: string;
}

// From: /Users/zxc/Desktop/analyzer-ts/ts/bundle/index2.ts (original: Package)


export type Package_index2 = {
  name: string;
  version: string;
}

