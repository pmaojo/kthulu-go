import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SecurityPanel } from '@/components/SecurityPanel';
import { kthuluApi } from '@/services/kthuluApi';
import type { SecurityConfig } from '@/types/kthulu';

vi.mock('@/services/kthuluApi', () => ({
  kthuluApi: {
    getSecurityConfig: vi.fn(),
    updateSecurityConfig: vi.fn(),
  },
}));

const mockedApi = kthuluApi as {
  getSecurityConfig: ReturnType<typeof vi.fn>;
  updateSecurityConfig: ReturnType<typeof vi.fn>;
};

const baseConfig: SecurityConfig = {
  rbac: {
    enabled: true,
    default_deny_policy: true,
    strict_mode: true,
    contextual_security: false,
    hierarchical_roles: false,
    cache_enabled: false,
    cache_ttl: '10m',
    audit_enabled: false,
  },
  audit: {
    enabled: true,
    log_level: 'info',
    retention_days: 30,
    storage_type: 's3',
  },
  session: {
    secure_cookie: true,
    same_site: 'lax',
    max_age: 3600,
  },
};

describe('SecurityPanel', () => {
  beforeEach(() => {
    mockedApi.getSecurityConfig.mockReset();
    mockedApi.updateSecurityConfig.mockReset();
    mockedApi.getSecurityConfig.mockResolvedValue(baseConfig);
    mockedApi.updateSecurityConfig.mockResolvedValue(baseConfig);
  });

  it('allows toggling RBAC and saves through the API', async () => {
    const user = userEvent.setup();
    render(<SecurityPanel />);

    const rbacSwitch = await screen.findByRole('switch', { name: /rbac habilitado/i });
    await user.click(rbacSwitch);
    await user.click(screen.getByRole('button', { name: /guardar cambios/i }));

    await waitFor(() => {
      expect(mockedApi.updateSecurityConfig).toHaveBeenCalledWith(
        expect.objectContaining({ rbac: expect.objectContaining({ enabled: false }) }),
      );
    });
  });

  it('blocks invalid session max age values', async () => {
    const user = userEvent.setup();
    render(<SecurityPanel />);

    const maxAgeInput = await screen.findByLabelText(/duración máxima/i);
    fireEvent.change(maxAgeInput, { target: { value: '-5' } });
    await user.click(screen.getByRole('button', { name: /guardar cambios/i }));

    expect(mockedApi.updateSecurityConfig).not.toHaveBeenCalled();
    expect(await screen.findByText(/vida de la sesión debe ser positiva/i)).toBeInTheDocument();
  });
});
