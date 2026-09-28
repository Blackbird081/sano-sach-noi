<script setup lang="ts">
// Hộp "Tạo video cả cuốn" (wireframe D15). Trái: xem trước từng cảnh vẽ bằng canvas
// (giống file xuất). Phải: phần sách, trích đoạn mở đầu, khung, nền, file kèm; lúc tạo
// hiện tiến độ theo bước; xong: mở thư mục, chép mô tả YouTube.
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Check, Clapperboard, Copy, FileText, FolderOpen, Image as ImageIcon, Loader2, Pause, Play, Square, Subtitles, Volume2, X } from 'lucide-vue-next'
import { usePreview } from '../lib/previewAudio'
import { Button } from '@/components/ui/button'
import { bv, bvBusy, buildTimeline, cancelBookVideo, chapterAt, clock, quickLyrics, startBookVideo, CHAPTER_SEC, END_SEC, TITLE_SEC, type BookVideoOptions, type Timeline } from '../lib/bookVideo'
import { useVideoPreview } from '../lib/videoPreview'
import { drawScene, type BVRatio, type Scene } from '../lib/bookVideoCard'
import { BACKGROUNDS, audioBars, loadImage, prepareCard } from '../lib/shareCard'
import { bookTexts, copyText, errText, openBookVideoFolder, type SectionText } from '../lib/backend'
import { buildLyrics, decodeSilences, sentenceBounds, snapToSilences } from '../lib/lyrics'
import { lyricIndex, player } from '../lib/player'

const KEY = 'sano.bookVideo'
function read() {
  try {
    const v = JSON.parse(localStorage.getItem(KEY) ?? 'null')
    return { ratio: (v?.ratio === 'tall' ? 'tall' : 'wide') as BVRatio, bg: Number.isInteger(v?.bg) && v.bg >= 0 && v.bg < 5 ? v.bg : 0, intro: v?.intro !== false }
  } catch {
    return { ratio: 'wide' as BVRatio, bg: 0, intro: true }
  }
}
const saved = read()
const ratio = ref<BVRatio>(saved.ratio)
const bg = ref(saved.bg)
const introOn = ref(saved.intro)
watch([ratio, bg, introOn], () => {
  try {
    localStorage.setItem(KEY, JSON.stringify({ ratio: ratio.value, bg: bg.value, intro: introOn.value }))
  } catch {
    // không lưu được thì thôi
  }
})
const extras = ref({ thumb: true, srt: true, desc: true })

const d = computed(() => (player.detail && player.detail.slug === bv.slug ? player.detail : null))
const texts = ref<SectionText[]>([])

// ── Phần sách: cả cuốn / từ chương … đến chương … ──
const groups = computed(() => {
  const b = d.value
  if (!b) return []
  const out: { name: string; first: number; last: number }[] = []
  b.tracks.forEach((t, i) => {
    if (i === 0 || t.chapterStart) out.push({ name: t.chapter || t.title, first: i, last: i })
    else out[out.length - 1].last = i
  })
  return out
})
const range = ref<'all' | 'some'>('all')
const fromG = ref(0)
const toG = ref(0)
const span = computed(() => {
  const g = groups.value
  if (!g.length) return { from: 0, to: 0 }
  if (range.value === 'all') return { from: 0, to: g[g.length - 1].last }
  const a = Math.min(fromG.value, toG.value)
  const b = Math.max(fromG.value, toG.value)
  return { from: g[a].first, to: g[b].last }
})
const bookSec = computed(() => (d.value ? d.value.tracks.slice(span.value.from, span.value.to + 1).reduce((n, t) => n + t.durationSec, 0) : 0))

// ── Trích đoạn mở đầu ──
const introTrack = ref(0)
const introPicked = ref<number[]>([])
const MAX_INTRO = 8
const MAX_INTRO_SEC = 30
// Khoảng lặng thật của tiểu mục trích đoạn: giờ câu bám giọng thật, cắt giữa khoảng lặng
// (không dính chữ câu kế như khi chỉ ước lượng theo độ dài chữ).
const introSil = ref<{ end: number; len: number }[] | null>(null)
const silCache = new Map<string, { end: number; len: number }[] | null>()
watch([introTrack, d], async () => {
  const t = d.value?.tracks[introTrack.value]
  introSil.value = null
  if (!t) return
  if (!silCache.has(t.url)) silCache.set(t.url, t.durationSec <= 20 * 60 ? await decodeSilences(t.url) : null)
  if (d.value?.tracks[introTrack.value]?.url === t.url) introSil.value = silCache.get(t.url) ?? null
}, { immediate: true })
const introLyrics = computed(() => {
  const b = d.value
  const tx = texts.value[introTrack.value]
  if (!b || !tx?.text) return null
  const l = buildLyrics(tx.text, tx.script, b.tracks[introTrack.value].durationSec)
  return introSil.value ? snapToSilences(l, introSil.value) : l
})
const introSentences = computed(() => {
  const l = introLyrics.value
  const t = d.value?.tracks[introTrack.value]
  if (!l || !t) return []
  const b = sentenceBounds(l, introSil.value, t.durationSec)
  return l.sentences.map((s, i) => ({ text: s.text, start: b[i].start, end: b[i].end }))
})
const lo = computed(() => Math.min(...introPicked.value))
const hi = computed(() => Math.max(...introPicked.value))
const introSpan = (a: number, b: number) => (introSentences.value[b]?.end ?? 0) - (introSentences.value[a]?.start ?? 0)
const introDur = computed(() => (introPicked.value.length ? introSpan(lo.value, hi.value) : 0))
function canAdd(i: number) {
  const p = introPicked.value
  return p.length < MAX_INTRO && (i === lo.value - 1 || i === hi.value + 1) && introSpan(Math.min(i, lo.value), Math.max(i, hi.value)) <= MAX_INTRO_SEC
}
function toggle(i: number) {
  const p = introPicked.value
  if (p.includes(i)) {
    if (p.length > 1 && (i === lo.value || i === hi.value)) introPicked.value = p.filter((x) => x !== i)
    return
  }
  introPicked.value = canAdd(i) ? [...p, i].sort((a, b) => a - b) : [i]
}
function autoFill(i: number) {
  const n = introSentences.value.length
  if (!n) {
    introPicked.value = []
    return
  }
  const p = [Math.min(i, n - 1)]
  while (p.length < MAX_INTRO && introSpan(p[0], p[p.length - 1]) < 15 && p[p.length - 1] + 1 < n && introSpan(p[0], p[p.length - 1] + 1) <= MAX_INTRO_SEC) p.push(p[p.length - 1] + 1)
  introPicked.value = p
}
const withText = computed(() => (d.value ? d.value.tracks.map((t, i) => ({ t, i })).filter(({ i }) => !!texts.value[i]?.text) : []))
function pickIntroTrack(i: number) {
  introTrack.value = i
  void nextTick(() => autoFill(0))
}

const listEl = ref<HTMLElement | null>(null)
// Nghe thử để dò đoạn hay: ▶ ở từng câu = nghe từ câu đó chạy tiếp; nút trên = đúng đoạn đã chọn.
const pv = usePreview(() => {
  const t = d.value?.tracks[introTrack.value]
  return t ? { url: t.url, sentences: introSentences.value } : null
})
watch(introTrack, () => pv.stop())
watch(() => pv.current.value, (i) => {
  if (i >= 0 && pv.mode.value === 'from') listEl.value?.querySelector(`[data-i="${i}"]`)?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
})
watch(() => bv.open, (o) => {
  if (!o) {
    pv.stop()
    vp.pause()
  }
})

// Nghe thử như video — ngay trong hộp thoại (không tạo video, không rời hộp): phát nối liền
// đúng dòng thời gian của video; ô xem trước vẽ khung hình theo giờ đang phát, tua được.
const vp = useVideoPreview()
const live = ref(false) // ô xem trước đang theo trình nghe thử
let tl: Timeline | null = null
function options(): BookVideoOptions {
  return {
    from: span.value.from, to: span.value.to, ratio: ratio.value, bg: bg.value, extras: { ...extras.value },
    intro: introOn.value && introPicked.value.length ? { track: introTrack.value, from: lo.value, to: hi.value } : null,
  }
}
function buildPreview() {
  const b = d.value
  if (!b) return null
  const lyr = quickLyrics(b, texts.value)
  // trích đoạn: dùng đúng giờ câu đang hiện trong danh sách chọn
  if (introLyrics.value) lyr.set(introTrack.value, introLyrics.value)
  tl = buildTimeline(b, options(), lyr, bars.value.length ? bars.value : Array.from({ length: 48 }, () => 0.35), new Map([[introTrack.value, introSil.value]]))
  vp.load(tl.segs)
  return tl
}
function listenLikeVideo() {
  const b = d.value
  if (!b) return
  pv.stop()
  if (!live.value || !tl) buildPreview()
  live.value = true
  const urls = new Map(b.tracks.map((t) => [t.file, t.url]))
  vp.play(vp.pos.value, (f) => urls.get(f) ?? f)
}
function stopListen() {
  vp.stop()
  live.value = false
  tl = null
  void nextTick(redraw)
}
// Đổi lựa chọn → dòng thời gian khác: dừng nghe thử (bấm lại để nghe bản mới).
watch([span, introOn, introPicked, introTrack, ratio], () => live.value && stopListen())
watch(() => pv.playing.value, (p) => p && vp.playing.value && vp.pause()) // ▶ ở một câu: tạm dừng nghe thử
let lastDraw = 0
let lastFrame = -1
watch(() => vp.pos.value, () => {
  if (!live.value) return
  const now = performance.now()
  const k = frameAt(vp.pos.value)
  if (k === lastFrame && now - lastDraw < 120) return
  lastDraw = now
  lastFrame = k
  drawLive(k)
})
function frameAt(t: number) {
  const f = tl?.frames ?? []
  let lo2 = 0
  let hi2 = f.length - 1
  while (lo2 < hi2) {
    const m = (lo2 + hi2 + 1) >> 1
    if (f[m].at <= t) lo2 = m
    else hi2 = m - 1
  }
  return lo2
}
function drawLive(k = frameAt(vp.pos.value)) {
  const b = d.value
  const f = tl?.frames[k]
  if (!canvas.value || !b || !f || !tl) return
  const t = vp.pos.value
  const introP = tl.introDur ? Math.min(1, t / tl.introDur) : 0
  const barP = tl.bookEnd > tl.bookStart ? Math.min(1, Math.max(0, (t - tl.bookStart) / (tl.bookEnd - tl.bookStart))) : 0
  drawScene(canvas.value, { ratio: ratio.value, bg: bg.value, cover: cover.value, title: b.title, voice: b.voice, totalLabel: tl.totalLabel }, f.scene, { progress: introP, bar: barP })
}
const liveLabel = computed(() => {
  if (!live.value || !tl) return ''
  const f = tl.frames[frameAt(vp.pos.value)]?.scene
  if (!f) return ''
  return { intro: 'Trích đoạn', title: 'Màn tựa', chapter: 'Thẻ chương', main: f.kind === 'main' ? f.section : '', end: 'Màn kết', thumb: '' }[f.kind]
})
function seekTo(e: MouseEvent) {
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect()
  vp.seek(((e.clientX - r.left) / r.width) * vp.total.value)
  drawLive()
}
async function init() {
  const b = d.value
  if (!b) return
  if (!texts.value.length || bv.slug !== loadedSlug) {
    try {
      texts.value = await bookTexts(b.slug)
      loadedSlug = b.slug
    } catch {
      texts.value = []
    }
  }
  // Mở lại cùng cuốn: giữ nguyên lựa chọn (vd vừa "Nghe thử như video" rồi quay lại tạo).
  if (initedSlug !== b.slug) {
    initedSlug = b.slug
    range.value = 'all'
    fromG.value = 0
    toG.value = Math.max(0, groups.value.length - 1)
    const cur = texts.value[player.current]?.text ? player.current : (withText.value[0]?.i ?? 0)
    introTrack.value = cur
    await nextTick()
    autoFill(cur === player.current ? lyricIndex.value : 0)
  }
  await prepareCard()
  cover.value = b.coverUrl ? await loadImage(b.coverUrl) : null
  void refreshBars()
  redraw()
  void nextTick(() => listEl.value?.querySelector(`[data-i="${lo.value}"]`)?.scrollIntoView({ block: 'center' }))
}
let loadedSlug = ''
let initedSlug = ''
watch(() => bv.open, (o) => o && void init(), { immediate: true })

// ── Xem trước ──
type Tab = 'intro' | 'title' | 'main' | 'end' | 'thumb'
const tab = ref<Tab>('main')
const tabs: { key: Tab; label: string }[] = [
  { key: 'intro', label: '1. Trích đoạn' },
  { key: 'title', label: '2. Màn tựa' },
  { key: 'main', label: '3. Trong sách' },
  { key: 'end', label: '4. Màn kết' },
  { key: 'thumb', label: 'Thumbnail' },
]
const canvas = ref<HTMLCanvasElement | null>(null)
const cover = ref<HTMLImageElement | null>(null)
const bars = ref<number[]>([])
async function refreshBars() {
  const t = d.value?.tracks[introTrack.value]
  if (!t || !introPicked.value.length) return
  const a = introSentences.value[lo.value]?.start ?? 0
  bars.value = await audioBars(t.url, a, a + introDur.value, 48).catch(() => Array.from({ length: 48 }, () => 0.35))
  redraw()
}
watch([introPicked, introTrack], () => void refreshBars())
const minutes = computed(() => Math.max(1, Math.round(bookSec.value / 60)))
const chapterCount = computed(() => {
  const b = d.value
  if (!b) return 0
  let n = 0
  for (let i = span.value.from; i <= span.value.to; i++) if (chapterAt(b, i, span.value.from)) n++
  return n
})
const ticks = computed(() => {
  const b = d.value
  if (!b) return []
  const out: number[] = []
  let x = 0
  for (let i = span.value.from; i <= span.value.to; i++) {
    if (chapterAt(b, i, span.value.from)) out.push(x)
    x += b.tracks[i].durationSec
  }
  return out.map((v) => (x ? v / x : 0))
})
function sample(): Scene {
  const b = d.value!
  const ss = introSentences.value
  if (tab.value === 'intro') {
    const s = ss[introPicked.value[0]]?.text ?? b.title
    return { kind: 'intro', line: s, bars: bars.value.length ? bars.value : Array.from({ length: 48 }, () => 0.35), reserve: 1 }
  }
  const badge = span.value.from > 0 || span.value.to < b.tracks.length - 1 ? 'NGHE THỬ SÁCH NÓI' : 'SÁCH NÓI ĐẦY ĐỦ'
  if (tab.value === 'title') return { kind: 'title', meta: `Giọng ${b.voice} · ${minutes.value} phút${chapterCount.value ? ` · ${chapterCount.value} chương` : ''}`, badge }
  if (tab.value === 'thumb') return { kind: 'thumb', meta: `${minutes.value} phút · Giọng ${b.voice}`, badge }
  if (tab.value === 'end') return { kind: 'end' }
  const i = Math.min(Math.max(span.value.from, player.current), span.value.to)
  const t = b.tracks[i]
  const tx = texts.value[i]
  const l = tx?.text ? buildLyrics(tx.text, tx.script, t.durationSec) : null
  const k = i === player.current ? lyricIndex.value : 0
  let ch = ''
  for (let j = i; j >= span.value.from; j--) if ((ch = chapterAt(b, j, span.value.from))) break
  return { kind: 'main', chapter: ch || t.chapter || b.title, section: t.title, line: l?.sentences[k]?.text ?? t.title, next: l?.sentences[k + 1]?.text ?? '', ticks: ticks.value }
}
function redraw() {
  const b = d.value
  if (!canvas.value || !b) return
  if (live.value) return drawLive()
  drawScene(canvas.value, { ratio: ratio.value, bg: bg.value, cover: cover.value, title: b.title, voice: b.voice, totalLabel: clock(bookSec.value) }, sample(), { progress: tab.value === 'intro' ? 0.4 : undefined })
}
watch([tab, ratio, bg, span, introPicked, introTrack, canvas, texts], () => void nextTick(redraw))
const previewWide = computed(() => ratio.value === 'wide' || (tab.value === 'thumb' && !live.value))

// ── Ước lượng ──
const totalSec = computed(() => bookSec.value + (introOn.value ? introDur.value : 0) + TITLE_SEC + END_SEC + chapterCount.value * CHAPTER_SEC)
// Đo thật: cuốn 29 phút ra 64 MB (~2,3 MB/phút, hình gần như tĩnh).
const estMB = computed(() => Math.max(3, Math.round((totalSec.value / 60) * 2.5)))

function create() {
  const b = d.value
  if (!b) return
  stopListen()
  pv.stop()
  void startBookVideo(b, {
    from: span.value.from, to: span.value.to, ratio: ratio.value, bg: bg.value, extras: { ...extras.value },
    intro: introOn.value && introPicked.value.length ? { track: introTrack.value, from: lo.value, to: hi.value } : null,
  })
}
const steps = computed(() => {
  const order = ['prep', 'frames', 'audio', 'video', 'files']
  const cur = order.indexOf(bv.step)
  return [
    { t: bv.step === 'prep' ? `Dò lời từng tiểu mục · ${bv.done}/${bv.total}` : 'Dò lời từng tiểu mục' },
    { t: bv.step === 'frames' ? `Vẽ khung hình · ${bv.done}/${bv.total}` : 'Vẽ khung hình' },
    { t: 'Ghép tiếng cả cuốn' },
    { t: bv.step === 'video' ? `Mã hoá video · ${bv.pct}%` : 'Mã hoá video' },
  ].map((s, i) => ({ ...s, state: i < cur ? 'done' : i === cur || (cur === 4 && i === 3) ? (cur === 4 ? 'done' : 'run') : 'wait' }))
})
const overall = computed(() => {
  const f = bv.total ? bv.done / bv.total : 0
  switch (bv.step) {
    case 'prep': return 5 * f
    case 'frames': return 5 + 35 * f
    case 'audio': return 40 + 10 * (bv.pct / 100)
    case 'video': return 50 + 48 * (bv.pct / 100)
    case 'files': return 99
    default: return 0
  }
})
const copied = ref(false)
async function copyDesc() {
  copied.value = await copyText(bv.desc)
  setTimeout(() => (copied.value = false), 2500)
}
const actionErr = ref('')
async function openFolder() {
  actionErr.value = ''
  try {
    await openBookVideoFolder()
  } catch (e) {
    actionErr.value = errText(e)
  }
}
function again() {
  bv.step = 'idle'
  void nextTick(redraw)
}
function close() {
  bv.open = false
}
function onKey(e: KeyboardEvent) {
  if (bv.open && e.key === 'Escape') close()
}
document.addEventListener('keydown', onKey)
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
const busy = computed(bvBusy)
const fmtSize = (b: number) => (b > 1e9 ? (b / 1e9).toLocaleString('vi-VN', { maximumFractionDigits: 1 }) + ' GB' : Math.round(b / 1e6) + ' MB')
</script>

<template>
  <div v-if="bv.open && d" class="absolute inset-0 bg-background/70 backdrop-blur-sm grid place-items-center z-20" @mousedown.self="close">
    <div role="dialog" aria-modal="true" aria-labelledby="bv-title" class="relative w-[1060px] max-w-[calc(100vw-2rem)] h-[690px] max-h-[calc(100vh-2rem)] rounded-2xl border border-border bg-card text-card-foreground shadow-2xl flex overflow-hidden">
      <!-- Trái: xem trước -->
      <div class="w-[540px] max-w-[48%] shrink-0 bg-muted/50 flex flex-col items-center justify-center gap-3 p-5 min-h-0">
        <div class="flex gap-1 p-1 rounded-lg bg-muted text-xs shrink-0">
          <button v-for="s in tabs" :key="s.key" class="h-7 px-2.5 rounded-md" :class="[tab === s.key && !live ? 'bg-background shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground', s.key === 'intro' && !introOn && 'line-through opacity-50']" @click="live && stopListen(); tab = s.key">{{ s.label }}</button>
        </div>
        <canvas ref="canvas" class="rounded-xl shadow-2xl object-contain shrink min-h-0" :class="previewWide ? 'w-[500px] max-w-full h-auto' : 'h-[470px] max-h-[calc(100vh-13rem)] w-auto'" aria-label="Xem trước cảnh video"></canvas>
        <!-- Nghe thử như video: phát / dừng, tua, giờ -->
        <div class="w-full max-w-[500px] shrink-0 flex items-center gap-2.5">
          <button class="h-9 w-9 shrink-0 rounded-full bg-primary text-primary-foreground grid place-items-center shadow disabled:opacity-50" :disabled="!d.tracks.length || busy"
            :aria-label="vp.playing.value ? 'Dừng nghe thử' : 'Nghe thử như video'" :title="vp.playing.value ? 'Dừng nghe thử' : 'Nghe thử như video — không cần tạo'" @click="vp.playing.value ? vp.pause() : listenLikeVideo()">
            <Pause v-if="vp.playing.value" class="w-4 h-4" /><Play v-else class="w-4 h-4 ml-0.5" />
          </button>
          <div class="flex-1 min-w-0">
            <div class="h-1.5 rounded-full bg-muted-foreground/20 cursor-pointer relative" role="slider" aria-label="Vị trí nghe thử" :aria-valuenow="Math.round(vp.pos.value)" @click="live ? seekTo($event) : (listenLikeVideo(), vp.pause(), seekTo($event))">
              <div class="absolute inset-y-0 left-0 rounded-full bg-primary" :style="{ width: (vp.total.value ? (vp.pos.value / vp.total.value) * 100 : 0) + '%' }"></div>
            </div>
            <div class="mt-1 flex justify-between gap-2 text-[11px] text-muted-foreground tabular-nums">
              <span class="truncate">{{ live ? liveLabel : 'Nghe thử như video — bấm ▶, không cần tạo' }}</span>
              <span class="shrink-0">{{ clock(vp.pos.value) }} / {{ clock(live ? vp.total.value : totalSec) }}</span>
            </div>
          </div>
          <button v-if="live" class="h-8 w-8 shrink-0 rounded-md grid place-items-center text-muted-foreground hover:bg-muted hover:text-foreground" aria-label="Thôi nghe thử" title="Thôi nghe thử" @click="stopListen"><Square class="w-3.5 h-3.5" /></button>
        </div>
        <p class="text-[11px] text-muted-foreground text-center max-w-[500px] shrink-0">
          {{ { intro: 'Trích đoạn: câu đổi theo lời đọc, sóng âm sáng dần — người xem nghe thử trước khi vào sách.', title: `Màn tựa ${TITLE_SEC} giây.`, main: 'Suốt cuốn: câu đang đọc chữ to, câu kế nhạt; tên chương; thanh tiến độ cả cuốn có vạch chương, sáng dần theo thời gian. Mỗi chương mở bằng thẻ tên chương.', end: `Màn kết ${END_SEC} giây, kêu gọi dùng Sano.`, thumb: 'Ảnh thumbnail 1280×720 (file riêng, luôn khung ngang): chữ to, đọc được cả khi YouTube thu nhỏ.' }[tab] }}
        </p>
      </div>

      <!-- Phải -->
      <div class="flex-1 min-w-0 flex flex-col">
        <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
          <div class="min-w-0">
            <h2 id="bv-title" class="font-semibold flex items-center gap-2"><Clapperboard class="w-4 h-4" /> Tạo video cả cuốn</h2>
            <p class="text-xs text-muted-foreground mt-0.5 truncate">Đăng YouTube, Facebook · {{ d.title }}</p>
          </div>
          <button aria-label="Đóng" class="h-8 w-8 grid place-items-center rounded-md hover:bg-muted" @click="close"><X class="w-4 h-4" /></button>
        </div>

        <!-- Tuỳ chọn -->
        <div v-if="!busy && bv.step !== 'done'" class="flex-1 min-h-0 overflow-auto px-5 py-3 space-y-3 text-sm">
          <div>
            <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Phần sách</span>
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button class="h-10 rounded-lg border px-3 text-left" :class="range === 'all' ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="range = 'all'">
                <span class="block font-medium">Cả cuốn</span><span class="block text-[11px] text-muted-foreground">{{ groups.length }} chương · {{ Math.round(d.tracks.reduce((n, t) => n + t.durationSec, 0) / 60) }} phút</span>
              </button>
              <button class="h-10 rounded-lg border px-3 text-left" :class="range === 'some' ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="range = 'some'">
                <span class="block font-medium">Một vài chương</span><span class="block text-[11px] text-muted-foreground">Ví dụ chương 1 làm mồi</span>
              </button>
            </div>
            <div v-if="range === 'some'" class="mt-2 flex items-center gap-2 text-xs">
              Từ <select v-model.number="fromG" class="h-8 min-w-0 flex-1 rounded-md border border-border bg-background px-2"><option v-for="(g, i) in groups" :key="i" :value="i">{{ g.name }}</option></select>
              đến <select v-model.number="toG" class="h-8 min-w-0 flex-1 rounded-md border border-border bg-background px-2"><option v-for="(g, i) in groups" :key="i" :value="i" :disabled="i < fromG">{{ g.name }}</option></select>
            </div>
          </div>

          <div>
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Mở đầu bằng đoạn hay nhất</span>
              <button role="switch" :aria-checked="introOn" aria-label="Mở đầu bằng đoạn hay nhất" class="h-5 w-9 rounded-full relative transition-colors" :class="introOn ? 'bg-primary' : 'bg-muted-foreground/30'" @click="introOn = !introOn">
                <span class="absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all" :class="introOn ? 'left-[18px]' : 'left-0.5'"></span>
              </button>
            </div>
            <template v-if="introOn">
              <div class="mt-2 flex items-center justify-between gap-2 text-xs">
                <select :value="introTrack" class="h-8 min-w-0 flex-1 rounded-md border border-border bg-background px-2" aria-label="Tiểu mục của trích đoạn" @change="pickIntroTrack(Number(($event.target as HTMLSelectElement).value))">
                  <option v-for="x in withText" :key="x.i" :value="x.i">{{ x.t.title }}</option>
                </select>
                <button class="h-8 px-2.5 rounded-md border text-xs flex items-center gap-1.5 shrink-0" :class="pv.playing.value && pv.mode.value === 'span' ? 'border-primary text-primary bg-primary/5' : 'border-border hover:bg-muted'"
                  :disabled="!introPicked.length" @click="pv.playing.value && pv.mode.value === 'span' ? pv.stop() : pv.play(lo, hi)">
                  <component :is="pv.playing.value && pv.mode.value === 'span' ? Square : Play" class="w-3.5 h-3.5" /> Nghe đoạn
                </button>
                <span class="text-muted-foreground tabular-nums shrink-0">{{ introPicked.length }}/{{ MAX_INTRO }} câu · {{ Math.round(introDur) }} giây</span>
              </div>
              <div ref="listEl" class="mt-2 max-h-[112px] overflow-auto rounded-lg border border-border divide-y divide-border">
                <div v-for="(s, i) in introSentences" :key="i" :data-i="i" role="checkbox" tabindex="0" :aria-checked="introPicked.includes(i)"
                  class="group w-full flex items-start gap-2.5 px-3 py-1.5 text-left cursor-pointer outline-none focus-visible:bg-muted" :class="introPicked.includes(i) ? 'bg-primary/5' : 'hover:bg-muted/60'"
                  @click="toggle(i)" @keydown.enter.prevent="toggle(i)" @keydown.space.prevent="toggle(i)">
                  <span class="mt-0.5 h-4 w-4 rounded border grid place-items-center shrink-0" :class="introPicked.includes(i) ? 'bg-primary border-primary text-primary-foreground' : canAdd(i) ? 'border-primary/50' : 'border-border'"><Check v-if="introPicked.includes(i)" class="w-3 h-3" /></span>
                  <span class="flex-1 leading-snug text-[13px]" :class="pv.current.value === i ? 'text-primary font-medium' : introPicked.includes(i) ? '' : 'text-muted-foreground'">{{ s.text }}</span>
                  <button class="h-6 w-6 -my-0.5 shrink-0 rounded-full grid place-items-center" :class="pv.current.value === i ? 'text-primary' : 'text-muted-foreground opacity-0 group-hover:opacity-100 focus:opacity-100 hover:bg-muted hover:text-foreground'"
                    :title="pv.current.value === i ? 'Dừng nghe thử' : 'Nghe từ câu này'" :aria-label="pv.current.value === i ? 'Dừng nghe thử' : 'Nghe từ câu này'" @click.stop="pv.current.value === i ? pv.stop() : pv.play(i)">
                    <Volume2 v-if="pv.current.value === i" class="w-3.5 h-3.5" /><Play v-else class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
              <p class="mt-1 text-[11px] text-muted-foreground truncate">Bấm ▶ ở một câu để nghe tiếp từ đó, dò đoạn hay · tối đa {{ MAX_INTRO_SEC }} giây</p>
            </template>
            <p v-else class="mt-1 text-[11px] text-muted-foreground">Tắt: video phát từ đầu đến cuối như thường.</p>
          </div>

          <div class="flex items-start gap-5">
            <div class="flex-1 min-w-0">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Khung</span>
              <div class="mt-2 grid grid-cols-2 gap-1.5">
                <button v-for="r in (['wide', 'tall'] as const)" :key="r" class="h-11 rounded-lg border flex items-center gap-2 px-2.5 text-left min-w-0" :class="ratio === r ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="ratio = r">
                  <span class="border-2 rounded-sm shrink-0" :class="[r === 'wide' ? 'w-6 h-4' : 'w-3.5 h-6', ratio === r ? 'border-primary' : 'border-muted-foreground/50']"></span>
                  <span class="leading-tight min-w-0"><span class="block text-[13px] font-medium">{{ r === 'wide' ? 'Ngang 16:9' : 'Dọc 9:16' }}</span><span class="block text-[10px] text-muted-foreground truncate">{{ r === 'wide' ? 'YouTube · 1920×1080' : 'Điện thoại · 1080×1920' }}</span></span>
                </button>
              </div>
            </div>
            <div class="shrink-0">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Nền · {{ BACKGROUNDS[bg].label }}</span>
              <div class="mt-2 flex gap-1.5">
                <button v-for="(b, i) in BACKGROUNDS" :key="i" class="h-11 w-7 rounded-lg overflow-hidden relative ring-offset-2 ring-offset-card" :class="[b.swatch, bg === i ? 'ring-2 ring-primary' : 'ring-1 ring-border']" :title="b.label" :aria-label="b.label" @click="bg = i">
                  <img v-if="!b.stops.length && d.coverUrl" :src="d.thumbUrl || d.coverUrl" alt="" class="absolute -inset-2 w-[calc(100%+1rem)] h-[calc(100%+1rem)] max-w-none object-cover blur-md" />
                </button>
              </div>
            </div>
          </div>

          <div>
            <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Kèm theo để đăng YouTube</span>
            <div class="mt-2 rounded-lg border border-border divide-y divide-border">
              <button v-for="x in ([{ k: 'thumb', i: ImageIcon, t: 'Ảnh thumbnail 1280×720', d: 'bìa + “Sách nói đầy đủ”' }, { k: 'srt', i: Subtitles, t: 'Phụ đề .srt từ lời đọc', d: 'dễ lên tìm kiếm YouTube' }, { k: 'desc', i: FileText, t: 'Mô tả có mốc chương', d: 'YouTube tự chia chương, kèm link' }] as const)" :key="x.k"
                role="checkbox" :aria-checked="extras[x.k]" class="w-full flex items-center gap-2.5 px-3 h-8 text-left hover:bg-muted/40" @click="extras[x.k] = !extras[x.k]">
                <span class="h-4 w-4 rounded border grid place-items-center shrink-0" :class="extras[x.k] ? 'bg-primary border-primary text-primary-foreground' : 'border-border'"><Check v-if="extras[x.k]" class="w-3 h-3" /></span>
                <component :is="x.i" class="w-4 h-4 shrink-0 text-muted-foreground" />
                <span class="text-[13px] font-medium whitespace-nowrap">{{ x.t }}</span><span class="text-[11px] text-muted-foreground truncate">· {{ x.d }}</span>
              </button>
            </div>
          </div>
          <p v-if="bv.step === 'error'" class="text-xs text-destructive">{{ bv.error }}</p>
        </div>

        <!-- Đang tạo -->
        <div v-else-if="busy" class="flex-1 min-h-0 px-5 py-6 text-sm">
          <p class="font-medium">Đang tạo video {{ Math.round(totalSec / 60) }} phút…</p>
          <p class="text-xs text-muted-foreground mt-0.5">Vẫn nghe sách, dùng app bình thường được. Đóng hộp này thì video vẫn tạo tiếp ở nền.</p>
          <div class="mt-4 h-2 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full transition-all" :style="{ width: overall + '%' }"></div></div>
          <p class="mt-1 text-xs text-muted-foreground tabular-nums">{{ Math.round(overall) }}%</p>
          <ol class="mt-5 space-y-2.5">
            <li v-for="s in steps" :key="s.t" class="flex items-center gap-2.5 tabular-nums" :class="s.state === 'wait' && 'text-muted-foreground'">
              <span class="h-5 w-5 rounded-full grid place-items-center shrink-0" :class="s.state === 'done' ? 'bg-primary text-primary-foreground' : s.state === 'run' ? 'border-2 border-primary' : 'border border-border'">
                <Check v-if="s.state === 'done'" class="w-3 h-3" /><Loader2 v-else-if="s.state === 'run'" class="w-3 h-3 animate-spin text-primary" />
              </span>{{ s.t }}
            </li>
          </ol>
        </div>

        <!-- Xong -->
        <div v-else class="flex-1 min-h-0 overflow-auto px-5 py-4 text-sm space-y-3">
          <p class="flex items-center gap-2 font-medium"><span class="h-5 w-5 rounded-full bg-primary text-primary-foreground grid place-items-center"><Check class="w-3 h-3" /></span> Đã tạo xong · {{ clock(bv.durSec) }} · {{ fmtSize(bv.bytes) }}</p>
          <div class="rounded-lg border border-border p-3 text-xs space-y-1">
            <p class="text-[11px] text-muted-foreground mb-1 truncate" :title="bv.dir">{{ bv.dir }}</p>
            <p class="flex items-center gap-1.5 font-mono"><Clapperboard class="w-3.5 h-3.5 shrink-0" /><span class="truncate" :title="bv.video">{{ bv.video.split(/[\\/]/).pop() }}</span></p>
            <p v-if="extras.thumb" class="flex items-center gap-1.5 font-mono"><ImageIcon class="w-3.5 h-3.5" /> thumbnail.png</p>
            <p v-if="extras.srt" class="flex items-center gap-1.5 font-mono"><Subtitles class="w-3.5 h-3.5" /> phu-de.srt</p>
            <p v-if="extras.desc" class="flex items-center gap-1.5 font-mono"><FileText class="w-3.5 h-3.5" /> mo-ta-youtube.txt</p>
          </div>
          <div v-if="bv.desc">
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Mô tả YouTube</span>
              <button class="h-7 px-2.5 rounded-md border border-border text-xs flex items-center gap-1.5 hover:bg-muted" @click="copyDesc"><component :is="copied ? Check : Copy" class="w-3.5 h-3.5" /> {{ copied ? 'Đã chép' : 'Chép mô tả' }}</button>
            </div>
            <pre class="mt-2 max-h-[220px] overflow-auto rounded-lg bg-muted/60 p-3 text-[11.5px] leading-relaxed whitespace-pre-wrap font-sans">{{ bv.desc }}</pre>
          </div>
          <p class="text-[11px] text-muted-foreground">Đăng YouTube: tải video lên, chọn thumbnail.png làm hình thu nhỏ, dán mô tả, vào Phụ đề → Tải tệp lên → phu-de.srt.</p>
          <p v-if="actionErr" class="text-xs text-destructive">{{ actionErr }}</p>
        </div>

        <div class="shrink-0 border-t border-border px-5 py-3.5">
          <template v-if="!busy && bv.step !== 'done'">
            <Button class="w-full" :disabled="!d.tracks.length" @click="create"><Clapperboard class="w-4 h-4" /> Tạo video {{ Math.round(totalSec / 60) }} phút</Button>
            <p class="mt-2 text-[11px] text-muted-foreground text-center">MP4 {{ ratio === 'wide' ? '1920×1080' : '1080×1920' }} · khoảng {{ estMB }} MB · tạo trên máy, không cần mạng.</p>
          </template>
          <Button v-else-if="busy" variant="outline" class="w-full" @click="cancelBookVideo">Huỷ</Button>
          <div v-else class="flex gap-2">
            <Button class="flex-1" @click="openFolder"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
            <Button variant="outline" @click="again">Tạo video khác</Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
