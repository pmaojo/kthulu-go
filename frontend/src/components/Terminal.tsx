import { useState, useRef, useEffect } from 'react';
import { Play, Square, RotateCcw, Download, Upload, FileCode, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { kthuluApi } from '@/services/kthuluApi';
import { useToast } from '@/hooks/use-toast';

interface LogEntry {
  time: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
}

const initialLogEntries: LogEntry[] = [
  { time: new Date().toLocaleTimeString(), type: 'info', message: 'Terminal Kthulu inicializada' },
];

export function Terminal() {
  const [currentCommand, setCurrentCommand] = useState('');
  const [isRunning, setIsRunning] = useState(false);
  const [logEntries, setLogEntries] = useState<LogEntry[]>(initialLogEntries);
  const [consoleOutput, setConsoleOutput] = useState<string[]>([
    '$ kthulu --version',
    'Kthulu CLI v2.1.0 - Arquitectura Brutalista',
  ]);
  const scrollRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [consoleOutput, logEntries]);

  const addLog = (type: LogEntry['type'], message: string) => {
    setLogEntries(prev => [...prev, {
      time: new Date().toLocaleTimeString(),
      type,
      message,
    }]);
  };

  const addConsoleOutput = (text: string) => {
    setConsoleOutput(prev => [...prev, text]);
  };

  const executeCommand = async () => {
    if (!currentCommand.trim()) return;

    setIsRunning(true);
    addConsoleOutput(`$ ${currentCommand}`);
    addLog('info', `Ejecutando: ${currentCommand}`);

    try {
      const parts = currentCommand.trim().split(' ');
      const command = parts[0];
      const args = parts.slice(1);

      switch (command) {
        case 'health':
          const health = await kthuluApi.health();
          addConsoleOutput(`✓ API Status: ${health.status}`);
          addConsoleOutput(`  Timestamp: ${health.timestamp}`);
          addLog('success', 'Health check exitoso');
          break;

        case 'modules':
          if (args[0] === 'list') {
            const modules = await kthuluApi.listModules();
            addConsoleOutput(`✓ ${modules.length} módulos disponibles:`);
            modules.forEach(m => {
              addConsoleOutput(`  - ${m.name} (${m.version || 'latest'}): ${m.description || ''}`);
            });
            addLog('success', `${modules.length} módulos listados`);
          } else if (args[0] === 'validate' && args[1]) {
            const moduleList = args.slice(1);
            const result = await kthuluApi.validateModules(moduleList);
            addConsoleOutput(`✓ Validación: ${result.valid ? 'VÁLIDA' : 'INVÁLIDA'}`);
            if (result.missing?.length) {
              addConsoleOutput(`  Faltantes: ${result.missing.join(', ')}`);
            }
            if (result.conflicts?.length) {
              addConsoleOutput(`  Conflictos: ${result.conflicts.length}`);
            }
            addLog(result.valid ? 'success' : 'error', 'Validación completada');
          }
          break;

        case 'templates':
          if (args[0] === 'list') {
            const templates = await kthuluApi.listTemplates();
            addConsoleOutput(`✓ ${templates.length} templates disponibles:`);
            templates.forEach(t => {
              addConsoleOutput(`  - ${t.name} (${t.version || 'latest'}): ${t.description || ''}`);
            });
            addLog('success', `${templates.length} templates listados`);
          }
          break;

        case 'audit':
          const audit = await kthuluApi.runAudit({});
          addConsoleOutput(`✓ Audit completado en ${audit.duration}`);
          addConsoleOutput(`  Archivos escaneados: ${audit.findings.length}`);
          Object.entries(audit.counts).forEach(([kind, count]) => {
            addConsoleOutput(`  - ${kind}: ${count}`);
          });
          addLog('success', 'Audit completado');
          break;

        case 'clear':
          setConsoleOutput([]);
          addLog('info', 'Terminal limpiada');
          break;

        case 'help':
          addConsoleOutput('Comandos disponibles:');
          addConsoleOutput('  health              - Verifica estado del API');
          addConsoleOutput('  modules list        - Lista módulos disponibles');
          addConsoleOutput('  modules validate    - Valida módulos');
          addConsoleOutput('  templates list      - Lista templates');
          addConsoleOutput('  audit              - Ejecuta audit del proyecto');
          addConsoleOutput('  clear              - Limpia la consola');
          addConsoleOutput('  help               - Muestra esta ayuda');
          break;

        default:
          addConsoleOutput(`✗ Comando no reconocido: ${command}`);
          addConsoleOutput('  Escribe "help" para ver comandos disponibles');
          addLog('error', `Comando no reconocido: ${command}`);
      }
    } catch (error: any) {
      const errorMsg = error.message || 'Error desconocido';
      addConsoleOutput(`✗ Error: ${errorMsg}`);
      addLog('error', errorMsg);
      
      toast({
        title: 'Error ejecutando comando',
        description: 'Verifica que el servidor Kthulu esté corriendo en localhost:8080',
        variant: 'destructive',
      });
    } finally {
      setIsRunning(false);
      setCurrentCommand('');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !isRunning) {
      executeCommand();
    }
  };

  const clearTerminal = () => {
    setConsoleOutput([]);
    setLogEntries(initialLogEntries);
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'error': return 'text-destructive';
      case 'warning': return 'text-accent';
      case 'success': return 'text-primary';
      default: return 'text-muted-foreground';
    }
  };

  const getTypeBadge = (type: string) => {
    switch (type) {
      case 'error': return 'destructive';
      case 'warning': return 'secondary';
      case 'success': return 'default';
      default: return 'outline';
    }
  };

  return (
    <div className="h-full bg-kthulu-surface1 flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-primary/20 bg-kthulu-surface2">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-mono font-bold text-primary text-lg">KTHULU TERMINAL</h2>
            <p className="text-xs text-muted-foreground font-mono">Control directo del CLI</p>
          </div>
          
          <div className="flex items-center gap-2">
            <Button 
              variant="outline" 
              size="sm"
              onClick={clearTerminal}
              className="bg-kthulu-surface1 border-accent/30 hover:bg-accent/10 font-mono"
            >
              <RotateCcw className="w-3 h-3" />
              Clear
            </Button>
          </div>
        </div>
      </div>

      <Tabs defaultValue="console" className="flex-1 flex flex-col">
        <div className="px-4 pt-2">
          <TabsList className="bg-kthulu-surface2 border border-primary/20">
            <TabsTrigger value="console" className="font-mono text-xs">Console</TabsTrigger>
            <TabsTrigger value="commands" className="font-mono text-xs">Comandos</TabsTrigger>
            <TabsTrigger value="logs" className="font-mono text-xs">Logs</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="console" className="flex-1 p-4 space-y-4">
          {/* Console Output */}
          <div 
            ref={scrollRef}
            className="bg-black/50 border border-primary/20 rounded-sm p-4 h-64 font-mono text-sm overflow-y-auto"
          >
            <div className="space-y-1">
              {consoleOutput.map((line, index) => (
                <div 
                  key={index}
                  className={line.startsWith('$') ? 'text-primary' : 'text-muted-foreground'}
                >
                  {line}
                </div>
              ))}
              {isRunning && (
                <div className="flex items-center gap-2 text-accent">
                  <Loader2 className="w-3 h-3 animate-spin" />
                  <span>Ejecutando...</span>
                </div>
              )}
            </div>
          </div>

          {/* Command Input */}
          <div className="flex gap-2">
            <div className="flex-1 bg-kthulu-surface2 border border-primary/30 rounded-sm p-3 font-mono text-sm flex items-center">
              <span className="text-primary mr-2">$</span>
              <input 
                type="text"
                value={currentCommand}
                onChange={(e) => setCurrentCommand(e.target.value)}
                onKeyPress={handleKeyPress}
                disabled={isRunning}
                placeholder="Escribe comando (ej: help, modules list, audit)..."
                className="flex-1 bg-transparent border-none outline-none text-foreground disabled:opacity-50"
              />
            </div>
            <Button 
              onClick={executeCommand}
              disabled={isRunning || !currentCommand.trim()}
              variant="outline"
              size="sm" 
              className="bg-primary text-primary-foreground hover:bg-primary/90 font-mono px-6 disabled:opacity-50"
            >
              {isRunning ? <Loader2 className="w-3 h-3 animate-spin" /> : 'Ejecutar'}
            </Button>
          </div>

          {/* Quick Commands */}
          <div className="space-y-2">
            <div className="text-xs text-muted-foreground font-mono">COMANDOS RÁPIDOS:</div>
            <div className="flex flex-wrap gap-2">
              {['health', 'modules list', 'templates list', 'audit', 'help'].map((cmd, index) => (
                <Button
                  key={index}
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentCommand(cmd)}
                  disabled={isRunning}
                  className="bg-kthulu-surface2 border-primary/30 hover:bg-primary/10 font-mono text-xs disabled:opacity-50"
                >
                  {cmd}
                </Button>
              ))}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="commands" className="flex-1 p-4">
          <ScrollArea className="h-full">
            <div className="space-y-4">
              <div className="text-sm font-mono text-primary mb-4">COMANDOS DISPONIBLES:</div>
              
              {[
                { cmd: 'kthulu generate', desc: 'Genera código para servicios, entidades o casos de uso', category: 'Generación' },
                { cmd: 'kthulu migrate', desc: 'Ejecuta migraciones de base de datos', category: 'Base de Datos' },
                { cmd: 'kthulu build', desc: 'Compila el proyecto completo', category: 'Construcción' },
                { cmd: 'kthulu deploy', desc: 'Despliega a plataforma especificada', category: 'Despliegue' },
                { cmd: 'kthulu test', desc: 'Ejecuta tests unitarios y de integración', category: 'Testing' },
                { cmd: 'kthulu validate', desc: 'Valida arquitectura y dependencias', category: 'Validación' },
              ].map((item, index) => (
                <div key={index} className="p-3 bg-kthulu-surface2 border border-primary/20 rounded-sm">
                  <div className="flex items-center justify-between mb-2">
                    <code className="text-primary font-mono text-sm">{item.cmd}</code>
                    <Badge variant="outline" className="text-xs font-mono">{item.category}</Badge>
                  </div>
                  <p className="text-xs text-muted-foreground font-mono">{item.desc}</p>
                </div>
              ))}
            </div>
          </ScrollArea>
        </TabsContent>

        <TabsContent value="logs" className="flex-1 p-4">
          <ScrollArea className="h-full">
            <div className="space-y-2">
              {logEntries.map((entry, index) => (
                <div key={index} className="flex items-start gap-3 p-2 hover:bg-kthulu-surface2 rounded-sm">
                  <Badge variant={getTypeBadge(entry.type)} className="text-xs font-mono min-w-fit">
                    {entry.type.toUpperCase()}
                  </Badge>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-muted-foreground font-mono">{entry.time}</span>
                      <span className={`text-sm font-mono ${getTypeColor(entry.type)}`}>
                        {entry.message}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </div>
  );
}