'use client';

import { useSupabaseClient } from '@supabase/auth-helpers-react';
import { useEffect, useState } from 'react';

interface Scan {
  id: string;
  target_url: string;
  scan_started_at: string;
  status: string;
  result?: any;
}

export default function ScanHistory() {
  const supabase = useSupabaseClient();
  const [scans, setScans] = useState<Scan[]>([]);
  const [loading, setLoading] = useState(true);
  const [reloadTrigger, setReloadTrigger] = useState(0);
  const refreshProjects = () => setReloadTrigger((prev) => prev + 1);

  useEffect(() => {
    const fetchScans = async () => {
      setLoading(true);
      const { data, error } = await supabase
        .from('scans')
        .select('id, target_url, scan_started_at, status, result')
        .order('scan_started_at', { ascending: false });
  
      if (error) {
        console.error('Error fetching scans:', error);
        setScans([]);
      } else {
        setScans(data as Scan[]);
      }
      setLoading(false);
    };
  
    fetchScans();
  }, [supabase, reloadTrigger]);

  if (loading) return <p className="text-gray-400">Loading scan history...</p>;

  return (
    <div className="bg-[#0d1117] border border-gray-800 rounded-lg p-6 mx-auto w-full sm:w-3/5 mt-8">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-orange-400">Scan History</h2>
        <button
          onClick={refreshProjects}
          className="text-green-400 hover:text-white transition p-1"
          title="Refresh Scan History"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="23 4 23 10 17 10" />
            <polyline points="1 20 1 14 7 14" />
            <path d="M3.51 9a9 9 0 0114.13-3.36L23 10" />
            <path d="M20.49 15a9 9 0 01-14.13 3.36L1 14" />
          </svg>
        </button>
      </div>
      {scans.length === 0 ? (
        <p className="text-gray-400">No scans found.</p>
      ) : (
        <ul className="space-y-4">
          {scans.map(scan => (
            <li
              key={scan.id}
              className="p-4 border rounded bg-gray-900 border-gray-700 flex flex-col sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p className="font-mono text-sm text-gray-400 break-all">{scan.target_url}</p>
                <p className="text-xs text-gray-500">
                  Started: {new Date(scan.scan_started_at).toLocaleString()}
                </p>
                {scan.result && (
                  <p className="text-xs text-gray-400 mt-1">
                    Result: {typeof scan.result === 'string' ? scan.result : JSON.stringify(scan.result)}
                  </p>
                )}
              </div>
              <div className="mt-2 sm:mt-0">
                <span
                  className={
                    scan.status === 'completed'
                      ? 'text-green-400 font-semibold'
                      : scan.status === 'failed'
                      ? 'text-red-400 font-semibold'
                      : 'text-yellow-400 font-semibold'
                  }
                >
                  {scan.status.charAt(0).toUpperCase() + scan.status.slice(1)}
                </span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}