import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Workflow, GitBranch, TimerReset } from 'lucide-react';

interface WorkflowNodeData {
  label: string;
  stages?: string[];
  owner?: string;
  duration?: string;
}

interface WorkflowNodeProps {
  data: WorkflowNodeData;
  id: string;
}

export const WorkflowNode = memo(({ data }: WorkflowNodeProps) => {
  return (
    <div className="min-w-[240px] bg-kthulu-surface2 border-2 border-kthulu-purple/30 rounded-sm kthulu-transition hover:border-kthulu-purple hover:shadow-[0_0_20px_hsl(var(--kthulu-neon-purple)/0.5)]">
      <Handle
        type="target"
        position={Position.Left}
        className="w-3 h-3 bg-kthulu-purple border-0"
      />

      <div className="px-3 py-2 bg-kthulu-purple/10 border-b border-kthulu-purple/20">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-kthulu-purple rounded-sm flex items-center justify-center">
            <Workflow className="w-3 h-3 text-background" />
          </div>
          <h3 className="font-mono font-bold text-kthulu-purple text-sm">
            {data.label}
          </h3>
        </div>
        <div className="text-xs text-kthulu-purple/70 font-mono mt-1">
          WORKFLOW
        </div>
      </div>

      <div className="p-3 space-y-2">
        {data.owner && (
          <div className="flex items-center gap-2 text-xs font-mono">
            <GitBranch className="w-3 h-3 text-kthulu-purple" />
            <span className="text-muted-foreground">Owner:</span>
            <span className="text-foreground">{data.owner}</span>
          </div>
        )}

        {data.duration && (
          <div className="flex items-center gap-2 text-xs font-mono">
            <TimerReset className="w-3 h-3 text-kthulu-purple" />
            <span className="text-muted-foreground">Duración:</span>
            <span className="text-foreground">{data.duration}</span>
          </div>
        )}

        {data.stages && data.stages.length > 0 ? (
          <div className="space-y-1">
            {data.stages.slice(0, 3).map((stage, index) => (
              <div key={index} className="text-xs font-mono text-muted-foreground flex items-center gap-2">
                <div className="w-1.5 h-1.5 bg-kthulu-purple rounded-full" />
                {stage}
              </div>
            ))}
            {data.stages.length > 3 && (
              <div className="text-xs font-mono text-kthulu-purple/60">
                +{data.stages.length - 3} etapas más
              </div>
            )}
          </div>
        ) : (
          <div className="text-xs font-mono text-muted-foreground italic">
            Sin etapas definidas
          </div>
        )}
      </div>

      <Handle
        type="source"
        position={Position.Right}
        className="w-3 h-3 bg-kthulu-purple border-0"
      />
    </div>
  );
});

WorkflowNode.displayName = 'WorkflowNode';
