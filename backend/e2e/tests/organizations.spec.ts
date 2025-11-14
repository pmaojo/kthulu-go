import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { TestHelpers } from '../utils/test-helpers';

test.describe('Organizations Management', () => {
  let apiClient: ApiClient;
  let helpers: TestHelpers;

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    helpers = new TestHelpers(page, apiClient);
    
    // Login as admin for all tests
    await helpers.loginAsAdmin();
  });

  test('should display organizations list', async ({ page }) => {
    await helpers.navigateToOrganizations();
    
    // Should show organizations page
    await helpers.expectElementToBeVisible('organizations-page');
    await helpers.expectElementToBeVisible('create-organization-button');
    
    // Should show table or empty state
    const hasOrganizations = await page.locator('[data-testid="organizations-table"]').isVisible();
    if (!hasOrganizations) {
      await helpers.expectElementToBeVisible('empty-state');
    }
  });

  test('should create a new organization', async ({ page }) => {
    await helpers.navigateToOrganizations();
    
    // Click create button
    await page.click('[data-testid="create-organization-button"]');
    await helpers.waitForModal('create-organization-modal');
    
    // Fill form
    const orgName = `Test Organization ${Date.now()}`;
    await helpers.fillForm({
      name: orgName,
      description: 'Test organization created by E2E test',
    });
    
    // Submit form
    await helpers.submitForm('create-organization-form');
    
    // Should close modal and show success message
    await helpers.waitForToast('Organization created successfully');
    
    // Should appear in the list
    await helpers.expectElementToContainText('organizations-table', orgName);
  });

  test('should edit an existing organization', async ({ page }) => {
    // Create test organization first
    const orgId = await helpers.createTestOrganization('Edit Test Org');
    
    await helpers.navigateToOrganizations();
    
    // Find and click edit button for the organization
    await page.click(`[data-testid="edit-organization-${orgId}"]`);
    await helpers.waitForModal('edit-organization-modal');
    
    // Update form
    const updatedName = `Updated Organization ${Date.now()}`;
    await helpers.fillForm({
      name: updatedName,
      description: 'Updated description',
    });
    
    await helpers.submitForm('edit-organization-form');
    
    // Should show success message
    await helpers.waitForToast('Organization updated successfully');
    
    // Should show updated name in list
    await helpers.expectElementToContainText('organizations-table', updatedName);
  });

  test('should delete an organization', async ({ page }) => {
    // Create test organization first
    const orgId = await helpers.createTestOrganization('Delete Test Org');
    
    await helpers.navigateToOrganizations();
    
    // Click delete button
    await page.click(`[data-testid="delete-organization-${orgId}"]`);
    await helpers.waitForModal('confirm-delete-modal');
    
    // Confirm deletion
    await page.click('[data-testid="confirm-delete-button"]');
    
    // Should show success message
    await helpers.waitForToast('Organization deleted successfully');
    
    // Should not appear in the list anymore
    const orgExists = await page.locator(`[data-testid="organization-row-${orgId}"]`).isVisible();
    expect(orgExists).toBeFalsy();
  });

  test('should search organizations', async ({ page }) => {
    // Create multiple test organizations
    await helpers.createTestOrganization('Searchable Org Alpha');
    await helpers.createTestOrganization('Searchable Org Beta');
    await helpers.createTestOrganization('Different Company');
    
    await helpers.navigateToOrganizations();
    
    // Search for specific organizations
    await helpers.searchInTable('Searchable');
    
    // Should show only matching results
    await helpers.expectElementToContainText('organizations-table', 'Searchable Org Alpha');
    await helpers.expectElementToContainText('organizations-table', 'Searchable Org Beta');
    
    // Should not show non-matching results
    const differentCompanyExists = await page.locator('text=Different Company').isVisible();
    expect(differentCompanyExists).toBeFalsy();
  });

  test('should handle pagination', async ({ page }) => {
    // Create many organizations to test pagination
    const orgPromises = [];
    for (let i = 0; i < 25; i++) {
      orgPromises.push(helpers.createTestOrganization(`Pagination Test Org ${i}`));
    }
    await Promise.all(orgPromises);
    
    await helpers.navigateToOrganizations();
    
    // Should show pagination controls
    await helpers.expectElementToBeVisible('pagination');
    
    // Should show first page
    await helpers.expectElementToContainText('current-page', '1');
    
    // Click next page
    await page.click('[data-testid="next-page-button"]');
    
    // Should show second page
    await helpers.expectElementToContainText('current-page', '2');
    
    // Should show different organizations
    const firstPageOrg = await page.locator('text=Pagination Test Org 0').isVisible();
    expect(firstPageOrg).toBeFalsy();
  });

  test('should validate organization form', async ({ page }) => {
    await helpers.navigateToOrganizations();
    
    // Click create button
    await page.click('[data-testid="create-organization-button"]');
    await helpers.waitForModal('create-organization-modal');
    
    // Try to submit empty form
    await helpers.submitForm('create-organization-form');
    
    // Should show validation errors
    await helpers.expectFormError('name', 'Organization name is required');
    
    // Test name too short
    await helpers.fillForm({
      name: 'AB',
    });
    
    await helpers.submitForm('create-organization-form');
    await helpers.expectFormError('name', 'Name must be at least 3 characters');
    
    // Test name too long
    await helpers.fillForm({
      name: 'A'.repeat(101),
    });
    
    await helpers.submitForm('create-organization-form');
    await helpers.expectFormError('name', 'Name must be less than 100 characters');
  });

  test('should handle organization details view', async ({ page }) => {
    // Create test organization
    const orgId = await helpers.createTestOrganization('Details Test Org');
    
    await helpers.navigateToOrganizations();
    
    // Click on organization row to view details
    await page.click(`[data-testid="organization-row-${orgId}"]`);
    
    // Should navigate to organization details page
    await helpers.expectToBeOnPage(`/organizations/${orgId}`);
    
    // Should show organization details
    await helpers.expectElementToBeVisible('organization-details');
    await helpers.expectElementToContainText('organization-name', 'Details Test Org');
    
    // Should show related data sections
    await helpers.expectElementToBeVisible('organization-contacts');
    await helpers.expectElementToBeVisible('organization-products');
    await helpers.expectElementToBeVisible('organization-invoices');
  });

  test('should handle organization switching', async ({ page }) => {
    // Create multiple organizations
    const org1Id = await helpers.createTestOrganization('Organization One');
    const org2Id = await helpers.createTestOrganization('Organization Two');
    
    await helpers.navigateToOrganizations();
    
    // Switch to first organization
    await page.click(`[data-testid="switch-to-org-${org1Id}"]`);
    
    // Should show organization context
    await helpers.expectElementToContainText('current-organization', 'Organization One');
    
    // Switch to second organization
    await page.click('[data-testid="organization-switcher"]');
    await page.click(`[data-testid="switch-to-org-${org2Id}"]`);
    
    // Should update organization context
    await helpers.expectElementToContainText('current-organization', 'Organization Two');
  });

  test('should handle organization permissions', async ({ page }) => {
    // Create test organization as admin
    const orgId = await helpers.createTestOrganization('Permission Test Org');
    
    // Logout and login as regular user
    await helpers.logout();
    await helpers.loginAsUser();
    
    await helpers.navigateToOrganizations();
    
    // Regular user should not see admin actions
    const deleteButtonExists = await page.locator(`[data-testid="delete-organization-${orgId}"]`).isVisible();
    expect(deleteButtonExists).toBeFalsy();
    
    // Should not be able to create organizations
    const createButtonExists = await page.locator('[data-testid="create-organization-button"]').isVisible();
    expect(createButtonExists).toBeFalsy();
  });
});