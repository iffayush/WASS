'use client';

import { useSession, useSupabaseClient } from '@supabase/auth-helpers-react';
import { Button } from './ui/button';
import { ProjectForm } from './ProjectForm';
import { ScanList } from './ScanList';
import { useState } from 'react';

export function Dashboard() {
  const session = useSession();
  const supabase = useSupabaseClient();
  const [reloadTrigger, setReloadTrigger] = useState(0);

  const refreshProjects = () => setReloadTrigger((prev) => prev + 1);

  if (!session) {
    return (
      <div className="text-center py-16">
        <h1 className="text-3xl font-bold mb-4">Welcome to WASS</h1>
        <p className="mb-6">Sign in to start scanning your projects for security risks.</p>
        <Button onClick={() => supabase.auth.signInWithOAuth({ provider: 'github', options: {
        redirectTo: 'https://wass-ebon.vercel.app',
  }})}>
          Sign in with GitHub
        </Button>
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-2xl font-semibold mb-6">Your Projects</h2>
      <ProjectForm onSuccess={refreshProjects} />
      <ScanList userId={session.user.id} reloadTrigger={reloadTrigger} />
    </div>
  );
}
