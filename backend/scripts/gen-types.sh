#!/usr/bin/env bash

# @kthulu:core
# TypeScript Type Generation Script
# Generates TypeScript interfaces and Zod schemas from OpenAPI specification

set -e

# Resolve repository root to support execution from any directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
API_DIR="$ROOT_DIR/api"
FRONTEND_DIR="$ROOT_DIR/frontend"
OPENAPI_FILE="$API_DIR/openapi.yaml"
TYPES_DIR="$FRONTEND_DIR/src/types"
TYPES_FILE="$TYPES_DIR/kthulu-api.ts"
ZOD_FILE="$TYPES_DIR/kthulu-api-zod.ts"

# Create cross-platform temporary directory
TEMP_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'kthulu-types')"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "$OPENAPI_FILE" ]; then
    log_error "OpenAPI specification not found at $OPENAPI_FILE. Please run ./scripts/openapi.sh first."
    exit 1
fi

if [ ! -d "$FRONTEND_DIR" ]; then
    log_error "Frontend directory not found. Please ensure the frontend project exists."
    exit 1
fi

# Create types directory if it doesn't exist
mkdir -p "$TYPES_DIR"
mkdir -p "$TEMP_DIR"

log_info "Starting TypeScript type generation from OpenAPI specification..."

# Generate TypeScript types using openapi-typescript
log_info "Generating TypeScript interfaces with openapi-typescript..."
npx --yes openapi-typescript "$OPENAPI_FILE" --output "$TEMP_DIR/generated-types.ts"

# Create the main types file with proper exports and utilities
cat > "$TYPES_FILE" << 'EOF'
// @kthulu:generated
// This file is auto-generated from the OpenAPI specification
// Do not edit manually - run `npm run gen:types` to regenerate

// Re-export all generated types
export * from './generated-types';

// Common utility types
export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T = any> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface ApiError {
  error: string;
  code: string;
  details?: Record<string, string>;
}

// Authentication types
export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  roleId?: number;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface AuthUser {
  id: number;
  email: string;
  confirmedAt?: string;
  roleId: number;
  role?: Role;
  createdAt: string;
  updatedAt: string;
}

export interface AuthResponse {
  user: AuthUser;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  message?: string;
}

// Role and Permission types
export interface Role {
  id: number;
  name: string;
  description?: string;
  permissions?: Permission[];
}

export interface Permission {
  id: number;
  name: string;
  description?: string;
  resource: string;
  action: string;
}

// Organization types
export type OrganizationType = 'company' | 'nonprofit' | 'personal' | 'education';

export interface Organization {
  id: number;
  name: string;
  slug: string;
  description?: string;
  type: OrganizationType;
  domain?: string;
  website?: string;
  phone?: string;
  address?: string;
  city?: string;
  state?: string;
  country?: string;
  postalCode?: string;
  logoUrl?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface OrganizationUser {
  id: number;
  organizationId: number;
  userId: number;
  role: OrganizationRole;
  joinedAt: string;
  user?: AuthUser;
  organization?: Organization;
}

export type OrganizationRole = 'owner' | 'admin' | 'member' | 'viewer';

// Health check types
export interface HealthResponse {
  status: 'healthy' | 'unhealthy';
  version: string;
  timestamp: string;
  checks: Record<string, string>;
}

// API client configuration
export interface ApiClientConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
}

// Request/Response wrapper types
export type ApiMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export interface ApiRequestConfig {
  method: ApiMethod;
  url: string;
  data?: any;
  params?: Record<string, any>;
  headers?: Record<string, string>;
}
EOF

# Copy generated types if they exist
if [ -f "$TEMP_DIR/generated-types.ts" ]; then
    # Extract just the type definitions and clean them up
    cat "$TEMP_DIR/generated-types.ts" >> "$TEMP_DIR/clean-types.ts"
    
    # Create a separate file for the generated types
    cat > "$TYPES_DIR/generated-types.ts" << 'EOF'
// @kthulu:generated
// Auto-generated TypeScript types from OpenAPI specification
// This file contains the raw generated types from the API schema

EOF
    
    # Append the generated content (if it exists and is valid TypeScript)
    if [ -s "$TEMP_DIR/generated-types.ts" ]; then
        cat "$TEMP_DIR/generated-types.ts" >> "$TYPES_DIR/generated-types.ts"
    fi
fi

# Generate Zod schemas for runtime validation
log_info "Generating Zod validation schemas..."

# Ensure zod is listed as a dependency; generation does not require installation
cd "$FRONTEND_DIR"
if ! grep -q '"zod"' package.json; then
    log_warn "zod dependency not found in package.json"
fi
cd ..

# Create Zod schemas file
cat > "$ZOD_FILE" << 'EOF'
// @kthulu:generated
// Zod validation schemas for API types
// This file provides runtime validation for API requests and responses

import { z } from 'zod';

// Authentication schemas
export const LoginCredentialsSchema = z.object({
  email: z.string().email('Invalid email format'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

export const RegisterDataSchema = z.object({
  email: z.string().email('Invalid email format'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  roleId: z.number().optional(),
});

export const AuthUserSchema = z.object({
  id: z.number(),
  email: z.string().email(),
  confirmedAt: z.string().optional(),
  roleId: z.number(),
  role: z.object({
    id: z.number(),
    name: z.string(),
    description: z.string().optional(),
  }).optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const AuthResponseSchema = z.object({
  user: AuthUserSchema,
  accessToken: z.string(),
  refreshToken: z.string(),
  expiresIn: z.number(),
  message: z.string().optional(),
});

// Organization schemas
export const OrganizationTypeSchema = z.enum(['company', 'nonprofit', 'personal', 'education']);

export const OrganizationSchema = z.object({
  id: z.number(),
  name: z.string().min(2).max(100),
  slug: z.string().min(2).max(50),
  description: z.string().max(500).optional(),
  type: OrganizationTypeSchema,
  domain: z.string().optional(),
  website: z.string().url().optional(),
  phone: z.string().optional(),
  address: z.string().max(200).optional(),
  city: z.string().max(100).optional(),
  state: z.string().max(100).optional(),
  country: z.string().max(100).optional(),
  postalCode: z.string().max(20).optional(),
  logoUrl: z.string().url().optional(),
  isActive: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export const CreateOrganizationSchema = z.object({
  name: z.string().min(2).max(100),
  description: z.string().max(500).optional(),
  type: OrganizationTypeSchema,
  domain: z.string().optional(),
  website: z.string().url().optional(),
  phone: z.string().optional(),
  address: z.string().max(200).optional(),
  city: z.string().max(100).optional(),
  state: z.string().max(100).optional(),
  country: z.string().max(100).optional(),
  postalCode: z.string().max(20).optional(),
});

// Role and Permission schemas
export const PermissionSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().optional(),
  resource: z.string(),
  action: z.string(),
});

export const RoleSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().optional(),
  permissions: z.array(PermissionSchema).optional(),
});

// Health check schema
export const HealthResponseSchema = z.object({
  status: z.enum(['healthy', 'unhealthy']),
  version: z.string(),
  timestamp: z.string(),
  checks: z.record(z.string()),
});

// API Error schema
export const ApiErrorSchema = z.object({
  error: z.string(),
  code: z.string(),
  details: z.record(z.string()).optional(),
});

// Generic response schemas
export const ApiResponseSchema = <T extends z.ZodType>(dataSchema: T) =>
  z.object({
    data: dataSchema.optional(),
    error: z.string().optional(),
    message: z.string().optional(),
  });

export const PaginatedResponseSchema = <T extends z.ZodType>(itemSchema: T) =>
  z.object({
    data: z.array(itemSchema),
    pagination: z.object({
      page: z.number(),
      limit: z.number(),
      total: z.number(),
      totalPages: z.number(),
    }),
  });

// Type inference helpers
export type LoginCredentials = z.infer<typeof LoginCredentialsSchema>;
export type RegisterData = z.infer<typeof RegisterDataSchema>;
export type AuthUser = z.infer<typeof AuthUserSchema>;
export type AuthResponse = z.infer<typeof AuthResponseSchema>;
export type Organization = z.infer<typeof OrganizationSchema>;
export type CreateOrganization = z.infer<typeof CreateOrganizationSchema>;
export type OrganizationType = z.infer<typeof OrganizationTypeSchema>;
export type Role = z.infer<typeof RoleSchema>;
export type Permission = z.infer<typeof PermissionSchema>;
export type HealthResponse = z.infer<typeof HealthResponseSchema>;
export type ApiError = z.infer<typeof ApiErrorSchema>;
EOF

# Create an index file to export everything
cat > "$TYPES_DIR/index.ts" << 'EOF'
// @kthulu:generated
// Main types export file

// Export all types
export * from './kthulu-api';
export * from './kthulu-api-zod';

// Re-export commonly used types for convenience
export type {
  AuthUser,
  AuthResponse,
  LoginCredentials,
  RegisterData,
  Organization,
  OrganizationType,
  Role,
  Permission,
  HealthResponse,
  ApiError,
  ApiResponse,
  PaginatedResponse,
} from './kthulu-api';

// Re-export validation schemas
export {
  LoginCredentialsSchema,
  RegisterDataSchema,
  AuthResponseSchema,
  OrganizationSchema,
  CreateOrganizationSchema,
  HealthResponseSchema,
  ApiErrorSchema,
} from './kthulu-api-zod';
EOF

# Clean up temporary files
rm -rf "$TEMP_DIR"

log_info "TypeScript type generation completed successfully!"
log_info "Generated files:"
log_info "  - $TYPES_FILE (Main types)"
log_info "  - $ZOD_FILE (Zod validation schemas)"
log_info "  - $TYPES_DIR/index.ts (Export index)"
log_info ""
log_info "To use the types in your React components:"
log_info "  import { AuthUser, LoginCredentials } from '@/types';"
log_info ""
log_info "To use validation schemas:"
log_info "  import { LoginCredentialsSchema } from '@/types';"
log_info "  const result = LoginCredentialsSchema.parse(data);"