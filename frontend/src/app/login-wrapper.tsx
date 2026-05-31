"use client";

import { useEffect, useState } from "react";
import { LoginScreen } from "./login-screen";

export function LoginWrapper() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <main className="login-shell" style={{ display: "flex", alignItems: "center", justifyContent: "center", minHeight: "100vh", width: "100%" }}>
        <div style={{ fontSize: "1.2rem", fontWeight: "bold", color: "var(--muted-foreground)" }}>Loading...</div>
      </main>
    );
  }

  return <LoginScreen />;
}
