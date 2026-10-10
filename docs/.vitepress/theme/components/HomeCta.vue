<script setup lang="ts">
import { computed } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = computed(() => String(lang.value).toLowerCase().startsWith('zh'))

const copy = computed(() =>
  isZh.value
    ? {
        title: '运行你自己的仓库',
        lead: '一分钟内部署 sigma，或阅读文档规划生产环境。',
        primary: '快速开始',
        primaryHref: '/zh/quickstart',
        secondary: 'Star on GitHub',
        secondaryHref: 'https://github.com/go-sigma/sigma',
      }
    : {
        title: 'Run your own registry',
        lead: 'Deploy sigma in a minute, or read the docs to plan a production setup.',
        primary: 'Get started',
        primaryHref: '/quickstart',
        secondary: 'Star on GitHub',
        secondaryHref: 'https://github.com/go-sigma/sigma',
      },
)
</script>

<template>
  <section class="hs hs-cta">
    <div class="hs__inner">
      <div class="hs-cta__panel">
        <h2 class="hs-cta__title">{{ copy.title }}</h2>
        <p class="hs-cta__lead">{{ copy.lead }}</p>
        <div class="hs-cta__actions">
          <a class="hs-btn hs-btn--brand" :href="copy.primaryHref">{{ copy.primary }}</a>
          <a
            class="hs-btn hs-btn--ghost"
            :href="copy.secondaryHref"
            target="_blank"
            rel="noreferrer"
          >
            {{ copy.secondary }}
          </a>
        </div>
      </div>
    </div>
  </section>
</template>
