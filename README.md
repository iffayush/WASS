# ⚡ WASS – Web App Security Scanner

**WASS** is a lightweight, developer-friendly web application scanner built for modern SaaS teams. It scans websites and repositories for common security risks and maps them to compliance requirements (like ISO 27001, SOC 2).

## 🔍 Features

- Upload GitHub repo or website URL
- Auto-scans for security issues using [Nuclei](https://github.com/projectdiscovery/nuclei)
- Real-time scan status & scoring
- Supabase-backed auth, data, and session management
- Fully Dockerized FastAPI backend
- Vercel-hosted frontend with Next.js App Router

## 📦 Tech Stack

- **Frontend**: Next.js 14 (App Router), Radix UI, Tailwind CSS
- **Backend**: FastAPI, Nuclei, Python
- **Auth & DB**: Supabase (PostgreSQL + Auth)
- **Infra**: Docker, AWS Lightsail, Vercel

## 🚀 Getting Started

```bash
# Frontend
cd frontend
npm install
npm run dev

# Backend
cd backend
docker build -t wass-backend .
docker run --env-file .env -p 8000:8000 wass-backend
