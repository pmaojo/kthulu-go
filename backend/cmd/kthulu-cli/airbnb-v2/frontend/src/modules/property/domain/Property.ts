export interface Property {
  id: number;
  created_at?: string;
  updated_at?: string;
  name: string;
}

export interface PropertyFilter {
  query?: string;
  page?: number;
  pageSize?: number;
}
