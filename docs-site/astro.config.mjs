import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeNova from 'starlight-theme-nova';

export default defineConfig({
  site: 'https://docs.remna-user-panel.local',
  integrations: [
    starlight({
      title: 'Remna User Panel',
      favicon: '/favicon.png',
      description: 'Remna User Panel 中文部署、配置和维护文档。',
      plugins: [
        starlightThemeNova({
          nav: [
            { label: '部署', href: '/getting-started/deployment/' },
            { label: 'GitHub', href: 'https://github.com/remna-user-panel/remna-user-panel' },
          ],
        }),
      ],
      customCss: ['./src/styles/custom.css'],
      components: {
        Header: './src/components/Header.astro',
      },
      lastUpdated: false,
      locales: {
        root: {
          label: '中文',
          lang: 'zh',
        },
      },
      sidebar: [
        {
          label: '开始',
          items: [
            { label: '首页', slug: 'index' },
            { label: 'Docker 部署', slug: 'getting-started/deployment' },
          ],
        },
        {
          label: '配置',
          items: [
            { label: '后台配置', slug: 'configuration/settings' },
            { label: '安全', slug: 'configuration/security' },
          ],
        },
        {
          label: '功能',
          items: [
            { label: '支付', slug: 'features/payments' },
            { label: '套餐', slug: 'features/tariffs' },
          ],
        },
        {
          label: '维护',
          items: [{ label: '维护、更新与备份', slug: 'troubleshooting/maintenance' }],
        },
      ],
    }),
  ],
});
