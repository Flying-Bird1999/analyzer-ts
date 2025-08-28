// src/external.ts
import { ExternalType } from 'some-package';

export interface LocalTypeWithExternal {
  localProp: string;
  externalProp: ExternalType;
}