import { useState } from "react";
import { Node } from "@xyflow/react";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { KthuluSidebar } from "@/components/KthuluSidebar";
import { ServiceCanvas } from "@/components/ServiceCanvas";
import { PropertiesPanel } from "@/components/PropertiesPanel";
import { Terminal } from "@/components/Terminal";
import { CodeEditor } from "@/components/CodeEditor";
import { Dashboard } from "@/components/Dashboard";
import { ProjectGeneratorDialog } from "@/components/ProjectGeneratorDialog";
import { ModuleCatalog } from "@/components/ModuleCatalog";
import { ComponentScaffolder } from "@/components/ComponentScaffolder";
import { TemplateManager } from "@/components/TemplateManager";
import { AuditWorkbench } from "@/components/AuditWorkbench";
import { AIAssistant } from "@/components/AIAssistant";
import { AIChat } from "@/components/AIChat";
import { Terminal as TerminalIcon, Layers, Code2, BarChart3, Eye, Zap, WifiOff, Wifi, Command } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useKthuluConnection } from "@/hooks/useKthuluConnection";
import { ElementProperties, ElementType } from "@/types/properties";
import CommandPalette from "@/components/CommandPalette";

const Index = () => {
  const [activeSection, setActiveSection] = useState("services");
  const [showProperties, setShowProperties] = useState(false);
  const [showGenerator, setShowGenerator] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const [selectedElement, setSelectedElement] = useState<ElementProperties | null>({
    id: "service-1",
    type: "service",
    name: "Servicio Principal",
    description: "Servicio base para la orquestación de módulos",
    status: "active",
  });
  const { isConnected, isChecking } = useKthuluConnection();

  const handleApplyProperties = (updatedElement: ElementProperties) => {
    setSelectedElement(updatedElement);
  };

  const handleDeleteElement = (elementId: string) => {
    setSelectedElement((current) => (current?.id === elementId ? null : current));
    setShowProperties(false);
  };

  const handleNodeSelect = (node: Node) => {
    const element: ElementProperties = {
      id: node.id,
      type: node.type as ElementType,
      name: node.data.label || node.data.name || '',
      description: node.data.description || '',
      fields: node.data.fields,
      actor: node.data.actor,
      action: node.data.action,
      status: node.data.status,
    };
    setSelectedElement(element);
    setShowProperties(true);
  };

  const renderMainContent = () => {
    switch (activeSection) {
      case "services":
      case "entities":
      case "architecture":
        return <ServiceCanvas className="flex-1" onNodeSelect={handleNodeSelect} />;
      
      case "terminal":
        return <Terminal />;

      case "generate":
      case "code":
        return <CodeEditor className="flex-1" />;

      case "preview":
        return <Dashboard />;

      case "modules":
        return <ModuleCatalog />;

      case "components":
        return <ComponentScaffolder />;

      case "templates":
        return <TemplateManager />;

      case "audit":
        return <AuditWorkbench />;

      case "ai":
        return (
          <div className="flex-1 bg-kthulu-surface1 p-4">
            <AIChat />
          </div>
        );
      
      case "settings":
        return (
          <div className="flex-1 bg-kthulu-surface1 flex items-center justify-center">
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-gradient-cyber rounded-lg flex items-center justify-center mx-auto">
                <BarChart3 className="w-8 h-8 text-background" />
              </div>
              <h2 className="text-xl font-mono text-accent">CONFIGURACIÓN</h2>
              <p className="text-muted-foreground font-mono">Panel de configuración próximamente...</p>
            </div>
          </div>
        );
      
      default:
        return <ServiceCanvas className="flex-1" onNodeSelect={handleNodeSelect} />;
    }
  };

  return (
    <SidebarProvider>
      <div className="min-h-screen w-full bg-background">
        {/* Header */}
        <header className="h-14 bg-kthulu-surface2 border-b border-primary/20 flex items-center px-4">
          <SidebarTrigger className="mr-4 hover:bg-primary/10 hover:text-primary" />
          <div className="flex items-center gap-3 flex-1">
            <div className="w-8 h-8 bg-gradient-neon rounded-sm flex items-center justify-center">
              <Layers className="w-4 h-4 text-background" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-primary font-mono">KTHULU UI</h1>
              <p className="text-xs text-muted-foreground font-mono">Arquitectura Visual del Engendro</p>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            {!isChecking && (
              <Badge 
                variant={isConnected ? "default" : "destructive"}
                className="font-mono text-xs"
              >
                {isConnected ? (
                  <>
                    <Wifi className="w-3 h-3 mr-1" />
                    API Conectada
                  </>
                ) : (
                  <>
                    <WifiOff className="w-3 h-3 mr-1" />
                    Sin conexión
                  </>
                )}
              </Badge>
            )}
            
            <Button
              variant="outline"
              size="sm"
              onClick={() => setCommandPaletteOpen(true)}
              className="bg-kthulu-surface1 border-secondary/30 hover:bg-secondary/10 hover:border-secondary font-mono flex items-center gap-2"
            >
              <Command className="w-4 h-4" />
              Paleta
              <span className="text-[10px] font-mono text-muted-foreground">⌘K</span>
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowProperties(!showProperties)}
              className="bg-kthulu-surface1 border-primary/30 hover:bg-primary/10 hover:border-primary font-mono"
            >
              <Layers className="w-4 h-4 mr-2" />
              Propiedades
            </Button>
            
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowGenerator(true)}
              className="bg-kthulu-surface1 border-accent/30 hover:bg-accent/10 hover:border-accent font-mono"
            >
              <Zap className="w-4 h-4 mr-2" />
              Generar
            </Button>
          </div>
        </header>

        <CommandPalette
          open={commandPaletteOpen}
          onOpenChange={setCommandPaletteOpen}
          activeSection={activeSection}
          onNavigate={setActiveSection}
          onToggleProperties={() => setShowProperties((current) => !current)}
          onOpenGenerator={() => setShowGenerator(true)}
        />

        {/* Main Layout */}
        <div className="flex h-[calc(100vh-3.5rem)] w-full">
          <KthuluSidebar 
            activeSection={activeSection} 
            onSectionChange={setActiveSection}
          />
          
          {/* Main Content */}
          <div className="flex flex-1">
            {renderMainContent()}
            
            {/* Properties Panel */}
            {showProperties && (
              <PropertiesPanel
                isOpen={showProperties}
                onClose={() => setShowProperties(false)}
                selectedElement={selectedElement ?? undefined}
                onApply={handleApplyProperties}
                onDelete={handleDeleteElement}
              />
            )}
          </div>
        </div>

        {/* Project Generator Dialog */}
        <ProjectGeneratorDialog 
          open={showGenerator}
          onOpenChange={setShowGenerator}
        />
      </div>
    </SidebarProvider>
  );
};

export default Index;
