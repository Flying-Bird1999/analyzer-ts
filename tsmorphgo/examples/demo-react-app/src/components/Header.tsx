import React from 'react';
import { Product } from '@/components/App';

interface User {
  id: number;
  name: string;
  email: string;
  avatar: string;
}

interface HeaderProps {
  user: User | null;
  currentDate: string;
  products?: Product[];
}

export const Header: React.FC<HeaderProps> = ({ user, currentDate }) => {
  return (
    <header className="header">
      <div className="header-container">
        <div className="logo">
          <h1>TSMorphGo Demo</h1>
          <p>React App Analysis</p>
        </div>

        <div className="user-info">
          {user ? (
            <div className="user-details">
              <img
                src={user.avatar}
                alt={user.name}
                className="avatar"
              />
              <div className="user-text">
                <span className="user-name">{user.name}</span>
                <span className="user-email">{user.email}</span>
              </div>
            </div>
          ) : (
            <div className="guest-user">
              <span>Guest User</span>
            </div>
          )}
        </div>

        <div className="date-time">
          <span>{currentDate}</span>
        </div>
      </div>
    </header>
  );
};