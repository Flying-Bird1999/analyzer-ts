import React, { useEffect, useCallback } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import type { User, UserRole, UserProfile as UserProfileType } from '../types';

// 表单数据接口
interface UserFormData {
    name: string;
    email: string;
    role: UserRole;
    active: boolean;
    profile: {
        avatar?: string;
        bio?: string;
        website?: string;
        social?: {
            twitter?: string;
            github?: string;
            linkedin?: string;
        };
    };
}

// 表单验证模式
const userFormSchema = yup.object({
    name: yup
        .string()
        .required('Name is required')
        .min(2, 'Name must be at least 2 characters')
        .max(50, 'Name must not exceed 50 characters')
        .matches(/^[a-zA-Z\s'-]+$/, 'Name can only contain letters, spaces, hyphens, and apostrophes'),

    email: yup
        .string()
        .required('Email is required')
        .email('Please enter a valid email address')
        .max(100, 'Email must not exceed 100 characters'),

    role: yup
        .mixed<UserRole>()
        .required('Role is required')
        .oneOf(Object.values(UserRole), 'Please select a valid role'),

    active: yup
        .boolean()
        .required('Active status is required'),

    profile: yup.object({
        avatar: yup
            .string()
            .url('Please enter a valid URL for avatar')
            .max(500, 'Avatar URL must not exceed 500 characters')
            .optional(),

        bio: yup
            .string()
            .max(500, 'Bio must not exceed 500 characters')
            .optional(),

        website: yup
            .string()
            .url('Please enter a valid URL for website')
            .max(500, 'Website URL must not exceed 500 characters')
            .optional(),

        social: yup.object({
            twitter: yup
                .string()
                .matches(/^@?[\w]{1,15}$/, 'Twitter handle must be a valid Twitter username (1-15 characters, letters, numbers, underscores only)')
                .optional(),

            github: yup
                .string()
                .matches(/^[\w](?:[\w-]*[\w])?$/, 'GitHub username must be a valid GitHub username (letters, numbers, hyphens only, cannot start or end with hyphen)')
                .max(39, 'GitHub username must not exceed 39 characters')
                .optional(),

            linkedin: yup
                .string()
                .matches(/^[\w-]{2,100}$/, 'LinkedIn username must be a valid LinkedIn username (2-100 characters, letters, numbers, hyphens only)')
                .optional()
        }).optional()
    })
});

// 表单组件Props接口
interface UserFormProps {
    // 初始用户数据（用于编辑模式）
    initialData?: User | null;
    // 表单提交回调
    onSubmit: (data: UserFormData) => Promise<void>;
    // 表单取消回调
    onCancel?: () => void;
    // 是否显示高级字段
    showAdvancedFields?: boolean;
    // 表单加载状态
    loading?: boolean;
    // 表单模式
    mode?: 'create' | 'edit';
    // 表单样式类名
    className?: string;
    // 字段权限控制
    fieldPermissions?: {
        readonly?: boolean;
        required?: Partial<Record<keyof UserFormData, boolean>>;
        disabled?: Partial<Record<keyof UserFormData, boolean>>;
    };
}

// 字段配置接口
interface FieldConfig {
    label: string;
    placeholder?: string;
    helpText?: string;
    required?: boolean;
    disabled?: boolean;
    readonly?: boolean;
    className?: string;
    inputProps?: React.InputHTMLAttributes<HTMLInputElement>;
}

// 表单字段组件
const FormField: React.FC<{
    label: string;
    error?: string;
    helpText?: string;
    required?: boolean;
    children: React.ReactNode;
    className?: string;
}> = ({ label, error, helpText, required, children, className }) => (
    <div className={`form-field ${className || ''}`}>
        <label className={`field-label ${required ? 'required' : ''}`}>
            {label}
        </label>
        <div className="field-input-container">
            {children}
            {error && <div className="field-error">{error}</div>}
            {helpText && <div className="field-help">{helpText}</div>}
        </div>
    </div>
);

// 高级表单组件 - 支持复杂的用户数据管理
const UserForm: React.FC<UserFormProps> = ({
    initialData,
    onSubmit,
    onCancel,
    showAdvancedFields = false,
    loading = false,
    mode = 'create',
    className = '',
    fieldPermissions
}) => {
    // 初始化React Hook Form
    const {
        register,
        control,
        handleSubmit,
        reset,
        formState: { errors, isDirty, isValid, isSubmitting },
        watch,
        setValue,
        trigger
    } = useForm<UserFormData>({
        resolver: yupResolver(userFormSchema),
        mode: 'onChange',
        defaultValues: {
            name: '',
            email: '',
            role: UserRole.USER,
            active: true,
            profile: {
                avatar: '',
                bio: '',
                website: '',
                social: {
                    twitter: '',
                    github: '',
                    linkedin: ''
                }
            }
        }
    });

    // 监听角色变化
    const selectedRole = watch('role');

    // 表单字段配置
    const fieldConfigs: Record<keyof UserFormData, FieldConfig> = {
        name: {
            label: 'Full Name',
            placeholder: 'Enter full name',
            helpText: 'Use your real name for better recognition',
            required: true,
            readonly: fieldPermissions?.readonly
        },
        email: {
            label: 'Email Address',
            placeholder: 'Enter email address',
            helpText: 'We\'ll never share your email with anyone else',
            required: true,
            readonly: fieldPermissions?.readonly
        },
        role: {
            label: 'User Role',
            helpText: 'Select the appropriate role for this user',
            required: true,
            disabled: fieldPermissions?.disabled?.role
        },
        active: {
            label: 'Account Status',
            helpText: 'Active users can access the system',
            disabled: fieldPermissions?.disabled?.active
        },
        profile: {
            label: 'Profile Information',
            helpText: 'Optional profile details'
        }
    };

    // 初始化表单数据
    useEffect(() => {
        if (initialData) {
            reset({
                name: initialData.name,
                email: initialData.email,
                role: initialData.role,
                active: initialData.active,
                profile: {
                    avatar: initialData.profile?.avatar || '',
                    bio: initialData.profile?.bio || '',
                    website: initialData.profile?.website || '',
                    social: {
                        twitter: initialData.profile?.social?.twitter || '',
                        github: initialData.profile?.social?.github || '',
                        linkedin: initialData.profile?.social?.linkedin || ''
                    }
                }
            });
        }
    }, [initialData, reset]);

    // 表单提交处理
    const handleFormSubmit = useCallback(async (data: UserFormData) => {
        try {
            await onSubmit(data);
            if (mode === 'create') {
                reset(); // 创建成功后重置表单
            }
        } catch (error) {
            console.error('Form submission failed:', error);
        }
    }, [onSubmit, mode, reset]);

    // 表单取消处理
    const handleCancel = useCallback(() => {
        if (mode === 'edit' && initialData) {
            // 恢复到初始数据
            reset({
                name: initialData.name,
                email: initialData.email,
                role: initialData.role,
                active: initialData.active,
                profile: {
                    avatar: initialData.profile?.avatar || '',
                    bio: initialData.profile?.bio || '',
                    website: initialData.profile?.website || '',
                    social: {
                        twitter: initialData.profile?.social?.twitter || '',
                        github: initialData.profile?.social?.github || '',
                        linkedin: initialData.profile?.social?.linkedin || ''
                    }
                });
            }
        } else {
            reset();
        }
        onCancel?.();
    }, [mode, initialData, reset, onCancel]);

    // 自定义社交链接验证
    const validateSocialLink = useCallback((platform: keyof UserProfileType['social'], value: string) => {
        if (!value.trim()) return true;

        switch (platform) {
            case 'twitter':
                return /^@?[\w]{1,15}$/.test(value) || 'Invalid Twitter handle';
            case 'github':
                return /^[\w](?:[\w-]*[\w])?$/.test(value) || 'Invalid GitHub username';
            case 'linkedin':
                return /^[\w-]{2,100}$/.test(value) || 'Invalid LinkedIn username';
            default:
                return true;
        }
    }, []);

    // 表单重置处理
    const handleReset = useCallback(() => {
        if (mode === 'edit' && initialData) {
            handleCancel();
        } else {
            reset();
        }
    }, [mode, initialData, handleCancel, reset]);

    // 渲染基础信息字段
    const renderBasicFields = () => (
        <div className="form-section">
            <h3>Basic Information</h3>

            <FormField
                label={fieldConfigs.name.label}
                error={errors.name?.message}
                helpText={fieldConfigs.name.helpText}
                required={fieldConfigs.name.required}
                className="field-name"
            >
                <input
                    type="text"
                    {...register('name')}
                    placeholder={fieldConfigs.name.placeholder}
                    disabled={fieldConfigs.name.readonly || isSubmitting}
                    className={`form-input ${errors.name ? 'error' : ''}`}
                />
            </FormField>

            <FormField
                label={fieldConfigs.email.label}
                error={errors.email?.message}
                helpText={fieldConfigs.email.helpText}
                required={fieldConfigs.email.required}
                className="field-email"
            >
                <input
                    type="email"
                    {...register('email')}
                    placeholder={fieldConfigs.email.placeholder}
                    disabled={fieldConfigs.email.readonly || isSubmitting}
                    className={`form-input ${errors.email ? 'error' : ''}`}
                />
            </FormField>

            <FormField
                label={fieldConfigs.role.label}
                error={errors.role?.message}
                helpText={fieldConfigs.role.helpText}
                required={fieldConfigs.role.required}
                className="field-role"
            >
                <select
                    {...register('role')}
                    disabled={fieldConfigs.role.disabled || isSubmitting}
                    className={`form-select ${errors.role ? 'error' : ''}`}
                >
                    {Object.values(UserRole).map(role => (
                        <option key={role} value={role}>
                            {role}
                        </option>
                    ))}
                </select>
            </FormField>

            <FormField
                label={fieldConfigs.active.label}
                error={errors.active?.message}
                helpText={fieldConfigs.active.helpText}
                className="field-active"
            >
                <label className="checkbox-label">
                    <input
                        type="checkbox"
                        {...register('active')}
                        disabled={fieldConfigs.active.disabled || isSubmitting}
                        className="form-checkbox"
                    />
                    <span>Account is active</span>
                </label>
            </FormField>
        </div>
    );

    // 渲染高级字段
    const renderAdvancedFields = () => (
        <div className="form-section">
            <h3>Profile Information</h3>

            <FormField
                label="Avatar URL"
                error={errors.profile?.avatar?.message}
                helpText="URL to your profile picture"
                className="field-avatar"
            >
                <Controller
                    name="profile.avatar"
                    control={control}
                    render={({ field }) => (
                        <div className="avatar-input-container">
                            <input
                                type="url"
                                {...field}
                                placeholder="https://example.com/avatar.jpg"
                                disabled={isSubmitting}
                                className={`form-input ${errors.profile?.avatar ? 'error' : ''}`}
                            />
                            {field.value && (
                                <img
                                    src={field.value}
                                    alt="Avatar preview"
                                    className="avatar-preview"
                                    onError={(e) => {
                                        e.currentTarget.style.display = 'none';
                                    }}
                                />
                            )}
                        </div>
                    )}
                />
            </FormField>

            <FormField
                label="Bio"
                error={errors.profile?.bio?.message}
                helpText="Tell us about yourself (max 500 characters)"
                className="field-bio"
            >
                <Controller
                    name="profile.bio"
                    control={control}
                    render={({ field }) => (
                        <textarea
                            {...field}
                            placeholder="Write a short bio..."
                            disabled={isSubmitting}
                            rows={4}
                            maxLength={500}
                            className={`form-textarea ${errors.profile?.bio ? 'error' : ''}`}
                        />
                    )}
                />
            </FormField>

            <FormField
                label="Website"
                error={errors.profile?.website?.message}
                helpText="Your personal or professional website"
                className="field-website"
            >
                <Controller
                    name="profile.website"
                    control={control}
                    render={({ field }) => (
                        <input
                            type="url"
                            {...field}
                            placeholder="https://yourwebsite.com"
                            disabled={isSubmitting}
                            className={`form-input ${errors.profile?.website ? 'error' : ''}`}
                        />
                    )}
                />
            </FormField>

            <div className="form-section">
                <h3>Social Media Links</h3>

                <FormField
                    label="Twitter"
                    error={errors.profile?.social?.twitter?.message}
                    helpText="Your Twitter username (optional)"
                    className="field-twitter"
                >
                    <Controller
                        name="profile.social.twitter"
                        control={control}
                        render={({ field }) => (
                            <input
                                type="text"
                                {...field}
                                placeholder="@username"
                                disabled={isSubmitting}
                                className={`form-input ${errors.profile?.social?.twitter ? 'error' : ''}`}
                            />
                        )}
                    />
                </FormField>

                <FormField
                    label="GitHub"
                    error={errors.profile?.social?.github?.message}
                    helpText="Your GitHub username (optional)"
                    className="field-github"
                >
                    <Controller
                        name="profile.social.github"
                        control={control}
                        render={({ field }) => (
                            <input
                                type="text"
                                {...field}
                                placeholder="username"
                                disabled={isSubmitting}
                                className={`form-input ${errors.profile?.social?.github ? 'error' : ''}`}
                            />
                        )}
                    />
                </FormField>

                <FormField
                    label="LinkedIn"
                    error={errors.profile?.social?.linkedin?.message}
                    helpText="Your LinkedIn username (optional)"
                    className="field-linkedin"
                >
                    <Controller
                        name="profile.social.linkedin"
                        control={control}
                        render={({ field }) => (
                            <input
                                type="text"
                                {...field}
                                placeholder="username"
                                disabled={isSubmitting}
                                className={`form-input ${errors.profile?.social?.linkedin ? 'error' : ''}`}
                            />
                        )}
                    />
                </FormField>
            </div>
        </div>
    );

    // 渲染表单操作按钮
    const renderFormActions = () => (
        <div className="form-actions">
            <button
                type="button"
                onClick={handleCancel}
                disabled={isSubmitting}
                className="btn btn-secondary"
            >
                Cancel
            </button>

            <button
                type="button"
                onClick={handleReset}
                disabled={isSubmitting || !isDirty}
                className="btn btn-outline"
            >
                Reset
            </button>

            <button
                type="submit"
                disabled={!isValid || !isDirty || isSubmitting}
                className="btn btn-primary"
            >
                {isSubmitting ? 'Saving...' : mode === 'create' ? 'Create User' : 'Update User'}
            </button>
        </div>
    );

    return (
        <form onSubmit={handleSubmit(handleFormSubmit)} className={`user-form ${className}`}>
            <div className="form-header">
                <h2>{mode === 'create' ? 'Create New User' : 'Edit User'}</h2>
                {isDirty && (
                    <div className="form-status">
                        <span className="status-changed">Unsaved changes</span>
                    </div>
                )}
            </div>

            {renderBasicFields()}

            {showAdvancedFields && renderAdvancedFields()}

            {renderFormActions()}

            {loading && (
                <div className="form-overlay">
                    <div className="loading-spinner">Processing...</div>
                </div>
            )}
        </form>
    );
};

export default UserForm;