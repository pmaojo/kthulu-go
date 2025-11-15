import { describe, expect, it, vi, afterEach } from 'vitest';
import { kthuluApi } from '@/services/kthuluApi';

const originalFetch = global.fetch;

describe('kthuluApi.generateComponent', () => {
  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('sends the payload and returns the parsed response', async () => {
    const responseBody = { status: 'ok', componentId: 'cmp-42' };
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(responseBody),
    });
    global.fetch = mockFetch as unknown as typeof global.fetch;

    const payload = { type: 'handler', name: 'Test', projectPath: '.' };
    const result = await kthuluApi.generateComponent(payload);

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/components', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    expect(result).toEqual(responseBody);
  });
});
