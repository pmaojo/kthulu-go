import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { useComponentInventory } from '@/hooks/useComponentInventory';
import { kthuluApi } from '@/services/kthuluApi';

describe('useComponentInventory', () => {
  let updateSpy: ReturnType<typeof vi.spyOn>;
  let deleteSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.spyOn(kthuluApi, 'listComponents').mockResolvedValue([
      { id: 'cmp-1', name: 'UserHandler', type: 'handler' },
    ]);
    vi.spyOn(kthuluApi, 'getComponent').mockResolvedValue({
      id: 'cmp-1',
      name: 'UserHandler',
      type: 'handler',
      config: { retries: 1 },
    });
    updateSpy = vi.spyOn(kthuluApi, 'updateComponent').mockResolvedValue({ status: 'ok' });
    deleteSpy = vi.spyOn(kthuluApi, 'deleteComponent').mockResolvedValue({ status: 'ok' });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('loads component inventory and fetches detail on demand', async () => {
    const { result } = renderHook(() => useComponentInventory());

    await act(async () => {
      await result.current.loadComponents();
    });
    expect(result.current.components).toHaveLength(1);

    await act(async () => {
      await result.current.selectComponent('cmp-1');
    });

    expect(kthuluApi.getComponent).toHaveBeenCalledWith('cmp-1');
    expect(result.current.selectedComponent?.config).toEqual({ retries: 1 });
  });

  it('performs optimistic updates and rolls back on failure', async () => {
    const { result } = renderHook(() => useComponentInventory());
    await act(async () => {
      await result.current.loadComponents();
    });
    await act(async () => {
      await result.current.selectComponent('cmp-1');
    });

    updateSpy.mockRejectedValueOnce(new Error('boom'));

    await expect(
      act(async () => {
        await result.current.updateComponent({ retries: 5 });
      }),
    ).rejects.toThrow();

    expect(result.current.selectedComponent?.config).toEqual({ retries: 1 });
  });

  it('removes components optimistically and keeps selection consistent', async () => {
    const { result } = renderHook(() => useComponentInventory());
    await act(async () => {
      await result.current.loadComponents();
    });
    await act(async () => {
      await result.current.selectComponent('cmp-1');
    });

    await act(async () => {
      await result.current.deleteComponent();
    });

    expect(deleteSpy).toHaveBeenCalledWith('cmp-1');
    expect(result.current.components).toHaveLength(0);
    expect(result.current.selectedComponent).toBeUndefined();
  });
});
