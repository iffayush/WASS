'use client'

import { FaGithub } from 'react-icons/fa'
import Navbar from '@/components/Navbar'
import { useSupabaseClient } from '@supabase/auth-helpers-react';

export default function Home() {
  const supabase = useSupabaseClient();

  // Supabase auth
  const handleLogin = async () => {
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'github',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    });
    if (error) console.error('GitHub login error:', error.message);
  };

  return (
    <main className="min-h-screen flex flex-col bg-black text-white font-inter">
      <Navbar />

      <div className="flex flex-1 flex-col md:flex-row">
        {/* Left Side - Info */}
        <div className="flex-1 bg-[#031a17] p-10 flex flex-col justify-center items-start text-left">
          <h1 className="text-4xl md:text-5xl font-bold text-orange-500 mb-6">
            WASS: Web App Security Scanner
          </h1>
          <p className="text-lg text-gray-300 leading-relaxed mb-6 max-w-lg">
            Scan public web apps and GitHub repositories for vulnerabilities.  
            Lightweight, developer-first, and made for speed.
          </p>
          <a href="/documentation" className="underline text-orange-300 font-medium hover:text-orange-400">
            Learn more →
          </a>
        </div>

        {/* Right Side - Login Box */}
        <div className="w-full md:w-[400px] bg-[#0d1117] border-l border-gray-800 flex items-center justify-center px-6 py-12">
          <div className="bg-[#0d1117] border border-[#30363d] rounded-xl w-full max-w-xs shadow-lg px-6 py-8">
            <h2 className="text-center text-white text-sm font-medium tracking-widest mb-5">
              SIGN IN
            </h2>
            <button onClick={handleLogin} className="bg-white text-black flex items-center justify-center gap-2 py-2 px-3 rounded-md w-full hover:bg-gray-200 transition text-sm">
              <FaGithub className="text-base" />
              Sign in with GitHub
            </button>
          </div>
        </div>
      </div>
    </main>
  )
}
