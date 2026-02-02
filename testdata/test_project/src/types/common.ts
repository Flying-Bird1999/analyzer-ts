// 通用类型定义

export interface Option {
  label: string;
  value: string;
  disabled?: boolean;
}

export interface Pagination {
  page: number;
  pageSize: number;
  total: number;
}

export interface SortConfig {
  key: string;
  direction: 'asc' | 'desc';
}

export interface FilterConfig {
  key: string;
  value: any;
  operator?: 'eq' | 'ne' | 'gt' | 'lt' | 'contains';
}
