import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

const Products = lazy(() => import('@/components/views/products/Products'));

const module: Module = {
  routes: [
    {
      path: '/products',
      Component: Products,
    },
  ],
  components: { Products },
};

registerModule(module);
