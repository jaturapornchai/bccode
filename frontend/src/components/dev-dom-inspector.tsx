"use client";

import { useEffect, useRef, useState } from "react";
import { toast } from "@/lib/toast";
import { Code2, Check, Crosshair, X } from "lucide-react";

type HighlightBox = {
  top: number;
  left: number;
  width: number;
  height: number;
  tagName: string;
  selector: string;
};

async function copyTextSafely(text: string): Promise<boolean> {
  // 1. Try modern Clipboard API if supported and available
  if (typeof navigator !== "undefined" && navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // Fall through to fallback
    }
  }

  // 2. Fallback: textarea + document.execCommand('copy') (works inside iframes, non-secure contexts, etc.)
  try {
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.left = "-9999px";
    textarea.style.top = "-9999px";
    textarea.style.opacity = "0";
    textarea.setAttribute("readonly", "");
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();
    textarea.setSelectionRange(0, textarea.value.length);
    const success = document.execCommand("copy");
    document.body.removeChild(textarea);
    if (success) return true;
  } catch (err) {
    console.error("Fallback execCommand copy failed:", err);
  }

  return false;
}

export function DevDomInspector() {
  const [mounted, setMounted] = useState(false);
  const [active, setActive] = useState(false);
  const [altHeld, setAltHeld] = useState(false);
  const [copiedRecently, setCopiedRecently] = useState(false);
  const [highlight, setHighlight] = useState<HighlightBox | null>(null);

  const activeRef = useRef(active);
  activeRef.current = active;

  const altHeldRef = useRef(altHeld);
  altHeldRef.current = altHeld;

  const isInspecting = active || altHeld;

  useEffect(() => {
    setMounted(true);
  }, []);

  // Listen for Alt key hold
  useEffect(() => {
    if (!mounted) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Alt") {
        setAltHeld(true);
      } else if (e.key === "Escape" && activeRef.current) {
        setActive(false);
        setHighlight(null);
      }
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.key === "Alt") {
        setAltHeld(false);
        if (!activeRef.current) {
          setHighlight(null);
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, [mounted]);

  // Hover and Click interception in capture phase
  useEffect(() => {
    if (!mounted) return;

    const handleMouseMove = (e: MouseEvent) => {
      if (!activeRef.current && !altHeldRef.current) {
        if (highlight) setHighlight(null);
        return;
      }

      const target = e.target as HTMLElement | null;
      if (!target || target.closest("[data-dev-dom-inspector]")) {
        setHighlight(null);
        return;
      }

      const rect = target.getBoundingClientRect();
      const tagName = target.tagName.toLowerCase();
      const id = target.id ? `#${target.id}` : "";
      const classList = Array.from(target.classList)
        .filter((c) => !c.startsWith("dev-dom-") && c.length < 24)
        .slice(0, 2)
        .map((c) => `.${c}`)
        .join("");

      setHighlight({
        top: rect.top,
        left: rect.left,
        width: rect.width,
        height: rect.height,
        tagName,
        selector: `${tagName}${id}${classList}`,
      });
    };

    const handleClick = (e: MouseEvent) => {
      if (!activeRef.current && !altHeldRef.current) return;

      const target = e.target as HTMLElement | null;
      if (!target || target.closest("[data-dev-dom-inspector]")) return;

      // Intercept and prevent original click action
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation();

      const html = target.outerHTML;
      copyTextSafely(html).then((success) => {
        if (success) {
          setCopiedRecently(true);
          setTimeout(() => setCopiedRecently(false), 800);
          toast.success(`คัดลอก DOM แล้ว: <${target.tagName.toLowerCase()}> (${html.length.toLocaleString()} ตัวอักษร)`);
          console.log("🎯 [DOM Inspector] Copied element:", target, "\n", html);
        } else {
          toast.error("ไม่สามารถคัดลอกลง Clipboard ได้ (ลองเลือกองค์ประกอบใหม่อีกครั้ง)");
        }
      });
    };

    window.addEventListener("mousemove", handleMouseMove, { passive: true });
    window.addEventListener("click", handleClick, true); // Capture phase!

    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("click", handleClick, true);
    };
  }, [mounted]);

  if (!mounted) {
    return null;
  }

  return (
    <>
      {/* Element Highlight Overlay */}
      {isInspecting && highlight && (
        <div
          data-dev-dom-inspector
          style={{
            position: "fixed",
            top: highlight.top,
            left: highlight.left,
            width: highlight.width,
            height: highlight.height,
            pointerEvents: "none",
            zIndex: 999998,
            border: copiedRecently ? "2px solid #22c55e" : "2px dashed #3b82f6",
            backgroundColor: copiedRecently ? "rgba(34, 197, 94, 0.25)" : "rgba(59, 130, 246, 0.18)",
            borderRadius: "4px",
            transition: "background-color 0.15s ease, border-color 0.15s ease",
            boxShadow: copiedRecently ? "0 0 15px rgba(34, 197, 94, 0.6)" : "0 0 10px rgba(59, 130, 246, 0.4)",
          }}
        >
          {/* Label Tooltip */}
          <div
            style={{
              position: "absolute",
              top: highlight.top < 30 ? "100%" : "-26px",
              left: 0,
              backgroundColor: copiedRecently ? "#16a34a" : "#1e293b",
              color: "#ffffff",
              fontSize: "11px",
              fontFamily: "monospace",
              padding: "2px 8px",
              borderRadius: "4px",
              whiteSpace: "nowrap",
              display: "flex",
              alignItems: "center",
              gap: "6px",
              boxShadow: "0 2px 6px rgba(0,0,0,0.3)",
              border: "1px solid rgba(255,255,255,0.15)",
            }}
          >
            {copiedRecently ? (
              <>
                <Check size={12} className="text-white" />
                <span>คัดลอก DOM สำเร็จ!</span>
              </>
            ) : (
              <>
                <Crosshair size={12} className="text-blue-400" />
                <span className="font-bold text-blue-200">{highlight.selector}</span>
                <span className="text-slate-400 text-[10px]">— คลิกเพื่อ Copy</span>
              </>
            )}
          </div>
        </div>
      )}

      {/* Floating Control Button */}
      <div
        data-dev-dom-inspector
        className="fixed bottom-4 left-4 z-[999999] flex items-center gap-1.5 select-none"
      >
        <button
          type="button"
          onClick={() => {
            const next = !active;
            setActive(next);
            if (!next) setHighlight(null);
            if (next) {
              toast.info("เปิดโหมด Copy DOM แล้ว — ชี้แล้วคลิกที่องค์ประกอบใดก็ได้บนจอ");
            }
          }}
          title={
            active
              ? "คลิกเพื่อปิดโหมด Copy DOM (หรือกด Esc)"
              : "คลิกเพื่อเปิดโหมด Copy DOM (หรือกดปุ่ม Alt ค้างไว้แล้วคลิกองค์ประกอบใดก็ได้บนจอ)"
          }
          className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium shadow-lg transition-all duration-200 border ${
            active
              ? "bg-blue-600 text-white border-blue-400 ring-2 ring-blue-400/40 shadow-blue-500/20"
              : altHeld
              ? "bg-amber-600 text-white border-amber-400 ring-2 ring-amber-400/40 animate-pulse"
              : "bg-slate-900/80 hover:bg-slate-900 text-slate-300 hover:text-white border-slate-700/80 backdrop-blur-md"
          }`}
        >
          {active ? (
            <>
              <Crosshair size={14} className="animate-spin text-blue-200" style={{ animationDuration: "3s" }} />
              <span>โหมด Copy DOM: เปิดอยู่</span>
              <X size={13} className="text-blue-200 hover:text-white ml-0.5" />
            </>
          ) : altHeld ? (
            <>
              <Crosshair size={14} className="text-amber-200" />
              <span>กด Alt ค้าง: พร้อมคลิก Copy</span>
            </>
          ) : (
            <>
              <Code2 size={14} className="text-blue-400" />
              <span>Copy DOM</span>
              <kbd className="px-1.5 py-0.5 text-[10px] font-mono bg-slate-800 rounded border border-slate-700 text-slate-400">
                Alt+คลิก
              </kbd>
            </>
          )}
        </button>
      </div>
    </>
  );
}
