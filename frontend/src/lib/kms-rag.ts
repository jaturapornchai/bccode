import kmsData from "./kms-knowledge.json";

export interface KnowledgeChunk {
  source: string;
  title: string;
  content: string;
}

const chunks: KnowledgeChunk[] = kmsData as KnowledgeChunk[];

/**
 * Tokenize and normalize search terms
 */
function tokenize(text: string): string[] {
  return text
    .toLowerCase()
    .replace(/[^\w\sก-๙]/g, " ")
    .split(/\s+/)
    .filter((w) => w.length > 1);
}

/**
 * Perform keyword-based RAG search on the KMS knowledge base
 */
export function searchKnowledge(query: string, topK = 4): KnowledgeChunk[] {
  const queryTokens = tokenize(query);
  if (queryTokens.length === 0) {
    return chunks.slice(0, topK);
  }

  const scored = chunks.map((chunk) => {
    let score = 0;
    const titleLower = chunk.title.toLowerCase();
    const contentLower = chunk.content.toLowerCase();
    const sourceLower = chunk.source.toLowerCase();

    // Exact query match bonus
    const qLower = query.toLowerCase().trim();
    if (titleLower.includes(qLower)) score += 50;
    if (contentLower.includes(qLower)) score += 20;

    // Token matching
    for (const token of queryTokens) {
      if (titleLower.includes(token)) {
        score += 10;
      }
      if (sourceLower.includes(token)) {
        score += 5;
      }
      // Count frequency in content
      const regex = new RegExp(token, "g");
      const matches = contentLower.match(regex);
      if (matches) {
        score += Math.min(matches.length, 5); // cap frequency score
      }
    }

    return { chunk, score };
  });

  scored.sort((a, b) => b.score - a.score);

  return scored
    .filter((item) => item.score > 0)
    .slice(0, topK)
    .map((item) => item.chunk);
}

const DEEPSEEK_API_KEY = process.env.DEEPSEEK_API_KEY || "sk-221af3a3b84047c5991fd66a792805af";
const DEEPSEEK_API_URL = "https://api.deepseek.com/chat/completions";

export interface ChatMessage {
  role: "system" | "user" | "assistant";
  content: string;
}

export async function askKmsChatbot(userMessages: ChatMessage[]): Promise<{ answer: string; sources: string[] }> {
  const lastUserMsg = [...userMessages].reverse().find((m) => m.role === "user");
  const query = lastUserMsg ? lastUserMsg.content : "";

  const relevantChunks = searchKnowledge(query, 5);
  const sources = Array.from(new Set(relevantChunks.map((c) => c.source)));

  const contextText = relevantChunks
    .map((c, i) => `--- เอกสารอ้างอิงที่ ${i + 1} (${c.source} : ${c.title}) ---\n${c.content}`)
    .join("\n\n");

  const systemPrompt = `คุณคือผู้ช่วยอัจฉริยะ "BC Ai Assistant" ประจำระบบ BC Ai Account
ระบบนี้เป็นระบบบัญชีและการจัดการองค์กรสำหรับธุรกิจไทย (SMEs ไทย) ซึ่งพัฒนาต่อยอดจากโปรแกรม Champ (บน Windows) มาสู่ระบบ Web Application
คุณมีความเชี่ยวชาญด้าน:
1. การใช้งานระบบ BC Ai Account (ผังบัญชี, สมุดรายวัน, การผ่านรายการ, ทะเบียนทรัพย์สิน, คลังสินค้า, ภาษี ภ.พ.30 / WHT)
2. สถาปัตยกรรมระบบ (Next.js, Go Backend, PostgreSQL แบบบริสุทธิ์ 100%, Redis)
3. การตั้งค่าสิทธิ์ผู้ใช้งาน สาขา และโครงสร้างองค์กร

แนวทางการตอบคำถาม:
- ตอบเป็นภาษาไทยอย่างสุภาพ กระชับ ชัดเจน และตรงประเด็น
- อ้างอิงข้อมูลจากเอกสารความรู้ (KMS) ที่ระบุด้านล่างเป็นหลัก
- หากมีขั้นตอน ให้ระบุเป็นข้อๆ 1, 2, 3 ให้เข้าใจและปฏิบัติตามได้ง่าย
- ในตอนท้ายของการตอบ ให้ระบุเอกสารอ้างอิงจากระบบ KMS ให้ผู้ใช้ทราบด้วย

ข้อมูลความรู้จากระบบ KMS:
${contextText}`;

  const messages: ChatMessage[] = [
    { role: "system", content: systemPrompt },
    ...userMessages.slice(-6), // เก็บประวัติล่าสุด 6 ข้อความ
  ];

  const response = await fetch(DEEPSEEK_API_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${DEEPSEEK_API_KEY}`,
    },
    body: JSON.stringify({
      model: "deepseek-chat",
      messages: messages,
      temperature: 0.3,
      max_tokens: 1500,
    }),
  });

  if (!response.ok) {
    const errText = await response.text();
    throw new Error(`DeepSeek API error (${response.status}): ${errText}`);
  }

  const data = await response.json();
  const answer = data.choices?.[0]?.message?.content || "ขออภัยครับ ไม่สามารถสร้างคำตอบได้ในขณะนี้";

  return { answer, sources };
}
