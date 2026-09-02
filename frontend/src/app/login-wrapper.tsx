"use client";

import { useEffect, useState } from "react";
import { LoginScreen } from "./login-screen";
import { MotionConfigProvider } from "./motion-config";

export function LoginWrapper() {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // The login page keeps its approved (pre-×1.5) root scale: globals.css
  // resolves html[data-login-scale] back to the historical 10/12/14px ladder.
  // Every other page renders at the new system scale (15/18/21px).
  useEffect(() => {
    document.documentElement.dataset.loginScale = "1";
    return () => {
      delete document.documentElement.dataset.loginScale;
    };
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
            "radial-gradient(circle at 30% 20%, color-mix(in srgb, var(--primary, #812920) 14%, transparent) 0, transparent 42%), radial-gradient(circle at 75% 80%, color-mix(in srgb, var(--primary, #812920) 10%, transparent) 0, transparent 40%), var(--background, #fbf9f8)",
        }}
      >
        <div
          aria-hidden="true"
          style={{
            width: 56,
            height: 56,
            borderRadius: 8,
            display: "grid",
            placeItems: "center",
            fontWeight: 900,
            color: "var(--primary-foreground, #fff)",
            background:
              "linear-gradient(135deg, var(--primary, #812920), color-mix(in srgb, var(--primary, #812920) 70%, #000))",
            boxShadow: "0 10px 28px rgba(160, 64, 53, 0.28)",
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
