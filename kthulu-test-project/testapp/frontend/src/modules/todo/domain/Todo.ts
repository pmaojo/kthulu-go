export interface Todo {
  id: number;
  created_at?: string;
  updated_at?: string;
  title: string;
  completed: boolean;
}

export interface TodoFilter {
  query?: string;
  page?: number;
  pageSize?: number;
}
