// 导入基础类型
import { User } from './types';

// 高级类型定义示例
export interface ApiResponse<T> {
  data: T;
  status: number;
  message: string;
  success: boolean;
}

// 泛型工具类型
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

// 条件类型
export type NonNullable<T> = T extends null | undefined ? never : T;

// 映射类型
export type ReadonlyUser = {
  readonly [K in keyof User]: User[K];
};

// 联合类型
export type Theme = 'light' | 'dark' | 'auto';

// 字面量类型
export type UserRole = 'admin' | 'user' | 'moderator';

// 递归类型
export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue };

// 模板字面量类型
export type EventName = `on${Capitalize<string>}`;

// 复杂组合类型
export interface AdvancedUser extends User {
  roles: UserRole[];
  preferences: {
    theme: Theme;
    notifications: boolean;
    language: string;
  };
  metadata: JsonValue;
}

// 装饰器（实验性功能）
function sealed(constructor: Function) {
  Object.seal(constructor);
  Object.seal(constructor.prototype);
}

@sealed
export class UserService {
  private static instance: UserService;

  public static getInstance(): UserService {
    if (!UserService.instance) {
      UserService.instance = new UserService();
    }
    return UserService.instance;
  }

  public async updateUser<T extends Partial<User>>(
    userId: number,
    updates: T
  ): Promise<ApiResponse<User>> {
    // 实现更新逻辑
    return {
      data: {} as User,
      status: 200,
      message: 'Success',
      success: true
    };
  }
}

// 导出基础类型 - 使用路径别名
export type { User } from './types';