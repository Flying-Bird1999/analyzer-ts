// src/circular-b.ts
import { CircularAType } from './circular-a';

export interface CircularBType {
  a?: CircularAType;
  bName: string;
}