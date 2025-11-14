// Kthulu API Types based on OpenAPI 3.1 schema

export interface ProjectRequest {
  name: string;
  modules?: string[];
  template?: string;
  database?: string;
  frontend?: string;
  skipGit?: boolean;
  skipDocker?: boolean;
  author?: string;
  license?: string;
  description?: string;
  path?: string;
  dryRun?: boolean;
}

export interface ProjectStructure {
  name: string;
  path: string;
  backend?: {
    packageName: string;
    modules: string[];
    architecture: string;
    template: string;
  };
  frontend?: {
    framework: string;
    language: string;
    modules: string[];
    template: string;
  };
  database?: {
    type: string;
    migrations: string[];
  };
  docker?: {
    enabled: boolean;
    services: string[];
  };
  config?: Record<string, any>;
  modules?: ModuleInfo[];
  author?: string;
  license?: string;
  description?: string;
}

export interface ProjectPlan {
  options: ProjectRequest;
  structure: ProjectStructure;
  modules: string[];
  projectDirectories: string[];
  backendTemplate?: string;
  backendTemplateVersion?: string;
  backendFiles?: string[];
  frontendTemplate?: string;
  frontendTemplateVersion?: string;
  frontendFiles?: string[];
  staticFiles?: string[];
  configFiles?: string[];
  migrationFiles?: string[];
  dockerServices?: string[];
}

export interface ModuleInfo {
  name: string;
  description?: string;
  version?: string;
  dependencies?: string[];
  optional?: boolean;
  category?: string;
  tags?: string[];
  entities?: any[];
  routes?: any[];
  migrations?: string[];
  frontend?: boolean;
  backend?: boolean;
  config?: Record<string, any>;
  conflicts?: string[];
  minVersion?: string;
  maxVersion?: string;
}

export interface ModuleValidationResult {
  valid: boolean;
  missing?: string[];
  circular?: { chain: string[] }[];
  conflicts?: { module: string; conflicts: string[]; reason: string }[];
  resolved?: string[];
  warnings?: string[];
}

export interface ModuleInjectionPlan {
  requested_modules: string[];
  resolved_modules: string[];
  injected_modules: string[];
  execution_order: string[];
  module_details: Record<string, ModuleInfo>;
  warnings?: string[];
  errors?: string[];
}

export interface ComponentRequest {
  type: string;
  name: string;
  module?: string;
  withTests?: boolean;
  withMigration?: boolean;
  fields?: string;
  relations?: string;
  force?: boolean;
  projectPath: string;
}

export interface TemplateInfo {
  name: string;
  version?: string;
  latest_version?: string;
  description?: string;
  author?: string;
  category?: string;
  tags?: string[];
  remote?: boolean;
  url?: string;
}

export interface Template {
  [key: string]: any;
}

export interface TemplateRenderRequest {
  name: string;
  vars?: Record<string, any>;
}

export interface TemplateRenderResult {
  files: Record<string, string>; // base64 encoded
}

export interface TemplateSyncResult {
  source: string;
  destination: string;
  filesCopied: number;
  templatesRegistered: number;
  manifestPath: string;
}

export interface TemplateDriftReport {
  added: string[];
  removed: string[];
  changed: {
    path: string;
    expectedChecksum: string;
    actualChecksum: string;
  }[];
}

export interface AuditRequest {
  path?: string;
  onlyKinds?: string[];
  extensions?: string[];
  ignore?: string[];
  strict?: boolean;
  jobs?: number;
}

export interface AuditResult {
  path: string;
  duration: string;
  counts: Record<string, number>;
  findings: {
    file: string;
    line: number;
    kind: string;
    detail: string;
  }[];
  strict: boolean;
  warnings?: string[];
}

export interface ApiError {
  error: string;
  details?: string;
}

export interface AISuggestionRequest {
  prompt: string;
  include_context?: boolean;
  project_path?: string;
  model?: string;
  provider?: string;
}

export interface AISuggestionResponse {
  result: string;
  model?: string;
  provider?: string;
  timestamp?: string;
  usage?: {
    prompt_tokens?: number;
    completion_tokens?: number;
    total_tokens?: number;
  };
}

export interface HealthResponse {
  status: string;
  timestamp: string;
  version?: string;
  uptime?: string;
  checks?: Record<string, 'healthy' | 'degraded' | 'unhealthy'>;
}

export interface SecurityConfig {
  rbac?: {
    enabled: boolean;
    default_deny_policy: boolean;
    strict_mode: boolean;
    contextual_security: boolean;
    hierarchical_roles: boolean;
    cache_enabled: boolean;
    cache_ttl: string;
    audit_enabled: boolean;
  };
  audit?: {
    enabled: boolean;
    log_level: string;
    retention_days: number;
    storage_type: string;
  };
  session?: {
    secure_cookie: boolean;
    same_site: string;
    max_age?: number;
  };
}
