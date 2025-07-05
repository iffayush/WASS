'use client';

import { useCallback } from 'react';
import Particles from 'react-tsparticles';
import { loadSlim } from 'tsparticles-slim';
import { useSupabaseClient } from '@supabase/auth-helpers-react';
import { useRouter } from 'next/navigation';
import { Button } from './ui/button';
import Navbar from './Navbar';

export default function Hero() {
  const router = useRouter();
  const supabase = useSupabaseClient();

  const particlesInit = useCallback(async (engine: any) => {
    await loadSlim(engine);
  }, []);

  const handleStartScan = () => {
    supabase.auth.signInWithOAuth({
      provider: 'github',
      options: {
        redirectTo: 'https://wass-ebon.vercel.app/',
      },
    });
  };

  return (
    <>
    <Navbar />
    <section className="relative h-screen w-full bg-[#0d0d0d] text-white overflow-hidden">
      {/* Particles */}
      <Particles
        className="absolute inset-0 z-0"
        id="tsparticles"
        init={particlesInit}
        options={{
          background: { color: { value: '#0d0d0d' } },
          fpsLimit: 60,
          particles: {
            number: {
              value: 60,
              density: { enable: true, area: 800 },
            },
            color: { value: '#ffffff' },
            shape: { type: 'circle' },
            opacity: { value: 0.2 },
            size: { value: 2, random: true },
            move: {
              enable: true,
              speed: 0.3,
              direction: 'none',
              straight: false,
              outModes: { default: 'out' },
            },
          },
        }}
      />

      {/* Content */}
      <div className="relative z-10 flex flex-col items-center justify-center h-full text-center px-4">
        <h1 className="text-4xl md:text-6xl font-mono font-bold text-green-400">
          &gt;_ WASS <span className="animate-blink">|</span>
        </h1>
        <p className="mt-4 text-white/70 max-w-xl">
          Web App Security Scanner. Scan, detect, and secure your web projects — the ethical way.
        </p>
        <Button
          onClick={handleStartScan}
          className="mt-8 px-6 py-3 border border-green-500 bg-transparent text-green-400 hover:bg-green-500 hover:text-black transition font-mono"
        >
          Start Scanning
        </Button>
      </div>
    </section>
    </>
  );
}
