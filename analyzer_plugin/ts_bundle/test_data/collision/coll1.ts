import { SameName as SameNameFromColl2 } from './coll2';

export interface SameName {
  id: string;
}

export interface Container {
  item1: SameName;
  item2: SameNameFromColl2;
}