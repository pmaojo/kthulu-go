import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

// Lazy load the main page component
const BookingPage = lazy(() => import('./presentation/BookingPage'));

const module: Module = {
  routes: [
    {
      path: '/bookings',
      Component: BookingPage,
    },
  ],
  components: {},
};

registerModule(module);
