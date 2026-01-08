import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

// Lazy load the main page component
const ReviewPage = lazy(() => import('./presentation/ReviewPage'));

const module: Module = {
  routes: [
    {
      path: '/reviews',
      Component: ReviewPage,
    },
  ],
  components: {},
};

registerModule(module);
