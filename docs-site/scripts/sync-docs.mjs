import { copyFile, mkdir, readdir, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const siteRoot = path.resolve(fileURLToPath(new URL('..', import.meta.url)));
const repoRoot = path.resolve(siteRoot, '..');
const sourceDir = path.join(repoRoot, 'docs');
const outputDir = path.join(siteRoot, 'src', 'content', 'docs');

const descriptions = {
  'getting-started/demo.md': 'Remnawave Minishop 静态演示模式的工作原理, 以及为什么它只随文档一起构建。',
  'index.md': 'Remnawave Telegram Mini App 的启动、配置和维护文档。',
  'getting-started/overview.md': 'Remnawave Minishop 的组成, 以及机器人、Mini App、backend、worker 和 Remnawave Panel 之间的关系。',
  'getting-started/setup.md': '通过 Docker Compose 启动 Remnawave Minishop 的最简路径。',
  'getting-started/configuration.md': '最小 .env、引导密钥和通过 Web App 管理界面配置。',
  'getting-started/deployment.md': 'Docker Compose、反向代理、TLS、镜像、更新和备份。',
  'configuration/security.md': 'Minishop 的密钥、公开 URL、管理员访问和基本安全措施。',
  'configuration/env-vars.md': 'Remnawave Minishop 环境变量完整参考。',
  'features/core.md': 'Remnawave Minishop 的用户和管理场景。',
  'features/payments.md': '支付商、支付按钮和 Webhook 处理。',
  'features/subscriptions.md': '周期和流量套餐、Premium Squads、HWID 设备和订阅生命周期。',
  'features/notifications.md': 'Remnawave Minishop 用户、管理员和服务的 Telegram 和 Email 通知渠道。',
  'features/tariffs.md': '套餐目录、周期/流量模式、Premium Squads 和 HWID 设备。',
  'features/web-app.md': 'Telegram Mini App、公开指南、代理和推荐链接。',
  'features/telegram-auth.md': 'Telegram Mini Apps initData、Telegram OAuth、BotFather 和 Telegram 登录设置。',
  'features/email-login.md': 'SMTP、一次性验证码、Magic Link、密码登录和邮箱账户绑定。',
  'features/webapp-themes.md': '自定义主题、CSS Token、资源和主题创建流程。',
  'features/admin-panel.md': '管理面板功能、用户管理、设置、套餐和客服。',
  'features/backups.md': '自动备份、归档发送到 Telegram、本地存储和从管理界面恢复数据库/Compose 文件夹。',
  'features/support.md': '用户工单、管理端工单列表、通知和客服限制。',
  'migrations/index.md': '从其他机器人迁移到 Remnawave Minishop 的现成方案。',
  'migrations/remnawave-tg-shop.md': '从旧 remnawave-tg-shop 迁移数据到 Minishop 分离架构。',
  'migrations/remnashop.md': '通过 install wizard 或 Go importer 从 Remnashop 导入数据。',
  'troubleshooting/issues.md': '启动、Webhook、Mini App 和支付常见问题的快速检查清单。',
  'troubleshooting/logs.md': '诊断 backend、worker、frontend、迁移和 Webhook 时应查看哪些日志。',
  'troubleshooting/maintenance.md': '更新、迁移、备份和生产环境检查。',
  'architecture.md': 'Backend、Frontend、Worker 和基础设施服务的简要架构。',
};

const imageExtensions = new Set(['.avif', '.gif', '.jpeg', '.jpg', '.png', '.svg', '.webp']);

function yamlString(value) {
  return JSON.stringify(value);
}

function toPosix(relativePath) {
  return relativePath.split(path.sep).join('/');
}

function outputRelativePath(sourceRelativePath) {
  if (sourceRelativePath === 'index.md') {
    return 'index.md';
  }
  if (!sourceRelativePath.includes('/')) {
    return `reference/${sourceRelativePath}`;
  }
  return sourceRelativePath;
}

function pagePathForSource(sourceRelativePath, hash = '') {
  const output = outputRelativePath(sourceRelativePath).replace(/\.md$/i, '');
  const route = output === 'index' ? '/' : `/${output.replace(/\/index$/u, '')}/`;
  return `${route}${hash}`;
}

function titleForRelativePath(relativePath) {
  const baseName = path.posix.basename(relativePath, '.md');
  return baseName;
}

function extractTitle(relativePath, content) {
  const match = content.match(/^#\s+(.+?)\s*$/m);
  return match?.[1] ?? titleForRelativePath(relativePath);
}

function stripFirstHeading(content) {
  return content.replace(/^#\s+.+?\s*\r?\n+/, '');
}

function rewriteMarkdownLinks(markdown, sourceRelativePath) {
  const sourceDirectory = path.posix.dirname(sourceRelativePath);
  return markdown.replace(/\]\((?!https?:\/\/|mailto:|tel:|\/|#)([^)\s]+\.md)(#[^)]+)?\)/g, (match, target, hash = '') => {
    const resolvedTarget = path.posix.normalize(path.posix.join(sourceDirectory, target));
    return `](${pagePathForSource(resolvedTarget, hash)})`;
  });
}

function normalizeCodeFences(markdown) {
  return markdown
    .replace(/^```env\s*$/gim, '```ini')
    .replace(/^```caddyfile\s*$/gim, '```txt');
}

function extraFrontmatter(sourceRelativePath) {
  if (sourceRelativePath !== 'index.md') {
    return [];
  }

  return [
    'template: splash',
    'hero:',
    '  tagline: "Telegram-бот и Mini App для продажи подписок Remnawave: платежи, тарифы, админка, поддержка и инструкции подключения."',
    '  image:',
    '    html: \'<img class="minishop-hero-screenshot" src="/remnawave-minishop.webp" alt="Интерфейс Remnawave Minishop" width="1920" height="1080" loading="eager" decoding="async" />\'',
    '  actions:',
    '    - text: "Демо"',
    '      link: /demo/home',
    '      icon: right-arrow',
    '    - text: "Установка"',
    '      link: /getting-started/setup/',
    '      icon: setting',
    '      variant: minimal',
  ];
}

function frontmatter({ title, description, sourceRelativePath }) {
  const editPath = sourceRelativePath
    .split('/')
    .map((segment) => encodeURIComponent(segment))
    .join('/');
  const editUrl = `https://github.com/remna-user-panel/remna-user-panel/edit/main/docs/${editPath}`;
  return [
    '---',
    `title: ${yamlString(title)}`,
    `description: ${yamlString(description)}`,
    `editUrl: ${yamlString(editUrl)}`,
    ...extraFrontmatter(sourceRelativePath),
    '---',
    '',
  ].join('\n');
}

async function walk(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const absolutePath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(absolutePath)));
      continue;
    }
    if (entry.isFile()) {
      files.push(absolutePath);
    }
  }
  return files;
}

async function syncMarkdown(files) {
  for (const sourcePath of files.filter((file) => file.endsWith('.md'))) {
    const sourceRelativePath = toPosix(path.relative(sourceDir, sourcePath));
    const outputRelative = outputRelativePath(sourceRelativePath);
    const outputPath = path.join(outputDir, ...outputRelative.split('/'));
    const content = await readFile(sourcePath, 'utf8');
    const title = extractTitle(sourceRelativePath, content);
    const body = normalizeCodeFences(
      rewriteMarkdownLinks(stripFirstHeading(content).trimStart(), sourceRelativePath),
    );
    const output = frontmatter({
      title,
      description: descriptions[sourceRelativePath] ?? title,
      sourceRelativePath,
    });

    await mkdir(path.dirname(outputPath), { recursive: true });
    await writeFile(outputPath, `${output}${body}\n`, 'utf8');
  }
}

async function syncAssets(files) {
  for (const sourcePath of files.filter((file) => imageExtensions.has(path.extname(file).toLowerCase()))) {
    const sourceRelativePath = toPosix(path.relative(sourceDir, sourcePath));
    const outputRelative = !sourceRelativePath.includes('/')
      ? sourceRelativePath
      : sourceRelativePath;
    const outputPath = path.join(outputDir, ...outputRelative.split('/'));
    await mkdir(path.dirname(outputPath), { recursive: true });
    await copyFile(sourcePath, outputPath);

    if (!sourceRelativePath.includes('/')) {
      const referenceOutputPath = path.join(outputDir, 'reference', sourceRelativePath);
      await mkdir(path.dirname(referenceOutputPath), { recursive: true });
      await copyFile(sourcePath, referenceOutputPath);
    }
  }
}

await rm(outputDir, { recursive: true, force: true });
await mkdir(outputDir, { recursive: true });

const files = await walk(sourceDir);
await syncMarkdown(files);
await syncAssets(files);

console.log(`Synced documentation from ${path.relative(repoRoot, sourceDir)} to ${path.relative(repoRoot, outputDir)}`);
