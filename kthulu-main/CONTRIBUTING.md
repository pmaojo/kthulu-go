# Contributing

Thank you for considering a contribution to Kthulu! This document outlines a few
guidelines to help you get started.

## Testing with fixtures

To simplify unit tests in the frontend, reusable fixtures are provided at
`frontend/src/data/fixtures.ts`. These fixtures contain sample users,
organizations and roles that can be imported directly into test files.

```ts
import { users, organizations, getUserFixture } from '@/data/fixtures';

const user = getUserFixture(1);
```

Use these predefined objects instead of manually crafting mock data in each
test. This keeps tests concise and consistent across the project.

