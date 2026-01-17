// @kthulu:frontend:hook
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Todo } from '../domain/Todo';
import { TodoService } from '../infrastructure/TodoService';

export const useTodos = () => {
  const queryClient = useQueryClient();

  // Retrieve Layer (Read)
  const { data, isLoading, error } = useQuery({
    queryKey: ['todos'],
    queryFn: TodoService.list,
  });

  // Mutations (Write)
  const createMutation = useMutation({
    mutationFn: (todo: Omit<Todo, 'id'>) => TodoService.create(todo),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Todo> }) => 
      TodoService.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => TodoService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });

  return {
    data: data || [],
    loading: isLoading,
    error,
    create: createMutation.mutateAsync,
    update: (id: number, data: Partial<Todo>) => updateMutation.mutateAsync({ id, data }),
    remove: deleteMutation.mutateAsync,
  };
};
