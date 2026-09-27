// Màn Sửa sách (wireframe D11): cuốn đang sửa, bản sửa chưa đọc lại, lượt đọc
// lại / đổi giọng đang chạy (chạy nền, xem ở thanh bên từ màn nào cũng được).
import { computed, reactive } from 'vue'
import {
  cancelEdit, discardVoiceChange, editBook, editStatus, errText, onEvent, startReread, startVoiceChange,
  type EditSection, type EditStatus, type EditView, type SectionEdit,
} from './backend'
import { pause, player, reloadBook } from './player'
import { refreshLibrary, state } from './store'

export type EditTab = 'content' | 'info' | 'voice'

/** Bản sửa của một tiểu mục (chưa đọc lại). stem để nhận ra khi sách đã đổi mục lục. */
interface Draft {
  stem: string
  title: string
  text: string
}

export const edit = reactive({
  slug: '',
  tab: 'content' as EditTab,
  index: 0,
  view: null as EditView | null,
  loading: false,
  error: '',
  drafts: {} as Record<number, Draft>,
  status: null as EditStatus | null,
  actionError: '',
  back: 'library' as 'library' | 'player',
})

// ── Bản sửa: lưu trên máy theo từng cuốn (rời màn, tắt app vẫn còn) ─────────

const draftKey = (slug: string) => 'sano.sua.' + slug

function readDrafts(slug: string): Record<number, Draft> {
  try {
    const d = JSON.parse(localStorage.getItem(draftKey(slug)) || '{}')
    return d && typeof d === 'object' ? d : {}
  } catch {
    return {}
  }
}

function saveDrafts() {
  if (!edit.slug) return
  try {
    if (Object.keys(edit.drafts).length) localStorage.setItem(draftKey(edit.slug), JSON.stringify(edit.drafts))
    else localStorage.removeItem(draftKey(edit.slug))
  } catch {
    // không lưu được thì bản sửa chỉ còn tới khi đóng app
  }
}

/** Tiêu đề / chữ đang hiện của tiểu mục (bản sửa nếu có). */
export function current(sec: EditSection) {
  const d = edit.drafts[sec.index]
  return { title: d?.title ?? sec.title, text: d?.text ?? sec.text }
}

/** Ghi bản sửa; trùng bản đang có thì bỏ bản sửa. */
export function setDraft(sec: EditSection, title: string, text: string) {
  if (title === sec.title && text.trim() === sec.text.trim()) delete edit.drafts[sec.index]
  else edit.drafts[sec.index] = { stem: sec.stem, title, text }
  saveDrafts()
}

export function dropDraft(index: number) {
  delete edit.drafts[index]
  saveDrafts()
}

export const pendingIdx = computed(() =>
  Object.keys(edit.drafts).map(Number).filter((i) => !!edit.view?.sections[i]).sort((a, b) => a - b),
)

// ── Lượt đang chạy ────────────────────────────────────────────────────────

export const running = computed(() => !!edit.status?.running)
/** Lượt đọc lại / đổi giọng đang chạy là của cuốn đang mở ở màn Sửa sách. */
export const runningHere = computed(() => running.value && edit.status?.slug === edit.slug)
/** Tiểu mục đang nằm trong lượt đọc lại (đang đọc hoặc chờ) → khoá ô sửa. */
export function busySection(i: number) {
  const st = edit.status
  if (!st?.running || st.slug !== edit.slug || st.kind !== 'sections') return false
  return (st.queuedIdx ?? []).includes(i) && !st.doneIdx.includes(i)
}

function onProgress(st: EditStatus) {
  edit.status = st
  if (st.kind === 'sections' && st.slug === edit.slug) for (const i of st.doneIdx) dropDraftIfSent(i)
}

// Chữ đã gửi đi đọc của từng mục; người dùng sửa tiếp trong lúc đọc thì giữ bản sửa mới.
const sent = new Map<number, string>()
function dropDraftIfSent(i: number) {
  const d = edit.drafts[i]
  if (d && sent.get(i) === d.title + '\u0000' + d.text) dropDraft(i)
  sent.delete(i)
}

async function onFinished(st: EditStatus) {
  onProgress(st)
  void refreshLibrary()
  void reloadBook(st.slug)
  if (st.slug === edit.slug) await load(false)
}

let inited = false
export async function initEdit() {
  if (inited) return
  inited = true
  onEvent<EditStatus>('edit:progress', onProgress)
  onEvent<EditStatus>('edit:finished', (st) => void onFinished(st))
  try {
    edit.status = await editStatus()
  } catch {
    // giữ trống
  }
}

// ── Mở / nạp ──────────────────────────────────────────────────────────────

/** Mở màn Sửa sách. index: chọn sẵn tiểu mục (vd mục đang nghe). */
export function openEdit(slug: string, opts: { tab?: EditTab; index?: number } = {}) {
  if (player.slug === slug) pause() // sửa cuốn đang nghe: dừng để khỏi nghe lẫn bản cũ
  edit.back = state.view === 'player' ? 'player' : 'library'
  if (edit.slug !== slug) {
    edit.view = null
    edit.drafts = readDrafts(slug)
  }
  edit.slug = slug
  edit.tab = opts.tab ?? 'content'
  edit.index = opts.index ?? -1 // -1: chọn tiểu mục đầu (bỏ qua lời mở đầu)
  edit.actionError = ''
  state.view = 'edit'
  void initEdit()
  void load(true)
}

export async function load(reset: boolean) {
  const slug = edit.slug
  if (!slug) return
  if (reset) edit.loading = true
  edit.error = ''
  try {
    const v = await editBook(slug)
    if (edit.slug !== slug) return
    edit.view = v
    // Bản sửa của mục không còn (mục lục đã đổi) thì bỏ.
    for (const [k, d] of Object.entries(edit.drafts)) {
      if (v.sections[Number(k)]?.stem !== d.stem) delete edit.drafts[Number(k)]
    }
    saveDrafts()
    if (reset) {
      const firstReal = v.sections.findIndex((s) => !s.intro)
      if (edit.index < 0 || !v.sections[edit.index]) edit.index = Math.max(0, firstReal)
    }
  } catch (e) {
    edit.error = errText(e)
  } finally {
    edit.loading = false
  }
}

export function closeEdit() {
  state.view = edit.back
  if (edit.back === 'library') void refreshLibrary()
}

// ── Việc ──────────────────────────────────────────────────────────────────

async function act(fn: () => Promise<EditStatus>) {
  edit.actionError = ''
  try {
    edit.status = await fn()
  } catch (e) {
    edit.actionError = errText(e)
  }
}

/** Đọc lại các tiểu mục (mặc định: mọi mục đã sửa). */
export async function reread(indexes: number[] = pendingIdx.value) {
  const v = edit.view
  if (!v || !indexes.length) return
  const edits: SectionEdit[] = []
  for (const i of indexes) {
    const sec = v.sections[i]
    if (!sec) continue
    const c = current(sec)
    edits.push({ index: i, title: c.title, text: c.text })
    sent.set(i, c.title + '\u0000' + c.text)
  }
  await act(() => startReread(edit.slug, edits))
}

/** Đọc lại một tiểu mục với chữ cho sẵn (vd lời mở đầu theo tên mới). */
export async function rereadWith(e: SectionEdit) {
  const sec = edit.view?.sections[e.index]
  if (!sec) return
  setDraft(sec, e.title, e.text)
  await reread([e.index])
}

export async function changeVoice(voice: string) {
  await act(() => startVoiceChange(edit.slug, voice))
}

/** Dừng lượt đang chạy. Đổi giọng: bỏ phần đã đọc, giữ giọng cũ. */
export async function stopEdit() {
  await cancelEdit(true)
}

export async function discardVoice() {
  edit.actionError = ''
  try {
    await discardVoiceChange(edit.slug)
    await load(false)
  } catch (e) {
    edit.actionError = errText(e)
  }
}

/** Giây → "khoảng N phút" / "N giây". */
export function fmtRemain(sec: number) {
  if (sec <= 0) return ''
  if (sec < 60) return `còn khoảng ${Math.max(5, Math.round(sec / 5) * 5)} giây`
  return `còn khoảng ${Math.round(sec / 60)} phút`
}
