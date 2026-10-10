import { defineConfig } from 'vitepress'

const guideSidebar = [
  {
    text: 'Getting Started',
    items: [
      { text: 'Quick Start', link: '/quickstart' },
      { text: 'Configuration', link: '/configuration' },
      { text: 'MCP Server', link: '/mcp' },
      { text: 'Cosign', link: '/cosign' }
    ]
  },
  {
    text: 'Push to sigma',
    items: [
      { text: 'Docker', link: '/push/docker' },
      { text: 'Helm', link: '/push/helm' },
      { text: 'Apptainer', link: '/push/apptainer' }
    ]
  }
]

const zhGuideSidebar = [
  {
    text: '开始使用',
    items: [
      { text: '快速开始', link: '/zh/quickstart' },
      { text: '配置', link: '/zh/configuration' },
      { text: 'MCP Server', link: '/zh/mcp' },
      { text: 'Cosign', link: '/zh/cosign' }
    ]
  },
  {
    text: '推送到 sigma',
    items: [
      { text: 'Docker', link: '/zh/push/docker' },
      { text: 'Helm', link: '/zh/push/helm' },
      { text: 'Apptainer', link: '/zh/push/apptainer' }
    ]
  }
]

export default defineConfig({
  title: 'sigma',
  description: 'A lightweight OCI artifact storage and distribution system',
  cleanUrls: true,
  head: [
    ['link', { rel: 'preconnect', href: 'https://cdn.jsdelivr.net', crossorigin: '' }],
    [
      'link',
      { rel: 'stylesheet', href: 'https://cdn.jsdelivr.net/npm/@fontsource/nunito@5/latin-400.css' }
    ],
    [
      'link',
      { rel: 'stylesheet', href: 'https://cdn.jsdelivr.net/npm/@fontsource/nunito@5/latin-500.css' }
    ],
    [
      'link',
      { rel: 'stylesheet', href: 'https://cdn.jsdelivr.net/npm/@fontsource/nunito@5/latin-600.css' }
    ],
    [
      'link',
      { rel: 'stylesheet', href: 'https://cdn.jsdelivr.net/npm/@fontsource/nunito@5/latin-700.css' }
    ],
    ['link', { rel: 'icon', href: '/img/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#fbfaf9', media: '(prefers-color-scheme: light)' }],
    ['meta', { name: 'theme-color', content: '#1b1a18', media: '(prefers-color-scheme: dark)' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'sigma' }],
    [
      'meta',
      {
        property: 'og:description',
        content: 'A lightweight, self-hosted OCI artifact storage and distribution system.'
      }
    ],
    ['meta', { property: 'og:image', content: '/img/logo.svg' }]
  ],
  themeConfig: {
    logo: '/img/logo.svg',
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Docs', link: '/quickstart' },
      { text: 'MCP', link: '/mcp' }
    ],
    sidebar: guideSidebar,
    socialLinks: [
      { icon: 'github', link: 'https://github.com/go-sigma/sigma' }
    ],
    search: {
      provider: 'local'
    }
  },
  locales: {
    root: {
      label: 'English',
      lang: 'en-US'
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      title: 'sigma',
      description: '轻量级 OCI 制品存储与分发系统',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '文档', link: '/zh/quickstart' },
          { text: 'MCP', link: '/zh/mcp' }
        ],
        sidebar: zhGuideSidebar,
        outline: {
          label: '页面导航'
        },
        docFooter: {
          prev: '上一页',
          next: '下一页'
        },
        darkModeSwitchLabel: '外观',
        sidebarMenuLabel: '菜单',
        returnToTopLabel: '返回顶部',
        langMenuLabel: '切换语言'
      }
    }
  }
})
