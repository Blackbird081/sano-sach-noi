// Tính số liệu trang "Hành trình nghe" (wireframe D7) từ nhật ký nghe ~/Sano/.nghe.json.
// Hàm thuần: nhận nhật ký + ngày hôm nay, không đọc gì bên ngoài (dễ kiểm).
import type { ListenLog } from './backend'

export type StatsRange = 'week' | 'month' | 'year' | 'all'

/** Ngày nghe ít hơn mức này không tính vào chuỗi ngày nghe. */
export const STREAK_MIN_SEC = 60

export const dayKey = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
const addDays = (d: Date, n: number) => new Date(d.getFullYear(), d.getMonth(), d.getDate() + n)
const parseDay = (k: string) => {
  const [y, m, d] = k.split('-').map(Number)
  return new Date(y, m - 1, d)
}
/** Thứ hai của tuần chứa ngày d. */
const weekStart = (d: Date) => addDays(d, -((d.getDay() + 6) % 7))

export interface DaySum { listen: number; audio: number }

function daySum(log: ListenLog, key: string): DaySum {
  const day = log.days[key]
  let listen = 0
  let audio = 0
  for (const b of Object.values(day?.books ?? {})) {
    listen += b.listen || 0
    audio += b.audio || 0
  }
  return { listen, audio }
}

/** Ngày đầu tiên có số liệu, hoặc null. */
export function firstDay(log: ListenLog): string | null {
  const keys = Object.keys(log.days).filter((k) => daySum(log, k).listen > 0).sort()
  return keys[0] ?? null
}

/** Khoảng [from, to] (khoá ngày) của range tính tới hôm nay; kèm khoảng kỳ trước để so sánh. */
export function rangeBounds(range: StatsRange, today: Date, log: ListenLog) {
  switch (range) {
    case 'week': {
      const s = weekStart(today)
      return { from: dayKey(s), to: dayKey(addDays(s, 6)), prevFrom: dayKey(addDays(s, -7)), prevTo: dayKey(addDays(s, -1)) }
    }
    case 'month': {
      const s = new Date(today.getFullYear(), today.getMonth(), 1)
      const e = new Date(today.getFullYear(), today.getMonth() + 1, 0)
      const ps = new Date(today.getFullYear(), today.getMonth() - 1, 1)
      return { from: dayKey(s), to: dayKey(e), prevFrom: dayKey(ps), prevTo: dayKey(addDays(s, -1)) }
    }
    case 'year':
      return { from: `${today.getFullYear()}-01-01`, to: `${today.getFullYear()}-12-31`, prevFrom: `${today.getFullYear() - 1}-01-01`, prevTo: `${today.getFullYear() - 1}-12-31` }
    default:
      return { from: firstDay(log) ?? dayKey(today), to: dayKey(today), prevFrom: '', prevTo: '' }
  }
}

const inRange = (k: string, from: string, to: string) => k >= from && k <= to

/** Tổng giây nghe thật + nội dung trong khoảng. */
export function sumRange(log: ListenLog, from: string, to: string): DaySum {
  let listen = 0
  let audio = 0
  for (const k of Object.keys(log.days)) {
    if (!inRange(k, from, to)) continue
    const s = daySum(log, k)
    listen += s.listen
    audio += s.audio
  }
  return { listen, audio }
}

export interface Bar { label: string; title: string; min: number; today: boolean; future: boolean }

/** Cột biểu đồ: tuần/tháng mỗi ngày một cột; năm mỗi tháng; tất cả mỗi tháng từ tháng đầu có số liệu. */
export function bars(log: ListenLog, range: StatsRange, today: Date): Bar[] {
  const tk = dayKey(today)
  const WD = ['CN', 'T2', 'T3', 'T4', 'T5', 'T6', 'T7']
  if (range === 'week' || range === 'month') {
    const { from, to } = rangeBounds(range, today, log)
    const out: Bar[] = []
    for (let d = parseDay(from); dayKey(d) <= to; d = addDays(d, 1)) {
      const k = dayKey(d)
      const n = d.getDate()
      out.push({
        label: range === 'week' ? WD[d.getDay()] : n === 1 || n % 5 === 0 ? String(n) : '',
        title: `${WD[d.getDay()]} ${n}/${d.getMonth() + 1}`,
        min: daySum(log, k).listen / 60,
        today: k === tk,
        future: k > tk,
      })
    }
    return out
  }
  // theo tháng
  let start = new Date(today.getFullYear(), 0, 1)
  let end = new Date(today.getFullYear(), 11, 1)
  if (range === 'all') {
    const f = firstDay(log)
    start = f ? new Date(parseDay(f).getFullYear(), parseDay(f).getMonth(), 1) : new Date(today.getFullYear(), today.getMonth(), 1)
    end = new Date(today.getFullYear(), today.getMonth(), 1)
  }
  const out: Bar[] = []
  for (let m = start; m <= end; m = new Date(m.getFullYear(), m.getMonth() + 1, 1)) {
    const prefix = `${m.getFullYear()}-${String(m.getMonth() + 1).padStart(2, '0')}-`
    let sec = 0
    for (const k of Object.keys(log.days)) if (k.startsWith(prefix)) sec += daySum(log, k).listen
    const isNow = m.getFullYear() === today.getFullYear() && m.getMonth() === today.getMonth()
    const multiYear = range === 'all' && start.getFullYear() !== end.getFullYear()
    out.push({
      label: multiYear ? `${m.getMonth() + 1}/${String(m.getFullYear()).slice(2)}` : `T${m.getMonth() + 1}`,
      title: `Tháng ${m.getMonth() + 1}/${m.getFullYear()}`,
      min: sec / 60,
      today: isNow,
      future: m > new Date(today.getFullYear(), today.getMonth(), 1),
    })
  }
  return out
}

/** Chuỗi ngày nghe hiện tại (tính tới hôm nay, hoặc tới hôm qua nếu hôm nay chưa nghe) và kỷ lục. */
export function streaks(log: ListenLog, today: Date): { current: number; best: number } {
  const heard = (k: string) => daySum(log, k).listen >= STREAK_MIN_SEC
  let current = 0
  let d = heard(dayKey(today)) ? today : addDays(today, -1)
  while (heard(dayKey(d))) {
    current++
    d = addDays(d, -1)
  }
  const keys = Object.keys(log.days).filter(heard).sort()
  let best = 0
  let run = 0
  let prev = ''
  for (const k of keys) {
    run = prev && dayKey(addDays(parseDay(prev), 1)) === k ? run + 1 : 1
    best = Math.max(best, run)
    prev = k
  }
  return { current, best: Math.max(best, current) }
}

/** Phút nghe theo giờ trong ngày (24 ô) trong khoảng. */
export function hours(log: ListenLog, from: string, to: string): number[] {
  const out = Array<number>(24).fill(0)
  for (const [k, day] of Object.entries(log.days)) {
    if (!inRange(k, from, to)) continue
    ;(day.hours ?? []).forEach((s, h) => {
      if (h < 24) out[h] += (s || 0) / 60
    })
  }
  return out
}

/** Các giờ nghe nhiều nhất (tối đa 2 khung, cách nhau ≥ 3 giờ), để viết câu "Bạn hay nghe lúc…". */
export function peakHours(h: number[]): number[] {
  const order = h.map((m, i) => [m, i] as const).filter(([m]) => m > 0).sort((a, b) => b[0] - a[0])
  const picked: number[] = []
  for (const [, i] of order) {
    if (picked.every((p) => Math.abs(p - i) >= 3)) picked.push(i)
    if (picked.length === 2) break
  }
  return picked.sort((a, b) => a - b)
}

/** Giây nghe theo từng cuốn trong khoảng, nhiều nhất trước. */
export function topBooks(log: ListenLog, from: string, to: string, limit = 5): { slug: string; sec: number }[] {
  const m = new Map<string, number>()
  for (const [k, day] of Object.entries(log.days)) {
    if (!inRange(k, from, to)) continue
    for (const [slug, b] of Object.entries(day.books ?? {})) m.set(slug, (m.get(slug) ?? 0) + (b.listen || 0))
  }
  return [...m.entries()].filter(([, s]) => s >= 1).sort((a, b) => b[1] - a[1]).slice(0, limit).map(([slug, sec]) => ({ slug, sec }))
}

/** Cuốn nghe xong trong khoảng (mới nhất trước). */
export function finishedIn(log: ListenLog, from: string, to: string): { slug: string; day: string }[] {
  return Object.entries(log.finished ?? {})
    .filter(([, d]) => inRange(d, from, to))
    .sort((a, b) => (a[1] < b[1] ? 1 : -1))
    .map(([slug, day]) => ({ slug, day }))
}

/** Lịch nghe: `weeks` tuần tính tới tuần này, mỗi tuần 7 ô (T2→CN): mức 0..4, -1 = tương lai. */
export function heatmap(log: ListenLog, today: Date, weeks = 26): { level: number; key: string; min: number }[][] {
  const tk = dayKey(today)
  const start = addDays(weekStart(today), -7 * (weeks - 1))
  return Array.from({ length: weeks }, (_, w) =>
    Array.from({ length: 7 }, (_, d) => {
      const k = dayKey(addDays(start, w * 7 + d))
      const min = daySum(log, k).listen / 60
      const level = k > tk ? -1 : min <= 0 ? 0 : min < 15 ? 1 : min < 30 ? 2 : min < 60 ? 3 : 4
      return { level, key: k, min }
    }),
  )
}

/** Nhãn tháng trên lịch nghe: ghi "T<tháng>" ở cột có thứ Hai rơi vào ngày 1–7 (tuần đầu tháng). */
export function heatmapMonths(cols: { key: string }[][]): string[] {
  return cols.map((w) => (Number(w[0].key.slice(8)) <= 7 ? `T${Number(w[0].key.slice(5, 7))}` : ''))
}

/** "3 giờ 30 phút" / "45 phút" / "0 phút" từ số giây. */
export function fmtDur(sec: number): string {
  const m = Math.round(sec / 60)
  if (m < 60) return `${m} phút`
  const h = Math.floor(m / 60)
  return m % 60 ? `${h} giờ ${m % 60} phút` : `${h} giờ`
}

/** "12/7/2026" từ khoá ngày. */
export const fmtDay = (k: string) => {
  const d = parseDay(k)
  return `${d.getDate()}/${d.getMonth() + 1}/${d.getFullYear()}`
}

/** "Hôm nay" / "Hôm qua" / "21/9". */
export function fmtRecent(k: string, today: Date): string {
  if (k === dayKey(today)) return 'Hôm nay'
  if (k === dayKey(addDays(today, -1))) return 'Hôm qua'
  const d = parseDay(k)
  return d.getFullYear() === today.getFullYear() ? `${d.getDate()}/${d.getMonth() + 1}` : fmtDay(k)
}

// ── Mục tiêu mỗi ngày (phút; 0 = tắt), nhớ theo máy ──
const GOAL_KEY = 'sano.dailyGoal'
export function loadGoal(): number {
  try {
    const v = localStorage.getItem(GOAL_KEY)
    if (v === null) return 30
    const n = Number(v)
    return Number.isFinite(n) && n >= 0 && n <= 600 ? Math.round(n) : 30
  } catch {
    return 30
  }
}
export function saveGoal(min: number) {
  try {
    localStorage.setItem(GOAL_KEY, String(Math.max(0, Math.min(600, Math.round(min)))))
  } catch {
    // bộ nhớ trình duyệt bị chặn: chỉ mất tính năng nhớ mục tiêu
  }
}
