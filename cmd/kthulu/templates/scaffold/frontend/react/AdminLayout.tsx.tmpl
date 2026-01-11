// @kthulu:frontend:layout
import { ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { User, LogOut } from 'lucide-react';
import { navigation } from '../../config/navigation';

interface AdminLayoutProps {
  children: ReactNode;
}

export default function AdminLayout({ children }: AdminLayoutProps) {
  const location = useLocation();

  const isActive = (path: string) => location.pathname === path;

  const navClass = (path: string) => `
    flex items-center gap-3 p-3 rounded-lg transition-all duration-200
    ${isActive(path) 
      ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/20' 
      : 'text-slate-400 hover:bg-slate-800 hover:text-white'}
  `;

  return (
    <div className="min-h-screen flex bg-slate-50">
      {/* Sidebar */}
      <aside className="w-72 bg-slate-900 text-white flex flex-col fixed h-full z-10">
        <div className="p-6 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center font-bold">K</div>
            <span className="text-xl font-bold tracking-tight">Kthulu Admin</span>
          </div>
        </div>

        <nav className="flex-1 p-4 space-y-2 overflow-y-auto">
          {navigation.map((item) => (
            <div key={item.path}>
              {item.category && (
                <div className="px-3 py-2 text-xs font-semibold text-slate-500 uppercase tracking-wider mt-6 first:mt-0">
                  {item.category}
                </div>
              )}
              <Link to={item.path} className={navClass(item.path)}>
                <item.icon size={20} />
                <span className="font-medium">{item.title}</span>
              </Link>
            </div>
          ))}
        </nav>

        <div className="p-4 border-t border-slate-800">
          <div className="flex items-center gap-3 p-3 rounded-lg bg-slate-800/50">
            <div className="w-10 h-10 rounded-full bg-slate-700 flex items-center justify-center">
              <User size={20} className="text-slate-400" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">Admin User</p>
              <p className="text-xs text-slate-400 truncate">admin@kthulu.dev</p>
            </div>
            <button className="p-2 text-slate-400 hover:text-red-400 transition-colors">
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 ml-72">
        <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-8 sticky top-0 z-10 backdrop-blur-sm bg-white/80">
          <div className="flex items-center gap-4">
            <h2 className="text-lg font-semibold text-slate-800">
              {location.pathname === '/' ? 'Dashboard' : location.pathname.substring(1).charAt(0).toUpperCase() + location.pathname.slice(2)}
            </h2>
          </div>
        </header>
        
        <div className="p-8 max-w-7xl mx-auto animate-in fade-in slide-in-from-bottom-4 duration-500">
          {children}
        </div>
      </main>
    </div>
  );
}
