import { Activity, Database, GitBranch, Layers, AlertTriangle, CheckCircle, Clock } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { useEffect, useState } from 'react';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { AuditResult } from '@/types/kthulu';

const metrics = {
  services: 5,
  entities: 12,
  usecases: 18,
  dependencies: 23,
  coverage: 87,
  issues: 3,
};

const dependencies = [
  { from: 'auth-service', to: 'user-entity', type: 'direct', status: 'ok' },
  { from: 'payment-service', to: 'auth-service', type: 'api', status: 'warning' },
  { from: 'order-service', to: 'payment-service', type: 'event', status: 'ok' },
  { from: 'notification-service', to: 'order-service', type: 'async', status: 'error' },
];

const issues = [
  { type: 'warning', message: 'Dependencia circular detectada: payment ↔ order', severity: 'medium' },
  { type: 'error', message: 'Entidad User sin validaciones de email', severity: 'high' },
  { type: 'info', message: 'Caso de uso LoginUser podría optimizarse', severity: 'low' },
];

export function Dashboard() {
  const [auditData, setAuditData] = useState<AuditResult | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    const loadAuditData = async () => {
      try {
        setIsLoading(true);
        const result = await kthuluApi.runAudit({});
        setAuditData(result);
      } catch (error) {
        console.error('Failed to load audit data:', error);
        toast({
          title: 'Info',
          description: 'No se pudo conectar con Kthulu API. Usando datos de ejemplo.',
        });
      } finally {
        setIsLoading(false);
      }
    };

    loadAuditData();
  }, [toast]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ok': return <CheckCircle className="w-3 h-3 text-primary" />;
      case 'warning': return <AlertTriangle className="w-3 h-3 text-accent" />;
      case 'error': return <AlertTriangle className="w-3 h-3 text-destructive" />;
      default: return <Clock className="w-3 h-3 text-muted-foreground" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ok': return 'border-primary/30';
      case 'warning': return 'border-accent/30';
      case 'error': return 'border-destructive/30';
      default: return 'border-muted/30';
    }
  };

  const getSeverityVariant = (severity: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (severity) {
      case 'high': return 'destructive';
      case 'medium': return 'secondary';
      case 'low': return 'outline';
      default: return 'default';
    }
  };

  const renderSkeletonMetrics = () => (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {Array.from({ length: 3 }).map((_, index) => (
        <Card key={index} className="bg-kthulu-surface2 border-primary/10">
          <CardHeader className="pb-2 space-y-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-3 w-24" />
          </CardHeader>
          <CardContent className="space-y-3">
            <Skeleton className="h-8 w-16" />
            <Skeleton className="h-3 w-36" />
          </CardContent>
        </Card>
      ))}
    </div>
  );

  const renderSkeletonDependencies = () => (
    <Card className="bg-kthulu-surface2 border-primary/10">
      <CardHeader className="space-y-2">
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-3 w-32" />
      </CardHeader>
      <CardContent className="space-y-3">
        {Array.from({ length: 4 }).map((_, index) => (
          <Skeleton key={index} className="h-12 w-full" />
        ))}
      </CardContent>
    </Card>
  );

  const renderSkeletonIssues = () => (
    <Card className="bg-kthulu-surface2 border-destructive/10">
      <CardHeader className="space-y-2">
        <Skeleton className="h-4 w-52" />
        <Skeleton className="h-3 w-28" />
      </CardHeader>
      <CardContent className="space-y-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Skeleton key={index} className="h-16 w-full" />
        ))}
      </CardContent>
    </Card>
  );

  return (
    <div className="h-full bg-kthulu-surface1 p-6 space-y-6">
      {/* Header */}
      <div>
        <h2 className="font-mono font-bold text-primary text-2xl">DASHBOARD ARQUITECTURA</h2>
        <p className="text-muted-foreground font-mono text-sm">Métricas y análisis del proyecto</p>
      </div>

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList className="bg-kthulu-surface2 border border-primary/20">
          <TabsTrigger value="overview" className="font-mono text-sm">Overview</TabsTrigger>
          <TabsTrigger value="dependencies" className="font-mono text-sm">Dependencias</TabsTrigger>
          <TabsTrigger value="issues" className="font-mono text-sm">Issues</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {isLoading && !auditData ? (
            renderSkeletonMetrics()
          ) : (
            <>
              {/* Métricas principales */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card className="bg-kthulu-surface2 border-primary/20">
                  <CardHeader className="pb-2">
                    <CardTitle className="flex items-center gap-2 text-primary font-mono text-sm">
                      <Layers className="w-4 h-4" />
                      SERVICIOS
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-primary font-mono">
                      {auditData?.counts?.service || metrics.services}
                    </div>
                    <div className="text-xs text-muted-foreground font-mono">microservicios activos</div>
                  </CardContent>
                </Card>

                <Card className="bg-kthulu-surface2 border-secondary/20">
                  <CardHeader className="pb-2">
                    <CardTitle className="flex items-center gap-2 text-secondary font-mono text-sm">
                      <Database className="w-4 h-4" />
                      ENTIDADES
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-secondary font-mono">
                      {auditData?.counts?.entity || metrics.entities}
                    </div>
                    <div className="text-xs text-muted-foreground font-mono">entidades de dominio</div>
                  </CardContent>
                </Card>

                <Card className="bg-kthulu-surface2 border-accent/20">
                  <CardHeader className="pb-2">
                    <CardTitle className="flex items-center gap-2 text-accent font-mono text-sm">
                      <Activity className="w-4 h-4" />
                      CASOS DE USO
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-accent font-mono">
                      {auditData?.counts?.usecase || metrics.usecases}
                    </div>
                    <div className="text-xs text-muted-foreground font-mono">flujos de negocio</div>
                  </CardContent>
                </Card>
              </div>

              {/* Cobertura y Calidad */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Card className="bg-kthulu-surface2 border-primary/20">
                  <CardHeader>
                    <CardTitle className="font-mono text-primary text-sm">COBERTURA DE TESTS</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="font-mono text-sm text-foreground">Tests unitarios</span>
                      <span className="font-mono text-sm text-primary">{metrics.coverage}%</span>
                    </div>
                    <Progress value={metrics.coverage} className="h-2" />
                    <div className="text-xs text-muted-foreground font-mono">
                      87% de cobertura total • 124/142 casos cubiertos
                    </div>
                  </CardContent>
                </Card>

                <Card className="bg-kthulu-surface2 border-accent/20">
                  <CardHeader>
                    <CardTitle className="font-mono text-accent text-sm">CALIDAD DEL CÓDIGO</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <div className="text-lg font-bold font-mono text-primary">A</div>
                        <div className="text-xs text-muted-foreground font-mono">Maintainability</div>
                      </div>
                      <div>
                        <div className="text-lg font-bold font-mono text-destructive">{metrics.issues}</div>
                        <div className="text-xs text-muted-foreground font-mono">Issues críticos</div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </>
          )}
        </TabsContent>

        <TabsContent value="dependencies" className="space-y-4">
          {isLoading && !auditData ? (
            renderSkeletonDependencies()
          ) : (
            <Card className="bg-kthulu-surface2 border-primary/20">
              <CardHeader>
                <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                  <GitBranch className="w-4 h-4" />
                  GRAFO DE DEPENDENCIAS
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {dependencies.map((dep, index) => (
                    <div key={index} className={`p-3 border rounded-sm ${getStatusColor(dep.status)}`}>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          {getStatusIcon(dep.status)}
                          <span className="font-mono text-sm text-foreground">
                            {dep.from} → {dep.to}
                          </span>
                        </div>
                        <Badge variant="outline" className="font-mono text-xs">
                          {dep.type}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="issues" className="space-y-4">
          {isLoading && !auditData ? (
            renderSkeletonIssues()
          ) : (
            <Card className="bg-kthulu-surface2 border-destructive/20">
              <CardHeader>
                <CardTitle className="font-mono text-destructive text-sm flex items-center gap-2">
                  <AlertTriangle className="w-4 h-4" />
                  ISSUES DETECTADOS
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {(auditData?.findings.length ? auditData.findings : issues).map((issue, index) => {
                    const displayIssue = 'message' in issue ? issue : {
                      type: 'info',
                      message: `${issue.kind}: ${issue.detail} (${issue.file}:${issue.line})`,
                      severity: 'medium'
                    };
                    return (
                      <div key={index} className="p-3 border border-primary/20 rounded-sm">
                        <div className="flex items-start gap-3">
                          <AlertTriangle className={`w-4 h-4 mt-0.5 ${
                            displayIssue.severity === 'high' ? 'text-destructive' :
                            displayIssue.severity === 'medium' ? 'text-accent' : 'text-muted-foreground'
                          }`} />
                          <div className="flex-1">
                            <div className="font-mono text-sm text-foreground">{displayIssue.message}</div>
                            <div className="flex items-center gap-2 mt-1">
                              <Badge variant={getSeverityVariant(displayIssue.severity)} className="font-mono text-xs">
                                {displayIssue.severity.toUpperCase()}
                              </Badge>
                              <span className="text-xs text-muted-foreground font-mono">{displayIssue.type}</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}