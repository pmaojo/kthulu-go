import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Zap, User, ArrowRight } from 'lucide-react';

interface UseCaseNodeData {
  label: string;
  actor?: string;
  action?: string;
}

interface UseCaseNodeProps {
  data: UseCaseNodeData;
  id: string;
}

export const UseCaseNode = memo(({ data, id }: UseCaseNodeProps) => {
  return (
    <div className="min-w-[220px] bg-kthulu-surface2 border-2 border-accent/30 rounded-sm kthulu-transition hover:border-accent hover:shadow-[0_0_20px_hsl(var(--accent)/0.5)]">
      <Handle 
        type="target" 
        position={Position.Left} 
        className="w-3 h-3 bg-accent border-0"
      />
      
      {/* Header */}
      <div className="px-3 py-2 bg-accent/10 border-b border-accent/20">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-accent rounded-sm flex items-center justify-center">
            <Zap className="w-3 h-3 text-background" />
          </div>
          <h3 className="font-mono font-bold text-accent text-sm">
            {data.label}
          </h3>
        </div>
        <div className="text-xs text-accent/70 font-mono mt-1">
          CASO DE USO
        </div>
      </div>
      
      {/* Content */}
      <div className="p-3 space-y-2">
        {data.actor && (
          <div className="flex items-center gap-2 text-xs font-mono">
            <User className="w-3 h-3 text-accent" />
            <span className="text-muted-foreground">Actor:</span>
            <span className="text-foreground">{data.actor}</span>
          </div>
        )}
        
        {data.action && (
          <div className="flex items-start gap-2 text-xs font-mono">
            <ArrowRight className="w-3 h-3 text-accent mt-0.5 flex-shrink-0" />
            <div>
              <span className="text-muted-foreground block">Acción:</span>
              <span className="text-foreground">{data.action}</span>
            </div>
          </div>
        )}
        
        {(!data.actor && !data.action) && (
          <div className="text-xs font-mono text-muted-foreground italic">
            Sin detalles definidos
          </div>
        )}
      </div>

      <Handle 
        type="source" 
        position={Position.Right} 
        className="w-3 h-3 bg-accent border-0"
      />
    </div>
  );
});

UseCaseNode.displayName = 'UseCaseNode';