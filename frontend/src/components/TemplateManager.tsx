import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Box,
  CheckCircle2,
  CloudDownload,
  DatabaseZap,
  FileJson,
  Globe,
  Loader2,
  RefreshCcw,
  Search,
  ServerCog,
  Settings2,
  ShieldAlert,
  ShieldCheck,
  UploadCloud,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type {
  TemplateInfo,
  TemplateRenderResult,
  TemplateSyncResult,
  TemplateDriftReport,
} from '@/types/kthulu';

interface TaskLogEntry {
  title: string;
  timestamp: string;
  success: boolean;
  payload?: unknown;
  response?: unknown;
  error?: string;
}

const parseJsonInput = (value: string): Record<string, any> | undefined => {
  if (!value.trim()) {
    return undefined;
  }

  try {
    return JSON.parse(value);
  } catch (error) {
    throw new Error('JSON inválido en variables de render.');
  }
};

export function TemplateManager() {
  const { toast } = useToast();
  const [search, setSearch] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState<TemplateInfo | null>(null);
  const [renderVariables, setRenderVariables] = useState('');
  const [cacheUrl, setCacheUrl] = useState('');
  const [registryName, setRegistryName] = useState('');
  const [registryUrl, setRegistryUrl] = useState('');
  const [cleanMaxAge, setCleanMaxAge] = useState('');
  const [syncSource, setSyncSource] = useState('');
  const [isRunning, setIsRunning] = useState(false);
  const [taskLog, setTaskLog] = useState<TaskLogEntry[]>([]);
  const [renderPreview, setRenderPreview] = useState<TemplateRenderResult | null>(null);

  const { data: templates = [], isLoading, refetch } = useQuery({
    queryKey: ['templates'],
    queryFn: async () => {
      try {
        return await kthuluApi.listTemplates();
      } catch (error) {
        console.error('Failed to load templates', error);
        toast({
          title: 'No se pudieron cargar los templates',
          description: 'Verifica la conexión con el backend de Kthulu.',
          variant: 'destructive',
        });
        throw error;
      }
    },
  });

  const filteredTemplates = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) return templates;

    return templates.filter((template) =>
      template.name.toLowerCase().includes(term) ||
      (template.description && template.description.toLowerCase().includes(term)) ||
      (template.tags && template.tags.some((tag) => tag.toLowerCase().includes(term)))
    );
  }, [search, templates]);

  const logTask = (entry: Omit<TaskLogEntry, 'timestamp'>) => {
    setTaskLog((previous) => [
      { ...entry, timestamp: new Date().toISOString() },
      ...previous,
    ]);
  };

  const handleTask = async <T,>(options: {
    title: string;
    payload?: Record<string, any>;
    executor: () => Promise<T>;
    onSuccess?: (response: T) => void;
  }) => {
    try {
      setIsRunning(true);
      const response = await options.executor();
      options.onSuccess?.(response);
      logTask({
        title: options.title,
        success: true,
        payload: options.payload,
        response,
      });
      toast({
        title: options.title,
        description: 'Operación completada correctamente.',
      });
    } catch (error) {
      console.error(options.title, error);
      const message = error instanceof Error ? error.message : 'Error desconocido';
      logTask({
        title: options.title,
        success: false,
        payload: options.payload,
        error: message,
      });
      toast({
        title: options.title,
        description: message,
        variant: 'destructive',
      });
    } finally {
      setIsRunning(false);
    }
  };

  const handleValidate = () => {
    if (!selectedTemplate) {
      toast({
        title: 'Selecciona un template',
        description: 'Elige un template para ejecutar la validación.',
      });
      return;
    }

    handleTask({
      title: `Validar template ${selectedTemplate.name}`,
      executor: () => kthuluApi.validateTemplate(selectedTemplate.name),
    });
  };

  const handleRender = () => {
    if (!selectedTemplate) {
      toast({
        title: 'Selecciona un template',
        description: 'Elige un template para poder renderizarlo.',
      });
      return;
    }

    try {
      const vars = parseJsonInput(renderVariables);
      handleTask<TemplateRenderResult>({
        title: `Render de ${selectedTemplate.name}`,
        payload: vars,
        executor: () => kthuluApi.renderTemplate({ name: selectedTemplate.name, vars }),
        onSuccess: (result) => setRenderPreview(result),
      });
    } catch (error) {
      toast({
        title: 'Variables inválidas',
        description: error instanceof Error ? error.message : 'No fue posible leer el JSON proporcionado.',
        variant: 'destructive',
      });
    }
  };

  const handleCache = () => {
    if (!cacheUrl.trim()) {
      toast({
        title: 'URL requerida',
        description: 'Proporciona la URL del template remoto a cachear.',
        variant: 'destructive',
      });
      return;
    }

    handleTask({
      title: 'Cachear template remoto',
      payload: { url: cacheUrl },
      executor: () => kthuluApi.cacheTemplate(cacheUrl.trim()),
      onSuccess: () => setCacheUrl(''),
    });
  };

  const handleAddRegistry = () => {
    if (!registryName.trim() || !registryUrl.trim()) {
      toast({
        title: 'Datos incompletos',
        description: 'Necesitas especificar nombre y URL del registry.',
        variant: 'destructive',
      });
      return;
    }

    handleTask({
      title: `Registrar ${registryName}`,
      payload: { name: registryName, url: registryUrl },
      executor: () => kthuluApi.addRegistry(registryName.trim(), registryUrl.trim()),
      onSuccess: () => {
        setRegistryName('');
        setRegistryUrl('');
      },
    });
  };

  const handleRemoveRegistry = () => {
    if (!registryName.trim()) {
      toast({
        title: 'Nombre requerido',
        description: 'Indica el nombre del registry que deseas eliminar.',
        variant: 'destructive',
      });
      return;
    }

    handleTask({
      title: `Eliminar registry ${registryName}`,
      payload: { name: registryName },
      executor: () => kthuluApi.removeRegistry(registryName.trim()),
      onSuccess: () => setRegistryName(''),
    });
  };

  const handleCleanCache = () => {
    handleTask({
      title: 'Limpiar cache de templates',
      payload: cleanMaxAge ? { maxAge: cleanMaxAge } : undefined,
      executor: () => kthuluApi.cleanTemplates(cleanMaxAge || undefined),
      onSuccess: () => setCleanMaxAge(''),
    });
  };

  const handleSyncSource = () => {
    if (!syncSource.trim()) {
      toast({
        title: 'Origen requerido',
        description: 'Indica un repositorio o ruta a sincronizar.',
        variant: 'destructive',
      });
      return;
    }

    handleTask<TemplateSyncResult>({
      title: 'Sync desde origen',
      payload: { source: syncSource },
      executor: () => kthuluApi.syncTemplatesFromSource(syncSource.trim()),
      onSuccess: () => setSyncSource(''),
    });
  };

  const handleVerify = () => {
    handleTask<TemplateDriftReport>({
      title: 'Verificar drift de templates',
      executor: () => kthuluApi.verifyTemplates(),
    });
  };

  const handleSyncRegistries = () => {
    handleTask({
      title: 'Sincronizar registries',
      executor: () => kthuluApi.syncRegistries(),
    });
  };

  const handleUpdateTemplates = () => {
    handleTask({
      title: 'Actualizar templates registrados',
      executor: () => kthuluApi.updateTemplates(),
    });
  };

  return (
    <div className="h-full bg-kthulu-surface1 grid grid-cols-1 lg:grid-cols-[320px_1fr]">
      <div className="border-r border-primary/20 bg-kthulu-surface2 flex flex-col">
        <div className="p-4 border-b border-primary/20 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="font-mono font-bold text-primary text-sm flex items-center gap-2">
              <Box className="w-4 h-4" />
              TEMPLATES
            </h2>
            <Button variant="ghost" size="icon" onClick={() => refetch()} disabled={isLoading || isRunning}>
              <RefreshCcw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
          <div className="relative">
            <Search className="w-4 h-4 text-muted-foreground absolute left-3 top-3" />
            <Input
              placeholder="Buscar template..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="pl-9 bg-kthulu-surface1 border-primary/30 font-mono text-sm"
            />
          </div>
        </div>
        <ScrollArea className="flex-1">
          <div className="p-4 space-y-2">
            {filteredTemplates.map((template) => (
              <Card
                key={template.name}
                className={`bg-kthulu-surface1 border ${selectedTemplate?.name === template.name ? 'border-primary/60' : 'border-primary/10'} cursor-pointer hover:border-primary/60 transition-colors`}
                onClick={() => setSelectedTemplate(template)}
              >
                <CardHeader className="pb-2">
                  <CardTitle className="font-mono text-sm text-primary flex items-center justify-between">
                    <span>{template.name}</span>
                    {template.version && (
                      <Badge variant="outline" className="text-[10px] font-mono">v{template.version}</Badge>
                    )}
                  </CardTitle>
                </CardHeader>
                {template.description && (
                  <CardContent className="pt-0 text-xs text-muted-foreground font-mono">
                    {template.description}
                  </CardContent>
                )}
              </Card>
            ))}
            {!isLoading && filteredTemplates.length === 0 && (
              <div className="text-xs text-muted-foreground font-mono">
                No se encontraron templates para la búsqueda realizada.
              </div>
            )}
          </div>
        </ScrollArea>
      </div>

      <div className="flex flex-col gap-4 p-6 overflow-y-auto">
        <Card className="bg-kthulu-surface2 border-primary/20">
          <CardHeader>
            <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
              <ServerCog className="w-4 h-4" />
              Operaciones con template
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {selectedTemplate ? (
              <div className="space-y-3 text-xs font-mono">
                <p className="text-muted-foreground">
                  {selectedTemplate.description || 'Template sin descripción detallada.'}
                </p>
                <div className="flex flex-wrap gap-2">
                  {selectedTemplate.tags?.map((tag) => (
                    <Badge key={tag} variant="outline" className="text-[11px]">
                      #{tag}
                    </Badge>
                  ))}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <span className="text-muted-foreground">Versión:</span>
                    <span className="ml-2 text-primary">{selectedTemplate.version ?? 'latest'}</span>
                  </div>
                  {selectedTemplate.author && (
                    <div>
                      <span className="text-muted-foreground">Autor:</span>
                      <span className="ml-2 text-primary">{selectedTemplate.author}</span>
                    </div>
                  )}
                  {selectedTemplate.url && (
                    <div className="col-span-2 truncate">
                      <span className="text-muted-foreground">Origen:</span>
                      <span className="ml-2 text-primary">{selectedTemplate.url}</span>
                    </div>
                  )}
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={handleValidate}
                    disabled={isRunning}
                    className="bg-kthulu-surface1 border-primary/30"
                  >
                    <ShieldCheck className="w-4 h-4 mr-2" /> Validar
                  </Button>
                  <Button size="sm" onClick={handleRender} disabled={isRunning} className="bg-primary text-background">
                    <FileJson className="w-4 h-4 mr-2" /> Render
                  </Button>
                </div>
                <div className="space-y-2">
                  <Textarea
                    rows={4}
                    placeholder='{"service":"billing"}'
                    value={renderVariables}
                    onChange={(event) => setRenderVariables(event.target.value)}
                    className="bg-kthulu-surface1 border-primary/30 font-mono"
                  />
                  <p className="text-[11px] text-muted-foreground">
                    Variables opcionales en formato JSON para personalizar el render.
                  </p>
                </div>
              </div>
            ) : (
              <p className="text-xs text-muted-foreground font-mono">
                Selecciona un template del catálogo para habilitar las acciones de validación y render.
              </p>
            )}
          </CardContent>
        </Card>

        <div className="grid md:grid-cols-2 gap-4">
          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                <CloudDownload className="w-4 h-4" /> Cache remoto
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Input
                value={cacheUrl}
                onChange={(event) => setCacheUrl(event.target.value)}
                placeholder="https://github.com/org/template.zip"
                className="bg-kthulu-surface1 border-primary/30 font-mono text-xs"
              />
              <Button onClick={handleCache} disabled={isRunning} className="bg-primary text-background font-mono">
                <UploadCloud className="w-4 h-4 mr-2" /> Cachear
              </Button>
            </CardContent>
          </Card>

          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                <Globe className="w-4 h-4" /> Registries
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Input
                value={registryName}
                onChange={(event) => setRegistryName(event.target.value)}
                placeholder="kthulu-official"
                className="bg-kthulu-surface1 border-primary/30 font-mono text-xs"
              />
              <Input
                value={registryUrl}
                onChange={(event) => setRegistryUrl(event.target.value)}
                placeholder="https://registry.kthulu.dev"
                className="bg-kthulu-surface1 border-primary/30 font-mono text-xs"
              />
              <div className="flex gap-2">
                <Button onClick={handleAddRegistry} disabled={isRunning} className="bg-primary text-background font-mono flex-1">
                  {isRunning ? (
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  ) : (
                    <Globe className="w-4 h-4 mr-2" />
                  )}
                  Registrar
                </Button>
                <Button variant="outline" onClick={handleRemoveRegistry} disabled={isRunning} className="font-mono flex-1">
                  <ShieldAlert className="w-4 h-4 mr-2" /> Eliminar
                </Button>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" onClick={handleSyncRegistries} disabled={isRunning} className="font-mono flex-1">
                  <RefreshCcw className="w-4 h-4 mr-2" /> Sync
                </Button>
                <Button variant="outline" onClick={handleUpdateTemplates} disabled={isRunning} className="font-mono flex-1">
                  <Settings2 className="w-4 h-4 mr-2" /> Update
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid md:grid-cols-2 gap-4">
          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                <DatabaseZap className="w-4 h-4" /> Limpieza de cache
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Input
                value={cleanMaxAge}
                onChange={(event) => setCleanMaxAge(event.target.value)}
                placeholder="72h"
                className="bg-kthulu-surface1 border-primary/30 font-mono text-xs"
              />
              <Button variant="outline" onClick={handleCleanCache} disabled={isRunning} className="font-mono">
                Limpiar cache
              </Button>
            </CardContent>
          </Card>

          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                <UploadCloud className="w-4 h-4" /> Sync desde origen
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Input
                value={syncSource}
                onChange={(event) => setSyncSource(event.target.value)}
                placeholder="git@github.com:org/templates.git"
                className="bg-kthulu-surface1 border-primary/30 font-mono text-xs"
              />
              <Button onClick={handleSyncSource} disabled={isRunning} className="bg-primary text-background font-mono">
                <UploadCloud className="w-4 h-4 mr-2" /> Sync
              </Button>
            </CardContent>
          </Card>
        </div>

        {renderPreview && (
          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                <FileJson className="w-4 h-4" /> Resultado del render
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Textarea
                value={JSON.stringify(renderPreview.files, null, 2)}
                readOnly
                rows={10}
                className="bg-kthulu-surface1 border-primary/20 font-mono text-[11px]"
              />
            </CardContent>
          </Card>
        )}

        <Card className="bg-kthulu-surface2 border-primary/20">
          <CardHeader>
            <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4" /> Registro de tareas
            </CardTitle>
          </CardHeader>
          <CardContent>
            {taskLog.length === 0 ? (
              <p className="text-xs text-muted-foreground font-mono">
                Ejecuta alguna operación para visualizar los resultados estructurados.
              </p>
            ) : (
              <ScrollArea className="max-h-64">
                <div className="space-y-3 text-xs font-mono">
                  {taskLog.map((entry, index) => (
                    <div
                      key={`${entry.timestamp}-${index}`}
                      className={`border rounded-sm p-3 ${entry.success ? 'border-primary/40' : 'border-destructive/40'}`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-primary font-semibold">{entry.title}</span>
                        <span className="text-muted-foreground text-[10px]">{entry.timestamp}</span>
                      </div>
                      {entry.payload && (
                        <div className="space-y-1">
                          <span className="text-muted-foreground">Payload:</span>
                          <pre className="bg-kthulu-surface1 border border-primary/10 rounded-sm p-2 text-[10px] overflow-x-auto">
{JSON.stringify(entry.payload, null, 2)}
                          </pre>
                        </div>
                      )}
                      {entry.response && (
                        <div className="space-y-1">
                          <span className="text-muted-foreground">Respuesta:</span>
                          <pre className="bg-kthulu-surface1 border border-primary/10 rounded-sm p-2 text-[10px] overflow-x-auto">
{JSON.stringify(entry.response, null, 2)}
                          </pre>
                        </div>
                      )}
                      {entry.error && (
                        <p className="text-destructive">{entry.error}</p>
                      )}
                    </div>
                  ))}
                </div>
              </ScrollArea>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
