import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from 'react-query';
import { useState, useCallback, useMemo } from 'react';
import type { User, UserRole, UserProfile as UserProfileType } from '../types';

// 用户服务接口
interface UserService {
    // 获取用户列表
    getUsers: (params: UserListParams) => Promise<UserListResponse>;
    // 获取单个用户
    getUser: (id: string) => Promise<User>;
    // 创建用户
    createUser: (userData: CreateUserRequest) => Promise<User>;
    // 更新用户
    updateUser: (id: string, userData: UpdateUserRequest) => Promise<User>;
    // 删除用户
    deleteUser: (id: string) => Promise<void>;
    // 批量删除用户
    bulkDeleteUsers: (ids: string[]) => Promise<void>;
    // 搜索用户
    searchUsers: (query: string, limit?: number) => Promise<User[]>;
    // 获取用户统计信息
    getUserStats: () => Promise<UserStats>;
}

// 用户列表参数
interface UserListParams {
    page: number;
    pageSize: number;
    search?: string;
    role?: UserRole;
    status?: 'active' | 'inactive';
    sortBy?: keyof User;
    sortOrder?: 'asc' | 'desc';
}

// 用户列表响应
interface UserListResponse {
    users: User[];
    pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
    };
}

// 创建用户请求
interface CreateUserRequest {
    name: string;
    email: string;
    role: UserRole;
    profile?: Partial<UserProfileType>;
}

// 更新用户请求
interface UpdateUserRequest {
    name?: string;
    email?: string;
    role?: UserRole;
    active?: boolean;
    profile?: Partial<UserProfileType>;
}

// 用户统计信息
interface UserStats {
    totalUsers: number;
    activeUsers: number;
    inactiveUsers: number;
    newUsersThisMonth: number;
    userByRole: Record<UserRole, number>;
    userGrowth: {
        date: string;
        count: number;
    }[];
}

// 查询键生成器
const userKeys = {
    all: ['users'] as const,
    lists: () => [...userKeys.all, 'list'] as const,
    list: (filters: UserListParams) => [...userKeys.lists(), filters] as const,
    details: () => [...userKeys.all, 'detail'] as const,
    detail: (id: string) => [...userKeys.details(), id] as const,
    infinite: (filters: Omit<UserListParams, 'page'>) => [...userKeys.lists(), 'infinite', filters] as const,
    search: (query: string) => [...userKeys.all, 'search', query] as const,
    stats: () => [...userKeys.all, 'stats'] as const
};

// 模拟用户服务实现
class MockUserService implements UserService {
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
        },
        // 添加更多模拟用户...
        {
            id: '3',
            name: 'Bob Johnson',
            email: 'bob.johnson@example.com',
            active: false,
            role: UserRole.MODERATOR,
            createdAt: new Date('2023-03-01'),
            updatedAt: new Date('2023-03-15'),
            profile: {
                bio: 'Inactive moderator',
                social: {
                    twitter: '@bobjohnson'
                }
            }
        },
        {
            id: '4',
            name: 'Alice Brown',
            email: 'alice.brown@example.com',
            active: true,
            role: UserRole.USER,
            createdAt: new Date('2023-04-01'),
            updatedAt: new Date('2023-04-15'),
            profile: {
                avatar: 'https://example.com/avatar4.jpg',
                bio: 'Active community member',
                website: 'https://alicebrown.com',
                social: {
                    twitter: '@alicebrown',
                    github: 'alicebrown',
                    linkedin: 'alicebrown'
                }
            }
        }
    ];

    async getUsers(params: UserListParams): Promise<UserListResponse> {
        await new Promise(resolve => setTimeout(resolve, 500));

        let filteredUsers = [...this.mockUsers];

        // 应用搜索筛选
        if (params.search) {
            filteredUsers = filteredUsers.filter(user =>
                user.name.toLowerCase().includes(params.search!.toLowerCase()) ||
                user.email.toLowerCase().includes(params.search!.toLowerCase())
            );
        }

        // 应用角色筛选
        if (params.role) {
            filteredUsers = filteredUsers.filter(user => user.role === params.role);
        }

        // 应用状态筛选
        if (params.status) {
            filteredUsers = filteredUsers.filter(user =>
                params.status === 'active' ? user.active : !user.active
            );
        }

        // 应用排序
        if (params.sortBy) {
            filteredUsers.sort((a, b) => {
                const aValue = a[params.sortBy!];
                const bValue = b[params.sortBy!];

                if (typeof aValue === 'string' && typeof bValue === 'string') {
                    return params.sortOrder === 'asc'
                        ? aValue.localeCompare(bValue)
                        : bValue.localeCompare(aValue);
                }

                if (aValue instanceof Date && bValue instanceof Date) {
                    return params.sortOrder === 'asc'
                        ? aValue.getTime() - bValue.getTime()
                        : bValue.getTime() - aValue.getTime();
                }

                return 0;
            });
        }

        // 应用分页
        const startIndex = (params.page - 1) * params.pageSize;
        const endIndex = startIndex + params.pageSize;
        const paginatedUsers = filteredUsers.slice(startIndex, endIndex);

        return {
            users: paginatedUsers,
            pagination: {
                page: params.page,
                pageSize: params.pageSize,
                total: filteredUsers.length,
                totalPages: Math.ceil(filteredUsers.length / params.pageSize)
            }
        };
    }

    async getUser(id: string): Promise<User> {
        await new Promise(resolve => setTimeout(resolve, 200));

        const user = this.mockUsers.find(u => u.id === id);
        if (!user) {
            throw new Error('User not found');
        }

        return user;
    }

    async createUser(userData: CreateUserRequest): Promise<User> {
        await new Promise(resolve => setTimeout(resolve, 300));

        const newUser: User = {
            id: Date.now().toString(),
            ...userData,
            active: true,
            createdAt: new Date(),
            updatedAt: new Date(),
            profile: {
                bio: '',
                social: {},
                ...userData.profile
            }
        };

        this.mockUsers.push(newUser);
        return newUser;
    }

    async updateUser(id: string, userData: UpdateUserRequest): Promise<User> {
        await new Promise(resolve => setTimeout(resolve, 300));

        const userIndex = this.mockUsers.findIndex(u => u.id === id);
        if (userIndex === -1) {
            throw new Error('User not found');
        }

        this.mockUsers[userIndex] = {
            ...this.mockUsers[userIndex],
            ...userData,
            updatedAt: new Date(),
            profile: {
                ...this.mockUsers[userIndex].profile,
                ...userData.profile
            }
        };

        return this.mockUsers[userIndex];
    }

    async deleteUser(id: string): Promise<void> {
        await new Promise(resolve => setTimeout(resolve, 300));

        const userIndex = this.mockUsers.findIndex(u => u.id === id);
        if (userIndex === -1) {
            throw new Error('User not found');
        }

        this.mockUsers.splice(userIndex, 1);
    }

    async bulkDeleteUsers(ids: string[]): Promise<void> {
        await new Promise(resolve => setTimeout(resolve, 500));

        ids.forEach(id => {
            const userIndex = this.mockUsers.findIndex(u => u.id === id);
            if (userIndex !== -1) {
                this.mockUsers.splice(userIndex, 1);
            }
        });
    }

    async searchUsers(query: string, limit = 10): Promise<User[]> {
        await new Promise(resolve => setTimeout(resolve, 300));

        const results = this.mockUsers.filter(user =>
            user.name.toLowerCase().includes(query.toLowerCase()) ||
            user.email.toLowerCase().includes(query.toLowerCase()) ||
            user.profile?.bio?.toLowerCase().includes(query.toLowerCase())
        );

        return results.slice(0, limit);
    }

    async getUserStats(): Promise<UserStats> {
        await new Promise(resolve => setTimeout(resolve, 400));

        const now = new Date();
        const thisMonth = now.getMonth();
        const thisYear = now.getFullYear();

        const userByRole: Record<UserRole, number> = {
            [UserRole.ADMIN]: 0,
            [UserRole.USER]: 0,
            [UserRole.MODERATOR]: 0
        };

        let newUsersThisMonth = 0;

        this.mockUsers.forEach(user => {
            userByRole[user.role]++;

            if (user.createdAt.getMonth() === thisMonth &&
                user.createdAt.getFullYear() === thisYear) {
                newUsersThisMonth++;
            }
        });

        // 生成用户增长数据（模拟）
        const userGrowth = Array.from({ length: 30 }, (_, i) => {
            const date = new Date(now);
            date.setDate(date.getDate() - (29 - i));

            return {
                date: date.toISOString().split('T')[0],
                count: Math.floor(Math.random() * 5) + (i < 10 ? 0 : i < 20 ? 2 : 3)
            };
        });

        return {
            totalUsers: this.mockUsers.length,
            activeUsers: this.mockUsers.filter(u => u.active).length,
            inactiveUsers: this.mockUsers.filter(u => !u.active).length,
            newUsersThisMonth,
            userByRole,
            userGrowth
        };
    }
}

// 创建用户服务实例
const userService = new MockUserService();

// 用户数据获取Hook
export const useUsers = (params: UserListParams) => {
    return useQuery(
        userKeys.list(params),
        () => userService.getUsers(params),
        {
            // 查询配置
            keepPreviousData: true,
            staleTime: 5 * 60 * 1000, // 5分钟
            cacheTime: 10 * 60 * 1000, // 10分钟

            // 错误处理
            retry: (failureCount, error) => {
                // 最多重试3次
                if (failureCount >= 3) return false;
                // 404错误不重试
                if (error instanceof Error && error.message.includes('not found')) return false;
                return true;
            },

            // 查询成功时的处理
            onSuccess: (data) => {
                console.log('Users loaded successfully:', data.users.length);
            },

            // 查询失败时的处理
            onError: (error) => {
                console.error('Failed to load users:', error);
            }
        }
    );
};

// 单个用户数据获取Hook
export const useUser = (id: string, enabled = true) => {
    return useQuery(
        userKeys.detail(id),
        () => userService.getUser(id),
        {
            enabled: id !== '' && enabled,
            staleTime: 10 * 60 * 1000, // 10分钟
            cacheTime: 30 * 60 * 1000, // 30分钟

            // 当用户不存在时返回null而不是抛出错误
            useErrorBoundary: false,

            // 重试配置
            retry: (failureCount, error) => {
                if (failureCount >= 2) return false;
                if (error instanceof Error && error.message.includes('not found')) return false;
                return true;
            }
        }
    );
};

// 无限滚动用户列表Hook
export const useInfiniteUsers = (filters: Omit<UserListParams, 'page'>) => {
    return useInfiniteQuery(
        userKeys.infinite(filters),
        ({ pageParam = 1 }) =>
            userService.getUsers({ ...filters, page: pageParam }),
        {
            getNextPageParam: (lastPage) => {
                // 如果还有更多页，返回下一页码，否则返回undefined
                return lastPage.pagination.page < lastPage.pagination.totalPages
                    ? lastPage.pagination.page + 1
                    : undefined;
            },

            // 每页10条数据，预加载2页
            getPreviousPageParam: (firstPage) => {
                return firstPage.pagination.page > 1
                    ? firstPage.pagination.page - 1
                    : undefined;
            }
        }
    );
};

// 用户搜索Hook（带防抖）
export const useUserSearch = (query: string, enabled = true) => {
    return useQuery(
        userKeys.search(query),
        () => userService.searchUsers(query),
        {
            enabled: query.trim().length >= 2 && enabled,
            staleTime: 30 * 1000, // 30秒
            cacheTime: 5 * 60 * 1000, // 5分钟

            // 空查询时重置
            placeholderData: (previousData) =>
                query.trim().length >= 2 ? undefined : previousData
        }
    );
};

// 用户统计信息Hook
export const useUserStats = (enabled = true) => {
    return useQuery(
        userKeys.stats(),
        () => userService.getUserStats(),
        {
            enabled,
            staleTime: 2 * 60 * 1000, // 2分钟
            cacheTime: 10 * 60 * 1000, // 10分钟

            // 轮询更新（每30秒）
            refetchInterval: 30 * 1000,

            // 窗口获得焦点时刷新
            refetchOnWindowFocus: true,

            // 网络重连时刷新
            refetchOnReconnect: true
        }
    );
};

// 用户创建Mutation Hook
export const useCreateUser = () => {
    const queryClient = useQueryClient();

    return useMutation(
        (userData: CreateUserRequest) => userService.createUser(userData),
        {
            // 创建成功后的处理
            onSuccess: (newUser) => {
                // 更新用户列表缓存
                queryClient.invalidateQueries(userKeys.lists());

                // 更新统计信息缓存
                queryClient.invalidateQueries(userKeys.stats());

                // 显示成功消息
                console.log('User created successfully:', newUser);
            },

            // 创建失败时的处理
            onError: (error) => {
                console.error('Failed to create user:', error);
            },

            // 乐观更新
            onMutate: async (newUserData) => {
                // 取消所有进行中的用户列表查询
                await queryClient.cancelQueries(userKeys.lists());

                // 获取当前的用户列表快照
                const previousUsersList = queryClient.getQueryData(
                    userKeys.lists()
                );

                return { previousUsersList };
            },

            // 回滚操作
            onError: (error, newUserData, context) => {
                if (context?.previousUsersList) {
                    queryClient.setQueryData(
                        userKeys.lists(),
                        context.previousUsersList
                    );
                }
            }
        }
    );
};

// 用户更新Mutation Hook
export const useUpdateUser = () => {
    const queryClient = useQueryClient();

    return useMutation(
        ({ id, userData }: { id: string; userData: UpdateUserRequest }) =>
            userService.updateUser(id, userData),
        {
            onSuccess: (updatedUser) => {
                // 更新单个用户缓存
                queryClient.setQueryData(
                    userKeys.detail(updatedUser.id),
                    updatedUser
                );

                // 更新用户列表缓存
                queryClient.invalidateQueries(userKeys.lists());

                // 更新统计信息缓存
                queryClient.invalidateQueries(userKeys.stats());

                console.log('User updated successfully:', updatedUser);
            },

            onError: (error) => {
                console.error('Failed to update user:', error);
            }
        }
    );
};

// 用户删除Mutation Hook
export const useDeleteUser = () => {
    const queryClient = useQueryClient();

    return useMutation(
        (id: string) => userService.deleteUser(id),
        {
            onSuccess: (deletedUserId) => {
                // 从缓存中移除已删除的用户
                queryClient.removeQueries(userKeys.detail(deletedUserId));

                // 更新用户列表缓存
                queryClient.invalidateQueries(userKeys.lists());

                // 更新统计信息缓存
                queryClient.invalidateQueries(userKeys.stats());

                console.log('User deleted successfully:', deletedUserId);
            },

            onError: (error) => {
                console.error('Failed to delete user:', error);
            }
        }
    );
};

// 批量删除用户Mutation Hook
export const useBulkDeleteUsers = () => {
    const queryClient = useQueryClient();

    return useMutation(
        (ids: string[]) => userService.bulkDeleteUsers(ids),
        {
            onSuccess: (deletedIds) => {
                // 从缓存中移除已删除的用户
                deletedIds.forEach(id => {
                    queryClient.removeQueries(userKeys.detail(id));
                });

                // 更新用户列表缓存
                queryClient.invalidateQueries(userKeys.lists());

                // 更新统计信息缓存
                queryClient.invalidateQueries(userKeys.stats());

                console.log('Bulk delete completed:', deletedIds);
            },

            onError: (error) => {
                console.error('Bulk delete failed:', error);
            }
        }
    );
};

// 复合Hook：用户管理（结合查询和操作）
export const useUserManagement = () => {
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize] = useState(10);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedRole, setSelectedRole] = useState<UserRole | 'all'>('all');
    const [selectedStatus, setSelectedStatus] = useState<'all' | 'active' | 'inactive'>('all');

    // 用户列表查询
    const usersQuery = useUsers({
        page: currentPage,
        pageSize,
        search: searchQuery || undefined,
        role: selectedRole === 'all' ? undefined : selectedRole,
        status: selectedStatus === 'all' ? undefined : selectedStatus,
        sortBy: 'createdAt',
        sortOrder: 'desc'
    });

    // 用户统计查询
    const userStatsQuery = useUserStats();

    // 操作mutations
    const createUserMutation = useCreateUser();
    const updateUserMutation = useUpdateUser();
    const deleteUserMutation = useDeleteUser();
    const bulkDeleteUsersMutation = useBulkDeleteUsers();

    // 计算分页信息
    const paginationInfo = useMemo(() => {
        if (!usersQuery.data) return null;

        return {
            currentPage: usersQuery.data.pagination.page,
            totalPages: usersQuery.data.pagination.totalPages,
            totalItems: usersQuery.data.pagination.total,
            hasNextPage: usersQuery.data.pagination.page < usersQuery.data.pagination.totalPages,
            hasPreviousPage: usersQuery.data.pagination.page > 1
        };
    }, [usersQuery.data]);

    // 事件处理函数
    const handlePageChange = useCallback((newPage: number) => {
        setCurrentPage(newPage);
    }, []);

    const handleSearch = useCallback((query: string) => {
        setSearchQuery(query);
        setCurrentPage(1); // 重置到第一页
    }, []);

    const handleRoleFilter = useCallback((role: UserRole | 'all') => {
        setSelectedRole(role);
        setCurrentPage(1); // 重置到第一页
    }, []);

    const handleStatusFilter = useCallback((status: 'all' | 'active' | 'inactive') => {
        setSelectedStatus(status);
        setCurrentPage(1); // 重置到第一页
    }, []);

    // 导出所有数据和方法
    return {
        // 状态
        currentPage,
        pageSize,
        searchQuery,
        selectedRole,
        selectedStatus,

        // 查询数据
        users: usersQuery.data?.users || [],
        isLoading: usersQuery.isLoading,
        isError: usersQuery.isError,
        error: usersQuery.error,

        // 统计数据
        userStats: userStatsQuery.data,
        statsLoading: userStatsQuery.isLoading,
        statsError: userStatsQuery.error,

        // 分页信息
        pagination: paginationInfo,

        // Mutations
        createUser: createUserMutation.mutate,
        updateUser: updateUserMutation.mutate,
        deleteUser: deleteUserMutation.mutate,
        bulkDeleteUsers: bulkDeleteUsersMutation.mutate,

        // Mutation状态
        isCreating: createUserMutation.isLoading,
        isUpdating: updateUserMutation.isLoading,
        isDeleting: deleteUserMutation.isLoading,
        isBulkDeleting: bulkDeleteUsersMutation.isLoading,

        // 事件处理
        handlePageChange,
        handleSearch,
        handleRoleFilter,
        handleStatusFilter
    };
};

// 导出服务和类型
export { userService, userKeys };
export type {
    UserService,
    UserListParams,
    UserListResponse,
    CreateUserRequest,
    UpdateUserRequest,
    UserStats
};