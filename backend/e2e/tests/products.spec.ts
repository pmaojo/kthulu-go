import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { ProductsPage } from '../pages/products-page';

test.describe('Products Management', () => {
  let apiClient: ApiClient;
  let productsPage: ProductsPage;
  let organizationId: number;

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    productsPage = new ProductsPage(page, apiClient);

    // Login and create test organization
    await productsPage.loginAsAdmin();
    organizationId = await productsPage.createTestOrganization('Products Test Org');

    await productsPage.goto(organizationId);
  });

  test('should display products list', async () => {
    await productsPage.goto();

    // Should show products page
    await productsPage.expectElementToBeVisible('products-page');
    await productsPage.expectElementToBeVisible('create-product-button');

    // Should show table or empty state
    const hasProducts = await productsPage.locator('products-table').isVisible();
    if (!hasProducts) {
      await productsPage.expectElementToBeVisible('empty-state');
    }
  });

  test('should create a new product', async () => {
    await productsPage.goto();

    // Click create button
    await productsPage.click('create-product-button');
    await productsPage.waitForModal('create-product-modal');
    
    // Fill form
    const productName = `Test Product ${Date.now()}`;
    const sku = `TEST-${Date.now()}`;
    await productsPage.fillForm({
      name: productName,
      sku: sku,
      description: 'Test product created by E2E test',
      price: '99.99',
      category: 'Electronics',
    });
    
    // Submit form
    await productsPage.submitForm('create-product-form');
    
    // Should close modal and show success message
    await productsPage.waitForToast('Product created successfully');
    
    // Should appear in the list
    await productsPage.expectElementToContainText('products-table', productName);
    await productsPage.expectElementToContainText('products-table', sku);
  });

  test('should edit an existing product', async () => {
    // Create test product first
    const productId = await productsPage.createTestProduct(organizationId, 'Edit Test Product');
    
    await productsPage.goto();
    
    // Find and click edit button
    await productsPage.click(`edit-product-${productId}`);
    await productsPage.waitForModal('edit-product-modal');
    
    // Update form
    const updatedName = `Updated Product ${Date.now()}`;
    await productsPage.fillForm({
      name: updatedName,
      description: 'Updated description',
      price: '149.99',
    });
    
    await productsPage.submitForm('edit-product-form');
    
    await productsPage.waitForToast('Product updated successfully');
    await productsPage.expectElementToContainText('products-table', updatedName);
  });

  test('should delete a product', async () => {
    // Create test product first
    const productId = await productsPage.createTestProduct(organizationId, 'Delete Test Product');
    
    await productsPage.goto();
    
    // Click delete button
    await productsPage.click(`delete-product-${productId}`);
    await productsPage.waitForModal('confirm-delete-modal');
    
    // Confirm deletion
    await productsPage.click('confirm-delete-button');
    
    await productsPage.waitForToast('Product deleted successfully');
    
    // Should not appear in the list anymore
    const productExists = await productsPage.locator(`product-row-${productId}`).isVisible();
    expect(productExists).toBeFalsy();
  });

  test('should search products', async () => {
    // Create multiple test products
    await productsPage.createTestProduct(organizationId, 'Searchable Product Alpha');
    await productsPage.createTestProduct(organizationId, 'Searchable Product Beta');
    await productsPage.createTestProduct(organizationId, 'Different Item');
    
    await productsPage.goto();
    
    // Search for specific products
    await productsPage.searchInTable('Searchable');
    
    // Should show only matching results
    await productsPage.expectElementToContainText('products-table', 'Searchable Product Alpha');
    await productsPage.expectElementToContainText('products-table', 'Searchable Product Beta');
    
    // Should not show non-matching results
    const differentExists = await productsPage.locatorByText('Different Item').isVisible();
    expect(differentExists).toBeFalsy();
  });

  test('should filter products by category', async () => {
    // Create products in different categories
    await apiClient.createProduct(organizationId, {
      name: 'Electronics Product',
      sku: 'ELEC-001',
      description: 'Electronics item',
      isActive: true,
    });
    
    await apiClient.createProduct(organizationId, {
      name: 'Clothing Product',
      sku: 'CLOTH-001',
      description: 'Clothing item',
      isActive: true,
    });
    
    await productsPage.goto();

    // Filter by electronics
    await productsPage.selectOption('category-filter', 'Electronics');
    await productsPage.waitForTimeout(500);
    
    // Should show only electronics products
    await productsPage.expectElementToContainText('products-table', 'Electronics Product');
    
    const clothingExists = await productsPage.locatorByText('Clothing Product').isVisible();
    expect(clothingExists).toBeFalsy();
  });

  test('should handle product details view', async () => {
    // Create test product
    const productId = await productsPage.createTestProduct(organizationId, 'Details Test Product');
    
    await productsPage.goto();
    
    // Click on product row to view details
    await productsPage.click(`product-row-${productId}`);
    
    // Should navigate to product details page
    await productsPage.expectToBeOnPage(`/products/${productId}`);
    
    // Should show product details
    await productsPage.expectElementToBeVisible('product-details');
    await productsPage.expectElementToContainText('product-name', 'Details Test Product');
    
    // Should show product information sections
    await productsPage.expectElementToBeVisible('product-info');
    await productsPage.expectElementToBeVisible('product-pricing');
    await productsPage.expectElementToBeVisible('product-inventory');
  });

  test('should manage product variants', async () => {
    // Create test product
    const productId = await productsPage.createTestProduct(organizationId, 'Variant Test Product');

    await productsPage.navigateTo(`/products/${productId}`);

    // Click add variant button
    await productsPage.click('add-variant-button');
    await productsPage.waitForModal('add-variant-modal');
    
    // Fill variant form
    await productsPage.fillForm({
      name: 'Size Large',
      sku: 'TEST-L',
      price: '109.99',
      'attribute-name': 'Size',
      'attribute-value': 'Large',
    });
    
    await productsPage.submitForm('add-variant-form');
    
    await productsPage.waitForToast('Product variant added successfully');
    await productsPage.expectElementToContainText('product-variants', 'Size Large');
  });

  test('should manage product pricing', async () => {
    // Create test product
    const productId = await productsPage.createTestProduct(organizationId, 'Pricing Test Product');

    await productsPage.navigateTo(`/products/${productId}`);

    // Click add price button
    await productsPage.click('add-price-button');
    await productsPage.waitForModal('add-price-modal');
    
    // Fill pricing form
    await productsPage.fillForm({
      'price-type': 'wholesale',
      amount: '79.99',
      'min-quantity': '10',
      currency: 'EUR',
    });
    
    await productsPage.submitForm('add-price-form');
    
    await productsPage.waitForToast('Product price added successfully');
    await productsPage.expectElementToContainText('product-pricing', 'wholesale');
    await productsPage.expectElementToContainText('product-pricing', '79.99');
  });

  test('should validate product form', async () => {
    await productsPage.goto();

    await productsPage.click('create-product-button');
    await productsPage.waitForModal('create-product-modal');
    
    // Try to submit empty form
    await productsPage.submitForm('create-product-form');
    
    // Should show validation errors
    await productsPage.expectFormError('name', 'Product name is required');
    await productsPage.expectFormError('sku', 'SKU is required');
    
    // Test duplicate SKU
    const existingProduct = await productsPage.createTestProduct(organizationId, 'Existing Product');
    const existingSku = `TEST-${Date.now() - 1000}`;
    
    await productsPage.fillForm({
      name: 'New Product',
      sku: existingSku,
    });
    
    await productsPage.submitForm('create-product-form');
    await productsPage.expectFormError('sku', 'SKU already exists');
    
    // Test invalid price
    await productsPage.fillForm({
      name: 'Valid Product',
      sku: `VALID-${Date.now()}`,
      price: 'invalid-price',
    });
    
    await productsPage.submitForm('create-product-form');
    await productsPage.expectFormError('price', 'Invalid price format');
  });

  test('should handle product status toggle', async () => {
    // Create test product
    const productId = await productsPage.createTestProduct(organizationId, 'Status Test Product');
    
    await productsPage.goto();
    
    // Product should be active by default
    await productsPage.expectElementToBeVisible(`[data-testid="product-status-${productId}"][data-active="true"]`);
    
    // Click status toggle
    await productsPage.click(`toggle-product-status-${productId}`);
    
    await productsPage.waitForToast('Product status updated');
    
    // Should now be inactive
    await productsPage.expectElementToBeVisible(`[data-testid="product-status-${productId}"][data-active="false"]`);
  });

  test('should bulk update products', async () => {
    // Create multiple test products
    const product1Id = await productsPage.createTestProduct(organizationId, 'Bulk Product 1');
    const product2Id = await productsPage.createTestProduct(organizationId, 'Bulk Product 2');
    
    await productsPage.goto();
    
    // Select multiple products
    await productsPage.check(`select-product-${product1Id}`);
    await productsPage.check(`select-product-${product2Id}`);
    
    // Click bulk actions button
    await productsPage.click('bulk-actions-button');
    await productsPage.click('bulk-update-prices');
    
    await productsPage.waitForModal('bulk-update-modal');
    
    // Update prices
    await productsPage.fillForm({
      'price-adjustment': '10',
      'adjustment-type': 'percentage',
    });
    
    await productsPage.submitForm('bulk-update-form');
    
    await productsPage.waitForToast('Products updated successfully');
  });

  test('should export products', async () => {
    // Create some test products
    await productsPage.createTestProduct(organizationId, 'Export Product 1');
    await productsPage.createTestProduct(organizationId, 'Export Product 2');
    
    await productsPage.goto();
    
    // Click export button
    const downloadPromise = productsPage.page.waitForEvent('download');
    await productsPage.click('export-products-button');
    const download = await downloadPromise;
    
    // Should download a file
    expect(download.suggestedFilename()).toMatch(/products.*\.(csv|xlsx)$/);
  });

  test('should import products', async () => {
    await productsPage.goto();

    // Click import button
    await productsPage.click('import-products-button');
    await productsPage.waitForModal('import-products-modal');
    
    // Upload CSV file (mock file upload)
    const fileInput = productsPage.locator('file-input');
    await fileInput.setInputFiles({
      name: 'products.csv',
      mimeType: 'text/csv',
      buffer: Buffer.from('name,sku,price\nImported Product,IMP-001,99.99'),
    });
    
    await productsPage.submitForm('import-products-form');
    
    await productsPage.waitForToast('Products imported successfully');
    await productsPage.expectElementToContainText('products-table', 'Imported Product');
  });
});