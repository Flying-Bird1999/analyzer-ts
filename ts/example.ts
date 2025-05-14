// import
import Bird from './type2';
import * as allTypes from './type';
import { School, School2 } from './school';
import type { CurrentRes } from './type';
import { School as NewSchool } from './school';




// interface

interface Class {
  teacher: string | number;
  count?: number[];
  name: Name;
  student: Person[];
}

interface Person {
  age: number
}

// interface smallClass extends Class {
//   ho: boolean;
// }

// export interface LinearModel {
//   /**单价 抽佣场景该值为比例,万分为抽佣,0.05%=5 其他场景为分 必填 */
//   price: number;
// }

// export interface LinearWithFixedModel {
//   /**单价 抽佣场景该值为比例,万分为抽佣,0.05%=5 其他场景为分 必填 */
//   price: number;
//   /**固额 必填 */
//   fixedPrice: number;
// }





// type
export type Name = {
  age: number;
  name: string;
  bird: Bird;
  class: Class;
  school: School;
  school2: NewSchool;
  storeDetailReq: allTypes.StoreDetailReq;
  currentRes: CurrentRes
}
