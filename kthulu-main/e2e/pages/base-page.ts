import { Page, expect, Locator } from '@playwright/test';
import { ApiClient } from '../utils/api-client';

export class BasePage {
  constructor(
    public page: Page,
    protected apiClient: ApiClient
  ) {}

  locator(testId: string): Locator {
    return this.page.locator(`[data-testid="${testId}"]`);
  }

  locatorByText(text: string): Locator {
    return this.page.locator(`text=${text}`);
  }

  async click(testId: string): Promise<void> {
    await this.locator(testId).click();
  }

  async fill(testId: string, value: string): Promise<void> {
    await this.locator(testId).fill(value);
  }

  async selectOption(testId: string, value: string): Promise<void> {
    await this.locator(testId).selectOption(value);
  }

  async check(testId: string): Promise<void> {
    await this.locator(testId).check();
  }

  async waitForTimeout(ms: number): Promise<void> {
    await this.page.waitForTimeout(ms);
  }

  // Authentication helpers
  async loginAsAdmin(): Promise<void> {
    await this.login('admin@test.com', 'TestPassword123!');
  }

  async loginAsUser(): Promise<void> {
    await this.login('user@test.com', 'TestPassword123!');
  }

  async loginAsManager(): Promise<void> {
    await this.login('manager@test.com', 'TestPassword123!');
  }

  async login(email: string, password: string): Promise<void> {
    await this.page.goto('/login');
    
    await this.page.fill('[data-testid="email-input"]', email);
    await this.page.fill('[data-testid="password-input"]', password);
    await this.page.click('[data-testid="login-button"]');
    
    // Wait for redirect to dashboard
    await this.page.waitForURL('/dashboard', { timeout: 10000 });
    
    // Also authenticate the API client
    await this.apiClient.login(email, password);
  }

  async logout(): Promise<void> {
    await this.page.click('[data-testid="user-menu"]');
    await this.page.click('[data-testid="logout-button"]');
    
    // Wait for redirect to login
    await this.page.waitForURL('/login', { timeout: 10000 });
    
    // Clear API client auth
    this.apiClient.clearAuthToken();
  }

  // Navigation helpers
  async navigateTo(path: string): Promise<void> {
    await this.page.goto(path);
    await this.page.waitForLoadState('networkidle');
  }

  // Form helpers
  async fillForm(formData: Record<string, string>): Promise<void> {
    for (const [field, value] of Object.entries(formData)) {
      await this.page.fill(`[data-testid="${field}-input"]`, value);
    }
  }

  async submitForm(formTestId = 'form'): Promise<void> {
    await this.page.click(`[data-testid="${formTestId}"] [type="submit"]`);
  }

  // Wait helpers
  async waitForToast(message?: string): Promise<void> {
    const toastSelector = '[data-testid="toast"]';
    await this.page.waitForSelector(toastSelector, { timeout: 5000 });
    
    if (message) {
      await expect(this.page.locator(toastSelector)).toContainText(message);
    }
  }

  async waitForModal(modalTestId: string): Promise<void> {
    await this.page.waitForSelector(`[data-testid="${modalTestId}"]`, { timeout: 5000 });
  }

  async closeModal(modalTestId: string): Promise<void> {
    await this.page.click(`[data-testid="${modalTestId}"] [data-testid="close-button"]`);
    await this.page.waitForSelector(`[data-testid="${modalTestId}"]`, { state: 'hidden' });
  }

  // Table helpers
  async getTableRowCount(tableTestId = 'data-table'): Promise<number> {
    const rows = await this.page.locator(`[data-testid="${tableTestId}"] tbody tr`).count();
    return rows;
  }

  async clickTableRow(rowIndex: number, tableTestId = 'data-table'): Promise<void> {
    await this.page.click(`[data-testid="${tableTestId}"] tbody tr:nth-child(${rowIndex + 1})`);
  }

  async searchInTable(query: string, searchTestId = 'search-input'): Promise<void> {
    await this.page.fill(`[data-testid="${searchTestId}"]`, query);
    await this.page.press(`[data-testid="${searchTestId}"]`, 'Enter');
    await this.page.waitForTimeout(500); // Wait for search to complete
  }

  // Assertion helpers
  async expectToBeOnPage(path: string): Promise<void> {
    await expect(this.page).toHaveURL(new RegExp(path));
  }

  async expectElementToBeVisible(testId: string): Promise<void> {
    await expect(this.page.locator(`[data-testid="${testId}"]`)).toBeVisible();
  }

  async expectElementToContainText(testId: string, text: string): Promise<void> {
    await expect(this.page.locator(`[data-testid="${testId}"]`)).toContainText(text);
  }

  async expectFormError(fieldTestId: string, errorMessage: string): Promise<void> {
    await expect(this.page.locator(`[data-testid="${fieldTestId}-error"]`)).toContainText(errorMessage);
  }

  // Screenshot helpers
  async takeScreenshot(name: string): Promise<void> {
    await this.page.screenshot({ 
      path: `test-results/screenshots/${name}.png`,
      fullPage: true 
    });
  }

  // Data setup helpers
  async createTestOrganization(name = 'Test Organization'): Promise<number> {
    const org = await this.apiClient.createOrganization(name, 'Test organization for E2E tests');
    return org.id;
  }

  async createTestContact(organizationId: number, name = 'Test Contact'): Promise<number> {
    const contact = await this.apiClient.createContact(organizationId, {
      companyName: name,
      email: 'test@example.com',
      type: 'customer',
      isActive: true,
    });
    return contact.id;
  }

  async createTestProduct(organizationId: number, name = 'Test Product'): Promise<number> {
    const product = await this.apiClient.createProduct(organizationId, {
      name,
      sku: `TEST-${Date.now()}`,
      description: 'Test product for E2E tests',
      isActive: true,
    });
    return product.id;
  }

  async createTestInvoice(organizationId: number, customerName = 'Test Customer'): Promise<number> {
    const invoice = await this.apiClient.createInvoice(organizationId, {
      number: `INV-${Date.now()}`,
      customerName,
      customerEmail: 'customer@example.com',
      status: 'draft',
      subtotal: 100.00,
      taxAmount: 21.00,
      total: 121.00,
      currency: 'EUR',
      issueDate: new Date().toISOString(),
      dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    });
    return invoice.id;
  }

  // Performance helpers
  async measurePageLoadTime(): Promise<number> {
    const startTime = Date.now();
    await this.page.waitForLoadState('networkidle');
    return Date.now() - startTime;
  }

  async measureApiResponseTime(apiCall: () => Promise<any>): Promise<number> {
    const startTime = Date.now();
    await apiCall();
    return Date.now() - startTime;
  }
}