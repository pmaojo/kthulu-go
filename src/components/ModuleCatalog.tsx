import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Layers, RefreshCw, Search, ShieldCheck, Workflow, Network } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { ModuleInfo, ModuleInjectionPlan, ModuleValidationResult } from '@/types/kthulu';

const sanitizeModuleList = (value: string) =>
  value
    .split(/[\n,]/)
    .map((module) => module.trim())
    .filter(Boolean);

export function ModuleCatalog() {
  const { toast } = useToast();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<string>('all');
  const [selectedModule, setSelectedModule] = useState<ModuleInfo | null>(null);
  const [validationInput, setValidationInput] = useState('');
  const [validationResult, setValidationResult] = useState<ModuleValidationResult | null>(null);
  const [planResult, setPlanResult] = useState<ModuleInjectionPlan | null>(null);

  const { data: modules = [], isLoading, refetch, isError } = useQuery({
    queryKey: ['modules', category],
    queryFn: async () => {
      try {
        const result = await kthuluApi.listModules(category === 'all' ? undefined : category);
        if (selectedModule) {
          const refreshed = result.find((module) => module.name === selectedModule.name);
          setSelectedModule(refreshed ?? null);
        }
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
    try {
      const detail = await kthuluApi.getModule(module.name);
      setSelectedModule(detail);
    } catch (error) {
      console.error('Failed to load module detail', error);
      toast({
        title: 'No se pudo cargar el módulo',
        description: 'Mostrando información del catálogo.',
      });
    }
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

  return (
    <div className="h-full bg-kthulu-surface1 text-foreground grid grid-cols-1 lg:grid-cols-[320px_1fr]">
      <div className="border-r border-primary/20 bg-kthulu-surface2 flex flex-col">
        <div className="p-4 border-b border-primary/20 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="font-mono font-bold text-primary text-sm flex items-center gap-2">
              <Layers className="w-4 h-4" />
              CATÁLOGO DE MÓDULOS
            </h2>
            <Button variant="ghost" size="icon" onClick={() => refetch()} disabled={isLoading}>
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
          <div className="relative">
            <Search className="w-4 h-4 text-muted-foreground absolute left-3 top-3" />
            <Input
              placeholder="Buscar módulo..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="pl-9 bg-kthulu-surface1 border-primary/30 font-mono text-sm"
            />
          </div>
          <div className="flex flex-wrap gap-2">
            <Badge
              variant={category === 'all' ? 'default' : 'outline'}
              className="cursor-pointer font-mono text-[11px]"
              onClick={() => setCategory('all')}
            >
              Todos
            </Badge>
            {categories.map((item) => (
              <Badge
                key={item}
                variant={category === item ? 'default' : 'outline'}
                className="cursor-pointer font-mono text-[11px]"
                onClick={() => setCategory(item)}
              >
                {item}
              </Badge>
            ))}
          </div>
        </div>
        <ScrollArea className="flex-1">
          <div className="p-4 space-y-2">
            {isError && (
              <div className="text-xs text-destructive font-mono">
                No se pudo cargar la información desde el API.
              </div>
            )}
            {!isLoading && filteredModules.length === 0 && (
              <div className="text-xs text-muted-foreground font-mono">
                No se encontraron módulos que coincidan con la búsqueda.
              </div>
            )}
            {filteredModules.map((module) => (
              <Card
                key={module.name}
                className={`bg-kthulu-surface1 border ${
                  selectedModule?.name === module.name ? 'border-primary/60' : 'border-primary/10'
                } hover:border-primary/60 transition-colors cursor-pointer`}
                onClick={() => handleModuleClick(module)}
              >
                <CardHeader className="pb-2">
                  <CardTitle className="font-mono text-sm text-primary flex items-center justify-between">
                    <span>{module.name}</span>
                    {module.version && (
                      <Badge variant="outline" className="text-[10px] font-mono">
                        v{module.version}
                      </Badge>
                    )}
                  </CardTitle>
                </CardHeader>
                {module.description && (
                  <CardContent className="pt-0 text-xs text-muted-foreground font-mono">
                    {module.description}
                  </CardContent>
                )}
              </Card>
            ))}
          </div>
        </ScrollArea>
      </div>

      <div className="flex flex-col gap-4 p-6 overflow-y-auto">
        <Card className="bg-kthulu-surface2 border-primary/20">
          <CardHeader>
            <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
              <Network className="w-4 h-4" />
              Dependencias y Validación
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid md:grid-cols-[1fr_auto_auto] gap-3">
              <Input
                placeholder="Auth, User, Payment"
                value={validationInput}
                onChange={(event) => setValidationInput(event.target.value)}
                className="bg-kthulu-surface1 border-primary/30 font-mono"
              />
              <Button variant="outline" className="bg-kthulu-surface1 border-secondary/40 font-mono" onClick={handleValidateModules}>
                <ShieldCheck className="w-4 h-4 mr-2" />
                Validar
              </Button>
              <Button className="bg-primary text-background font-mono" onClick={handlePlanModules}>
                <Workflow className="w-4 h-4 mr-2" />
                Planear
              </Button>
            </div>

            {validationResult && (
              <div className="space-y-2 text-xs font-mono">
                <div className="flex items-center gap-2">
                  <Badge variant={validationResult.valid ? 'default' : 'destructive'}>
                    {validationResult.valid ? 'CONJUNTO VÁLIDO' : 'PROBLEMAS DETECTADOS'}
                  </Badge>
                  {validationResult.warnings && validationResult.warnings.length > 0 && (
                    <Badge variant="secondary">{validationResult.warnings.length} warning(s)</Badge>
                  )}
                </div>
                {validationResult.missing?.length ? (
                  <p className="text-destructive">
                    Faltan dependencias: {validationResult.missing.join(', ')}
                  </p>
                ) : null}
                {validationResult.conflicts?.length ? (
                  <div className="text-destructive space-y-1">
                    <p>Conflictos detectados:</p>
                    <ul className="list-disc list-inside space-y-1">
                      {validationResult.conflicts.map((conflict) => (
                        <li key={conflict.module}>
                          {conflict.module}: {conflict.conflicts.join(', ')} — {conflict.reason}
                        </li>
                      ))}
                    </ul>
                  </div>
                ) : null}
                {validationResult.circular?.length ? (
                  <div className="text-accent space-y-1">
                    <p>Cadenas circulares:</p>
                    <ul className="list-disc list-inside space-y-1">
                      {validationResult.circular.map((cycle, index) => (
                        <li key={index}>{cycle.chain.join(' → ')}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}
              </div>
            )}

            {planResult && (
              <div className="space-y-2 text-xs font-mono">
                <div className="flex items-center gap-2">
                  <Badge variant="outline">Orden de ejecución</Badge>
                  <span className="text-muted-foreground">
                    {planResult.execution_order.length} pasos
                  </span>
                </div>
                <div className="flex flex-wrap gap-2">
                  {planResult.execution_order.map((name) => (
                    <Badge key={name} variant="secondary">
                      {name}
                    </Badge>
                  ))}
                </div>
                {planResult.warnings && planResult.warnings.length > 0 ? (
                  <div className="text-accent space-y-1">
                    <p>Warnings:</p>
                    <ul className="list-disc list-inside space-y-1">
                      {planResult.warnings.map((warning, index) => (
                        <li key={index}>{warning}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}
                {planResult.errors && planResult.errors.length > 0 ? (
                  <div className="text-destructive space-y-1">
                    <p>Errores:</p>
                    <ul className="list-disc list-inside space-y-1">
                      {planResult.errors.map((error, index) => (
                        <li key={index}>{error}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="bg-kthulu-surface2 border-primary/20 flex-1">
          <CardHeader>
            <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
              <Layers className="w-4 h-4" />
              {selectedModule ? `Detalle de ${selectedModule.name}` : 'Selecciona un módulo'}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-sm font-mono">
            {!selectedModule && (
              <p className="text-muted-foreground text-xs">
                Elige un módulo del catálogo para inspeccionar su metadata y dependencias.
              </p>
            )}

            {selectedModule && (
              <div className="space-y-4">
                {selectedModule.description && (
                  <p className="text-muted-foreground text-xs leading-5">
                    {selectedModule.description}
                  </p>
                )}

                <div className="flex flex-wrap gap-2">
                  {selectedModule.tags?.map((tag) => (
                    <Badge key={tag} variant="outline" className="text-[11px]">
                      #{tag}
                    </Badge>
                  ))}
                </div>

                <div className="grid md:grid-cols-2 gap-4">
                  <div>
                    <h3 className="text-primary text-xs font-semibold mb-2">Dependencias</h3>
                    {selectedModule.dependencies?.length ? (
                      <ul className="space-y-1 text-xs">
                        {selectedModule.dependencies.map((dependency) => (
                          <li key={dependency}>{dependency}</li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-muted-foreground text-xs">Sin dependencias declaradas.</p>
                    )}
                  </div>
                  <div>
                    <h3 className="text-primary text-xs font-semibold mb-2">Conflictos</h3>
                    {selectedModule.conflicts?.length ? (
                      <ul className="space-y-1 text-xs">
                        {selectedModule.conflicts.map((conflict) => (
                          <li key={conflict}>{conflict}</li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-muted-foreground text-xs">Sin conflictos registrados.</p>
                    )}
                  </div>
                </div>

                {selectedModule.entities && selectedModule.entities.length > 0 && (
                  <div>
                    <h3 className="text-primary text-xs font-semibold mb-2">Entidades</h3>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="text-xs">Nombre</TableHead>
                          <TableHead className="text-xs">Tipo</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {selectedModule.entities.map((entity: any, index) => (
                          <TableRow key={entity.name ?? index}>
                            <TableCell className="text-xs">{entity.name ?? '—'}</TableCell>
                            <TableCell className="text-xs text-muted-foreground">
                              {entity.type ?? '—'}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}

                {selectedModule.routes && selectedModule.routes.length > 0 && (
                  <div>
                    <h3 className="text-primary text-xs font-semibold mb-2">Endpoints</h3>
                    <ul className="space-y-1 text-xs">
                      {selectedModule.routes.map((route: any, index) => (
                        <li key={route.path ?? index}>
                          <Badge variant="outline" className="mr-2 text-[10px]">
                            {(route.method ?? 'ANY').toUpperCase()}
                          </Badge>
                          {route.path ?? '/'}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {selectedModule.migrations && selectedModule.migrations.length > 0 && (
                  <div>
                    <h3 className="text-primary text-xs font-semibold mb-2">Migraciones</h3>
                    <ul className="space-y-1 text-xs">
                      {selectedModule.migrations.map((migration) => (
                        <li key={migration}>{migration}</li>
                      ))}
                    </ul>
                  </div>
                )}

                {selectedModule.config && (
                  <div className="space-y-2">
                    <h3 className="text-primary text-xs font-semibold">Configuración</h3>
                    <pre className="bg-kthulu-surface1 border border-primary/20 rounded-sm p-3 text-[11px] overflow-x-auto">
{JSON.stringify(selectedModule.config, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
