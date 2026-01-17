import { test, expect } from '@playwright/test';

test.describe('Todo Module', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the module page
    await page.goto('/todo');
  });

  test('should display empty state or list', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 2 })).toHaveText('Todos');
  });

  test('should open create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Todo' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { level: 3, name: 'Create Todo' })).toBeVisible();
  });

  test('should create a new todo', async ({ page }) => {
    await page.getByRole('button', { name: 'Add Todo' }).click();
    
    // Fill form fields
    
    await page.getByLabel('Title').fill('Test title');
    
    await page.getByLabel('Completed').fill('Test completed');
    

    await page.getByRole('button', { name: 'Save' }).click();

    // Verify it appears in the list
    await expect(page.getByText('Test title')).toBeVisible();
  });
});
