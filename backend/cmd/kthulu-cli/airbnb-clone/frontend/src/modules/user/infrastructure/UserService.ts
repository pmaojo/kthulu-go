import { User, UserFilter } from '../domain/User';
import { api } from '@/services/api';

const BASE_URL = '/users';

export const UserService = {
  list: async (filter?: UserFilter): Promise<User[]> => {
    const params = new URLSearchParams();
    if (filter?.query) params.append('q', filter.query);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

    const response = await api.get<User[]>(`+ "`" + `${BASE_URL}?${params.toString()}`+ "`" + `);
    return response.data;
  },

  get: async (id: number): Promise<User> => {
    const response = await api.get<User>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
    return response.data;
  },

  create: async (data: Omit<User, 'id'>): Promise<User> => {
    const response = await api.post<User>(BASE_URL, data);
    return response.data;
  },

  update: async (id: number, data: Partial<User>): Promise<User> => {
    const response = await api.put<User>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
  },
};
