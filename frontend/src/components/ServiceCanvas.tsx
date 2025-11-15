import { useCallback, useEffect } from 'react';
import {
  ReactFlow,
  addEdge,
  useNodesState,
  useEdgesState,
  Connection,
  ReactFlowProvider,
  useReactFlow,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  type Node,
  type Edge,
} from '@xyflow/react';

import '@xyflow/react/dist/style.css';

import { ServiceNode } from './nodes/ServiceNode';
import { EntityNode } from './nodes/EntityNode';
import { UseCaseNode } from './nodes/UseCaseNode';
import { ActorNode } from './nodes/ActorNode';
import { WorkflowNode } from './nodes/WorkflowNode';
import { CanvasToolbar } from './CanvasToolbar';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';

const nodeTypes = {
  service: ServiceNode,
  entity: EntityNode,
  usecase: UseCaseNode,
  actor: ActorNode,
  workflow: WorkflowNode,
};

const initialNodes: Node[] = [
  {
    id: '1',
    type: 'service',
    position: { x: 100, y: 100 },
    data: { 
      label: 'Auth Service',
      description: 'Manejo de autenticación y autorización',
      status: 'active'
    },
  },
  {
    id: '2',
    type: 'entity',
    position: { x: 400, y: 50 },
    data: { 
      label: 'User',
      fields: ['id', 'email', 'password', 'created_at'],
      type: 'aggregate'
    },
  },
  {
    id: '3',
    type: 'usecase',
    position: { x: 700, y: 100 },
    data: { 
      label: 'Login User',
      actor: 'User',
      action: 'Authenticate credentials and generate token'
    },
  },
  {
    id: '4',
    type: 'entity',
    position: { x: 400, y: 250 },
    data: { 
      label: 'Session',
      fields: ['id', 'user_id', 'token', 'expires_at'],
      type: 'entity'
    },
  },
  {
    id: '5',
    type: 'usecase',
    position: { x: 700, y: 300 },
    data: {
      label: 'Register User',
      actor: 'User',
      action: 'Create new user account with validation'
    },
  },
  {
    id: '6',
    type: 'actor',
    position: { x: 150, y: 320 },
    data: {
      label: 'Customer',
      role: 'Cliente externo',
      responsibility: 'Interactúa con los casos de uso de autenticación'
    },
  },
  {
    id: '7',
    type: 'workflow',
    position: { x: 950, y: 220 },
    data: {
      label: 'User Onboarding',
      owner: 'Product Team',
      duration: '3 días',
      stages: ['Registro', 'Verificación de email', 'Activación de cuenta']
    },
  },
];

const initialEdges: Edge[] = [
  {
    id: 'e1-2',
    source: '1',
    target: '2',
    type: 'smoothstep',
    style: { stroke: 'hsl(var(--primary))' },
  },
  {
    id: 'e1-3',
    source: '1',
    target: '3',
    type: 'smoothstep',
    style: { stroke: 'hsl(var(--accent))' },
  },
  {
    id: 'e2-4',
    source: '2',
    target: '4',
    type: 'smoothstep',
    style: { stroke: 'hsl(var(--secondary))' },
  },
  {
    id: 'e1-5',
    source: '1',
    target: '5',
    type: 'smoothstep',
    style: { stroke: 'hsl(var(--accent))' },
  },
];

interface ServiceCanvasProps {
  className?: string;
  onNodeSelect?: (node: Node) => void;
}

function ServiceCanvasContent({ className, onNodeSelect }: ServiceCanvasProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const { toast } = useToast();
  const { fitView: fitViewport, setNodes: setFlowNodes, setEdges: setFlowEdges } = useReactFlow<Node, Edge>();

  // Load modules from API
  useEffect(() => {
    const loadModules = async () => {
      try {
        const modules = await kthuluApi.listModules();
        
        // Convert modules to nodes
        const moduleNodes: Node[] = modules.map((module, index) => ({
          id: module.name,
          type: 'service',
          position: { x: 100 + (index * 200), y: 100 + (index % 3) * 150 },
          data: {
            label: module.name,
            description: module.description || '',
            status: 'active',
          },
        }));

        if (moduleNodes.length > 0) {
          setNodes(moduleNodes);
        }
      } catch (error) {
        console.error('Failed to load modules:', error);
        toast({
          title: 'Error',
          description: 'No se pudo conectar con Kthulu API. Usando datos de ejemplo.',
          variant: 'destructive',
        });
      }
    };

    loadModules();
  }, [setNodes, toast]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  const handleAddNode = useCallback((type: string, data: any) => {
    const newNode: Node = {
      id: `${Date.now()}`,
      type,
      position: { x: Math.random() * 500 + 100, y: Math.random() * 300 + 100 },
      data,
    };
    setNodes((nds) => [...nds, newNode]);
  }, [setNodes]);

  const handleTemplateApply = useCallback((templateNodes: Node[], templateEdges: Edge[] = []) => {
    setNodes(templateNodes);
    setEdges(templateEdges);
  }, [setEdges, setNodes]);

  const handleClearCanvas = useCallback(() => {
    setNodes([]);
    setEdges([]);
    setFlowNodes([]);
    setFlowEdges([]);
  }, [setEdges, setFlowEdges, setFlowNodes, setNodes]);

  const handleFitView = useCallback(() => {
    fitViewport({ padding: 0.2 });
  }, [fitViewport]);

  const handleNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    if (onNodeSelect) {
      onNodeSelect(node);
    }
  }, [onNodeSelect]);

  return (
    <div className={`w-full h-full relative ${className}`}>
      <CanvasToolbar
        onAddNode={handleAddNode}
        onTemplateApply={handleTemplateApply}
        onClearCanvas={handleClearCanvas}
        onFitView={handleFitView}
      />

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={handleNodeClick}
        nodeTypes={nodeTypes}
        fitView
        className="bg-kthulu-surface1"
        style={{
          backgroundColor: 'hsl(var(--kthulu-surface-1))',
        }}
      >
        <Controls 
          className="bg-kthulu-surface2 border border-primary/20"
          style={{
            background: 'hsl(var(--kthulu-surface-2))',
            border: '1px solid hsl(var(--primary) / 0.2)',
          }}
        />
        <MiniMap 
          className="bg-kthulu-surface2 border border-primary/20"
          style={{
            background: 'hsl(var(--kthulu-surface-2))',
            border: '1px solid hsl(var(--primary) / 0.2)',
          }}
          nodeColor={(node) => {
            switch (node.type) {
              case 'service': return 'hsl(var(--primary))';
              case 'entity': return 'hsl(var(--secondary))';
              case 'usecase': return 'hsl(var(--accent))';
              case 'actor': return 'hsl(var(--kthulu-neon-cyan))';
              case 'workflow': return 'hsl(var(--kthulu-neon-purple))';
              default: return 'hsl(var(--muted))';
            }
          }}
        />
        <Background 
          variant={BackgroundVariant.Lines}
          gap={20}
          size={1}
          color="hsl(var(--primary) / 0.1)"
        />
      </ReactFlow>
    </div>
  );
}

export function ServiceCanvas(props: ServiceCanvasProps) {
  return (
    <ReactFlowProvider>
      <ServiceCanvasContent {...props} />
    </ReactFlowProvider>
  );
}
