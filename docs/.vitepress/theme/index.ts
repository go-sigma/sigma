import { h } from 'vue'
import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'

import HeroInstall from './components/HeroInstall.vue'
import HomeArchitecture from './components/HomeArchitecture.vue'
import HomeCta from './components/HomeCta.vue'
import HomeQuickStart from './components/HomeQuickStart.vue'

import './custom.css'

export default {
  extends: DefaultTheme,
  Layout() {
    return h(DefaultTheme.Layout, null, {
      'home-hero-actions-after': () => h(HeroInstall),
    })
  },
  enhanceApp({ app }) {
    app.component('HeroInstall', HeroInstall)
    app.component('HomeQuickStart', HomeQuickStart)
    app.component('HomeArchitecture', HomeArchitecture)
    app.component('HomeCta', HomeCta)
  },
} satisfies Theme
