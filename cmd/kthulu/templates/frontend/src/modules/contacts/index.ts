import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

const Contacts = lazy(() => import('@/components/views/contacts/Contacts'));

const module: Module = {
  routes: [
    {
      path: '/contacts',
      Component: Contacts,
    },
  ],
  components: { Contacts },
};

registerModule(module);
