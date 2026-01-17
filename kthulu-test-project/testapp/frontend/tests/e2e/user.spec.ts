import { test, expect } from '@playwright/test';

test.describe('User Module', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the module page
    await page.goto('/user');
  });

  test('should display empty state or list', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 2 })).toHaveText('Users');
  });

  test('should open create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Add User' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { level: 3, name: 'Create User' })).toBeVisible();
  });

  test('should create a new user', async ({ page }) => {
    await page.getByRole('button', { name: 'Add User' }).click();
    
    // Fill form fields
    
    await page.getByLabel('Name').fill('Test name');
    

    await page.getByRole('button', { name: 'Save' }).click();

    // Verify it appears in the list
    await expect(page.getByText('Test name')).toBeVisible();
  });
});
