import { FullConfig } from '@playwright/test';
import { ApiClient } from './utils/api-client';
import { waitForServer, setupTestData, createTestUsers } from './utils/test-data';

async function globalSetup(config: FullConfig) {
  console.log('🚀 Starting E2E test setup...');
  
  const baseURL = config.projects[0].use.baseURL || 'http://localhost:8080';
  const apiClient = new ApiClient(baseURL);
  
  // Wait for the server to be ready
  console.log('⏳ Waiting for server to be ready...');
  await waitForServer(apiClient);
  
  // Setup test data
  console.log('📊 Setting up test data...');
  await setupTestData(apiClient);
  
  // Create admin user for tests
  console.log('👤 Creating test users...');
  await createTestUsers(apiClient);
  
  console.log('✅ E2E test setup completed');
}

export default globalSetup;
