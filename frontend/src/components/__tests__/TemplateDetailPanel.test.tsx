import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { TemplateDetailPanel } from '@/components/TemplateManager';
import type { Template } from '@/types/kthulu';

const detail: Template = {
  name: 'hexagonal-go',
  content: { 'main.go': 'package main' },
  variables: {
    service: { description: 'Nombre del servicio', default: 'billing' },
  },
  updatedAt: '2024-06-01T00:00:00Z',
  metadata: { language: 'go' },
};

describe('TemplateDetailPanel', () => {
  it('renders metadata and variables for the selected template', () => {
    render(<TemplateDetailPanel detail={detail} isLoading={false} />);

    expect(screen.getByText(/Archivos:/i)).toBeInTheDocument();
    expect(screen.getByText(/Default:\s+billing/i)).toBeInTheDocument();
    expect(screen.getByText(/language/i)).toBeInTheDocument();
  });

  it('renders loading and error states', () => {
    const { rerender } = render(<TemplateDetailPanel isLoading detail={undefined} />);
    expect(screen.getByText(/Obteniendo detalles/i)).toBeInTheDocument();

    rerender(<TemplateDetailPanel isLoading={false} errorMessage="falló" />);
    expect(screen.getByText('falló')).toBeInTheDocument();
  });
});
