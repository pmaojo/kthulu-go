import { FullConfig } from '@playwright/test';
import { ApiClient } from './utils/api-client';

async function globalTeardown(config: FullConfig) {
  console.log('🧹 Starting E2E test teardown...');
  
  const baseURL = config.projects[0].use.baseURL || 'http://localhost:8080';
  const apiClient = new ApiClient(baseURL);
  
  try {
    // Clean up test data
    console.log('🗑️ Cleaning up test data...');
    await apiClient.resetDatabase();
    
    console.log('✅ E2E test teardown completed');
  } catch (error) {
    console.warn('⚠️ Teardown failed:', error);
  }
}

export default globalTeardown;