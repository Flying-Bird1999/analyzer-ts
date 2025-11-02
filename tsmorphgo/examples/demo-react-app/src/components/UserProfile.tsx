import React, { useState, useEffect, useCallback } from 'react';
import { User, type UserProfile as UserProfileType, type UserRole } from '../types';
import { formatDate } from '../utils/date';
import { Validation } from '../utils/validation';

interface UserProfileProps {
    userId?: string;
    user?: User;
    onEdit?: (user: User) => void;
    onDelete?: (userId: string) => void;
    showAdvancedDetails?: boolean;
    allowEdit?: boolean;
    className?: string;
}

interface UserProfileState {
    user: User | null;
    profile: UserProfileType | null;
    loading: boolean;
    error: string | null;
    isEditing: boolean;
    editedProfile: Partial<UserProfileType>;
    showConfirmDelete: boolean;
}

const UserProfile: React.FC<UserProfileProps> = ({
    userId,
    user: propUser,
    onEdit,
    onDelete,
    showAdvancedDetails = false,
    allowEdit = true,
    className = ''
}) => {
    const [state, setState] = useState<UserProfileState>({
        user: propUser || null,
        profile: null,
        loading: false,
        error: null,
        isEditing: false,
        editedProfile: {},
        showConfirmDelete: false
    });

    // 模拟用户服务
    const userService = {
        getUserById: async (id: string): Promise<User | null> => {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 500));
            return {
                id,
                name: 'John Doe',
                email: 'john.doe@example.com',
                active: true,
                role: UserRole.USER,
                createdAt: new Date('2023-01-01'),
                updatedAt: new Date('2023-01-15'),
                profile: {
                    avatar: 'https://example.com/avatar.jpg',
                    bio: 'Software developer with 5+ years of experience in React and TypeScript.',
                    website: 'https://johndoe.com',
                    social: {
                        twitter: '@johndoe',
                        github: 'johndoe',
                        linkedin: 'johndoe'
                    }
                }
            };
        },

        updateUserProfile: async (userId: string, profile: Partial<UserProfileType>): Promise<User> => {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 300));
            return {
                ...state.user!,
                profile: { ...state.user!.profile, ...profile }
            } as User;
        },

        deleteUser: async (userId: string): Promise<void> => {
            // 模拟API调用
            await new Promise(resolve => setTimeout(resolve, 800));
        }
    };

    // 加载用户数据
    const loadUserData = useCallback(async (): Promise<void> => {
        if (!userId || propUser) return;

        setState(prev => ({ ...prev, loading: true, error: null }));

        try {
            const userData = await userService.getUserById(userId);
            setState(prev => ({
                ...prev,
                user: userData,
                profile: userData?.profile || null,
                loading: false
            }));
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load user data'
            }));
        }
    }, [userId, propUser]);

    // 初始化加载
    useEffect(() => {
        if (userId && !propUser) {
            loadUserData();
        }
    }, [loadUserData, userId, propUser]);

    // 处理编辑开始
    const handleEditStart = useCallback((): void => {
        setState(prev => ({
            ...prev,
            isEditing: true,
            editedProfile: { ...prev.profile } || {}
        }));
    }, []);

    // 处理编辑取消
    const handleEditCancel = useCallback((): void => {
        setState(prev => ({
            ...prev,
            isEditing: false,
            editedProfile: {}
        }));
    }, []);

    // 处理输入变更
    const handleInputChange = useCallback((
        field: keyof UserProfileType,
        value: string
    ): void => {
        setState(prev => ({
            ...prev,
            editedProfile: {
                ...prev.editedProfile,
                [field]: value
            }
        }));
    }, []);

    // 处理社交链接变更
    const handleSocialInputChange = useCallback((
        platform: keyof UserProfileType['social'],
        value: string
    ): void => {
        setState(prev => ({
            ...prev,
            editedProfile: {
                ...prev.editedProfile,
                social: {
                    ...prev.editedProfile.social!,
                    [platform]: value
                }
            }
        }));
    }, []);

    // 处理保存
    const handleSave = useCallback(async (): Promise<void> => {
        if (!state.user || !state.editedProfile) return;

        setState(prev => ({ ...prev, loading: true, error: null }));

        try {
            // 验证输入
            const validationErrors: string[] = [];

            if (state.editedProfile.website && !Validation.isValidUrl(state.editedProfile.website)) {
                validationErrors.push('Invalid website URL');
            }

            if (state.editedProfile.avatar && !Validation.isValidUrl(state.editedProfile.avatar)) {
                validationErrors.push('Invalid avatar URL');
            }

            if (state.editedProfile.social) {
                Object.entries(state.editedProfile.social).forEach(([platform, value]) => {
                    if (value && !Validation.isNonEmptyString(value)) {
                        validationErrors.push(`Invalid ${platform} handle`);
                    }
                });
            }

            if (validationErrors.length > 0) {
                setState(prev => ({
                    ...prev,
                    loading: false,
                    error: validationErrors.join(', ')
                }));
                return;
            }

            // 更新用户档案
            const updatedUser = await userService.updateUserProfile(
                state.user.id,
                state.editedProfile
            );

            setState(prev => ({
                ...prev,
                user: updatedUser,
                profile: updatedUser.profile,
                isEditing: false,
                editedProfile: {},
                loading: false
            }));

            onEdit?.(updatedUser);
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to update profile'
            }));
        }
    }, [state.user, state.editedProfile, onEdit]);

    // 处理删除开始
    const handleDeleteStart = useCallback((): void => {
        setState(prev => ({ ...prev, showConfirmDelete: true }));
    }, []);

    // 处理删除确认
    const handleDeleteConfirm = useCallback(async (): Promise<void> => {
        if (!state.user) return;

        setState(prev => ({ ...prev, loading: true }));

        try {
            await userService.deleteUser(state.user.id);
            onDelete?.(state.user.id);
        } catch (error) {
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to delete user',
                showConfirmDelete: false
            }));
        }
    }, [state.user, onDelete]);

    // 处理删除取消
    const handleDeleteCancel = useCallback((): void => {
        setState(prev => ({ ...prev, showConfirmDelete: false }));
    }, []);

    // 渲染加载状态
    if (state.loading) {
        return (
            <div className={`user-profile-loading ${className}`}>
                <div className="loading-spinner">Loading user profile...</div>
            </div>
        );
    }

    // 渲染错误状态
    if (state.error) {
        return (
            <div className={`user-profile-error ${className}`}>
                <div className="error-message">{state.error}</div>
                <button onClick={loadUserData} className="retry-button">
                    Retry
                </button>
            </div>
        );
    }

    // 渲染用户不存在状态
    if (!state.user) {
        return (
            <div className={`user-profile-not-found ${className}`}>
                <div className="not-found-message">User not found</div>
            </div>
        );
    }

    // 渲染主界面
    return (
        <div className={`user-profile ${className}`}>
            {/* 头部操作栏 */}
            <div className="profile-header">
                <h2>{state.user.name}</h2>
                <div className="header-actions">
                    {allowEdit && !state.isEditing && (
                        <button onClick={handleEditStart} className="edit-button">
                            Edit Profile
                        </button>
                    )}
                    {onDelete && (
                        <button onClick={handleDeleteStart} className="delete-button">
                            Delete User
                        </button>
                    )}
                </div>
            </div>

            {/* 基本信息 */}
            <div className="profile-basic-info">
                <div className="info-item">
                    <label>Email:</label>
                    <span>{state.user.email}</span>
                </div>
                <div className="info-item">
                    <label>Role:</label>
                    <span className={`role-badge role-${state.user.role}`}>
                        {state.user.role}
                    </span>
                </div>
                <div className="info-item">
                    <label>Status:</label>
                    <span className={`status-badge status-${state.user.active ? 'active' : 'inactive'}`}>
                        {state.user.active ? 'Active' : 'Inactive'}
                    </span>
                </div>
                <div className="info-item">
                    <label>Joined:</label>
                    <span>{formatDate(state.user.createdAt)}</span>
                </div>
            </div>

            {/* 档案详情 */}
            {showAdvancedDetails && state.profile && (
                <div className="profile-details">
                    {state.isEditing ? (
                        <div className="profile-edit-form">
                            <h3>Edit Profile</h3>

                            <div className="form-group">
                                <label htmlFor="avatar">Avatar URL:</label>
                                <input
                                    type="url"
                                    id="avatar"
                                    value={state.editedProfile.avatar || ''}
                                    onChange={(e) => handleInputChange('avatar', e.target.value)}
                                    placeholder="https://example.com/avatar.jpg"
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="bio">Bio:</label>
                                <textarea
                                    id="bio"
                                    value={state.editedProfile.bio || ''}
                                    onChange={(e) => handleInputChange('bio', e.target.value)}
                                    placeholder="Tell us about yourself..."
                                    rows={4}
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="website">Website:</label>
                                <input
                                    type="url"
                                    id="website"
                                    value={state.editedProfile.website || ''}
                                    onChange={(e) => handleInputChange('website', e.target.value)}
                                    placeholder="https://yourwebsite.com"
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="twitter">Twitter:</label>
                                <input
                                    type="text"
                                    id="twitter"
                                    value={state.editedProfile.social?.twitter || ''}
                                    onChange={(e) => handleSocialInputChange('twitter', e.target.value)}
                                    placeholder="@username"
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="github">GitHub:</label>
                                <input
                                    type="text"
                                    id="github"
                                    value={state.editedProfile.social?.github || ''}
                                    onChange={(e) => handleSocialInputChange('github', e.target.value)}
                                    placeholder="username"
                                />
                            </div>

                            <div className="form-group">
                                <label htmlFor="linkedin">LinkedIn:</label>
                                <input
                                    type="text"
                                    id="linkedin"
                                    value={state.editedProfile.social?.linkedin || ''}
                                    onChange={(e) => handleSocialInputChange('linkedin', e.target.value)}
                                    placeholder="username"
                                />
                            </div>

                            <div className="form-actions">
                                <button onClick={handleEditCancel} className="cancel-button">
                                    Cancel
                                </button>
                                <button onClick={handleSave} className="save-button" disabled={state.loading}>
                                    {state.loading ? 'Saving...' : 'Save Changes'}
                                </button>
                            </div>
                        </div>
                    ) : (
                        <div className="profile-display">
                            {state.profile.avatar && (
                                <div className="profile-avatar">
                                    <img
                                        src={state.profile.avatar}
                                        alt={`${state.user.name}'s avatar`}
                                        className="avatar-image"
                                    />
                                </div>
                            )}

                            {state.profile.bio && (
                                <div className="profile-bio">
                                    <h3>Bio</h3>
                                    <p>{state.profile.bio}</p>
                                </div>
                            )}

                            {state.profile.website && (
                                <div className="profile-website">
                                    <h3>Website</h3>
                                    <a
                                        href={state.profile.website}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="website-link"
                                    >
                                        Visit Website
                                    </a>
                                </div>
                            )}

                            {(state.profile.social?.twitter || state.profile.social?.github || state.profile.social?.linkedin) && (
                                <div className="profile-social-links">
                                    <h3>Social Links</h3>
                                    <div className="social-links">
                                        {state.profile.social.twitter && (
                                            <a
                                                href={`https://twitter.com/${state.profile.social.twitter}`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="social-link twitter"
                                            >
                                                Twitter
                                            </a>
                                        )}
                                        {state.profile.social.github && (
                                            <a
                                                href={`https://github.com/${state.profile.social.github}`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="social-link github"
                                            >
                                                GitHub
                                            </a>
                                        )}
                                        {state.profile.social.linkedin && (
                                            <a
                                                href={`https://linkedin.com/in/${state.profile.social.linkedin}`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="social-link linkedin"
                                            >
                                                LinkedIn
                                            </a>
                                        )}
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}

            {/* 删除确认对话框 */}
            {state.showConfirmDelete && (
                <div className="delete-confirm-modal">
                    <div className="modal-content">
                        <h3>Confirm Delete</h3>
                        <p>Are you sure you want to delete this user? This action cannot be undone.</p>
                        <div className="modal-actions">
                            <button onClick={handleDeleteCancel} className="cancel-button">
                                Cancel
                            </button>
                            <button onClick={handleDeleteConfirm} className="confirm-button">
                                Delete
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default UserProfile;
