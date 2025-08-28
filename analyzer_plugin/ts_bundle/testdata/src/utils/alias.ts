// src/utils/alias.ts
export interface AliasUser {
  aliasId: number;
  aliasName: string;
}

export type AliasRole = 'admin' | 'user';

const sideEffect = 'sideEffect';
console.log('Side effect from alias.ts', sideEffect);

export default AliasUser;