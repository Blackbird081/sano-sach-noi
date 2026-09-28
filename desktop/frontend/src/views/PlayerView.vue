<script setup lang="ts">
// Màn nghe: mục lục tiểu mục, tua −15s/+30s, đổi tốc độ, nhớ vị trí nghe; xuất M4B
// (tiến độ ngay dưới hàng nút), xuất gói zip, xoá vào Thùng rác. Việc phát nằm ở
// lib/player.ts (dùng chung với thanh nghe nhỏ) nên rời màn này vẫn nghe tiếp.
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Check, ChevronDown, ChevronLeft, Download, Expand, FolderOpen, Gauge, Loader2, Maximize2, Mic, Minimize2, Package, Pause, Play, RotateCcw, RotateCw,
  Pencil, Settings, SkipBack, SkipForward, Smartphone, Timer, Trash2, Volume2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import M4BProgress from '../components/M4BProgress.vue'
import LyricsPanel from '../components/LyricsPanel.vue'
import LyricsStage from '../components/LyricsStage.vue'
import PauseCustomInputs from '../components/PauseCustomInputs.vue'
import { PAUSE_PRESETS, bookPause, fmtGap, levelLabel, makeSetting, pauseState, setBookPause, type Gaps, type PauseLevel } from '../lib/pause'
import { deleteBook, errText, openBookFolder, revealBookZip } from '../lib/backend'
import { fmtClock, fmtLong } from '../lib/position'
import { go, state } from '../lib/store'
import { openEdit } from '../lib/edit'
import { m4bBusy, startM4B, useM4B } from '../lib/m4b'
import { openPhone } from '../lib/phone'
import {
  SPEEDS as speeds, bookHasLyrics, forgetBook, lyricIndex, lyrics, pause, pctTrack, pick, player, seek, seekFrac, setSpeed as applySpeed, skip, toggle, totalSec,
  track, tracks,
} from '../lib/player'

useM4B()

const actionError = ref('')
const lyricsOpen = ref(false)
watch(() => state.playerSlug, () => (lyricsOpen.value = false))
const tocEl = ref<HTMLElement | null>(null)

// Mục lục tự cuộn tới tiểu mục đang phát (mở màn nghe, chuyển tiểu mục, sang cuốn khác).
function scrollToCurrent(smooth: boolean) {
  void nextTick(() => {
    const el = tocEl.value?.querySelector<HTMLElement>('[data-current="true"]')
    el?.scrollIntoView({ block: 'center', behavior: smooth ? 'smooth' : 'auto' })
  })
}
onMounted(() => scrollToCurrent(false))
watch(() => player.current, () => scrollToCurrent(true))
watch(() => player.detail?.slug, () => scrollToCurrent(false))
watch(lyricsOpen, (open) => !open && scrollToCurrent(false))

// Lời đọc phóng to trong khung (wireframe D13): nhớ lựa chọn cho lần mở sau.
const FRAME_KEY = 'sano.lyricsFrame'
function loadFrame() {
  try {
    return localStorage.getItem(FRAME_KEY) === '1'
  } catch {
    return false
  }
}
const frameOn = ref(loadFrame())
function setFrame(on: boolean) {
  frameOn.value = on
  try {
    localStorage.setItem(FRAME_KEY, on ? '1' : '0')
  } catch {
    // không lưu được thì thôi
  }
}
const framed = computed(() => frameOn.value && bookHasLyrics.value)

// Tiến độ cả cuốn ở đầu mục lục: số tiểu mục đã nghe, thời gian còn lại.
const remainSec = computed(() => {
  let n = 0
  tracks.value.forEach((t, i) => {
    if (player.heard.includes(i)) return
    n += i === player.current ? Math.max(0, t.durationSec - player.time) : t.durationSec
  })
  return n
})
const pctBook = computed(() => (totalSec.value ? Math.min(100, ((totalSec.value - remainSec.value) / totalSec.value) * 100) : 0))
const speedOpen = ref(false)
// Menu bánh răng góc trên phải: các việc phụ của cuốn sách (zip, thư mục, lưu M4B chỗ khác, xoá).
const bookMenu = ref(false)
function menuAct(fn: () => unknown) {
  bookMenu.value = false
  void fn()
}

// Cách xếp cột giữa theo chỗ trống thật (wireframe D11). Ngưỡng tính theo nội dung
// cao nhất của từng cách xếp; sách không có chữ chạy thì cần ít chỗ hơn.
const midEl = ref<HTMLElement | null>(null)
const midH = ref(9999)
const midW = ref(9999)
let midRO: ResizeObserver | null = null
watch(midEl, (el, old) => {
  if (old) midRO?.unobserve(old)
  if (!el) return
  midRO ??= new ResizeObserver(([e]) => {
    midH.value = e.contentRect.height
    midW.value = e.contentRect.width
  })
  midRO.observe(el)
})
onBeforeUnmount(() => midRO?.disconnect())
// Bìa to nhất vừa chỗ trống. Đo thật phần còn lại của cách xếp Rộng: khối tên sách
// (titleH) và phần dưới (lời đọc, thanh thời gian, hàng nút). Rộng: bìa giữa, trên
// khối tên. Vừa: bìa bên trái khối tên nên được cao hơn. Chọn cách cho bìa to hơn;
// bìa Rộng từ WIDE_MIN px là giữ Rộng (bìa giữa đẹp hơn khi đủ lớn).
const WIDE_MIN = 180
const SIDE_MIN = 96
const restH = ref(0) // phần còn lại của cách xếp Rộng (không tính bìa), đo thật
const titleH = ref(0) // khối tên sách, giọng, tiểu mục (kể cả khoảng cách trên)
const coverMax = computed(() => (bookHasLyrics.value ? 240 : 300))
const rest = computed(() => restH.value || (bookHasLyrics.value ? 310 : 210))
const coverH = computed(() => Math.floor(Math.min(coverMax.value, midH.value - rest.value - 16)))
const sideCoverH = computed(() => Math.floor(Math.min(220, midH.value - (rest.value - (titleH.value || 84)) - 16)))
const fit = computed<'wide' | 'medium' | 'narrow'>(() => {
  if (coverH.value >= WIDE_MIN || (coverH.value >= SIDE_MIN && coverH.value >= sideCoverH.value)) return 'wide'
  return sideCoverH.value >= SIDE_MIN ? 'medium' : 'narrow'
})
// Đo lại phần còn lại mỗi khi đang ở cách xếp Rộng (đổi sách, có / không lời đọc, lỗi).
function measureRest() {
  const el = midEl.value
  if (!el || framed.value || fit.value !== 'wide') return
  let h = 0 // mọi thứ trừ bìa
  let t = 0 // khối tên sách: các dòng H2/P đứng trước khung Lời đọc
  let beforeLyrics = true
  for (const c of Array.from(el.children) as HTMLElement[]) {
    const cs = getComputedStyle(c)
    if (c.dataset.cover !== undefined || cs.position === 'absolute' || cs.display === 'none') continue
    const ch = c.offsetHeight + parseFloat(cs.marginTop) + parseFloat(cs.marginBottom)
    h += ch
    if (c.tagName === 'BUTTON' || c.tagName === 'DIV') beforeLyrics = false
    else if (beforeLyrics) t += ch
  }
  if (Math.abs(h - restH.value) > 1) restH.value = h
  if (Math.abs(t - titleH.value) > 1) titleH.value = t
}
watch([midH, fit, framed, bookHasLyrics, () => player.detail?.slug, () => player.error], () => void nextTick(measureRest))
// Chế độ phóng to: bìa hàng trên ~22% chiều cao cột giữa (80–150 px); khung hẹp thu gọn nút.
const frameCoverH = computed(() => Math.round(Math.max(80, Math.min(150, midH.value * 0.22))))
const compact = computed(() => midW.value < 520)

// Quãng nghỉ riêng cuốn đang nghe (wireframe D9); null = theo cài đặt chung.
const pauseOpen = ref(false)
const pauseMsg = ref('')
const mine = computed(() => (state.playerSlug ? bookPause(state.playerSlug) : null))
const custom = ref(false) // đang mở ô Tuỳ chỉnh
const pauseLabel = computed(() => {
  const m = mine.value
  if (!m) return 'Nghỉ: theo chung'
  return m.level === 'custom' ? `Nghỉ: ${fmtGap(m.section)}` : `Nghỉ: ${levelLabel(m.level)}`
})
const customValue = computed<Gaps>(() => mine.value ?? pauseState.global)
watch(() => state.playerSlug, () => {
  pauseOpen.value = false
  pauseMsg.value = ''
})
watch(pauseOpen, (open) => (custom.value = open && mine.value?.level === 'custom'))

let pauseMsgTimer = 0
function savePause(v: ReturnType<typeof makeSetting> | null) {
  setBookPause(state.playerSlug, v)
  clearTimeout(pauseMsgTimer)
  pauseMsgTimer = window.setTimeout(() => (pauseMsg.value = ''), 4000)
  pauseMsg.value = v
    ? `Cuốn này nghỉ ${fmtGap(v.section)} giữa tiểu mục, ${fmtGap(v.chapter)} sang chương`
    : 'Cuốn này theo cài đặt chung'
}
function pickPause(level: PauseLevel | 'global') {
  if (level === 'custom') {
    custom.value = true
    savePause(makeSetting('custom', customValue.value))
    return
  }
  savePause(level === 'global' ? null : makeSetting(level))
  pauseOpen.value = false
}

function seekTo(e: MouseEvent) {
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  seekFrac((e.clientX - r.left) / r.width)
}

function setSpeed(v: number) {
  applySpeed(v)
  speedOpen.value = false
}

// Đóng menu tốc độ / quãng nghỉ khi bấm ra ngoài hoặc nhấn Esc.
function closeSpeed(e: Event) {
  const esc = e instanceof KeyboardEvent
  if (esc ? e.key === 'Escape' : !(e.target as HTMLElement).closest('[data-speed-menu]')) speedOpen.value = false
  if (esc ? e.key === 'Escape' : !(e.target as HTMLElement).closest('[data-pause-menu]')) pauseOpen.value = false
  if (esc ? e.key === 'Escape' : !(e.target as HTMLElement).closest('[data-book-menu]')) bookMenu.value = false
}
document.addEventListener('mousedown', closeSpeed)
document.addEventListener('keydown', closeSpeed)
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', closeSpeed)
  document.removeEventListener('keydown', closeSpeed)
})

const fmtSpeed = (s: number) => s.toLocaleString('vi-VN') + '×'

async function removeBook() {
  actionError.value = ''
  try {
    pause()
    const slug = state.playerSlug
    const where = await deleteBook(slug)
    if (!where) return // người dùng huỷ
    forgetBook(slug)
    go('library')
  } catch (e) {
    actionError.value = errText(e)
  }
}

async function act(fn: (slug: string) => Promise<void>) {
  try {
    await fn(state.playerSlug)
  } catch (e) {
    actionError.value = errText(e)
  }
}
</script>

<template>
  <LyricsPanel v-if="lyricsOpen && player.detail" @close="lyricsOpen = false" />
  <section v-else class="flex-1 flex min-h-0">
    <div class="relative isolate flex-1 flex flex-col p-6 min-w-0 min-h-0">
      <!-- Phóng to (D13): lớp màu bìa thật nhẹ phủ cả cột giữa (bìa phóng to, làm mờ), tan dần về nền app -->
      <div v-if="framed && player.detail" aria-hidden="true" class="absolute inset-0 -z-10 overflow-hidden pointer-events-none">
        <div class="absolute -inset-10 blur-3xl saturate-150 opacity-[0.18] dark:opacity-[0.22]">
          <BookCover :cover-image="player.detail.thumbUrl || player.detail.coverUrl" :title="player.detail.title" class="h-full w-full rounded-none shadow-none" />
        </div>
        <div class="absolute inset-0 bg-gradient-to-b from-transparent via-background/40 to-background"></div>
      </div>
      <div class="relative shrink-0 flex items-center justify-between gap-2">
        <button class="text-sm text-muted-foreground flex items-center gap-1 hover:text-foreground w-fit" @click="go('library')"><ChevronLeft class="w-4 h-4" /> Thư viện</button>
        <!-- Phóng to (D13): thương hiệu giữa thanh trên như đầu trình nghe (ảnh chụp / chia sẻ có sẵn Sano) -->
        <span v-if="framed && player.detail" class="absolute left-1/2 -translate-x-1/2 flex items-center gap-2 pointer-events-none select-none">
          <img src="@/assets/favicon.svg" alt="" class="h-6 w-6 rounded-md shadow-sm" />
          <span class="text-[15px] font-semibold tracking-tight">Sano</span>
          <span class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">Sách nói</span>
        </span>
        <div v-if="player.detail" class="relative" data-book-menu>
          <button class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" :class="bookMenu && 'bg-muted text-foreground'"
            aria-label="Tuỳ chọn cuốn sách" aria-haspopup="menu" :aria-expanded="bookMenu" @click="bookMenu = !bookMenu"><Settings class="w-4 h-4" /></button>
          <div v-if="bookMenu" role="menu" class="absolute top-full right-0 mt-1 w-64 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-30 text-sm">
            <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" title="Sửa chữ, đọc lại mục đang nghe (wireframe D11)" @click="menuAct(() => openEdit(state.playerSlug, { index: player.current }))"><Pencil class="w-4 h-4" /> Sửa mục đang nghe</button>
            <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted disabled:opacity-50" :disabled="!player.detail.zip" title="Sao lưu hoặc chuyển sách sang máy khác" @click="menuAct(() => act(revealBookZip))"><Package class="w-4 h-4" /> Xuất gói zip</button>
            <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted disabled:opacity-50" :disabled="m4bBusy()" title="Tạo file M4B và tự chọn nơi lưu" @click="menuAct(() => startM4B(state.playerSlug, true))"><Download class="w-4 h-4" /> Lưu file M4B vào chỗ khác…</button>
            <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted" @click="menuAct(() => act(openBookFolder))"><FolderOpen class="w-4 h-4" /> Mở thư mục sách</button>
            <div class="my-1 border-t border-border"></div>
            <button role="menuitem" class="w-full flex items-center gap-2 px-3 py-1.5 text-left hover:bg-muted text-destructive" title="Chuyển cuốn sách vào Thùng rác (lấy lại được)" @click="menuAct(removeBook)"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</button>
          </div>
        </div>
      </div>
      <div v-if="!player.detail || player.slug !== state.playerSlug" class="flex-1 grid place-items-center text-sm text-muted-foreground">
        <span v-if="player.error" class="text-destructive">{{ player.error }}</span>
        <span v-else-if="!state.playerSlug">Chọn một cuốn trong thư viện để nghe.</span>
        <span v-else class="flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang mở sách…</span>
      </div>
      <!-- Cột giữa (wireframe D11): đo chỗ trống thật, chọn cách xếp Rộng / Vừa / Hẹp để luôn vừa,
           không cuộn, không cắt. Menu tốc độ / Nghỉ được tràn ra ngoài cột (không overflow). -->
      <div v-else ref="midEl" data-fit-check="mid" class="relative flex-1 min-h-0 flex flex-col items-center" :class="framed ? 'pt-4' : 'justify-center py-2'">
        <!-- Phóng to (D13): hàng trên bìa nhỏ + tên sách; lời đọc chiếm phần giữa; hàng nút giữ nguyên -->
        <template v-if="framed">
          <div class="shrink-0 w-full max-w-2xl flex items-center gap-4">
            <div class="rounded-lg shadow-lg overflow-hidden shrink-0" :style="{ height: frameCoverH + 'px', width: Math.round(frameCoverH * 0.75) + 'px' }">
              <BookCover :cover-image="player.detail.coverUrl" :title="player.detail.title" :author="player.detail.author" class="h-full w-full rounded-lg shadow-none" />
            </div>
            <div class="min-w-0 flex-1">
              <h2 class="text-lg font-semibold leading-snug line-clamp-2" :title="player.detail.title">{{ player.detail.title }}</h2>
              <p v-if="player.detail.author || player.detail.voice" class="mt-0.5 text-sm text-muted-foreground flex items-center gap-1 min-w-0">
                <span v-if="player.detail.author" class="truncate">{{ player.detail.author }}</span><template v-if="player.detail.author && player.detail.voice"> ·</template>
                <span v-if="player.detail.voice" class="inline-flex items-center gap-1 min-w-0"><Mic class="w-3.5 h-3.5 shrink-0" /><span class="truncate">Giọng {{ player.detail.voice }}</span></span>
              </p>
              <p class="text-sm text-primary font-medium truncate" :title="track?.title">{{ track?.title }}</p>
            </div>
            <div class="self-start shrink-0 flex items-center gap-1 text-xs text-muted-foreground">
              <button class="h-8 rounded-md border border-border bg-background/70 hover:bg-muted hover:text-foreground flex items-center justify-center gap-1.5" :class="compact ? 'w-8' : 'px-2.5'"
                title="Thu lại màn nghe thường" aria-label="Thu lại" @click="setFrame(false)"><Minimize2 class="w-3.5 h-3.5" /><template v-if="!compact">Thu lại</template></button>
              <button class="h-8 w-8 rounded-md border border-border bg-background/70 hover:bg-muted hover:text-foreground grid place-items-center disabled:opacity-50"
                title="Xem cả lời" aria-label="Xem cả lời" :disabled="!lyrics" @click="lyricsOpen = true"><Expand class="w-3.5 h-3.5" /></button>
            </div>
          </div>
          <LyricsStage class="mt-3 flex-1 w-full max-w-2xl" />
        </template>
        <template v-else-if="fit === 'wide'">
          <div data-cover class="rounded-xl shadow-2xl overflow-hidden shrink-0" :style="{ height: coverH + 'px', width: Math.round(coverH * 0.75) + 'px' }">
            <BookCover :cover-image="player.detail.coverUrl" :title="player.detail.title" :author="player.detail.author" class="h-full w-full rounded-xl shadow-none" />
          </div>
          <h2 class="mt-4 text-lg font-semibold text-center max-w-md line-clamp-2">{{ player.detail.title }}</h2>
          <p v-if="player.detail.author || player.detail.voice" class="text-sm text-muted-foreground flex items-center gap-1">
            {{ player.detail.author }}<template v-if="player.detail.author && player.detail.voice"> ·</template>
            <span v-if="player.detail.voice" class="inline-flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng {{ player.detail.voice }}</span>
          </p>
          <p class="text-sm text-muted-foreground max-w-md truncate" :title="track?.title">{{ track?.title }}</p>
        </template>
        <div v-else class="w-full max-w-md flex items-center gap-4 shrink-0">
          <div v-if="fit === 'medium'" class="rounded-lg shadow-lg overflow-hidden shrink-0" :style="{ height: sideCoverH + 'px', width: Math.round(sideCoverH * 0.75) + 'px' }">
            <BookCover :cover-image="player.detail.coverUrl" :title="player.detail.title" :author="player.detail.author" class="h-full w-full rounded-xl shadow-none" />
          </div>
          <div class="min-w-0">
            <h2 class="text-lg font-semibold truncate" :title="player.detail.title">{{ player.detail.title }}</h2>
            <p v-if="player.detail.author || player.detail.voice" class="text-sm text-muted-foreground flex items-center gap-1 min-w-0">
              <span class="truncate">{{ player.detail.author }}</span><template v-if="player.detail.author && player.detail.voice"> ·</template>
              <span v-if="player.detail.voice" class="inline-flex items-center gap-1 shrink-0"><Mic class="w-3.5 h-3.5" /> Giọng {{ player.detail.voice }}</span>
            </p>
            <p class="text-sm text-muted-foreground truncate" :title="track?.title">{{ track?.title }}</p>
          </div>
        </div>
        <!-- Chiều cao cố định để không đẩy nút phía dưới; cách xếp Hẹp chỉ 1 dòng.
             Đầu khung: Phóng to (lời đọc to ngay trong khung, D13) · Xem cả lời (toàn vùng nội dung). -->
        <div v-if="bookHasLyrics && !framed" class="mt-4 w-full max-w-md shrink-0 rounded-lg border border-border bg-muted/30 px-4 py-3 hover:bg-muted/60">
          <span class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            Lời đọc
            <span class="flex items-center gap-3 normal-case font-normal tracking-normal text-primary">
              <button class="flex items-center gap-1 hover:underline" title="Phóng to lời đọc ngay trong khung" @click="setFrame(true)"><Maximize2 class="w-3 h-3" /> Phóng to</button>
              <button v-if="lyrics" class="hover:underline" @click="lyricsOpen = true">Xem cả lời →</button>
            </span>
          </span>
          <button class="lyrics-body block w-full text-left disabled:cursor-default" :disabled="!lyrics" @click="lyricsOpen = true">
            <template v-if="lyrics">
              <span class="mt-1 text-sm font-medium leading-snug" :class="fit === 'narrow' ? 'h-[1.375em] line-clamp-1' : 'h-[2.75em] line-clamp-2'">{{ lyrics.sentences[lyricIndex]?.text }}</span>
              <span v-if="fit !== 'narrow'" class="mt-0.5 block h-[1.375em] text-sm text-muted-foreground leading-snug line-clamp-1">{{ lyrics.sentences[lyricIndex + 1]?.text }}</span>
            </template>
            <template v-else>
              <span class="mt-1 block text-sm text-muted-foreground leading-snug" :class="fit === 'narrow' ? 'h-[1.375em] truncate' : 'h-[2.75em]'">Tiểu mục này không có chữ để hiện.</span>
              <span v-if="fit !== 'narrow'" class="mt-0.5 block h-[1.375em]"></span>
            </template>
          </button>
        </div>
        <div class="w-full max-w-md shrink-0" :class="framed ? 'mt-2' : 'mt-4'">
          <div class="h-1.5 rounded-full bg-muted overflow-hidden cursor-pointer" role="slider" aria-label="Vị trí nghe" :aria-valuenow="Math.round(pctTrack)" @click="seekTo">
            <div class="h-full bg-primary rounded-full" :style="{ width: pctTrack + '%' }"></div>
          </div>
          <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>{{ fmtClock(player.time) }}</span>
            <span>-{{ fmtClock(Math.max(0, player.duration - player.time)) }}</span></div>
        </div>
        <!-- Một hàng: tốc độ (trái) · nút phát (giữa) · Nghỉ (phải) -->
        <div class="mt-3 w-full max-w-md shrink-0 grid grid-cols-[1fr_auto_1fr] items-center gap-2 text-xs">
          <div class="justify-self-start">
          <div class="relative" data-speed-menu>
            <button class="h-8 px-2.5 rounded-full border border-border flex items-center gap-1 whitespace-nowrap hover:bg-muted" :class="framed && 'bg-background/70'" aria-haspopup="listbox" :aria-expanded="speedOpen" @click="speedOpen = !speedOpen">
              <Gauge class="w-3.5 h-3.5" />{{ fmtSpeed(player.speed) }}<ChevronDown class="w-3 h-3 text-muted-foreground" />
            </button>
            <div v-if="speedOpen" role="listbox" aria-label="Tốc độ phát" class="absolute bottom-full left-0 mb-2 w-40 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
              <div class="px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Tốc độ phát</div>
              <button v-for="v in speeds" :key="v" role="option" :aria-selected="v === player.speed"
                class="w-full flex items-center justify-between px-3 py-1.5 text-sm text-left hover:bg-muted"
                :class="v === player.speed && 'text-primary font-medium'" @click="setSpeed(v)">
                {{ fmtSpeed(v) }}<Check v-if="v === player.speed" class="w-4 h-4" />
              </button>
            </div>
          </div>
          </div>
          <div class="flex items-center gap-0.5">
            <button aria-label="Tiểu mục trước" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="skip(-1)"><SkipBack class="w-5 h-5" /></button>
            <button aria-label="Lùi 15 giây" title="Lùi 15 giây" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="seek(-15)"><RotateCcw class="w-5 h-5" /></button>
            <button :aria-label="player.playing ? 'Dừng' : 'Phát'" class="mx-1 h-12 w-12 grid place-items-center rounded-full bg-primary text-primary-foreground shadow" @click="toggle">
              <Pause v-if="player.playing" class="w-5 h-5" /><Play v-else class="w-5 h-5 ml-0.5" />
            </button>
            <button aria-label="Tới 30 giây" title="Tới 30 giây" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="seek(30)"><RotateCw class="w-5 h-5" /></button>
            <button aria-label="Tiểu mục sau" class="h-9 w-9 grid place-items-center rounded-full hover:bg-muted" @click="skip(1)"><SkipForward class="w-5 h-5" /></button>
          </div>
          <div class="justify-self-end">
          <div class="relative" data-pause-menu>
            <button class="h-8 px-2.5 rounded-full border flex items-center gap-1 whitespace-nowrap hover:bg-muted" aria-haspopup="listbox" :aria-expanded="pauseOpen"
              :class="pauseOpen ? 'border-foreground/40 bg-muted' : framed ? 'border-border bg-background/70' : 'border-border'" :title="pauseLabel" @click="pauseOpen = !pauseOpen">
              <Timer class="w-3.5 h-3.5" /><template v-if="!(framed && compact)">{{ pauseLabel }}</template><ChevronDown class="w-3 h-3 text-muted-foreground" />
            </button>
            <div v-if="pauseOpen" role="listbox" aria-label="Quãng nghỉ cho cuốn này" class="absolute bottom-full right-0 mb-2 w-72 max-h-[calc(100vh-8rem)] overflow-y-auto rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
              <div class="px-3 pt-1.5 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Quãng nghỉ cho cuốn này</div>
              <button role="option" :aria-selected="!mine" class="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-muted" :class="!mine && 'font-medium'" @click="pickPause('global')">
                <span class="text-sm">Theo cài đặt chung ({{ levelLabel(pauseState.global.level) }})<span class="block text-[11px] font-normal text-muted-foreground">{{ fmtGap(pauseState.global.section) }} giữa tiểu mục · {{ fmtGap(pauseState.global.chapter) }} sang chương</span></span>
                <Check v-if="!mine" class="w-4 h-4 shrink-0" />
              </button>
              <button v-for="p in PAUSE_PRESETS" :key="p.level" role="option" :aria-selected="mine?.level === p.level"
                class="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-muted" :class="mine?.level === p.level && 'font-medium'" @click="pickPause(p.level)">
                <span class="text-sm">{{ p.label }}<span class="block text-[11px] font-normal text-muted-foreground">{{ fmtGap(p.gaps.section) }} · {{ fmtGap(p.gaps.chapter) }}</span></span>
                <Check v-if="mine?.level === p.level" class="w-4 h-4 shrink-0" />
              </button>
              <button role="option" :aria-selected="mine?.level === 'custom'" class="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-muted" :class="mine?.level === 'custom' && 'font-medium'" @click="pickPause('custom')">
                <span class="text-sm">Tuỳ chỉnh<span class="block text-[11px] font-normal text-muted-foreground">{{ mine?.level === 'custom' ? `${fmtGap(mine.section)} · ${fmtGap(mine.chapter)}` : 'Tự đặt số giây' }}</span></span>
                <Check v-if="mine?.level === 'custom'" class="w-4 h-4 shrink-0" />
              </button>
              <PauseCustomInputs v-if="custom" compact class="mx-3 mb-1.5 mt-0.5 rounded-md bg-muted/60 p-2.5 text-sm" :value="customValue" @save="(g) => savePause(makeSetting('custom', g))" />
              <div class="mt-1 border-t border-border px-3 pt-2 pb-1.5 text-[11px] text-muted-foreground leading-relaxed">Lưu riêng cho cuốn này, dùng cả khi Xuất M4B. Sách truyện hợp mức Dài, sách kiến thức hợp Vừa hoặc Ngắn.</div>
            </div>
          </div>
          </div>
        </div>
        <p v-if="player.error || actionError" class="mt-2 max-w-md text-sm text-destructive line-clamp-2 shrink-0">{{ player.error || actionError }}</p>
        <!-- Thông báo nhỏ tự tắt (thay dòng xác nhận cố định) -->
        <div v-if="pauseMsg" role="status" class="absolute bottom-1 left-1/2 -translate-x-1/2 max-w-[calc(100%-1rem)] rounded-full bg-foreground text-background text-xs px-3 py-1.5 shadow-lg flex items-center gap-1.5 z-10">
          <Check class="w-3.5 h-3.5 shrink-0" /><span class="truncate">{{ pauseMsg }}</span>
        </div>
      </div>
      <div class="shrink-0 flex flex-wrap items-center gap-2 border-t border-border pt-4">
        <Button size="sm" variant="outline" class="border-primary/30 bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary" :disabled="!player.detail" title="Tạo một file sách nói (M4B) có bìa, mục lục chương, kèm hướng dẫn chép sang điện thoại" @click="openPhone(state.playerSlug, player.detail?.title ?? '', totalSec)"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</Button>
        <M4BProgress v-if="state.playerSlug" :slug="state.playerSlug" />
      </div>
    </div>
    <div ref="tocEl" class="w-72 shrink-0 border-l border-border overflow-auto">
      <!-- Đầu mục lục (D13): tiến độ cả cuốn -->
      <div class="sticky top-0 z-10 bg-background px-4 pt-3 pb-2.5 border-b border-border/60">
        <div class="flex items-baseline justify-between gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục
          <span class="normal-case font-normal tracking-normal tabular-nums">{{ tracks.length }} mục · {{ fmtLong(totalSec) }}</span></div>
        <div class="mt-2 h-1 rounded-full bg-muted overflow-hidden" role="progressbar" aria-label="Đã nghe cả cuốn" :aria-valuenow="Math.round(pctBook)"><div class="h-full bg-primary rounded-full" :style="{ width: pctBook + '%' }"></div></div>
        <div class="mt-1.5 text-[11px] text-muted-foreground tabular-nums">
          <template v-if="remainSec > 0">Đã nghe <b class="font-semibold text-foreground">{{ player.heard.length }}/{{ tracks.length }}</b> mục · còn {{ fmtLong(remainSec) }}</template>
          <template v-else>Đã nghe hết cuốn</template>
        </div>
      </div>
      <template v-for="(c, i) in tracks" :key="c.file">
        <!-- Tên chương đứng trước tiểu mục đầu của chương; chương chỉ có 1 tiểu mục trùng tên thì không lặp. -->
        <div v-if="c.chapterStart && c.chapter && c.chapter !== c.title"
          class="px-4 pt-4 pb-1 text-[11px] font-semibold uppercase tracking-wider truncate"
          :class="c.chapter === track?.chapter ? 'text-primary' : 'text-muted-foreground'"
          :title="c.chapter">{{ c.chapter }}</div>
        <button :data-current="i === player.current"
          class="w-full flex items-baseline justify-between gap-2 px-4 py-2 text-sm leading-snug text-left hover:bg-muted/60"
          :class="i === player.current ? 'bg-primary/10 text-primary font-medium' : player.heard.includes(i) ? 'text-muted-foreground' : ''"
          @click="pick(i)">
          <!-- Tên xuống tối đa 2 dòng (không cắt cụt); giờ thẳng hàng dòng đầu -->
          <span class="flex items-baseline gap-2 min-w-0">
            <span class="self-start w-3.5 h-[1.375em] shrink-0 grid place-items-center">
              <Volume2 v-if="i === player.current" class="w-3.5 h-3.5" />
              <Check v-else-if="player.heard.includes(i)" class="w-3.5 h-3.5" aria-label="Đã nghe" />
            </span>
            <span class="line-clamp-2" :title="c.title">{{ c.title }}</span>
          </span>
          <span class="text-xs tabular-nums shrink-0">{{ fmtClock(c.durationSec) }}</span>
        </button>
      </template>
    </div>
  </section>
</template>
