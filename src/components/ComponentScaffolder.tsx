import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Box, Code2, Database, FolderGit2, Loader2, Rocket } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { ComponentRequest, ModuleInfo } from '@/types/kthulu';

const componentTypes: { value: string; label: string; description: string; icon: typeof Box }[] = [
  { value: 'handler', label: 'HTTP Handler', description: 'Controlador HTTP/REST', icon: Code2 },
  { value: 'usecase', label: 'Caso de Uso', description: 'Aplicación o servicio de dominio', icon: Rocket },
  { value: 'entity', label: 'Entidad', description: 'Modelo de dominio con reglas', icon: Database },
  { value: 'repository', label: 'Repositorio', description: 'Puerto de persistencia', icon: FolderGit2 },
  { value: 'service', label: 'Servicio', description: 'Componente de dominio', icon: Box },
];

const defaultRequest: ComponentRequest = {
  type: 'handler',
  name: '',
  withTests: true,
  withMigration: false,
  fields: '',
  relations: '',
  projectPath: '',
};

export function ComponentScaffolder() {
  const { toast } = useToast();
  const [form, setForm] = useState<ComponentRequest>(defaultRequest);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [lastResponse, setLastResponse] = useState<string | null>(null);

  const { data: modules = [], isLoading: loadingModules } = useQuery({
    queryKey: ['modules-for-component'],
    queryFn: async () => {
      try {
        return await kthuluApi.listModules();
      } catch (error) {
        console.error('Failed to load modules', error);
        toast({
          title: 'No se pudieron cargar los módulos',
          description: 'Puedes continuar ingresando el nombre manualmente.',
          variant: 'destructive',
        });
        return [];
      }
    },
    retry: 1,
  });

  const selectedType = useMemo(
    () => componentTypes.find((type) => type.value === form.type) ?? componentTypes[0],
    [form.type],
  );

  const handleSubmit = async () => {
    if (!form.name.trim()) {
      toast({
        title: 'Nombre requerido',
        description: 'Debes indicar el nombre del componente a generar.',
        variant: 'destructive',
      });
      return;
    }

    if (!form.projectPath.trim()) {
      toast({
        title: 'Ruta del proyecto requerida',
        description: 'Indica dónde se encuentra el proyecto existente.',
        variant: 'destructive',
      });
      return;
    }

    const payload: ComponentRequest = {
      ...form,
      fields: form.fields?.trim() || undefined,
      relations: form.relations?.trim() || undefined,
    };

    try {
      setIsSubmitting(true);
      const response = await kthuluApi.generateComponent(payload);
      setLastResponse(JSON.stringify(response, null, 2));
      toast({
        title: 'Componente enviado',
        description: 'La generación fue solicitada al motor Kthulu.',
      });
    } catch (error) {
      console.error('Failed to generate component', error);
      toast({
        title: 'Error generando componente',
        description: 'Verifica que el servicio de scaffolding esté disponible.',
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReset = () => {
    setForm(defaultRequest);
    setLastResponse(null);
  };

  const moduleValue = form.module ?? 'none';

  return (
    <div className="h-full bg-kthulu-surface1 p-6 overflow-y-auto">
      <div className="max-w-5xl mx-auto space-y-6">
        <div>
          <h2 className="font-mono font-bold text-primary text-2xl">Scaffolding de Componentes</h2>
          <p className="text-muted-foreground font-mono text-sm">
            Define el contexto del componente y envía el trabajo al generador del proyecto existente.
          </p>
        </div>

        <div className="grid lg:grid-cols-[360px_1fr] gap-6">
          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm">Configuración</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Tipo de componente</Label>
                <Select value={form.type} onValueChange={(value) => setForm((prev) => ({ ...prev, type: value }))}>
                  <SelectTrigger className="bg-kthulu-surface1 border-primary/30 font-mono">
                    <SelectValue placeholder="Selecciona tipo" />
                  </SelectTrigger>
                  <SelectContent className="bg-kthulu-surface2 border-primary/20">
                    {componentTypes.map((type) => (
                      <SelectItem key={type.value} value={type.value} className="font-mono text-sm">
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <div className="flex items-center gap-2 text-xs text-muted-foreground font-mono">
                  <selectedType.icon className="w-4 h-4 text-primary" />
                  <span>{selectedType.description}</span>
                </div>
              </div>

              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Nombre</Label>
                <Input
                  value={form.name}
                  onChange={(event) => setForm((prev) => ({ ...prev, name: event.target.value }))}
                  placeholder="CreateUser"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>

              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Módulo</Label>
                <Select
                  value={moduleValue}
                  onValueChange={(value) =>
                    setForm((prev) => ({ ...prev, module: value === 'none' ? undefined : value }))
                  }
                >
                  <SelectTrigger className="bg-kthulu-surface1 border-primary/30 font-mono">
                    <SelectValue placeholder="Selecciona módulo" />
                  </SelectTrigger>
                  <SelectContent className="bg-kthulu-surface2 border-primary/20">
                    <SelectItem value="none" className="font-mono text-sm">
                      Sin módulo específico
                    </SelectItem>
                    {modules.map((module) => (
                      <SelectItem key={module.name} value={module.name} className="font-mono text-sm">
                        {module.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground font-mono">
                  Los módulos disponibles se extraen del catálogo activo.
                </p>
              </div>

              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Ruta del proyecto</Label>
                <Input
                  value={form.projectPath}
                  onChange={(event) => setForm((prev) => ({ ...prev, projectPath: event.target.value }))}
                  placeholder="~/workspace/mi-servicio"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>

              <div className="flex items-center justify-between border border-primary/20 rounded-sm px-3 py-2">
                <div>
                  <p className="font-mono text-xs text-primary">Incluir pruebas</p>
                  <p className="text-[11px] text-muted-foreground font-mono">Genera archivos de test unitario.</p>
                </div>
                <Switch
                  checked={!!form.withTests}
                  onCheckedChange={(checked) => setForm((prev) => ({ ...prev, withTests: checked }))}
                />
              </div>

              <div className="flex items-center justify-between border border-primary/20 rounded-sm px-3 py-2">
                <div>
                  <p className="font-mono text-xs text-primary">Incluir migración</p>
                  <p className="text-[11px] text-muted-foreground font-mono">Crea una migración SQL vinculada.</p>
                </div>
                <Switch
                  checked={!!form.withMigration}
                  onCheckedChange={(checked) => setForm((prev) => ({ ...prev, withMigration: checked }))}
                />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm">Contexto y opciones</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label className="font-mono text-xs text-primary">Campos</Label>
                  <Textarea
                    rows={6}
                    placeholder="name:string\nemail:string"
                    value={form.fields ?? ''}
                    onChange={(event) => setForm((prev) => ({ ...prev, fields: event.target.value }))}
                    className="bg-kthulu-surface1 border-primary/30 font-mono"
                  />
                  <p className="text-[11px] text-muted-foreground font-mono">
                    Define los campos principales del componente o entidad (formato libre).
                  </p>
                </div>
                <div className="space-y-2">
                  <Label className="font-mono text-xs text-primary">Relaciones</Label>
                  <Textarea
                    rows={6}
                    placeholder="User -> Session"
                    value={form.relations ?? ''}
                    onChange={(event) => setForm((prev) => ({ ...prev, relations: event.target.value }))}
                    className="bg-kthulu-surface1 border-primary/30 font-mono"
                  />
                  <p className="text-[11px] text-muted-foreground font-mono">
                    Describe relaciones o dependencias requeridas por el componente.
                  </p>
                </div>
              </div>

              <div className="flex gap-2">
                <Button onClick={handleSubmit} disabled={isSubmitting} className="bg-primary text-background font-mono">
                  {isSubmitting && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                  Generar componente
                </Button>
                <Button variant="outline" onClick={handleReset} className="font-mono">
                  Limpiar
                </Button>
              </div>

              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Payload enviado</Label>
                <pre className="bg-kthulu-surface1 border border-primary/20 rounded-sm p-4 text-[11px] overflow-x-auto">
{JSON.stringify({ ...form, fields: form.fields || undefined, relations: form.relations || undefined }, null, 2)}
                </pre>
              </div>

              {lastResponse && (
                <div className="space-y-2">
                  <Label className="font-mono text-xs text-primary">Respuesta del servicio</Label>
                  <pre className="bg-kthulu-surface1 border border-primary/20 rounded-sm p-4 text-[11px] overflow-x-auto">
{lastResponse}
                  </pre>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {modules.length > 0 && (
          <Card className="bg-kthulu-surface2 border-primary/20">
            <CardHeader>
              <CardTitle className="font-mono text-primary text-sm">Módulos disponibles</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-40">
                <div className="flex flex-wrap gap-2">
                  {modules.map((module) => (
                    <Badge key={module.name} variant="outline" className="font-mono text-[11px]">
                      {module.name}
                    </Badge>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
