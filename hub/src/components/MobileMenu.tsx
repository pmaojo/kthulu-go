"use client";

import { useState } from "react";
import Link from "next/link";
import { Menu, X, Book, ShoppingBag, Zap, Shield, Cloud } from "lucide-react";

export function MobileMenu() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <button
        className="md:hidden p-2 hover:bg-white/10 rounded-full transition-colors"
        onClick={() => setIsOpen(true)}
        aria-label="Open menu"
      >
        <Menu size={20} />
      </button>

      {/* Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 md:hidden"
          onClick={() => setIsOpen(false)}
        />
      )}

      {/* Drawer */}
      <div
        style={{ backgroundColor: 'hsl(var(--background))' }}
        className={`fixed top-0 right-0 w-3/4 max-w-sm h-full border-l border-white/10 shadow-2xl z-50 transform transition-transform duration-300 ease-in-out md:hidden ${
          isOpen ? "translate-x-0" : "translate-x-full"
        }`}
      >
        <div className="p-6 h-full flex flex-col">
          <div className="flex justify-between items-center mb-8">
            <span className="font-bold text-lg tracking-tight kthulu-glow">Menu</span>
            <button
              onClick={() => setIsOpen(false)}
              className="p-2 hover:bg-white/10 rounded-full transition-colors"
              aria-label="Close menu"
            >
              <X size={20} />
            </button>
          </div>

          <nav className="flex flex-col gap-6">
            <div>
              <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-widest mb-4">Navigation</h3>
              <ul className="space-y-4 text-base font-medium">
                <li>
                  <Link
                    href="/docs"
                    className="flex items-center gap-3 hover:text-primary transition-colors"
                    onClick={() => setIsOpen(false)}
                  >
                    <Book size={18} /> Docs
                  </Link>
                </li>
                <li>
                  <Link
                    href="/marketplace"
                    className="flex items-center gap-3 hover:text-primary transition-colors"
                    onClick={() => setIsOpen(false)}
                  >
                    <ShoppingBag size={18} /> Marketplace
                  </Link>
                </li>
              </ul>
            </div>

            <div className="h-px bg-white/10 w-full" />

            <div>
              <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-widest mb-4">Guides</h3>
              <ul className="space-y-4 text-sm">
                <li>
                  <Link
                    href="/docs/guide/introduction"
                    className="hover:text-primary text-primary transition-colors block"
                    onClick={() => setIsOpen(false)}
                  >
                    Introduction
                  </Link>
                </li>
                <li>
                  <Link
                    href="/docs/guide/installation"
                    className="hover:text-primary transition-colors block"
                    onClick={() => setIsOpen(false)}
                  >
                    Installation
                  </Link>
                </li>
              </ul>
            </div>

            <div>
              <h3 className="text-xs font-bold text-muted-foreground uppercase tracking-widest mb-4">Marketplace</h3>
              <ul className="space-y-4 text-sm">
                <li>
                  <Link
                    href="/marketplace?type=starters"
                    className="flex items-center gap-3 hover:text-primary transition-colors"
                    onClick={() => setIsOpen(false)}
                  >
                    <Zap size={16} /> Starters
                  </Link>
                </li>
                <li>
                  <Link
                    href="/marketplace?type=modules"
                    className="flex items-center gap-3 hover:text-primary transition-colors"
                    onClick={() => setIsOpen(false)}
                  >
                    <Shield size={16} /> Modules
                  </Link>
                </li>
                <li>
                  <Link
                    href="/marketplace?type=plugins"
                    className="flex items-center gap-3 hover:text-primary transition-colors"
                    onClick={() => setIsOpen(false)}
                  >
                    <Cloud size={16} /> Plugins
                  </Link>
                </li>
              </ul>
            </div>
          </nav>

          <div className="mt-auto pt-6 border-t border-white/10">
            <p className="text-xs text-muted-foreground text-center">
              &copy; {new Date().getFullYear()} Kthulu Go
            </p>
          </div>
        </div>
      </div>
    </>
  );
}
