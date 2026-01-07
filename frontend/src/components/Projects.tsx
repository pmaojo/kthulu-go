import type { Project } from '@/types/kthulu';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Input } from '@/components/ui/input';
import { Folder, FolderOpen, Clock, Plus, Search, Trash2, ArrowRight, Settings } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { es } from 'date-fns/locale';
import { useProjects } from '@/hooks/useProjects';

interface ProjectsProps {
  onSelectProject?: (project: Project) => void;
  onCreateProject?: () => void;
}

export function Projects({ onSelectProject, onCreateProject }: ProjectsProps) {
  const {
    projects,
    isLoading,
    searchQuery,
    setSearchQuery,
    deleteProject
  } = useProjects();

  const handleDelete = async (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (!confirm('¿Estás seguro de que deseas eliminar este proyecto? Esta acción no se puede deshacer.')) return;
    await deleteProject(id);
  };

  return (
    <div className="h-full flex flex-col bg-kthulu-surface1 p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h2 className="font-mono font-bold text-primary text-2xl flex items-center gap-2">
            <FolderOpen className="w-6 h-6" />
            PROYECTOS
          </h2>
          <p className="text-muted-foreground font-mono text-sm">Gestiona tus proyectos generados con Kthulu</p>
        </div>

        <Button onClick={onCreateProject} className="font-mono gap-2">
          <Plus className="w-4 h-4" />
          Nuevo Proyecto
        </Button>
      </div>

      {/* Filters */}
      <div className="relative max-w-md">
        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Buscar proyectos..."
          className="pl-9 font-mono bg-kthulu-surface2 border-primary/20"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="bg-kthulu-surface2 border-primary/10">
              <CardHeader>
                <Skeleton className="h-6 w-3/4 mb-2" />
                <Skeleton className="h-4 w-1/2" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-20 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {projects.map((project) => (
            <Card
              key={project.id}
              className="bg-kthulu-surface2 border-primary/10 hover:border-primary/50 transition-colors cursor-pointer group flex flex-col"
              onClick={() => onSelectProject?.(project)}
            >
              <CardHeader className="pb-3">
                <div className="flex justify-between items-start">
                  <div className="flex items-center gap-2">
                    <Folder className="w-5 h-5 text-primary" />
                    <CardTitle className="font-mono text-lg text-foreground group-hover:text-primary transition-colors">
                      {project.name}
                    </CardTitle>
                  </div>
                  <Badge variant={project.status === 'active' ? 'default' : 'secondary'} className="font-mono text-xs">
                    {project.status || 'draft'}
                  </Badge>
                </div>
                <CardDescription className="font-mono text-xs truncate">
                  {project.path}
                </CardDescription>
              </CardHeader>

              <CardContent className="flex-1">
                <p className="text-sm text-muted-foreground line-clamp-3 mb-4 font-mono">
                  {project.description || "Sin descripción"}
                </p>

                {project.modules && project.modules.length > 0 && (
                  <div className="flex flex-wrap gap-1 mb-4">
                    {project.modules.slice(0, 3).map((mod) => (
                      <Badge key={mod} variant="outline" className="text-[10px] bg-background/50">
                        {mod}
                      </Badge>
                    ))}
                    {project.modules.length > 3 && (
                      <Badge variant="outline" className="text-[10px] bg-background/50">
                        +{project.modules.length - 3}
                      </Badge>
                    )}
                  </div>
                )}
              </CardContent>

              <CardFooter className="pt-3 border-t border-primary/10 flex justify-between items-center text-xs text-muted-foreground font-mono">
                <div className="flex items-center gap-1" title={project.updatedAt}>
                  <Clock className="w-3 h-3" />
                  {project.updatedAt ? formatDistanceToNow(new Date(project.updatedAt), { addSuffix: true, locale: es }) : 'N/A'}
                </div>

                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                   <Button variant="ghost" size="icon" className="h-7 w-7 hover:bg-destructive/20 hover:text-destructive" onClick={(e) => handleDelete(e, project.id)}>
                      <Trash2 className="w-4 h-4" />
                   </Button>
                   <Button variant="ghost" size="icon" className="h-7 w-7 hover:bg-primary/20 hover:text-primary">
                      <Settings className="w-4 h-4" />
                   </Button>
                   <Button variant="ghost" size="icon" className="h-7 w-7 hover:bg-primary/20 hover:text-primary">
                      <ArrowRight className="w-4 h-4" />
                   </Button>
                </div>
              </CardFooter>
            </Card>
          ))}

          {/* Empty State */}
          {projects.length === 0 && !isLoading && (
             <div className="col-span-full flex flex-col items-center justify-center py-12 text-muted-foreground">
                <FolderOpen className="w-12 h-12 mb-4 opacity-20" />
                <p className="font-mono">No se encontraron proyectos.</p>
                {searchQuery && <Button variant="link" onClick={() => setSearchQuery('')}>Limpiar búsqueda</Button>}
             </div>
          )}
        </div>
      )}
    </div>
  );
}
