import { ApiResponse, PaginatedResponse, User, Product, Order, SearchFilters } from '../types/types';

// API 配置
const API_BASE_URL = 'https://api.example.com';

// 错误类型
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// HTTP 方法枚举
export enum HttpMethod {
  GET = 'GET',
  POST = 'POST',
  PUT = 'PUT',
  DELETE = 'DELETE',
  PATCH = 'PATCH'
}

// 请求配置接口
export interface RequestConfig {
  method: HttpMethod;
  url: string;
  data?: any;
  params?: Record<string, any>;
  headers?: Record<string, string>;
}

// API 客户端类
export class ApiClient {
  private baseURL: string;
  private defaultHeaders: Record<string, string>;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    };
  }

  private async request<T>(config: RequestConfig): Promise<ApiResponse<T>> {
    const url = new URL(config.url, this.baseURL);

    // 添加查询参数
    if (config.params) {
      Object.entries(config.params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, String(value));
        }
      });
    }

    const response = await fetch(url.toString(), {
      method: config.method,
      headers: { ...this.defaultHeaders, ...config.headers },
      body: config.data ? JSON.stringify(config.data) : undefined,
    });

    const responseData = await response.json();

    if (!response.ok) {
      throw new ApiError(
        responseData.message || 'Request failed',
        response.status,
        responseData.code
      );
    }

    return responseData as ApiResponse<T>;
  }

  // 用户相关 API
  async getUser(userId: number): Promise<ApiResponse<User>> {
    return this.request<User>({
      method: HttpMethod.GET,
      url: `/users/${userId}`,
    });
  }

  async updateUser(userId: number, data: Partial<User>): Promise<ApiResponse<User>> {
    return this.request<User>({
      method: HttpMethod.PUT,
      url: `/users/${userId}`,
      data,
    });
  }

  async getUsers(page = 1, pageSize = 10): Promise<ApiResponse<PaginatedResponse<User>>> {
    return this.request<PaginatedResponse<User>>({
      method: HttpMethod.GET,
      url: '/users',
      params: { page, pageSize },
    });
  }

  // 产品相关 API
  async getProducts(filters?: SearchFilters): Promise<ApiResponse<PaginatedResponse<Product>>> {
    return this.request<PaginatedResponse<Product>>({
      method: HttpMethod.GET,
      url: '/products',
      params: filters,
    });
  }

  async getProduct(productId: number): Promise<ApiResponse<Product>> {
    return this.request<Product>({
      method: HttpMethod.GET,
      url: `/products/${productId}`,
    });
  }

  async createProduct(data: Omit<Product, 'id'>): Promise<ApiResponse<Product>> {
    return this.request<Product>({
      method: HttpMethod.POST,
      url: '/products',
      data,
    });
  }

  // 订单相关 API
  async getOrders(userId: number): Promise<ApiResponse<Order[]>> {
    return this.request<Order[]>({
      method: HttpMethod.GET,
      url: `/users/${userId}/orders`,
    });
  }

  async createOrder(data: Omit<Order, 'id' | 'createdAt'>): Promise<ApiResponse<Order>> {
    return this.request<Order>({
      method: HttpMethod.POST,
      url: '/orders',
      data,
    });
  }

  // 搜索 API
  async search(query: string, type: 'users' | 'products' = 'products'): Promise<ApiResponse<any[]>> {
    return this.request<any[]>({
      method: HttpMethod.GET,
      url: `/search/${type}`,
      params: { q: query },
    });
  }
}

// 默认 API 客户端实例
export const apiClient = new ApiClient();

// 便捷函数
export const api = {
  // 用户
  getUser: (id: number) => apiClient.getUser(id),
  updateUser: (id: number, data: Partial<User>) => apiClient.updateUser(id, data),
  getUsers: (page?: number, pageSize?: number) => apiClient.getUsers(page, pageSize),

  // 产品
  getProducts: (filters?: SearchFilters) => apiClient.getProducts(filters),
  getProduct: (id: number) => apiClient.getProduct(id),
  createProduct: (data: Omit<Product, 'id'>) => apiClient.createProduct(data),

  // 订单
  getOrders: (userId: number) => apiClient.getOrders(userId),
  createOrder: (data: Omit<Order, 'id' | 'createdAt'>) => apiClient.createOrder(data),

  // 搜索
  search: (query: string, type?: 'users' | 'products') => apiClient.search(query, type),
};