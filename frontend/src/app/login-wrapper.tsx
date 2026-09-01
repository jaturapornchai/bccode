"use client";

import { useEffect, useState } from "react";
import { LoginScreen } from "./login-screen";
import { MotionConfigProvider } from "./motion-config";

export function LoginWrapper() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    // Keep the pre-hydration tree static so React hydrates identical HTML.
    return (
      <main
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          minHeight: "100vh",
          width: "100%",
          flexDirection: "column",
          gap: 18,
          background:
            "radial-gradient(70% 45% at 8% 0%, color-mix(in srgb, var(--primary, #812920) 18%, transparent) 0, transparent 70%), radial-gradient(60% 40% at 100% 100%, color-mix(in srgb, var(--primary, #812920) 12%, transparent) 0, transparent 70%), var(--background, #fbf9f8)",
        }}
      >
        <div
          aria-hidden="true"
          style={{
            width: 56,
            height: 56,
            borderRadius: 16,
            display: "grid",
            placeItems: "center",
            fontWeight: 800,
            letterSpacing: "0.02em",
            color: "#fff",
            background:
              "linear-gradient(140deg, var(--primary, #812920), color-mix(in srgb, var(--primary, #812920) 62%, #1a0e0c))",
            boxShadow: "inset 0 1px 0 rgba(255,255,255,0.35), 0 14px 30px -10px rgba(129, 41, 32, 0.6)",
          }}
        >
          BC
        </div>
        <div
          aria-hidden="true"
          style={{
            width: 140,
            height: 8,
            borderRadius: 999,
            background:
              "linear-gradient(90deg, color-mix(in srgb, var(--primary, #812920) 12%, transparent), color-mix(in srgb, var(--primary, #812920) 36%, transparent), color-mix(in srgb, var(--primary, #812920) 12%, transparent))",
          }}
        />
      </main>
    );
  }

  return (
    <MotionConfigProvider>
      <LoginScreen />
    </MotionConfigProvider>
  );
}
