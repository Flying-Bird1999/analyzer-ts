import _ from 'lodash';

// Utility functions
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): T => {
  return _.debounce(func, wait) as T;
};

export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): T => {
  return _.throttle(func, wait) as T;
};

export const formatNumber = (num: number): string => {
  return _.toNumber(num.toFixed(2)).toLocaleString();
};

export const deepClone = <T>(obj: T): T => {
  return _.cloneDeep(obj);
};

export const mergeObjects = <T>(target: T, source: Partial<T>): T => {
  return _.merge(target, source);
};

// Array utilities
export const chunk = <T>(array: T[], size: number): T[][] => {
  return _.chunk(array, size);
};

export const groupBy = <T>(array: T[], key: keyof T): Record<string, T[]> => {
  return _.groupBy(array, key);
};

export const sortBy = <T>(array: T[], key: keyof T): T[] => {
  return _.sortBy(array, key);
};

// String utilities
export const capitalize = (str: string): string => {
  return _.capitalize(str);
};

export const camelCase = (str: string): string => {
  return _.camelCase(str);
};

export const kebabCase = (str: string): string => {
  return _.kebabCase(str);
};

export const snakeCase = (str: string): string => {
  return _.snakeCase(str);
};