// Nút "Tạo video" ở màn nghe (wireframe D17): một hộp, đầu hộp chọn Video ngắn (phần video
// của hộp Chia sẻ D14) hoặc Video cả cuốn (D15/D16). Hai loại vẫn là hai hộp thoại riêng,
// đổi loại thì đóng hộp này mở hộp kia tại chỗ. Nhớ loại người dùng chọn lần trước.
import { ref, watch } from 'vue'
import { shareUI } from './share'
import { bv, openBookVideo } from './bookVideo'
import { lyricIndex, lyrics } from './player'

export type VideoKind = 'short' | 'full'

const KEY = 'sano.video.kind'
function read(): VideoKind {
  try {
    return localStorage.getItem(KEY) === 'full' ? 'full' : 'short' // lần đầu: Video ngắn
  } catch {
    return 'short'
  }
}
export const videoKind = ref<VideoKind>(read())
watch(videoKind, (k) => {
  try {
    localStorage.setItem(KEY, k)
  } catch {
    // không lưu được thì thôi
  }
})

/** Video ngắn cần lời đọc theo câu của mục đang nghe. */
export const canShortVideo = () => !!lyrics.value?.sentences.length

function show(kind: VideoKind, slug: string) {
  if (kind === 'short') {
    bv.open = false
    shareUI.kind = 'video'
    shareUI.sentence = lyricIndex.value
    shareUI.open = true
  } else {
    shareUI.open = false
    openBookVideo(slug)
  }
}

/** Mở hộp Tạo video với loại dùng lần trước (mục đang nghe chưa có lời thì mở Video cả cuốn). */
export function openVideo(slug: string) {
  show(videoKind.value === 'short' && !canShortVideo() ? 'full' : videoKind.value, slug)
}

/** Người dùng đổi loại ngay trong hộp. */
export function switchVideo(kind: VideoKind, slug: string) {
  if (kind === 'short' && !canShortVideo()) return
  videoKind.value = kind
  show(kind, slug)
}
