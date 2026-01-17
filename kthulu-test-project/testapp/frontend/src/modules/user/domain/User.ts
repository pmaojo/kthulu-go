export interface User {
  id: number;
  created_at?: string;
  updated_at?: string;
  name: string;
}

export interface UserFilter {
  query?: string;
  page?: number;
  pageSize?: number;
}
