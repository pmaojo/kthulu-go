import axios, { AxiosInstance, AxiosResponse } from 'axios';

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  user: {
    id: number;
    email: string;
  };
}

export interface Organization {
  id: number;
  name: string;
  slug: string;
  description?: string;
}

export interface Contact {
  id: number;
  organizationId: number;
  companyName?: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  type: 'customer' | 'supplier' | 'lead' | 'partner';
  isActive: boolean;
}

export interface Product {
  id: number;
  organizationId: number;
  name: string;
  sku: string;
  description?: string;
  isActive: boolean;
}

export interface Invoice {
  id: number;
  organizationId: number;
  number: string;
  customerName: string;
  customerEmail?: string;
  status: 'draft' | 'sent' | 'paid' | 'overdue' | 'cancelled';
  subtotal: number;
  taxAmount: number;
  total: number;
  currency: string;
  issueDate: string;
  dueDate: string;
}

export class ApiClient {
  private client: AxiosInstance;
  private accessToken?: string;

  constructor(baseURL: string) {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use((config) => {
      if (this.accessToken) {
        config.headers.Authorization = `Bearer ${this.accessToken}`;
      }
      return config;
    });
  }

  // Health check
  async healthCheck(): Promise<void> {
    await this.client.get('/health');
  }

  // Database management
  async resetDatabase(): Promise<void> {
    try {
      await this.client.post('/test/reset-db');
    } catch (error) {
      // Ignore if endpoint doesn't exist
      console.warn('Reset database endpoint not available');
    }
  }

  // Authentication
  async register(email: string, password: string): Promise<AuthResponse> {
    const response: AxiosResponse<AuthResponse> = await this.client.post('/auth/register', {
      email,
      password,
    });
    
    this.accessToken = response.data.accessToken;
    return response.data;
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    const response: AxiosResponse<AuthResponse> = await this.client.post('/auth/login', {
      email,
      password,
    });
    
    this.accessToken = response.data.accessToken;
    return response.data;
  }

  async logout(): Promise<void> {
    await this.client.post('/auth/logout');
    this.accessToken = undefined;
  }

  async getProfile(): Promise<any> {
    const response = await this.client.get('/users/me');
    return response.data;
  }

  // Organizations
  async createOrganization(name: string, description?: string): Promise<Organization> {
    const response: AxiosResponse<Organization> = await this.client.post('/organizations', {
      name,
      description,
    });
    return response.data;
  }

  async getOrganizations(): Promise<Organization[]> {
    const response = await this.client.get('/organizations');
    return response.data.organizations || response.data;
  }

  async getOrganization(id: number): Promise<Organization> {
    const response: AxiosResponse<Organization> = await this.client.get(`/organizations/${id}`);
    return response.data;
  }

  // Contacts
  async createContact(organizationId: number, contact: Partial<Contact>): Promise<Contact> {
    const response: AxiosResponse<Contact> = await this.client.post('/contacts', {
      ...contact,
      organizationId,
    }, {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data;
  }

  async getContacts(organizationId: number): Promise<Contact[]> {
    const response = await this.client.get('/contacts', {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data.contacts || response.data;
  }

  // Products
  async createProduct(organizationId: number, product: Partial<Product>): Promise<Product> {
    const response: AxiosResponse<Product> = await this.client.post('/products', {
      ...product,
      organizationId,
    }, {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data;
  }

  async getProducts(organizationId: number): Promise<Product[]> {
    const response = await this.client.get('/products', {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data.products || response.data;
  }

  // Invoices
  async createInvoice(organizationId: number, invoice: Partial<Invoice>): Promise<Invoice> {
    const response: AxiosResponse<Invoice> = await this.client.post('/invoices', {
      ...invoice,
      organizationId,
    }, {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data;
  }

  async getInvoices(organizationId: number): Promise<Invoice[]> {
    const response = await this.client.get('/invoices', {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data.invoices || response.data;
  }

  async getInvoiceStats(organizationId: number): Promise<any> {
    const response = await this.client.get('/invoices/stats', {
      headers: {
        'X-Organization-ID': organizationId.toString(),
      },
    });
    return response.data;
  }

  // Utility methods
  setAuthToken(token: string): void {
    this.accessToken = token;
  }

  clearAuthToken(): void {
    this.accessToken = undefined;
  }
}