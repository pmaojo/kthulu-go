import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { ContactsPage } from '../pages/contacts-page';

test.describe('Contacts Management', () => {
  let apiClient: ApiClient;
  let contactsPage: ContactsPage;
  let organizationId: number;

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    contactsPage = new ContactsPage(page, apiClient);

    // Login and create test organization
    await contactsPage.loginAsAdmin();
    organizationId = await contactsPage.createTestOrganization('Contacts Test Org');

    await contactsPage.goto(organizationId);
  });

  test('should display contacts list', async () => {
    await contactsPage.goto();

    // Should show contacts page
    await contactsPage.expectElementToBeVisible('contacts-page');
    await contactsPage.expectElementToBeVisible('create-contact-button');

    // Should show table or empty state
    const hasContacts = await contactsPage.locator('contacts-table').isVisible();
    if (!hasContacts) {
      await contactsPage.expectElementToBeVisible('empty-state');
    }
  });

  test('should create a new customer contact', async () => {
    await contactsPage.goto();

    // Click create button
    await contactsPage.click('create-contact-button');
    await contactsPage.waitForModal('create-contact-modal');
    
    // Fill form for customer
    const contactName = `Test Customer ${Date.now()}`;
    await contactsPage.fillForm({
      'company-name': contactName,
      email: 'customer@test.com',
      phone: '+1234567890',
      type: 'customer',
    });
    
    // Submit form
    await contactsPage.submitForm('create-contact-form');
    
    // Should close modal and show success message
    await contactsPage.waitForToast('Contact created successfully');
    
    // Should appear in the list
    await contactsPage.expectElementToContainText('contacts-table', contactName);
    await contactsPage.expectElementToContainText('contacts-table', 'customer@test.com');
  });

  test('should create a new supplier contact', async () => {
    await contactsPage.goto();

    await contactsPage.click('create-contact-button');
    await contactsPage.waitForModal('create-contact-modal');
    
    // Fill form for supplier
    const supplierName = `Test Supplier ${Date.now()}`;
    await contactsPage.fillForm({
      'company-name': supplierName,
      email: 'supplier@test.com',
      phone: '+1234567891',
      type: 'supplier',
    });
    
    await contactsPage.submitForm('create-contact-form');
    
    await contactsPage.waitForToast('Contact created successfully');
    await contactsPage.expectElementToContainText('contacts-table', supplierName);
  });

  test('should create a contact with individual name', async () => {
    await contactsPage.goto();

    await contactsPage.click('create-contact-button');
    await contactsPage.waitForModal('create-contact-modal');
    
    // Fill form with individual name instead of company
    await contactsPage.fillForm({
      'first-name': 'John',
      'last-name': 'Doe',
      email: 'john.doe@test.com',
      phone: '+1234567892',
      type: 'lead',
    });
    
    await contactsPage.submitForm('create-contact-form');
    
    await contactsPage.waitForToast('Contact created successfully');
    await contactsPage.expectElementToContainText('contacts-table', 'John Doe');
  });

  test('should edit an existing contact', async () => {
    // Create test contact first
    const contactId = await contactsPage.createTestContact(organizationId, 'Edit Test Contact');
    
    await contactsPage.goto();
    
    // Find and click edit button
    await contactsPage.click(`edit-contact-${contactId}`);
    await contactsPage.waitForModal('edit-contact-modal');
    
    // Update form
    const updatedName = `Updated Contact ${Date.now()}`;
    await contactsPage.fillForm({
      'company-name': updatedName,
      email: 'updated@test.com',
    });
    
    await contactsPage.submitForm('edit-contact-form');
    
    await contactsPage.waitForToast('Contact updated successfully');
    await contactsPage.expectElementToContainText('contacts-table', updatedName);
  });

  test('should delete a contact', async () => {
    // Create test contact first
    const contactId = await contactsPage.createTestContact(organizationId, 'Delete Test Contact');
    
    await contactsPage.goto();
    
    // Click delete button
    await contactsPage.click(`delete-contact-${contactId}`);
    await contactsPage.waitForModal('confirm-delete-modal');
    
    // Confirm deletion
    await contactsPage.click('confirm-delete-button');
    
    await contactsPage.waitForToast('Contact deleted successfully');
    
    // Should not appear in the list anymore
    const contactExists = await contactsPage.locator(`contact-row-${contactId}`).isVisible();
    expect(contactExists).toBeFalsy();
  });

  test('should filter contacts by type', async () => {
    // Create contacts of different types
    await contactsPage.createTestContact(organizationId, 'Customer Contact');
    await apiClient.createContact(organizationId, {
      companyName: 'Supplier Contact',
      email: 'supplier@test.com',
      type: 'supplier',
      isActive: true,
    });
    await apiClient.createContact(organizationId, {
      companyName: 'Lead Contact',
      email: 'lead@test.com',
      type: 'lead',
      isActive: true,
    });
    
    await contactsPage.goto();

    // Filter by customer
    await contactsPage.selectOption('type-filter', 'customer');
    await contactsPage.waitForTimeout(500);

    // Should show only customers
    await contactsPage.expectElementToContainText('contacts-table', 'Customer Contact');

    const supplierExists = await contactsPage.locatorByText('Supplier Contact').isVisible();
    expect(supplierExists).toBeFalsy();

    // Filter by supplier
    await contactsPage.selectOption('type-filter', 'supplier');
    await contactsPage.waitForTimeout(500);

    await contactsPage.expectElementToContainText('contacts-table', 'Supplier Contact');

    const customerExists = await contactsPage.locatorByText('Customer Contact').isVisible();
    expect(customerExists).toBeFalsy();
  });

  test('should search contacts', async () => {
    // Create multiple test contacts
    await contactsPage.createTestContact(organizationId, 'Searchable Contact Alpha');
    await contactsPage.createTestContact(organizationId, 'Searchable Contact Beta');
    await contactsPage.createTestContact(organizationId, 'Different Company');
    
    await contactsPage.goto();
    
    // Search for specific contacts
    await contactsPage.searchInTable('Searchable');
    
    // Should show only matching results
    await contactsPage.expectElementToContainText('contacts-table', 'Searchable Contact Alpha');
    await contactsPage.expectElementToContainText('contacts-table', 'Searchable Contact Beta');
    
    // Should not show non-matching results
    const differentExists = await contactsPage.locatorByText('').isVisible();
    expect(differentExists).toBeFalsy();
  });

  test('should handle contact details view', async () => {
    // Create test contact
    const contactId = await contactsPage.createTestContact(organizationId, 'Details Test Contact');
    
    await contactsPage.goto();
    
    // Click on contact row to view details
    await contactsPage.click(`contact-row-${contactId}`);
    
    // Should navigate to contact details page
    await contactsPage.expectToBeOnPage(`/contacts/${contactId}`);
    
    // Should show contact details
    await contactsPage.expectElementToBeVisible('contact-details');
    await contactsPage.expectElementToContainText('contact-name', 'Details Test Contact');
    
    // Should show contact information sections
    await contactsPage.expectElementToBeVisible('contact-info');
    await contactsPage.expectElementToBeVisible('contact-addresses');
    await contactsPage.expectElementToBeVisible('contact-phones');
  });

  test('should add contact address', async () => {
    // Create test contact
    const contactId = await contactsPage.createTestContact(organizationId, 'Address Test Contact');
    
    await contactsPage.navigateTo(`/contacts/${contactId}`);

    // Click add address button
    await contactsPage.click('add-address-button');
    await contactsPage.waitForModal('add-address-modal');
    
    // Fill address form
    await contactsPage.fillForm({
      'address-line-1': '123 Test Street',
      'address-line-2': 'Suite 100',
      city: 'Test City',
      state: 'Test State',
      'postal-code': '12345',
      country: 'Test Country',
      type: 'billing',
    });
    
    await contactsPage.submitForm('add-address-form');
    
    await contactsPage.waitForToast('Address added successfully');
    await contactsPage.expectElementToContainText('contact-addresses', '123 Test Street');
  });

  test('should add contact phone', async () => {
    // Create test contact
    const contactId = await contactsPage.createTestContact(organizationId, 'Phone Test Contact');
    
    await contactsPage.navigateTo(`/contacts/${contactId}`);

    // Click add phone button
    await contactsPage.click('add-phone-button');
    await contactsPage.waitForModal('add-phone-modal');
    
    // Fill phone form
    await contactsPage.fillForm({
      number: '+1234567890',
      extension: '123',
      type: 'work',
    });
    
    await contactsPage.submitForm('add-phone-form');
    
    await contactsPage.waitForToast('Phone added successfully');
    await contactsPage.expectElementToContainText('contact-phones', '+1234567890');
  });

  test('should validate contact form', async () => {
    await contactsPage.goto();

    await contactsPage.click('create-contact-button');
    await contactsPage.waitForModal('create-contact-modal');
    
    // Try to submit empty form
    await contactsPage.submitForm('create-contact-form');
    
    // Should show validation errors
    await contactsPage.expectFormError('type', 'Contact type is required');
    
    // Test invalid email
    await contactsPage.fillForm({
      'company-name': 'Test Company',
      email: 'invalid-email',
      type: 'customer',
    });
    
    await contactsPage.submitForm('create-contact-form');
    await contactsPage.expectFormError('email', 'Invalid email format');
    
    // Test missing name (both company and individual)
    await contactsPage.fillForm({
      email: 'test@example.com',
      type: 'customer',
    });
    
    await contactsPage.submitForm('create-contact-form');
    await contactsPage.expectFormError('name', 'Either company name or first/last name is required');
  });

  test('should handle contact status toggle', async () => {
    // Create test contact
    const contactId = await contactsPage.createTestContact(organizationId, 'Status Test Contact');
    
    await contactsPage.goto();
    
    // Contact should be active by default
    await contactsPage.expectElementToBeVisible(`[data-testid="contact-status-${contactId}"][data-active="true"]`);
    
    // Click status toggle
    await contactsPage.click(`toggle-contact-status-${contactId}`);
    
    await contactsPage.waitForToast('Contact status updated');
    
    // Should now be inactive
    await contactsPage.expectElementToBeVisible(`[data-testid="contact-status-${contactId}"][data-active="false"]`);
  });

  test('should export contacts', async () => {
    // Create some test contacts
    await contactsPage.createTestContact(organizationId, 'Export Contact 1');
    await contactsPage.createTestContact(organizationId, 'Export Contact 2');
    
    await contactsPage.goto();

    // Click export button
    const downloadPromise = contactsPage.page.waitForEvent('download');
    await contactsPage.click('export-contacts-button');
    const download = await downloadPromise;
    
    // Should download a file
    expect(download.suggestedFilename()).toMatch(/contacts.*\.(csv|xlsx)$/);
  });
});