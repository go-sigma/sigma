<script setup lang="ts">
import { computed } from 'vue'
import { useData } from 'vitepress'

const { lang } = useData()
const isZh = computed(() => String(lang.value).toLowerCase().startsWith('zh'))

const copy = computed(() =>
  isZh.value
    ? {
        title: '架构',
        lead: '四个进程，一个二进制。既可以在单机上一起运行，也可以把 server、distribution、worker 和 builder 拆分到集群中。',
        caption: '客户端请求经过各进程与后端服务的流转图。',
        alt: 'sigma 架构图',
      }
    : {
        title: 'Architecture',
        lead: 'Four processes, one binary. Run them together on a laptop, or split the server, distribution, worker, and builder across a cluster.',
        caption: 'Request flow from clients through the registry processes and their backing services.',
        alt: 'sigma architecture diagram',
      },
)
</script>

<template>
  <section class="hs hs-arch">
    <div class="hs__inner">
      <div class="hs__head">
        <h2 class="hs__title">{{ copy.title }}</h2>
        <p class="hs__lead">{{ copy.lead }}</p>
      </div>

      <figure class="hs-arch__frame">
        <img
          class="hs-arch__img"
          src="/img/architecture.svg"
          :alt="copy.alt"
          width="1840"
          height="1016"
          loading="lazy"
          decoding="async"
        />
        <figcaption class="hs-arch__caption">{{ copy.caption }}</figcaption>
      </figure>
    </div>
  </section>
</template>
