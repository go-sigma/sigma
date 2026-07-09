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
    ['link', { rel: 'icon', href: '/img/favicon.svg' }]
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
