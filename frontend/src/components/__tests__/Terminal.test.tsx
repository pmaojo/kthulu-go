import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Terminal } from '@/components/Terminal';
import { kthuluApi } from '@/services/kthuluApi';

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({ toast: vi.fn() }),
}));

vi.mock('@/services/kthuluApi', () => ({
  kthuluApi: {
    health: vi.fn(),
    listModules: vi.fn(),
    validateModules: vi.fn(),
    listTemplates: vi.fn(),
    runAudit: vi.fn(),
    runGenerateCommand: vi.fn(),
    runMigrateCommand: vi.fn(),
    runBuildCommand: vi.fn(),
    runDeployCommand: vi.fn(),
    runTestCommand: vi.fn(),
    runValidateCommand: vi.fn(),
  },
}));

const mockedApi = kthuluApi as unknown as {
  health: ReturnType<typeof vi.fn>;
  listModules: ReturnType<typeof vi.fn>;
  validateModules: ReturnType<typeof vi.fn>;
  listTemplates: ReturnType<typeof vi.fn>;
  runAudit: ReturnType<typeof vi.fn>;
  runGenerateCommand: ReturnType<typeof vi.fn>;
  runMigrateCommand: ReturnType<typeof vi.fn>;
  runBuildCommand: ReturnType<typeof vi.fn>;
  runDeployCommand: ReturnType<typeof vi.fn>;
  runTestCommand: ReturnType<typeof vi.fn>;
  runValidateCommand: ReturnType<typeof vi.fn>;
};

const cliResult = (command: string) => ({
  command,
  status: 'success',
  output: ['mock-output'],
  warnings: [],
  errors: [],
  duration: '1s',
});

beforeEach(() => {
  vi.clearAllMocks();
  mockedApi.health.mockResolvedValue({ status: 'ok', timestamp: '2024-01-01T00:00:00Z' });
  mockedApi.listModules.mockResolvedValue([]);
  mockedApi.validateModules.mockResolvedValue({ valid: true });
  mockedApi.listTemplates.mockResolvedValue([]);
  mockedApi.runAudit.mockResolvedValue({ duration: '1s', findings: [], counts: {} });
  mockedApi.runGenerateCommand.mockResolvedValue(cliResult('generate'));
  mockedApi.runMigrateCommand.mockResolvedValue(cliResult('migrate'));
  mockedApi.runBuildCommand.mockResolvedValue(cliResult('build'));
  mockedApi.runDeployCommand.mockResolvedValue(cliResult('deploy'));
  mockedApi.runTestCommand.mockResolvedValue(cliResult('test'));
  mockedApi.runValidateCommand.mockResolvedValue(cliResult('validate'));
});

describe('Terminal quick commands', () => {
  it('prints help entries sourced from the command catalog', async () => {
    const user = userEvent.setup();
    render(<Terminal />);

    const input = screen.getByPlaceholderText(/escribe comando/i);
    await user.type(input, 'help');
    await user.keyboard('{Enter}');

    expect(
      await screen.findByText((content) => content.includes('kthulu deploy') && content.includes('Despliega artefactos')),
    ).toBeInTheDocument();
    expect(
      screen.getByText((content) => content.includes('health') && content.includes('Verifica estado del API')),
    ).toBeInTheDocument();
  });

  it('executes the deploy CLI command when the quick button is used', async () => {
    const user = userEvent.setup();
    render(<Terminal />);

    const deployButton = screen.getByRole('button', {
      name: /kthulu deploy --cloud=aws --region=us-east-1/i,
    });
    await user.click(deployButton);

    await waitFor(() => {
      expect(mockedApi.runDeployCommand).toHaveBeenCalledWith({
        args: ['--cloud=aws', '--region=us-east-1'],
        options: { cloud: 'aws', region: 'us-east-1' },
      });
    });

    expect(await screen.findByText(/kthulu deploy \(success\)/i)).toBeInTheDocument();
  });
});
