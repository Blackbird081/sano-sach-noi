// Nghe thử trong hộp chọn câu (Chia sẻ đoạn hay, Tạo video cả cuốn): phát tiếng tiểu mục
// bằng một thẻ audio riêng (không đụng trình nghe chính, chỉ tạm dừng nó), từ câu bấm chạy
// tiếp — để dò đoạn hay — hoặc đúng đoạn đã chọn rồi dừng. Báo câu đang đọc để tô sáng.
import { onBeforeUnmount, ref } from 'vue'
import { pause as pauseMain, player } from './player'

export interface PreviewSentence {
  start: number
  end: number
}

export function usePreview(get: () => { url: string; sentences: PreviewSentence[] } | null) {
  const audio = new Audio()
  audio.preload = 'auto'
  const playing = ref(false)
  const current = ref(-1) // câu đang đọc (-1 = không phát)
  const mode = ref<'from' | 'span'>('from')
  let stopAt = Infinity
  let onDone: (() => void) | null = null // gọi khi phát hết đoạn (không gọi khi bấm dừng)
  let raf = 0

  function tick() {
    const src = get()
    if (!playing.value || !src) return
    const t = audio.currentTime
    if (t >= stopAt) {
      const cb = onDone
      stop()
      cb?.()
      return
    }
    let i = src.sentences.findIndex((s) => t < s.end)
    if (i < 0) i = src.sentences.length - 1
    current.value = i
    raf = requestAnimationFrame(tick)
  }

  async function play(from: number, to?: number, done?: () => void) {
    const src = get()
    if (!src?.sentences[from]) return
    if (player.playing) pauseMain()
    mode.value = to === undefined ? 'from' : 'span'
    onDone = done ?? null
    stopAt = to === undefined ? Infinity : src.sentences[to].end
    if (!audio.src.endsWith(src.url)) audio.src = src.url
    audio.currentTime = src.sentences[from].start
    current.value = from
    try {
      await audio.play()
      playing.value = true
      cancelAnimationFrame(raf)
      raf = requestAnimationFrame(tick)
    } catch {
      playing.value = false
      current.value = -1
    }
  }

  function stop() {
    onDone = null
    audio.pause()
    playing.value = false
    current.value = -1
    cancelAnimationFrame(raf)
  }

  // Hết file (đoạn chọn kéo tới cuối tiểu mục): coi như phát hết đoạn.
  audio.addEventListener('ended', () => {
    const cb = stopAt < Infinity ? onDone : null
    stop()
    cb?.()
  })
  onBeforeUnmount(() => {
    stop()
    audio.removeAttribute('src')
  })
  return { playing, current, mode, play, stop }
}
