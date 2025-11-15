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

interface CommandDescriptor {
  id: string;
  command: string;
  helpLabel: string;
  description: string;
  quick?: boolean;
  includeInHelp?: boolean;
  category?: string;
}

const commandCatalog: CommandDescriptor[] = [
  { id: 'health', command: 'health', helpLabel: 'health', description: 'Verifica estado del API', quick: true },
  { id: 'modules-list', command: 'modules list', helpLabel: 'modules list', description: 'Lista módulos disponibles', quick: true },
  { id: 'modules-validate', command: 'modules validate core', helpLabel: 'modules validate', description: 'Valida módulos' },
  { id: 'templates-list', command: 'templates list', helpLabel: 'templates list', description: 'Lista templates', quick: true },
  { id: 'audit', command: 'audit', helpLabel: 'audit', description: 'Ejecuta audit del proyecto', quick: true },
  {
    id: 'kthulu-generate',
    command: 'kthulu generate --module core --with-tests',
    helpLabel: 'kthulu generate',
    description: 'Genera código y artefactos',
    quick: true,
    category: 'Generación',
  },
  {
    id: 'kthulu-migrate',
    command: 'kthulu migrate --database postgres',
    helpLabel: 'kthulu migrate',
    description: 'Ejecuta migraciones pendientes',
    quick: true,
    category: 'Base de Datos',
  },
  {
    id: 'kthulu-build',
    command: 'kthulu build',
    helpLabel: 'kthulu build',
    description: 'Compila el proyecto',
    quick: true,
    category: 'Construcción',
  },
  {
    id: 'kthulu-deploy',
    command: 'kthulu deploy --cloud=aws --region=us-east-1',
    helpLabel: 'kthulu deploy',
    description: 'Despliega artefactos',
    quick: true,
    category: 'Despliegue',
  },
  {
    id: 'kthulu-test',
    command: 'kthulu test --suite smoke',
    helpLabel: 'kthulu test',
    description: 'Corre suites de pruebas',
    quick: true,
    category: 'Testing',
  },
  {
    id: 'kthulu-validate',
    command: 'kthulu validate',
    helpLabel: 'kthulu validate',
    description: 'Revisa arquitectura y dependencias',
    quick: true,
    category: 'Validación',
  },
  { id: 'clear', command: 'clear', helpLabel: 'clear', description: 'Limpia la consola' },
  { id: 'help', command: 'help', helpLabel: 'help', description: 'Muestra esta ayuda', quick: true },
];

const helpEntries = commandCatalog.filter((entry) => entry.includeInHelp !== false);
const helpLabelPadding = helpEntries.reduce((max, entry) => Math.max(max, entry.helpLabel.length), 0) + 2;
const quickCommandEntries = commandCatalog.filter((entry) => entry.quick);
const panelCommandEntries = commandCatalog.filter((entry) => entry.category);

const initialLogEntries: LogEntry[] = [
  { time: new Date().toLocaleTimeString(), type: 'info', message: 'Terminal Kthulu inicializada' },
];

const parseCliArgs = (cliArgs: string[]) => {
  const payload: Record<string, unknown> = { args: cliArgs };
  const options: Record<string, string | boolean> = {};
  const targets: string[] = [];

  for (let i = 0; i < cliArgs.length; i += 1) {
    const token = cliArgs[i];
    if (token.startsWith('--')) {
      const withoutPrefix = token.slice(2);
      const [flag, value] = withoutPrefix.split('=');
      if (value !== undefined) {
        options[flag] = value;
      } else if (cliArgs[i + 1] && !cliArgs[i + 1].startsWith('--')) {
        options[flag] = cliArgs[i + 1];
        i += 1;
      } else {
        options[flag] = true;
      }
    } else {
      targets.push(token);
    }
  }

  if (Object.keys(options).length) payload.options = options;
  if (targets.length) payload.targets = targets;
  return payload;
};

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

  const executeCommand = async (manualCommand?: string) => {
    const commandToRun = manualCommand?.trim() ?? currentCommand.trim();
    if (!commandToRun) return;

    setIsRunning(true);
    addConsoleOutput(`$ ${commandToRun}`);
    addLog('info', `Ejecutando: ${commandToRun}`);

    try {
      const parts = commandToRun.split(' ');
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

        case 'kthulu': {
          const subCommand = args[0];
          if (!subCommand) {
            addConsoleOutput('✗ Uso: kthulu <comando> [opciones]');
            break;
          }

          const cliArgs = args.slice(1);
          const payload = parseCliArgs(cliArgs);
          let result;

          switch (subCommand) {
            case 'generate':
              result = await kthuluApi.runGenerateCommand(payload);
              break;
            case 'migrate':
              result = await kthuluApi.runMigrateCommand(payload);
              break;
            case 'build':
              result = await kthuluApi.runBuildCommand(payload);
              break;
            case 'deploy':
              result = await kthuluApi.runDeployCommand(payload);
              break;
            case 'test':
              result = await kthuluApi.runTestCommand(payload);
              break;
            case 'validate':
              result = await kthuluApi.runValidateCommand(payload);
              break;
            default:
              addConsoleOutput(`✗ Subcomando no soportado: ${subCommand}`);
              addLog('warning', `kthulu ${subCommand} no implementado`);
              result = null;
          }

          if (result) {
            addConsoleOutput(`✓ kthulu ${subCommand} (${result.status})`);
            result.output.forEach((line) => addConsoleOutput(`  ${line}`));
            result.warnings?.forEach((line) => addConsoleOutput(`  ⚠️ ${line}`));
            result.errors?.forEach((line) => addConsoleOutput(`  ✗ ${line}`));
            if (result.duration) {
              addConsoleOutput(`  Duración: ${result.duration}`);
            }
            addLog('success', `Comando kthulu ${subCommand} finalizado`);
          }
          break;
        }

        case 'help':
          addConsoleOutput('Comandos disponibles:');
          helpEntries.forEach((entry) => {
            const label = entry.helpLabel.padEnd(helpLabelPadding, ' ');
            addConsoleOutput(`  ${label}- ${entry.description}`);
          });
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
              {quickCommandEntries.map((entry) => (
                <Button
                  key={entry.id}
                  variant="outline"
                  size="sm"
                  onClick={() => executeCommand(entry.command)}
                  disabled={isRunning}
                  className="bg-kthulu-surface2 border-primary/30 hover:bg-primary/10 font-mono text-xs disabled:opacity-50"
                >
                  {entry.command}
                </Button>
              ))}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="commands" className="flex-1 p-4">
          <ScrollArea className="h-full">
            <div className="space-y-4">
              <div className="text-sm font-mono text-primary mb-4">COMANDOS DISPONIBLES:</div>
              
              {panelCommandEntries.map((item) => (
                <div key={item.id} className="p-3 bg-kthulu-surface2 border border-primary/20 rounded-sm">
                  <div className="flex items-center justify-between mb-2">
                    <code className="text-primary font-mono text-sm">{item.helpLabel}</code>
                    <Badge variant="outline" className="text-xs font-mono">{item.category}</Badge>
                  </div>
                  <p className="text-xs text-muted-foreground font-mono">{item.description}</p>
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