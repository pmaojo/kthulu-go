import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

const Inventory = lazy(() => import('@/components/views/inventory/Inventory'));

const module: Module = {
  routes: [
    {
      path: '/inventory',
      Component: Inventory,
    },
  ],
  components: { Inventory },
};

registerModule(module);
