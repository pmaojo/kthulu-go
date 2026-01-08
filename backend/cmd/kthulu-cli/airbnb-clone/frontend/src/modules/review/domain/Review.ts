export interface Review {
  id: number;
  created_at?: string;
  updated_at?: string;
  name: string;
}

export interface ReviewFilter {
  query?: string;
  page?: number;
  pageSize?: number;
}
