'use client';

import { useSession } from '@supabase/auth-helpers-react';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function AuthCallback() {
  const session = useSession();
  const router = useRouter();

  useEffect(() => {
    // Wait for session to resolve
    if (session === undefined) return;
    if (session) {
      router.replace('/dashboard');
    } else if (session === null) {
      // Only redirect to home if we are sure there is no session
      router.replace('/');
    }
  }, [session, router]);

  return <p className="text-white p-4">Completing sign-in...</p>;
}
