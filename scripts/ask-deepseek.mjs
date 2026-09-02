#!/usr/bin/env node
// ที่ปรึกษา DeepSeek — ใช้ขอความเห็นที่สองจาก DeepSeek Chat API (OpenAI-compatible)
// ใช้งาน: node scripts/ask-deepseek.mjs "คำถามหรือบริบทที่จะปรึกษา"
//        echo "โจทย์" | node scripts/ask-deepseek.mjs
//        node scripts/ask-deepseek.mjs --model deepseek-reasoner "โจทย์"
// Key อยู่ที่ deepseek.env (git-ignored ผ่าน pattern *.env — ห้าม commit)

import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const env = Object.fromEntries(
  readFileSync(join(root, "deepseek.env"), "utf8")
    .split(/\r?\n/)
    .filter((line) => line.includes("="))
    .map((line) => [line.slice(0, line.indexOf("=")).trim(), line.slice(line.indexOf("=") + 1).trim()]),
);

const apiKey = process.env.DEEPSEEK_API_KEY || env.DEEPSEEK_API_KEY;
if (!apiKey) {
  console.error("ไม่พบ DEEPSEEK_API_KEY (ตรวจ deepseek.env)");
  process.exit(1);
}
const baseUrl = env.DEEPSEEK_BASE_URL || "https://api.deepseek.com";
const argModelIdx = process.argv.indexOf("--model");
const model = argModelIdx > -1 ? process.argv[argModelIdx + 1] : env.DEEPSEEK_MODEL || "deepseek-chat";
const argv = process.argv.slice(2).filter((a, i) => !(argModelIdx > -1 && (i === argModelIdx - 2 || a === model)));

let prompt = argv.join(" ").trim();
if (!prompt) {
  prompt = readFileSync(0, "utf8").trim();
}
if (!prompt) {
  console.error('ใส่คำถาม เช่น node scripts/ask-deepseek.mjs "ประเมินแนวทาง..."');
  process.exit(1);
}

const response = await fetch(`${baseUrl}/chat/completions`, {
  method: "POST",
  headers: { "Content-Type": "application/json", Authorization: `Bearer ${apiKey}` },
  body: JSON.stringify({
    model,
    messages: [
      {
        role: "system",
        content:
          "คุณเป็นที่ปรึกษาให้ทีมพัฒนา BC Ai Account (ระบบบัญชีสำหรับ SME ไทย, Next.js + Go + MongoDB) ตอบสั้น ตรงประเด็น ชี้ข้อเสียจริง ไม่ประจบ และระบุตำแหน่งโค้ด/ไฟล์เมื่อเกี่ยวข้อง",
      },
      { role: "user", content: prompt },
    ],
    stream: false,
  }),
});

if (!response.ok) {
  console.error(`DeepSeek API error ${response.status}: ${await response.text()}`);
  process.exit(1);
}
const data = await response.json();
process.stdout.write(data.choices?.[0]?.message?.content ?? JSON.stringify(data));
