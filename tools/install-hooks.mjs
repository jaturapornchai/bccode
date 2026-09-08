#!/usr/bin/env node
// Install the tracked hooks in .githooks/ into this clone's .git/hooks/.
//
// Why copy instead of `git config core.hooksPath .githooks`: hooksPath REPLACES the hook
// directory wholesale, which silently disables any hook that already lives in .git/hooks —
// this repo's owner has a post-commit hook that refreshes an Obsidian vault, and switching
// hooksPath killed it without a word. Copying leaves local hooks alone.
//
// Trade-off: because these are copies, re-run this after pulling a change to .githooks/.
//
// Usage: npm run hooks:install   (or: node tools/install-hooks.mjs)

import { chmodSync, copyFileSync, existsSync, mkdirSync, readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = dirname(dirname(fileURLToPath(import.meta.url)));
const srcDir = join(repoRoot, ".githooks");
const gitDir = join(repoRoot, ".git");

if (!existsSync(srcDir)) {
  console.error(`install-hooks: ${srcDir} not found`);
  process.exit(1);
}
if (!existsSync(gitDir)) {
  console.error("install-hooks: no .git directory — run this from a git clone");
  process.exit(1);
}

// A worktree or submodule has .git as a file pointing at the real git dir.
let hooksParent = gitDir;
if (statSync(gitDir).isFile()) {
  const pointer = readFileSync(gitDir, "utf8").trim();
  const match = /^gitdir:\s*(.+)$/.exec(pointer);
  if (!match) {
    console.error(`install-hooks: cannot parse .git file: ${pointer}`);
    process.exit(1);
  }
  hooksParent = match[1];
}

const destDir = join(hooksParent, "hooks");
mkdirSync(destDir, { recursive: true });

const installed = [];
for (const name of readdirSync(srcDir)) {
  if (name.startsWith(".") || name.endsWith(".sample") || name.endsWith(".md")) continue;
  const dest = join(destDir, name);
  copyFileSync(join(srcDir, name), dest);
  try {
    chmodSync(dest, 0o755);
  } catch {
    // Windows filesystems ignore the mode; git for Windows runs the hook regardless.
  }
  installed.push(name);
}

if (installed.length === 0) {
  console.error("install-hooks: nothing to install");
  process.exit(1);
}

console.log(`installed hooks -> ${destDir}: ${installed.join(", ")}`);
console.log("re-run this after pulling changes to .githooks/");
