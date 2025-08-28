// src/complex.ts
import { FullUser } from './index';
import { LocalTypeWithExternal } from './external';
import * as UserUtils from './utils/user';

// Type using indexed access
export type UserName = FullUser['name'];

// Type using mapped type
export type UserFields = {
  [K in keyof FullUser]?: FullUser[K];
};

// Type using Omit
export type UserWithoutAddress = Omit<FullUser, 'address'>;

// Type using Pick
export type UserBasicInfo = Pick<FullUser, 'id' | 'name'>;

// Type using namespace import
export interface UserTypeCheck {
  userId: UserUtils.User['id'];
  userRole: UserUtils.UserRole;
}