// src/path-alias.ts
import { AliasUser } from '@utils/alias';
import { AliasRole } from '@alias';

export interface PathAliasUser extends AliasUser {
  role: AliasRole;
}