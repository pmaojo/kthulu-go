import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AIChat } from '@/components/AIChat';
import { kthuluApi } from '@/services/kthuluApi';

vi.mock('@/services/kthuluApi', () => ({
  kthuluApi: {
    getAIProviders: vi.fn(),
    setAIProvider: vi.fn(),
    suggestAI: vi.fn(),
  },
}));

const mockedApi = kthuluApi as {
  getAIProviders: ReturnType<typeof vi.fn>;
  setAIProvider: ReturnType<typeof vi.fn>;
  suggestAI: ReturnType<typeof vi.fn>;
};

const providersResponse = {
  providers: [
    { id: 'litellm', name: 'LiteLLM', enabled: true },
    { id: 'bedrock', name: 'Bedrock', enabled: true },
  ],
};

describe('AIChat provider selector', () => {
  beforeEach(() => {
    localStorage.clear();
    mockedApi.getAIProviders.mockResolvedValue(providersResponse);
    mockedApi.setAIProvider.mockResolvedValue({ status: 'ok' });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('persists provider changes through the API', async () => {
    const user = userEvent.setup();
    render(<AIChat />);

    await user.click(screen.getByRole('button', { name: /configuración/i }));
    const providerSelect = await screen.findByLabelText(/provider/i);
    await user.selectOptions(providerSelect, 'bedrock');

    await waitFor(() => {
      expect(mockedApi.setAIProvider).toHaveBeenCalledWith('bedrock');
    });

    expect(localStorage.getItem('kthulu.ai.provider')).toBe('bedrock');
    expect(screen.getByText(/Provider: bedrock/i)).toBeInTheDocument();
  });

  it('falls back to local cache when backend rejects the change', async () => {
    mockedApi.setAIProvider.mockRejectedValueOnce(new Error('offline'));
    const user = userEvent.setup();
    render(<AIChat />);

    await user.click(screen.getByRole('button', { name: /configuración/i }));
    const providerSelect = await screen.findByLabelText(/provider/i);
    await user.selectOptions(providerSelect, 'bedrock');

    expect(await screen.findByText(/Error: offline/i)).toBeInTheDocument();
    expect(screen.getByText(/Provider: litellm/i)).toBeInTheDocument();
    expect(localStorage.getItem('kthulu.ai.provider')).toBe('bedrock');
  });
});
