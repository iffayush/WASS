'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useSession } from '@supabase/auth-helpers-react';
import { useState } from 'react';

const navItems = [
  { label: 'home', href: '/', type: 'link' },
  { label: 'report', href: '/report', type: 'link' },
  { label: 'documentation', href: 'https://github.com/iffayush/WASS', type: 'external' },
  { label: 'about', href: '/about', type: 'link' },
  { label: 'community', href: 'https://discord.gg/your-server', type: 'external' },
];

export default function Navbar() {
  const pathname = usePathname();
  const session = useSession();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 w-full z-50 bg-[#0d0d0d] border-b border-white/10">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Logo / Brand */}
        <Link href="/" className="text-green-400 font-mono text-xl font-bold">
          WASS
        </Link>

        {/* Desktop Nav */}
        <nav className="hidden md:flex gap-6 text-sm font-mono text-white/70">
          {navItems.map(({ label, href, type }) => {
            if (type === 'external') {
              return (
                <a
                  key={label}
                  href={href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-green-400 transition"
                >
                  {label}
                </a>
              );
            }

            if (label === 'report' && !session) {
              return null; // hide "report" if user is not authenticated
            }

            return (
              <Link
                key={label}
                href={href}
                className={`hover:text-green-400 transition ${
                  pathname === href ? 'text-green-400' : ''
                }`}
              >
                {label}
              </Link>
            );
          })}
        </nav>

        {/* Mobile Hamburger */}
        <div className="md:hidden">
          <button
            onClick={() => setMobileOpen(!mobileOpen)}
            className="text-white text-xl"
            aria-label="Toggle Navigation"
          >
            ☰
          </button>
        </div>
      </div>

      {/* Mobile Nav Dropdown */}
      {mobileOpen && (
        <div className="md:hidden bg-[#0d0d0d] px-4 py-2 space-y-2 border-t border-white/10">
          {navItems.map(({ label, href, type }) => {
            if (type === 'external') {
              return (
                <a
                  key={label}
                  href={href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block text-white/70 hover:text-green-400"
                >
                  {label}
                </a>
              );
            }

            if (label === 'report' && !session) {
              return null;
            }

            return (
              <Link
                key={label}
                href={href}
                className="block text-white/70 hover:text-green-400"
                onClick={() => setMobileOpen(false)}
              >
                {label}
              </Link>
            );
          })}
        </div>
      )}
    </header>
  );
}
