import { useState, useEffect, useCallback } from 'react';
import { User } from '../domain/User';
import { UserService } from '../infrastructure/UserService';
import { message } from 'antd';

export const useUsers = () => {
  const [data, setData] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const result = await UserService.list();
      setData(result);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch users');
      message.error('Failed to load users');
    } finally {
      setLoading(false);
    }
  }, []);

  const createUser = async (user: Omit<User, 'id'>) => {
    try {
      await UserService.create(user);
      message.success('User created successfully');
      fetchUsers();
    } catch (err: any) {
      message.error('Failed to create User');
      throw err;
    }
  };

  const updateUser = async (id: number, user: Partial<User>) => {
    try {
      await UserService.update(id, user);
      message.success('User updated successfully');
      fetchUsers();
    } catch (err: any) {
      message.error('Failed to update User');
      throw err;
    }
  };

  const deleteUser = async (id: number) => {
    try {
      await UserService.delete(id);
      message.success('User deleted successfully');
      fetchUsers();
    } catch (err: any) {
      message.error('Failed to delete User');
      throw err;
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  return {
    data,
    loading,
    error,
    refresh: fetchUsers,
    create: createUser,
    update: updateUser,
    remove: deleteUser,
  };
};
