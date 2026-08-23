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
const manualLanguages = LANGUAGES.filter((item) => item.code === "th" || item.code === "en");

const indexUi = {
  th: {
    title: "สารบัญคู่มือที่พร้อมใช้งาน",
    subtitle: "รวมคู่มือเบื้องต้นที่ตรวจเนื้อหาแล้ว เลือกหน้าที่ต้องการจากรายการด้านล่าง",
    language: "ภาษา",
    open: "เปิดคู่มือ",
    updated: "อัปเดต",
    coverage: "คู่มือชุดนี้ประกอบด้วย",
    coverageItems: ["แนวคิดและเป้าหมาย", "สิ่งที่ต้องเตรียมและขั้นตอนใช้งาน", "ผลลัพธ์ ข้อผิดพลาดที่พบบ่อย และข้อจำกัด", "ลิงก์แหล่งอ้างอิงของกฎและพฤติกรรมระบบ"],
  },
  en: {
    title: "Available BC Ai Account Guides",
    subtitle: "Basic guides whose content has been reviewed. Choose the screen you need below.",
    language: "Language",
    open: "Open guide",
    updated: "Updated",
    coverage: "These guides include",
    coverageItems: ["Concept and objective", "Prerequisites and workflow", "Results, common mistakes, and limitations", "Source links for system rules and behavior"],
  },
} as const;

export default async function ManualIndexPage({ searchParams }: ManualIndexProps) {
  const query = searchParams ? await searchParams : {};
  const requestedLanguage = normalizeLanguage(query.lang);
  const language: LanguageCode = requestedLanguage === "th" ? "th" : "en";
  const currentScreen = isManualScreen(query.screen) ? query.screen : null;
  const labels = language === "th" ? indexUi.th : indexUi.en;
  const manuals = (
    await Promise.all(
      manualScreens.map(async (screen) => {
        try {
          return await readManual(screen);
        } catch (error) {
          if (isFileNotFound(error)) return null;
          throw error;
        }
      }),
    )
  ).filter((manual): manual is ManualFile => manual !== null);

  return (
    <main className="min-h-dvh w-full max-w-none bg-background px-2 py-2 text-foreground sm:px-3 lg:px-4" lang={language}>
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
              {manualLanguages.map((item) => (
                <Link
                  aria-current={item.code === language ? "page" : undefined}
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
  const manualPath = path.resolve(process.cwd(), "manual", `${screen}.json`);
  const raw = await readFile(manualPath, "utf8");
  return JSON.parse(raw) as ManualFile;
}

function isManualScreen(screen: string | undefined): screen is typeof manualScreens[number] {
  return Boolean(screen && (manualScreens as readonly string[]).includes(screen));
}

function isFileNotFound(error: unknown): boolean {
  return (error as { code?: unknown } | null)?.code === "ENOENT";
}
