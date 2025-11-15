import { useEffect } from "react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";
import { cn } from "@/lib/utils";
import { Layers, Terminal, Zap, Eye, LayoutDashboard, Sparkles } from "lucide-react";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  activeSection: string;
  onNavigate: (section: string) => void;
  onToggleProperties: () => void;
  onOpenGenerator: () => void;
}

const navigationCommands = [
  {
    id: "services",
    section: "services",
    label: "Canvas de servicios",
    description: "Visualiza y conecta tus módulos principales",
    icon: <Layers className="h-4 w-4" />,
  },
  {
    id: "terminal",
    section: "terminal",
    label: "Terminal",
    description: "Consulta logs y lanza comandos directamente",
    icon: <Terminal className="h-4 w-4" />,
  },
  {
    id: "generate",
    section: "generate",
    label: "Generador",
    description: "Activa el diálogo de scaffolding con IA",
    icon: <Zap className="h-4 w-4" />,
  },
  {
    id: "preview",
    section: "preview",
    label: "Panel de métricas",
    description: "Revisa dashboards, auditoría y trazabilidad",
    icon: <LayoutDashboard className="h-4 w-4" />,
  },
  {
    id: "modules",
    section: "modules",
    label: "Catálogo de módulos",
    description: "Explora módulos reutilizables y plantillas",
    icon: <Sparkles className="h-4 w-4" />,
  },
];

const quickActions = [
  {
    id: "toggle-properties",
    label: "Alternar panel de propiedades",
    description: "Muestra u oculta el panel lateral contextual",
    icon: <Eye className="h-4 w-4" />,
    shortcut: "⌘P",
    action: "properties",
  },
  {
    id: "open-generator",
    label: "Abrir generador de proyectos",
    description: "Despliega el asistente de scaffolding",
    icon: <Zap className="h-4 w-4" />,
    shortcut: "⌘G",
    action: "generator",
  },
];

export function CommandPalette({
  open,
  onOpenChange,
  activeSection,
  onNavigate,
  onToggleProperties,
  onOpenGenerator,
}: CommandPaletteProps) {
  useEffect(() => {
    const handleShortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        onOpenChange(!open);
      }
    };

    window.addEventListener("keydown", handleShortcut);
    return () => window.removeEventListener("keydown", handleShortcut);
  }, [open, onOpenChange]);

  const handleNavigation = (section: string) => {
    onNavigate(section);
    onOpenChange(false);
  };

  const handleAction = (action: string) => {
    if (action === "properties") {
      onToggleProperties();
    }
    if (action === "generator") {
      onOpenGenerator();
    }
    onOpenChange(false);
  };

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange}>
      <CommandInput placeholder="Buscar comando… (⌘K)" />
      <CommandList>
        <CommandEmpty>No se encontraron comandos.</CommandEmpty>
        <CommandGroup heading="Navegar">
          {navigationCommands.map((command) => {
            const isActive = activeSection === command.section;
            return (
              <CommandItem
                key={command.id}
                value={command.label}
                onSelect={() => handleNavigation(command.section)}
                className={cn(
                  "flex flex-col gap-1 rounded-md px-2 py-2 text-sm leading-tight transition",
                  isActive
                    ? "bg-kthulu-surface2 border border-accent/50 text-accent-foreground"
                    : "hover:bg-kthulu-surface1"
                )}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="flex items-center gap-2">
                    {command.icon}
                    {command.label}
                  </span>
                  {isActive && (
                    <CommandShortcut>Activo</CommandShortcut>
                  )}
                </div>
                <p className="text-xs text-muted-foreground">
                  {command.description}
                </p>
              </CommandItem>
            );
          })}
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Acciones rápidas">
          {quickActions.map((action) => (
            <CommandItem
              key={action.id}
              value={action.label}
              onSelect={() => handleAction(action.action)}
              className="flex flex-col gap-1 rounded-md px-2 py-2 text-sm leading-tight transition hover:bg-kthulu-surface1"
            >
              <div className="flex items-center justify-between gap-2">
                <span className="flex items-center gap-2">
                  {action.icon}
                  {action.label}
                </span>
                <CommandShortcut>{action.shortcut}</CommandShortcut>
              </div>
              <p className="text-xs text-muted-foreground">{action.description}</p>
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  );
}

export default CommandPalette;
