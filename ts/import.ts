import Bird from './type2';
import * as allTypes from './type';
import All, { School, School2 } from './school';
import type { CurrentRes } from './type';
import { School as NewSchool } from './school';




export type MediasItem = {
  images: {
    [key in ImagesType]: ImagesAttribute | ImagesAttribute2;
  };
  _id: string;
};

export type TranslationModel = {
  [key in SupportedLanguages]?: ImagesAttribute | ImagesAttribute2;
};
