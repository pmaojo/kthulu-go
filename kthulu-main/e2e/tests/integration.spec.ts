import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { TestHelpers } from '../utils/test-helpers';

test.describe('Full Application Integration', () => {
  let apiClient: ApiClient;
  let helpers: TestHelpers;

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    helpers = new TestHelpers(page, apiClient);
  });

  test('should complete full business workflow', async ({ page }) => {
    // 1. Register and login
    await page.goto('/register');
    const userEmail = `workflow@test-${Date.now()}.com`;
    
    await helpers.fillForm({
      email: userEmail,
      password: 'TestPassword123!',
      'confirm-password': 'TestPassword123!',
    });
    
    await helpers.submitForm('register-form');
    await helpers.expectToBeOnPage('/dashboard');
    
    // 2. Create organization
    await helpers.navigateToOrganizations();
    await page.click('[data-testid="create-organization-button"]');
    await helpers.waitForModal('create-organization-modal');
    
    const orgName = `Workflow Org ${Date.now()}`;
    await helpers.fillForm({
      name: orgName,
      description: 'Organization for workflow test',
    });
    
    await helpers.submitForm('create-organization-form');
    await helpers.waitForToast('Organization created successfully');
    
    // 3. Create customer contact
    await helpers.navigateToContacts();
    await page.click('[data-testid="create-contact-button"]');
    await helpers.waitForModal('create-contact-modal');
    
    const customerName = `Workflow Customer ${Date.now()}`;
    await helpers.fillForm({
      'company-name': customerName,
      email: 'customer@workflow.com',
      type: 'customer',
    });
    
    await helpers.submitForm('create-contact-form');
    await helpers.waitForToast('Contact created successfully');
    
    // 4. Create product
    await helpers.navigateToProducts();
    await page.click('[data-testid="create-product-button"]');
    await helpers.waitForModal('create-product-modal');
    
    const productName = `Workflow Product ${Date.now()}`;
    const productSku = `WF-${Date.now()}`;
    await helpers.fillForm({
      name: productName,
      sku: productSku,
      description: 'Product for workflow test',
      price: '99.99',
    });
    
    await helpers.submitForm('create-product-form');
    await helpers.waitForToast('Product created successfully');
    
    // 5. Create invoice
    await helpers.navigateToInvoices();
    await page.click('[data-testid="create-invoice-button"]');
    await helpers.waitForModal('create-invoice-modal');
    
    const invoiceNumber = `WF-INV-${Date.now()}`;
    await helpers.fillForm({
      number: invoiceNumber,
      'customer-name': customerName,
      'customer-email': 'customer@workflow.com',
      subtotal: '99.99',
      'tax-amount': '21.00',
      total: '120.99',
      currency: 'EUR',
      'due-date': '2024-12-31',
    });
    
    await helpers.submitForm('create-invoice-form');
    await helpers.waitForToast('Invoice created successfully');
    
    // 6. Verify all data is connected
    await helpers.expectElementToContainText('invoices-table', invoiceNumber);
    await helpers.expectElementToContainText('invoices-table', customerName);
    
    // 7. Navigate to dashboard and verify summary
    await page.click('[data-testid="nav-dashboard"]');
    await helpers.expectToBeOnPage('/dashboard');
    
    // Should show recent activity
    await helpers.expectElementToBeVisible('recent-activity');
    await helpers.expectElementToContainText('recent-activity', 'Invoice created');
    await helpers.expectElementToContainText('recent-activity', 'Product created');
    await helpers.expectElementToContainText('recent-activity', 'Contact created');
  });

  test('should handle multi-organization workflow', async ({ page }) => {
    // Login as admin
    await helpers.loginAsAdmin();
    
    // Create first organization
    const org1Id = await helpers.createTestOrganization('Multi Org 1');
    const org2Id = await helpers.createTestOrganization('Multi Org 2');
    
    // Switch to first organization
    await page.goto(`/organizations/${org1Id}`);
    
    // Create data in first organization
    const contact1Id = await helpers.createTestContact(org1Id, 'Org 1 Contact');
    const product1Id = await helpers.createTestProduct(org1Id, 'Org 1 Product');
    
    // Switch to second organization
    await page.goto(`/organizations/${org2Id}`);
    
    // Create data in second organization
    const contact2Id = await helpers.createTestContact(org2Id, 'Org 2 Contact');
    const product2Id = await helpers.createTestProduct(org2Id, 'Org 2 Product');
    
    // Verify data isolation
    await helpers.navigateToContacts();
    await helpers.expectElementToContainText('contacts-table', 'Org 2 Contact');
    
    const org1ContactExists = await page.locator('text=Org 1 Contact').isVisible();
    expect(org1ContactExists).toBeFalsy();
    
    // Switch back to first organization
    await page.goto(`/organizations/${org1Id}`);
    await helpers.navigateToContacts();
    
    await helpers.expectElementToContainText('contacts-table', 'Org 1 Contact');
    
    const org2ContactExists = await page.locator('text=Org 2 Contact').isVisible();
    expect(org2ContactExists).toBeFalsy();
  });

  test('should handle concurrent user sessions', async ({ page, context }) => {
    // First user session
    await helpers.loginAsAdmin();
    const org1Id = await helpers.createTestOrganization('Concurrent Org 1');
    
    // Second user session
    const secondPage = await context.newPage();
    const secondApiClient = new ApiClient(page.url());
    const secondHelpers = new TestHelpers(secondPage, secondApiClient);
    
    await secondHelpers.loginAsUser();
    const org2Id = await secondHelpers.createTestOrganization('Concurrent Org 2');
    
    // Both users work independently
    await helpers.createTestContact(org1Id, 'Admin Contact');
    await secondHelpers.createTestContact(org2Id, 'User Contact');
    
    // Verify isolation
    await helpers.navigateToContacts();
    await helpers.expectElementToContainText('contacts-table', 'Admin Contact');
    
    await secondHelpers.navigateToContacts();
    await secondHelpers.expectElementToContainText('contacts-table', 'User Contact');
    
    await secondPage.close();
  });

  test('should handle error scenarios gracefully', async ({ page }) => {
    await helpers.loginAsAdmin();
    const orgId = await helpers.createTestOrganization('Error Test Org');
    
    // Test network error handling
    await page.route('**/api/contacts', route => route.abort());
    
    await helpers.navigateToContacts();
    
    // Should show error message
    await helpers.expectElementToBeVisible('error-message');
    await helpers.expectElementToContainText('error-message', 'Failed to load contacts');
    
    // Should have retry button
    await helpers.expectElementToBeVisible('retry-button');
    
    // Restore network and retry
    await page.unroute('**/api/contacts');
    await page.click('[data-testid="retry-button"]');
    
    // Should load successfully
    await helpers.expectElementToBeVisible('contacts-table');
  });

  test('should handle offline scenarios', async ({ page, context }) => {
    await helpers.loginAsAdmin();
    const orgId = await helpers.createTestOrganization('Offline Test Org');
    
    // Create some data while online
    await helpers.createTestContact(orgId, 'Online Contact');
    await helpers.navigateToContacts();
    
    // Go offline
    await context.setOffline(true);
    
    // Try to create contact while offline
    await page.click('[data-testid="create-contact-button"]');
    await helpers.waitForModal('create-contact-modal');
    
    await helpers.fillForm({
      'company-name': 'Offline Contact',
      email: 'offline@test.com',
      type: 'customer',
    });
    
    await helpers.submitForm('create-contact-form');
    
    // Should show offline message
    await helpers.waitForToast('You are offline. Changes will be saved when connection is restored.');
    
    // Go back online
    await context.setOffline(false);
    
    // Should sync changes
    await helpers.waitForToast('Changes synced successfully');
    await helpers.expectElementToContainText('contacts-table', 'Offline Contact');
  });

  test('should handle performance under load', async ({ page }) => {
    await helpers.loginAsAdmin();
    const orgId = await helpers.createTestOrganization('Performance Test Org');
    
    // Create many contacts to test performance
    const contactPromises = [];
    for (let i = 0; i < 50; i++) {
      contactPromises.push(
        helpers.createTestContact(orgId, `Performance Contact ${i}`)
      );
    }
    await Promise.all(contactPromises);
    
    // Measure page load time
    const startTime = Date.now();
    await helpers.navigateToContacts();
    const loadTime = Date.now() - startTime;
    
    // Should load within reasonable time (5 seconds)
    expect(loadTime).toBeLessThan(5000);
    
    // Should show pagination for large datasets
    await helpers.expectElementToBeVisible('pagination');
    
    // Test search performance
    const searchStartTime = Date.now();
    await helpers.searchInTable('Performance Contact 1');
    const searchTime = Date.now() - searchStartTime;
    
    // Search should be fast (2 seconds)
    expect(searchTime).toBeLessThan(2000);
    
    // Should show filtered results
    await helpers.expectElementToContainText('contacts-table', 'Performance Contact 1');
  });

  test('should handle data consistency across modules', async ({ page }) => {
    await helpers.loginAsAdmin();
    const orgId = await helpers.createTestOrganization('Consistency Test Org');
    
    // Create interconnected data
    const contactId = await helpers.createTestContact(orgId, 'Consistency Customer');
    const productId = await helpers.createTestProduct(orgId, 'Consistency Product');
    
    // Create invoice with the contact and product
    const invoiceId = await helpers.createTestInvoice(orgId, 'Consistency Customer');
    
    // Update contact name
    await helpers.navigateToContacts();
    await page.click(`[data-testid="edit-contact-${contactId}"]`);
    await helpers.waitForModal('edit-contact-modal');
    
    const updatedName = 'Updated Consistency Customer';
    await helpers.fillForm({
      'company-name': updatedName,
    });
    
    await helpers.submitForm('edit-contact-form');
    await helpers.waitForToast('Contact updated successfully');
    
    // Verify invoice reflects the updated contact name
    await helpers.navigateToInvoices();
    await page.click(`[data-testid="invoice-row-${invoiceId}"]`);
    
    // Should show updated customer name
    await helpers.expectElementToContainText('customer-name', updatedName);
    
    // Update product name
    await helpers.navigateToProducts();
    await page.click(`[data-testid="edit-product-${productId}"]`);
    await helpers.waitForModal('edit-product-modal');
    
    const updatedProductName = 'Updated Consistency Product';
    await helpers.fillForm({
      name: updatedProductName,
    });
    
    await helpers.submitForm('edit-product-form');
    await helpers.waitForToast('Product updated successfully');
    
    // Verify consistency across all modules
    await helpers.navigateToContacts();
    await helpers.expectElementToContainText('contacts-table', updatedName);
    
    await helpers.navigateToProducts();
    await helpers.expectElementToContainText('products-table', updatedProductName);
    
    await helpers.navigateToInvoices();
    await helpers.expectElementToContainText('invoices-table', updatedName);
  });

  test('should handle browser refresh and state persistence', async ({ page }) => {
    await helpers.loginAsAdmin();
    const orgId = await helpers.createTestOrganization('Persistence Test Org');
    
    // Navigate to contacts and apply filters
    await helpers.navigateToContacts();
    await helpers.createTestContact(orgId, 'Persistence Customer');
    await helpers.createTestContact(orgId, 'Persistence Supplier');
    
    // Apply filter
    await page.selectOption('[data-testid="type-filter"]', 'customer');
    await page.waitForTimeout(500);
    
    // Refresh page
    await page.reload();
    
    // Should maintain authentication
    await helpers.expectToBeOnPage('/contacts');
    
    // Should maintain filter state
    const filterValue = await page.locator('[data-testid="type-filter"]').inputValue();
    expect(filterValue).toBe('customer');
    
    // Should show filtered results
    await helpers.expectElementToContainText('contacts-table', 'Persistence Customer');
    
    const supplierExists = await page.locator('text=Persistence Supplier').isVisible();
    expect(supplierExists).toBeFalsy();
  });
});