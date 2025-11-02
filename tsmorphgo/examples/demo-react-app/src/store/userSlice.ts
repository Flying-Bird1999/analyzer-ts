import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import type { User, UserRole } from '../types';

// 用户状态接口
interface UserState {
    // 当前用户信息
    currentUser: User | null;
    // 用户列表
    users: User[];
    // 加载状态
    loading: boolean;
    // 错误信息
    error: string | null;
    // 选中用户ID
    selectedUserId: string | null;
    // 分页信息
    pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
    };
    // 筛选条件
    filters: {
        search: string;
        role: UserRole | 'all';
        status: 'all' | 'active' | 'inactive';
    };
    // 排序配置
    sorting: {
        field: keyof User;
        order: 'asc' | 'desc';
    };
}

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
    deleteUser: (id: string) Promise<void>;
    // 批量删除用户
    bulkDeleteUsers: (ids: string[]) Promise<void>;
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
    profile?: Partial<User['profile']>;
}

// 更新用户请求
interface UpdateUserRequest {
    name?: string;
    email?: string;
    role?: UserRole;
    active?: boolean;
    profile?: Partial<User['profile']>;
}

// 初始状态
const initialState: UserState = {
    currentUser: null,
    users: [],
    loading: false,
    error: null,
    selectedUserId: null,
    pagination: {
        page: 1,
        pageSize: 10,
        total: 0,
        totalPages: 0
    },
    filters: {
        search: '',
        role: 'all',
        status: 'all'
    },
    sorting: {
        field: 'createdAt',
        order: 'desc'
    }
};

// 模拟用户服务
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
        }
    ];

    async getUsers(params: UserListParams): Promise<UserListResponse> {
        // 模拟网络延迟
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
}

// 创建用户服务实例
const userService = new MockUserService();

// 异步Thunk actions
export const fetchUsers = createAsyncThunk(
    'users/fetchUsers',
    async (params: UserListParams, { rejectWithValue }) => {
        try {
            const response = await userService.getUsers(params);
            return response;
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch users');
        }
    }
);

export const fetchUser = createAsyncThunk(
    'users/fetchUser',
    async (id: string, { rejectWithValue }) => {
        try {
            const user = await userService.getUser(id);
            return user;
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch user');
        }
    }
);

export const createUser = createAsyncThunk(
    'users/createUser',
    async (userData: CreateUserRequest, { rejectWithValue }) => {
        try {
            const user = await userService.createUser(userData);
            return user;
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to create user');
        }
    }
);

export const updateUser = createAsyncThunk(
    'users/updateUser',
    async ({ id, userData }: { id: string; userData: UpdateUserRequest }, { rejectWithValue }) => {
        try {
            const user = await userService.updateUser(id, userData);
            return user;
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to update user');
        }
    }
);

export const deleteUser = createAsyncThunk(
    'users/deleteUser',
    async (id: string, { rejectWithValue }) => {
        try {
            await userService.deleteUser(id);
            return id;
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to delete user');
        }
    }
);

export const bulkDeleteUsers = createAsyncThunk(
    'users/bulkDeleteUsers',
    async (ids: string[], { rejectWithValue }) => {
        try {
            await userService.bulkDeleteUsers(ids);
            return ids;
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to delete users');
        }
    }
);

// 用户Slice
const userSlice = createSlice({
    name: 'users',
    initialState,
    reducers: {
        // 设置当前用户
        setCurrentUser: (state, action: PayloadAction<User | null>) => {
            state.currentUser = action.payload;
        },
        // 设置选中用户
        setSelectedUser: (state, action: PayloadAction<string | null>) => {
            state.selectedUserId = action.payload;
        },
        // 设置筛选条件
        setFilters: (state, action: PayloadAction<Partial<UserState['filters']>>) => {
            state.filters = { ...state.filters, ...action.payload };
        },
        // 设置排序条件
        setSorting: (state, action: PayloadAction<UserState['sorting']>) => {
            state.sorting = action.payload;
        },
        // 清除错误
        clearError: (state) => {
            state.error = null;
        },
        // 重置状态
        resetState: (state) => {
            Object.assign(state, initialState);
        },
        // 从列表中添加用户（用于创建后的实时更新）
        addUserToList: (state, action: PayloadAction<User>) => {
            state.users.unshift(action.payload);
        },
        // 更新列表中的用户
        updateUserInList: (state, action: PayloadAction<User>) => {
            const index = state.users.findIndex(u => u.id === action.payload.id);
            if (index !== -1) {
                state.users[index] = action.payload;
            }
        },
        // 从列表中删除用户
        removeUserFromList: (state, action: PayloadAction<string>) => {
            state.users = state.users.filter(u => u.id !== action.payload);
        }
    },
    extraReducers: (builder) => {
        // fetchUsers
        builder.addCase(fetchUsers.pending, (state) => {
            state.loading = true;
            state.error = null;
        });
        builder.addCase(fetchUsers.fulfilled, (state, action) => {
            state.loading = false;
            state.users = action.payload.users;
            state.pagination = action.payload.pagination;
        });
        builder.addCase(fetchUsers.rejected, (state, action) => {
            state.loading = false;
            state.error = action.payload as string;
        });

        // fetchUser
        builder.addCase(fetchUser.pending, (state) => {
            state.loading = true;
            state.error = null;
        });
        builder.addCase(fetchUser.fulfilled, (state, action) => {
            state.loading = false;
            state.currentUser = action.payload;
        });
        builder.addCase(fetchUser.rejected, (state, action) => {
            state.loading = false;
            state.error = action.payload as string;
        });

        // createUser
        builder.addCase(createUser.pending, (state) => {
            state.loading = true;
            state.error = null;
        });
        builder.addCase(createUser.fulfilled, (state, action) => {
            state.loading = false;
            state.users.unshift(action.payload);
        });
        builder.addCase(createUser.rejected, (state, action) => {
            state.loading = false;
            state.error = action.payload as string;
        });

        // updateUser
        builder.addCase(updateUser.pending, (state) => {
            state.loading = true;
            state.error = null;
        });
        builder.addCase(updateUser.fulfilled, (state, action) => {
            state.loading = false;
            const index = state.users.findIndex(u => u.id === action.payload.id);
            if (index !== -1) {
                state.users[index] = action.payload;
            }
            if (state.currentUser?.id === action.payload.id) {
                state.currentUser = action.payload;
            }
        });
        builder.addCase(updateUser.rejected, (state, action) => {
            state.loading = false;
            state.error = action.payload as string;
        });

        // deleteUser
        builder.addCase(deleteUser.pending, (state) => {
            state.loading = true;
            state.error = null;
        });
        builder.addCase(deleteUser.fulfilled, (state, action) => {
            state.loading = false;
            state.users = state.users.filter(u => u.id !== action.payload);
            if (state.currentUser?.id === action.payload) {
                state.currentUser = null;
            }
        });
        builder.addCase(deleteUser.rejected, (state, action) => {
            state.loading = false;
            state.error = action.payload as string;
        });

        // bulkDeleteUsers
        builder.addCase(bulkDeleteUsers.pending, (state) => {
            state.loading = true;
            state.error = null;
        });
        builder.addCase(bulkDeleteUsers.fulfilled, (state, action) => {
            state.loading = false;
            state.users = state.users.filter(u => !action.payload.includes(u.id));
            if (state.currentUser && action.payload.includes(state.currentUser.id)) {
                state.currentUser = null;
            }
        });
        builder.addCase(bulkDeleteUsers.rejected, (state, action) => {
            state.loading = false;
            state.error = action.payload as string;
        });
    }
});

// 导出actions
export const {
    setCurrentUser,
    setSelectedUser,
    setFilters,
    setSorting,
    clearError,
    resetState,
    addUserToList,
    updateUserInList,
    removeUserFromList
} = userSlice.actions;

// 导出reducer
export default userSlice.reducer;

// 选择器
export const selectCurrentUser = (state: { users: UserState }) => state.users.currentUser;
export const selectUsers = (state: { users: UserState }) => state.users.users;
export const selectUserLoading = (state: { users: UserState }) => state.users.loading;
export const selectUserError = (state: { users: UserState }) => state.users.error;
export const selectUserPagination = (state: { users: UserState }) => state.users.pagination;
export const selectUserFilters = (state: { users: UserState }) => state.users.filters;
export const selectUserSorting = (state: { users: UserState }) => state.users.sorting;

// 导出类型
export type { UserState, UserService, UserListParams, UserListResponse, CreateUserRequest, UpdateUserRequest };