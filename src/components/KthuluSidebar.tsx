import {
  Box,
  Database,
  GitBranch,
  Settings,
  Terminal,
  Zap,
  Plus,
  Layers,
  Code,
  Eye,
  PackageSearch,
  Puzzle,
  FileStack,
  ShieldCheck,
  Sparkles
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";

import { Button } from "@/components/ui/button";

const mainItems = [
  { title: "Servicios", icon: Box, id: "services" },
  { title: "Entidades", icon: Database, id: "entities" },
  { title: "Arquitectura", icon: GitBranch, id: "architecture" },
  { title: "Terminal", icon: Terminal, id: "terminal" },
];

const toolsItems = [
  { title: "Catálogo", icon: PackageSearch, id: "modules" },
  { title: "Componentes", icon: Puzzle, id: "components" },
  { title: "Templates", icon: FileStack, id: "templates" },
  { title: "Auditoría", icon: ShieldCheck, id: "audit" },
  { title: "IA Asistente", icon: Sparkles, id: "ai" },
  { title: "Generar", icon: Zap, id: "generate" },
  { title: "Vista Previa", icon: Eye, id: "preview" },
  { title: "Código", icon: Code, id: "code" },
  { title: "Configuración", icon: Settings, id: "settings" },
];

interface KthuluSidebarProps {
  activeSection: string;
  onSectionChange: (section: string) => void;
}

export function KthuluSidebar({ activeSection, onSectionChange }: KthuluSidebarProps) {
  const { state } = useSidebar();
  const collapsed = state === "collapsed";

  const isActive = (id: string) => activeSection === id;

  return (
    <Sidebar className={collapsed ? "w-14" : "w-64"} collapsible="icon">
      <SidebarContent className="bg-kthulu-surface1 border-r border-primary/20">
        {/* Header */}
        <div className="p-4 border-b border-primary/20">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-gradient-neon rounded-sm flex items-center justify-center">
              <Layers className="w-4 h-4 text-background" />
            </div>
            {!collapsed && (
              <div>
                <h1 className="text-lg font-bold text-primary">KTHULU</h1>
                <p className="text-xs text-muted-foreground">Arquitectura Visual</p>
              </div>
            )}
          </div>
        </div>

        {/* Nuevo Servicio */}
        <div className="p-4">
          <Button 
            variant="outline" 
            className="w-full bg-kthulu-surface2 border-primary/30 hover:bg-primary/10 hover:border-primary kthulu-transition"
            size={collapsed ? "icon" : "default"}
          >
            <Plus className="w-4 h-4" />
            {!collapsed && <span className="ml-2">Nuevo Servicio</span>}
          </Button>
        </div>

        {/* Main Navigation */}
        <SidebarGroup>
          <SidebarGroupLabel className="text-primary/70 font-mono">NAVEGACIÓN</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainItems.map((item) => (
                <SidebarMenuItem key={item.id}>
                  <SidebarMenuButton
                    onClick={() => onSectionChange(item.id)}
                    className={`
                      kthulu-transition font-mono
                      ${isActive(item.id) 
                        ? 'bg-primary/20 text-primary border-r-2 border-primary' 
                        : 'hover:bg-kthulu-surface2 hover:text-primary'
                      }
                    `}
                  >
                    <item.icon className="w-4 h-4" />
                    {!collapsed && <span>{item.title}</span>}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {/* Tools */}
        <SidebarGroup>
          <SidebarGroupLabel className="text-accent/70 font-mono">HERRAMIENTAS</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {toolsItems.map((item) => (
                <SidebarMenuItem key={item.id}>
                  <SidebarMenuButton
                    onClick={() => onSectionChange(item.id)}
                    className={`
                      kthulu-transition font-mono
                      ${isActive(item.id) 
                        ? 'bg-accent/20 text-accent border-r-2 border-accent' 
                        : 'hover:bg-kthulu-surface2 hover:text-accent'
                      }
                    `}
                  >
                    <item.icon className="w-4 h-4" />
                    {!collapsed && <span>{item.title}</span>}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {/* Status Footer */}
        <div className="mt-auto p-4 border-t border-primary/20">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-primary rounded-full animate-glow"></div>
            {!collapsed && (
              <span className="text-xs text-primary font-mono">Sistema Activo</span>
            )}
          </div>
        </div>
      </SidebarContent>
    </Sidebar>
  );
}