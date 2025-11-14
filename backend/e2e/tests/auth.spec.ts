import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { TestHelpers } from '../utils/test-helpers';

test.describe('Authentication Flow', () => {
  let apiClient: ApiClient;
  let helpers: TestHelpers;

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    helpers = new TestHelpers(page, apiClient);
  });

  test('should register a new user successfully', async ({ page }) => {
    await page.goto('/register');

    // Fill registration form
    await helpers.fillForm({
      email: 'newuser@test.com',
      password: 'TestPassword123!',
      'confirm-password': 'TestPassword123!',
    });

    await helpers.submitForm('register-form');

    // Should redirect to dashboard or confirmation page
    await expect(page).toHaveURL(/\/(dashboard|confirm)/);
    
    // Check for success message
    await helpers.waitForToast('Registration successful');
  });

  test('should login with valid credentials', async ({ page }) => {
    await page.goto('/login');

    await helpers.fillForm({
      email: 'admin@test.com',
      password: 'TestPassword123!',
    });

    await helpers.submitForm('login-form');

    // Should redirect to dashboard
    await helpers.expectToBeOnPage('/dashboard');
    
    // Should show user info
    await helpers.expectElementToBeVisible('user-menu');
    await helpers.expectElementToContainText('user-email', 'admin@test.com');
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await helpers.fillForm({
      email: 'invalid@test.com',
      password: 'wrongpassword',
    });

    await helpers.submitForm('login-form');

    // Should show error message
    await helpers.expectFormError('login-form', 'Invalid credentials');
    
    // Should stay on login page
    await helpers.expectToBeOnPage('/login');
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await helpers.loginAsAdmin();
    
    // Verify we're on dashboard
    await helpers.expectToBeOnPage('/dashboard');
    
    // Logout
    await helpers.logout();
    
    // Should redirect to login
    await helpers.expectToBeOnPage('/login');
    
    // Should not be able to access protected routes
    await page.goto('/dashboard');
    await helpers.expectToBeOnPage('/login');
  });

  test('should handle password reset flow', async ({ page }) => {
    await page.goto('/login');
    
    // Click forgot password link
    await page.click('[data-testid="forgot-password-link"]');
    await helpers.expectToBeOnPage('/forgot-password');
    
    // Enter email
    await helpers.fillForm({
      email: 'admin@test.com',
    });
    
    await helpers.submitForm('forgot-password-form');
    
    // Should show success message
    await helpers.waitForToast('Password reset email sent');
  });

  test('should validate form fields', async ({ page }) => {
    await page.goto('/register');

    // Try to submit empty form
    await helpers.submitForm('register-form');

    // Should show validation errors
    await helpers.expectFormError('email', 'Email is required');
    await helpers.expectFormError('password', 'Password is required');

    // Test invalid email
    await helpers.fillForm({
      email: 'invalid-email',
      password: 'TestPassword123!',
      'confirm-password': 'TestPassword123!',
    });

    await helpers.submitForm('register-form');
    await helpers.expectFormError('email', 'Invalid email format');

    // Test password mismatch
    await helpers.fillForm({
      email: 'test@example.com',
      password: 'TestPassword123!',
      'confirm-password': 'DifferentPassword123!',
    });

    await helpers.submitForm('register-form');
    await helpers.expectFormError('confirm-password', 'Passwords do not match');
  });

  test('should persist authentication across page reloads', async ({ page }) => {
    // Login
    await helpers.loginAsAdmin();
    await helpers.expectToBeOnPage('/dashboard');

    // Reload page
    await page.reload();
    
    // Should still be authenticated
    await helpers.expectToBeOnPage('/dashboard');
    await helpers.expectElementToBeVisible('user-menu');
  });

  test('should handle token expiration gracefully', async ({ page }) => {
    // Login
    await helpers.loginAsAdmin();
    await helpers.expectToBeOnPage('/dashboard');

    // Simulate token expiration by clearing localStorage
    await page.evaluate(() => {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
    });

    // Try to navigate to a protected route
    await page.goto('/organizations');
    
    // Should redirect to login
    await helpers.expectToBeOnPage('/login');
    await helpers.waitForToast('Session expired. Please login again.');
  });

  test('should handle concurrent login sessions', async ({ page, context }) => {
    // Login in first tab
    await helpers.loginAsAdmin();
    await helpers.expectToBeOnPage('/dashboard');

    // Open second tab and login with different user
    const secondPage = await context.newPage();
    const secondHelpers = new TestHelpers(secondPage, new ApiClient(page.url()));
    
    await secondHelpers.loginAsUser();
    await secondHelpers.expectToBeOnPage('/dashboard');

    // Both sessions should be independent
    await helpers.expectElementToContainText('user-email', 'admin@test.com');
    await secondHelpers.expectElementToContainText('user-email', 'user@test.com');

    await secondPage.close();
  });
});