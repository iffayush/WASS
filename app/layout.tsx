// app/layout.tsx
import './globals.css'
import { Inter } from 'next/font/google'
import { AuthProvider } from '@/components/AuthProvider';

const inter = Inter({ subsets: ['latin'] })

export const metadata = {
  title: 'WASS',
  description: 'Web App Security Scanner',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <AuthProvider>
          <main>{children}</main>
        </AuthProvider>
      </body>
    </html>
  )
}
