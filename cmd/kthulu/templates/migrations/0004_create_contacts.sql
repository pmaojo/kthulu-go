-- +goose Up
-- Optimal-style migration: Cross-database compatible contacts

-- Create contacts table
CREATE TABLE contacts (
    id INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('customer', 'supplier', 'lead', 'partner')),
    company_name TEXT,
    first_name TEXT,
    last_name TEXT,
    email TEXT,
    phone TEXT,
    mobile TEXT,
    website TEXT,
    tax_number TEXT,
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Create indexes for contacts table
CREATE INDEX idx_contacts_organization_id ON contacts(organization_id);
CREATE INDEX idx_contacts_type ON contacts(type);
CREATE INDEX idx_contacts_company_name ON contacts(company_name);
CREATE INDEX idx_contacts_email ON contacts(email);
CREATE INDEX idx_contacts_is_active ON contacts(is_active);
CREATE INDEX idx_contacts_created_at ON contacts(created_at);
CREATE INDEX idx_contacts_full_name ON contacts(first_name, last_name);
CREATE INDEX idx_contacts_org_type_active ON contacts(organization_id, type, is_active);

-- Create contact_addresses table
CREATE TABLE contact_addresses (
    id INTEGER PRIMARY KEY,
    contact_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('billing', 'shipping', 'office', 'home', 'other')),
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    city TEXT NOT NULL,
    state TEXT,
    country TEXT NOT NULL,
    postal_code TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Create indexes for contact_addresses table
CREATE INDEX idx_contact_addresses_contact_id ON contact_addresses(contact_id);
CREATE INDEX idx_contact_addresses_type ON contact_addresses(type);
CREATE INDEX idx_contact_addresses_is_primary ON contact_addresses(is_primary);
CREATE INDEX idx_contact_addresses_country ON contact_addresses(country);
CREATE INDEX idx_contact_addresses_city ON contact_addresses(city);
CREATE INDEX idx_contact_addresses_created_at ON contact_addresses(created_at);

-- Create contact_phones table
CREATE TABLE contact_phones (
    id INTEGER PRIMARY KEY,
    contact_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('work', 'mobile', 'home', 'fax', 'other')),
    number TEXT NOT NULL,
    extension TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);

-- Create indexes for contact_phones table
CREATE INDEX idx_contact_phones_contact_id ON contact_phones(contact_id);
CREATE INDEX idx_contact_phones_type ON contact_phones(type);
CREATE INDEX idx_contact_phones_number ON contact_phones(number);
CREATE INDEX idx_contact_phones_is_primary ON contact_phones(is_primary);
CREATE INDEX idx_contact_phones_created_at ON contact_phones(created_at);

-- +goose Down
-- Optimal-style rollback: Clean and simple

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS contact_phones;
DROP TABLE IF EXISTS contact_addresses;
DROP TABLE IF EXISTS contacts;