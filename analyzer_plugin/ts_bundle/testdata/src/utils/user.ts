// src/utils/user.ts
export interface User {
  id: number;
  name: string;
}

export type UserRole = 'admin' | 'user';

export enum UserStatus {
  Active = 'active',
  Inactive = 'inactive'
}

export interface AdminUser extends User {
  role: UserRole;
  status: UserStatus;
}