import { useEffect, useState } from 'react';
import { Settings, X, Plus, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ElementProperties, ElementType } from '@/types/properties';

interface PropertiesPanelProps {
  isOpen: boolean;
  onClose: () => void;
  selectedElement?: ElementProperties;
  onApply?: (element: ElementProperties) => void;
  onDelete?: (elementId: string) => void;
}


export function PropertiesPanel({
  isOpen,
  onClose,
  selectedElement,
  onApply,
  onDelete,
}: PropertiesPanelProps) {
  const [selectedType, setSelectedType] = useState<ElementType>('service');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [actor, setActor] = useState('');
  const [action, setAction] = useState('');
  const [status, setStatus] = useState<ElementProperties['status']>('active');
  const [fields, setFields] = useState(['id', 'name', 'email']);

  useEffect(() => {
    if (!selectedElement) {
      setSelectedType('service');
      setName('');
      setDescription('');
      setActor('');
      setAction('');
      setStatus('active');
      setFields(['id', 'name', 'email']);
      return;
    }

    setSelectedType(selectedElement.type);
    setName(selectedElement.name);
    setDescription(selectedElement.description);
    setActor(selectedElement.actor ?? '');
    setAction(selectedElement.action ?? '');
    setStatus(selectedElement.status ?? 'active');
    setFields(
      selectedElement.fields && selectedElement.fields.length > 0
        ? [...selectedElement.fields]
        : ['id', 'name', 'email']
    );
  }, [selectedElement]);

  const isEntity = selectedType === 'entity';
  const isUseCase = selectedType === 'usecase';
  const isService = selectedType === 'service';

  const isReadOnly = !selectedElement;
  const isApplyDisabled = isReadOnly || !onApply;
  const isDeleteDisabled = isReadOnly || !onDelete;

  const addField = () => {
    if (isReadOnly) return;
    setFields([...fields, `field_${fields.length + 1}`]);
  };

  const removeField = (index: number) => {
    if (isReadOnly) return;
    setFields(fields.filter((_, i) => i !== index));
  };

  const updateField = (index: number, value: string) => {
    if (isReadOnly) return;
    const newFields = [...fields];
    newFields[index] = value;
    setFields(newFields);
  };

  const handleApply = () => {
    if (!selectedElement || !onApply) {
      return;
    }

    const updatedElement: ElementProperties = {
      ...selectedElement,
      type: selectedType,
      name,
      description,
      fields: isEntity ? fields : undefined,
      actor: isUseCase ? actor : undefined,
      action: isUseCase ? action : undefined,
      status: isService ? status ?? 'active' : undefined,
    };

    onApply(updatedElement);
  };

  const handleDelete = () => {
    if (!selectedElement || !onDelete) {
      return;
    }

    onDelete(selectedElement.id);
  };

  if (!isOpen) return null;

  return (
    <div className="w-80 bg-kthulu-surface1 border-l border-primary/20 h-full flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-primary/20 bg-kthulu-surface2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Settings className="w-4 h-4 text-primary" />
            <h2 className="font-mono font-bold text-primary">PROPIEDADES</h2>
          </div>
          <Button 
            variant="ghost" 
            size="icon"
            onClick={onClose}
            className="hover:bg-primary/10 hover:text-primary"
          >
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 p-4 space-y-6 overflow-y-auto">
        {isReadOnly && (
          <div className="rounded-md border border-dashed border-primary/30 bg-kthulu-surface2/50 p-3 text-xs font-mono text-muted-foreground">
            Selecciona un elemento del lienzo para editar sus propiedades.
          </div>
        )}

        {/* Tipo de Elemento */}
        <div className="space-y-2">
          <Label className="text-primary font-mono text-sm">TIPO</Label>
          <Select value={selectedType} onValueChange={(value) => setSelectedType(value as ElementType)} disabled={isReadOnly}>
            <SelectTrigger className="bg-kthulu-surface2 border-primary/30 text-foreground font-mono">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="bg-kthulu-surface2 border-primary/30">
              <SelectItem value="service" className="font-mono">Servicio</SelectItem>
              <SelectItem value="entity" className="font-mono">Entidad</SelectItem>
              <SelectItem value="usecase" className="font-mono">Caso de Uso</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Nombre */}
        <div className="space-y-2">
          <Label className="text-primary font-mono text-sm">NOMBRE</Label>
          <Input
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Nombre del elemento"
            className="bg-kthulu-surface2 border-primary/30 text-foreground font-mono"
            disabled={isReadOnly}
          />
        </div>

        {/* Descripción */}
        <div className="space-y-2">
          <Label className="text-primary font-mono text-sm">DESCRIPCIÓN</Label>
          <Textarea
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            placeholder="Descripción detallada..."
            className="bg-kthulu-surface2 border-primary/30 text-foreground font-mono resize-none"
            rows={3}
            disabled={isReadOnly}
          />
        </div>

        {/* Campos (solo para entidades) */}
        {isEntity && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <Label className="text-secondary font-mono text-sm">CAMPOS</Label>
              <Button
                variant="outline"
                size="sm"
                onClick={addField}
                className="bg-kthulu-surface2 border-secondary/30 hover:bg-secondary/10 hover:border-secondary"
                disabled={isReadOnly}
              >
                <Plus className="w-3 h-3" />
              </Button>
            </div>

            <div className="space-y-2 max-h-40 overflow-y-auto">
              {fields.map((field, index) => (
                <div key={index} className="flex items-center gap-2">
                  <Input
                    value={field}
                    onChange={(e) => updateField(index, e.target.value)}
                    className="bg-kthulu-surface2 border-secondary/30 text-foreground font-mono text-sm"
                    disabled={isReadOnly}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => removeField(index)}
                    className="hover:bg-destructive/10 hover:text-destructive h-8 w-8"
                    disabled={isReadOnly}
                  >
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Actor y Acción (solo para casos de uso) */}
        {isUseCase && (
          <>
            <div className="space-y-2">
              <Label className="text-accent font-mono text-sm">ACTOR</Label>
              <Input
                value={actor}
                onChange={(event) => setActor(event.target.value)}
                placeholder="Usuario, Admin, Sistema..."
                className="bg-kthulu-surface2 border-accent/30 text-foreground font-mono"
                disabled={isReadOnly}
              />
            </div>

            <div className="space-y-2">
              <Label className="text-accent font-mono text-sm">ACCIÓN</Label>
              <Textarea
                value={action}
                onChange={(event) => setAction(event.target.value)}
                placeholder="Descripción de la acción..."
                className="bg-kthulu-surface2 border-accent/30 text-foreground font-mono resize-none"
                rows={2}
                disabled={isReadOnly}
              />
            </div>
          </>
        )}

        {/* Status (solo para servicios) */}
        {isService && (
          <div className="space-y-2">
            <Label className="text-primary font-mono text-sm">STATUS</Label>
            <Select value={status ?? 'active'} onValueChange={(value) => setStatus(value as ElementProperties['status'])} disabled={isReadOnly}>
              <SelectTrigger className="bg-kthulu-surface2 border-primary/30 text-foreground font-mono">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-kthulu-surface2 border-primary/30">
                <SelectItem value="active" className="font-mono">Activo</SelectItem>
                <SelectItem value="inactive" className="font-mono">Inactivo</SelectItem>
                <SelectItem value="error" className="font-mono">Error</SelectItem>
              </SelectContent>
            </Select>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-4 border-t border-primary/20 bg-kthulu-surface2">
        <div className="flex gap-2">
          <Button
            variant="outline"
            className="flex-1 bg-kthulu-surface1 border-primary/30 hover:bg-primary/10 hover:border-primary font-mono"
            onClick={handleApply}
            disabled={isApplyDisabled}
          >
            Aplicar
          </Button>
          <Button
            variant="outline"
            className="flex-1 bg-kthulu-surface1 border-destructive/30 hover:bg-destructive/10 hover:border-destructive font-mono"
            onClick={handleDelete}
            disabled={isDeleteDisabled}
          >
            Eliminar
          </Button>
        </div>
      </div>
    </div>
  );
}
