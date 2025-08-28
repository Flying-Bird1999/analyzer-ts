// src/advanced.ts
import { AliasUser as AliasUserType, AliasRole as AliasRoleType } from './utils/alias';
import AliasDefault, { AliasUser } from './utils/alias';
import { default as AliasDefault2 } from './utils/alias';
import type { AliasRole } from './utils/alias';
import './utils/alias'; // Side effect import

// Use dynamic import (commented out as it's not executed in tests)
// const dynamicImport = () => import('./utils/user');

// Type using import type
export type AdvancedRole = AliasRole;

// Type using aliased import
export interface AdvancedUser extends AliasUserType {
  role: AliasRoleType;
}

// Type using default import
export interface AdvancedDefaultUser extends AliasDefault {
  defaultProp: string;
}

// Type using default import with alias
export interface AdvancedDefaultUser2 extends AliasDefault2 {
  defaultProp2: string;
}

// Re-export with alias
export { AliasUser as RenamedAliasUser } from './utils/alias';