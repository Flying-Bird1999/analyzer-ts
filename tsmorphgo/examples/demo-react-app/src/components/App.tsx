import React, { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { UserProfile } from './components/UserProfile';
import { ProductList } from './components/ProductList';
import { useUserData } from './hooks/useUserData';
import { useApiService } from './hooks/useApiService';
import { formatDate } from './utils/dateUtils';

// 导入样式
import './App.css';

interface Product {
  id: number;
  name: string;
  price: number;
  category: string;
}

interface User {
  id: number;
  name: string;
  email: string;
  avatar: string;
}

export const App: React.FC = () => {
  // 使用自定义Hook获取用户数据
  const { user, loading: userLoading, error: userError } = useUserData(1);

  // 使用API服务Hook
  const { fetchProducts, fetchUser } = useApiService();

  // 组件状态
  const [products, setProducts] = useState<Product[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState<'name' | 'price'>('name');

  // 模拟商品数据
  useEffect(() => {
    const mockProducts: Product[] = [
      { id: 1, name: 'iPhone 15', price: 999, category: 'electronics' },
      { id: 2, name: 'MacBook Pro', price: 1999, category: 'electronics' },
      { id: 3, name: 'AirPods Pro', price: 249, category: 'electronics' },
      { id: 4, name: 'iPad Air', price: 599, category: 'electronics' },
      { id: 5, name: 'Apple Watch', price: 399, category: 'electronics' },
    ];

    setProducts(mockProducts);
  }, []);

  // 处理搜索
  const handleSearch = (term: string) => {
    setSearchTerm(term);
  };

  // 处理排序
  const handleSort = (criteria: 'name' | 'price') => {
    setSortBy(criteria);
  };

  // 过滤和排序商品
  const filteredProducts = products
    .filter(product =>
      product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      product.category.toLowerCase().includes(searchTerm.toLowerCase())
    )
    .sort((a, b) => {
      if (sortBy === 'name') {
        return a.name.localeCompare(b.name);
      }
      return a.price - b.price;
    });

  // 获取格式化的当前日期
  const currentDate = formatDate(new Date());

  if (userLoading) {
    return <div className="loading">Loading...</div>;
  }

  if (userError) {
    return <div className="error">Error: {userError}</div>;
  }

  return (
    <div className="app">
      <Header
        user={user}
        currentDate={currentDate}
      />

      <main className="main-content">
        <div className="user-section">
          {user && <UserProfile user={user} />}
        </div>

        <div className="products-section">
          <div className="products-header">
            <h2>Products</h2>
            <div className="products-controls">
              <input
                type="text"
                placeholder="Search products..."
                value={searchTerm}
                onChange={(e) => handleSearch(e.target.value)}
                className="search-input"
              />
              <button
                onClick={() => handleSort('name')}
                className={`sort-btn ${sortBy === 'name' ? 'active' : ''}`}
              >
                Sort by Name
              </button>
              <button
                onClick={() => handleSort('price')}
                className={`sort-btn ${sortBy === 'price' ? 'active' : ''}`}
              >
                Sort by Price
              </button>
            </div>
          </div>

          <ProductList products={filteredProducts} />
        </div>
      </main>
    </div>
  );
};

export default App;