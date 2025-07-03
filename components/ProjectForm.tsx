'use client';

import { useState } from 'react';
import { useSupabaseClient, useSession } from '@supabase/auth-helpers-react';
import { Button } from './ui/button';

export function ProjectForm({ onSuccess }: { onSuccess?: () => void }) {
  const [url, setUrl] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const [loading, setLoading] = useState(false);
  const supabase = useSupabaseClient();
  const session = useSession();

  const isValidUrl = (string: string): boolean => {
    try {
      const parsed = new URL(string);
      return ['http:', 'https:'].includes(parsed.protocol);
    } catch {
      return false;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg('');
    const trimmedUrl = url.trim();
  
    if (!trimmedUrl) {
      setErrorMsg('Please enter a URL.');
      return;
    }
  
    if (!isValidUrl(trimmedUrl)) {
      setErrorMsg('Enter a valid URL starting with http:// or https://');
      return;
    }
  
    setLoading(true);
  
    try {
      // 1. Create project in Supabase
      const { data: project, error: projectError } = await supabase
        .from('projects')
        .insert({
          url: trimmedUrl,
          user_id: session?.user.id,
        })
        .select()
        .single();
  
      if (projectError) throw new Error(projectError.message);
  
      const projectId = project.id;
  

      // 2. Call backend to queue scan
      const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://3.111.40.124:8000";
      const response = await fetch(`${BASE_URL}/api/scan`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          url: trimmedUrl,
          user_id: session?.user.id,
          project_id: projectId,
        }),
      });
  
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.detail || 'Failed to start scan');
      }
  
      const { scan_id, status } = await response.json();
      console.log('Scan started:', scan_id, status);
  
      onSuccess?.(); // refresh ScanList
      setUrl('');
    } catch (err: any) {
      setErrorMsg(err.message || 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };
  

  return (
    <form onSubmit={handleSubmit} className="mb-8">
      <label className="block text-sm font-medium mb-2">
        GitHub Repo / Website URL
      </label>
      <input
        type="text"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        placeholder="https://github.com/user/repo or https://yoursite.com"
        className="w-full p-2 rounded bg-gray-800 border border-gray-700 text-white"
      />
      {errorMsg && (
        <p className="text-red-500 text-sm mt-2">{errorMsg}</p>
      )}
      <Button type="submit" className="mt-4" disabled={loading}>
        {loading ? 'Submitting...' : 'Submit'}
      </Button>
    </form>
  );
}
