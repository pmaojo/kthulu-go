import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Users, ShieldCheck, MessageCircle } from 'lucide-react';

interface ActorNodeData {
  label: string;
  role?: string;
  responsibility?: string;
  contact?: string;
}

interface ActorNodeProps {
  data: ActorNodeData;
  id: string;
}

export const ActorNode = memo(({ data }: ActorNodeProps) => {
  return (
    <div className="min-w-[200px] bg-kthulu-surface2 border-2 border-kthulu-cyan/30 rounded-sm kthulu-transition hover:border-kthulu-cyan hover:shadow-[0_0_20px_hsl(var(--kthulu-neon-cyan)/0.5)]">
      <Handle
        type="target"
        position={Position.Left}
        className="w-3 h-3 bg-kthulu-cyan border-0"
      />

      <div className="px-3 py-2 bg-kthulu-cyan/10 border-b border-kthulu-cyan/20">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-kthulu-cyan rounded-sm flex items-center justify-center">
            <Users className="w-3 h-3 text-background" />
          </div>
          <h3 className="font-mono font-bold text-kthulu-cyan text-sm">
            {data.label}
          </h3>
        </div>
        <div className="text-xs text-kthulu-cyan/70 font-mono mt-1">
          ACTOR
        </div>
      </div>

      <div className="p-3 space-y-2">
        {data.role && (
          <div className="flex items-center gap-2 text-xs font-mono text-muted-foreground">
            <ShieldCheck className="w-3 h-3 text-kthulu-cyan" />
            <span className="text-foreground">{data.role}</span>
          </div>
        )}

        {data.responsibility && (
          <div className="flex items-start gap-2 text-xs font-mono">
            <ArrowBullet />
            <span className="text-foreground">{data.responsibility}</span>
          </div>
        )}

        {data.contact && (
          <div className="flex items-center gap-2 text-xs font-mono text-muted-foreground">
            <MessageCircle className="w-3 h-3 text-kthulu-cyan" />
            <span className="text-foreground">{data.contact}</span>
          </div>
        )}

        {!data.role && !data.responsibility && !data.contact && (
          <div className="text-xs font-mono text-muted-foreground italic">
            Sin información definida
          </div>
        )}
      </div>

      <Handle
        type="source"
        position={Position.Right}
        className="w-3 h-3 bg-kthulu-cyan border-0"
      />
    </div>
  );
});

const ArrowBullet = () => (
  <div className="w-2 h-2 rounded-full bg-kthulu-cyan mt-1" />
);

ActorNode.displayName = 'ActorNode';
