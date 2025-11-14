import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Database, Table } from 'lucide-react';

interface EntityNodeData {
  label: string;
  fields?: string[];
  type?: 'entity' | 'aggregate' | 'value_object';
}

interface EntityNodeProps {
  data: EntityNodeData;
  id: string;
}

export const EntityNode = memo(({ data, id }: EntityNodeProps) => {
  const getTypeColor = () => {
    switch (data.type) {
      case 'aggregate': return 'bg-secondary';
      case 'value_object': return 'bg-accent';
      case 'entity': return 'bg-kthulu-cyan';
      default: return 'bg-secondary';
    }
  };

  const getTypeIcon = () => {
    switch (data.type) {
      case 'aggregate': return Database;
      case 'value_object': return Table;
      case 'entity': return Database;
      default: return Database;
    }
  };

  const TypeIcon = getTypeIcon();

  return (
    <div className="min-w-[180px] bg-kthulu-surface2 border-2 border-secondary/30 rounded-sm kthulu-transition hover:border-secondary hover:shadow-[0_0_20px_hsl(var(--secondary)/0.5)]">
      <Handle 
        type="target" 
        position={Position.Left} 
        className="w-3 h-3 bg-secondary border-0"
      />
      
      {/* Header */}
      <div className="px-3 py-2 bg-secondary/10 border-b border-secondary/20">
        <div className="flex items-center gap-2">
          <div className={`w-6 h-6 ${getTypeColor()} rounded-sm flex items-center justify-center`}>
            <TypeIcon className="w-3 h-3 text-background" />
          </div>
          <h3 className="font-mono font-bold text-secondary text-sm">
            {data.label}
          </h3>
        </div>
        <div className="text-xs text-secondary/70 font-mono mt-1">
          {(data.type || 'entity').toUpperCase()}
        </div>
      </div>
      
      {/* Fields */}
      <div className="p-3">
        {data.fields && data.fields.length > 0 ? (
          <div className="space-y-1">
            {data.fields.slice(0, 4).map((field, index) => (
              <div key={index} className="text-xs font-mono text-muted-foreground flex items-center gap-2">
                <div className="w-1 h-1 bg-secondary rounded-full" />
                {field}
              </div>
            ))}
            {data.fields.length > 4 && (
              <div className="text-xs font-mono text-secondary/50">
                +{data.fields.length - 4} más...
              </div>
            )}
          </div>
        ) : (
          <div className="text-xs font-mono text-muted-foreground italic">
            Sin campos definidos
          </div>
        )}
      </div>

      <Handle 
        type="source" 
        position={Position.Right} 
        className="w-3 h-3 bg-secondary border-0"
      />
    </div>
  );
});

EntityNode.displayName = 'EntityNode';