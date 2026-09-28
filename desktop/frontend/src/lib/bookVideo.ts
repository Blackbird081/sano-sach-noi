// Tạo video cả cuốn (wireframe D15): dựng dòng thời gian (trích đoạn → màn tựa → các
// tiểu mục, thẻ chương, quãng nghỉ như khi nghe → màn kết), tính giờ từng câu (ước lượng
// rồi bám khoảng lặng thật như màn nghe), vẽ từng khung hình gửi sang Go, kèm thumbnail,
// phụ đề .srt, mô tả YouTube có mốc chương. Chạy ở module (không ở hộp thoại) nên đóng
// hộp thoại giữa chừng vẫn tạo tiếp.
import { reactive } from 'vue'
import {
  bookTexts, bookVideoBegin, bookVideoCancel, bookVideoFinish, bookVideoFrame, bookVideoState, errText, onEvent,
  type BookDetail, type BookVideoOverlay, type BookVideoSeg, type BookVideoStatus, type Track,
} from './backend'
import { buildLyrics, findSilences, snapToSilences, type Lyrics } from './lyrics'
import { gapsFor } from './pause'
import { audioBars, canvasPNG, loadImage, prepareCard } from './shareCard'
import { drawBrightBar, drawBrightWave, drawScene, introLines, sceneSize, type BVRatio, type Rect, type Scene, type SceneCommon } from './bookVideoCard'

export const TITLE_SEC = 3
export const CHAPTER_SEC = 2.5
export const END_SEC = 5
const MAX_DECODE_SEC = 20 * 60

export interface BookVideoOptions {
  from: number // chỉ số tiểu mục đầu (gồm)
  to: number // chỉ số tiểu mục cuối (gồm)
  intro: { track: number; from: number; to: number } | null // câu from..to (gồm) của tiểu mục
  ratio: BVRatio
  bg: number
  extras: { thumb: boolean; srt: boolean; desc: boolean }
}

type Step = 'idle' | 'prep' | 'frames' | 'audio' | 'video' | 'files' | 'done' | 'error'
export const bv = reactive({
  open: false,
  slug: '',
  step: 'idle' as Step,
  done: 0, // tiểu mục đã dò / khung hình đã vẽ
  total: 0,
  pct: 0,
  error: '',
  dir: '',
  bytes: 0,
  durSec: 0,
  desc: '',
  title: '',
  video: '', // đường dẫn file video vừa tạo
})
export const bvBusy = () => ['prep', 'frames', 'audio', 'video', 'files'].includes(bv.step)

function applyStatus(st: BookVideoStatus) {
  if (!st) return
  bv.slug = st.slug
  bv.title = st.title
  if (st.running) {
    bv.step = st.phase
    bv.pct = st.pct
  } else if (st.done) {
    bv.step = 'done'
    bv.pct = 100
    bv.dir = st.dir
    bv.video = st.video
    bv.bytes = st.bytes
    bv.durSec = st.durSec
  } else if (st.error) {
    bv.step = st.error === 'đã huỷ' ? 'idle' : 'error'
    bv.error = st.error === 'đã huỷ' ? '' : st.error
  }
}
onEvent<BookVideoStatus>('bookvideo:progress', (st) => st.phase !== 'frames' && applyStatus(st))
onEvent<BookVideoStatus>('bookvideo:finished', applyStatus)

export function openBookVideo(slug: string) {
  if (bv.slug !== slug && !bvBusy()) {
    bv.step = 'idle'
    bv.error = ''
  }
  bv.slug = bvBusy() ? bv.slug : slug
  bv.open = true
  void bookVideoState().then((st) => st && st.slug === bv.slug && (st.running || bv.step === 'idle') && st.phase !== 'frames' && applyStatus(st))
}

let cancelled = false
export async function cancelBookVideo() {
  cancelled = true
  await bookVideoCancel()
  if (bv.step === 'prep' || bv.step === 'frames') bv.step = 'idle'
}

// ── Lời từng tiểu mục (giờ câu bám khoảng lặng thật như màn nghe) ──
async function trackLyrics(t: Track, text: string, script: string): Promise<Lyrics | null> {
  if (!text) return null
  const l = buildLyrics(text, script, t.durationSec)
  if (t.durationSec > MAX_DECODE_SEC) return l
  try {
    const buf = await (await fetch(t.url)).arrayBuffer()
    const Ctx = window.OfflineAudioContext || (window as unknown as { webkitOfflineAudioContext: typeof OfflineAudioContext }).webkitOfflineAudioContext
    const a = await new Ctx(1, 1, 16000).decodeAudioData(buf)
    return snapToSilences(l, findSilences(a.getChannelData(0), a.sampleRate))
  } catch {
    return l
  }
}

/** Chương mới bắt đầu ở tiểu mục i (hiện thẻ chương / mốc mô tả)? */
export function chapterAt(d: BookDetail, i: number, from: number): string {
  const t = d.tracks[i]
  if (!t?.chapter || t.chapter === d.title) return ''
  return t.chapterStart || i === from ? t.chapter : ''
}

export const clock = (sec: number) => {
  const s = Math.max(0, Math.floor(sec))
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = String(s % 60).padStart(2, '0')
  return h ? `${h}:${String(m).padStart(2, '0')}:${ss}` : `${m}:${ss}`
}
function srtTime(sec: number) {
  const ms = Math.max(0, Math.round(sec * 1000))
  const h = Math.floor(ms / 3600000)
  const m = Math.floor((ms % 3600000) / 60000)
  const s = Math.floor((ms % 60000) / 1000)
  const p = (n: number, w = 2) => String(n).padStart(w, '0')
  return `${p(h)}:${p(m)}:${p(s)},${p(ms % 1000, 3)}`
}

interface Frame {
  scene: Scene
  sec: number
}

/** Chạy cả lượt: dò lời → vẽ khung hình → Go ghép + mã hoá (tiến độ qua sự kiện). */
export async function startBookVideo(d: BookDetail, o: BookVideoOptions) {
  cancelled = false
  bv.slug = d.slug
  bv.title = d.title
  bv.error = ''
  bv.step = 'prep'
  bv.done = 0
  bv.pct = 0
  const range = d.tracks.slice(o.from, o.to + 1)
  bv.total = range.length + (o.intro ? 1 : 0)
  try {
    const texts = await bookTexts(d.slug)
    await prepareCard()
    const cover = d.coverUrl ? await loadImage(d.coverUrl) : null
    const common: SceneCommon = { ratio: o.ratio, bg: o.bg, cover, title: d.title, voice: d.voice, totalLabel: '' }

    // 1. Lời + giờ câu của từng tiểu mục trong phần chọn.
    const lyr = new Map<number, Lyrics | null>()
    const need = [...new Set([...(o.intro ? [o.intro.track] : []), ...range.map((_, k) => o.from + k)])]
    for (const i of need) {
      if (cancelled) return
      lyr.set(i, await trackLyrics(d.tracks[i], texts[i]?.text ?? '', texts[i]?.script ?? ''))
      bv.done++
    }

    // 2. Dòng thời gian.
    const segs: BookVideoSeg[] = []
    const frames: Frame[] = []
    const srt: { a: number; b: number; t: string }[] = []
    const marks: { t: number; label: string }[] = []
    let t = 0
    const silence = (dur: number) => segs.push({ file: '', start: 0, dur, silence: true })
    let introDur = 0
    let introBars: number[] = []
    if (o.intro) {
      const it = d.tracks[o.intro.track]
      const ss = lyr.get(o.intro.track)?.sentences ?? []
      const a = ss[o.intro.from]?.start ?? 0
      const endAt = (i: number) => (i + 1 < ss.length ? ss[i + 1].start : it.durationSec)
      introDur = Math.max(1, endAt(o.intro.to) - a)
      introBars = await audioBars(it.url, a, a + introDur, 48).catch(() => Array.from({ length: 48 }, () => 0.35))
      const pick = ss.slice(o.intro.from, o.intro.to + 1)
      const reserve = Math.max(1, ...pick.map((s) => introLines(o.ratio, s.text)))
      segs.push({ file: it.file, start: a, dur: introDur, silence: false })
      pick.forEach((s, k) => {
        const sec = endAt(o.intro!.from + k) - s.start
        frames.push({ scene: { kind: 'intro', line: s.text, bars: introBars, reserve }, sec })
        srt.push({ a: t, b: t + sec, t: s.text })
        t += sec
      })
      t = introDur
      marks.push({ t: 0, label: 'Trích đoạn' })
    }
    const bookSec = range.reduce((n, x) => n + x.durationSec, 0)
    const minutes = Math.max(1, Math.round(bookSec / 60))
    const chapters = range.filter((_, k) => chapterAt(d, o.from + k, o.from)).length
    common.totalLabel = clock(bookSec)
    const meta = `Giọng ${d.voice} · ${minutes} phút${chapters ? ` · ${chapters} chương` : ''}`
    // Vài chương: ghi "Nghe thử", tên file / thư mục kèm tên chương để không đè video cả cuốn.
    const partial = o.from > 0 || o.to < d.tracks.length - 1
    const badge = partial ? 'NGHE THỬ SÁCH NÓI' : 'SÁCH NÓI ĐẦY ĐỦ'
    const names = range.map((_, k) => chapterAt(d, o.from + k, o.from)).filter(Boolean)
    const part = names.length > 1 ? `${names[0]} đến ${names[names.length - 1]}` : names[0] || range[0].title
    const outName = partial ? `${d.title} - ${part}` : d.title
    frames.push({ scene: { kind: 'title', meta, badge }, sec: TITLE_SEC })
    silence(TITLE_SEC)
    t += TITLE_SEC
    const bookStart = t

    // vạch chương: tính trước theo cùng cách cộng thời gian
    const gaps = gapsFor(d.slug)
    const ticks: number[] = []
    const plan: { i: number; chapter: string; gap: number }[] = []
    {
      let x = 0
      range.forEach((tr, k) => {
        const i = o.from + k
        const ch = chapterAt(d, i, o.from)
        if (ch) {
          ticks.push(x)
          x += CHAPTER_SEC
        }
        x += tr.durationSec
        const nextCh = k + 1 < range.length && !!chapterAt(d, i + 1, o.from)
        const gap = k + 1 < range.length ? (nextCh ? gaps.chapter : gaps.section) : 0
        x += gap
        plan.push({ i, chapter: ch, gap })
      })
      const total = x
      for (let n = 0; n < ticks.length; n++) ticks[n] = total ? ticks[n] / total : 0
    }
    let curChapter = ''
    for (const { i, chapter, gap } of plan) {
      const tr = d.tracks[i]
      if (chapter) {
        curChapter = chapter
        frames.push({ scene: { kind: 'chapter', chapter, ticks }, sec: CHAPTER_SEC })
        silence(CHAPTER_SEC)
        marks.push({ t, label: chapter })
        t += CHAPTER_SEC
      } else if (i === o.from && !marks.length) marks.push({ t, label: tr.title })
      segs.push({ file: tr.file, start: 0, dur: tr.durationSec, silence: false })
      const ss = lyr.get(i)?.sentences ?? []
      const chLabel = curChapter || tr.chapter || d.title
      if (!ss.length) frames.push({ scene: { kind: 'main', chapter: chLabel, section: tr.title, line: tr.title, next: '', ticks }, sec: tr.durationSec + gap })
      ss.forEach((s, k) => {
        const a = k === 0 ? 0 : s.start
        const b = k + 1 < ss.length ? ss[k + 1].start : tr.durationSec
        frames.push({ scene: { kind: 'main', chapter: chLabel, section: tr.title, line: s.text, next: ss[k + 1]?.text ?? '', ticks }, sec: Math.max(0.05, b - a) + (k === ss.length - 1 ? gap : 0) })
        srt.push({ a: t + s.start, b: t + b, t: s.text })
      })
      t += tr.durationSec
      if (gap) {
        silence(gap)
        t += gap
      }
    }
    const bookEnd = t
    frames.push({ scene: { kind: 'end' }, sec: END_SEC })
    silence(END_SEC)
    t += END_SEC

    // 3. Vẽ + gửi khung hình.
    if (cancelled) return
    const id = await bookVideoBegin(d.slug, d.title)
    bv.step = 'frames'
    bv.done = 0
    bv.total = frames.length
    const c = document.createElement('canvas')
    let wave: Rect | null = null
    let bar: Rect | null = null
    for (let n = 0; n < frames.length; n++) {
      if (cancelled) return
      const r = drawScene(c, common, frames[n].scene)
      wave ??= r.wave
      bar ??= r.bar
      await bookVideoFrame(id, n, c.toDataURL('image/jpeg', 0.9).split(',')[1])
      bv.done = n + 1
      if (n % 8 === 0) await new Promise((res) => setTimeout(res)) // nhường giao diện
    }
    const overlays: BookVideoOverlay[] = []
    if (wave && introDur) {
      drawBrightWave(c, wave, introBars)
      overlays.push({ png: (await canvasPNG(c)).b64, ...wave, t0: 0, t1: introDur })
    }
    if (bar && bookEnd > bookStart) {
      drawBrightBar(c, bar)
      overlays.push({ png: (await canvasPNG(c)).b64, ...bar, t0: bookStart, t1: bookEnd })
    }
    let thumb = ''
    if (o.extras.thumb) {
      drawScene(c, common, { kind: 'thumb', meta: `${minutes} phút · Giọng ${d.voice}`, badge })
      thumb = (await canvasPNG(c)).b64
    }
    const desc = [
      partial ? `${d.title} — Nghe thử sách nói (${part}), giọng ${d.voice}.` : `${d.title} — Sách nói đầy đủ, giọng ${d.voice}.`,
      '',
      ...marks.map((m, k) => `${clock(k === 0 ? 0 : m.t)} ${m.label}`),
      '',
      'Sách nói tạo bằng Sano — tự tạo sách nói miễn phí từ file Word: https://sanobook.com',
    ].join('\n')
    bv.desc = o.extras.desc ? desc : ''
    bv.title = outName
    const { w, h } = sceneSize(o.ratio, 'main')
    if (cancelled) return
    await bookVideoFinish(id, {
      frames: frames.map((f) => f.sec), segs, overlays, width: w, height: h, title: outName, thumb,
      srt: o.extras.srt ? srt.map((s, k) => `${k + 1}\n${srtTime(s.a)} --> ${srtTime(s.b)}\n${s.t}\n`).join('\n') : '',
      desc: o.extras.desc ? desc + '\n' : '',
    })
    bv.step = 'audio'
  } catch (e) {
    if (cancelled) return
    bv.step = 'error'
    bv.error = errText(e)
  }
}
