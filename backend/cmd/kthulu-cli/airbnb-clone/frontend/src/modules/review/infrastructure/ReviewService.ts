import { Review, ReviewFilter } from '../domain/Review';
import { api } from '@/services/api';

const BASE_URL = '/reviews';

export const ReviewService = {
  list: async (filter?: ReviewFilter): Promise<Review[]> => {
    const params = new URLSearchParams();
    if (filter?.query) params.append('q', filter.query);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

    const response = await api.get<Review[]>(`+ "`" + `${BASE_URL}?${params.toString()}`+ "`" + `);
    return response.data;
  },

  get: async (id: number): Promise<Review> => {
    const response = await api.get<Review>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
    return response.data;
  },

  create: async (data: Omit<Review, 'id'>): Promise<Review> => {
    const response = await api.post<Review>(BASE_URL, data);
    return response.data;
  },

  update: async (id: number, data: Partial<Review>): Promise<Review> => {
    const response = await api.put<Review>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
  },
};
