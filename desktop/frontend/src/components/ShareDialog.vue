<script setup lang="ts">
// Hộp "Chia sẻ đoạn hay" (wireframe D14). Trái: xem trước vẽ bằng canvas đúng như
// file sẽ lưu. Phải: chọn câu liền nhau, khung, nền; lưu, chép, AirDrop.
// D17: mở theo một loại — "Chia sẻ ảnh" (chỉ ảnh) hoặc "Tạo video" · Video ngắn (đầu hộp có
// thẻ chọn sang Video cả cuốn, cùng cỡ hộp D15 để đổi loại không nhảy khung).
// Chụp lại tiểu mục lúc mở (câu, tiếng, bìa) nên nghe tiếp sang mục khác không đổi thẻ.
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Check, Clapperboard, Copy, Download, Film, Image as ImageIcon, Loader2, Play, Share2, Pause, Square, Volume2, X } from 'lucide-vue-next'
import { usePreview } from '../lib/previewAudio'
import { Button } from '@/components/ui/button'
import VideoKindTabs from './VideoKindTabs.vue'
import { shareUI } from '../lib/share'
import { BACKGROUNDS, audioBars, canvasPNG, drawBrightBars, drawCard, loadImage, prepareCard, videoLines, type ShareRatio } from '../lib/shareCard'
import {
  airDropShareImage, airDropShareVideo, canAirDrop, cancelShareVideo, copyShareImage, errText, makeShareVideo, onEvent, saveShareImage, saveShareVideo,
} from '../lib/backend'
import { lyrics, player, track, trackSilences } from '../lib/player'
import { decodeSilences, sentenceBounds, snapToSilences } from '../lib/lyrics'

// Ảnh: tối đa 4 câu (nhiều hơn chữ phải thu nhỏ, khó đọc). Video: tối đa 8 câu / 30 giây,
// mở ra chọn sẵn ~15 giây (Reels 15–30 giây dễ được xem hết nhất; mỗi câu đọc ~3,3 giây).
const MAX_IMAGE = 4
const MAX_VIDEO = 8
const MAX_VIDEO_SEC = 30
const AUTO_VIDEO_SEC = 15
const MAX = computed(() => (kind.value === 'video' ? MAX_VIDEO : MAX_IMAGE))

interface Snap {
  slug: string
  file: string
  url: string
  section: string
  title: string
  voice: string
  sentences: { text: string; start: number; end: number }[]
}
const snap = ref<Snap | null>(null)
const cover = ref<HTMLImageElement | null>(null)
const picked = ref<number[]>([])
const airdrop = ref(false)
void canAirDrop().then((v) => (airdrop.value = v))

async function init() {
  const d = player.detail
  const t = track.value
  const ly = lyrics.value
  if (!d || !t || !ly?.sentences.length) {
    shareUI.open = false
    return
  }
  // Giờ câu bám khoảng lặng thật; điểm cắt ở giữa khoảng lặng để video không dính chữ câu kế.
  // Màn nghe chưa kịp dò khoảng lặng thì dò ngay (trên bản sao, không đụng lời đang hiện).
  let sil = trackSilences.value
  let l = { paragraphs: ly.paragraphs, sentences: ly.sentences.map((x) => ({ ...x })) }
  if (!sil && t.durationSec <= 20 * 60) {
    sil = await decodeSilences(t.url)
    if (sil) l = snapToSilences(l, sil)
  }
  const ss = l.sentences
  const bounds = sentenceBounds(l, sil, t.durationSec)
  snap.value = {
    slug: d.slug, file: t.file, url: t.url, section: t.title, title: d.title, voice: d.voice,
    sentences: ss.map((x, i) => ({ text: x.text, start: bounds[i].start, end: bounds[i].end })),
  }
  const i = Math.min(Math.max(0, shareUI.sentence), ss.length - 1)
  picked.value = [i]
  if (shareUI.kind === 'video') autoFill()
  vstate.value = 'idle'
  err.value = ''
  cover.value = null
  await prepareCard()
  if (d.coverUrl) cover.value = await loadImage(d.coverUrl)
  redraw()
  void nextTick(() => listEl.value?.querySelector(`[data-i="${i}"]`)?.scrollIntoView({ block: 'center' }))
}
watch(() => shareUI.open, (o) => (o ? void init() : stopPreview()))
function onKey(e: KeyboardEvent) {
  if (shareUI.open && e.key === 'Escape') close()
}
document.addEventListener('keydown', onKey)
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))

const kind = computed(() => shareUI.kind)
const ratio = computed({ get: () => shareUI.ratio, set: (v: ShareRatio) => (shareUI.ratio = v) })

const lo = computed(() => Math.min(...picked.value))
const hi = computed(() => Math.max(...picked.value))
const span = (a: number, b: number) => (snap.value ? snap.value.sentences[b].end - snap.value.sentences[a].start : 0)
const dur = computed(() => (picked.value.length ? Math.min(MAX_VIDEO_SEC, span(lo.value, hi.value)) : 0))
function canAdd(i: number) {
  const p = picked.value
  if (p.length >= MAX.value || !(i === lo.value - 1 || i === hi.value + 1)) return false
  return kind.value === 'image' || span(Math.min(i, lo.value), Math.max(i, hi.value)) <= MAX_VIDEO_SEC
}
function toggle(i: number) {
  const p = picked.value
  if (p.includes(i)) {
    // chỉ bỏ được câu ở hai đầu để đoạn luôn liền nhau
    if (p.length > 1 && (i === lo.value || i === hi.value)) picked.value = p.filter((x) => x !== i)
    return
  }
  picked.value = canAdd(i) ? [...p, i].sort((a, b) => a - b) : [i] // câu không liền kề → chọn lại từ đầu
}
// Video: thêm dần các câu kế tiếp (rồi câu trước nếu hết tiểu mục) tới ~15 giây.
function autoFill() {
  const s = snap.value
  if (!s) return
  const p = [...picked.value]
  const len = () => span(Math.min(...p), Math.max(...p))
  while (p.length < MAX_VIDEO && len() < AUTO_VIDEO_SEC) {
    const next = Math.max(...p) + 1
    const prev = Math.min(...p) - 1
    const add = next < s.sentences.length ? next : prev >= 0 ? prev : -1
    if (add < 0) break
    const q = [...p, add]
    if (span(Math.min(...q), Math.max(...q)) > MAX_VIDEO_SEC) break
    p.push(add)
  }
  picked.value = p.sort((a, b) => a - b)
}
// Sang video: đoạn ngắn thì chọn thêm cho đủ ~15 giây, dài quá thì bớt câu cuối cho vừa 30 giây.
// Sang ảnh: giữ tối đa 4 câu đầu.
watch(kind, (k) => {
  if (!snap.value) return
  if (k === 'image') {
    if (picked.value.length > MAX_IMAGE) picked.value = picked.value.slice(0, MAX_IMAGE)
    return
  }
  while (picked.value.length > 1 && span(lo.value, hi.value) > MAX_VIDEO_SEC) picked.value = picked.value.slice(0, -1)
  if (span(lo.value, hi.value) < AUTO_VIDEO_SEC - 5) autoFill()
})

// ── Sóng âm thật của đoạn đã chọn (video) ──
const BARS = 40
const bars = ref<number[]>([])
let barsSeq = 0
watch([kind, picked, snap], async () => {
  const s = snap.value
  if (kind.value !== 'video' || !s || !picked.value.length) return
  const seq = ++barsSeq
  const start = s.sentences[lo.value].start
  try {
    const b = await audioBars(s.url, start, start + dur.value, BARS)
    if (seq === barsSeq) bars.value = b
  } catch {
    if (seq === barsSeq) bars.value = Array.from({ length: BARS }, () => 0.35) // không đọc được tiếng: cột đều
  }
  redraw()
})
const reserve = computed(() => (snap.value ? Math.max(0, ...picked.value.map((i) => videoLines(ratio.value, snap.value!.sentences[i].text))) : 0))

// ── Xem trước ──
const canvas = ref<HTMLCanvasElement | null>(null)
const listEl = ref<HTMLElement | null>(null)
// Nghe thử: ▶ ở từng câu = nghe từ câu đó chạy tiếp (dò đoạn hay); nút trên = đúng đoạn đã chọn.
const pv = usePreview(() => (snap.value ? { url: snap.value.url, sentences: snap.value.sentences } : null))
watch(() => pv.current.value, (i) => {
  if (i >= 0 && pv.mode.value === 'from') listEl.value?.querySelector(`[data-i="${i}"]`)?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
})
watch(() => shareUI.open, (o) => !o && pv.stop())
const previewing = ref(false)
const vIdx = ref(0)
const pprog = ref(0) // phần đã chạy của bản xem thử video (0–1)
let raf = 0
let t0 = 0
function cardOpts(lines: string[], progress = 0) {
  const s = snap.value!
  return {
    kind: kind.value, ratio: ratio.value, bg: shareUI.bg, lines, title: s.title, voice: s.voice, cover: cover.value,
    bars: bars.value, progress, reserveLines: reserve.value,
  }
}
const pickedText = computed(() => picked.value.map((i) => snap.value?.sentences[i].text ?? ''))
function redraw() {
  if (!canvas.value || !snap.value || !picked.value.length) return
  if (kind.value === 'image') drawCard(canvas.value, cardOpts(pickedText.value))
  else drawCard(canvas.value, cardOpts([pickedText.value[Math.min(vIdx.value, pickedText.value.length - 1)]], pprog.value))
}
watch([kind, ratio, () => shareUI.bg, picked, cover, canvas], () => {
  vIdx.value = 0
  pprog.value = 0
  redraw()
})
// Xem thử video (không tiếng): chữ đổi theo thời lượng thật của từng câu, sóng âm tô dần.
function loop(t: number) {
  const s = snap.value
  if (!previewing.value || !s) return
  if (!t0) t0 = t
  const start = s.sentences[lo.value].start
  let el = (t - t0) / 1000
  if (el > dur.value) {
    t0 = t
    el = 0
  }
  pprog.value = dur.value ? el / dur.value : 0
  const at = start + el
  let k = picked.value.findIndex((i) => at < s.sentences[i].end)
  if (k < 0) k = picked.value.length - 1
  vIdx.value = k
  redraw()
  raf = requestAnimationFrame(loop)
}
function togglePreview() {
  previewing.value = !previewing.value
  t0 = 0
  if (previewing.value) raf = requestAnimationFrame(loop)
  else {
    cancelAnimationFrame(raf)
    pprog.value = 0
    vIdx.value = 0
    redraw()
  }
}
function stopPreview() {
  previewing.value = false
  cancelAnimationFrame(raf)
  pprog.value = 0
}
watch(kind, (k) => k === 'image' && stopPreview())
onBeforeUnmount(stopPreview)

// ── Hành động ──
const busy = ref('')
const err = ref('')
const toast = ref('')
let toastTimer = 0
function flash(t: string) {
  toast.value = t
  clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => (toast.value = ''), 3500)
}
async function run(name: string, fn: () => Promise<void>) {
  err.value = ''
  busy.value = name
  try {
    await fn()
  } catch (e) {
    err.value = errText(e)
  } finally {
    busy.value = ''
  }
}
async function imagePNG() {
  const c = document.createElement('canvas')
  drawCard(c, cardOpts(pickedText.value))
  return canvasPNG(c)
}
const saveImage = () => run('save', async () => {
  const { b64 } = await imagePNG()
  const path = await saveShareImage(b64, snap.value!.title)
  if (path) flash('Đã lưu ảnh: ' + path)
})
const copyImage = () => run('copy', async () => {
  const { b64, blob } = await imagePNG()
  await copyShareImage(b64, blob)
  flash('Đã chép ảnh — dán thẳng vào Facebook, Zalo (⌘V / Ctrl+V)')
})
const airImage = () => run('air', async () => {
  const { b64 } = await imagePNG()
  await airDropShareImage(b64, snap.value!.title)
})

type VState = 'idle' | 'busy' | 'done'
const vstate = ref<VState>('idle')
const progress = ref(0)
const offProgress = onEvent<number>('share:progress', (p) => (progress.value = p))
onBeforeUnmount(offProgress)
watch([kind, ratio, () => shareUI.bg, picked], () => vstate.value === 'done' && (vstate.value = 'idle'))
async function makeVideo() {
  const s = snap.value!
  err.value = ''
  vstate.value = 'busy'
  progress.value = 0
  stopPreview()
  try {
    const start = s.sentences[lo.value].start
    const frames: { png: string; sec: number }[] = []
    let wave = { x: 0, y: 0, w: 0, h: 0 }
    for (const i of picked.value) {
      const c = document.createElement('canvas')
      const r = drawCard(c, cardOpts([s.sentences[i].text]))
      if (r.wave) wave = r.wave
      frames.push({ png: (await canvasPNG(c)).b64, sec: s.sentences[i].end - s.sentences[i].start })
    }
    let wavePNG = ''
    if (bars.value.length && wave.w) {
      const c = document.createElement('canvas')
      drawBrightBars(c, ratio.value, wave, bars.value)
      wavePNG = (await canvasPNG(c)).b64
    }
    await makeShareVideo({
      slug: s.slug, file: s.file, start, end: start + dur.value, frames, wavePNG,
      waveX: wave.x, waveY: wave.y, waveW: wave.w, waveH: wave.h, title: s.title,
    })
    if (!shareUI.open) return
    vstate.value = 'done'
  } catch (e) {
    vstate.value = 'idle'
    if (shareUI.open) err.value = errText(e)
  }
  redraw()
}
const saveVideo = () => run('save', async () => {
  const path = await saveShareVideo()
  if (path) flash('Đã lưu video: ' + path)
})
const airVideo = () => run('air', () => airDropShareVideo())

function close() {
  if (vstate.value === 'busy') void cancelShareVideo()
  shareUI.open = false
}
</script>

<template>
  <div v-if="shareUI.open && snap" class="absolute inset-0 bg-background/70 backdrop-blur-sm grid place-items-center z-20" @mousedown.self="close">
    <div role="dialog" aria-modal="true" aria-labelledby="share-title" tabindex="-1"
      class="relative max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)] rounded-2xl border border-border bg-card text-card-foreground shadow-2xl flex overflow-hidden"
      :class="kind === 'video' ? 'w-[1060px] h-[690px]' : 'w-[980px] h-[684px]'">
      <!-- Trái: xem trước -->
      <div class="shrink-0 bg-muted/50 grid place-items-center p-6 relative" :class="kind === 'video' ? 'w-[540px] max-w-[48%]' : 'w-[460px]'">
        <canvas ref="canvas" class="rounded-2xl shadow-2xl object-contain max-h-[calc(100vh-8rem)]"
          :class="ratio === 'square' ? 'w-[360px] h-[360px]' : 'w-[280px] h-[498px]'" aria-label="Xem trước thẻ chia sẻ"></canvas>
        <button v-if="kind === 'video' && vstate !== 'busy'" class="absolute bottom-3 left-1/2 -translate-x-1/2 h-7 px-3 rounded-full bg-background border border-border text-xs flex items-center gap-1.5 shadow-sm hover:bg-muted" @click="togglePreview">
          <component :is="previewing ? Pause : Play" class="w-3 h-3" /> {{ previewing ? 'Dừng xem thử' : 'Xem thử (không tiếng)' }}
        </button>
      </div>

      <!-- Phải: tuỳ chọn -->
      <div class="flex-1 min-w-0 flex flex-col">
        <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
          <div class="min-w-0">
            <h2 id="share-title" class="font-semibold flex items-center gap-2">
              <template v-if="kind === 'video'"><Clapperboard class="w-4 h-4" /> Tạo video</template>
              <template v-else><ImageIcon class="w-4 h-4" /> Chia sẻ ảnh</template>
            </h2>
            <p class="text-xs text-muted-foreground mt-0.5 truncate">{{ kind === 'video' ? 'Đăng Reels, TikTok, Shorts' : 'Đăng Facebook, Zalo, Story' }} · {{ snap.title }}</p>
          </div>
          <button aria-label="Đóng" class="h-8 w-8 grid place-items-center rounded-md hover:bg-muted" @click="close"><X class="w-4 h-4" /></button>
        </div>

        <!-- Danh sách câu co giãn theo chỗ còn lại để Khung, Nền luôn hiện đủ (cửa sổ thấp vẫn thấy). -->
        <div class="flex-1 min-h-0 flex flex-col px-5 py-4 gap-4 text-sm overflow-auto">
          <VideoKindTabs v-if="kind === 'video'" class="shrink-0" current="short" :slug="snap.slug" :disabled="vstate === 'busy'" />

          <div class="flex-1 min-h-[104px] flex flex-col">
            <div class="flex items-baseline justify-between gap-2">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground truncate" :title="snap.section">Chọn câu · {{ snap.section }}</span>
              <span class="flex items-center gap-2 shrink-0">
                <button class="h-6 px-2 rounded-md border text-[11px] flex items-center gap-1" :class="pv.playing.value && pv.mode.value === 'span' ? 'border-primary text-primary bg-primary/5' : 'border-border hover:bg-muted'"
                  :disabled="!picked.length || vstate === 'busy'" @click="pv.playing.value && pv.mode.value === 'span' ? pv.stop() : pv.play(lo, hi)">
                  <component :is="pv.playing.value && pv.mode.value === 'span' ? Square : Play" class="w-3 h-3" /> Nghe đoạn đã chọn
                </button>
                <span class="text-xs text-muted-foreground tabular-nums">{{ picked.length }}/{{ MAX }} câu<template v-if="kind === 'video'"> · {{ Math.round(dur) }} giây</template></span>
              </span>
            </div>
            <div ref="listEl" class="mt-2 flex-1 min-h-0 max-h-[260px] overflow-auto rounded-lg border border-border divide-y divide-border">
              <div v-for="(s, i) in snap.sentences" :key="i" :data-i="i" role="checkbox" tabindex="0" :aria-checked="picked.includes(i)" :aria-disabled="vstate === 'busy'"
                class="group w-full flex items-start gap-2.5 px-3 py-2 text-left cursor-pointer outline-none focus-visible:bg-muted"
                :class="picked.includes(i) ? 'bg-primary/5' : 'hover:bg-muted/60'" @click="vstate !== 'busy' && toggle(i)" @keydown.enter.prevent="vstate !== 'busy' && toggle(i)" @keydown.space.prevent="vstate !== 'busy' && toggle(i)">
                <span class="mt-0.5 h-4 w-4 rounded border grid place-items-center shrink-0"
                  :class="picked.includes(i) ? 'bg-primary border-primary text-primary-foreground' : canAdd(i) ? 'border-primary/50' : 'border-border'">
                  <Check v-if="picked.includes(i)" class="w-3 h-3" />
                </span>
                <span class="flex-1 leading-snug" :class="pv.current.value === i ? 'text-primary font-medium' : picked.includes(i) ? 'text-foreground' : 'text-muted-foreground'">{{ s.text }}</span>
                <button class="h-6 w-6 -my-0.5 shrink-0 rounded-full grid place-items-center" :class="pv.current.value === i ? 'text-primary' : 'text-muted-foreground opacity-0 group-hover:opacity-100 focus:opacity-100 hover:bg-muted hover:text-foreground'"
                  :title="pv.current.value === i ? 'Dừng nghe thử' : 'Nghe từ câu này'" :aria-label="pv.current.value === i ? 'Dừng nghe thử' : 'Nghe từ câu này'" @click.stop="pv.current.value === i ? pv.stop() : pv.play(i)">
                  <Volume2 v-if="pv.current.value === i" class="w-3.5 h-3.5" /><Play v-else class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            <p class="mt-1.5 text-[11px] text-muted-foreground truncate" title="Chọn các câu liền nhau; bấm câu ở xa để chọn lại từ đầu">
              <template v-if="kind === 'video'">Câu liền nhau, tối đa {{ MAX }} câu / {{ MAX_VIDEO_SEC }} giây · video 15–30 giây dễ được xem hết nhất</template>
              <template v-else>Câu liền nhau, tối đa {{ MAX }} câu · bấm ▶ ở một câu để nghe tiếp từ đó</template>
            </p>
          </div>

          <div class="shrink-0">
            <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Khung</span>
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button v-for="r in (['story', 'square'] as const)" :key="r" :disabled="vstate === 'busy'" class="h-12 rounded-lg border flex items-center gap-3 px-3 text-left"
                :class="ratio === r ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="ratio = r">
                <span class="border-2 rounded-sm shrink-0" :class="[r === 'square' ? 'w-5 h-5' : 'w-4 h-7', ratio === r ? 'border-primary' : 'border-muted-foreground/50']"></span>
                <span><span class="block font-medium">{{ r === 'square' ? 'Vuông 1:1' : 'Dọc 9:16' }}</span><span class="block text-[11px] text-muted-foreground">{{ r === 'square' ? 'Bài đăng Facebook, Zalo' : 'Story, Reels, TikTok' }}</span></span>
              </button>
            </div>
          </div>

          <div class="shrink-0">
            <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Nền</span>
            <div class="mt-2 flex flex-wrap gap-2">
              <button v-for="(b, i) in BACKGROUNDS" :key="i" :disabled="vstate === 'busy'" class="flex flex-col items-center gap-1 text-[11px]" :class="shareUI.bg === i ? 'text-foreground font-medium' : 'text-muted-foreground'" @click="shareUI.bg = i">
                <span class="h-11 w-11 rounded-lg overflow-hidden ring-offset-2 ring-offset-card relative" :class="[shareUI.bg === i ? 'ring-2 ring-primary' : 'ring-1 ring-border', b.swatch]">
                  <img v-if="!b.stops.length && player.detail?.coverUrl" :src="player.detail.thumbUrl || player.detail.coverUrl" alt="" class="absolute -inset-2 w-[calc(100%+1rem)] h-[calc(100%+1rem)] max-w-none object-cover blur-md" />
                  <span v-else-if="!b.stops.length" class="absolute inset-0 bg-stone-700"></span>
                </span>
                {{ b.label }}
              </button>
            </div>
          </div>
        </div>

        <div class="shrink-0 border-t border-border px-5 py-3.5">
          <p v-if="err" class="mb-2 text-xs text-destructive line-clamp-2">{{ err }}</p>
          <div v-if="kind === 'image'" class="flex items-center gap-2">
            <Button class="flex-1" :disabled="!!busy" @click="saveImage"><Loader2 v-if="busy === 'save'" class="w-4 h-4 animate-spin" /><Download v-else class="w-4 h-4" /> Lưu ảnh</Button>
            <Button variant="outline" :disabled="!!busy" @click="copyImage"><Loader2 v-if="busy === 'copy'" class="w-4 h-4 animate-spin" /><Copy v-else class="w-4 h-4" /> Sao chép ảnh</Button>
            <Button v-if="airdrop" variant="outline" :disabled="!!busy" @click="airImage"><Share2 class="w-4 h-4" /> AirDrop</Button>
          </div>
          <template v-else>
            <div v-if="vstate === 'busy'" class="flex items-center gap-2">
              <div class="flex-1 h-9 rounded-md bg-muted flex items-center justify-center gap-2 text-sm relative overflow-hidden">
                <div class="absolute inset-y-0 left-0 bg-primary/15 transition-all" :style="{ width: progress + '%' }"></div>
                <Loader2 class="w-4 h-4 animate-spin relative" /><span class="relative tabular-nums">Đang tạo video {{ progress }}%</span>
              </div>
              <Button variant="outline" @click="close">Huỷ</Button>
            </div>
            <div v-else-if="vstate === 'done'" class="flex items-center gap-2">
              <Button class="flex-1" :disabled="!!busy" @click="saveVideo"><Loader2 v-if="busy === 'save'" class="w-4 h-4 animate-spin" /><Download v-else class="w-4 h-4" /> Lưu video</Button>
              <Button v-if="airdrop" variant="outline" :disabled="!!busy" @click="airVideo"><Share2 class="w-4 h-4" /> AirDrop</Button>
              <Button variant="ghost" @click="makeVideo">Tạo lại</Button>
            </div>
            <Button v-else class="w-full" @click="makeVideo"><Film class="w-4 h-4" /> Tạo video {{ Math.round(dur) }} giây</Button>
          </template>
          <p class="mt-2 text-[11px] text-muted-foreground text-center">
            {{ kind === 'image' ? 'Ảnh PNG sắc nét 1080 px. Facebook trên máy tính: bấm Sao chép ảnh rồi dán vào ô đăng bài.' : 'Video MP4 1080 px kèm tiếng đọc đúng các câu đã chọn. Tạo ngay trên máy, không cần mạng.' }}
          </p>
        </div>
      </div>

      <div v-if="toast" role="status" class="absolute bottom-4 left-4 max-w-[380px] rounded-full bg-foreground text-background text-xs px-3 py-1.5 shadow-lg flex items-center gap-1.5">
        <Check class="w-3.5 h-3.5 shrink-0" /><span class="truncate">{{ toast }}</span>
      </div>
    </div>
  </div>
</template>
