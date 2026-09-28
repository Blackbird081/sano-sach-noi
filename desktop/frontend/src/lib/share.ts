// Hộp "Chia sẻ đoạn hay" (wireframe D14), từ D17 dùng cho hai việc: Chia sẻ ảnh (nút ở màn
// nghe, Phóng to lời đọc, Xem cả lời) và Video ngắn (hộp Tạo video, lib/video.ts). Chọn sẵn
// câu đang đọc. Nhớ khung / nền đã chọn cho lần sau.
import { reactive, watch } from 'vue'
import type { ShareKind, ShareRatio } from './shareCard'

// v2: đổi thứ tự nền (Hoàng hôn đầu) nên không đọc lựa chọn của bản cũ.
const KEY = 'sano.share.v2'
function read(): { ratio: ShareRatio; bg: number } {
  const def = { ratio: 'story' as ShareRatio, bg: 0 } // Dọc 9:16 + Hoàng hôn
  try {
    const v = JSON.parse(localStorage.getItem(KEY) ?? 'null')
    return {
      ratio: v?.ratio === 'square' ? 'square' : def.ratio,
      bg: Number.isInteger(v?.bg) && v.bg >= 0 && v.bg < 5 ? v.bg : def.bg,
    }
  } catch {
    return def
  }
}

export const shareUI = reactive({
  open: false,
  sentence: 0, // câu chọn sẵn khi mở
  kind: 'image' as ShareKind, // theo chỗ mở: Chia sẻ ảnh hay Video ngắn
  ...read(),
})

watch(
  () => [shareUI.ratio, shareUI.bg],
  () => {
    try {
      localStorage.setItem(KEY, JSON.stringify({ ratio: shareUI.ratio, bg: shareUI.bg }))
    } catch {
      // không lưu được thì thôi
    }
  },
)

/** Chia sẻ ảnh có lời. */
export function openShare(sentence: number) {
  shareUI.kind = 'image'
  shareUI.sentence = sentence
  shareUI.open = true
}
