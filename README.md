# ⚡ WASS – Web App Security Scanner (Frontend)

This is the frontend for **WASS**, a modern Web App Security Scanner that helps developers and teams detect vulnerabilities and stay compliant.

## ✨ Features

- Supabase GitHub OAuth login
- Project submission (by URL)
- Live scan queue tracking
- Security score & report visualization
- TailwindCSS + Radix UI + Next.js 14 (App Router)

## 🔧 Tech Stack

- Next.js (App Router)
- Tailwind CSS + Radix UI
- Supabase (Auth + DB)
- Deployed on Vercel

## 🚀 Getting Started

```bash
# Clone the repo
git clone https://github.com/your-username/wass.git
cd wass

# Install deps
npm install

# Add environment variables to .env.local
NEXT_PUBLIC_SUPABASE_URL=...
NEXT_PUBLIC_SUPABASE_ANON_KEY=...
NEXT_PUBLIC_API_URL=http://your-backend-url:8000

# Run the app
npm run dev
