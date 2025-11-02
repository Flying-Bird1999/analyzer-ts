import { useContext, useEffect, useCallback, createContext } from 'react';
import type { User, UserRole } from '../types';
import type { UserProfile as UserProfileType } from '../types';

// 认证状态类型定义
export interface AuthState {
    user: User | null;
    loading: boolean;
    error: string | null;
    isAuthenticated: boolean;
    permissions: string[];
    lastAuthenticated: Date | null;
}

// 认证动作类型定义
export interface AuthActions {
    login: (credentials: Credentials) => Promise<void>;
    logout: () => Promise<void>;
    updateUserProfile: (profile: Partial<UserProfileType>) => Promise<void>;
    refreshUser: () => Promise<void>;
    clearError: () => void;
}

// 合并认证状态和动作
export type AuthContextType = AuthState & AuthActions;

// 默认认证状态
const defaultAuthState: AuthState = {
    user: null,
    loading: false,
    error: null,
    isAuthenticated: false,
    permissions: [],
    lastAuthenticated: null
};

// 认证上下文
const AuthContext = createContext<AuthContextType>({
    ...defaultAuthState,
    login: async () => {},
    logout: async () => {},
    updateUserProfile: async () => {},
    refreshUser: async () => {},
    clearError: () => {}
});

// 凭证类型定义
export interface Credentials {
    email: string;
    password: string;
    rememberMe?: boolean;
}

// 登录响应类型定义
export interface LoginResponse {
    user: User;
    token: string;
    refreshToken: string;
    expiresIn: number;
    permissions: string[];
}

// 认证服务接口
interface AuthService {
    login: (credentials: Credentials) => Promise<LoginResponse>;
    logout: () => Promise<void>;
    refreshToken: (refreshToken: string) => Promise<LoginResponse>;
    getCurrentUser: () => Promise<User>;
    updateUserProfile: (userId: string, profile: Partial<UserProfileType>) => Promise<User>;
}

// 本地存储管理器
interface StorageManager {
    getItem: (key: string) => string | null;
    setItem: (key: string, value: string) => void;
    removeItem: (key: string) => void;
    clear: () => void;
}

// 认证提供器Props
interface AuthProviderProps {
    children: React.ReactNode;
    authService: AuthService;
    storageManager: StorageManager;
    onAuthChange?: (user: User | null) => void;
}

// 默认认证服务实现
class DefaultAuthService implements AuthService {
    private mockUsers: User[] = [
        {
            id: '1',
            name: 'John Doe',
            email: 'john.doe@example.com',
            active: true,
            role: UserRole.ADMIN,
            createdAt: new Date('2023-01-01'),
            updatedAt: new Date('2023-01-15'),
            profile: {
                avatar: 'https://example.com/avatar1.jpg',
                bio: 'System administrator with full access',
                website: 'https://johndoe.com',
                social: {
                    twitter: '@johndoe',
                    github: 'johndoe'
                }
            }
        },
        {
            id: '2',
            name: 'Jane Smith',
            email: 'jane.smith@example.com',
            active: true,
            role: UserRole.USER,
            createdAt: new Date('2023-02-01'),
            updatedAt: new Date('2023-02-15'),
            profile: {
                avatar: 'https://example.com/avatar2.jpg',
                bio: 'Regular user with standard permissions',
                website: 'https://janesmith.com',
                social: {
                    github: 'janesmith'
                }
            }
        }
    ];

    async login(credentials: Credentials): Promise<LoginResponse> {
        // 模拟网络延迟
        await new Promise(resolve => setTimeout(resolve, 500));

        const user = this.mockUsers.find(u => u.email === credentials.email);

        if (!user || credentials.password !== 'password') {
            throw new Error('Invalid email or password');
        }

        if (!user.active) {
            throw new Error('Account is disabled');
        }

        // 模拟权限计算
        const permissions = this.calculatePermissions(user.role);

        return {
            user,
            token: this.generateToken(user),
            refreshToken: this.generateRefreshToken(),
            expiresIn: 3600,
            permissions
        };
    }

    async logout(): Promise<void> {
        await new Promise(resolve => setTimeout(resolve, 100));
        // 清理服务器端的会话（模拟）
    }

    async refreshToken(refreshToken: string): Promise<LoginResponse> {
        await new Promise(resolve => setTimeout(resolve, 300));

        if (!this.isValidToken(refreshToken)) {
            throw new Error('Invalid refresh token');
        }

        // 模拟刷新成功
        const user = this.mockUsers[0]; // 简化实现
        const permissions = this.calculatePermissions(user.role);

        return {
            user,
            token: this.generateToken(user),
            refreshToken: this.generateRefreshToken(),
            expiresIn: 3600,
            permissions
        };
    }

    async getCurrentUser(): Promise<User> {
        await new Promise(resolve => setTimeout(resolve, 200));

        // 模拟从服务器获取当前用户
        return this.mockUsers[0];
    }

    async updateUserProfile(userId: string, profile: Partial<UserProfileType>): Promise<User> {
        await new Promise(resolve => setTimeout(resolve, 300));

        const userIndex = this.mockUsers.findIndex(u => u.id === userId);
        if (userIndex === -1) {
            throw new Error('User not found');
        }

        this.mockUsers[userIndex] = {
            ...this.mockUsers[userIndex],
            profile: { ...this.mockUsers[userIndex].profile, ...profile },
            updatedAt: new Date()
        };

        return this.mockUsers[userIndex];
    }

    private calculatePermissions(role: UserRole): string[] {
        const rolePermissions: Record<UserRole, string[]> = {
            [UserRole.ADMIN]: [
                'users.read',
                'users.write',
                'users.delete',
                'profile.read',
                'profile.write',
                'admin.dashboard',
                'system.settings'
            ],
            [UserRole.USER]: [
                'users.read',
                'profile.read',
                'profile.write'
            ],
            [UserRole.MODERATOR]: [
                'users.read',
                'users.write',
                'profile.read',
                'profile.write',
                'moderate.content'
            ]
        };

        return rolePermissions[role] || [];
    }

    private generateToken(user: User): string {
        return `mock-jwt-token-${user.id}-${Date.now()}`;
    }

    private generateRefreshToken(): string {
        return `mock-refresh-token-${Date.now()}`;
    }

    private isValidToken(token: string): boolean {
        // 简化实现，实际应该验证token的有效性和过期时间
        return token.startsWith('mock-');
    }
}

// 默认存储管理器实现
class DefaultStorageManager implements StorageManager {
    getItem(key: string): string | null {
        return localStorage.getItem(key);
    }

    setItem(key: string, value: string): void {
        localStorage.setItem(key, value);
    }

    removeItem(key: string): void {
        localStorage.removeItem(key);
    }

    clear(): void {
        localStorage.clear();
    }
}

// 认证提供器组件
export const AuthProvider: React.FC<AuthProviderProps> = ({
    children,
    authService = new DefaultAuthService(),
    storageManager = new DefaultStorageManager(),
    onAuthChange
}) => {
    const [state, setState] = useState<AuthState>(() => {
        const savedUser = storageManager.getItem('auth_user');
        const savedToken = storageManager.getItem('auth_token');
        const savedPermissions = storageManager.getItem('auth_permissions');
        const lastAuthenticated = storageManager.getItem('auth_last_authenticated');

        return {
            user: savedUser ? JSON.parse(savedUser) : null,
            loading: false,
            error: null,
            isAuthenticated: !!(savedToken && savedUser),
            permissions: savedPermissions ? JSON.parse(savedPermissions) : [],
            lastAuthenticated: lastAuthenticated ? new Date(lastAuthenticated) : null
        };
    });

    // 保存认证信息到本地存储
    const saveAuthInfo = useCallback((user: User, token: string, permissions: string[]) => {
        storageManager.setItem('auth_user', JSON.stringify(user));
        storageManager.setItem('auth_token', token);
        storageManager.setItem('auth_permissions', JSON.stringify(permissions));
        storageManager.setItem('auth_last_authenticated', new Date().toISOString());
    }, [storageManager]);

    // 清除认证信息
    const clearAuthInfo = useCallback(() => {
        storageManager.removeItem('auth_user');
        storageManager.removeItem('auth_token');
        storageManager.removeItem('auth_permissions');
        storageManager.removeItem('auth_last_authenticated');
        storageManager.removeItem('auth_refresh_token');
    }, [storageManager]);

    // 登录函数
    const login = useCallback(async (credentials: Credentials): Promise<void> => {
        setState(prev => ({ ...prev, loading: true, error: null }));

        try {
            const response = await authService.login(credentials);

            saveAuthInfo(response.user, response.token, response.permissions);

            setState(prev => ({
                ...prev,
                user: response.user,
                loading: false,
                isAuthenticated: true,
                permissions: response.permissions,
                lastAuthenticated: new Date(),
                error: null
            }));

            onAuthChange?.(response.user);
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Login failed',
                isAuthenticated: false
            }));
            throw error;
        }
    }, [authService, saveAuthInfo, onAuthChange]);

    // 登出函数
    const logout = useCallback(async (): Promise<void> => {
        setState(prev => ({ ...prev, loading: true }));

        try {
            await authService.logout();
            clearAuthInfo();

            setState(prev => ({
                ...prev,
                user: null,
                loading: false,
                isAuthenticated: false,
                permissions: [],
                lastAuthenticated: null,
                error: null
            }));

            onAuthChange?.(null);
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Logout failed'
            }));
            throw error;
        }
    }, [authService, clearAuthInfo, onAuthChange]);

    // 更新用户档案
    const updateUserProfile = useCallback(async (profile: Partial<UserProfileType>): Promise<void> => {
        if (!state.user) {
            throw new Error('No authenticated user');
        }

        setState(prev => ({ ...prev, loading: true }));

        try {
            const updatedUser = await authService.updateUserProfile(state.user.id, profile);

            setState(prev => ({
                ...prev,
                user: updatedUser,
                loading: false,
                error: null
            }));

            onAuthChange?.(updatedUser);
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to update profile'
            }));
            throw error;
        }
    }, [authService, state.user, onAuthChange]);

    // 刷新用户信息
    const refreshUser = useCallback(async (): Promise<void> => {
        if (!state.isAuthenticated) {
            return;
        }

        setState(prev => ({ ...prev, loading: true }));

        try {
            const currentUser = await authService.getCurrentUser();

            setState(prev => ({
                ...prev,
                user: currentUser,
                loading: false,
                error: null
            }));

            onAuthChange?.(currentUser);
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to refresh user'
            }));
            throw error;
        }
    }, [authService, state.isAuthenticated, onAuthChange]);

    // 清除错误
    const clearError = useCallback((): void => {
        setState(prev => ({ ...prev, error: null }));
    }, []);

    // 自动令牌刷新
    useEffect(() => {
        if (!state.isAuthenticated) {
            return;
        }

        const refreshTokenInterval = setInterval(async () => {
            const refreshToken = storageManager.getItem('auth_refresh_token');
            if (refreshToken) {
                try {
                    await authService.refreshToken(refreshToken);
                } catch (error) {
                    console.warn('Failed to refresh token:', error);
                    // 刷新失败，可能是token过期，触发登出
                    logout().catch(console.error);
                }
            }
        }, 30 * 60 * 1000); // 每30分钟刷新一次

        return () => clearInterval(refreshTokenInterval);
    }, [state.isAuthenticated, authService, logout, storageManager]);

    // 页面加载时验证认证状态
    useEffect(() => {
        if (state.isAuthenticated && !state.user) {
            refreshUser().catch(console.error);
        }
    }, [state.isAuthenticated, state.user, refreshUser]);

    // 权限检查函数
    const hasPermission = useCallback((permission: string): boolean => {
        return state.permissions.includes(permission);
    }, [state.permissions]);

    const hasAnyPermission = useCallback((permissions: string[]): boolean => {
        return permissions.some(perm => state.permissions.includes(perm));
    }, [state.permissions]);

    const hasAllPermissions = useCallback((permissions: string[]): boolean => {
        return permissions.every(perm => state.permissions.includes(perm));
    }, [state.permissions]);

    // 角色检查函数
    const hasRole = useCallback((role: UserRole): boolean => {
        return state.user?.role === role;
    }, [state.user]);

    const hasAnyRole = useCallback((roles: UserRole[]): boolean => {
        return state.user ? roles.includes(state.user.role) : false;
    }, [state.user]);

    const contextValue: AuthContextType = {
        ...state,
        login,
        logout,
        updateUserProfile,
        refreshUser,
        clearError,
        hasPermission,
        hasAnyPermission,
        hasAllPermissions,
        hasRole,
        hasAnyRole
    };

    return (
        <AuthContext.Provider value={contextValue}>
            {children}
        </AuthContext.Provider>
    );
};

// 使用认证上下文的Hook
export const useAuth = (): AuthContextType => {
    const context = useContext(AuthContext);

    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }

    return context;
};

// 认证守卫HOC
export interface AuthGuardProps {
    children: React.ReactNode;
    requiredPermissions?: string[];
    requiredRoles?: UserRole[];
    redirectTo?: string;
    fallback?: React.ReactNode;
}

export const AuthGuard: React.FC<AuthGuardProps> = ({
    children,
    requiredPermissions = [],
    requiredRoles = [],
    redirectTo = '/login',
    fallback = null
}) => {
    const { isAuthenticated, hasPermission, hasAnyPermission, hasRole, hasAnyRole } = useAuth();

    // 检查用户是否已认证
    if (!isAuthenticated) {
        if (fallback) {
            return <>{fallback}</>;
        }
        // 这里应该使用 React Router 进行重定向
        // 简化实现，直接返回fallback
        return <div>Please log in to access this page.</div>;
    }

    // 检查权限
    if (requiredPermissions.length > 0) {
        if (requiredPermissions.length === 1) {
            if (!hasPermission(requiredPermissions[0])) {
                if (fallback) {
                    return <>{fallback}</>;
                }
                return <div>You don't have permission to access this page.</div>;
            }
        } else {
            if (!hasAnyPermission(requiredPermissions)) {
                if (fallback) {
                    return <>{fallback}</>;
                }
                return <div>You don't have the required permissions to access this page.</div>;
            }
        }
    }

    // 检查角色
    if (requiredRoles.length > 0) {
        if (requiredRoles.length === 1) {
            if (!hasRole(requiredRoles[0])) {
                if (fallback) {
                    return <>{fallback}</>;
                }
                return <div>You don't have the required role to access this page.</div>;
            }
        } else {
            if (!hasAnyRole(requiredRoles)) {
                if (fallback) {
                    return <>{fallback}</>;
                }
                return <div>You don't have any of the required roles to access this page.</div>;
            }
        }
    }

    // 所有检查通过，渲染子组件
    return <>{children}</>;
};

// 权限Hook
export const usePermissions = (): {
    hasPermission: (permission: string) => boolean;
    hasAnyPermission: (permissions: string[]) => boolean;
    hasAllPermissions: (permissions: string[]) => boolean;
    permissions: string[];
} => {
    const { permissions: userPermissions, hasPermission, hasAnyPermission, hasAllPermissions } = useAuth();

    return {
        hasPermission,
        hasAnyPermission,
        hasAllPermissions,
        permissions: userPermissions
    };
};

// 角色Hook
export const useRoles = (): {
    hasRole: (role: UserRole) => boolean;
    hasAnyRole: (roles: UserRole[]) => boolean;
    currentRole: UserRole | null;
    roles: UserRole[];
} => {
    const { user, hasRole, hasAnyRole } = useAuth();
    const availableRoles: UserRole[] = [UserRole.ADMIN, UserRole.USER, UserRole.MODERATOR];

    return {
        hasRole,
        hasAnyRole,
        currentRole: user?.role || null,
        roles: availableRoles
    };
};

export default AuthProvider;