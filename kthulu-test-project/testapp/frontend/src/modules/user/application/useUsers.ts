// @kthulu:frontend:hook
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { User } from '../domain/User';
import { UserService } from '../infrastructure/UserService';

export const useUsers = () => {
  const queryClient = useQueryClient();

  // Retrieve Layer (Read)
  const { data, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: UserService.list,
  });

  // Mutations (Write)
  const createMutation = useMutation({
    mutationFn: (user: Omit<User, 'id'>) => UserService.create(user),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<User> }) => 
      UserService.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => UserService.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });

  return {
    data: data || [],
    loading: isLoading,
    error,
    create: createMutation.mutateAsync,
    update: (id: number, data: Partial<User>) => updateMutation.mutateAsync({ id, data }),
    remove: deleteMutation.mutateAsync,
  };
};
