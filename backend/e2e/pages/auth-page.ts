import { Page } from '@playwright/test';
import { ApiClient } from '../utils/api-client';
import { BasePage } from './base-page';

export class AuthPage extends BasePage {
  constructor(page: Page, apiClient: ApiClient) {
    super(page, apiClient);
  }

  async gotoLogin(): Promise<void> {
    await this.navigateTo('/login');
  }

  async gotoRegister(): Promise<void> {
    await this.navigateTo('/register');
  }
}
