#!/usr/bin/env node
// authFetch() only *refreshes* an Authorization header that is already on the
// request — if the caller does not attach one, the request goes out anonymous,
// the backend answers 401, and a screen that reads `res?.items` renders as
// "no data" instead of failing. This finds callers that never attach it.
//
//   node tools/audit-auth-fetch.mjs
//
// Exits non-zero when a call has no Authorization in scope, so it can guard a
// change as well as report on one.
import { readFileSync, readdirSync } from "node:fs";
import { resolve, dirname, join, sep } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const root = resolve(repoRoot, "frontend/src");

const files = [];
const walk = (dir) => {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) walk(full);
    else if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) files.push(full);
  }
};
walk(root);

// Read the argument list of a call, respecting nesting and strings.
function callArguments(source, openIndex) {
  let depth = 0;
  let quote = "";
  for (let i = openIndex; i < source.length; i += 1) {
    const char = source[i];
    if (quote) {
      if (char === "\\") i += 1;
      else if (char === quote) quote = "";
      continue;
    }
    if (char === '"' || char === "'" || char === "`") {
      quote = char;
      continue;
    }
    if (char === "(") depth += 1;
    else if (char === ")") {
      depth -= 1;
      if (depth === 0) return source.slice(openIndex + 1, i);
    }
  }
  return "";
}

const findings = [];
for (const file of files) {
  const source = readFileSync(file, "utf8");
  if (!source.includes("authFetch(")) continue;
  const relative = file.slice(repoRoot.length + 1).split(sep).join("/");
  if (relative.endsWith("lib/client-auth-session.ts")) continue; // the helper itself

  for (const match of source.matchAll(/\bauthFetch\(/g)) {
    const open = (match.index ?? 0) + match[0].length - 1;
    const args = callArguments(source, open);
    const line = source.slice(0, open).split(/\r?\n/).length;
    const text = source.split(/\r?\n/)[line - 1] ?? "";
    if (/^\s*(\/\/|\*|\/\*)/.test(text)) continue; // a comment mentioning it

    if (/Authorization/i.test(args)) continue;
    // requestHeaders(auth) builds Content-Type + x-bc-backend-url + Authorization.
    if (/\brequestHeaders\(/.test(args)) continue;
    // Routes that are the way *in*: they cannot require a session yet.
    if (/["'`]\/api\/auth\//.test(args)) continue;
    // A wrapper forwarding its own init: the callers set the header (checked by
    // the file carrying an Authorization of its own).
    if (/^\s*\w+\s*,\s*\w+\s*$/.test(args) && /Authorization/i.test(source)) continue;
    // Headers assembled in the same function, just above the call.
    const before = source.slice(Math.max(0, open - 900), open);
    if (/Authorization/i.test(before) && /headers/i.test(args)) continue;

    findings.push({ file: relative, line, call: args.replace(/\s+/g, " ").slice(0, 90) });
  }
}

if (findings.length === 0) {
  console.log("every authFetch call attaches an Authorization header");
  process.exit(0);
}

const byFile = new Map();
for (const finding of findings) byFile.set(finding.file, (byFile.get(finding.file) ?? 0) + 1);
console.log(`authFetch calls with no Authorization header: ${findings.length} in ${byFile.size} files\n`);
for (const [file, count] of [...byFile].sort((a, b) => b[1] - a[1])) {
  console.log(`${String(count).padStart(3)} ${file}`);
  for (const finding of findings.filter((f) => f.file === file).slice(0, 3)) {
    console.log(`      :${finding.line}  authFetch(${finding.call})`);
  }
}
process.exit(1);
