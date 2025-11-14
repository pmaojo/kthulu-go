import type {
  ProjectRequest,
  ProjectPlan,
  ModuleInfo,
  ModuleValidationResult,
  ModuleInjectionPlan,
  ComponentRequest,
  TemplateInfo,
  Template,
  TemplateRenderRequest,
  TemplateRenderResult,
  TemplateSyncResult,
  TemplateDriftReport,
  AuditRequest,
  AuditResult,
  ApiError,
} from '@/types/kthulu';

const API_BASE_URL = 'http://localhost:8080';

class KthuluApiError extends Error {
  constructor(public status: number, message: string, public details?: string) {
    super(message);
    this.name = 'KthuluApiError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error: ApiError = await response.json().catch(() => ({
      error: response.statusText,
    }));
    throw new KthuluApiError(response.status, error.error, error.details);
  }
  return response.json();
}

export const kthuluApi = {
  // System
  async health() {
    const response = await fetch(`${API_BASE_URL}/health`);
    return handleResponse<{ status: string; timestamp: string }>(response);
  },

  // Projects
  async planProject(request: ProjectRequest) {
    const response = await fetch(`${API_BASE_URL}/api/v1/projects/plan`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return handleResponse<ProjectPlan>(response);
  },

  async generateProject(request: ProjectRequest) {
    const response = await fetch(`${API_BASE_URL}/api/v1/projects`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return handleResponse<ProjectPlan>(response);
  },

  // Modules
  async listModules(category?: string) {
    const url = new URL(`${API_BASE_URL}/api/v1/modules`);
    if (category) url.searchParams.set('category', category);
    const response = await fetch(url.toString());
    return handleResponse<ModuleInfo[]>(response);
  },

  async getModule(name: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/modules/${name}`);
    return handleResponse<ModuleInfo>(response);
  },

  async validateModules(modules: string[]) {
    const response = await fetch(`${API_BASE_URL}/api/v1/modules/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ modules }),
    });
    return handleResponse<ModuleValidationResult>(response);
  },

  async planModules(modules: string[]) {
    const response = await fetch(`${API_BASE_URL}/api/v1/modules/plan`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ modules }),
    });
    return handleResponse<ModuleInjectionPlan>(response);
  },

  // Components
  async generateComponent(request: ComponentRequest) {
    const response = await fetch(`${API_BASE_URL}/api/v1/components`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return handleResponse<{ status: string }>(response);
  },

  // Templates
  async listTemplates() {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates`);
    return handleResponse<TemplateInfo[]>(response);
  },

  async getTemplate(name: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/${name}`);
    return handleResponse<Template>(response);
  },

  async validateTemplate(name: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/validate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    });
    return handleResponse<{ status: string }>(response);
  },

  async renderTemplate(request: TemplateRenderRequest) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/render`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return handleResponse<TemplateRenderResult>(response);
  },

  async cacheTemplate(url: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/cache`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url }),
    });
    return handleResponse<{ status: string }>(response);
  },

  async addRegistry(name: string, url: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/registries`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, url }),
    });
    return handleResponse<{ status: string }>(response);
  },

  async removeRegistry(name: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/registries/${name}`, {
      method: 'DELETE',
    });
    return handleResponse<{ status: string }>(response);
  },

  async updateTemplates() {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/update`, {
      method: 'POST',
    });
    return handleResponse<{ status: string }>(response);
  },

  async syncRegistries() {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/sync`, {
      method: 'POST',
    });
    return handleResponse<{ status: string }>(response);
  },

  async cleanTemplates(maxAge?: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/clean`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ maxAge }),
    });
    return handleResponse<{ status: string }>(response);
  },

  async syncTemplatesFromSource(source: string) {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/sync-from`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ source }),
    });
    return handleResponse<TemplateSyncResult>(response);
  },

  async verifyTemplates() {
    const response = await fetch(`${API_BASE_URL}/api/v1/templates/verify`, {
      method: 'POST',
    });
    return handleResponse<TemplateDriftReport>(response);
  },

  // Audit
  async runAudit(request: AuditRequest = {}) {
    const response = await fetch(`${API_BASE_URL}/api/v1/audit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return handleResponse<AuditResult>(response);
  },
};

export type { KthuluApiError };
