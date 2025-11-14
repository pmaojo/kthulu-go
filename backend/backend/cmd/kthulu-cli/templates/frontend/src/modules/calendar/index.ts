import { lazy } from 'react';
import { registerModule, type Module } from '../registry';

const Calendar = lazy(() => import('@/components/views/calendar/Calendar'));

const module: Module = {
  routes: [
    {
      path: '/calendar',
      Component: Calendar,
    },
  ],
  components: { Calendar },
};

registerModule(module);
