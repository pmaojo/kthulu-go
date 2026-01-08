import { useState, useEffect, useCallback } from 'react';
import { Property } from '../domain/Property';
import { PropertyService } from '../infrastructure/PropertyService';
import { message } from 'antd';

export const usePropertys = () => {
  const [data, setData] = useState<Property[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchPropertys = useCallback(async () => {
    setLoading(true);
    try {
      const result = await PropertyService.list();
      setData(result);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch properties');
      message.error('Failed to load properties');
    } finally {
      setLoading(false);
    }
  }, []);

  const createProperty = async (property: Omit<Property, 'id'>) => {
    try {
      await PropertyService.create(property);
      message.success('Property created successfully');
      fetchPropertys();
    } catch (err: any) {
      message.error('Failed to create Property');
      throw err;
    }
  };

  const updateProperty = async (id: number, property: Partial<Property>) => {
    try {
      await PropertyService.update(id, property);
      message.success('Property updated successfully');
      fetchPropertys();
    } catch (err: any) {
      message.error('Failed to update Property');
      throw err;
    }
  };

  const deleteProperty = async (id: number) => {
    try {
      await PropertyService.delete(id);
      message.success('Property deleted successfully');
      fetchPropertys();
    } catch (err: any) {
      message.error('Failed to delete Property');
      throw err;
    }
  };

  useEffect(() => {
    fetchPropertys();
  }, [fetchPropertys]);

  return {
    data,
    loading,
    error,
    refresh: fetchPropertys,
    create: createProperty,
    update: updateProperty,
    remove: deleteProperty,
  };
};
