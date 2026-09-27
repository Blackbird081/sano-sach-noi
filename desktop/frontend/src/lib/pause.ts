// Quãng nghỉ giữa các phần (wireframe D8, D9): mức chung ở Cài đặt, mức riêng
// từng cuốn chọn ở màn nghe. Dùng cho trình phát và khi Xuất M4B. Lưu trong bộ
// nhớ của cửa sổ app như tốc độ và vị trí nghe (mất thì về mặc định Vừa).
import { reactive } from 'vue'

export type PauseLevel = 'short' | 'medium' | 'long' | 'custom'

/** Số giây nghỉ trước tiểu mục kế (section) và trước chương mới (chapter). */
export interface Gaps {
  section: number
  chapter: number
}

export interface PauseSetting extends Gaps {
  level: PauseLevel
}

export const PAUSE_PRESETS: { level: Exclude<PauseLevel, 'custom'>; label: string; gaps: Gaps }[] = [
  { level: 'short', label: 'Ngắn', gaps: { section: 0.5, chapter: 1 } },
  { level: 'medium', label: 'Vừa', gaps: { section: 1.5, chapter: 2 } },
  { level: 'long', label: 'Dài', gaps: { section: 3, chapter: 4 } },
]
export const MAX_GAP_SEC = 10
const DEFAULT: PauseSetting = { level: 'medium', section: 1.5, chapter: 2 }

/**
 * Đọc số giây người dùng gõ ("1,2" hoặc "1.2"): hợp lệ khi là số từ 0 đến 10,
 * làm tròn 0,1. Sai → null.
 */
export function parseGapSec(raw: string | number): number | null {
  const s = String(raw).trim().replace(',', '.')
  if (!/^\d+(\.\d+)?$/.test(s)) return null
  const n = Number(s)
  if (!Number.isFinite(n) || n < 0 || n > MAX_GAP_SEC) return null
  return Math.round(n * 10) / 10
}

export const fmtGap = (n: number) => n.toLocaleString('vi-VN', { maximumFractionDigits: 1 }) + ' giây'
export const levelLabel = (l: PauseLevel) => PAUSE_PRESETS.find((p) => p.level === l)?.label ?? 'Tuỳ chỉnh'

/** Chuẩn hoá dữ liệu đọc từ bộ nhớ; hỏng → null. */
function clean(v: unknown): PauseSetting | null {
  if (!v || typeof v !== 'object') return null
  const o = v as Partial<PauseSetting>
  const preset = PAUSE_PRESETS.find((p) => p.level === o.level)
  if (preset) return { level: preset.level, ...preset.gaps }
  if (o.level !== 'custom') return null
  const section = parseGapSec(o.section ?? '')
  const chapter = parseGapSec(o.chapter ?? '')
  return section === null || chapter === null ? null : { level: 'custom', section, chapter }
}

function read(key: string): PauseSetting | null {
  try {
    const raw = localStorage.getItem(key)
    return raw ? clean(JSON.parse(raw)) : null
  } catch {
    return null
  }
}

function write(key: string, v: PauseSetting | null) {
  try {
    if (v) localStorage.setItem(key, JSON.stringify(v))
    else localStorage.removeItem(key)
  } catch {
    // bộ nhớ bị chặn: dùng được trong phiên này, lần sau về mặc định
  }
}

const GLOBAL_KEY = 'sano.pause'
const bookKey = (slug: string) => `sano:pause:${slug}`

export const pauseState = reactive({
  global: read(GLOBAL_KEY) ?? { ...DEFAULT },
  book: {} as Record<string, PauseSetting | null>, // đã đọc: mức riêng hoặc null (theo chung)
})

/** Tạo mức từ level (preset) hoặc số giây tuỳ chỉnh đã kiểm. */
export function makeSetting(level: PauseLevel, custom?: Gaps): PauseSetting {
  const preset = PAUSE_PRESETS.find((p) => p.level === level)
  if (preset) return { level: preset.level, ...preset.gaps }
  const section = parseGapSec(custom?.section ?? '')
  const chapter = parseGapSec(custom?.chapter ?? '')
  if (section === null || chapter === null) throw new Error(`Quãng nghỉ phải từ 0 đến ${MAX_GAP_SEC} giây`)
  return { level: 'custom', section, chapter }
}

export function setGlobalPause(v: PauseSetting) {
  pauseState.global = v
  write(GLOBAL_KEY, v)
}

/** Mức riêng của cuốn slug; null = theo cài đặt chung. */
export function bookPause(slug: string): PauseSetting | null {
  if (!(slug in pauseState.book)) pauseState.book[slug] = read(bookKey(slug))
  return pauseState.book[slug]
}

export function setBookPause(slug: string, v: PauseSetting | null) {
  pauseState.book[slug] = v
  write(bookKey(slug), v)
}

/** Quãng nghỉ dùng cho cuốn slug: mức riêng, không có thì mức chung. */
export function gapsFor(slug: string): Gaps {
  const v = (slug && bookPause(slug)) || pauseState.global
  return { section: v.section, chapter: v.chapter }
}
