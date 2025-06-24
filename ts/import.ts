import Bird from './type2';
import * as allTypes from './type';
import All, { School, School2 } from './school';
import type { CurrentRes } from './type';
import { School as NewSchool } from './school';


type Translations = {
  name: string
}

type Translations_2 ={
  age: number
}

type PersonName = {
  name: Translations["name"] | Translations_2['age']
}