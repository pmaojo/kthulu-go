import { useState, useEffect } from 'react';
import { kthuluApi } from '@/services/kthuluApi';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import { Play, FileText, FlaskConical, CheckCircle, XCircle, Clock } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

interface FeatureFile {
  path: string;
  name: string;
  scenarios: string[];
}

export function BehaviorLab() {
  const [features, setFeatures] = useState<FeatureFile[]>([]);
  const [selectedFeature, setSelectedFeature] = useState<FeatureFile | null>(null);
  const [testResults, setTestResults] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    loadFeatures();
  }, []);

  const loadFeatures = async () => {
    try {
      setIsLoading(true);
      const result = await kthuluApi.listFeatures();

      if (result.output && result.output.length > 0) {
        // Parsing logic based on CLI output structure
        // Assuming line-separated paths for now
        const rawOutput = result.output.join('\n');
        // Simple heuristic: split by newlines and look for .feature files
        const paths = rawOutput.split('\n')
            .map(s => s.trim())
            .filter(s => s.endsWith('.feature'));

        const featureList = paths.map(p => ({
          path: p,
          name: p.split('/').pop() || p,
          scenarios: ['Scenario 1', 'Scenario 2'] // detailed parsing would go here
        }));
        setFeatures(featureList);
      }
    } catch (e) {
      console.error("Failed to load features", e);
      toast({
        title: "Error",
        description: "No se pudieron cargar los archivos de features.",
        variant: "destructive",
      });
      // Fallback mock data for demo purposes if backend fails
      setFeatures([
        { path: 'features/auth.feature', name: 'auth.feature', scenarios: ['Login', 'Logout'] },
        { path: 'features/products.feature', name: 'products.feature', scenarios: ['List Products', 'Get Product'] },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  const runTests = async (filter: string = '') => {
    setIsRunning(true);
    setTestResults('Ejecutando tests...');
    try {
      const result = await kthuluApi.runScenario(filter);
      setTestResults(result.output.join('\n'));

      if (result.status === 'success' || (result.metadata?.exitCode === 0)) {
        toast({
          title: "Tests Pasaron",
          description: "La ejecución de features fue exitosa.",
          variant: "default",
          className: "bg-green-500/10 border-green-500/50 text-green-500",
        });
      } else {
        toast({
          title: "Tests Fallaron",
          description: "Hubo errores en la ejecución de features.",
          variant: "destructive",
        });
      }
    } catch (e) {
      setTestResults('Error al ejecutar tests.');
      toast({
        title: "Error",
        description: "Fallo al invocar el comando de tests.",
        variant: "destructive",
      });
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <div className="h-full flex flex-col bg-kthulu-surface1 p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-mono font-bold text-primary text-2xl flex items-center gap-2">
            <FlaskConical className="w-6 h-6" />
            LABORATORIO DE COMPORTAMIENTO
          </h2>
          <p className="text-muted-foreground font-mono text-sm">Gestión y ejecución de tests BDD (Gherkin)</p>
        </div>
        <Button
          variant="outline"
          onClick={loadFeatures}
          disabled={isLoading}
          className="font-mono"
        >
          Refrescar Lista
        </Button>
      </div>

      <ResizablePanelGroup direction="horizontal" className="flex-1 rounded-lg border border-primary/20 bg-kthulu-surface2 overflow-hidden">

        {/* Sidebar List */}
        <ResizablePanel defaultSize={25} minSize={20} maxSize={40}>
          <div className="h-full flex flex-col">
            <div className="p-4 border-b border-primary/10">
              <h3 className="font-mono text-sm text-muted-foreground font-semibold">FEATURES DISPONIBLES</h3>
            </div>
            <ScrollArea className="flex-1 p-2">
              <div className="space-y-1">
                {features.map((feature) => (
                  <Button
                    key={feature.path}
                    variant={selectedFeature?.path === feature.path ? "secondary" : "ghost"}
                    className={`w-full justify-start font-mono text-sm ${
                        selectedFeature?.path === feature.path
                        ? 'bg-primary/10 text-primary hover:bg-primary/20'
                        : 'text-muted-foreground hover:text-foreground'
                    }`}
                    onClick={() => setSelectedFeature(feature)}
                  >
                    <FileText className="w-4 h-4 mr-2" />
                    {feature.name}
                  </Button>
                ))}
                {features.length === 0 && !isLoading && (
                    <div className="text-center p-4 text-xs text-muted-foreground font-mono">
                        No se encontraron features.
                    </div>
                )}
              </div>
            </ScrollArea>
          </div>
        </ResizablePanel>

        <ResizableHandle className="bg-primary/10" />

        {/* Main Content */}
        <ResizablePanel defaultSize={75}>
          {selectedFeature ? (
            <div className="h-full flex flex-col p-4 space-y-4">
              {/* Feature Header */}
              <div className="flex items-center justify-between bg-kthulu-surface1 p-4 rounded-md border border-primary/10">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-sm">
                        <FileText className="w-5 h-5 text-primary" />
                    </div>
                    <div>
                        <h3 className="font-mono font-bold text-lg text-foreground">{selectedFeature.name}</h3>
                        <p className="font-mono text-xs text-muted-foreground">{selectedFeature.path}</p>
                    </div>
                </div>
                <Button
                    onClick={() => runTests(selectedFeature.path)}
                    disabled={isRunning}
                    className="font-mono bg-primary/10 text-primary hover:bg-primary/20 border border-primary/50"
                >
                    {isRunning ? <Clock className="w-4 h-4 mr-2 animate-spin" /> : <Play className="w-4 h-4 mr-2" />}
                    Ejecutar Feature
                </Button>
              </div>

              {/* Editor & Output Split */}
              <ResizablePanelGroup direction="vertical" className="flex-1 rounded-md border border-primary/10">

                {/* Feature Content Viewer */}
                <ResizablePanel defaultSize={60}>
                    <div className="h-full flex flex-col bg-kthulu-surface1">
                        <div className="p-2 border-b border-primary/10 flex items-center justify-between">
                            <span className="font-mono text-xs text-muted-foreground ml-2">CONTENIDO DEL ARCHIVO</span>
                            <Badge variant="outline" className="font-mono text-[10px]">READ ONLY</Badge>
                        </div>
                        <ScrollArea className="flex-1 p-4">
                            <pre className="font-mono text-sm text-foreground/80 leading-relaxed whitespace-pre-wrap">
{`Feature: ${selectedFeature.name}

  Scenario: Basic Interaction
    Given the system is running
    When the user requests "${selectedFeature.name}"
    Then the response should be successful

  Scenario: Error Handling
    Given the system is unavailable
    When the user requests "${selectedFeature.name}"
    Then the system should return an error
`}
{/* TODO: Implement real file reading via API */}
                            </pre>
                        </ScrollArea>
                    </div>
                </ResizablePanel>

                <ResizableHandle className="bg-primary/10" />

                {/* Test Output Console */}
                <ResizablePanel defaultSize={40}>
                    <div className="h-full flex flex-col bg-black/40">
                         <div className="p-2 border-b border-primary/10 flex items-center gap-2">
                            <TerminalIcon className="w-3 h-3 text-muted-foreground ml-2" />
                            <span className="font-mono text-xs text-muted-foreground">SALIDA DE TESTS</span>
                        </div>
                        <ScrollArea className="flex-1 p-4">
                            <pre className={`font-mono text-xs leading-relaxed ${
                                testResults.includes('FAIL') ? 'text-red-400' : 'text-green-400'
                            }`}>
                                {testResults || <span className="text-muted-foreground/50">Esperando ejecución...</span>}
                            </pre>
                        </ScrollArea>
                    </div>
                </ResizablePanel>

              </ResizablePanelGroup>
            </div>
          ) : (
            <div className="h-full flex flex-col items-center justify-center text-muted-foreground space-y-4">
                <FlaskConical className="w-16 h-16 opacity-20" />
                <p className="font-mono text-sm">Selecciona un archivo .feature para comenzar</p>
            </div>
          )}
        </ResizablePanel>

      </ResizablePanelGroup>
    </div>
  );
}

function TerminalIcon(props: React.SVGProps<SVGSVGElement>) {
    return (
        <svg
          {...props}
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="4 17 10 11 4 5" />
          <line x1="12" x2="20" y1="19" y2="19" />
        </svg>
      )
}
