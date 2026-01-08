import { useMemo, useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Layers, RefreshCw, Search, ShieldCheck, Workflow, Network, Package, ChevronRight, ArrowLeft } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { ModuleInfo, ModuleInjectionPlan, ModuleValidationResult } from '@/types/kthulu';
import { Skeleton } from '@/components/ui/skeleton';

const sanitizeModuleList = (value: string) =>
  value
    .split(/[\n,]/)
    .map((module) => module.trim())
    .filter(Boolean);

interface ModuleCatalogProps {
  initialModuleId?: string;
}

export function ModuleCatalog({ initialModuleId }: ModuleCatalogProps) {
  const { toast } = useToast();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<string>('all');
  const [selectedModule, setSelectedModule] = useState<ModuleInfo | null>(null);
  const [validationInput, setValidationInput] = useState('');
  const [validationResult, setValidationResult] = useState<ModuleValidationResult | null>(null);
  const [planResult, setPlanResult] = useState<ModuleInjectionPlan | null>(null);
  const [view, setView] = useState<'grid' | 'detail'>('grid');

  const { data: modules = [], isLoading, refetch, isError } = useQuery({
    queryKey: ['modules', category],
    queryFn: async () => {
      try {
        const result = await kthuluApi.listModules(category === 'all' ? undefined : category);
        return result;
      } catch (error) {
        console.error('Failed to fetch modules', error);
        toast({
          title: 'Error al cargar módulos',
          description: 'No fue posible conectar con el catálogo de módulos.',
          variant: 'destructive',
        });
        throw error;
      }
    },
  });

  // Handle initialModuleId
  useEffect(() => {
    if (initialModuleId && modules.length > 0) {
      const found = modules.find(m => m.name === initialModuleId || m.name.toLowerCase() === initialModuleId.toLowerCase());
      if (found) {
        handleModuleClick(found);
      } else {
        // If not found in list, maybe try fetching directly?
        // But for now, just show a toast or stay in grid
      }
    }
  }, [initialModuleId, modules]);

  const filteredModules = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) {
      return modules;
    }

    return modules.filter((module) =>
      module.name.toLowerCase().includes(term) ||
      (module.description && module.description.toLowerCase().includes(term)) ||
      (module.tags && module.tags.some((tag) => tag.toLowerCase().includes(term)))
    );
  }, [modules, search]);

  const categories = useMemo(() => {
    const names = new Set<string>();
    modules.forEach((module) => {
      if (module.category) {
        names.add(module.category);
      }
    });
    return Array.from(names.values()).sort();
  }, [modules]);

  const handleModuleClick = async (module: ModuleInfo) => {
    setSelectedModule(module);
    setView('detail');
    try {
      const detail = await kthuluApi.getModule(module.name);
      setSelectedModule(detail);
    } catch (error) {
      console.error('Failed to load module detail', error);
      toast({
        title: 'No se pudo cargar el detalle completo',
        description: 'Mostrando información básica del catálogo.',
      });
    }
  };

  const handleBackToGrid = () => {
    setSelectedModule(null);
    setView('grid');
    // Optionally clear URL param via navigation if using React Router hooks deeper here,
    // but the parent handles navigation. For now, local state is enough.
  };

  const handleValidateModules = async () => {
    const modulesToValidate = sanitizeModuleList(validationInput);
    if (modulesToValidate.length === 0) {
      toast({
        title: 'Agrega módulos a validar',
        description: 'Ingresa módulos separados por coma o nueva línea.',
      });
      return;
    }

    try {
      const result = await kthuluApi.validateModules(modulesToValidate);
      setValidationResult(result);
      toast({
        title: result.valid ? 'Conjunto válido' : 'Validación con problemas',
        description: result.valid
          ? 'No se detectaron conflictos en la selección.'
          : 'Revisa los detalles de la validación para corregir el conjunto.',
        variant: result.valid ? 'default' : 'destructive',
      });
    } catch (error) {
      console.error('Failed to validate modules', error);
      toast({
        title: 'Error validando módulos',
        description: 'No fue posible ejecutar la validación con el API.',
        variant: 'destructive',
      });
    }
  };

  const handlePlanModules = async () => {
    const modulesToPlan = sanitizeModuleList(validationInput);
    if (modulesToPlan.length === 0) {
      toast({
        title: 'Agrega módulos a planificar',
        description: 'Ingresa módulos separados por coma o nueva línea.',
      });
      return;
    }

    try {
      const result = await kthuluApi.planModules(modulesToPlan);
      setPlanResult(result);
      toast({
        title: 'Plan de inyección generado',
        description: `Se inyectarán ${result.injected_modules.length} módulos en ${result.execution_order.length} pasos.`,
      });
    } catch (error) {
      console.error('Failed to plan modules', error);
      toast({
        title: 'Error generando plan',
        description: 'No fue posible construir el plan de inyección.',
        variant: 'destructive',
      });
    }
  };

  if (view === 'detail' && selectedModule) {
    return (
      <div className="h-full flex flex-col bg-kthulu-surface1 p-6 space-y-6">
        <div className="flex items-center gap-4">
           <Button variant="ghost" size="icon" onClick={handleBackToGrid}>
              <ArrowLeft className="w-5 h-5" />
           </Button>
           <div>
              <h2 className="font-mono font-bold text-primary text-2xl flex items-center gap-2">
                 <Package className="w-6 h-6" />
                 {selectedModule.name}
              </h2>
              <p className="text-muted-foreground font-mono text-sm">Detalle del módulo y dependencias</p>
           </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 flex-1 overflow-hidden">
             {/* Main Detail */}
             <div className="md:col-span-2 flex flex-col gap-6 overflow-y-auto pr-2">
                 <Card className="bg-kthulu-surface2 border-primary/20">
                     <CardHeader>
                         <CardTitle className="font-mono text-lg text-primary">Descripción</CardTitle>
                     </CardHeader>
                     <CardContent className="font-mono text-sm text-muted-foreground">
                         {selectedModule.description || "Sin descripción disponible."}
                     </CardContent>
                     <CardFooter className="gap-2">
                        {selectedModule.tags?.map((tag) => (
                           <Badge key={tag} variant="outline" className="font-mono">#{tag}</Badge>
                        ))}
                     </CardFooter>
                 </Card>

                 {selectedModule.entities && selectedModule.entities.length > 0 && (
                    <Card className="bg-kthulu-surface2 border-primary/20">
                      <CardHeader>
                        <CardTitle className="font-mono text-lg text-primary">Entidades del Dominio</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead className="font-mono text-xs">Nombre</TableHead>
                              <TableHead className="font-mono text-xs">Tipo</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {selectedModule.entities.map((entity: any, index) => (
                              <TableRow key={entity.name ?? index}>
                                <TableCell className="font-mono text-xs">{entity.name ?? '—'}</TableCell>
                                <TableCell className="font-mono text-xs text-muted-foreground">
                                  {entity.type ?? '—'}
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </CardContent>
                    </Card>
                 )}

                 {selectedModule.routes && selectedModule.routes.length > 0 && (
                  <Card className="bg-kthulu-surface2 border-primary/20">
                    <CardHeader>
                      <CardTitle className="font-mono text-lg text-primary">Endpoints API</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <ul className="space-y-2 font-mono text-xs">
                        {selectedModule.routes.map((route: any, index) => (
                            <li key={route.path ?? index} className="flex items-center bg-kthulu-surface1 p-2 rounded border border-primary/10">
                            <Badge variant="outline" className="mr-3 w-16 justify-center">
                                {(route.method ?? 'ANY').toUpperCase()}
                            </Badge>
                            <span className="text-primary">{route.path ?? '/'}</span>
                            </li>
                        ))}
                        </ul>
                    </CardContent>
                  </Card>
                )}
             </div>

             {/* Sidebar Info */}
             <div className="flex flex-col gap-6">
                <Card className="bg-kthulu-surface2 border-primary/20">
                   <CardHeader>
                      <CardTitle className="font-mono text-sm text-primary">Metadata</CardTitle>
                   </CardHeader>
                   <CardContent className="space-y-3 font-mono text-xs">
                       <div className="flex justify-between">
                           <span className="text-muted-foreground">Versión</span>
                           <span className="text-foreground">{selectedModule.version || 'latest'}</span>
                       </div>
                       <div className="flex justify-between">
                           <span className="text-muted-foreground">Categoría</span>
                           <span className="text-foreground">{selectedModule.category || 'General'}</span>
                       </div>
                       {selectedModule.author && (
                           <div className="flex justify-between">
                               <span className="text-muted-foreground">Autor</span>
                               <span className="text-foreground">{selectedModule.author}</span>
                           </div>
                       )}
                   </CardContent>
                </Card>

                <Card className="bg-kthulu-surface2 border-primary/20">
                   <CardHeader>
                      <CardTitle className="font-mono text-sm text-primary">Dependencias</CardTitle>
                   </CardHeader>
                   <CardContent className="space-y-2">
                       {selectedModule.dependencies?.length ? (
                           selectedModule.dependencies.map(dep => (
                               <Badge key={dep} variant="secondary" className="font-mono mr-1 mb-1">{dep}</Badge>
                           ))
                       ) : (
                           <p className="font-mono text-xs text-muted-foreground">Sin dependencias.</p>
                       )}
                   </CardContent>
                </Card>

                 {selectedModule.config && (
                  <Card className="bg-kthulu-surface2 border-primary/20 flex-1">
                    <CardHeader>
                       <CardTitle className="font-mono text-sm text-primary">Configuración Default</CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1">
                      <pre className="bg-kthulu-surface1 border border-primary/20 rounded-sm p-3 text-[10px] overflow-auto max-h-60 font-mono">
                         {JSON.stringify(selectedModule.config, null, 2)}
                      </pre>
                    </CardContent>
                  </Card>
                )}
             </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col bg-kthulu-surface1 p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h2 className="font-mono font-bold text-primary text-2xl flex items-center gap-2">
            <Layers className="w-6 h-6" />
            CATÁLOGO DE MÓDULOS
          </h2>
          <p className="text-muted-foreground font-mono text-sm">Explora, valida y planifica la integración de módulos</p>
        </div>

        <Button onClick={() => refetch()} variant="outline" className="font-mono gap-2">
          <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          Actualizar
        </Button>
      </div>

      {/* Grid Layout Container */}
      <div className="flex flex-col lg:flex-row gap-6 flex-1 overflow-hidden">
         {/* Main Grid */}
         <div className="flex-1 flex flex-col gap-4 overflow-hidden">
             {/* Search & Filters */}
             <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                    placeholder="Buscar módulo..."
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    className="pl-9 bg-kthulu-surface2 border-primary/20 font-mono"
                    />
                </div>
                <div className="flex gap-2 overflow-x-auto pb-2 sm:pb-0">
                    <Badge
                        variant={category === 'all' ? 'default' : 'outline'}
                        className="cursor-pointer font-mono h-9 px-4 flex items-center hover:bg-primary/20"
                        onClick={() => setCategory('all')}
                        >
                        Todos
                    </Badge>
                    {categories.map((item) => (
                        <Badge
                            key={item}
                            variant={category === item ? 'default' : 'outline'}
                            className="cursor-pointer font-mono h-9 px-4 flex items-center hover:bg-primary/20"
                            onClick={() => setCategory(item)}
                        >
                            {item}
                        </Badge>
                    ))}
                </div>
             </div>

             {/* Cards Grid */}
             <ScrollArea className="flex-1 pr-4">
                 {isLoading ? (
                     <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                         {[1,2,3,4,5,6].map(i => (
                             <Skeleton key={i} className="h-40 w-full bg-kthulu-surface2" />
                         ))}
                     </div>
                 ) : (
                     <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pb-4">
                        {filteredModules.map((module) => (
                        <Card
                            key={module.name}
                            className="bg-kthulu-surface2 border-primary/10 hover:border-primary/50 transition-all cursor-pointer group flex flex-col h-full"
                            onClick={() => handleModuleClick(module)}
                        >
                            <CardHeader className="pb-2">
                            <div className="flex justify-between items-start">
                                <CardTitle className="font-mono text-lg text-foreground group-hover:text-primary transition-colors flex items-center gap-2">
                                    <Package className="w-4 h-4" />
                                    {module.name}
                                </CardTitle>
                                {module.version && (
                                    <Badge variant="secondary" className="text-[10px] font-mono">
                                        v{module.version}
                                    </Badge>
                                )}
                            </div>
                            </CardHeader>
                            <CardContent className="flex-1">
                                <CardDescription className="font-mono text-xs line-clamp-3">
                                    {module.description}
                                </CardDescription>
                            </CardContent>
                            <CardFooter className="pt-2 border-t border-primary/5 flex justify-between items-center">
                                <div className="flex gap-2">
                                    {module.tags?.slice(0, 2).map(tag => (
                                        <Badge key={tag} variant="outline" className="text-[10px] font-mono border-primary/20">{tag}</Badge>
                                    ))}
                                </div>
                                <ChevronRight className="w-4 h-4 text-primary opacity-0 group-hover:opacity-100 transition-opacity" />
                            </CardFooter>
                        </Card>
                        ))}
                        {!isLoading && filteredModules.length === 0 && (
                            <div className="col-span-full py-12 text-center text-muted-foreground font-mono">
                                No se encontraron módulos.
                            </div>
                        )}
                     </div>
                 )}
             </ScrollArea>
         </div>

         {/* Validation Panel (Right Side) */}
         <div className="w-full lg:w-80 flex flex-col gap-4">
             <Card className="bg-kthulu-surface2 border-primary/20 h-full flex flex-col">
                <CardHeader>
                    <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                    <Network className="w-4 h-4" />
                    Workbench
                    </CardTitle>
                    <CardDescription className="text-xs font-mono">Validador de dependencias</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4 flex-1">
                    <div className="space-y-2">
                        <label className="text-xs font-mono text-muted-foreground">Módulos (separados por coma)</label>
                        <Input
                            placeholder="Auth, User, Payment"
                            value={validationInput}
                            onChange={(event) => setValidationInput(event.target.value)}
                            className="bg-kthulu-surface1 border-primary/30 font-mono text-sm"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-2">
                        <Button variant="outline" size="sm" className="bg-kthulu-surface1 border-secondary/40 font-mono" onClick={handleValidateModules}>
                            <ShieldCheck className="w-3 h-3 mr-2" />
                            Validar
                        </Button>
                        <Button size="sm" className="bg-primary text-background font-mono" onClick={handlePlanModules}>
                            <Workflow className="w-3 h-3 mr-2" />
                            Planear
                        </Button>
                    </div>

                    <ScrollArea className="h-48 rounded border border-primary/10 bg-kthulu-surface1 p-2">
                        {!validationResult && !planResult && (
                            <p className="text-xs text-muted-foreground font-mono text-center pt-4">Resultados aparecerán aquí.</p>
                        )}

                        {validationResult && (
                        <div className="space-y-2 text-xs font-mono mb-4">
                            <div className="flex items-center gap-2">
                            <Badge variant={validationResult.valid ? 'default' : 'destructive'} className="text-[10px]">
                                {validationResult.valid ? 'VÁLIDO' : 'INVALIDO'}
                            </Badge>
                            </div>
                            {validationResult.missing?.length ? (
                            <p className="text-destructive">
                                Faltan: {validationResult.missing.join(', ')}
                            </p>
                            ) : null}
                            {validationResult.conflicts?.length ? (
                                <p className="text-destructive">Conflictos detectados.</p>
                            ) : null}
                        </div>
                        )}

                        {planResult && (
                        <div className="space-y-2 text-xs font-mono">
                            <div className="flex items-center gap-2">
                            <Badge variant="outline" className="text-[10px]">Plan generado</Badge>
                            </div>
                            <div className="flex flex-wrap gap-1">
                            {planResult.execution_order.map((name, i) => (
                                <div key={name} className="flex items-center">
                                    <span className="text-primary">{name}</span>
                                    {i < planResult.execution_order.length - 1 && <span className="mx-1 text-muted-foreground">→</span>}
                                </div>
                            ))}
                            </div>
                        </div>
                        )}
                    </ScrollArea>
                </CardContent>
            </Card>
         </div>
      </div>
    </div>
  );
}
