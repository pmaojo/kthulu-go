import { useEffect, useMemo, useState } from 'react';
import type { Edge, Node } from '@xyflow/react';
import { Plus, Box, Database, Zap, Users, Workflow, ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Badge } from '@/components/ui/badge';
import { kthuluApi } from '@/services/kthuluApi';
import type { TemplateInfo, TemplateRenderResult } from '@/types/kthulu';
import { useToast } from '@/hooks/use-toast';

const nodeTypes = [
  {
    type: 'service',
    label: 'Servicio',
    icon: Box,
    color: 'primary',
    description: 'Microservicio o bounded context'
  },
  {
    type: 'entity',
    label: 'Entidad',
    icon: Database,
    color: 'secondary',
    description: 'Entidad de dominio o agregado'
  },
  {
    type: 'usecase',
    label: 'Caso de Uso',
    icon: Zap,
    color: 'accent',
    description: 'Flujo de negocio o acción'
  },
  {
    type: 'actor',
    label: 'Actor',
    icon: Users,
    color: 'kthulu-cyan',
    description: 'Usuario o sistema externo'
  },
  {
    type: 'workflow',
    label: 'Workflow',
    icon: Workflow,
    color: 'kthulu-purple',
    description: 'Proceso o flujo complejo'
  },
];

interface CanvasToolbarProps {
  onAddNode: (type: string, data: any) => void;
  onTemplateApply: (nodes: Node[], edges?: Edge[]) => void;
  onClearCanvas?: () => void;
  onFitView?: () => void;
  className?: string;
}

type TemplateDiagram = {
  nodes: Node[];
  edges: Edge[];
};

const decodeBase64 = (value: string) => {
  if (typeof globalThis.atob === 'function') {
    return globalThis.atob(value);
  }

  throw new Error('Base64 decoding is not supported in this environment.');
};

const extractDiagramFromTemplate = (result: TemplateRenderResult): TemplateDiagram | null => {
  const candidate = result as TemplateRenderResult & {
    nodes?: Node[];
    edges?: Edge[];
    diagram?: {
      nodes?: Node[];
      edges?: Edge[];
    };
  };

  const normalizeNodes = (nodes: Node[]) =>
    nodes.map((node, index) => ({
      ...node,
      id: String(node.id ?? `template-node-${Date.now()}-${index}`),
      position: node.position ?? { x: 200 * index, y: 100 + (index % 3) * 160 },
      data: node.data ?? {},
    }));

  const normalizeEdges = (edges: Edge[]) =>
    edges
      .filter((edge) => edge.source && edge.target)
      .map((edge, index) => ({
        ...edge,
        type: edge.type ?? 'smoothstep',
        id: String(edge.id ?? `template-edge-${Date.now()}-${index}`),
      }));

  if (Array.isArray(candidate.nodes) && candidate.nodes.length > 0) {
    return {
      nodes: normalizeNodes(candidate.nodes),
      edges: Array.isArray(candidate.edges) ? normalizeEdges(candidate.edges) : [],
    };
  }

  if (candidate.diagram && Array.isArray(candidate.diagram.nodes) && candidate.diagram.nodes.length > 0) {
    return {
      nodes: normalizeNodes(candidate.diagram.nodes),
      edges: Array.isArray(candidate.diagram.edges) ? normalizeEdges(candidate.diagram.edges) : [],
    };
  }

  for (const [path, encodedContent] of Object.entries(result.files ?? {})) {
    if (!path.endsWith('.json')) continue;

    try {
      const decoded = decodeBase64(encodedContent);
      const parsed = JSON.parse(decoded);

      if (Array.isArray(parsed.nodes) && parsed.nodes.length > 0) {
        return {
          nodes: normalizeNodes(parsed.nodes),
          edges: Array.isArray(parsed.edges) ? normalizeEdges(parsed.edges) : [],
        };
      }
    } catch (error) {
      console.warn(`Unable to parse diagram from template file ${path}:`, error);
    }
  }

  return null;
};

export function CanvasToolbar({ onAddNode, onTemplateApply, onClearCanvas, onFitView, className }: CanvasToolbarProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const [templates, setTemplates] = useState<TemplateInfo[]>([]);
  const [isLoadingTemplates, setIsLoadingTemplates] = useState(false);
  const [activeTemplate, setActiveTemplate] = useState<string | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    let isMounted = true;

    const loadTemplates = async () => {
      try {
        setIsLoadingTemplates(true);
        const availableTemplates = await kthuluApi.listTemplates();
        if (isMounted) {
          setTemplates(availableTemplates);
        }
      } catch (error) {
        console.error('Failed to load templates:', error);
        toast({
          title: 'Error al cargar templates',
          description: error instanceof Error ? error.message : 'No se pudieron obtener los templates disponibles.',
          variant: 'destructive',
        });
      } finally {
        if (isMounted) {
          setIsLoadingTemplates(false);
        }
      }
    };

    loadTemplates();

    return () => {
      isMounted = false;
    };
  }, [toast]);

  const filteredNodeTypes = useMemo(() => (
    nodeTypes.filter(nodeType =>
      nodeType.label.toLowerCase().includes(searchTerm.toLowerCase()) ||
      nodeType.description.toLowerCase().includes(searchTerm.toLowerCase())
    )
  ), [searchTerm]);

  const buildDefaultData = (type: string, label: string, description: string) => {
    switch (type) {
      case 'service':
        return {
          label: `Nuevo ${label}`,
          description,
          status: 'active',
        };
      case 'entity':
        return {
          label: `Nueva ${label}`,
          fields: ['id', 'nombre'],
          type: 'entity',
        };
      case 'usecase':
        return {
          label: `Nuevo ${label}`,
          actor: 'Actor principal',
          action: description,
        };
      case 'actor':
        return {
          label: `Nuevo ${label}`,
          role: 'Rol principal',
          responsibility: description,
        };
      case 'workflow':
        return {
          label: `Nuevo ${label}`,
          owner: 'Equipo responsable',
          duration: 'Por definir',
          stages: ['Inicio', 'Proceso', 'Fin'],
        };
      default:
        return {
          label: `Nuevo ${label}`,
          description,
        };
    }
  };

  const handleNodeSelect = (nodeType: any) => {
    onAddNode(nodeType.type, buildDefaultData(nodeType.type, nodeType.label, nodeType.description));
    setIsPopoverOpen(false);
  };

  const handleTemplateSelect = async (template: TemplateInfo) => {
    try {
      setActiveTemplate(template.name);
      const renderResult = await kthuluApi.renderTemplate({ name: template.name });
      const diagram = extractDiagramFromTemplate(renderResult);

      if (!diagram || diagram.nodes.length === 0) {
        toast({
          title: 'Template sin nodos',
          description: 'No fue posible extraer nodos del template seleccionado.',
          variant: 'destructive',
        });
        return;
      }

      onTemplateApply(diagram.nodes, diagram.edges);
      toast({
        title: 'Template aplicado',
        description: `Se agregaron ${diagram.nodes.length} nodos al canvas.`,
      });
    } catch (error) {
      console.error('Failed to render template:', error);
      toast({
        title: 'Error al renderizar template',
        description: error instanceof Error ? error.message : 'No se pudo renderizar el template seleccionado.',
        variant: 'destructive',
      });
    } finally {
      setActiveTemplate(null);
    }
  };

  const getColorClass = (color: string) => {
    switch (color) {
      case 'primary': return 'text-primary border-primary/30';
      case 'secondary': return 'text-secondary border-secondary/30';
      case 'accent': return 'text-accent border-accent/30';
      case 'kthulu-cyan': return 'text-kthulu-cyan border-kthulu-cyan/30';
      case 'kthulu-purple': return 'text-kthulu-purple border-kthulu-purple/30';
      default: return 'text-primary border-primary/30';
    }
  };

  return (
    <div className={`absolute top-4 left-4 z-10 ${className}`}>
      <div className="bg-kthulu-surface2 border border-primary/20 rounded-sm p-4 space-y-4 min-w-[280px]">
        {/* Header */}
        <div>
          <h3 className="font-mono font-bold text-primary text-sm mb-1">TOOLBAR KTHULU</h3>
          <p className="text-xs text-muted-foreground font-mono">Arrastra para crear elementos</p>
        </div>

        {/* Quick Add */}
        <div className="space-y-2">
          <div className="text-xs text-primary font-mono">AGREGAR ELEMENTO:</div>
          <Popover open={isPopoverOpen} onOpenChange={setIsPopoverOpen}>
            <PopoverTrigger asChild>
              <Button 
                variant="outline" 
                className="w-full justify-start bg-kthulu-surface1 border-primary/30 hover:bg-primary/10 font-mono"
              >
                <Plus className="w-4 h-4 mr-2" />
                Nuevo Elemento
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 bg-kthulu-surface2 border-primary/20" align="start">
              <div className="space-y-3">
                <Input
                  placeholder="Buscar tipo de elemento..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="bg-kthulu-surface1 border-primary/30 font-mono text-sm"
                />
                
                <div className="space-y-2 max-h-60 overflow-y-auto">
                  {filteredNodeTypes.map((nodeType) => (
                    <Button
                      key={nodeType.type}
                      variant="ghost"
                      onClick={() => handleNodeSelect(nodeType)}
                      className="w-full justify-start p-3 h-auto hover:bg-kthulu-surface1"
                    >
                      <div className="flex items-start gap-3 w-full">
                        <div className={`w-8 h-8 border rounded-sm flex items-center justify-center ${getColorClass(nodeType.color)}`}>
                          <nodeType.icon className="w-4 h-4" />
                        </div>
                        <div className="flex-1 text-left">
                          <div className="font-mono text-sm text-foreground">{nodeType.label}</div>
                          <div className="font-mono text-xs text-muted-foreground">{nodeType.description}</div>
                        </div>
                      </div>
                    </Button>
                  ))}
                </div>
              </div>
            </PopoverContent>
          </Popover>
        </div>

        {/* Quick Templates */}
        <div className="space-y-2">
          <div className="text-xs text-accent font-mono">TEMPLATES:</div>
          <div className="space-y-1">
            {isLoadingTemplates && (
              <div className="text-xs text-muted-foreground font-mono px-2 py-3 border border-dashed border-accent/40 rounded-sm">
                Cargando templates...
              </div>
            )}

            {!isLoadingTemplates && templates.length === 0 && (
              <div className="text-xs text-muted-foreground font-mono px-2 py-3 border border-dashed border-accent/40 rounded-sm">
                No hay templates disponibles.
              </div>
            )}

            {templates.map((template) => (
              <Button
                key={template.name}
                variant="outline"
                onClick={() => handleTemplateSelect(template)}
                disabled={activeTemplate === template.name}
                className="w-full justify-start bg-kthulu-surface1 border-accent/30 hover:bg-accent/10 font-mono text-xs p-2 h-auto"
              >
                <div className="flex items-start justify-between w-full gap-3">
                  <div className="text-left">
                    <div className="font-bold text-accent flex items-center gap-2">
                      <span>{template.name}</span>
                      {template.version && (
                        <Badge variant="outline" className="text-[10px] font-mono">
                          v{template.version}
                        </Badge>
                      )}
                    </div>
                    {template.description && (
                      <div className="text-muted-foreground text-xs mt-1 line-clamp-2">
                        {template.description}
                      </div>
                    )}
                    {template.tags && template.tags.length > 0 && (
                      <div className="flex items-center gap-1 mt-2 flex-wrap">
                        {template.tags.slice(0, 3).map((tag) => (
                          <Badge key={tag} variant="outline" className="text-[10px] font-mono">
                            {tag}
                          </Badge>
                        ))}
                        {template.tags.length > 3 && (
                          <span className="text-[10px] text-muted-foreground">+{template.tags.length - 3}</span>
                        )}
                      </div>
                    )}
                  </div>
                  <ArrowRight className="w-4 h-4 text-accent shrink-0 mt-1" />
                </div>
              </Button>
            ))}
          </div>
        </div>

        {/* Quick Actions */}
        <div className="pt-2 border-t border-primary/20 space-y-2">
          <div className="text-xs text-muted-foreground font-mono">ACCIONES RÁPIDAS:</div>
          <div className="flex gap-1">
            <Button
              variant="outline"
              size="sm"
              className="flex-1 bg-kthulu-surface1 border-primary/30 hover:bg-primary/10 font-mono text-xs"
              onClick={() => onClearCanvas?.()}
            >
              Limpiar
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="flex-1 bg-kthulu-surface1 border-secondary/30 hover:bg-secondary/10 font-mono text-xs"
              onClick={() => onFitView?.()}
            >
              Centrar
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}