// app/layout.tsx
import './globals.css';
import { ReactNode } from 'react';
import { AuthProvider } from '@/components/AuthProvider';

export const metadata = {
  title: 'Web App Security Scanner',
  description: 'Security risk scanner for developers',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-gray-950 text-white">
        <AuthProvider>
          <main className="max-w-5xl mx-auto py-8 px-4">{children}</main>
        </AuthProvider>
      </body>
    </html>
  );
}
