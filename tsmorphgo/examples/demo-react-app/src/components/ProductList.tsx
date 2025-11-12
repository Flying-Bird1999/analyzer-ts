import React, { useState, useEffect, useMemo } from 'react';
import { Product, SearchFilters, ButtonProps } from '../types/types';
import { api } from '../services/api';

interface ProductListProps {
  category?: string;
  onProductSelect?: (product: Product) => void;
}

export const ProductList: React.FC<ProductListProps> = ({
  category,
  onProductSelect
}) => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<SearchFilters>({
    sortBy: 'name',
    sortOrder: 'asc'
  });

  // 获取产品列表
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoading(true);
        setError(null);

        const response = await api.getProducts({ ...filters, category });
        setProducts(response.data.items);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch products');
        setProducts([]);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, [filters, category]);

  // 过滤和排序后的产品列表
  const filteredAndSortedProducts = useMemo(() => {
    let result = [...products];

    // 应用搜索过滤
    if (filters.query) {
      result = result.filter(product =>
        product.name.toLowerCase().includes(filters.query!.toLowerCase()) ||
        product.category.toLowerCase().includes(filters.query!.toLowerCase())
      );
    }

    // 应用价格过滤
    if (filters.minPrice !== undefined) {
      result = result.filter(product => product.price >= filters.minPrice!);
    }
    if (filters.maxPrice !== undefined) {
      result = result.filter(product => product.price <= filters.maxPrice!);
    }

    // 应用库存过滤
    if (filters.inStock !== undefined) {
      result = result.filter(product => product.inStock === filters.inStock);
    }

    // 应用排序
    result.sort((a, b) => {
      const { sortBy = 'name', sortOrder = 'asc' } = filters;

      let aValue: any = a[sortBy as keyof Product];
      let bValue: any = b[sortBy as keyof Product];

      if (typeof aValue === 'string') {
        aValue = aValue.toLowerCase();
        bValue = (bValue as string).toLowerCase();
      }

      if (sortOrder === 'desc') {
        return aValue < bValue ? 1 : -1;
      }
      return aValue > bValue ? 1 : -1;
    });

    return result;
  }, [products, filters]);

  const handleFilterChange = (newFilters: Partial<SearchFilters>) => {
    setFilters(prev => ({ ...prev, ...newFilters }));
  };

  if (loading) {
    return <div className="product-list loading">Loading products...</div>;
  }

  if (error) {
    return <div className="product-list error">Error: {error}</div>;
  }

  return (
    <div className="product-list">
      <div className="product-list__filters">
        <input
          type="text"
          placeholder="Search products..."
          value={filters.query || ''}
          onChange={(e) => handleFilterChange({ query: e.target.value })}
          className="search-input"
        />

        <select
          value={filters.sortBy || 'name'}
          onChange={(e) => handleFilterChange({ sortBy: e.target.value as any })}
          className="sort-select"
        >
          <option value="name">Sort by Name</option>
          <option value="price">Sort by Price</option>
          <option value="createdAt">Sort by Date</option>
        </select>

        <select
          value={filters.sortOrder || 'asc'}
          onChange={(e) => handleFilterChange({ sortOrder: e.target.value as any })}
          className="order-select"
        >
          <option value="asc">Ascending</option>
          <option value="desc">Descending</option>
        </select>

        <label className="checkbox-label">
          <input
            type="checkbox"
            checked={filters.inStock ?? false}
            onChange={(e) => handleFilterChange({ inStock: e.target.checked })}
          />
          In Stock Only
        </label>
      </div>

      <div className="product-list__grid">
        {filteredAndSortedProducts.map((product) => (
          <div
            key={product.id}
            className="product-card"
            onClick={() => onProductSelect?.(product)}
          >
            <div className="product-card__image">
              <img
                src={product.images[0] || '/placeholder.jpg'}
                alt={product.name}
                loading="lazy"
              />
              {product.discount && (
                <span className="product-card__discount">
                  -{product.discount}%
                </span>
              )}
            </div>

            <div className="product-card__content">
              <h3 className="product-card__title">{product.name}</h3>
              <p className="product-card__category">{product.category}</p>
              {product.description && (
                <p className="product-card__description">{product.description}</p>
              )}

              <div className="product-card__price">
                {product.discount ? (
                  <>
                    <span className="original-price">${product.price}</span>
                    <span className="discounted-price">
                      ${(product.price * (1 - product.discount / 100)).toFixed(2)}
                    </span>
                  </>
                ) : (
                  <span className="current-price">${product.price}</span>
                )}
              </div>

              <div className="product-card__stock">
                <span className={`stock-status ${product.inStock ? 'in-stock' : 'out-of-stock'}`}>
                  {product.inStock ? 'In Stock' : 'Out of Stock'}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {filteredAndSortedProducts.length === 0 && (
        <div className="product-list__empty">
          <p>No products found matching your criteria.</p>
        </div>
      )}
    </div>
  );
};