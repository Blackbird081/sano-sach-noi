// Nhớ vị trí nghe từng cuốn (theo máy, trong bộ nhớ của cửa sổ app).
export interface Position {
  track: number
  time: number
  pct: number // % cả cuốn đã nghe, để hiện ở Thư viện
  at?: number // lúc nghe gần nhất (ms), để sắp "Nghe gần đây" / hàng "Nghe tiếp"
}

const key = (slug: string) => `sano:pos:${slug}`

export function loadPosition(slug: string): Position | null {
  try {
    const raw = localStorage.getItem(key(slug))
    if (!raw) return null
    const p = JSON.parse(raw) as Partial<Position>
    return { track: Number(p.track) || 0, time: Number(p.time) || 0, pct: Number(p.pct) || 0, at: Number(p.at) || 0 }
  } catch {
    return null
  }
}

export function savePosition(slug: string, p: Position) {
  try {
    localStorage.setItem(key(slug), JSON.stringify({ ...p, at: p.at || Date.now() }))
  } catch {
    // bộ nhớ trình duyệt bị chặn: bỏ qua, chỉ mất tính năng nhớ vị trí
  }
}

// Các tiểu mục đã nghe thật (không tính nhảy qua) → dấu ✓ ở mục lục.
const heardKey = (slug: string) => `sano:heard:${slug}`

/** Tiểu mục đã nghe của một cuốn; null = chưa từng ghi (sách nghe từ bản cũ). */
export function loadHeard(slug: string): number[] | null {
  try {
    const raw = localStorage.getItem(heardKey(slug))
    if (raw === null) return null
    const v = JSON.parse(raw)
    return Array.isArray(v) ? v.filter((n) => Number.isInteger(n) && n >= 0) : []
  } catch {
    return null
  }
}

export function saveHeard(slug: string, heard: number[]) {
  try {
    localStorage.setItem(heardKey(slug), JSON.stringify([...new Set(heard)].sort((a, b) => a - b)))
  } catch {
    // bộ nhớ trình duyệt bị chặn: chỉ mất dấu đã nghe
  }
}

const histKey = (slug: string) => `sano:hist:${slug}`

/** Một lượt nghe đã xoá khỏi màn hình, cất lại cho thống kê sau này. */
export interface ClearedPosition extends Position {
  clearedAt: number
}

/** Lịch sử nghe đã cất của một cuốn (cũ trước, mới sau). */
export function loadHistory(slug: string): ClearedPosition[] {
  try {
    const v = JSON.parse(localStorage.getItem(histKey(slug)) ?? '[]')
    return Array.isArray(v) ? v : []
  } catch {
    return []
  }
}

/** Xoá lịch sử nghe khỏi màn hình: mọi cuốn về "Chưa nghe", nghe lại từ đầu, bỏ dấu ✓. Vị trí
 *  cũ (tiến độ, lúc nghe) KHÔNG mất mà cất sang sano:hist:<slug> để làm thống kê.
 *  Trả số cuốn đã xoá. */
export function clearPositions(): number {
  try {
    const now = Date.now()
    const keys = Object.keys(localStorage).filter((k) => k.startsWith('sano:pos:'))
    for (const k of keys) {
      const slug = k.slice('sano:pos:'.length)
      const pos = loadPosition(slug)
      if (pos) localStorage.setItem(histKey(slug), JSON.stringify([...loadHistory(slug), { ...pos, clearedAt: now }]))
      localStorage.removeItem(k)
    }
    for (const k of Object.keys(localStorage).filter((k) => k.startsWith('sano:heard:'))) localStorage.removeItem(k)
    return keys.length
  } catch {
    return 0
  }
}

/** "1 giờ 12 phút" / "58 phút" / "40 giây". */
export function fmtLong(sec: number) {
  if (sec < 60) return `${Math.max(0, Math.round(sec))} giây`
  const m = Math.round(sec / 60)
  if (m < 60) return `${m} phút`
  return `${Math.floor(m / 60)} giờ ${String(m % 60).padStart(2, '0')} phút`
}

/** "3:05" / "1:02:07". */
export function fmtClock(sec: number) {
  const s = Math.max(0, Math.floor(sec))
  const h = Math.floor(s / 3600)
  const mm = Math.floor((s % 3600) / 60)
  const ss = String(s % 60).padStart(2, '0')
  return h ? `${h}:${String(mm).padStart(2, '0')}:${ss}` : `${mm}:${ss}`
}
