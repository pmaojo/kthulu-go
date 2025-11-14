import { Page } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { BasePage } from './base-page';

export class OrganizationsPage extends BasePage {
  constructor(page: Page, apiClient: ApiClient) {
    super(page, apiClient);
  }

  async goto(): Promise<void> {
    await this.click('nav-organizations');
    await this.page.waitForURL('/organizations');
  }
}
