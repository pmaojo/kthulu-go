import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { TestHelpers } from '../utils/test-helpers';

test.describe('Invoices Management', () => {
  let apiClient: ApiClient;
  let helpers: TestHelpers;
  let organizationId: number;

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    helpers = new TestHelpers(page, apiClient);
    
    // Login and create test organization
    await helpers.loginAsAdmin();
    organizationId = await helpers.createTestOrganization('Invoices Test Org');
    
    // Switch to the test organization context
    await page.goto(`/organizations/${organizationId}/invoices`);
  });

  test('should display invoices list', async ({ page }) => {
    await helpers.navigateToInvoices();
    
    // Should show invoices page
    await helpers.expectElementToBeVisible('invoices-page');
    await helpers.expectElementToBeVisible('create-invoice-button');
    
    // Should show table or empty state
    const hasInvoices = await page.locator('[data-testid="invoices-table"]').isVisible();
    if (!hasInvoices) {
      await helpers.expectElementToBeVisible('empty-state');
    }
  });

  test('should create a new invoice', async ({ page }) => {
    await helpers.navigateToInvoices();
    
    // Click create button
    await page.click('[data-testid="create-invoice-button"]');
    await helpers.waitForModal('create-invoice-modal');
    
    // Fill form
    const invoiceNumber = `INV-${Date.now()}`;
    await helpers.fillForm({
      number: invoiceNumber,
      'customer-name': 'Test Customer',
      'customer-email': 'customer@test.com',
      subtotal: '100.00',
      'tax-amount': '21.00',
      total: '121.00',
      currency: 'EUR',
      'due-date': '2024-12-31',
    });
    
    // Submit form
    await helpers.submitForm('create-invoice-form');
    
    // Should close modal and show success message
    await helpers.waitForToast('Invoice created successfully');
    
    // Should appear in the list
    await helpers.expectElementToContainText('invoices-table', invoiceNumber);
    await helpers.expectElementToContainText('invoices-table', 'Test Customer');
  });

  test('should edit an existing invoice', async ({ page }) => {
    // Create test invoice first
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Edit Test Customer');
    
    await helpers.navigateToInvoices();
    
    // Find and click edit button
    await page.click(`[data-testid="edit-invoice-${invoiceId}"]`);
    await helpers.waitForModal('edit-invoice-modal');
    
    // Update form
    const updatedCustomer = `Updated Customer ${Date.now()}`;
    await helpers.fillForm({
      'customer-name': updatedCustomer,
      'customer-email': 'updated@test.com',
      total: '150.00',
    });
    
    await helpers.submitForm('edit-invoice-form');
    
    await helpers.waitForToast('Invoice updated successfully');
    await helpers.expectElementToContainText('invoices-table', updatedCustomer);
  });

  test('should delete an invoice', async ({ page }) => {
    // Create test invoice first
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Delete Test Customer');
    
    await helpers.navigateToInvoices();
    
    // Click delete button
    await page.click(`[data-testid="delete-invoice-${invoiceId}"]`);
    await helpers.waitForModal('confirm-delete-modal');
    
    // Confirm deletion
    await page.click('[data-testid="confirm-delete-button"]');
    
    await helpers.waitForToast('Invoice deleted successfully');
    
    // Should not appear in the list anymore
    const invoiceExists = await page.locator(`[data-testid="invoice-row-${invoiceId}"]`).isVisible();
    expect(invoiceExists).toBeFalsy();
  });

  test('should filter invoices by status', async ({ page }) => {
    // Create invoices with different statuses
    await apiClient.createInvoice(organizationId, {
      number: 'INV-DRAFT-001',
      customerName: 'Draft Customer',
      status: 'draft',
      subtotal: 100,
      taxAmount: 21,
      total: 121,
      currency: 'EUR',
      issueDate: new Date().toISOString(),
      dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    });
    
    await apiClient.createInvoice(organizationId, {
      number: 'INV-PAID-001',
      customerName: 'Paid Customer',
      status: 'paid',
      subtotal: 200,
      taxAmount: 42,
      total: 242,
      currency: 'EUR',
      issueDate: new Date().toISOString(),
      dueDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    });
    
    await helpers.navigateToInvoices();
    
    // Filter by draft
    await page.selectOption('[data-testid="status-filter"]', 'draft');
    await page.waitForTimeout(500);
    
    // Should show only draft invoices
    await helpers.expectElementToContainText('invoices-table', 'Draft Customer');
    
    const paidExists = await page.locator('text=Paid Customer').isVisible();
    expect(paidExists).toBeFalsy();
    
    // Filter by paid
    await page.selectOption('[data-testid="status-filter"]', 'paid');
    await page.waitForTimeout(500);
    
    await helpers.expectElementToContainText('invoices-table', 'Paid Customer');
    
    const draftExists = await page.locator('text=Draft Customer').isVisible();
    expect(draftExists).toBeFalsy();
  });

  test('should search invoices', async ({ page }) => {
    // Create multiple test invoices
    await helpers.createTestInvoice(organizationId, 'Searchable Customer Alpha');
    await helpers.createTestInvoice(organizationId, 'Searchable Customer Beta');
    await helpers.createTestInvoice(organizationId, 'Different Client');
    
    await helpers.navigateToInvoices();
    
    // Search for specific invoices
    await helpers.searchInTable('Searchable');
    
    // Should show only matching results
    await helpers.expectElementToContainText('invoices-table', 'Searchable Customer Alpha');
    await helpers.expectElementToContainText('invoices-table', 'Searchable Customer Beta');
    
    // Should not show non-matching results
    const differentExists = await page.locator('text=Different Client').isVisible();
    expect(differentExists).toBeFalsy();
  });

  test('should handle invoice details view', async ({ page }) => {
    // Create test invoice
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Details Test Customer');
    
    await helpers.navigateToInvoices();
    
    // Click on invoice row to view details
    await page.click(`[data-testid="invoice-row-${invoiceId}"]`);
    
    // Should navigate to invoice details page
    await helpers.expectToBeOnPage(`/invoices/${invoiceId}`);
    
    // Should show invoice details
    await helpers.expectElementToBeVisible('invoice-details');
    await helpers.expectElementToContainText('customer-name', 'Details Test Customer');
    
    // Should show invoice sections
    await helpers.expectElementToBeVisible('invoice-header');
    await helpers.expectElementToBeVisible('invoice-items');
    await helpers.expectElementToBeVisible('invoice-totals');
    await helpers.expectElementToBeVisible('invoice-payments');
  });

  test('should add invoice items', async ({ page }) => {
    // Create test invoice and product
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Items Test Customer');
    const productId = await helpers.createTestProduct(organizationId, 'Test Product');
    
    await page.goto(`/invoices/${invoiceId}`);
    
    // Click add item button
    await page.click('[data-testid="add-item-button"]');
    await helpers.waitForModal('add-item-modal');
    
    // Fill item form
    await helpers.fillForm({
      'product-id': productId.toString(),
      quantity: '2',
      'unit-price': '50.00',
      description: 'Test item description',
    });
    
    await helpers.submitForm('add-item-form');
    
    await helpers.waitForToast('Invoice item added successfully');
    await helpers.expectElementToContainText('invoice-items', 'Test Product');
    await helpers.expectElementToContainText('invoice-items', '100.00'); // 2 * 50.00
  });

  test('should record invoice payment', async ({ page }) => {
    // Create test invoice
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Payment Test Customer');
    
    await page.goto(`/invoices/${invoiceId}`);
    
    // Click record payment button
    await page.click('[data-testid="record-payment-button"]');
    await helpers.waitForModal('record-payment-modal');
    
    // Fill payment form
    await helpers.fillForm({
      amount: '121.00',
      'payment-method': 'bank_transfer',
      'payment-date': '2024-01-15',
      reference: 'PAYMENT-001',
      notes: 'Full payment received',
    });
    
    await helpers.submitForm('record-payment-form');
    
    await helpers.waitForToast('Payment recorded successfully');
    await helpers.expectElementToContainText('invoice-payments', '121.00');
    await helpers.expectElementToContainText('invoice-status', 'paid');
  });

  test('should change invoice status', async ({ page }) => {
    // Create test invoice
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Status Test Customer');
    
    await page.goto(`/invoices/${invoiceId}`);
    
    // Should be draft by default
    await helpers.expectElementToContainText('invoice-status', 'draft');
    
    // Change status to sent
    await page.click('[data-testid="change-status-button"]');
    await page.selectOption('[data-testid="status-select"]', 'sent');
    await page.click('[data-testid="update-status-button"]');
    
    await helpers.waitForToast('Invoice status updated');
    await helpers.expectElementToContainText('invoice-status', 'sent');
  });

  test('should generate invoice PDF', async ({ page }) => {
    // Create test invoice
    const invoiceId = await helpers.createTestInvoice(organizationId, 'PDF Test Customer');
    
    await page.goto(`/invoices/${invoiceId}`);
    
    // Click generate PDF button
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-testid="generate-pdf-button"]');
    const download = await downloadPromise;
    
    // Should download a PDF file
    expect(download.suggestedFilename()).toMatch(/invoice.*\.pdf$/);
  });

  test('should send invoice by email', async ({ page }) => {
    // Create test invoice
    const invoiceId = await helpers.createTestInvoice(organizationId, 'Email Test Customer');
    
    await page.goto(`/invoices/${invoiceId}`);
    
    // Click send email button
    await page.click('[data-testid="send-email-button"]');
    await helpers.waitForModal('send-email-modal');
    
    // Fill email form
    await helpers.fillForm({
      to: 'customer@test.com',
      subject: 'Your Invoice',
      message: 'Please find your invoice attached.',
    });
    
    await helpers.submitForm('send-email-form');
    
    await helpers.waitForToast('Invoice sent successfully');
  });

  test('should display invoice statistics', async ({ page }) => {
    // Create multiple test invoices with different statuses
    await apiClient.createInvoice(organizationId, {
      number: 'INV-STATS-001',
      customerName: 'Stats Customer 1',
      status: 'paid',
      subtotal: 100,
      taxAmount: 21,
      total: 121,
      currency: 'EUR',
      issueDate: new Date().toISOString(),
      dueDate: new Date().toISOString(),
    });
    
    await apiClient.createInvoice(organizationId, {
      number: 'INV-STATS-002',
      customerName: 'Stats Customer 2',
      status: 'overdue',
      subtotal: 200,
      taxAmount: 42,
      total: 242,
      currency: 'EUR',
      issueDate: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toISOString(),
      dueDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
    });
    
    await helpers.navigateToInvoices();
    
    // Should show statistics
    await helpers.expectElementToBeVisible('invoice-stats');
    await helpers.expectElementToContainText('total-invoices', '2');
    await helpers.expectElementToContainText('total-amount', '363.00'); // 121 + 242
    await helpers.expectElementToContainText('paid-amount', '121.00');
    await helpers.expectElementToContainText('overdue-amount', '242.00');
  });

  test('should validate invoice form', async ({ page }) => {
    await helpers.navigateToInvoices();
    
    await page.click('[data-testid="create-invoice-button"]');
    await helpers.waitForModal('create-invoice-modal');
    
    // Try to submit empty form
    await helpers.submitForm('create-invoice-form');
    
    // Should show validation errors
    await helpers.expectFormError('number', 'Invoice number is required');
    await helpers.expectFormError('customer-name', 'Customer name is required');
    await helpers.expectFormError('total', 'Total amount is required');
    
    // Test invalid email
    await helpers.fillForm({
      number: 'INV-001',
      'customer-name': 'Test Customer',
      'customer-email': 'invalid-email',
      total: '100.00',
    });
    
    await helpers.submitForm('create-invoice-form');
    await helpers.expectFormError('customer-email', 'Invalid email format');
    
    // Test invalid amounts
    await helpers.fillForm({
      'customer-email': 'valid@test.com',
      subtotal: 'invalid-amount',
      total: '-50.00',
    });
    
    await helpers.submitForm('create-invoice-form');
    await helpers.expectFormError('subtotal', 'Invalid amount format');
    await helpers.expectFormError('total', 'Amount must be positive');
  });

  test('should handle overdue invoices', async ({ page }) => {
    // Create overdue invoice
    await apiClient.createInvoice(organizationId, {
      number: 'INV-OVERDUE-001',
      customerName: 'Overdue Customer',
      status: 'sent',
      subtotal: 100,
      taxAmount: 21,
      total: 121,
      currency: 'EUR',
      issueDate: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toISOString(),
      dueDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
    });
    
    await helpers.navigateToInvoices();
    
    // Click overdue filter
    await page.click('[data-testid="overdue-filter"]');
    
    // Should show overdue invoices
    await helpers.expectElementToContainText('invoices-table', 'Overdue Customer');
    await helpers.expectElementToBeVisible('[data-testid="overdue-badge"]');
    
    // Should show overdue actions
    await helpers.expectElementToBeVisible('[data-testid="send-reminder-button"]');
  });

  test('should export invoices', async ({ page }) => {
    // Create some test invoices
    await helpers.createTestInvoice(organizationId, 'Export Customer 1');
    await helpers.createTestInvoice(organizationId, 'Export Customer 2');
    
    await helpers.navigateToInvoices();
    
    // Click export button
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-testid="export-invoices-button"]');
    const download = await downloadPromise;
    
    // Should download a file
    expect(download.suggestedFilename()).toMatch(/invoices.*\.(csv|xlsx)$/);
  });
});