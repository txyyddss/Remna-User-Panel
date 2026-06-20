#!/usr/bin/env node
import { createHash } from "node:crypto";
import { readFile, readdir, unlink, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { brotliCompressSync, constants as zlibConstants, gzipSync } from "node:zlib";

import { transform } from "esbuild";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const JS_MINIFY_TARGET = "es2020";
const templatesDir = path.join(repoRoot, "internal", "webassets", "templates");
const JS_BUILDS = [
  {
    sourcePath: path.join(templatesDir, "subscription_webapp.js"),
    outputPrefix: "subscription_webapp.min",
  },
  {
    sourcePath: path.join(templatesDir, "subscription_webapp_admin.js"),
    outputPrefix: "subscription_webapp_admin.min",
  },
];
const CSS_BUILDS = [
  {
    sourcePath: path.join(templatesDir, "subscription_webapp.css"),
    outputPrefix: "subscription_webapp",
  },
  {
    sourcePath: path.join(templatesDir, "subscription_webapp_admin.css"),
    outputPrefix: "subscription_webapp_admin",
  },
];

function normalizeLineEndings(value) {
  return value.replace(/\r\n/g, "\n");
}

async function removeOldHashedAssets(assetDir, pattern, keepNames) {
  const keep = new Set(Array.isArray(keepNames) ? keepNames : [keepNames]);
  const entries = await readdir(assetDir, { withFileTypes: true });
  await Promise.all(
    entries
      .filter((entry) => entry.isFile() && pattern.test(entry.name) && !keep.has(entry.name))
      .map((entry) => unlink(path.join(assetDir, entry.name)))
  );
}

async function writePrecompressedAssets(outputPath, body) {
  const buffer = Buffer.isBuffer(body) ? body : Buffer.from(body, "utf8");
  const gzipBody = gzipSync(buffer, { level: 9 });
  const brotliBody = brotliCompressSync(buffer, {
    params: {
      [zlibConstants.BROTLI_PARAM_QUALITY]: 11,
    },
  });

  await Promise.all([
    writeFile(`${outputPath}.gz`, gzipBody),
    writeFile(`${outputPath}.br`, brotliBody),
  ]);

  return {
    gzip: gzipBody.length,
    brotli: brotliBody.length,
  };
}

async function buildJsAsset({ sourcePath, outputPrefix }) {
  const rawSource = await readFile(sourcePath, "utf8");
  const strippedSource = normalizeLineEndings(rawSource);
  const result = await transform(strippedSource, {
    charset: "utf8",
    legalComments: "none",
    loader: "js",
    minify: true,
    target: JS_MINIFY_TARGET,
  });

  const code = `${result.code.replace(/[ \t]+$/gm, "").trimEnd()}\n`;
  const hash = createHash("sha256").update(code, "utf8").digest("hex").slice(0, 8);
  const outputPath = path.join(path.dirname(sourcePath), `${outputPrefix}.${hash}.js`);
  const outputName = path.basename(outputPath);
  const escapedPrefix = outputPrefix.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

  await removeOldHashedAssets(
    path.dirname(sourcePath),
    new RegExp(`^${escapedPrefix}\\.[0-9a-f]{8}\\.js(?:\\.(?:br|gz))?$`),
    [outputName, `${outputName}.br`, `${outputName}.gz`]
  );
  await writeFile(outputPath, code, "utf8");
  const compressedJs = await writePrecompressedAssets(outputPath, code);
  console.log(
    `Wrote ${path.relative(repoRoot, outputPath)} (${Buffer.byteLength(code, "utf8")} bytes, gzip ${compressedJs.gzip}, br ${compressedJs.brotli})`
  );
}

async function buildCssAsset({ sourcePath, outputPrefix }) {
  const rawCss = await readFile(sourcePath, "utf8");
  const cssResult = await transform(rawCss, {
    legalComments: "none",
    loader: "css",
    minify: true,
  });
  const css = `${cssResult.code.replace(/[ \t]+$/gm, "").trimEnd()}\n`;
  const cssHash = createHash("sha256").update(css, "utf8").digest("hex").slice(0, 8);
  const cssOutputPath = path.join(path.dirname(sourcePath), `${outputPrefix}.${cssHash}.css`);
  const cssOutputName = path.basename(cssOutputPath);
  const escapedPrefix = outputPrefix.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  await removeOldHashedAssets(
    path.dirname(sourcePath),
    new RegExp(`^${escapedPrefix}\\.[0-9a-f]{8}\\.css(?:\\.(?:br|gz))?$`),
    [cssOutputName, `${cssOutputName}.br`, `${cssOutputName}.gz`]
  );
  await writeFile(cssOutputPath, css, "utf8");
  const compressedCss = await writePrecompressedAssets(cssOutputPath, css);
  console.log(
    `Wrote ${path.relative(repoRoot, cssOutputPath)} (${Buffer.byteLength(css, "utf8")} bytes, gzip ${compressedCss.gzip}, br ${compressedCss.brotli})`
  );
}

async function main() {
  for (const build of JS_BUILDS) {
    await buildJsAsset(build);
  }
  for (const build of CSS_BUILDS) {
    await buildCssAsset(build);
  }
}

await main();
