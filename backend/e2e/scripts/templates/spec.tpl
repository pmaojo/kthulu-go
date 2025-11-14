import { test, expect } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { {{PAGE_CLASS}} } from '../pages/{{SLUG}}-page';

test.describe('{{TITLE}}', () => {
  let apiClient: ApiClient;
  let pageObject: {{PAGE_CLASS}};

  test.beforeEach(async ({ page, baseURL }) => {
    apiClient = new ApiClient(baseURL!);
    pageObject = new {{PAGE_CLASS}}(page, apiClient);
    await pageObject.goto();
  });

  test('should have a placeholder test', async () => {
    // TODO: implement test
    await expect(pageObject.page).toBeTruthy();
  });
});
