import { readFile } from "node:fs/promises";
import path from "node:path";
import Link from "next/link";
import { BookOpen, FileText, Languages, ListChecks } from "lucide-react";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { SYSTEM_SETTING_SLUGS } from "@/lib/system-setting-screens";
import { ThemeToggle } from "../theme-toggle";

export const dynamic = "force-dynamic";

type ManualContent = {
  title: string;
  summary?: string;
  objective: string;
  workflow: string[];
  config: string[];
  limitations: string[];
};

type ManualFile = {
  screen: string;
  updatedAt: string;
  translations: Partial<Record<LanguageCode, ManualContent>>;
};

type ManualIndexProps = {
  searchParams?: Promise<{ lang?: string; screen?: string }>;
};

const manualScreens = ["login", "workspace", "settings", "menu", "currency", ...SYSTEM_SETTING_SLUGS] as const;

const indexUi: Record<LanguageCode, {
  title: string;
  subtitle: string;
  language: string;
  open: string;
  updated: string;
  coverage: string;
  coverageItems: string[];
}> = {
  th: { title: "สารบัญคู่มือ BC Ai Account", subtitle: "รวมคู่มือทุกหน้าจอ เรียงตามลำดับการใช้งานจริง อ่านจากหน้านี้ก่อน แล้วเลือกคู่มือที่ต้องการ", language: "ภาษา", open: "เปิดคู่มือ", updated: "อัปเดต", coverage: "คู่มือทุกหน้าต้องมี", coverageItems: ["ขั้นตอนใช้งานแบบละเอียด", "คำอธิบายช่องข้อมูลและปุ่ม", "ผลลัพธ์ที่ควรได้และวิธีแก้ปัญหา", "รองรับภาษาและธีมมืด/สว่าง"] },
  en: { title: "BC Ai Account Manual Contents", subtitle: "All screen manuals in real workflow order. Start here, then open the guide you need.", language: "Language", open: "Open manual", updated: "Updated", coverage: "Every manual includes", coverageItems: ["Detailed step-by-step workflow", "Field and button explanations", "Expected results and troubleshooting", "Language and dark/light theme support"] },
  cn: { title: "BC Ai Account 手册目录", subtitle: "按实际工作流程整理所有页面手册。请先从这里开始，再打开需要的手册。", language: "语言", open: "打开手册", updated: "更新", coverage: "每份手册包含", coverageItems: ["详细步骤", "字段和按钮说明", "预期结果和故障排除", "语言和深浅主题支持"] },
  ja: { title: "BC Ai Account マニュアル目次", subtitle: "実際の作業順にすべての画面マニュアルをまとめています。ここから必要な手順を開いてください。", language: "言語", open: "マニュアルを開く", updated: "更新", coverage: "各マニュアルの内容", coverageItems: ["詳細な操作手順", "項目とボタンの説明", "期待結果とトラブル対応", "言語とダーク/ライトテーマ対応"] },
  ko: { title: "BC Ai Account 매뉴얼 목차", subtitle: "실제 업무 순서에 맞춘 모든 화면 매뉴얼입니다. 여기에서 필요한 안내를 여세요.", language: "언어", open: "매뉴얼 열기", updated: "업데이트", coverage: "모든 매뉴얼 포함 항목", coverageItems: ["상세 단계별 작업", "필드와 버튼 설명", "예상 결과와 문제 해결", "언어 및 다크/라이트 테마 지원"] },
  lo: { title: "ສາລະບານຄູ່ມື BC Ai Account", subtitle: "ລວມຄູ່ມືທຸກໜ້າຕາມລໍາດັບໃຊ້ງານຈິງ ເລີ່ມຈາກໜ້ານີ້ແລ້ວເປີດຄູ່ມືທີ່ຕ້ອງການ.", language: "ພາສາ", open: "ເປີດຄູ່ມື", updated: "ອັບເດດ", coverage: "ຄູ່ມືທຸກໜ້າມີ", coverageItems: ["ຂັ້ນຕອນລະອຽດ", "ອະທິບາຍຊ່ອງຂໍ້ມູນແລະປຸ່ມ", "ຜົນທີ່ຄວນໄດ້ແລະແກ້ບັນຫາ", "ຮອງຮັບພາສາແລະທີມມືດ/ສະຫວ່າງ"] },
  my: { title: "BC Ai Account လမ်းညွှန် မာတိကာ", subtitle: "အသုံးပြုမှုအစဉ်အတိုင်း မျက်နှာပြင်လမ်းညွှန်အားလုံးကိုစုထားသည်။ ဤနေရာမှစပြီး လိုသောလမ်းညွှန်ကိုဖွင့်ပါ။", language: "ဘာသာစကား", open: "လမ်းညွှန်ဖွင့်ရန်", updated: "နောက်ဆုံးပြင်ဆင်", coverage: "လမ်းညွှန်တိုင်းတွင်", coverageItems: ["အသေးစိတ်အဆင့်လိုက်လုပ်ဆောင်နည်း", "အချက်အလက်ကွက်နှင့်ခလုတ်ရှင်းလင်းချက်", "ရရှိသင့်သောရလဒ်နှင့်ပြဿနာဖြေရှင်းနည်း", "ဘာသာစကားနှင့်အမှောင်/အလင်း theme"] },
  km: { title: "មាតិកាសៀវភៅណែនាំ BC Ai Account", subtitle: "ប្រមូលសៀវភៅណែនាំគ្រប់ទំព័រតាមលំដាប់ប្រើប្រាស់ពិត។ ចាប់ផ្តើមពីទីនេះ រួចបើកសៀវភៅណែនាំដែលត្រូវការ។", language: "ភាសា", open: "បើកសៀវភៅណែនាំ", updated: "បានធ្វើបច្ចុប្បន្នភាព", coverage: "សៀវភៅណែនាំនីមួយៗមាន", coverageItems: ["ជំហានប្រើប្រាស់លម្អិត", "ការពន្យល់វាល និងប៊ូតុង", "លទ្ធផលដែលគួរទទួលបាន និងដោះស្រាយបញ្ហា", "គាំទ្រភាសា និងរូបរាងងងឹត/ភ្លឺ"] },
  vi: { title: "Mục lục hướng dẫn BC Ai Account", subtitle: "Tập hợp hướng dẫn mọi màn hình theo đúng thứ tự sử dụng. Bắt đầu tại đây rồi mở phần cần đọc.", language: "Ngôn ngữ", open: "Mở hướng dẫn", updated: "Cập nhật", coverage: "Mỗi hướng dẫn có", coverageItems: ["Quy trình từng bước chi tiết", "Giải thích trường dữ liệu và nút", "Kết quả mong đợi và xử lý sự cố", "Hỗ trợ ngôn ngữ và giao diện tối/sáng"] },
  ms: { title: "Kandungan Manual BC Ai Account", subtitle: "Semua manual skrin disusun mengikut aliran kerja sebenar. Mula di sini, kemudian buka manual yang diperlukan.", language: "Bahasa", open: "Buka manual", updated: "Dikemas kini", coverage: "Setiap manual merangkumi", coverageItems: ["Aliran kerja langkah demi langkah", "Penjelasan medan dan butang", "Hasil dijangka dan penyelesaian masalah", "Sokongan bahasa dan tema gelap/cerah"] },
  id: { title: "Daftar Isi Panduan BC Ai Account", subtitle: "Semua panduan layar disusun sesuai alur kerja nyata. Mulai dari sini, lalu buka panduan yang diperlukan.", language: "Bahasa", open: "Buka panduan", updated: "Diperbarui", coverage: "Setiap panduan berisi", coverageItems: ["Alur kerja langkah demi langkah", "Penjelasan field dan tombol", "Hasil yang diharapkan dan pemecahan masalah", "Dukungan bahasa dan tema gelap/terang"] },
  fil: { title: "Nilalaman ng Manual ng BC Ai Account", subtitle: "Lahat ng screen manual ayon sa totoong workflow. Magsimula rito, pagkatapos buksan ang kailangan mong gabay.", language: "Wika", open: "Buksan ang manual", updated: "Na-update", coverage: "Kasama sa bawat manual", coverageItems: ["Detalyadong step-by-step workflow", "Paliwanag ng fields at buttons", "Inaasahang resulta at troubleshooting", "Suporta sa wika at dark/light theme"] },
};

export default async function ManualIndexPage({ searchParams }: ManualIndexProps) {
  const query = searchParams ? await searchParams : {};
  const language = normalizeLanguage(query.lang);
  const currentScreen = isManualScreen(query.screen) ? query.screen : null;
  const labels = indexUi[language] ?? indexUi.en;
  const manuals = await Promise.all(manualScreens.map((screen) => readManual(screen)));

  return (
    <main className="min-h-dvh w-full max-w-none bg-background px-2 py-2 text-foreground sm:px-3 lg:px-4">
      <article className="grid w-full max-w-none gap-3">
        <header className="grid gap-3 rounded-2xl border border-border bg-card p-3 shadow-sm sm:p-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="flex items-center gap-2 text-xs font-semibold uppercase tracking-normal text-muted-foreground">
                <BookOpen size={16} />
                <span>{labels.coverage}</span>
              </p>
              <h1 className="mt-1 text-2xl font-semibold tracking-normal sm:text-3xl">{labels.title}</h1>
              <p className="mt-2 max-w-[96ch] text-sm leading-6 text-muted-foreground sm:text-base">{labels.subtitle}</p>
            </div>
            <ThemeToggle language={language} />
          </div>

          <section className="grid gap-2 rounded-2xl border border-border bg-background p-3">
            <h2 className="flex items-center gap-2 text-sm font-semibold">
              <Languages size={17} />
              <span>{labels.language}</span>
            </h2>
            <nav className="flex flex-wrap gap-1.5" aria-label={labels.language}>
              {LANGUAGES.map((item) => (
                <Link
                  className={`min-h-8 rounded-full border px-3 py-1.5 text-xs font-semibold ${item.code === language ? "border-primary bg-primary text-primary-foreground" : "border-border bg-card text-muted-foreground"}`}
                  href={`/manual?lang=${item.code}`}
                  key={item.code}
                >
                  {item.name}
                </Link>
              ))}
            </nav>
          </section>
        </header>

        <section className="grid w-full gap-3 xl:grid-cols-[minmax(280px,0.28fr)_minmax(0,1fr)]">
          <aside className="grid h-fit gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm xl:sticky xl:top-3">
            <h2 className="flex items-center gap-2 text-sm font-semibold">
              <ListChecks size={17} />
              <span>{labels.coverage}</span>
            </h2>
            <ul className="grid gap-2 text-sm leading-6 text-muted-foreground">
              {labels.coverageItems.map((item) => (
                <li className="rounded-xl border border-border bg-background px-3 py-2" key={item}>{item}</li>
              ))}
            </ul>
          </aside>

          <div className="grid min-w-0 gap-3 md:grid-cols-2 2xl:grid-cols-4">
            {manuals.map((manual, index) => {
              const content = manual.translations[language] ?? manual.translations.en ?? manual.translations.th;
              return (
                <Link
                  className={`grid min-w-0 content-between gap-3 rounded-2xl border bg-card p-3 text-foreground shadow-sm transition hover:border-primary hover:bg-accent hover:no-underline ${manual.screen === currentScreen ? "border-primary ring-2 ring-ring/20" : "border-border"}`}
                  href={`/manual/${manual.screen}?lang=${language}`}
                  key={manual.screen}
                >
                  <div className="grid gap-2">
                    <div className="flex items-center justify-between gap-2">
                      <span className="grid size-10 place-items-center rounded-xl bg-primary text-sm font-semibold text-primary-foreground">{index + 1}</span>
                      <span className="text-xs font-semibold text-muted-foreground">{labels.updated}: {manual.updatedAt}</span>
                    </div>
                    <h2 className="text-lg font-semibold leading-snug">{content?.title ?? manual.screen}</h2>
                    <p className="text-sm leading-6 text-muted-foreground">{content?.summary ?? content?.objective}</p>
                  </div>
                  <span className="inline-flex min-h-9 items-center justify-center gap-2 rounded-xl border border-border bg-background px-3 text-sm font-semibold text-primary">
                    <FileText size={16} />
                    {labels.open}
                  </span>
                </Link>
              );
            })}
          </div>
        </section>
      </article>
    </main>
  );
}

async function readManual(screen: string): Promise<ManualFile> {
  const manualPath = path.resolve(process.cwd(), "..", "manual", `${screen}.json`);
  const raw = await readFile(manualPath, "utf8");
  return JSON.parse(raw) as ManualFile;
}

function isManualScreen(screen: string | undefined): screen is typeof manualScreens[number] {
  return Boolean(screen && (manualScreens as readonly string[]).includes(screen));
}
