#!/usr/bin/env node
import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { resolveLocaleKey } from "../src/lib/webapp/constants.js";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(scriptDir, "..");
const repoRoot = path.resolve(frontendRoot, "..");
const sourceRoot = path.join(frontendRoot, "src");
const sourceExtensions = new Set([".js", ".svelte", ".ts"]);

async function sourceFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const nested = await Promise.all(
    entries.map((entry) => {
      const entryPath = path.join(directory, entry.name);
      if (entry.isDirectory()) return sourceFiles(entryPath);
      return sourceExtensions.has(path.extname(entry.name)) ? [entryPath] : [];
    })
  );
  return nested.flat();
}

function staticLocaleKeys(source) {
  const keys = [];
  const pattern = /\b(t|at)\(\s*["']([^"']+)["']/g;
  for (const match of source.matchAll(pattern)) {
    const rawKey = match[1] === "at" ? `admin_${match[2]}` : match[2];
    keys.push(resolveLocaleKey(rawKey));
  }
  return keys;
}

const catalogs = Object.fromEntries(
  await Promise.all(
    ["zh", "en"].map(async (language) => [
      language,
      JSON.parse(await readFile(path.join(repoRoot, "locales", `${language}.json`), "utf8")),
    ])
  )
);
const usedBy = new Map();

for (const file of await sourceFiles(sourceRoot)) {
  const relative = path.relative(frontendRoot, file);
  for (const key of staticLocaleKeys(await readFile(file, "utf8"))) {
    if (!usedBy.has(key)) usedBy.set(key, new Set());
    usedBy.get(key).add(relative);
  }
}

const missing = [];
for (const [key, files] of usedBy) {
  for (const [language, catalog] of Object.entries(catalogs)) {
    if (!(key in catalog)) {
      missing.push(`${language}: ${key} (${[...files].sort().join(", ")})`);
    }
  }
}

if (missing.length) {
  console.error(`Missing ${missing.length} static locale entries:\n${missing.sort().join("\n")}`);
  process.exitCode = 1;
} else {
  console.log(`Validated ${usedBy.size} static locale keys in zh and en.`);
}
