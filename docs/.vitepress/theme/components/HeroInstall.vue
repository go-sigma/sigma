<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = computed(() => String(lang.value).toLowerCase().startsWith('zh'))

const command =
  'docker run -d --name sigma -p 3000:3000 ghcr.io/go-sigma/sigma:nightly-alpine'

const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

async function copy() {
  try {
    await navigator.clipboard.writeText(command)
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
</script>

<template>
  <div class="hero-install">
    <span class="hero-install__label">
      {{ isZh ? '一条命令启动' : 'Run in one command' }}
    </span>
    <code class="hero-install__code">
      <span class="hero-install__prompt">$</span>{{ command }}
    </code>
    <button
      class="hero-install__copy"
      type="button"
      :aria-label="isZh ? '复制命令' : 'Copy command'"
      @click="copy"
    >
      {{ copied ? (isZh ? '已复制' : 'Copied') : isZh ? '复制' : 'Copy' }}
    </button>
  </div>
</template>
