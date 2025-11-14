import { Page } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { BasePage } from './base-page';

export class ContactsPage extends BasePage {
  constructor(page: Page, apiClient: ApiClient) {
    super(page, apiClient);
  }

  async goto(organizationId?: number): Promise<void> {
    if (organizationId) {
      await this.navigateTo(`/organizations/${organizationId}/contacts`);
    } else {
      await this.click('nav-contacts');
      await this.page.waitForURL('/contacts');
    }
  }
}
