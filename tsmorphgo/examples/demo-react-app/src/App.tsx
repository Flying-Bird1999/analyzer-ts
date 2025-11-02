import React, { useState, useEffect } from 'react';
import axios from 'axios';
import _ from 'lodash';

interface User {
  id: number;
  name: string;
  email: string;
  active: boolean;
}

interface Msg {
  code: number;
  msg: string;
}

interface ApiResponse<T> {
  data: T;
  status: string;
  msg: Msg
}

const App: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async (): Promise<void> => {
    try {
      setLoading(true);
      const response = await axios.get<ApiResponse<User[]>>('/api/users');
      setUsers(response.data.data);
    } catch (error) {
      console.error('Error fetching users:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatUserName = (user: User): string => {
    return _.toUpper(user.name);
  };

  const filterActiveUsers = (users: User[]): User[] => {
    return _.filter(users, { active: true });
  };

  if (loading) {
    return <div className="loading">Loading...</div>;
  }

  return (
    <div className="app">
      <h1>User Management</h1>
      <div className="user-list">
        {filterActiveUsers(users).map((user) => (
          <div key={user.id} className="user-card">
            <h3>{formatUserName(user)}</h3>
            <p>{user.email}</p>
            <span className={`status ${user.active ? 'active' : 'inactive'}`}>
              {user.active ? 'Active' : 'Inactive'}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default App;