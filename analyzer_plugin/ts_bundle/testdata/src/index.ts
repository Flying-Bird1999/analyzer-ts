// src/index.ts
import { User, UserRole, AdminUser } from './utils/user';
import { Address } from './utils/address';
import * as Common from './utils/common';

// Re-export some types
export { User, UserRole } from './utils/user';
export type { Address } from './utils/address';

// Define a complex type that uses imports
export interface UserProfile extends User {
  address: Address;
  tags: Common.CommonType[];
}

// Define a type that uses namespace import
export type UserId = Common.CommonInterface['id'];

// Define a type that combines multiple imports
export type FullUser = UserProfile & AdminUser;

// Default export
export default UserProfile;