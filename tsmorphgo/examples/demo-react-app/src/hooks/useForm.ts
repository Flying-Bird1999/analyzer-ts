import { useState, useCallback, useEffect } from 'react';
import { FormState } from '../types/types';

// 验证函数类型
export type ValidatorFunction<T> = (value: T[keyof T]) => string | undefined;

// 验证器配置
export type Validators<T> = {
  [K in keyof T]?: ValidatorFunction<T>[];
};

// 钩子配置
export interface UseFormConfig<T> {
  initialValues: T;
  validators?: any;
  onSubmit?: (values: T) => Promise<void> | void;
  validateOnChange?: boolean;
  validateOnBlur?: boolean;
}

// 钩子返回值
export interface UseFormReturn<T> {
  formState: FormState<T>;
  setValue: <K extends keyof T>(field: K, value: T[K]) => void;
  setError: <K extends keyof T>(field: K, error: string | undefined) => void;
  setTouched: <K extends keyof T>(field: K, touched: boolean) => void;
  validateField: <K extends keyof T>(field: K) => string | undefined;
  validateForm: () => boolean;
  handleSubmit: (onSubmit?: (values: T) => Promise<void> | void) => (e?: React.FormEvent) => Promise<void>;
  resetForm: () => void;
  isDirty: boolean;
}

// 自定义表单钩子
export function useForm<T extends Record<string, any>>(
  config: UseFormConfig<T>
): UseFormReturn<T> {
  const { initialValues, validators = {}, validateOnChange = true, validateOnBlur = true } = config;

  // 初始化表单状态
  const initializeFormState = (): FormState<T> => ({
    values: { ...initialValues },
    errors: {},
    touched: {},
    isSubmitting: false,
    isValid: true,
  });

  const [formState, setFormState] = useState<FormState<T>>(initializeFormState);
  const [isDirty, setIsDirty] = useState(false);

  // 验证单个字段
  const validateField = useCallback(<K extends keyof T>(field: K): string | undefined => {
    const fieldValidators = validators as Validators<T>[K];
    if (!fieldValidators || fieldValidators.length === 0) {
      return undefined;
    }

    const value = formState.values[field];
    for (const validator of fieldValidators) {
      const error = validator(value);
      if (error) {
        return error;
      }
    }

    return undefined;
  }, [formState.values, validators]);

  // 验证整个表单
  const validateForm = useCallback((): boolean => {
    const newErrors: Partial<Record<keyof T, string>> = {};

    // 验证所有字段
    Object.keys(validators).forEach((field) => {
      const error = validateField(field as keyof T);
      if (error) {
        newErrors[field as keyof T] = error;
      }
    });

    setFormState(prev => ({
      ...prev,
      errors: newErrors,
      isValid: Object.keys(newErrors).length === 0,
    }));

    return Object.keys(newErrors).length === 0;
  }, [validators, validateField]);

  // 设置字段值
  const setValue = useCallback(<K extends keyof T>(field: K, value: T[K]) => {
    setFormState(prev => {
      const newValues = { ...prev.values, [field]: value };
      const newErrors = { ...prev.errors };

      // 如果启用了实时验证，验证当前字段
      if (validateOnChange) {
        const error = validators[field]?.reduce((acc: string | undefined, validator: any) => {
          return acc || validator(value);
        }, undefined);

        if (error) {
          newErrors[field] = error;
        } else {
          delete newErrors[field];
        }
      }

      return {
        ...prev,
        values: newValues,
        errors: newErrors,
        isValid: Object.keys(newErrors).length === 0,
      };
    });
    setIsDirty(true);
  }, [validators, validateOnChange]);

  // 设置字段错误
  const setError = useCallback(<K extends keyof T>(field: K, error: string | undefined) => {
    setFormState(prev => {
      const newErrors = { ...prev.errors };
      if (error) {
        newErrors[field] = error;
      } else {
        delete newErrors[field];
      }
      return {
        ...prev,
        errors: newErrors,
        isValid: Object.keys(newErrors).length === 0,
      };
    });
  }, []);

  // 设置字段触摸状态
  const setTouched = useCallback(<K extends keyof T>(field: K, touched: boolean) => {
    setFormState(prev => {
      const newTouched = { ...prev.touched, [field]: touched };
      const newErrors = { ...prev.errors };

      // 如果启用了失焦验证，验证当前字段
      if (touched && validateOnBlur) {
        const error = validateField(field);
        if (error) {
          newErrors[field] = error;
        } else {
          delete newErrors[field];
        }
      }

      return {
        ...prev,
        touched: newTouched,
        errors: newErrors,
        isValid: Object.keys(newErrors).length === 0,
      };
    });
  }, [validateField, validateOnBlur]);

  // 处理表单提交
  const handleSubmit = useCallback((onSubmit?: (values: T) => Promise<void> | void) => {
    return async (e?: React.FormEvent) => {
      e?.preventDefault();

      // 验证表单
      const isValid = validateForm();
      if (!isValid) {
        return;
      }

      setFormState(prev => ({ ...prev, isSubmitting: true }));

      try {
        const submitFn = onSubmit || config.onSubmit;
        if (submitFn) {
          await submitFn(formState.values);
        }
      } catch (error) {
        console.error('Form submission error:', error);
      } finally {
        setFormState(prev => ({ ...prev, isSubmitting: false }));
      }
    };
  }, [validateForm, formState.values, config.onSubmit]);

  // 重置表单
  const resetForm = useCallback(() => {
    setFormState(initializeFormState());
    setIsDirty(false);
  }, []);

  return {
    formState,
    setValue,
    setError,
    setTouched,
    validateField,
    validateForm,
    handleSubmit,
    resetForm,
    isDirty,
  };
}

// 常用验证器
export const validators = {
  required: (message = 'This field is required') => (value: any) => {
    if (value === undefined || value === null || value === '') {
      return message;
    }
    return undefined;
  },

  minLength: (min: number, message?: string) => (value: string) => {
    if (value && value.length < min) {
      return message || `Must be at least ${min} characters`;
    }
    return undefined;
  },

  maxLength: (max: number, message?: string) => (value: string) => {
    if (value && value.length > max) {
      return message || `Must be no more than ${max} characters`;
    }
    return undefined;
  },

  email: (message = 'Must be a valid email address') => (value: string) => {
    if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      return message;
    }
    return undefined;
  },

  min: (min: number, message?: string) => (value: number) => {
    if (value !== undefined && value < min) {
      return message || `Must be at least ${min}`;
    }
    return undefined;
  },

  max: (max: number, message?: string) => (value: number) => {
    if (value !== undefined && value > max) {
      return message || `Must be no more than ${max}`;
    }
    return undefined;
  },

  pattern: (regex: RegExp, message = 'Invalid format') => (value: string) => {
    if (value && !regex.test(value)) {
      return message;
    }
    return undefined;
  },
};