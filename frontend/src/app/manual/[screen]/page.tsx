import { readFile } from "node:fs/promises";
import path from "node:path";
import Link from "next/link";
import { redirect } from "next/navigation";
import { BookOpen, CheckCircle2, CircleAlert, ClipboardList, HelpCircle, Home, Info, ListChecks, Settings2 } from "lucide-react";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { SYSTEM_SETTING_SLUGS } from "@/lib/system-setting-screens";
import { ThemeToggle } from "../../theme-toggle";

export const dynamic = "force-dynamic";

type ManualContent = {
  title: string;
  summary?: string;
  sources?: ManualSource[];
  objective: string;
  prerequisites?: string[];
  workflow: string[];
  fields?: ManualPair[];
  actions?: ManualPair[];
  config: string[];
  expectedResults?: string[];
  commonMistakes?: string[];
  troubleshooting?: ManualPair[];
  limitations: string[];
  nextSteps?: string[];
};

type ManualPair = {
  name: string;
  description: string;
};

type ManualSource = {
  name: string;
  href: string;
};

type ManualFile = {
  screen: string;
  updatedAt: string;
  translations: Partial<Record<LanguageCode, ManualContent>>;
};

const manualScreens = new Set(["login", "workspace", "settings", "menu", "currency", ...SYSTEM_SETTING_SLUGS]);
const manualLanguages = LANGUAGES.filter((item) => item.code === "th" || item.code === "en");
const manualUi: Record<LanguageCode, {
  actions: string;
  commonMistakes: string;
  config: string;
  expectedResults: string;
  fields: string;
  limitations: string;
  nextSteps: string;
  home: string;
  prerequisites: string;
  quickLinks: string;
  screen: string;
  summary: string;
  troubleshooting: string;
  updated: string;
  workflow: string;
}> = {
  th: { actions: "ปุ่มและคำสั่ง", commonMistakes: "ข้อผิดพลาดที่พบบ่อย", config: "การตั้งค่า", expectedResults: "ผลลัพธ์ที่ควรได้", fields: "ความหมายของช่องข้อมูล", home: "หน้าแรกคู่มือ", limitations: "ข้อจำกัด", nextSteps: "ทำอะไรต่อ", prerequisites: "ก่อนเริ่ม", quickLinks: "สารบัญ", screen: "หน้าจอ", summary: "สรุป", troubleshooting: "แก้ปัญหาเบื้องต้น", updated: "อัปเดต", workflow: "ขั้นตอนใช้งาน" },
  en: { actions: "Buttons and actions", commonMistakes: "Common mistakes", config: "Configuration", expectedResults: "Expected results", fields: "Field guide", home: "Manual home", limitations: "Limitations", nextSteps: "Next steps", prerequisites: "Before you start", quickLinks: "Contents", screen: "Screen", summary: "Summary", troubleshooting: "Troubleshooting", updated: "Updated", workflow: "Step-by-step workflow" },
  cn: { actions: "按钮和操作", commonMistakes: "常见错误", config: "配置", expectedResults: "预期结果", fields: "字段说明", home: "手册首页", limitations: "限制", nextSteps: "下一步", prerequisites: "开始之前", quickLinks: "目录", screen: "页面", summary: "摘要", troubleshooting: "故障排除", updated: "更新", workflow: "操作步骤" },
  ja: { actions: "ボタンと操作", commonMistakes: "よくある間違い", config: "設定", expectedResults: "期待される結果", fields: "項目説明", home: "マニュアルホーム", limitations: "制限", nextSteps: "次の操作", prerequisites: "開始前", quickLinks: "目次", screen: "画面", summary: "概要", troubleshooting: "トラブル対応", updated: "更新", workflow: "操作手順" },
  ko: { actions: "버튼 및 작업", commonMistakes: "자주 하는 실수", config: "설정", expectedResults: "예상 결과", fields: "필드 안내", home: "매뉴얼 홈", limitations: "제한 사항", nextSteps: "다음 단계", prerequisites: "시작 전", quickLinks: "목차", screen: "화면", summary: "요약", troubleshooting: "문제 해결", updated: "업데이트", workflow: "단계별 작업" },
  lo: { actions: "ປຸ່ມແລະຄໍາສັ່ງ", commonMistakes: "ຂໍ້ຜິດພາດທີ່ພົບເລື້ອຍ", config: "ການຕັ້ງຄ່າ", expectedResults: "ຜົນທີ່ຄວນໄດ້", fields: "ຄວາມໝາຍຊ່ອງຂໍ້ມູນ", home: "ໜ້າຫຼັກຄູ່ມື", limitations: "ຂໍ້ຈໍາກັດ", nextSteps: "ຂັ້ນຕໍ່ໄປ", prerequisites: "ກ່ອນເລີ່ມ", quickLinks: "ສາລະບານ", screen: "ໜ້າຈໍ", summary: "ສະຫຼຸບ", troubleshooting: "ແກ້ໄຂບັນຫາ", updated: "ອັບເດດ", workflow: "ຂັ້ນຕອນໃຊ້ງານ" },
  my: { actions: "ခလုတ်များနှင့်လုပ်ဆောင်ချက်များ", commonMistakes: "တွေ့ရလေ့ရှိသောအမှားများ", config: "ဆက်တင်", expectedResults: "ရရှိသင့်သောရလဒ်", fields: "အချက်အလက်ကွက်ရှင်းလင်းချက်", home: "လမ်းညွှန် ပင်မ", limitations: "ကန့်သတ်ချက်များ", nextSteps: "နောက်လုပ်ဆောင်ရန်", prerequisites: "မစတင်မီ", quickLinks: "မာတိကာ", screen: "မျက်နှာပြင်", summary: "အကျဉ်းချုပ်", troubleshooting: "ပြဿနာဖြေရှင်းခြင်း", updated: "နောက်ဆုံးပြင်ဆင်", workflow: "အဆင့်လိုက်အသုံးပြုနည်း" },
  km: { actions: "ប៊ូតុង និងសកម្មភាព", commonMistakes: "កំហុសដែលជួបញឹកញាប់", config: "ការកំណត់", expectedResults: "លទ្ធផលដែលគួរទទួលបាន", fields: "ការពន្យល់វាល", home: "ទំព័រដើមសៀវភៅណែនាំ", limitations: "ដែនកំណត់", nextSteps: "ជំហានបន្ទាប់", prerequisites: "មុនចាប់ផ្តើម", quickLinks: "មាតិកា", screen: "ទំព័រ", summary: "សង្ខេប", troubleshooting: "ដោះស្រាយបញ្ហា", updated: "បានធ្វើបច្ចុប្បន្នភាព", workflow: "ជំហានប្រើប្រាស់" },
  vi: { actions: "Nút và thao tác", commonMistakes: "Lỗi thường gặp", config: "Cấu hình", expectedResults: "Kết quả mong đợi", fields: "Giải thích trường dữ liệu", home: "Trang chủ hướng dẫn", limitations: "Giới hạn", nextSteps: "Bước tiếp theo", prerequisites: "Trước khi bắt đầu", quickLinks: "Mục lục", screen: "Màn hình", summary: "Tóm tắt", troubleshooting: "Xử lý sự cố", updated: "Cập nhật", workflow: "Quy trình từng bước" },
  ms: { actions: "Butang dan tindakan", commonMistakes: "Kesilapan biasa", config: "Konfigurasi", expectedResults: "Hasil dijangka", fields: "Panduan medan", home: "Laman utama manual", limitations: "Had", nextSteps: "Langkah seterusnya", prerequisites: "Sebelum mula", quickLinks: "Kandungan", screen: "Skrin", summary: "Ringkasan", troubleshooting: "Penyelesaian masalah", updated: "Dikemas kini", workflow: "Aliran kerja langkah demi langkah" },
  id: { actions: "Tombol dan tindakan", commonMistakes: "Kesalahan umum", config: "Konfigurasi", expectedResults: "Hasil yang diharapkan", fields: "Panduan field", home: "Beranda panduan", limitations: "Batasan", nextSteps: "Langkah berikutnya", prerequisites: "Sebelum mulai", quickLinks: "Daftar isi", screen: "Layar", summary: "Ringkasan", troubleshooting: "Pemecahan masalah", updated: "Diperbarui", workflow: "Alur kerja langkah demi langkah" },
  fil: { actions: "Mga button at aksyon", commonMistakes: "Karaniwang pagkakamali", config: "Config", expectedResults: "Inaasahang resulta", fields: "Gabay sa field", home: "Manual home", limitations: "Limitasyon", nextSteps: "Susunod na hakbang", prerequisites: "Bago magsimula", quickLinks: "Nilalaman", screen: "Screen", summary: "Buod", troubleshooting: "Pag-troubleshoot", updated: "Na-update", workflow: "Step-by-step na daloy" },
};

type ManualPageProps = {
  params: Promise<{ screen: string }>;
  searchParams?: Promise<{ lang?: string }>;
};

export default async function ManualPage({ params, searchParams }: ManualPageProps) {
  const { screen } = await params;
  const query = searchParams ? await searchParams : {};
  const requestedLanguage = normalizeLanguage(query.lang);
  const language: LanguageCode = requestedLanguage === "th" ? "th" : "en";
  const manualHome = `/manual?lang=${language}&screen=${encodeURIComponent(screen)}`;
  if (!manualScreens.has(screen)) redirect(manualHome);

  const labels = manualUi[language] ?? manualUi.en;
  let manual: ManualFile;
  try {
    manual = await readManual(screen);
  } catch (error) {
    if (isFileNotFound(error)) redirect(manualHome);
    throw error;
  }
  const content = manual.translations[language] ?? manual.translations.en ?? manual.translations.th;
  if (!content) redirect(manualHome);
  const sections = [
    { id: "summary", title: labels.summary, show: true },
    { id: "sources", title: language === "th" ? "แหล่งอ้างอิง" : "Sources", show: Boolean(content.sources?.length) },
    { id: "prerequisites", title: labels.prerequisites, show: Boolean(content.prerequisites?.length) },
    { id: "workflow", title: labels.workflow, show: Boolean(content.workflow.length) },
    { id: "fields", title: labels.fields, show: Boolean(content.fields?.length) },
    { id: "actions", title: labels.actions, show: Boolean(content.actions?.length) },
    { id: "config", title: labels.config, show: Boolean(content.config.length) },
    { id: "expected-results", title: labels.expectedResults, show: Boolean(content.expectedResults?.length) },
    { id: "common-mistakes", title: labels.commonMistakes, show: Boolean(content.commonMistakes?.length) },
    { id: "troubleshooting", title: labels.troubleshooting, show: Boolean(content.troubleshooting?.length) },
    { id: "limitations", title: labels.limitations, show: Boolean(content.limitations.length) },
    { id: "next-steps", title: labels.nextSteps, show: Boolean(content.nextSteps?.length) },
  ].filter((item) => item.show);

  return (
    <main className="min-h-dvh w-full max-w-none bg-background px-2 py-2 text-foreground sm:px-3 lg:px-4" lang={language}>
      <article className="grid w-full max-w-none gap-3">
        <header className="grid w-full gap-3 rounded-2xl border border-border bg-card p-3 shadow-sm sm:p-4">
          <div className="flex w-full flex-wrap items-center justify-between gap-2">
            <Link className="manual-link w-fit" href={`/manual?lang=${language}&screen=${screen}`}>
              <Home size={16} />
              <span>{labels.home}</span>
            </Link>
            <ThemeToggle language={language} />
          </div>
          <div className="grid w-full gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(280px,360px)]">
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase tracking-normal text-muted-foreground">{labels.screen}: {manual.screen} · {labels.updated}: {manual.updatedAt}</p>
              <h1 className="mt-1 text-2xl font-semibold tracking-normal sm:text-3xl">{content.title}</h1>
              <p className="mt-2 max-w-[92ch] text-sm leading-6 text-muted-foreground sm:text-base">{content.summary ?? content.objective}</p>
            </div>
            <div className="grid min-w-0 gap-2 rounded-2xl border border-border bg-background p-3">
              <div className="flex items-center gap-2 text-sm font-semibold">
                <BookOpen size={17} />
                <span>{labels.quickLinks}</span>
              </div>
              <nav className="grid max-h-48 gap-1 overflow-auto text-sm" aria-label={labels.quickLinks}>
                {sections.map((section) => (
                  <a className="rounded-xl px-2 py-1.5 text-muted-foreground hover:bg-accent hover:text-foreground" href={`#${section.id}`} key={section.id}>
                    {section.title}
                  </a>
                ))}
              </nav>
            </div>
          </div>
          <nav className="flex w-full flex-wrap gap-1.5" aria-label="Manual languages">
            {manualLanguages.map((item) => (
              <Link
                aria-current={item.code === language ? "page" : undefined}
                className={`min-h-8 rounded-full border px-3 py-1.5 text-xs font-semibold ${item.code === language ? "border-primary bg-primary text-primary-foreground" : "border-border bg-background text-muted-foreground"}`}
                href={`/manual/${screen}?lang=${item.code}`}
                key={item.code}
              >
                {item.name}
              </Link>
            ))}
          </nav>
        </header>

        <div className="grid w-full max-w-none gap-3 xl:grid-cols-[280px_minmax(0,1fr)]">
          <aside className="hidden h-fit rounded-2xl border border-border bg-card p-3 shadow-sm xl:sticky xl:top-3 xl:grid">
            <h2 className="flex items-center gap-2 text-sm font-semibold"><ListChecks size={16} /> {labels.quickLinks}</h2>
            <nav className="mt-2 grid gap-1 text-sm" aria-label={labels.quickLinks}>
              {sections.map((section) => (
                <a className="rounded-xl px-2 py-1.5 text-muted-foreground hover:bg-accent hover:text-foreground" href={`#${section.id}`} key={section.id}>
                  {section.title}
                </a>
              ))}
            </nav>
          </aside>

          <div className="grid min-w-0 gap-3">
            <ManualHero id="summary" objective={content.objective} title={labels.summary} />
            <ManualSources
              id="sources"
              items={content.sources}
              title={language === "th" ? "แหล่งอ้างอิง" : "Sources"}
            />
            <ManualList icon="info" id="prerequisites" items={content.prerequisites} title={labels.prerequisites} />
            <ManualList icon="check" id="workflow" items={content.workflow} ordered title={labels.workflow} />
            <ManualPairs id="fields" items={content.fields} title={labels.fields} />
            <ManualPairs id="actions" items={content.actions} title={labels.actions} />
            <ManualList icon="settings" id="config" items={content.config} title={labels.config} />
            <ManualList icon="check" id="expected-results" items={content.expectedResults} title={labels.expectedResults} />
            <ManualList icon="alert" id="common-mistakes" items={content.commonMistakes} title={labels.commonMistakes} />
            <ManualPairs icon="help" id="troubleshooting" items={content.troubleshooting} title={labels.troubleshooting} />
            <ManualList icon="alert" id="limitations" items={content.limitations} title={labels.limitations} />
            <ManualList icon="check" id="next-steps" items={content.nextSteps} title={labels.nextSteps} />
          </div>
        </div>
      </article>
    </main>
  );
}

async function readManual(screen: string): Promise<ManualFile> {
  const manualPath = path.resolve(process.cwd(), "manual", `${screen}.json`);
  const raw = await readFile(manualPath, "utf8");
  return JSON.parse(raw) as ManualFile;
}

function ManualHero({ id, objective, title }: { id: string; objective: string; title: string }) {
  return (
    <section className="grid w-full gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm sm:p-4" id={id}>
      <h2 className="flex items-center gap-2 text-base font-semibold"><Info size={18} /> {title}</h2>
      <p className="max-w-[100ch] text-sm leading-6 text-muted-foreground sm:text-base">{objective}</p>
    </section>
  );
}

function ManualList({ icon = "info", id, items, ordered = false, title }: { icon?: "alert" | "check" | "info" | "settings"; id: string; items?: string[]; ordered?: boolean; title: string }) {
  if (!items?.length) return null;
  const ListTag = ordered ? "ol" : "ul";

  return (
    <section className="grid w-full gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm sm:p-4" id={id}>
      <h2 className="flex items-center gap-2 text-base font-semibold">{sectionIcon(icon)} {title}</h2>
      <ListTag className={`grid gap-2 text-sm leading-6 text-muted-foreground sm:grid-cols-2 ${ordered ? "list-decimal pl-5 sm:pl-6" : ""}`}>
        {items.map((item) => (
          <li className={`${ordered ? "rounded-xl border border-border bg-background px-3 py-2" : "rounded-xl border border-border bg-background px-3 py-2"}`} key={item}>
            {item}
          </li>
        ))}
      </ListTag>
    </section>
  );
}

function ManualPairs({ icon = "settings", id, items, title }: { icon?: "help" | "settings"; id: string; items?: ManualPair[]; title: string }) {
  if (!items?.length) return null;

  return (
    <section className="grid w-full gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm sm:p-4" id={id}>
      <h2 className="flex items-center gap-2 text-base font-semibold">{sectionIcon(icon)} {title}</h2>
      <div className="grid gap-2 md:grid-cols-2 2xl:grid-cols-3">
        {items.map((item) => (
          <div className="min-w-0 rounded-xl border border-border bg-background p-3" key={`${item.name}-${item.description}`}>
            <h3 className="text-sm font-semibold">{item.name}</h3>
            <p className="mt-1 text-sm leading-6 text-muted-foreground">{item.description}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

function sectionIcon(icon: "alert" | "check" | "help" | "info" | "settings") {
  if (icon === "alert") return <CircleAlert size={18} />;
  if (icon === "check") return <CheckCircle2 size={18} />;
  if (icon === "help") return <HelpCircle size={18} />;
  if (icon === "settings") return <Settings2 size={18} />;
  return <ClipboardList size={18} />;
}

function ManualSources({ id, items, title }: { id: string; items?: ManualSource[]; title: string }) {
  if (!items?.length) return null;

  return (
    <section className="grid w-full gap-2 rounded-2xl border border-border bg-card p-3 shadow-sm sm:p-4" id={id}>
      <h2 className="flex items-center gap-2 text-base font-semibold"><BookOpen size={18} /> {title}</h2>
      <div className="flex flex-wrap gap-2">
        {items.map((item) => (
          <a
            className="manual-link"
            href={item.href}
            key={item.href}
            rel="noreferrer"
            target="_blank"
          >
            <span>{item.name}</span>
          </a>
        ))}
      </div>
    </section>
  );
}

function isFileNotFound(error: unknown): boolean {
  return (error as { code?: unknown } | null)?.code === "ENOENT";
}
