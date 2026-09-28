// Nghe thử như video ngay trong hộp "Tạo video cả cuốn": phát nối liền các đoạn tiếng của
// dòng thời gian (trích đoạn → lặng màn tựa → thẻ chương → từng tiểu mục → quãng nghỉ →
// màn kết) bằng một thẻ audio riêng, báo giờ hiện tại trên cả video để hộp thoại vẽ đúng
// khung hình. Tua được tới bất kỳ đâu. Không đụng trình nghe chính (chỉ tạm dừng nó).
import { onBeforeUnmount, ref, shallowRef } from 'vue'
import type { BookVideoSeg } from './backend'
import { pause as pauseMain, player } from './player'

export function useVideoPreview() {
  const audio = new Audio()
  audio.preload = 'auto'
  const playing = ref(false)
  const pos = ref(0) // giây trên cả video
  const total = ref(0)
  const segs = shallowRef<BookVideoSeg[]>([])
  let at: number[] = [] // giờ bắt đầu từng đoạn
  let i = -1 // đoạn đang phát
  let clockFrom = 0 // performance.now() ứng với đầu phần lặng (đoạn lặng / hết file sớm)
  let silentMode = false
  let raf = 0
  let seq = 0

  function load(list: BookVideoSeg[]) {
    stop()
    segs.value = list
    at = []
    let t = 0
    for (const s of list) {
      at.push(t)
      t += s.dur
    }
    total.value = t
    pos.value = 0
  }

  function segAt(t: number) {
    let k = at.length - 1
    while (k > 0 && at[k] > t) k--
    return Math.max(0, k)
  }

  async function startSeg(k: number, offset: number) {
    const s = segs.value[k]
    i = k
    const my = ++seq
    if (!s) return
    if (s.silence) {
      audio.pause()
      silentMode = true
      clockFrom = performance.now() - offset * 1000
      return
    }
    silentMode = false
    const url = trackUrl(s.extra ? 'extra:' + s.extra : s.file)
    if (!audio.src.endsWith(url)) audio.src = url
    audio.currentTime = s.start + offset
    try {
      await audio.play()
    } catch {
      if (my === seq) pause()
    }
  }

  // Đoạn tiếng chỉ có tên file: hộp thoại cho biết URL phục vụ của từng file.
  let urlOf: (file: string) => string = (f) => f
  const trackUrl = (file: string) => urlOf(file)

  function tick() {
    if (!playing.value) return
    const s = segs.value[i]
    if (!s) return stop(true)
    let t: number
    if (silentMode) t = at[i] + (performance.now() - clockFrom) / 1000
    else t = at[i] + Math.max(0, audio.currentTime - s.start)
    if (t >= at[i] + s.dur) {
      if (i + 1 >= segs.value.length) return stop(true)
      void startSeg(i + 1, 0)
      t = at[i]
    }
    pos.value = Math.min(t, total.value)
    raf = requestAnimationFrame(tick)
  }

  // File hết sớm hơn độ dài đoạn: phần còn lại tính như lặng cho khớp video.
  audio.addEventListener('ended', () => {
    if (!playing.value || silentMode) return
    silentMode = true
    clockFrom = performance.now() - (pos.value - at[i]) * 1000
  })

  function play(from = pos.value, resolve?: (file: string) => string) {
    if (resolve) urlOf = resolve
    if (!segs.value.length) return
    if (player.playing) pauseMain()
    if (from >= total.value - 0.05) from = 0
    const k = segAt(from)
    playing.value = true
    pos.value = from
    void startSeg(k, from - at[k])
    cancelAnimationFrame(raf)
    raf = requestAnimationFrame(tick)
  }

  function pause() {
    playing.value = false
    audio.pause()
    seq++
    cancelAnimationFrame(raf)
  }

  function seek(t: number) {
    const was = playing.value
    pause()
    pos.value = Math.min(Math.max(0, t), total.value)
    if (was) play(pos.value)
  }

  function stop(atEnd = false) {
    pause()
    if (!atEnd) pos.value = 0
  }

  onBeforeUnmount(() => {
    pause()
    audio.removeAttribute('src')
  })
  return { playing, pos, total, load, play, pause, seek, stop }
}
