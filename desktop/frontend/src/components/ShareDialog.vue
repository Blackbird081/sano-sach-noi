<script setup lang="ts">
// Hộp "Chia sẻ đoạn hay" (wireframe D14). Trái: xem trước vẽ bằng canvas đúng như
// file sẽ lưu. Phải: Ảnh / Video, chọn câu liền nhau, khung, nền; lưu, chép, AirDrop.
// Chụp lại tiểu mục lúc mở (câu, tiếng, bìa) nên nghe tiếp sang mục khác không đổi thẻ.
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Check, Copy, Download, Film, Image as ImageIcon, Loader2, Play, Share2, Pause, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { shareUI } from '../lib/share'
import { BACKGROUNDS, audioBars, canvasPNG, drawBrightBars, drawCard, loadImage, prepareCard, videoLines, type ShareKind, type ShareRatio } from '../lib/shareCard'
import {
  airDropShareImage, airDropShareVideo, canAirDrop, cancelShareVideo, copyShareImage, errText, makeShareVideo, onEvent, saveShareImage, saveShareVideo,
} from '../lib/backend'
import { lyrics, player, track } from '../lib/player'

const MAX = 4 // số câu tối đa
const MAX_VIDEO_SEC = 30

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
  const ss = ly.sentences
  snap.value = {
    slug: d.slug, file: t.file, url: t.url, section: t.title, title: d.title, voice: d.voice,
    sentences: ss.map((s, i) => ({ text: s.text, start: s.start, end: i + 1 < ss.length ? ss[i + 1].start : t.durationSec })),
  }
  const i = Math.min(Math.max(0, shareUI.sentence), ss.length - 1)
  picked.value = [i]
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

const kind = computed({ get: () => shareUI.kind, set: (v: ShareKind) => (shareUI.kind = v) })
const ratio = computed({ get: () => shareUI.ratio, set: (v: ShareRatio) => (shareUI.ratio = v) })

const lo = computed(() => Math.min(...picked.value))
const hi = computed(() => Math.max(...picked.value))
const span = (a: number, b: number) => (snap.value ? snap.value.sentences[b].end - snap.value.sentences[a].start : 0)
const dur = computed(() => (picked.value.length ? Math.min(MAX_VIDEO_SEC, span(lo.value, hi.value)) : 0))
function canAdd(i: number) {
  const p = picked.value
  if (p.length >= MAX || !(i === lo.value - 1 || i === hi.value + 1)) return false
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
// Đổi sang video mà đoạn quá dài → giữ các câu đầu vừa 30 giây.
watch(kind, (k) => {
  if (k !== 'video' || !snap.value) return
  while (picked.value.length > 1 && span(lo.value, hi.value) > MAX_VIDEO_SEC) picked.value = picked.value.slice(0, -1)
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
      class="relative w-[880px] max-w-[calc(100%-2rem)] h-[600px] max-h-[calc(100%-2rem)] rounded-2xl border border-border bg-card text-card-foreground shadow-2xl flex overflow-hidden">
      <!-- Trái: xem trước -->
      <div class="w-[420px] shrink-0 bg-muted/50 grid place-items-center p-6 relative">
        <canvas ref="canvas" class="rounded-2xl shadow-2xl object-contain max-h-full"
          :class="ratio === 'square' ? 'w-[360px] h-[360px]' : 'w-[280px] h-[498px]'" aria-label="Xem trước thẻ chia sẻ"></canvas>
        <button v-if="kind === 'video' && vstate !== 'busy'" class="absolute bottom-3 left-1/2 -translate-x-1/2 h-7 px-3 rounded-full bg-background border border-border text-xs flex items-center gap-1.5 shadow-sm hover:bg-muted" @click="togglePreview">
          <component :is="previewing ? Pause : Play" class="w-3 h-3" /> {{ previewing ? 'Dừng xem thử' : 'Xem thử (không tiếng)' }}
        </button>
      </div>

      <!-- Phải: tuỳ chọn -->
      <div class="flex-1 min-w-0 flex flex-col">
        <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
          <h2 id="share-title" class="font-semibold flex items-center gap-2"><Share2 class="w-4 h-4" /> Chia sẻ đoạn hay</h2>
          <button aria-label="Đóng" class="h-8 w-8 grid place-items-center rounded-md hover:bg-muted" @click="close"><X class="w-4 h-4" /></button>
        </div>

        <div class="flex-1 min-h-0 overflow-auto px-5 py-4 space-y-4 text-sm">
          <div class="grid grid-cols-2 gap-2 p-1 rounded-lg bg-muted" role="tablist">
            <button v-for="k in (['image', 'video'] as const)" :key="k" role="tab" :aria-selected="kind === k" :disabled="vstate === 'busy'"
              class="h-9 rounded-md flex items-center justify-center gap-2 font-medium" :class="kind === k ? 'bg-background shadow-sm' : 'text-muted-foreground hover:text-foreground'" @click="kind = k">
              <component :is="k === 'image' ? ImageIcon : Film" class="w-4 h-4" /> {{ k === 'image' ? 'Ảnh có lời' : 'Video có tiếng đọc' }}
            </button>
          </div>

          <div>
            <div class="flex items-baseline justify-between gap-2">
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground truncate" :title="snap.section">Chọn câu · {{ snap.section }}</span>
              <span class="text-xs text-muted-foreground tabular-nums shrink-0">{{ picked.length }}/{{ MAX }} câu<template v-if="kind === 'video'"> · {{ Math.round(dur) }} giây</template></span>
            </div>
            <div ref="listEl" class="mt-2 max-h-[196px] overflow-auto rounded-lg border border-border divide-y divide-border">
              <button v-for="(s, i) in snap.sentences" :key="i" :data-i="i" :disabled="vstate === 'busy'" class="w-full flex items-start gap-2.5 px-3 py-2 text-left"
                :class="picked.includes(i) ? 'bg-primary/5' : 'hover:bg-muted/60'" @click="toggle(i)">
                <span class="mt-0.5 h-4 w-4 rounded border grid place-items-center shrink-0"
                  :class="picked.includes(i) ? 'bg-primary border-primary text-primary-foreground' : canAdd(i) ? 'border-primary/50' : 'border-border'">
                  <Check v-if="picked.includes(i)" class="w-3 h-3" />
                </span>
                <span class="leading-snug" :class="picked.includes(i) ? 'text-foreground' : 'text-muted-foreground'">{{ s.text }}</span>
              </button>
            </div>
            <p class="mt-1.5 text-[11px] text-muted-foreground">Chọn các câu liền nhau, tối đa {{ MAX }} câu<template v-if="kind === 'video'">, {{ MAX_VIDEO_SEC }} giây</template>. Bấm câu ở xa để chọn lại từ đầu.</p>
          </div>

          <div>
            <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Khung</span>
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button v-for="r in (['square', 'story'] as const)" :key="r" :disabled="vstate === 'busy'" class="h-14 rounded-lg border flex items-center gap-3 px-3 text-left"
                :class="ratio === r ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="ratio = r">
                <span class="border-2 rounded-sm shrink-0" :class="[r === 'square' ? 'w-5 h-5' : 'w-4 h-7', ratio === r ? 'border-primary' : 'border-muted-foreground/50']"></span>
                <span><span class="block font-medium">{{ r === 'square' ? 'Vuông 1:1' : 'Dọc 9:16' }}</span><span class="block text-[11px] text-muted-foreground">{{ r === 'square' ? 'Bài đăng Facebook, Zalo' : 'Story, Reels, TikTok' }}</span></span>
              </button>
            </div>
          </div>

          <div>
            <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Nền</span>
            <div class="mt-2 flex flex-wrap gap-2">
              <button v-for="(b, i) in BACKGROUNDS" :key="i" :disabled="vstate === 'busy'" class="flex flex-col items-center gap-1 text-[11px]" :class="shareUI.bg === i ? 'text-foreground font-medium' : 'text-muted-foreground'" @click="shareUI.bg = i">
                <span class="h-11 w-11 rounded-lg overflow-hidden ring-offset-2 ring-offset-card relative" :class="[shareUI.bg === i ? 'ring-2 ring-primary' : 'ring-1 ring-border', b.swatch]">
                  <img v-if="i === 0 && player.detail?.coverUrl" :src="player.detail.thumbUrl || player.detail.coverUrl" alt="" class="absolute -inset-2 w-[calc(100%+1rem)] h-[calc(100%+1rem)] max-w-none object-cover blur-md" />
                  <span v-else-if="i === 0" class="absolute inset-0 bg-stone-700"></span>
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
