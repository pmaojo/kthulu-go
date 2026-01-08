import { Booking, BookingFilter } from '../domain/Booking';
import { api } from '@/services/api';

const BASE_URL = '/bookings';

export const BookingService = {
  list: async (filter?: BookingFilter): Promise<Booking[]> => {
    const params = new URLSearchParams();
    if (filter?.query) params.append('q', filter.query);
    if (filter?.page) params.append('page', filter.page.toString());
    if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

    const response = await api.get<Booking[]>(`+ "`" + `${BASE_URL}?${params.toString()}`+ "`" + `);
    return response.data;
  },

  get: async (id: number): Promise<Booking> => {
    const response = await api.get<Booking>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
    return response.data;
  },

  create: async (data: Omit<Booking, 'id'>): Promise<Booking> => {
    const response = await api.post<Booking>(BASE_URL, data);
    return response.data;
  },

  update: async (id: number, data: Partial<Booking>): Promise<Booking> => {
    const response = await api.put<Booking>(`+ "`" + `${BASE_URL}/${id}`+ "`" + `, data);
    return response.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`+ "`" + `${BASE_URL}/${id}`+ "`" + `);
  },
};
