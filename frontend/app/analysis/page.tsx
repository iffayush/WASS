"use client";

import { useState } from "react";
import TechSelector from "@/components/TechSelector";
import StartScanButton from "@/components/StartScanButton";
import ScanResults from "@/components/ScanResults";

type ScanReport = {
  reportTitle: string;
  riskScore: number;
  top3Risks: string[];
  immediateActions: string[];
  remediationPlan: Array<{
    phase: string;
    steps: string[];
  }>;
  preventionMeasures: string[];
};

export default function AnalysisPage() {
  const [repoUrl, setRepoUrl] = useState("");
  const [branch, setBranch] = useState("main");
  const [isLoading, setIsLoading] = useState(false);
  const [scanReport, setScanReport] = useState<ScanReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const framework = "nextjs"; // Hardcoded React + Next.js

  const handleScanComplete = (report: ScanReport) => {
    setScanReport(report);
    setError(null);
  };

  const handleScanError = (errorMessage: string) => {
    setError(errorMessage);
    setScanReport(null);
  };

  return (
    <main className="min-h-screen bg-[#0a0a0a] px-6 py-16 lg:px-16">
      <div className="max-w-4xl mx-auto space-y-10">
        <header className="space-y-3 border-b border-[#2a2a2a] pb-6">
          <p className="text-sm tracking-[0.2em] uppercase text-[#9ca3af]">
            New Scan
          </p>
          <h1 className="text-4xl font-semibold tracking-[0.08em] uppercase">
            Analysis Configuration
          </h1>
          <p className="text-sm text-[#9ca3af]">
            Define scope, select runtime targets, and initiate a security-first execution.
          </p>
        </header>

        <section className="space-y-6 border border-[#2a2a2a] bg-[#0d0d0d] p-6 rounded-none">
          <div className="space-y-3">
            <label className="text-sm tracking-[0.14em] uppercase text-[#9ca3af]">
              GitHub Repository URL
            </label>
            <input
              type="url"
              placeholder="https://github.com/username/repo"
              value={repoUrl}
              onChange={(e) => setRepoUrl(e.target.value)}
              className="w-full px-4 py-3 bg-[#0a0a0a] border-2 border-[#2a2a2a] text-[#e5e5e5] placeholder:text-[#4b5563] focus:border-red-700 outline-none rounded-none"
            />
          </div>

          <div className="space-y-3">
            <h3 className="text-sm tracking-[0.16em] uppercase text-[#9ca3af]">
              Framework
            </h3>
            <div className="px-4 py-3 bg-[#111111] border-2 border-[#2a2a2a] text-[#e5e5e5] text-sm tracking-[0.12em] uppercase">
              React + Next.js
            </div>
          </div>

          <TechSelector value={branch} onChange={setBranch} />

          {error && (
            <div className="border-2 border-red-700 bg-[#0a0a0a] p-4">
              <p className="text-red-600 text-sm font-mono uppercase tracking-[0.12em]">
                Error: {error}
              </p>
            </div>
          )}

          <div className="pt-2">
            <StartScanButton
              payload={{ repoUrl, framework, branch }}
              onScanComplete={handleScanComplete}
              onScanError={handleScanError}
              isLoading={isLoading}
              setIsLoading={setIsLoading}
            />
          </div>
        </section>

        {scanReport && <ScanResults report={scanReport} />}
      </div>
    </main>
  );
}
