import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

// Lazy load the main page component
const PropertyPage = lazy(() => import('./presentation/PropertyPage'));

const module: Module = {
  routes: [
    {
      path: '/properties',
      Component: PropertyPage,
    },
  ],
  components: {},
};

registerModule(module);
