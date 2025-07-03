'use client';

import { useEffect, useState } from 'react';
import { useSupabaseClient } from '@supabase/auth-helpers-react';

interface Scan {
  id: string;
  status: string;
  score: number | null;
  timestamp: string;
}

interface Project {
  id: string;
  url: string;
  created_at: string;
  scans: Scan[];
}

export function ScanList({
  userId,
  reloadTrigger,
}: {
  userId: string;
  reloadTrigger: number;
}) {
  const supabase = useSupabaseClient();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchProjectsWithScans() {
      setLoading(true);
      const { data, error } = await supabase
        .from('projects')
        .select(`
          id,
          url,
          created_at,
          scans (
            id,
            status,
            score,
            created_at
          )
        `)
        .eq('user_id', userId)
        .order('created_at', { ascending: false });

      if (error) {
        console.error('Error fetching projects:', error);
      } else {
        setProjects(
          data.map((project: any) => ({
            ...project,
            scans: project.scans.map((scan: any) => ({
              ...scan,
              timestamp: scan.created_at,
            })),
          }))
        );
      }

      setLoading(false);
    }

    if (userId) fetchProjectsWithScans();
  }, [userId, reloadTrigger, supabase]);

  return (
    <div>
      <h3 className="text-xl font-medium mb-4">Your Submitted Projects</h3>
      {loading ? (
        <p>Loading...</p>
      ) : projects.length === 0 ? (
        <p className="text-gray-400">No projects submitted yet.</p>
      ) : (
        <ul className="space-y-4">
          {projects.map((project) => {
            const latestScan = [...project.scans].sort((a, b) =>
              new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
            )[0];

            return (
              <li
                key={project.id}
                className="p-4 border rounded bg-gray-900 border-gray-700"
              >
                <p className="font-mono text-sm text-gray-400">{project.url}</p>
                <p className="text-xs text-gray-500">
                  Submitted: {new Date(project.created_at).toLocaleString()}
                </p>

                {latestScan ? (
                  <>
                    <p className="text-sm mt-1">
                      Status:{' '}
                      <span
                        className={
                          latestScan.status === 'completed'
                            ? 'text-green-400'
                            : latestScan.status === 'error'
                            ? 'text-red-400'
                            : 'text-yellow-400'
                        }
                      >
                        {latestScan.status}
                      </span>
                    </p>
                    {latestScan.status === 'completed' && (
                      <p className="text-sm text-blue-300">
                        Score: {latestScan.score?.toFixed(1) ?? 'N/A'}
                      </p>
                    )}
                  </>
                ) : (
                  <p className="text-sm text-yellow-300 italic">No scans yet</p>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
