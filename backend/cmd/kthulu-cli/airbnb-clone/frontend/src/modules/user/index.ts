import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

// Lazy load the main page component
const UserPage = lazy(() => import('./presentation/UserPage'));

const module: Module = {
  routes: [
    {
      path: '/users',
      Component: UserPage,
    },
  ],
  components: {},
};

registerModule(module);
