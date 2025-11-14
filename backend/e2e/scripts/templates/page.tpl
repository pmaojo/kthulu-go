import { Page } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { BasePage } from './base-page';

export class {{PAGE_CLASS}} extends BasePage {
  constructor(page: Page, apiClient: ApiClient) {
    super(page, apiClient);
  }

  async goto(): Promise<void> {
    await this.navigateTo('/{{SLUG}}');
  }
}
