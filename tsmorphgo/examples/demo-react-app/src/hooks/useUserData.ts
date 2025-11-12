import { useState, useEffect } from 'react';

interface User {
  id: number;
  name: string;
  email: string;
  avatar: string;
}

export const useUserData = (userId: number) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Simulate API call
    const fetchUser = async () => {
      try {
        setLoading(true);
        // Mock API delay
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Mock user data
        const userData: User = {
          id: userId,
          name: 'John Doe',
          email: 'john.doe@example.com',
          avatar: 'https://via.placeholder.com/150/000000/FFFFFF?text=JD'
        };

        setUser(userData);
        setError(null);
      } catch (err) {
        setError('Failed to fetch user data');
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [userId]);

  return { user, loading, error };
};