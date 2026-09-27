// Hộp "Nghe trên điện thoại" (wireframe D10): chọn máy → cài BookPlayer → tạo
// file (M4B, lưu thẳng vào Tải về) → chép sang máy. Lần sau (đã làm xong một
// lần) mở thẳng bước tạo file. Nhớ loại máy trong bộ nhớ cửa sổ app.
import { reactive, watch } from 'vue'
import { cancelM4B, m4b, startM4B } from './m4b'

export type Phone = 'ios' | 'android'
export type PhoneStep = 'pick' | 'app' | 'make' | 'copy'

export const STORE_URL: Record<Phone, string> = {
  ios: 'https://apps.apple.com/app/bookplayer/id1138219998',
  android: 'https://play.google.com/store/apps/details?id=com.tortugapower.audiobookplayer',
}

const KEY = 'sano.phone'
function read(): { device: Phone | null; done: boolean } {
  try {
    const v = JSON.parse(localStorage.getItem(KEY) ?? 'null')
    return { device: v?.device === 'ios' || v?.device === 'android' ? v.device : null, done: v?.done === true }
  } catch {
    return { device: null, done: false }
  }
}
const saved = read()

export const phone = reactive({
  open: false,
  slug: '',
  title: '',
  totalSec: 0,
  step: 'pick' as PhoneStep,
  device: saved.device,
  done: saved.done, // đã đi hết một lần → lần sau vào thẳng bước tạo file
  background: false, // đóng hộp khi đang tạo file: xong thì mở lại ở bước chép
})

function save() {
  try {
    localStorage.setItem(KEY, JSON.stringify({ device: phone.device, done: phone.done }))
  } catch {
    // không lưu được: lần sau hỏi lại từ đầu
  }
}

const runningHere = () => !!m4b.status?.running && m4b.status.slug === phone.slug

/** Mở hộp cho cuốn slug (nút "Nghe trên điện thoại"). */
export function openPhone(slug: string, title: string, totalSec = 0) {
  phone.slug = slug
  phone.title = title
  phone.totalSec = totalSec
  phone.background = false
  phone.open = true
  if (runningHere()) phone.step = 'make'
  else if (phone.done && phone.device) void make()
  else phone.step = 'pick'
}

/** Mở thẳng bước chép cho file vừa tạo (nút "Cách chép sang điện thoại"). Chưa chọn máy thì hỏi trước. */
export function showCopy(slug: string, title: string, totalSec = 0) {
  phone.slug = slug
  phone.title = title
  phone.totalSec = totalSec
  phone.background = false
  phone.open = true
  phone.step = phone.device ? 'copy' : 'pick'
}

export function pickDevice(d: Phone) {
  phone.device = d
  save()
  phone.step = 'app'
}

/** Bước tạo file: lưu thẳng vào Tải về, hoặc hỏi nơi lưu (ask). */
export async function make(ask = false) {
  phone.step = 'make'
  if (runningHere()) return
  await startM4B(phone.slug, ask)
}

/** "Đổi chỗ lưu": huỷ lượt đang chạy (nếu có) rồi tạo lại, lần này hỏi nơi lưu. */
export async function changeLocation() {
  if (runningHere()) {
    await cancelM4B()
    await new Promise<void>((resolve) => {
      const stop = watch(
        () => m4b.status?.running,
        (running) => {
          if (running) return
          stop()
          resolve()
        },
        { immediate: true },
      )
    })
  }
  await make(true)
}

export function runInBackground() {
  phone.background = true
  phone.open = false
}

export function closePhone() {
  phone.open = false
  phone.background = false
}

export function restartGuide() {
  phone.step = 'pick'
}

// Tạo xong: sang bước chép (mở lại hộp nếu đã để chạy nền), lần sau vào thẳng bước tạo file.
watch(
  () => m4b.status,
  (st) => {
    if (!st?.done || st.slug !== phone.slug) return
    if (!(phone.open && phone.step === 'make') && !phone.background) return
    phone.step = 'copy'
    phone.open = true
    phone.background = false
    phone.done = true
    save()
  },
)
