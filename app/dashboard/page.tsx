'use client'

import { useSession, useSupabaseClient } from '@supabase/auth-helpers-react';
import { useRouter } from 'next/navigation';
import ScanHistory from '@/components/ScanHistory';
import { useEffect, useState } from 'react'

export default function Dashboard() {
  const session = useSession();
  const supabase = useSupabaseClient();
  const router = useRouter();

  // Wait for session to resolve before redirecting
  useEffect(() => {
    if (session === undefined) return; // Wait for session to resolve
    if (!session) {
      router.replace('/');
    }
  }, [session, router]);

  const [scanUrl, setScanUrl] = useState('');
  const [scanResult, setScanResult] = useState<any | null>(null);
  const [scanLoading, setScanLoading] = useState(false);
  const [scanError, setScanError] = useState<string | null>(null);

  // Scanner handler
  const handleScan = async () => {
    setScanLoading(true)
    setScanResult(null)
    setScanError(null)
    try {
      const res = await fetch('http://localhost:8000/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: scanUrl }), // changed to match new schema
      })
      const data = await res.json()
      if (res.ok) {
        setScanResult(data.result)
        // Insert scan result into Supabase
        const { error } = await supabase.from('scans').insert([
          {
            target_url: scanUrl,
            status: 'completed',
            result: data.result,
            scan_started_at: new Date().toISOString(),
          }
        ])
        if (error) {
          console.error('Error inserting scan into Supabase:', error)
        }
      } else {
        setScanError(data.detail || 'Scan failed')
      }
    } catch (err) {
      setScanError('Could not connect to scanner backend')
    }
    setScanLoading(false)
  }

  // Helper to check if all findings are zero
  const noFindings = scanResult && scanResult.critical === 0 && scanResult.medium === 0 && scanResult.low === 0;

  // Only render dashboard if session is resolved and valid
  if (session === undefined) {
    return <div className="text-gray-400 p-8">Loading...</div>;
  }
  if (!session) {
    return null; // Or a loading spinner
  }


  return (
    <div className="min-h-screen bg-black text-white p-8">
      {/* Scanner UI */}
      <div className="mb-8 bg-[#0d1117] border border-gray-800 rounded-lg p-6 mx-auto w-full sm:w-3/5">
        {/* Notice about scan duration */}
        <div className="mb-4 p-3 rounded bg-yellow-900 border border-yellow-700 text-yellow-300 text-sm font-medium">
          <span>
            <strong>Note:</strong> Scanning for vulnerabilities is a time-consuming process. Please wait patiently while the scan completes. The duration may vary depending on the target and network conditions.
          </span>
        </div>
        <h2 className="text-xl font-bold mb-4">Vulnerability Scanner</h2>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-4 w-full">
          <div className="w-full flex flex-col sm:flex-row items-center gap-4">
            <input
              type="text"
              value={scanUrl}
              onChange={e => setScanUrl(e.target.value)}
              placeholder="Enter target URL (e.g. https://example.com)"
              className="flex-1 px-4 py-2 rounded bg-gray-900 border border-gray-700 text-white focus:outline-none focus:ring-2 focus:ring-orange-400"
              disabled={scanLoading}
            />
            <button
              onClick={handleScan}
              disabled={scanLoading || !scanUrl}
              className="px-6 py-2 rounded bg-orange-500 hover:bg-orange-600 text-white font-semibold disabled:opacity-50 w-full sm:w-auto"
            >
              {scanLoading ? 'Scanning...' : 'Scan'}
            </button>
          </div>
        </div>
        {scanError && (
          <div className="text-red-400 mb-2">{scanError}</div>
        )}
        {scanResult && (
          <div className="mt-4">
            <h3 className="font-semibold mb-2">Scan Results:</h3>
            <div className="flex gap-6 mb-4">
              <div className="flex flex-col items-center">
                <span className="text-lg font-bold text-red-500">{scanResult.critical}</span>
                <span className="text-gray-400">Critical</span>
              </div>
              <div className="flex flex-col items-center">
                <span className="text-lg font-bold text-yellow-400">{scanResult.medium}</span>
                <span className="text-gray-400">Medium</span>
              </div>
              <div className="flex flex-col items-center">
                <span className="text-lg font-bold text-green-400">{scanResult.low}</span>
                <span className="text-gray-400">Low</span>
              </div>
            </div>
            <div className="text-xs text-gray-500 bg-gray-900 rounded p-3 border border-gray-800">
              {noFindings ? (
                <span>
                  No vulnerabilities were detected for the provided target. <br />
                  <br />
                </span>
              ) : null}
              <span>
                <strong>Disclaimer:</strong> This vulnerability scanning tool is experimental and provided for informational purposes only. The results may not be comprehensive or fully accurate, and the absence of reported vulnerabilities does not guarantee the security of the target. Use of this tool is at your own risk. For a complete security assessment, consult a qualified professional.
              </span>
            </div>
          </div>
        )}
      </div>
      {/* Scan History */}
    <ScanHistory />
    </div>
  )
}
