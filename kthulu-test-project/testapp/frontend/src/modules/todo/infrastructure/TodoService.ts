import { Todo, TodoFilter } from '../domain/Todo';
import { api } from '@/services/api';

const BASE_URL = '/todos';

export const TodoService = {
  list: async (filter?: TodoFilter): Promise<Todo[]> => {
    const params = new URLSearchParams();
    if (filter?.query) params.append('q', filter.query);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

    const response = await api.get<Todo[]>(`${BASE_URL}?${params.toString()}`);
    return response.data;
  },

  get: async (id: number): Promise<Todo> => {
    const response = await api.get<Todo>(`${BASE_URL}/${id}`);
    return response.data;
  },

  create: async (data: Omit<Todo, 'id'>): Promise<Todo> => {
    const response = await api.post<Todo>(BASE_URL, data);
    return response.data;
  },

  update: async (id: number, data: Partial<Todo>): Promise<Todo> => {
    const response = await api.put<Todo>(`${BASE_URL}/${id}`, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`${BASE_URL}/${id}`);
  },
};
