#!/usr/bin/env node
// ที่ปรึกษา Kimi (K3 Max) — ขอความเห็นที่สองจาก Kimi Coding gateway (Anthropic-compatible)
// ใช้งาน: node scripts/ask-kimi.mjs "คำถามหรือบริบทที่จะปรึกษา"
//        echo "โจทย์" | node scripts/ask-kimi.mjs
//        node scripts/ask-kimi.mjs --model kimi-k3-max "โจทย์"   (default)
// Key อยู่ที่ ~/.kimi/kimi-claw/openclaw.json (เครื่องผู้ใช้ — ห้าม copy key
// ลงไฟล์/log/commit ใด ๆ ตามกฎ consultant เดียวกับ DeepSeek)

import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";

const cfgPath = join(homedir(), ".kimi", "kimi-claw", "openclaw.json");
let baseUrl = "https://agent-gw.kimi.com/coding";
let apiKey = process.env.KIMI_API_KEY || "";
try {
  const cfg = JSON.parse(readFileSync(cfgPath, "utf8"));
  const prov = cfg?.models?.providers?.["kimi-coding"];
  if (prov?.baseUrl) baseUrl = String(prov.baseUrl).replace(/\/$/, "");
  if (prov?.apiKey) apiKey = prov.apiKey;
} catch {
  // fall through to env-only
}
if (!apiKey) {
  console.error("ไม่พบ Kimi API key (ตรวจ ~/.kimi/kimi-claw/openclaw.json หรือตั้ง KIMI_API_KEY)");
  process.exit(1);
}

const argModelIdx = process.argv.indexOf("--model");
const model = argModelIdx > -1 ? process.argv[argModelIdx + 1] : "kimi-k3-max";
const argv = process.argv.slice(2).filter((a, i) => !(argModelIdx > -1 && (i === argModelIdx - 2 || a === model)));

let prompt = argv.join(" ").trim();
if (!prompt) {
  prompt = readFileSync(0, "utf8").trim();
}
if (!prompt) {
  console.error('ใส่คำถาม เช่น node scripts/ask-kimi.mjs "ประเมินแนวทาง..."');
  process.exit(1);
}

const response = await fetch(`${baseUrl}/v1/messages`, {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "x-api-key": apiKey,
    Authorization: `Bearer ${apiKey}`,
    "anthropic-version": "2023-06-01",
  },
  body: JSON.stringify({
    model,
    max_tokens: 16384,
    system:
      "คุณเป็นที่ปรึกษาให้ทีมพัฒนา BC Ai Account (ระบบบัญชีสำหรับ SME ไทย, Next.js + Go + MongoDB) ผู้ใช้หลักคนไทยอายุ 40+ ตอบสั้น ตรงประเด็น ชี้ข้อเสียจริง ไม่ประจบ และระบุตำแหน่งโค้ด/ไฟล์เมื่อเกี่ยวข้อง",
    messages: [{ role: "user", content: prompt }],
  }),
});

const payload = await response.json().catch(() => null);
if (!response.ok || !payload) {
  console.error(`Kimi API error ${response.status}:`, typeof payload === "object" ? JSON.stringify(payload).slice(0, 300) : response.status);
  process.exit(1);
}

// Anthropic content array — เอาเฉพาะบล็อก type:text (ข้าม thinking)
const text = Array.isArray(payload.content)
  ? payload.content.filter((block) => block.type === "text").map((block) => block.text).join("\n").trim()
  : "";
if (!text) {
  console.error("Kimi ตอบกลับมาไม่มีเนื้อความ:", JSON.stringify(payload).slice(0, 300));
  process.exit(1);
}
console.log(text);
