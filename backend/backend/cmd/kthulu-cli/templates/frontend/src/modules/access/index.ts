import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

const Access = lazy(() => import('@/components/views/access/Access'));

const module: Module = {
  routes: [
    {
      path: '/access',
      Component: Access,
    },
  ],
  components: { Access },
};

registerModule(module);
