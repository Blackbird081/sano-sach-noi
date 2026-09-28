// Hộp "Chia sẻ đoạn hay" (wireframe D14): mở từ chế độ Phóng to lời đọc hoặc Xem cả
// lời, chọn sẵn câu đang đọc. Nhớ khung / nền / loại đã chọn cho lần sau.
import { reactive, watch } from 'vue'
import type { ShareKind, ShareRatio } from './shareCard'

const KEY = 'sano.share'
function read(): { kind: ShareKind; ratio: ShareRatio; bg: number } {
  const def = { kind: 'image' as ShareKind, ratio: 'square' as ShareRatio, bg: 0 }
  try {
    const v = JSON.parse(localStorage.getItem(KEY) ?? 'null')
    return {
      kind: v?.kind === 'video' ? 'video' : def.kind,
      ratio: v?.ratio === 'story' ? 'story' : def.ratio,
      bg: Number.isInteger(v?.bg) && v.bg >= 0 && v.bg < 5 ? v.bg : def.bg,
    }
  } catch {
    return def
  }
}

export const shareUI = reactive({
  open: false,
  sentence: 0, // câu chọn sẵn khi mở
  ...read(),
})

watch(
  () => [shareUI.kind, shareUI.ratio, shareUI.bg],
  () => {
    try {
      localStorage.setItem(KEY, JSON.stringify({ kind: shareUI.kind, ratio: shareUI.ratio, bg: shareUI.bg }))
    } catch {
      // không lưu được thì thôi
    }
  },
)

export function openShare(sentence: number) {
  shareUI.sentence = sentence
  shareUI.open = true
}
