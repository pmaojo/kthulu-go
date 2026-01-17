import { LayoutDashboard, Users, Settings } from 'lucide-react';

export interface NavItem {
  title: string;
  path: string;
  icon: any;
  category?: string;
}

export const navigation: NavItem[] = [
  {
    title: 'Dashboard',
    path: '/',
    icon: LayoutDashboard,
    category: 'Overview'
  },
  {
    title: 'Users',
    path: '/users',
    icon: Users,
    category: 'Management'
  },
  {
    title: 'Settings',
    path: '/settings',
    icon: Settings,
    category: 'System'
  },
];
