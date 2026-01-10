-- +goose Up
-- Optimal-style migration: Cross-database compatible

-- Create permissions table
CREATE TABLE permissions (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for permissions table
CREATE UNIQUE INDEX idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_permissions_created_at ON permissions(created_at);

-- Create role_permissions junction table (many-to-many relationship)
CREATE TABLE role_permissions (
    id INTEGER PRIMARY KEY,
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Create indexes for role_permissions table
CREATE UNIQUE INDEX idx_role_permissions_role_permission ON role_permissions(role_id, permission_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_created_at ON role_permissions(created_at);

-- Insert default permissions (Optimal: simple and clean)
INSERT INTO permissions (name, description, resource, action) VALUES 
    ('View Users', 'View user profiles and information', 'users', 'read'),
    ('Create Users', 'Create new user accounts', 'users', 'create'),
    ('Update Users', 'Update user profiles and information', 'users', 'update'),
    ('Delete Users', 'Delete user accounts', 'users', 'delete'),
    ('View Organizations', 'View organization information', 'organizations', 'read'),
    ('Create Organizations', 'Create new organizations', 'organizations', 'create'),
    ('Update Organizations', 'Update organization information', 'organizations', 'update'),
    ('Delete Organizations', 'Delete organizations', 'organizations', 'delete'),
    ('Manage Organization Members', 'Add/remove organization members', 'organizations', 'manage_members'),
    ('View Roles', 'View roles and permissions', 'roles', 'read'),
    ('Create Roles', 'Create new roles', 'roles', 'create'),
    ('Update Roles', 'Update roles and permissions', 'roles', 'update'),
    ('Delete Roles', 'Delete roles', 'roles', 'delete'),
    ('System Admin', 'Full system administration access', 'system', 'admin'),
    ('View System Logs', 'View system logs and audit trails', 'system', 'logs'),
    ('Manage System Settings', 'Manage system configuration', 'system', 'settings');

-- Assign permissions to default roles
-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'admin';

-- Moderator gets limited permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'moderator' 
AND p.action IN ('read', 'update') 
AND p.resource IN ('users', 'organizations');

-- Regular user gets basic read permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'user' 
AND p.action = 'read' 
AND p.resource IN ('users', 'organizations');

-- +goose Down
-- Optimal-style rollback: Clean and simple

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;