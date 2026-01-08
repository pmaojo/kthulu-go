export interface Booking {
  id: number;
  created_at?: string;
  updated_at?: string;
  name: string;
}

export interface BookingFilter {
  query?: string;
  page?: number;
  pageSize?: number;
}
