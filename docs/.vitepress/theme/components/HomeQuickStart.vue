<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = computed(() => String(lang.value).toLowerCase().startsWith('zh'))

const startCommand =
  'docker run -d --name sigma -p 3000:3000 ghcr.io/go-sigma/sigma:nightly-alpine'

const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

async function copy() {
  try {
    await navigator.clipboard.writeText(startCommand)
    copied.value = true
    clearTimeout(timer)
    timer = setTimeout(() => {
      copied.value = false
    }, 1600)
  } catch {
    copied.value = false
  }
}

onBeforeUnmount(() => clearTimeout(timer))

const content = computed(() =>
  isZh.value
    ? {
        title: '一条命令启动',
        lead: '拉取镜像即可运行。server、distribution、worker 和 builder 都在同一个容器中，默认使用内置数据库与本地存储。',
        link: '完整快速开始',
        linkHref: '/zh/quickstart',
      }
    : {
        title: 'Start in one command',
        lead: 'Pull the image and run. The server, distribution, worker, and builder all ship in one container, with an embedded database and local storage by default.',
        link: 'Full quickstart',
        linkHref: '/quickstart',
      },
)
</script>

<template>
  <section class="hs hs-quick">
    <div class="hs__inner">
      <div class="hs-quick__grid">
        <div class="hs-quick__intro">
          <h2 class="hs__title">{{ content.title }}</h2>
          <p class="hs__lead">{{ content.lead }}</p>
          <a class="hs-quick__link" :href="content.linkHref">
            {{ content.link }}
            <span aria-hidden="true">&rarr;</span>
          </a>
        </div>

        <figure class="hs-quick__terminal">
          <div class="hs-quick__term-head">
            <span class="hs-quick__term-name">bash</span>
            <button
              class="hs-quick__copy"
              type="button"
              :aria-label="isZh ? '复制启动命令' : 'Copy start command'"
              @click="copy"
            >
              {{ copied ? (isZh ? '已复制' : 'Copied') : isZh ? '复制' : 'Copy' }}
            </button>
          </div>
          <pre class="hs-quick__term-body"><code><span class="c-prompt">$</span> docker run -d --name sigma -p 3000:3000 \
      ghcr.io/go-sigma/sigma:nightly-alpine</code></pre>
        </figure>
      </div>
    </div>
  </section>
</template>
