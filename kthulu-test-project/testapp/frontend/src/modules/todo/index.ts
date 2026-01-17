import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

// Lazy load the main page component
const TodoPage = lazy(() => import('./presentation/TodoPage'));

const module: Module = {
  routes: [
    {
      path: '/todos',
      Component: TodoPage,
    },
  ],
  components: {},
};

registerModule(module);
