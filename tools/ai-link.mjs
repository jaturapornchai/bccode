#!/usr/bin/env node
// Link Claude Code's skill folder to the shared .agents/skills/ (single source for every AI).
//
// Codex, ZCode, Antigravity and Gemini CLI discover project skills from .agents/skills/ natively;
// Claude Code only scans .claude/skills/. Instead of keeping two copies (which drift), this creates
// .claude/skills -> .agents/skills as a directory junction on Windows (no admin / Developer Mode
// needed) or a symlink elsewhere. .claude/ is gitignored, so re-run this once per clone.
//
// Usage: npm run ai:link   (or: node tools/ai-link.mjs)

import { existsSync, lstatSync, mkdirSync, readdirSync, symlinkSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = dirname(dirname(fileURLToPath(import.meta.url)));
const source = join(repoRoot, ".agents", "skills");
const link = join(repoRoot, ".claude", "skills");

if (!existsSync(source)) {
  console.error(`ai-link: ${source} not found`);
  process.exit(1);
}

if (pathExists(link)) {
  // lstat reports a Windows junction as a symbolic link, so one check covers both platforms.
  if (lstatSync(link).isSymbolicLink()) {
    console.log(`ai-link: ${link} already linked`);
    process.exit(0);
  }
  console.error(`ai-link: ${link} is a real folder - move its contents into .agents/skills/ and delete it first`);
  process.exit(1);
}

mkdirSync(dirname(link), { recursive: true });
symlinkSync(source, link, process.platform === "win32" ? "junction" : "dir");
console.log(`ai-link: ${link} -> ${source} (${readdirSync(source).join(", ")})`);

// existsSync follows links, so a dangling link would look absent; lstat sees the link itself.
function pathExists(path) {
  try {
    lstatSync(path);
    return true;
  } catch {
    return false;
  }
}
