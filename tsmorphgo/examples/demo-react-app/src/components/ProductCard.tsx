import React from 'react';
import { Product } from '@/types/types';

interface ProductCardProps {
  product: Product;
  onClick?: (product: Product) => void;
}

export const ProductCard: React.FC<ProductCardProps> = ({ product, onClick }) => {
  const discountedPrice = product.discount
    ? product.price * (1 - product.discount / 100)
    : product.price;

  return (
    <div
      className="product-card"
      onClick={() => onClick?.(product)}
    >
      <div className="product-image">
        {product.images[0] && (
          <img src={product.images[0]} alt={product.name} />
        )}
        {product.discount && (
          <span className="discount-badge">-{product.discount}%</span>
        )}
      </div>

      <div className="product-info">
        <h3 className="product-name">{product.name}</h3>
        <p className="product-category">{product.category}</p>

        <div className="product-price">
          {product.discount ? (
            <>
              <span className="original-price">${product.price}</span>
              <span className="discounted-price">${discountedPrice.toFixed(2)}</span>
            </>
          ) : (
            <span className="current-price">${product.price}</span>
          )}
        </div>

        <div className={`stock-status ${product.inStock ? 'in-stock' : 'out-of-stock'}`}>
          {product.inStock ? 'In Stock' : 'Out of Stock'}
        </div>
      </div>
    </div>
  );
};