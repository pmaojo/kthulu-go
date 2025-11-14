import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { Loader2, Zap, Package, Database, Layout, GitBranch, FileCode } from 'lucide-react';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { ProjectRequest, ProjectPlan } from '@/types/kthulu';

interface ProjectGeneratorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ProjectGeneratorDialog({ open, onOpenChange }: ProjectGeneratorDialogProps) {
  const [isGenerating, setIsGenerating] = useState(false);
  const [isPreviewing, setIsPreviewing] = useState(false);
  const [projectPlan, setProjectPlan] = useState<ProjectPlan | null>(null);
  const { toast } = useToast();

  const [formData, setFormData] = useState<ProjectRequest>({
    name: '',
    modules: [],
    template: 'hexagonal-go',
    database: 'postgresql',
    frontend: 'react',
    skipGit: false,
    skipDocker: false,
    author: '',
    license: 'MIT',
    description: '',
    path: '',
    dryRun: false,
  });

  const availableModules = [
    { id: 'auth', name: 'Auth', description: 'Autenticación y autorización' },
    { id: 'user', name: 'User', description: 'Gestión de usuarios' },
    { id: 'payment', name: 'Payment', description: 'Procesamiento de pagos' },
    { id: 'notification', name: 'Notification', description: 'Sistema de notificaciones' },
    { id: 'storage', name: 'Storage', description: 'Almacenamiento de archivos' },
  ];

  const toggleModule = (moduleId: string) => {
    setFormData(prev => ({
      ...prev,
      modules: prev.modules?.includes(moduleId)
        ? prev.modules.filter(m => m !== moduleId)
        : [...(prev.modules || []), moduleId],
    }));
  };

  const handlePreview = async () => {
    try {
      setIsPreviewing(true);
      const plan = await kthuluApi.planProject(formData);
      setProjectPlan(plan);
      
      toast({
        title: 'Plan generado',
        description: `Se crearán ${plan.backendFiles?.length || 0} archivos backend y ${plan.frontendFiles?.length || 0} archivos frontend`,
      });
    } catch (error) {
      console.error('Preview failed:', error);
      toast({
        title: 'Error',
        description: 'No se pudo generar el plan del proyecto',
        variant: 'destructive',
      });
    } finally {
      setIsPreviewing(false);
    }
  };

  const handleGenerate = async () => {
    if (!formData.name) {
      toast({
        title: 'Error',
        description: 'El nombre del proyecto es requerido',
        variant: 'destructive',
      });
      return;
    }

    try {
      setIsGenerating(true);
      const result = await kthuluApi.generateProject(formData);
      
      toast({
        title: '¡Proyecto generado!',
        description: `${result.projectDirectories?.length || 0} directorios creados exitosamente`,
      });

      onOpenChange(false);
      setFormData({
        name: '',
        modules: [],
        template: 'hexagonal-go',
        database: 'postgresql',
        frontend: 'react',
        skipGit: false,
        skipDocker: false,
        author: '',
        license: 'MIT',
        description: '',
        path: '',
        dryRun: false,
      });
      setProjectPlan(null);
    } catch (error) {
      console.error('Generation failed:', error);
      toast({
        title: 'Error',
        description: 'No se pudo generar el proyecto. Verifica que el servidor Kthulu esté corriendo.',
        variant: 'destructive',
      });
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl bg-kthulu-surface2 border-primary/20 max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="font-mono text-primary flex items-center gap-2">
            <Zap className="w-5 h-5" />
            GENERAR PROYECTO KTHULU
          </DialogTitle>
          <DialogDescription className="font-mono text-muted-foreground">
            Configura y genera un nuevo microservicio con arquitectura hexagonal
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Basic Info */}
          <div className="space-y-4">
            <div className="space-y-2">
              <Label className="font-mono text-primary text-sm">NOMBRE DEL PROYECTO *</Label>
              <Input
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="mi-servicio"
                className="bg-kthulu-surface1 border-primary/30 font-mono"
              />
            </div>

            <div className="space-y-2">
              <Label className="font-mono text-primary text-sm">DESCRIPCIÓN</Label>
              <Textarea
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="Breve descripción del servicio..."
                className="bg-kthulu-surface1 border-primary/30 font-mono resize-none"
                rows={2}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="font-mono text-primary text-sm flex items-center gap-1">
                  <FileCode className="w-3 h-3" />
                  TEMPLATE
                </Label>
                <Select
                  value={formData.template}
                  onValueChange={(value) => setFormData(prev => ({ ...prev, template: value }))}
                >
                  <SelectTrigger className="bg-kthulu-surface1 border-primary/30 font-mono">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-kthulu-surface2 border-primary/20">
                    <SelectItem value="hexagonal-go" className="font-mono">Hexagonal Go</SelectItem>
                    <SelectItem value="clean-go" className="font-mono">Clean Go</SelectItem>
                    <SelectItem value="ddd-go" className="font-mono">DDD Go</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label className="font-mono text-primary text-sm flex items-center gap-1">
                  <Database className="w-3 h-3" />
                  BASE DE DATOS
                </Label>
                <Select
                  value={formData.database}
                  onValueChange={(value) => setFormData(prev => ({ ...prev, database: value }))}
                >
                  <SelectTrigger className="bg-kthulu-surface1 border-primary/30 font-mono">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-kthulu-surface2 border-primary/20">
                    <SelectItem value="postgresql" className="font-mono">PostgreSQL</SelectItem>
                    <SelectItem value="mysql" className="font-mono">MySQL</SelectItem>
                    <SelectItem value="sqlite" className="font-mono">SQLite</SelectItem>
                    <SelectItem value="mongodb" className="font-mono">MongoDB</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="font-mono text-primary text-sm flex items-center gap-1">
                  <Layout className="w-3 h-3" />
                  FRONTEND
                </Label>
                <Select
                  value={formData.frontend}
                  onValueChange={(value) => setFormData(prev => ({ ...prev, frontend: value }))}
                >
                  <SelectTrigger className="bg-kthulu-surface1 border-primary/30 font-mono">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="bg-kthulu-surface2 border-primary/20">
                    <SelectItem value="react" className="font-mono">React</SelectItem>
                    <SelectItem value="vue" className="font-mono">Vue</SelectItem>
                    <SelectItem value="none" className="font-mono">Ninguno</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label className="font-mono text-primary text-sm">AUTOR</Label>
                <Input
                  value={formData.author}
                  onChange={(e) => setFormData(prev => ({ ...prev, author: e.target.value }))}
                  placeholder="Tu nombre"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>
            </div>
          </div>

          {/* Modules */}
          <div className="space-y-3">
            <Label className="font-mono text-accent text-sm flex items-center gap-1">
              <Package className="w-3 h-3" />
              MÓDULOS
            </Label>
            <div className="grid grid-cols-2 gap-3">
              {availableModules.map((module) => (
                <div
                  key={module.id}
                  className="flex items-start gap-3 p-3 bg-kthulu-surface1 border border-accent/20 rounded-sm"
                >
                  <Checkbox
                    checked={formData.modules?.includes(module.id)}
                    onCheckedChange={() => toggleModule(module.id)}
                    className="mt-0.5"
                  />
                  <div className="flex-1">
                    <div className="font-mono text-sm text-foreground">{module.name}</div>
                    <div className="font-mono text-xs text-muted-foreground">{module.description}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Options */}
          <div className="space-y-3">
            <Label className="font-mono text-primary text-sm">OPCIONES</Label>
            <div className="space-y-2">
              <div className="flex items-center gap-3 p-2 bg-kthulu-surface1 rounded-sm">
                <Checkbox
                  checked={formData.skipGit}
                  onCheckedChange={(checked) => setFormData(prev => ({ ...prev, skipGit: !!checked }))}
                />
                <Label className="font-mono text-sm cursor-pointer">Sin Git</Label>
              </div>
              <div className="flex items-center gap-3 p-2 bg-kthulu-surface1 rounded-sm">
                <Checkbox
                  checked={formData.skipDocker}
                  onCheckedChange={(checked) => setFormData(prev => ({ ...prev, skipDocker: !!checked }))}
                />
                <Label className="font-mono text-sm cursor-pointer">Sin Docker</Label>
              </div>
            </div>
          </div>

          {/* Plan Preview */}
          {projectPlan && (
            <div className="space-y-3 p-4 bg-kthulu-surface1 border border-primary/20 rounded-sm">
              <div className="font-mono text-primary text-sm font-bold">PLAN DE GENERACIÓN</div>
              <div className="space-y-2 text-xs font-mono">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Directorios:</span>
                  <Badge variant="outline">{projectPlan.projectDirectories?.length || 0}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Archivos Backend:</span>
                  <Badge variant="outline">{projectPlan.backendFiles?.length || 0}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Archivos Frontend:</span>
                  <Badge variant="outline">{projectPlan.frontendFiles?.length || 0}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Módulos:</span>
                  <Badge variant="outline">{projectPlan.modules?.length || 0}</Badge>
                </div>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className="flex gap-2">
          <Button
            variant="outline"
            onClick={handlePreview}
            disabled={isPreviewing || isGenerating || !formData.name}
            className="bg-kthulu-surface1 border-accent/30 hover:bg-accent/10 font-mono"
          >
            {isPreviewing && <Loader2 className="w-3 h-3 mr-2 animate-spin" />}
            Vista Previa
          </Button>
          <Button
            onClick={handleGenerate}
            disabled={isGenerating || !formData.name}
            className="bg-gradient-neon text-background hover:opacity-90 font-mono"
          >
            {isGenerating && <Loader2 className="w-3 h-3 mr-2 animate-spin" />}
            Generar Proyecto
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
