import Bird from './type2';
import * as allTypes from './type';
import { School, School2 } from './school';
import type { CurrentRes } from './type';
import { School as NewSchool } from './school';


interface Person {
  age: number
}

export interface LinearModel {
  price: number;
}


export interface LinearModel2 {
  price: number;
}

export interface LinearModel3 {
  price: number;
}

export interface LinearModel4 {
  price: number;
  name: string;
}

export interface LinearModel5 {
  price: number;
  age: number;
}

export interface LinearModel6 {
  price: number;
  age: number;
}

interface CurrentRes2 {
  policyDimension: LinearModel5;
}

interface CurrentRes3<T> {
  number: T;
}

interface Class extends CurrentRes2 {
  test: LinearModel4 & LinearModel5;
  test2: LinearModel4 | LinearModel5;
  cccc: CurrentRes3<LinearModel6>
  count?: number[];
  name?: Name;
  student: Person[];
  student2: [LinearModel4, LinearModel2];
  school: {
    school_name: allTypes.StoreDetailReq;
    school_age: {
      color: {
        sc: NewSchool
      }
    };
    linearModel: LinearModel;
  };
  school2: LinearModel | Person;
}

type TestType = {
  test: LinearModel4 & LinearModel5;
  test2: LinearModel4 | LinearModel5;
  cccc: CurrentRes3<LinearModel6>
  count?: number[];
  name?: Name;
  student: Person[];
  student2: [LinearModel4, LinearModel2];
  school: {
    school_name: allTypes.StoreDetailReq;
    school_age: {
      color: {
        sc: NewSchool
      }
    };
    linearModel: LinearModel;
  };
  school2: LinearModel | Person;
}

type Name = string | number;
type Name2 = LinearModel;
type Name3 = LinearModel | Person;
type Name4 = LinearModel & Person;
