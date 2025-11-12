import { useState, useCallback } from 'react';
import { User } from '@/types/types';
import { api } from '@/services/api';

// 用户状态管理
interface UserState {
  user: User | null;
  loading: boolean;
  error: string | null;
}

// 使用简单的状态管理模式
let globalUserState: UserState = {
  user: null,
  loading: false,
  error: null,
};

let listeners: ((state: UserState) => void)[] = [];

// 状态更新函数
const updateUserState = (updater: (prev: UserState) => UserState) => {
  globalUserState = updater(globalUserState);
  listeners.forEach(listener => listener(globalUserState));
};

// 状态订阅函数
export const useUserStore = () => {
  const [state, setState] = useState(globalUserState);

  // 订阅状态变化
  const subscribe = useCallback(() => {
    const listener = (newState: UserState) => {
      setState(newState);
    };

    listeners.push(listener);

    // 返回取消订阅函数
    return () => {
      listeners = listeners.filter(l => l !== listener);
    };
  }, []);

  // 订阅状态
  React.useEffect(() => {
    return subscribe();
  }, [subscribe]);

  // Actions
  const actions = {
    // 获取用户
    fetchUser: useCallback(async (userId: number) => {
      updateUserState(prev => ({ ...prev, loading: true, error: null }));

      try {
        const response = await api.getUser(userId);
        updateUserState(prev => ({
          ...prev,
          user: response.data,
          loading: false,
          error: null,
        }));
      } catch (error) {
        updateUserState(prev => ({
          ...prev,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to fetch user',
        }));
      }
    }, []),

    // 更新用户
    updateUser: useCallback(async (userId: number, updates: Partial<User>) => {
      if (!state.user) return;

      updateUserState(prev => ({ ...prev, loading: true, error: null }));

      try {
        const response = await api.updateUser(userId, updates);
        updateUserState(prev => ({
          ...prev,
          user: response.data,
          loading: false,
          error: null,
        }));
      } catch (error) {
        updateUserState(prev => ({
          ...prev,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to update user',
        }));
      }
    }, [state.user]),

    // 清除错误
    clearError: useCallback(() => {
      updateUserState(prev => ({ ...prev, error: null }));
    }, []),

    // 登出
    logout: useCallback(() => {
      updateUserState(() => ({
        user: null,
        loading: false,
        error: null,
      }));
    }, []),
  };

  return {
    ...state,
    ...actions,
  };
};