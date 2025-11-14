import { useState } from 'react';
import { ClipboardList, Flame, FolderSearch, RefreshCcw } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';
import type { AuditRequest, AuditResult } from '@/types/kthulu';

const sanitizeList = (value: string) =>
  value
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean);

const defaultRequest: AuditRequest = {
  path: '',
  onlyKinds: [],
  extensions: [],
  ignore: [],
  strict: false,
  jobs: 4,
};

export function AuditWorkbench() {
  const { toast } = useToast();
  const [request, setRequest] = useState<AuditRequest>(defaultRequest);
  const [isRunning, setIsRunning] = useState(false);
  const [result, setResult] = useState<AuditResult | null>(null);

  const handleExecute = async () => {
    try {
      setIsRunning(true);
      const payload: AuditRequest = {
        ...request,
        path: request.path?.trim() || undefined,
        onlyKinds: request.onlyKinds && request.onlyKinds.length > 0 ? request.onlyKinds : undefined,
        extensions: request.extensions && request.extensions.length > 0 ? request.extensions : undefined,
        ignore: request.ignore && request.ignore.length > 0 ? request.ignore : undefined,
        strict: request.strict,
        jobs: request.jobs,
      };

      const response = await kthuluApi.runAudit(payload);
      setResult(response);
      toast({
        title: 'Audit completado',
        description: `Escaneo finalizado en ${response.duration}.`,
      });
    } catch (error) {
      console.error('Failed to run audit', error);
      toast({
        title: 'Error en audit',
        description: 'No fue posible ejecutar el análisis. Verifica la conexión con el backend.',
        variant: 'destructive',
      });
    } finally {
      setIsRunning(false);
    }
  };

  const resetForm = () => {
    setRequest(defaultRequest);
    setResult(null);
  };

  const countsEntries = result ? Object.entries(result.counts || {}) : [];

  return (
    <div className="h-full bg-kthulu-surface1 p-6 overflow-y-auto">
      <div className="max-w-6xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-mono font-bold text-primary text-2xl">Auditoría de repositorios</h2>
            <p className="text-muted-foreground font-mono text-sm">
              Configura el alcance del escaneo y visualiza desviaciones frente a las convenciones Kthulu.
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={resetForm} className="font-mono" disabled={isRunning}>
              <RefreshCcw className="w-4 h-4 mr-2" /> Reset
            </Button>
            <Button onClick={handleExecute} disabled={isRunning} className="bg-primary text-background font-mono">
              <ClipboardList className={`w-4 h-4 mr-2 ${isRunning ? 'animate-spin' : ''}`} /> Ejecutar audit
            </Button>
          </div>
        </div>

        <Card className="bg-kthulu-surface2 border-primary/20">
          <CardHeader>
            <CardTitle className="font-mono text-primary text-sm">Configuración del escaneo</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Ruta del proyecto</Label>
                <Input
                  value={request.path ?? ''}
                  onChange={(event) => setRequest((prev) => ({ ...prev, path: event.target.value }))}
                  placeholder="~/workspace/mi-servicio"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>
              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Paralelismo (jobs)</Label>
                <Input
                  type="number"
                  min={1}
                  value={request.jobs ?? 4}
                  onChange={(event) => setRequest((prev) => ({ ...prev, jobs: Number(event.target.value) }))}
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>
            </div>

            <div className="grid md:grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Tipos (kinds)</Label>
                <Input
                  value={(request.onlyKinds ?? []).join(', ')}
                  onChange={(event) => setRequest((prev) => ({ ...prev, onlyKinds: sanitizeList(event.target.value) }))}
                  placeholder="service, entity, docs"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>
              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Extensiones</Label>
                <Input
                  value={(request.extensions ?? []).join(', ')}
                  onChange={(event) => setRequest((prev) => ({ ...prev, extensions: sanitizeList(event.target.value) }))}
                  placeholder="go, ts, md"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>
              <div className="space-y-2">
                <Label className="font-mono text-xs text-primary">Ignorar rutas</Label>
                <Input
                  value={(request.ignore ?? []).join(', ')}
                  onChange={(event) => setRequest((prev) => ({ ...prev, ignore: sanitizeList(event.target.value) }))}
                  placeholder="node_modules, vendor"
                  className="bg-kthulu-surface1 border-primary/30 font-mono"
                />
              </div>
            </div>

            <div className="flex items-center justify-between border border-primary/20 rounded-sm px-3 py-2">
              <div>
                <p className="font-mono text-xs text-primary">Modo estricto</p>
                <p className="text-[11px] text-muted-foreground font-mono">
                  Reporta advertencias como errores críticos.
                </p>
              </div>
              <Switch
                checked={!!request.strict}
                onCheckedChange={(checked) => setRequest((prev) => ({ ...prev, strict: checked }))}
              />
            </div>
          </CardContent>
        </Card>

        {result && (
          <div className="grid lg:grid-cols-[320px_1fr] gap-6">
            <Card className="bg-kthulu-surface2 border-primary/20">
              <CardHeader>
                <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                  <FolderSearch className="w-4 h-4" /> Resumen del audit
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-xs font-mono">
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Ruta:</span>
                  <span className="text-primary truncate max-w-[180px]">{result.path}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Duración:</span>
                  <span className="text-primary">{result.duration}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-muted-foreground">Strict:</span>
                  <Badge variant={result.strict ? 'destructive' : 'outline'} className="text-[10px]">
                    {result.strict ? 'Sí' : 'No'}
                  </Badge>
                </div>
                <div className="space-y-2">
                  <p className="text-muted-foreground font-semibold">Totales por tipo</p>
                  <div className="space-y-1">
                    {countsEntries.length === 0 && (
                      <p className="text-muted-foreground">Sin datos de conteo.</p>
                    )}
                    {countsEntries.map(([kind, count]) => (
                      <div key={kind} className="flex items-center justify-between">
                        <span>{kind}</span>
                        <Badge variant="outline" className="text-[10px]">
                          {count}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </div>
                {result.warnings && result.warnings.length > 0 && (
                  <div className="space-y-1">
                    <p className="text-muted-foreground font-semibold">Warnings</p>
                    <ul className="list-disc list-inside space-y-1">
                      {result.warnings.map((warning, index) => (
                        <li key={index}>{warning}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className="bg-kthulu-surface2 border-primary/20">
              <CardHeader>
                <CardTitle className="font-mono text-primary text-sm flex items-center gap-2">
                  <Flame className="w-4 h-4" /> Findings detectados
                </CardTitle>
              </CardHeader>
              <CardContent>
                {result.findings.length === 0 ? (
                  <p className="text-xs text-muted-foreground font-mono">
                    Sin findings registrados. Todo está alineado con las convenciones.
                  </p>
                ) : (
                  <ScrollArea className="max-h-72">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="text-xs">Archivo</TableHead>
                          <TableHead className="text-xs">Línea</TableHead>
                          <TableHead className="text-xs">Tipo</TableHead>
                          <TableHead className="text-xs">Detalle</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {result.findings.map((finding, index) => (
                          <TableRow key={`${finding.file}-${finding.line}-${index}`}>
                            <TableCell className="text-xs font-mono">{finding.file}</TableCell>
                            <TableCell className="text-xs text-muted-foreground">{finding.line}</TableCell>
                            <TableCell className="text-xs">
                              <Badge variant="outline" className="text-[10px]">
                                {finding.kind}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-xs text-muted-foreground">{finding.detail}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </ScrollArea>
                )}
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </div>
  );
}
