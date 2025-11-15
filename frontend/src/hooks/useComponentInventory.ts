import { useCallback, useState } from 'react';
import { kthuluApi } from '@/services/kthuluApi';

export interface ComponentSummary {
  id: string;
  name: string;
  type: string;
}

export interface ComponentDetail extends ComponentSummary {
  config: Record<string, unknown>;
}

export function useComponentInventory() {
  const [components, setComponents] = useState<ComponentSummary[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [details, setDetails] = useState<Record<string, ComponentDetail>>({});
  const [loading, setLoading] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadComponents = useCallback(async () => {
    setLoading(true);
    try {
      const list = await kthuluApi.listComponents();
      setComponents(list);
      setError(null);
      if (list.length) {
        setSelectedId((current) => current ?? list[0].id);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'No se pudo cargar el inventario de componentes.';
      setError(message);
    } finally {
      setLoading(false);
    }
  }, []);

  const selectComponent = useCallback(async (id: string) => {
    setSelectedId(id);
    if (details[id]) {
      return details[id];
    }
    setDetailLoading(true);
    try {
      const detail = await kthuluApi.getComponent(id);
      setDetails((previous) => ({ ...previous, [id]: detail }));
      setError(null);
      return detail;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'No se pudo obtener el componente seleccionado.';
      setError(message);
      throw err;
    } finally {
      setDetailLoading(false);
    }
  }, [details]);

  const updateComponent = useCallback(async (config: Record<string, unknown>) => {
    if (!selectedId) {
      throw new Error('Selecciona un componente antes de actualizar.');
    }
    const previousDetail = details[selectedId];
    const optimisticDetail = previousDetail ? { ...previousDetail, config } : undefined;
    if (optimisticDetail) {
      setDetails((previous) => ({ ...previous, [selectedId]: optimisticDetail }));
    }
    try {
      await kthuluApi.updateComponent(selectedId, config);
      return optimisticDetail;
    } catch (err) {
      if (previousDetail) {
        setDetails((previous) => ({ ...previous, [selectedId]: previousDetail }));
      }
      throw err;
    }
  }, [details, selectedId]);

  const deleteComponent = useCallback(async () => {
    if (!selectedId) {
      throw new Error('No hay componente seleccionado.');
    }
    const previousList = components;
    const remaining = components.filter((component) => component.id !== selectedId);
    setComponents(remaining);
    try {
      await kthuluApi.deleteComponent(selectedId);
      setSelectedId(remaining[0]?.id ?? null);
      return true;
    } catch (err) {
      setComponents(previousList);
      throw err;
    }
  }, [components, selectedId]);

  const selectedComponent = selectedId ? details[selectedId] : undefined;

  return {
    components,
    selectedComponent,
    selectedId,
    loading,
    detailLoading,
    error,
    loadComponents,
    selectComponent,
    updateComponent,
    deleteComponent,
    setSelectedId,
  };
}
