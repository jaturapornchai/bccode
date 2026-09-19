"use client";

import React, { useState, useRef, useEffect } from "react";
import { 
  MessageSquare, 
  X, 
  Send, 
  Sparkles, 
  Maximize2, 
  Minimize2, 
  RotateCcw, 
  BookOpen, 
  HelpCircle,
  ChevronRight,
  Bot
} from "lucide-react";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  sources?: string[];
  time: string;
}

const QUICK_PROMPTS = [
  "โครงสร้างผังบัญชีและการตั้งค่า",
  "วิธีบันทึกสมุดรายวันทั่วไป",
  "การเลือก Holding บริษัท และสาขา",
  "สถาปัตยกรรมระบบและฐานข้อมูล PostgreSQL",
];

export function KmsChatbot() {
  const [isOpen, setIsOpen] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "welcome",
      role: "assistant",
      content: "สวัสดีครับลุงจืดและผู้ใช้งานทุกท่าน! ผมคือผู้ช่วยอัจฉริยะ **BC Ai Assistant** พร้อมตอบคำถามและให้คำแนะนำเกี่ยวกับคู่มือการใช้งานระบบ BC Ai Account และข้อมูลทางบัญชี ถามผมได้เลยครับ!",
      time: new Date().toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" }),
    },
  ]);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (isOpen) {
      scrollToBottom();
      setTimeout(() => inputRef.current?.focus(), 150);
    }
  }, [isOpen, messages, loading]);

  const handleSend = async (textToSend?: string) => {
    const text = (textToSend || input).trim();
    if (!text || loading) return;

    const userMsg: Message = {
      id: "user-" + Date.now(),
      role: "user",
      content: text,
      time: new Date().toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" }),
    };

    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setLoading(true);

    try {
      const history = messages
        .filter((m) => m.id !== "welcome")
        .concat(userMsg)
        .map((m) => ({ role: m.role, content: m.content }));

      const res = await fetch("/api/kms-chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ messages: history }),
      });

      const data = await res.json();
      if (res.ok && data.success) {
        const botMsg: Message = {
          id: "bot-" + Date.now(),
          role: "assistant",
          content: data.answer,
          sources: data.sources,
          time: new Date().toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" }),
        };
        setMessages((prev) => [...prev, botMsg]);
      } else {
        const errorMsg: Message = {
          id: "bot-err-" + Date.now(),
          role: "assistant",
          content: `ขออภัยครับ เกิดข้อผิดพลาด: ${data.error || "ไม่สามารถติดต่อระบบ AI ได้"}`,
          time: new Date().toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" }),
        };
        setMessages((prev) => [...prev, errorMsg]);
      }
    } catch (err: any) {
      const errorMsg: Message = {
        id: "bot-err-" + Date.now(),
        role: "assistant",
        content: `ขออภัยครับ เกิดปัญหาเครือข่าย: ${err?.message || "กรุณาลองใหม่อีกครั้ง"}`,
        time: new Date().toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" }),
      };
      setMessages((prev) => [...prev, errorMsg]);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setMessages([
      {
        id: "welcome",
        role: "assistant",
        content: "สวัสดีครับ! เริ่มการสนทนาใหม่แล้ว มีข้อสงสัยเรื่องการใช้งานหรือคู่มือระบบส่วนไหน สอบถามได้เลยครับ",
        time: new Date().toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit" }),
      },
    ]);
  };

  return (
    <div className="fixed bottom-5 right-5 z-[9999] font-sans antialiased">
      {/* Floating Trigger Button */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          className="group flex items-center gap-2.5 bg-gradient-to-r from-blue-600 via-indigo-600 to-sky-600 hover:from-blue-700 hover:to-sky-700 text-white px-4 py-3 rounded-full shadow-xl hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-1 active:translate-y-0 cursor-pointer border border-white/20"
          title="เปิดผู้ช่วยตอบคำถามคู่มือ (KMS AI)"
          aria-label="เปิดผู้ช่วยตอบคำถามคู่มือ"
        >
          <div className="relative flex items-center justify-center">
            <Bot className="w-6 h-6 animate-pulse" />
            <span className="absolute -top-1 -right-1 flex h-2.5 w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
            </span>
          </div>
          <span className="text-sm font-semibold tracking-wide pr-1">
            ถามคู่มือ AI (KMS)
          </span>
        </button>
      )}

      {/* Chat Window */}
      {isOpen && (
        <div
          className={`flex flex-col bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden transition-all duration-300 ${
            isExpanded
              ? "w-[92vw] sm:w-[680px] h-[85vh] max-h-[850px]"
              : "w-[92vw] sm:w-[420px] h-[580px] max-h-[80vh]"
          }`}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 bg-gradient-to-r from-blue-600 via-indigo-600 to-sky-600 text-white shadow-md">
            <div className="flex items-center gap-2.5">
              <div className="p-1.5 bg-white/15 rounded-lg">
                <Sparkles className="w-5 h-5 text-amber-300" />
              </div>
              <div>
                <h3 className="text-sm font-bold leading-tight flex items-center gap-1.5">
                  BC Ai Assistant
                  <span className="text-[10px] bg-white/20 text-white font-medium px-1.5 py-0.5 rounded-full">
                    KMS RAG
                  </span>
                </h3>
                <p className="text-[11px] text-blue-100/90 font-normal">
                  ผู้ช่วยตอบคำถามคู่มือและระบบบัญชี
                </p>
              </div>
            </div>

            <div className="flex items-center gap-1">
              <button
                onClick={handleReset}
                className="p-1.5 hover:bg-white/15 text-blue-100 hover:text-white rounded-lg transition-colors"
                title="เริ่มการสนทนาใหม่"
              >
                <RotateCcw className="w-4 h-4" />
              </button>
              <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="p-1.5 hover:bg-white/15 text-blue-100 hover:text-white rounded-lg transition-colors hidden sm:block"
                title={isExpanded ? "ย่อหน้าต่าง" : "ขยายหน้าต่าง"}
              >
                {isExpanded ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
              </button>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1.5 hover:bg-white/15 text-blue-100 hover:text-white rounded-lg transition-colors"
                title="ปิดหน้าต่าง"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* Message List */}
          <div className="flex-1 overflow-y-auto p-4 space-y-3.5 bg-slate-50/50 dark:bg-slate-950/40">
            {messages.map((m) => (
              <div
                key={m.id}
                className={`flex flex-col ${
                  m.role === "user" ? "items-end" : "items-start"
                }`}
              >
                <div
                  className={`max-w-[85%] rounded-2xl px-4 py-2.5 text-sm shadow-sm leading-relaxed ${
                    m.role === "user"
                      ? "bg-blue-600 text-white rounded-br-none"
                      : "bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 border border-slate-200/80 dark:border-slate-700/80 rounded-bl-none"
                  }`}
                >
                  <div className="whitespace-pre-wrap">{m.content}</div>

                  {m.sources && m.sources.length > 0 && (
                    <div className="mt-2.5 pt-2 border-t border-slate-200 dark:border-slate-700/60 text-[11px] text-slate-500 dark:text-slate-400">
                      <div className="flex items-center gap-1 font-medium text-slate-600 dark:text-slate-300 mb-1">
                        <BookOpen className="w-3.5 h-3.5 text-indigo-500" />
                        เอกสารอ้างอิง:
                      </div>
                      <ul className="list-disc list-inside space-y-0.5">
                        {m.sources.map((s, idx) => (
                          <li key={idx} className="truncate">
                            {s}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
                <span className="text-[10px] text-slate-400 mt-1 px-1">{m.time}</span>
              </div>
            ))}

            {/* Loading Indicator */}
            {loading && (
              <div className="flex items-center gap-2 text-slate-500 dark:text-slate-400 text-xs bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl rounded-bl-none px-4 py-2.5 w-fit shadow-sm">
                <span className="flex h-2 w-2 relative">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-blue-600"></span>
                </span>
                <span>กำลังค้นหาคู่มือและประมวลผลคำตอบ...</span>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Quick Prompts (shown when only welcome message) */}
          {messages.length === 1 && !loading && (
            <div className="p-3 bg-slate-100/70 dark:bg-slate-900 border-t border-slate-200/80 dark:border-slate-800">
              <p className="text-[11px] font-medium text-slate-500 dark:text-slate-400 mb-1.5 flex items-center gap-1">
                <HelpCircle className="w-3 h-3 text-blue-500" />
                คำถามแนะนำ:
              </p>
              <div className="flex flex-wrap gap-1.5">
                {QUICK_PROMPTS.map((prompt, i) => (
                  <button
                    key={i}
                    onClick={() => handleSend(prompt)}
                    className="text-xs bg-white dark:bg-slate-800 hover:bg-blue-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 rounded-lg px-2.5 py-1 text-left flex items-center gap-1 transition-colors"
                  >
                    <span>{prompt}</span>
                    <ChevronRight className="w-3 h-3 text-slate-400" />
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Input Box */}
          <div className="p-3 bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleSend();
              }}
              className="flex items-center gap-2"
            >
              <input
                ref={inputRef}
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="พิมพ์คำถามเกี่ยวกับระบบหรือคู่มือที่นี่..."
                disabled={loading}
                className="flex-1 text-sm bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-slate-100 px-3.5 py-2.5 rounded-xl border border-slate-200 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all disabled:opacity-50 placeholder:text-slate-400"
              />
              <button
                type="submit"
                disabled={loading || !input.trim()}
                className="p-2.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-40 text-white rounded-xl shadow-md transition-colors cursor-pointer flex-shrink-0"
                title="ส่งคำถาม"
              >
                <Send className="w-4 h-4" />
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
