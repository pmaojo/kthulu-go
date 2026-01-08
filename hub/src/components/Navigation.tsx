import Link from 'next/link';
import { Home, Book, ShoppingBag, Github, Menu, X, Zap, Shield, Cloud } from 'lucide-react';

export function Header() {
  return (
    <header className="glass fixed top-0 w-full z-50 px-6 h-16 flex items-center justify-between border-b border-white/10">
      <div className="flex items-center gap-8">
        <Link href="/" className="flex items-center gap-2 font-bold text-xl tracking-tighter kthulu-glow">
          <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-primary-foreground">K</div>
          KTHULU HUB
        </Link>
        <nav className="hidden md:flex items-center gap-6 text-sm font-medium">
          <Link href="/docs" className="hover:text-primary transition-colors flex items-center gap-2">
            <Book size={16} /> Docs
          </Link>
          <Link href="/marketplace" className="hover:text-primary transition-colors flex items-center gap-2">
            <ShoppingBag size={16} /> Marketplace
          </Link>
        </nav>
      </div>
      <div className="flex items-center gap-4">
        <a 
          href="https://github.com/pmaojo/kthulu-go" 
          target="_blank" 
          rel="noopener noreferrer"
          className="p-2 hover:bg-white/10 rounded-full transition-colors"
        >
          <Github size={20} />
        </a>
        <button className="md:hidden p-2">
          <Menu size={20} />
        </button>
      </div>
    </header>
  );
}

export function Sidebar() {
  return (
    <aside className="fixed left-0 top-16 w-64 h-[calc(100vh-64px)] overflow-y-auto border-r border-white/5 bg-background/50 backdrop-blur-sm hidden md:block p-6">
      <div className="space-y-8">
        <div>
          <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-widest mb-4">Guides</h3>
          <ul className="space-y-2 text-sm">
            <li>
              <Link href="/docs/guide/introduction" className="hover:text-primary text-primary transition-colors">
                Introduction
              </Link>
            </li>
            <li>
              <Link href="/docs/guide/installation" className="hover:text-primary transition-colors">
                Installation
              </Link>
            </li>
          </ul>
        </div>
        <div>
          <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-widest mb-4">Marketplace</h3>
          <ul className="space-y-2 text-sm">
            <li>
              <Link href="/marketplace?type=starters" className="hover:text-primary transition-colors flex items-center gap-2">
                <Zap size={14} /> Starters
              </Link>
            </li>
            <li>
              <Link href="/marketplace?type=modules" className="hover:text-primary transition-colors flex items-center gap-2">
                <Shield size={14} /> Modules
              </Link>
            </li>
            <li>
              <Link href="/marketplace?type=plugins" className="hover:text-primary transition-colors flex items-center gap-2">
                <Cloud size={14} /> Plugins
              </Link>
            </li>
          </ul>
        </div>
      </div>
    </aside>
  );
}
