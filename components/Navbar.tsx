// components/Navbar.tsx
import Link from 'next/link'

export default function Navbar() {
  return (
    <nav className="w-full px-6 py-4 text-sm text-orange-400 font-semibold flex gap-6">
      <Link href="/">Home</Link>
      <Link href="/documentation">Documentation</Link>
      <Link href="/about">About</Link>
    </nav>
  )
}
