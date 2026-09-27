// Nhật ký nghe cho thống kê: đếm số giây thật đang phát + số giây nội dung đã nghe,
// gom lại rồi ghi xuống ~/Sano/.nghe.json (phần Go) mỗi ~30 giây, khi tạm dừng,
// đổi cuốn hoặc đóng app. Tua (nhảy thời gian) không tính là nghe.
import { markFinished, recordListening } from './backend'

const FLUSH_MS = 30_000
const MAX_GAP_S = 2 // timeupdate chạy ~4 lần/giây; cách xa hơn là máy ngủ / tab treo

let lastWall = 0
let lastAudio = -1
let lastPlace = '' // cuốn + tiểu mục: đổi tiểu mục thì không so thời gian
let lastFlush = Date.now()
const pending = new Map<string, { listen: number; audio: number }>()
const finished = new Set<string>()

/** Gọi ở mỗi timeupdate khi đang phát. */
export function listenTick(slug: string, track: number, audioTime: number, rate: number) {
  const now = Date.now()
  const place = `${slug}#${track}`
  if (place === lastPlace && lastWall && lastAudio >= 0) {
    const wall = (now - lastWall) / 1000
    const aud = audioTime - lastAudio
    if (wall > 0 && wall <= MAX_GAP_S && aud > 0 && aud <= wall * Math.max(rate, 1) + 1) {
      const p = pending.get(slug) ?? { listen: 0, audio: 0 }
      p.listen += wall
      p.audio += aud
      pending.set(slug, p)
    }
  }
  lastPlace = place
  lastWall = now
  lastAudio = audioTime
  if (now - lastFlush > FLUSH_MS) flushListening()
}

/** Tạm dừng / đổi tiểu mục / đổi cuốn: ghi phần đã gom, lần phát sau đếm lại từ đầu. */
export function listenStop() {
  lastWall = 0
  lastAudio = -1
  flushListening()
}

export function flushListening() {
  lastFlush = Date.now()
  for (const [slug, p] of pending) {
    if (p.listen >= 1 || p.audio >= 1) void recordListening(slug, p.listen, p.audio).catch(() => {})
  }
  pending.clear()
}

/** Nghe tới cuối cuốn: ghi ngày nghe xong (phần Go chỉ nhận lần đầu). */
export function listenFinished(slug: string) {
  if (!slug || finished.has(slug)) return
  finished.add(slug)
  void markFinished(slug).catch(() => {})
}
