<script setup lang="ts">
// Lời đọc phóng to trong khung màn nghe (wireframe D13): mỗi câu một khối kiểu lời bài hát,
// câu đang đọc đậm, câu khác nhạt, mép trên / dưới mờ dần. Tự cuộn giữ câu đang đọc ở ~1/3
// trên; người dùng tự cuộn (chuột / vuốt / phím) → ngừng tự cuộn, hiện "Về câu đang đọc".
// Đổi cỡ cửa sổ không tính là tự cuộn: canh lại câu đang đọc. Bấm câu → nghe từ câu đó.
// D18: ngay dưới câu đang đọc có "Chia sẻ câu này" · "Video 15 giây" — hiện khi di chuột vào
// lời đọc hoặc lúc tạm dừng (chỗ luôn dành sẵn nên lời không nhảy khi nút hiện).
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Clapperboard, Crosshair, Sparkles } from 'lucide-vue-next'
import { lyricIndex, lyrics, player, seekTime, tracks } from '../lib/player'
import { openShare } from '../lib/share'
import { openShortVideo } from '../lib/video'

const box = ref<HTMLElement | null>(null)
const follow = ref(true)

function scrollToCurrent(smooth = true) {
  void nextTick(() => {
    const el = box.value?.querySelector<HTMLElement>(`[data-i="${lyricIndex.value}"]`)
    if (!el || !box.value) return
    box.value.scrollTo({ top: el.offsetTop - box.value.clientHeight / 3, behavior: smooth ? 'smooth' : 'auto' })
  })
}
function userScrolled() {
  follow.value = false
}
const SCROLL_KEYS = ['ArrowUp', 'ArrowDown', 'PageUp', 'PageDown', 'Home', 'End']
function keyScrolled(e: KeyboardEvent) {
  if (SCROLL_KEYS.includes(e.key)) userScrolled()
}
function backToCurrent() {
  follow.value = true
  scrollToCurrent()
}

let ro: ResizeObserver | null = null
onMounted(() => {
  scrollToCurrent(false)
  ro = new ResizeObserver(() => follow.value && scrollToCurrent(false))
  if (box.value) ro.observe(box.value)
})
onBeforeUnmount(() => ro?.disconnect())
watch(lyricIndex, () => follow.value && scrollToCurrent())
watch(
  () => player.current,
  () => {
    follow.value = true // sang tiểu mục mới: bám theo lại từ đầu
    scrollToCurrent(false)
  },
)

function pickSentence(start: number) {
  follow.value = true
  seekTime(start + 0.01)
}

const next = computed(() => tracks.value[player.current + 1])
</script>

<template>
  <div class="group/stage relative min-h-0 flex flex-col">
    <div ref="box" class="flex-1 min-h-0 overflow-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden [mask-image:linear-gradient(to_bottom,transparent,black_14%,black_80%,transparent)]"
      @wheel.passive="userScrolled" @touchmove.passive="userScrolled" @keydown="keyScrolled">
      <div v-if="lyrics" class="py-[16%] space-y-2.5">
        <template v-for="(p, pi) in lyrics.paragraphs" :key="pi">
          <template v-for="(s, si) in p" :key="s.index">
          <p :data-i="s.index" role="button" tabindex="0" title="Nghe từ câu này"
            class="text-[23px] leading-[1.35] font-bold tracking-[-0.01em] cursor-pointer transition-colors duration-500 outline-none focus-visible:underline"
            :class="[
              s.index === lyricIndex ? 'text-foreground' : s.index < lyricIndex ? 'text-foreground/20 hover:text-foreground/50' : 'text-foreground/35 hover:text-foreground/60',
              si === p.length - 1 && pi < lyrics.paragraphs.length - 1 && 'pb-4',
            ]"
            @click="pickSentence(s.start)" @keydown.enter="pickSentence(s.start)">{{ s.text }}</p>
            <div v-if="s.index === lyricIndex" class="-mt-1 h-9 flex items-center gap-2 transition-opacity duration-300 focus-within:opacity-100"
              :class="player.playing ? 'opacity-0 group-hover/stage:opacity-100' : 'opacity-100'">
              <button class="h-8 pl-2.5 pr-3 rounded-full text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-rose-500/30 bg-gradient-to-r from-orange-500 via-rose-600 to-purple-700 hover:brightness-110"
                title="Ảnh có lời từ câu này, đăng Facebook, Zalo, Story" @click="openShare(s.index)"><Sparkles class="w-3.5 h-3.5" /> Chia sẻ câu này</button>
              <button class="h-8 px-3 rounded-full border border-rose-500/40 text-rose-600 dark:text-rose-400 bg-background/70 text-xs font-semibold flex items-center gap-1.5 hover:bg-rose-500/10"
                title="Video ngắn có tiếng đọc từ câu này, đăng Reels, TikTok" @click="openShortVideo(s.index)"><Clapperboard class="w-3.5 h-3.5" /> Video 15 giây</button>
            </div>
          </template>
        </template>
        <p class="pt-4 text-sm text-muted-foreground">Hết tiểu mục<template v-if="next"> · tiếp theo: {{ next.title }}</template></p>
      </div>
      <p v-else class="h-full grid place-items-center text-sm text-muted-foreground">Tiểu mục này không có chữ để hiện.</p>
    </div>
    <button v-if="!follow && lyrics" class="absolute bottom-2 right-0 h-8 px-3 rounded-full bg-primary text-primary-foreground text-xs shadow-lg flex items-center gap-1.5" @click="backToCurrent">
      <Crosshair class="w-3.5 h-3.5" /> Về câu đang đọc
    </button>
  </div>
</template>
