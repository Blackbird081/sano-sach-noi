// Từ điển cách đọc (wireframe D12): tách từ giống phần Go (bookmaker pronunTokenRe)
// để đếm "N chỗ" và tìm mục có từ ngay ở giao diện.
import { reactive } from 'vue'
import { pronunciations, type DictEntry } from './backend'

const TOKEN = /[\p{L}\p{N}]+(?:&[\p{L}\p{N}]+)*/gu
const KEY = /^[\p{L}\p{N}]+(?:&[\p{L}\p{N}]+)*$/u

/** Một từ tra được trong từ điển: chữ và số liền nhau, có thể nối bằng & (KPI, Nielsen, R&D). */
export function isDictWord(w: string) {
  return KEY.test(w)
}

/** Số lần word xuất hiện nguyên từ trong text (phân biệt hoa thường). */
export function countWord(text: string, word: string) {
  if (!word) return 0
  let n = 0
  for (const m of text.matchAll(TOKEN)) if (m[0] === word) n++
  return n
}

export const WORD_HINT = 'Chỉ một từ: chữ và số liền nhau, có thể nối bằng & (vd KPI, Nielsen, R&D).'

/** Từ điển chung + bộ chuẩn (nạp một lần, nạp lại sau khi sửa). */
export const globalDict = reactive({ entries: [] as DictEntry[], loaded: false })

export async function loadGlobalDict(force = false) {
  if (globalDict.loaded && !force) return
  try {
    globalDict.entries = await pronunciations()
    globalDict.loaded = true
  } catch {
    // không đọc được thì coi như trống
  }
}

/** Cách đọc đang có của một từ ở bộ chuẩn / từ điển chung ('' = chưa có). */
export function globalReading(word: string) {
  return globalDict.entries.find((e) => e.word === word)?.reading ?? ''
}
