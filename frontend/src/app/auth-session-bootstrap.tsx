"use client";

import { useEffect, useState, type ReactNode } from "react";
import { getAuthSession, restoreAuthSession } from "@/lib/client-auth-session";

export function AuthSessionBootstrap({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(() => getAuthSession() !== null);

  useEffect(() => {
    if (ready) return;
    void restoreAuthSession().finally(() => setReady(true));
  }, [ready]);

  return ready ? children : null;
}
