// src/circular-a.ts
import { CircularBType } from './circular-b';

export interface CircularAType {
  b?: CircularBType;
  aName: string;
}