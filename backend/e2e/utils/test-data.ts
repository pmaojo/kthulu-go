import { ApiClient } from './api-client';

/**
 * Basic shape of a user generated for E2E tests.
 * Extend this interface or create additional ones when seeding
 * new entity types.
 */
export interface TestUser {
  email: string;
  password: string;
  role: string;
}

/**
 * Wait until the application server responds to health checks.
 * Retries the request a limited number of times before failing.
 */
export async function waitForServer(apiClient: ApiClient, maxRetries = 30): Promise<void> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      await apiClient.healthCheck();
      console.log('✅ Server is ready');
      return;
    } catch (error) {
      console.log(`⏳ Waiting for server... (${i + 1}/${maxRetries})`);
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  }
  throw new Error('Server failed to start within timeout');
}

/**
 * Prepare the database for testing.
 * Currently resets the database to a clean state, but this
 * function can be extended to seed other shared data such as
 * organizations, products, etc.
 */
export async function setupTestData(apiClient: ApiClient): Promise<void> {
  try {
    await apiClient.resetDatabase();
    console.log('🗄️ Database reset completed');
  } catch (error) {
    console.warn('⚠️ Database reset failed, continuing...', error);
  }
}

const defaultTestUsers: TestUser[] = [
  {
    email: 'admin@test.com',
    password: 'TestPassword123!',
    role: 'admin',
  },
  {
    email: 'user@test.com',
    password: 'TestPassword123!',
    role: 'user',
  },
  {
    email: 'manager@test.com',
    password: 'TestPassword123!',
    role: 'manager',
  },
];

/**
 * Create a list of test users in the system.
 * Pass a custom list to generate additional roles or entities.
 */
export async function createTestUsers(apiClient: ApiClient, users: TestUser[] = defaultTestUsers): Promise<void> {
  for (const user of users) {
    try {
      await apiClient.register(user.email, user.password);
      console.log(`✅ Created test user: ${user.email}`);
    } catch (error) {
      console.warn(`⚠️ Failed to create user ${user.email}:`, error);
    }
  }
}
