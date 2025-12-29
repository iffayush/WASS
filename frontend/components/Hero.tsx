"use client";
import { useRouter } from "next/navigation";

export default function Hero() {
  const router = useRouter();

  return (
    <section className="min-h-screen grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-12 px-8 py-16 lg:px-16 border-b border-[#2a2a2a] bg-[#0d0d0d]">
      <div className="flex flex-col justify-center space-y-8">
        <div className="space-y-3">
          <p className="text-sm tracking-[0.24em] text-[#9ca3af] uppercase">
            Web App Security Scanner
          </p>
          <h1 className="text-6xl font-bold tracking-[0.08em] uppercase leading-[1.05]">
            WASS
          </h1>
          <p className="text-lg max-w-2xl text-[#9ca3af] leading-relaxed">
            Automated, repeatable, and aggressive reconnaissance for modern web stacks.
            Designed for red teams and SOC operators who need audit-grade clarity.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row gap-4">
          <button
            onClick={() => router.push("/analysis")}
            className="px-8 py-4 uppercase tracking-[0.12em] bg-red-700 text-[#e5e5e5] border-2 border-[#2a2a2a] hover:bg-red-800 active:bg-red-900 transition-colors duration-150 rounded-none"
          >
            Start Scan
          </button>
          <div className="px-6 py-4 border-2 border-dashed border-[#2a2a2a] rounded-none text-sm text-[#9ca3af]">
            Ready to initiate live repository interrogation.
          </div>
        </div>
      </div>

      <div className="hidden lg:flex flex-col justify-center gap-4 border border-[#2a2a2a] bg-[#111111] px-6 py-8">
        <h3 className="text-sm tracking-[0.2em] text-[#9ca3af] uppercase">
          Scan Posture
        </h3>
        <div className="space-y-3 text-sm text-[#e5e5e5]">
          <div className="flex items-center justify-between border-b border-[#2a2a2a] pb-2">
            <span className="text-[#9ca3af]">Mode</span>
            <span>Automated Recon</span>
          </div>
          <div className="flex items-center justify-between border-b border-[#2a2a2a] pb-2">
            <span className="text-[#9ca3af]">Scope</span>
            <span>GitHub Repository</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-[#9ca3af]">Output</span>
            <span>Static + Dependency Audit</span>
          </div>
        </div>
      </div>
    </section>
  );
}
