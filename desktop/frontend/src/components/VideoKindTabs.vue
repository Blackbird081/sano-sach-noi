<script setup lang="ts">
// Đầu hộp Tạo video (wireframe D17): chọn Video ngắn hoặc Video cả cuốn.
import { Check, Film, MonitorPlay } from 'lucide-vue-next'
import { canShortVideo, switchVideo, type VideoKind } from '../lib/video'

const props = defineProps<{ current: VideoKind; slug: string; disabled?: boolean }>()
const kinds = [
  { k: 'short' as const, t: 'Video ngắn', d: 'Reels, TikTok, Shorts · 15–30 giây', icon: Film },
  { k: 'full' as const, t: 'Video cả cuốn', d: 'YouTube, Facebook · cả cuốn hoặc vài chương', icon: MonitorPlay },
]
const off = (k: VideoKind) => props.disabled || (k === 'short' && !canShortVideo())
</script>

<template>
  <div class="grid grid-cols-2 gap-2" role="tablist" aria-label="Loại video">
    <button v-for="x in kinds" :key="x.k" role="tab" :aria-selected="current === x.k" :disabled="off(x.k) && current !== x.k"
      class="rounded-xl border-2 px-3 py-2.5 text-left flex gap-2.5 items-start disabled:opacity-50 disabled:cursor-not-allowed"
      :class="current === x.k ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'"
      :title="x.k === 'short' && !canShortVideo() ? 'Mục đang nghe chưa có lời đọc theo câu' : undefined"
      @click="current !== x.k && switchVideo(x.k, slug)">
      <span class="mt-0.5 border-2 rounded-sm shrink-0" :class="[x.k === 'short' ? 'w-3.5 h-6' : 'w-6 h-3.5 mt-1.5', current === x.k ? 'border-primary' : 'border-muted-foreground/50']"></span>
      <span class="min-w-0 flex-1">
        <span class="font-semibold text-sm flex items-center gap-1.5"><component :is="x.icon" class="w-4 h-4" /> {{ x.t }}</span>
        <span class="block text-[11px] text-muted-foreground mt-0.5 leading-snug">{{ x.d }}</span>
      </span>
      <Check v-if="current === x.k" class="w-4 h-4 text-primary shrink-0" />
    </button>
  </div>
</template>
