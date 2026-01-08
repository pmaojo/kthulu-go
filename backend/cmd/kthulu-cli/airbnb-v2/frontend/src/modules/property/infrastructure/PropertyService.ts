import { Property, PropertyFilter } from '../domain/Property';
import { api } from '@/services/api';

const BASE_URL = '/properties';

export const PropertyService = {
  list: async (filter?: PropertyFilter): Promise<Property[]> => {
    const params = new URLSearchParams();
    if (filter?.query) params.append('q', filter.query);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

    const response = await api.get<Property[]>(`+ "`" + `${BASE_URL}?${params.toString()}`+ "`" + `);
    return response.data;
  },

  get: async (id: number): Promise<Property> => {
    const response = await api.get<Property>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
    return response.data;
  },

  create: async (data: Omit<Property, 'id'>): Promise<Property> => {
    const response = await api.post<Property>(BASE_URL, data);
    return response.data;
  },

  update: async (id: number, data: Partial<Property>): Promise<Property> => {
    const response = await api.put<Property>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
  },
};
