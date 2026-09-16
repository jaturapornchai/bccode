#!/usr/bin/env node
// Move hard-coded Thai out of a screen and into languages.tsv, one screen at a
// time. Replaces the throw-away scratch scripts that were rewritten per screen.
//
//   node tools/i18n-sweep.mjs scan  <file...>            what is still hard-coded
//   node tools/i18n-sweep.mjs plan  <file...> --prefix=fa  propose a key per string
//   node tools/i18n-sweep.mjs apply <file...>            rewrite code + append rows
//
// Working files live in .i18n-sweep/ (git-ignored). Read the plan before apply:
// reuse decisions are the part a human has to judge, because a row whose Thai
// matches can still have been translated for a different sense.
//
// While the project is in the Thai-only dev phase (AGENTS.md 2026-09-16), new
// rows carry the Thai in every language column. That keeps the row complete —
// an incomplete row makes the backend serve the raw key — and makes the set
// that still needs translating trivial to find: every column equals `th`.
import { readFileSync, writeFileSync, mkdirSync, existsSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const workDir = resolve(repoRoot, ".i18n-sweep");
const tsvPath = resolve(repoRoot, "backend/assets/language/languages.tsv");
const THAI = /[฀-฾เ-๿]/;
const LANGUAGE_COLUMNS = 12; // th en cn ja km ko lo my vi ms id fil

const args = process.argv.slice(2);
const command = args[0];
const files = args.slice(1).filter((a) => !a.startsWith("--"));
const flag = (name, fallback) => {
  const hit = args.find((a) => a.startsWith(`--${name}=`));
  return hit ? hit.slice(name.length + 3) : fallback;
};

const workFile = (name) => resolve(workDir, name);
const readWork = (name) => JSON.parse(readFileSync(workFile(name), "utf8"));
const writeWork = (name, value) => {
  mkdirSync(workDir, { recursive: true });
  writeFileSync(workFile(name), JSON.stringify(value, null, 1));
};

// ---------------------------------------------------------------- scan

// Three shapes can be rewritten mechanically; a template literal that
// interpolates a value has to become a {0} row by hand, so it is only reported.
function scanFile(file) {
  const lines = readFileSync(resolve(repoRoot, file), "utf8").split(/\r?\n/);
  const plain = [];
  const jsx = [];
  const manual = [];
  let inBlockComment = false;

  lines.forEach((line, index) => {
    const trimmed = line.trim();
    if (inBlockComment) {
      if (trimmed.includes("*/")) inBlockComment = false;
      return;
    }
    if (trimmed.startsWith("/*")) {
      if (!trimmed.includes("*/")) inBlockComment = true;
      return;
    }
    if (trimmed.startsWith("//") || trimmed.startsWith("*")) return;
    if (!THAI.test(line)) return;
    const at = index + 1;

    for (const match of line.matchAll(/`([^`]*)`/g)) {
      if (!THAI.test(match[1])) continue;
      // Thai that only sits inside an interpolated bilingual literal is already
      // routed by the helper reading it; the template itself has none.
      if (!THAI.test(match[1].replace(/\bth:\s*"[^"]*"/g, ""))) continue;
      if (match[1].includes("${")) manual.push({ file, at, text: match[1], code: trimmed });
      else plain.push({ file, at, text: match[1], backtick: true });
    }

    let matched = false;
    for (const match of line.matchAll(/"([^"\\]*)"|'([^'\\]*)'/g)) {
      const text = match[1] ?? match[2];
      if (!text || !THAI.test(text)) continue;
      // The fallback argument of an existing tr()/backendText() call is fine.
      const keyAt = line.search(/"[a-z0-9_.]+"\s*,/);
      if (keyAt >= 0 && match.index > keyAt) continue;
      // So is the Thai side of a bilingual literal — the catalog helper that
      // reads it already resolves the row, so wrapping it would resolve twice.
      if (/(?:^|[\s{,(])(?:th|en|labelTh|titleTh|helperTh):\s*$/.test(line.slice(0, match.index))) continue;
      plain.push({ file, at, text });
      matched = true;
    }

    for (const match of line.matchAll(/>([^<>{}]*)</g)) {
      const text = match[1].trim();
      if (!text || !THAI.test(text)) continue;
      jsx.push({ file, at, text });
      matched = true;
    }

    // A Thai run alone on its line inside a multi-line JSX element body.
    if (!matched && /^[^<>{}"'`]+$/.test(trimmed) && THAI.test(trimmed)) {
      jsx.push({ file, at, text: trimmed, bare: true });
    }
  });

  return { plain, jsx, manual };
}

function scan(report = true) {
  const found = { plain: [], jsx: [], manual: [] };
  for (const file of files) {
    const one = scanFile(file);
    for (const shape of ["plain", "jsx", "manual"]) found[shape].push(...one[shape]);
  }
  writeWork("found.json", found);
  if (!report) return found;
  const unique = new Set([...found.plain, ...found.jsx].map((r) => r.text));
  console.log(
    `plain ${found.plain.length} | jsx ${found.jsx.length} | template ${found.manual.length} | unique ${unique.size}`,
  );
  for (const row of found.manual) console.log(`  BY HAND ${row.file}:${row.at}  ${row.text}`);
  return found;
}

// ---------------------------------------------------------------- plan

function readTsv() {
  const raw = readFileSync(tsvPath, "utf8");
  return { raw, eol: raw.includes("\r\n") ? "\r\n" : "\n", lines: raw.split(/\r?\n/) };
}

function plan() {
  const found = scan(false);
  const prefix = flag("prefix");
  if (!prefix) throw new Error("plan needs --prefix=<short screen prefix>");

  const { lines } = readTsv();
  const byThai = new Map();
  const taken = new Set();
  for (const line of lines) {
    const cells = line.split("\t");
    if (cells.length !== 13) continue;
    taken.add(cells[0]);
    const list = byThai.get(cells[1]) ?? [];
    list.push({ key: cells[0], en: cells[2] });
    byThai.set(cells[1], list);
  }

  // Keep decisions already made by hand: `plan` is run repeatedly to re-read the
  // review list, and silently discarding an edited key is how a checked reuse
  // choice gets lost. --fresh starts over.
  const previous = !args.includes("--fresh") && existsSync(workFile("plan.json")) ? readWork("plan.json") : {};
  const plan = {};
  const review = [];
  let minted = 0;
  for (const row of [...found.plain, ...found.jsx]) {
    if (plan[row.text]) continue;
    if (previous[row.text]) {
      plan[row.text] = previous[row.text];
      taken.add(previous[row.text].key);
      review.push(`KEPT  ${previous[row.text].key}\t${row.text}`);
      continue;
    }
    const hits = byThai.get(row.text) ?? [];
    if (hits.length) {
      plan[row.text] = { key: hits[0].key, reuse: true };
      review.push(`REUSE ${hits[0].key}\t${row.text}\t[${hits.map((h) => `${h.key}=${h.en}`).join(" | ")}]`);
      continue;
    }
    let key = `${prefix}_${++minted}`;
    while (taken.has(key)) key = `${prefix}_${++minted}`;
    taken.add(key);
    plan[row.text] = { key, reuse: false };
    review.push(`MINT  ${key}\t${row.text}`);
  }

  writeWork("plan.json", plan);
  console.log(review.join("\n"));
  console.log(
    `\ntotal ${Object.keys(plan).length} | reuse ${review.filter((r) => r.startsWith("REUSE")).length}` +
      ` | kept ${review.filter((r) => r.startsWith("KEPT")).length} | mint ${minted}` +
      `\nCheck the English of every REUSE line before apply, then rename any mint key in .i18n-sweep/plan.json` +
      ` (edits survive a re-run; use --fresh to discard them).`,
  );
}

// ---------------------------------------------------------------- apply

function appendRows(plan) {
  const { eol, lines } = readTsv();
  const known = new Map(lines.map((l) => [l.split("\t")[0], l.split("\t")[1]]));
  const fresh = [];
  for (const [th, value] of Object.entries(plan)) {
    if (value.reuse) continue;
    const existing = known.get(value.key);
    if (existing !== undefined) {
      if (existing !== th) throw new Error(`${value.key} already exists with different Thai: "${existing}"`);
      continue;
    }
    // Thai in every language column: complete row, still obviously untranslated.
    fresh.push([value.key, ...Array(LANGUAGE_COLUMNS).fill(th)].join("\t"));
    known.set(value.key, th);
  }
  if (fresh.length) {
    const body = lines.filter((l, i) => l || i < lines.length - 1);
    writeFileSync(tsvPath, [...body, ...fresh].join(eol) + eol);
  }
  const check = readFileSync(tsvPath, "utf8").split(/\r?\n/).filter(Boolean);
  const keys = check.map((l) => l.split("\t")[0]);
  console.log(
    `rows added ${fresh.length} | rows ${check.length} | malformed ${check.filter((l) => l.split("\t").length !== 13).length}` +
      ` | duplicate keys ${keys.length - new Set(keys).size}`,
  );
}

function apply() {
  const found = scan(false);
  const plan = readWork("plan.json");
  const escape = (s) => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  let changed = 0;
  const skipped = [];

  for (const file of files) {
    const path = resolve(repoRoot, file);
    const src = readFileSync(path, "utf8");
    const eol = src.includes("\r\n") ? "\r\n" : "\n";
    const lines = src.split(/\r?\n/);

    for (const row of found.plain.filter((r) => r.file === file)) {
      const key = plan[row.text]?.key;
      if (!key) continue;
      const i = row.at - 1;
      const quoted = row.backtick
        ? new RegExp(`\`${escape(row.text)}\``)
        : new RegExp(`(["'])${escape(row.text)}\\1`);
      if (!quoted.test(lines[i])) {
        skipped.push(`${file}:${row.at} ${row.text}`);
        continue;
      }
      // A JSX attribute (placeholder="…") needs braces around the expression.
      const isAttribute = new RegExp(`=(["'\`])${escape(row.text)}\\1`).test(lines[i]);
      const call = `tr("${key}", "${row.text}")`;
      lines[i] = lines[i].replace(quoted, isAttribute ? `{${call}}` : call);
      changed += 1;
    }

    for (const row of found.jsx.filter((r) => r.file === file)) {
      const key = plan[row.text]?.key;
      if (!key) continue;
      const i = row.at - 1;
      const call = `{tr("${key}", "${row.text}")}`;
      if (row.bare) {
        if (lines[i].trim() !== row.text) {
          skipped.push(`${file}:${row.at} ${row.text}`);
          continue;
        }
        lines[i] = lines[i].replace(row.text, call);
        changed += 1;
        continue;
      }
      const node = new RegExp(`>(\\s*)${escape(row.text)}(\\s*)<`);
      if (!node.test(lines[i])) {
        skipped.push(`${file}:${row.at} ${row.text}`);
        continue;
      }
      lines[i] = lines[i].replace(node, `>$1${call}$2<`);
      changed += 1;
    }

    writeFileSync(path, lines.join(eol));
  }

  console.log(`rewrote ${changed} | skipped ${skipped.length}`);
  skipped.forEach((s) => console.log("  ", s));
  appendRows(plan);
  if (found.manual.length) {
    console.log(`\n${found.manual.length} template literals still need a {0} row by hand:`);
    found.manual.forEach((row) => console.log(`  ${row.file}:${row.at}  ${row.text}`));
  }
}

const commands = { scan: () => scan(), plan, apply };
if (!commands[command] || files.length === 0) {
  console.error("usage: node tools/i18n-sweep.mjs <scan|plan|apply> <file...> [--prefix=xx]");
  process.exit(1);
}
commands[command]();
