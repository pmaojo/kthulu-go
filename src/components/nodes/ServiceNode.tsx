import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Box, Circle } from 'lucide-react';

interface ServiceNodeData {
  label: string;
  description?: string;
  status?: 'active' | 'inactive' | 'error';
}

interface ServiceNodeProps {
  data: ServiceNodeData;
  id: string;
}

export const ServiceNode = memo(({ data, id }: ServiceNodeProps) => {
  const getStatusColor = () => {
    switch (data.status) {
      case 'active': return 'bg-primary';
      case 'error': return 'bg-destructive';
      case 'inactive': return 'bg-muted';
      default: return 'bg-primary';
    }
  };

  const getStatusTextColor = () => {
    switch (data.status) {
      case 'active': return 'text-primary';
      case 'error': return 'text-destructive';
      case 'inactive': return 'text-muted-foreground';
      default: return 'text-primary';
    }
  };

  return (
    <div className="min-w-[200px] bg-kthulu-surface2 border-2 border-primary/30 rounded-sm p-4 kthulu-transition hover:border-primary hover:shadow-neon">
      <Handle 
        type="target" 
        position={Position.Left} 
        className="w-3 h-3 bg-primary border-0"
      />
      
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          <div className="w-8 h-8 bg-gradient-neon rounded-sm flex items-center justify-center">
            <Box className="w-4 h-4 text-background" />
          </div>
        </div>
        
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <h3 className="font-mono font-bold text-primary text-sm">
              {data.label}
            </h3>
            <div className={`w-2 h-2 rounded-full ${getStatusColor()}`} />
          </div>
          
          {data.description && (
            <p className="text-xs text-muted-foreground font-mono leading-relaxed">
              {data.description}
            </p>
          )}
          
          <div className={`text-xs font-mono mt-2 ${getStatusTextColor()}`}>
            STATUS: {(data.status || 'active').toUpperCase()}
          </div>
        </div>
      </div>

      <Handle 
        type="source" 
        position={Position.Right} 
        className="w-3 h-3 bg-primary border-0"
      />
    </div>
  );
});

ServiceNode.displayName = 'ServiceNode';