import React from 'react';

interface User {
  id: number;
  name: string;
  email: string;
  avatar: string;
}

interface UserProfileProps {
  user: User;
}

export const UserProfile: React.FC<UserProfileProps> = ({ user }) => {
  return (
    <div className="user-profile">
      <div className="profile-header">
        <img
          src={user.avatar}
          alt={user.name}
          className="profile-avatar"
        />
        <div className="profile-info">
          <h2 className="profile-name">{user.name}</h2>
          <p className="profile-email">{user.email}</p>
          <span className="profile-id">ID: {user.id}</span>
        </div>
      </div>

      <div className="profile-stats">
        <div className="stat-item">
          <span className="stat-label">Status</span>
          <span className="stat-value active">Active</span>
        </div>
        <div className="stat-item">
          <span className="stat-label">Role</span>
          <span className="stat-value">Admin</span>
        </div>
        <div className="stat-item">
          <span className="stat-label">Join Date</span>
          <span className="stat-value">2024-01-15</span>
        </div>
      </div>

      <div className="profile-actions">
        <button className="btn btn-primary">Edit Profile</button>
        <button className="btn btn-secondary">Settings</button>
      </div>
    </div>
  );
};