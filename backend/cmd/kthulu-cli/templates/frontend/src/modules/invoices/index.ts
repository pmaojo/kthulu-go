import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

const Invoice = lazy(() => import('@/components/views/invoice/Invoice'));

const module: Module = {
  routes: [
    {
      path: '/invoice/:recordId',
      Component: Invoice,
    },
  ],
  components: {
    Invoice,
  },
};

registerModule(module);
