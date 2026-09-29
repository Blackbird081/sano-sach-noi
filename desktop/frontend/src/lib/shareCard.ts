// Vẽ thẻ chia sẻ đoạn hay (wireframe D14) bằng canvas: dùng chung cho bản xem
// trước và file xuất (PNG 1080 px, khung video) nên thấy gì lưu nấy. Toạ độ tính
// theo cỡ thẻ trong wireframe (vuông 380, dọc 300 px) rồi nhân lên 1080 px.
import logoUrl from '@/assets/favicon.svg'

export type ShareKind = 'image' | 'video'
export type ShareRatio = 'square' | 'story'

/** Nền: Hoàng hôn đầu tiên (mặc định); không có stops = theo bìa (bìa làm mờ, phủ tối). */
export const BACKGROUNDS: { label: string; stops: string[]; swatch: string }[] = [
  { label: 'Hoàng hôn', stops: ['#f97316', '#e11d48', '#6b21a8'], swatch: 'bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800' },
  { label: 'Theo bìa', stops: [], swatch: '' },
  { label: 'Biển', stops: ['#06b6d4', '#14b8a6', '#84cc16'], swatch: 'bg-gradient-to-br from-cyan-500 via-teal-500 to-lime-500' },
  { label: 'Đêm', stops: ['#4338ca', '#1e40af', '#a21caf'], swatch: 'bg-gradient-to-br from-indigo-700 via-blue-800 to-fuchsia-700' },
  { label: 'Trà', stops: ['#57534e', '#92400e', '#1c1917'], swatch: 'bg-gradient-to-br from-stone-600 via-amber-800 to-stone-900' },
]

// Màu bìa mặc định khi sách không có ảnh bìa (khớp BookCover.vue: cùng hàm băm tên sách).
const COVER_COLORS = ['#1e1b4b', '#292524', '#881337', '#92400e', '#1e293b', '#4c1d95', '#115e59', '#991b1b']
function coverColor(title: string) {
  let h = 2166136261
  for (const ch of title) h = Math.imul(h ^ (ch.codePointAt(0) ?? 0), 16777619) >>> 0
  return COVER_COLORS[h % COVER_COLORS.length]
}

export interface CardOptions {
  kind: ShareKind
  ratio: ShareRatio
  bg: number
  lines: string[] // ảnh: các câu đã chọn; video: câu của khung này
  title: string
  voice: string
  cover: HTMLImageElement | null
  /** Video: độ to từng cột sóng âm (0–1) của cả đoạn, tính từ tiếng đọc thật. */
  bars?: number[]
  /** Phần sóng âm đã đọc (0–1) tô sáng; file thật để 0, ffmpeg tô dần theo thời gian. */
  progress?: number
  /** Video: số dòng chữ dành chỗ (dòng nhiều nhất trong các câu) để bìa không nhảy giữa các câu. */
  reserveLines?: number
}

export interface CardResult {
  /** Khung sóng âm (px của ảnh) cho video; null với ảnh. */
  wave: { x: number; y: number; w: number; h: number } | null
}

const BASE = { square: { w: 380, h: 380 }, story: { w: 300, h: (300 * 16) / 9 } }
export const OUT_W = 1080

export function cardSize(ratio: ShareRatio) {
  const b = BASE[ratio]
  return { w: OUT_W, h: Math.round((b.h * OUT_W) / b.w) }
}

let logo: Promise<HTMLImageElement | null> | null = null
export function loadImage(url: string): Promise<HTMLImageElement | null> {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => resolve(null)
    img.src = url
  })
}
/** Chờ logo và font sẵn sàng trước lần vẽ đầu. */
export async function prepareCard(): Promise<HTMLImageElement | null> {
  logo ??= loadImage(logoUrl)
  try {
    await document.fonts?.ready
  } catch {
    // không có document.fonts: vẽ bằng font hệ thống
  }
  return logo
}
let logoImg: HTMLImageElement | null = null
/** Logo Sano đã nạp (sau prepareCard). */
export const cardLogo = () => logoImg
void prepareCard().then((l) => (logoImg = l))

export const SANS = () => getComputedStyle(document.body).fontFamily || 'system-ui, sans-serif'

export function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number) {
  ctx.beginPath()
  ctx.roundRect(x, y, w, h, r)
}

/** Ảnh vừa khít khung (object-fit: cover). */
export function drawCover(ctx: CanvasRenderingContext2D, img: HTMLImageElement | CanvasImageSource, iw: number, ih: number, x: number, y: number, w: number, h: number) {
  const s = Math.max(w / iw, h / ih)
  const sw = w / s
  const sh = h / s
  ctx.drawImage(img, (iw - sw) / 2, (ih - sh) / 2, sw, sh, x, y, w, h)
}

/** Bìa sách: ảnh bìa, hoặc màu bìa mặc định + tên sách. */
export function drawBook(ctx: CanvasRenderingContext2D, o: Pick<CardOptions, 'cover' | 'title'>, x: number, y: number, w: number, h: number, r: number) {
  ctx.save()
  roundRect(ctx, x, y, w, h, r)
  ctx.clip()
  if (o.cover) drawCover(ctx, o.cover, o.cover.naturalWidth, o.cover.naturalHeight, x, y, w, h)
  else {
    ctx.fillStyle = coverColor(o.title)
    ctx.fillRect(x, y, w, h)
    ctx.fillStyle = 'rgba(0,0,0,0.25)'
    ctx.fillRect(x, y, w * 0.07, h)
    ctx.fillStyle = '#fff'
    ctx.font = `600 ${Math.max(4, w * 0.13)}px Georgia, serif`
    const lines = wrap(ctx, o.title, w * 0.8).slice(0, 4)
    lines.forEach((l, i) => ctx.fillText(l, x + w * 0.12, y + h - w * 0.12 - (lines.length - 1 - i) * w * 0.15))
  }
  ctx.restore()
  ctx.save()
  roundRect(ctx, x, y, w, h, r)
  ctx.strokeStyle = 'rgba(255,255,255,0.2)'
  ctx.lineWidth = 1
  ctx.stroke()
  ctx.restore()
}

export function wrap(ctx: CanvasRenderingContext2D, text: string, maxW: number): string[] {
  const out: string[] = []
  let cur = ''
  for (const word of text.split(/\s+/).filter(Boolean)) {
    const next = cur ? cur + ' ' + word : word
    if (cur && ctx.measureText(next).width > maxW) {
      out.push(cur)
      cur = word
    } else cur = next
  }
  if (cur) out.push(cur)
  return out
}

/** Cắt còn `max` dòng, dòng cuối thêm "…". */
export function clampLines(ctx: CanvasRenderingContext2D, lines: string[], max: number, maxW: number) {
  if (lines.length <= max) return lines
  const keep = lines.slice(0, max)
  let last = keep[max - 1] + '…'
  while (last.length > 1 && ctx.measureText(last).width > maxW) last = last.slice(0, -2) + '…'
  keep[max - 1] = last
  return keep
}

export function drawBackground(ctx: CanvasRenderingContext2D, o: Pick<CardOptions, 'bg' | 'cover' | 'title'>, W: number, H: number) {
  const preset = BACKGROUNDS[o.bg]
  if (preset && preset.stops.length) {
    const g = ctx.createLinearGradient(0, 0, W, H)
    preset.stops.forEach((c, i) => g.addColorStop(i / (preset.stops.length - 1), c))
    ctx.fillStyle = g
    ctx.fillRect(0, 0, W, H)
    const d = ctx.createLinearGradient(0, 0, 0, H)
    d.addColorStop(0, 'rgba(0,0,0,0.10)')
    d.addColorStop(1, 'rgba(0,0,0,0.40)')
    ctx.fillStyle = d
    ctx.fillRect(0, 0, W, H)
    return
  }
  // Theo bìa: làm mờ bằng ctx.filter; WebView cũ không có filter thì thu bìa về
  // rất nhỏ rồi phóng to lại (mờ kém mịn hơn nhưng vẫn ra màu bìa).
  if (o.cover && typeof ctx.filter === 'string') {
    // Bìa phóng to, làm mờ mạnh (như nền màn nghe phóng to).
    ctx.save()
    ctx.filter = `blur(${Math.round(W / 14)}px) saturate(1.4)`
    const pad = W * 0.2
    drawCover(ctx, o.cover, o.cover.naturalWidth, o.cover.naturalHeight, -pad, -pad, W + pad * 2, H + pad * 2)
    ctx.restore()
  } else if (o.cover) {
    const tiny = document.createElement('canvas')
    tiny.width = 12
    tiny.height = 16
    const t = tiny.getContext('2d')!
    drawCover(t, o.cover, o.cover.naturalWidth, o.cover.naturalHeight, 0, 0, 12, 16)
    const mid = document.createElement('canvas')
    mid.width = 60
    mid.height = 80
    const m = mid.getContext('2d')!
    m.imageSmoothingQuality = 'high'
    m.drawImage(tiny, 0, 0, 60, 80)
    ctx.imageSmoothingQuality = 'high'
    const pad = W * 0.12
    drawCover(ctx, mid, 60, 80, -pad, -pad, W + pad * 2, H + pad * 2)
  } else {
    ctx.fillStyle = coverColor(o.title)
    ctx.fillRect(0, 0, W, H)
  }
  const d = ctx.createLinearGradient(0, 0, 0, H)
  d.addColorStop(0, 'rgba(0,0,0,0.35)')
  d.addColorStop(0.5, 'rgba(0,0,0,0.30)')
  d.addColorStop(1, 'rgba(0,0,0,0.60)')
  ctx.fillStyle = d
  ctx.fillRect(0, 0, W, H)
}

/** Vẽ thẻ lên canvas (đặt lại cỡ canvas = cỡ file xuất). */
export function drawCard(canvas: HTMLCanvasElement, o: CardOptions): CardResult {
  const { w: W, h: H } = cardSize(o.ratio)
  canvas.width = W
  canvas.height = H
  const ctx = canvas.getContext('2d')!
  const base = BASE[o.ratio]
  const k = W / base.w
  const sans = SANS()
  ctx.save()
  drawBackground(ctx, o, W, H)
  ctx.scale(k, k)
  const bw = base.w
  const bh = base.h
  const padX = o.ratio === 'square' ? 28 : 24
  const padY = o.ratio === 'square' ? 28 : 32
  const innerW = bw - padX * 2
  ctx.textBaseline = 'alphabetic'
  ctx.fillStyle = '#fff'
  if (o.kind === 'video' && o.ratio === 'story') {
    const r = drawReel(ctx, o, bw, k)
    ctx.restore()
    return r
  }

  // Dải thương hiệu dưới cùng (marketing lan truyền): logo + "Sano · Tự tạo sách nói" · sanobook.com,
  // vạch mảnh phía trên. Người xem biết ngay Sano để làm gì và tìm tới đâu.
  // Ảnh Story (Facebook / Instagram / Zalo): app phủ ~14% đầu (thanh tiến trình, tên người
  // đăng) và ~20% đáy (ô trả lời), nên khung dọc thu nội dung vào giữa.
  const storyImg = o.kind === 'image' && o.ratio === 'story'
  const safeTop = storyImg ? Math.round(bh * 0.14) : padY
  const brandRow = 24
  const brandBottom = storyImg ? bh * 0.78 : bh - padY
  const brandTop = brandBottom - brandRow
  ctx.fillStyle = 'rgba(255,255,255,0.2)'
  ctx.fillRect(padX, brandTop - 12 - 1, innerW, 1)
  if (logoImg) {
    ctx.save()
    ctx.shadowColor = 'rgba(0,0,0,0.3)'
    ctx.shadowBlur = 6
    ctx.drawImage(logoImg, padX, brandTop, brandRow, brandRow)
    ctx.restore()
  }
  const midY = brandTop + brandRow / 2 + 4.2
  ctx.fillStyle = '#fff'
  ctx.font = `700 12px ${sans}`
  const bx = padX + brandRow + 8
  ctx.fillText('Sano', bx, midY)
  const sanoW = ctx.measureText('Sano').width
  ctx.font = `400 12px ${sans}`
  ctx.fillStyle = 'rgba(255,255,255,0.8)'
  ctx.fillText(' · Tự tạo sách nói', bx + sanoW, midY)
  ctx.font = `600 11px ${sans}`
  ctx.fillStyle = 'rgba(255,255,255,0.9)'
  ctx.textAlign = 'right'
  ;(ctx as CanvasRenderingContext2D & { letterSpacing?: string }).letterSpacing = '0.3px'
  ctx.fillText('sanobook.com', bw - padX, midY)
  ;(ctx as CanvasRenderingContext2D & { letterSpacing?: string }).letterSpacing = '0px'
  ctx.textAlign = 'left'
  ctx.fillStyle = '#fff'

  // Chân thẻ: (bìa nhỏ) tên sách, giọng — nằm trên dải thương hiệu.
  const footBottom = brandTop - 12 - 1 - 12
  const hasMini = o.kind === 'image'
  const textX = padX + (hasMini ? 36 + 12 : 0)
  const textW = bw - padX - textX
  ctx.font = `600 13px ${sans}`
  const titleLines = clampLines(ctx, wrap(ctx, o.title, textW), 2, textW)
  const voiceH = o.voice ? 16 : 0
  const footTextH = titleLines.length * 16 + voiceH
  const footH = Math.max(hasMini ? 48 : 0, footTextH)
  const footTop = footBottom - footH
  if (hasMini) drawBook(ctx, o, padX, footBottom - 48, 36, 48, 4)
  let ty = footBottom - footTextH
  ctx.font = `600 13px ${sans}`
  for (const l of titleLines) {
    ctx.fillText(l, textX, ty + 12.5)
    ty += 16
  }
  if (o.voice) {
    ctx.font = `400 11px ${sans}`
    ctx.fillStyle = 'rgba(255,255,255,0.7)'
    ctx.fillText(clampLines(ctx, ['Giọng ' + o.voice], 1, textW)[0], textX, ty + 12)
    ctx.fillStyle = '#fff'
  }

  let wave: CardResult['wave'] = null
  if (o.kind === 'image') {
    // Dấu ngoặc kép + các câu, canh giữa theo chiều dọc vùng còn lại.
    ctx.font = `400 60px Georgia, serif`
    ctx.fillStyle = 'rgba(255,255,255,0.5)'
    ctx.fillText('“', padX - 2, safeTop + 44)
    ctx.fillStyle = '#fff'
    const top = safeTop + 44
    const bottom = footTop - 20
    let size = o.ratio === 'square' ? (o.lines.length > 2 ? 19 : 23) : 22
    let blocks: string[][] = []
    let total = 0
    for (; size >= 12; size--) {
      ctx.font = `700 ${size}px ${sans}`
      blocks = o.lines.map((l) => wrap(ctx, l, innerW))
      total = blocks.reduce((n, b) => n + b.length * size * 1.375, 0) + (blocks.length - 1) * 10
      if (total <= bottom - top) break
    }
    let y = top + Math.max(0, (bottom - top - total) / 2)
    for (const b of blocks) {
      for (const l of b) {
        ctx.fillText(l, padX, y + size * 1.05)
        y += size * 1.375
      }
      y += 10
    }
  } else {
    // Video: câu đang đọc ngay trên chân thẻ; phía trên là bìa lớn + sóng âm.
    const size = o.ratio === 'square' ? 19 : 21
    ctx.font = `700 ${size}px ${sans}`
    let lines = wrap(ctx, o.lines[0] ?? '', innerW)
    lines = clampLines(ctx, lines, o.ratio === 'square' ? 4 : 6, innerW)
    const lh = size * 1.375
    const blockH = Math.max(Math.max(lines.length, o.reserveLines ?? 0) * lh, size * 3.3)
    const textTop = footTop - 20 - blockH
    lines.forEach((l, i) => ctx.fillText(l, padX, textTop + i * lh + size * 1.05))

    const cw = o.ratio === 'square' ? 96 : 144
    const ch = (cw * 4) / 3
    const waveH = 32
    const waveW = o.ratio === 'square' ? 200 : 220
    const groupH = ch + 16 + waveH
    const areaTop = padY
    const areaBottom = textTop - 12
    const gy = areaTop + Math.max(0, (areaBottom - areaTop - groupH) / 2)
    const cx = (bw - cw) / 2
    ctx.save()
    ctx.shadowColor = 'rgba(0,0,0,0.5)'
    ctx.shadowBlur = 24
    ctx.shadowOffsetY = 10
    ctx.fillStyle = 'rgba(0,0,0,0.3)'
    roundRect(ctx, cx, gy, cw, ch, 8)
    ctx.fill()
    ctx.restore()
    drawBook(ctx, o, cx, gy, cw, ch, 8)
    const wy = gy + ch + 16
    const wx = (bw - waveW) / 2
    wave = { x: Math.round(wx * k), y: Math.round(wy * k), w: Math.round(waveW * k), h: Math.round(waveH * k) }
    if (o.bars?.length) {
      drawBars(ctx, wx, wy, waveW, waveH, o.bars, 'rgba(255,255,255,0.35)')
      if (o.progress) {
        ctx.save()
        ctx.beginPath()
        ctx.rect(wx, wy, waveW * o.progress, waveH)
        ctx.clip()
        drawBars(ctx, wx, wy, waveW, waveH, o.bars, 'rgba(255,255,255,0.95)')
        ctx.restore()
      }
    }
  }
  ctx.restore()
  return { wave }
}

/**
 * Video ngắn khung dọc (Reels / TikTok / Shorts): app phủ ~14% đầu, ~28% đáy (tên người đăng,
 * chú thích, ô bình luận) và cột nút bên phải, nên mọi thứ nằm trong REEL_SAFE: logo trên cùng,
 * rồi bìa nhỏ + tên sách, sóng âm, câu đang đọc.
 */
const REEL_SAFE = { top: 75, bottom: BASE.story.h * 0.72, left: 24, right: 54 }
const reelTextW = () => BASE.story.w - REEL_SAFE.left - REEL_SAFE.right

function drawReel(ctx: CanvasRenderingContext2D, o: CardOptions, bw: number, k: number): CardResult {
  const S = REEL_SAFE
  const sans = SANS()
  const x = S.left
  const tw = reelTextW()
  const sp = (v: string) => ((ctx as CanvasRenderingContext2D & { letterSpacing?: string }).letterSpacing = v)

  // Logo + "Sano · Tự tạo sách nói" · sanobook.com, vạch mảnh bên dưới
  const row = 24
  let y = S.top
  if (logoImg) {
    ctx.save()
    ctx.shadowColor = 'rgba(0,0,0,0.3)'
    ctx.shadowBlur = 6
    ctx.drawImage(logoImg, x, y, row, row)
    ctx.restore()
  }
  const midY = y + row / 2 + 4.2
  const bx = x + row + 8
  ctx.fillStyle = '#fff'
  ctx.font = `700 12px ${sans}`
  ctx.fillText('Sano', bx, midY)
  const sanoW = ctx.measureText('Sano').width
  ctx.font = `400 12px ${sans}`
  ctx.fillStyle = 'rgba(255,255,255,0.8)'
  ctx.fillText(' · Tự tạo sách nói', bx + sanoW, midY)
  ctx.font = `600 11px ${sans}`
  ctx.fillStyle = 'rgba(255,255,255,0.9)'
  ctx.textAlign = 'right'
  sp('0.3px')
  ctx.fillText('sanobook.com', bw - x, midY)
  sp('0px')
  ctx.textAlign = 'left'
  y += row + 12
  ctx.fillStyle = 'rgba(255,255,255,0.2)'
  ctx.fillRect(x, y, bw - x * 2, 1)
  const top = y + 1 + 16

  // Cỡ chữ câu đang đọc cố định cho cả video (theo câu dài nhất) để bố cục không nhảy.
  const cw = 54
  const ch = (cw * 4) / 3
  const waveH = 30
  const fixedH = ch + 16 + waveH + 18
  const avail = S.bottom - top - fixedH
  const reserve = Math.max(2, o.reserveLines ?? 0)
  let size = 21
  if (reserve * size * 1.375 > avail) size = Math.max(14, Math.floor(21 * Math.sqrt(avail / (reserve * 21 * 1.375))))
  const lh = size * 1.375
  const maxLines = Math.max(1, Math.floor(avail / lh))
  ctx.font = `700 ${size}px ${sans}`
  const lines = clampLines(ctx, wrap(ctx, o.lines[0] ?? '', tw), maxLines, tw)
  const blockH = Math.min(maxLines, Math.max(lines.length, reserve)) * lh
  // canh giữa cả nhóm theo chiều dọc vùng an toàn
  const gy = top + Math.max(0, (S.bottom - top - (fixedH + blockH)) / 2)

  // bìa nhỏ + tên sách, giọng
  ctx.save()
  ctx.shadowColor = 'rgba(0,0,0,0.45)'
  ctx.shadowBlur = 16
  ctx.shadowOffsetY = 6
  ctx.fillStyle = 'rgba(0,0,0,0.3)'
  roundRect(ctx, x, gy, cw, ch, 5)
  ctx.fill()
  ctx.restore()
  drawBook(ctx, o, x, gy, cw, ch, 5)
  const ix = x + cw + 12
  const iw = bw - ix - S.left
  ctx.font = `600 13px ${sans}`
  const titleLines = clampLines(ctx, wrap(ctx, o.title, iw), 2, iw)
  const infoH = titleLines.length * 16 + (o.voice ? 16 : 0)
  let ty = gy + (ch - infoH) / 2
  ctx.fillStyle = '#fff'
  for (const l of titleLines) {
    ctx.fillText(l, ix, ty + 12.5)
    ty += 16
  }
  if (o.voice) {
    ctx.font = `400 11px ${sans}`
    ctx.fillStyle = 'rgba(255,255,255,0.7)'
    ctx.fillText(clampLines(ctx, wrap(ctx, 'Giọng ' + o.voice, iw), 1, iw)[0], ix, ty + 12)
  }

  // sóng âm
  const wy = gy + ch + 16
  const waveW = tw
  const wave = { x: Math.round(x * k), y: Math.round(wy * k), w: Math.round(waveW * k), h: Math.round(waveH * k) }
  if (o.bars?.length) {
    drawBars(ctx, x, wy, waveW, waveH, o.bars, 'rgba(255,255,255,0.35)')
    if (o.progress) {
      ctx.save()
      ctx.beginPath()
      ctx.rect(x, wy, waveW * o.progress, waveH)
      ctx.clip()
      drawBars(ctx, x, wy, waveW, waveH, o.bars, 'rgba(255,255,255,0.95)')
      ctx.restore()
    }
  }

  // câu đang đọc
  const sy = wy + waveH + 18
  ctx.font = `700 ${size}px ${sans}`
  ctx.fillStyle = '#fff'
  lines.forEach((l, i) => ctx.fillText(l, x, sy + i * lh + size * 1.05))
  return { wave }
}

export function drawBars(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, bars: number[], color: string) {
  const step = w / bars.length
  const bw = Math.max(1, step * 0.55)
  ctx.fillStyle = color
  bars.forEach((v, i) => {
    const bh = Math.max(bw, h * v)
    roundRect(ctx, x + i * step + (step - bw) / 2, y + (h - bh) / 2, bw, bh, bw / 2)
    ctx.fill()
  })
}

/** Ảnh riêng các cột sóng âm sáng (nền trong suốt, đúng cỡ khung sóng âm) để ffmpeg tô dần. */
export function drawBrightBars(canvas: HTMLCanvasElement, ratio: ShareRatio, wave: { w: number; h: number }, bars: number[]) {
  const k = OUT_W / BASE[ratio].w
  canvas.width = wave.w
  canvas.height = wave.h
  const ctx = canvas.getContext('2d')!
  ctx.scale(k, k)
  drawBars(ctx, 0, 0, wave.w / k, wave.h / k, bars, 'rgba(255,255,255,0.95)')
}

/** Số dòng câu chiếm khi vẽ ở video (để dành chỗ cố định giữa các câu). */
export function videoLines(ratio: ShareRatio, text: string) {
  const ctx = document.createElement('canvas').getContext('2d')!
  const base = BASE[ratio]
  const size = ratio === 'square' ? 19 : 21
  ctx.font = `700 ${size}px ${SANS()}`
  if (ratio === 'story') return wrap(ctx, text, reelTextW()).length
  const padX = 28
  return Math.min(4, wrap(ctx, text, base.w - padX * 2).length)
}

/** Độ to từng cột sóng âm của đoạn [start, end) trong file tiếng (RMS, chuẩn hoá 0–1). */
const decoded = new Map<string, Promise<AudioBuffer>>()
function decode(url: string): Promise<AudioBuffer> {
  let p = decoded.get(url)
  if (!p) {
    p = (async () => {
      const buf = await (await fetch(url)).arrayBuffer()
      const AC = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
      const ac = new AC()
      try {
        return await ac.decodeAudioData(buf)
      } finally {
        void ac.close()
      }
    })()
    p.catch(() => decoded.delete(url))
    if (decoded.size > 4) decoded.delete(decoded.keys().next().value!)
    decoded.set(url, p)
  }
  return p
}
export async function audioBars(url: string, start: number, end: number, n: number): Promise<number[]> {
  {
    const audio = await decode(url)
    const data = audio.getChannelData(0)
    const a = Math.floor(start * audio.sampleRate)
    const b = Math.min(data.length, Math.floor(end * audio.sampleRate))
    const per = Math.max(1, Math.floor((b - a) / n))
    const out: number[] = []
    for (let i = 0; i < n; i++) {
      let sum = 0
      const s0 = a + i * per
      for (let j = s0; j < s0 + per && j < b; j++) sum += data[j] * data[j]
      out.push(Math.sqrt(sum / per))
    }
    const max = Math.max(...out) || 1
    return out.map((v) => Math.min(1, Math.max(0.08, Math.sqrt(v / max))))
  }
}

/** Canvas → PNG (base64 không tiền tố + Blob để chép bằng API trình duyệt). */
export async function canvasPNG(canvas: HTMLCanvasElement): Promise<{ b64: string; blob: Blob }> {
  const blob = await new Promise<Blob>((res, rej) => canvas.toBlob((b) => (b ? res(b) : rej(new Error('Không tạo được ảnh'))), 'image/png'))
  const b64 = canvas.toDataURL('image/png').split(',')[1]
  return { b64, blob }
}
