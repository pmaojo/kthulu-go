import { useState, useEffect, useCallback } from 'react';
import { kthuluApi } from '@/services/kthuluApi';
import type { Project } from '@/types/kthulu';
import { useToast } from '@/hooks/use-toast';

export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const { toast } = useToast();

  const loadProjects = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await kthuluApi.listProjects();
      setProjects(data);
    } catch (error) {
      console.error('Failed to load projects:', error);
      toast({
        title: "Error",
        description: "No se pudieron cargar los proyectos. ¿Está el backend ejecutándose?",
        variant: "destructive",
      });
      // Mock data for demo
      setProjects([
        {
          id: 1,
          name: 'kthulu-demo',
          path: '/projects/kthulu-demo',
          status: 'active',
          description: 'Demo project showing Kthulu capabilities',
          updatedAt: new Date().toISOString(),
          modules: ['auth', 'users']
        },
        {
          id: 2,
          name: 'e-commerce-core',
          path: '/projects/e-commerce-core',
          status: 'draft',
          description: 'Core backend for the new e-commerce platform',
          updatedAt: new Date(Date.now() - 86400000).toISOString(),
          modules: ['products', 'orders', 'inventory']
        }
      ]);
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  const deleteProject = async (id: number) => {
    try {
      await kthuluApi.deleteProject(id);
      toast({
        title: "Proyecto eliminado",
        description: "El proyecto ha sido eliminado correctamente.",
      });
      loadProjects();
    } catch (error) {
      toast({
        title: "Error",
        description: "No se pudo eliminar el proyecto.",
        variant: "destructive",
      });
    }
  };

  const filteredProjects = projects.filter(p =>
    p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    p.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return {
    projects: filteredProjects,
    isLoading,
    searchQuery,
    setSearchQuery,
    loadProjects,
    deleteProject
  };
}
