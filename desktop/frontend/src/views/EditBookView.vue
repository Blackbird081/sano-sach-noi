<script setup lang="ts">
// Màn Sửa sách (wireframe D11 đã duyệt): 3 tab Nội dung / Thông tin & bìa / Giọng đọc.
// - Nội dung: mục lục trái, ô sửa tiêu đề + chữ của tiểu mục phải. "Lưu & đọc lại mục
//   này" lưu chữ mới vào sách và đọc lại đúng mục đó; mục đã sửa chưa lưu có chấm vàng
//   (bản sửa giữ tạm khi rời màn). Dưới ô chữ luôn báo: chưa lưu / đang lưu / đã lưu xong.
// - Thông tin & bìa: đổi tên → bìa tự vẽ vẽ lại, lời mở đầu "Cuốn sách: …" đọc lại.
// - Giọng đọc: đổi giọng cả cuốn, chạy nền, xong hết mới thay.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  AlertCircle, Check, ChevronDown, ChevronLeft, ChevronRight, Headphones, ImagePlus, Loader2, Mic, Pause, Play,
  RotateCcw, Square, Wand2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '../components/sano/BookCover.vue'
import CategoryPicker from '../components/CategoryPicker.vue'
import { chooseCover, errText, saveBookInfo, setBookCover, speakSample, useAutoCover, type EditSection } from '../lib/backend'
import { useClipPlayer } from '../lib/audio'
import {
  busySection, changeVoice, closeEdit, current, discardVoice, dropDraft, edit, fmtRemain, load, pendingIdx, reread,
  rereadWith, runningHere, setDraft, stopEdit, type EditTab,
} from '../lib/edit'
import { categoryCounts, seriesKey } from '../lib/find'
import { refreshInfo } from '../lib/player'
import { loadVoices, refreshLibrary, state } from '../lib/store'

const clip = useClipPlayer()

const v = computed(() => edit.view)
const secs = computed(() => v.value?.sections ?? [])
const sec = computed<EditSection | undefined>(() => secs.value[edit.index])
const st = computed(() => (edit.status?.slug === edit.slug ? edit.status : null))
const voiceName = computed(() => v.value?.voice || 'Hải Đăng')
const fmtDur = (s: number) => `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
const totalLabel = computed(() => {
  const s = v.value?.durationSec ?? 0
  const h = Math.floor(s / 3600)
  const m = Math.round((s % 3600) / 60)
  return h ? `${h} giờ ${m} phút` : `${m} phút`
})

function setTab(t: EditTab) {
  edit.tab = t
  edit.actionError = ''
  clip.stop()
}

// ── Tab Nội dung ──────────────────────────────────────────────────────────

const chapters = computed(() => {
  const out: { index: number; title: string; items: EditSection[] }[] = []
  for (const s of secs.value) {
    const last = out[out.length - 1]
    if (last && last.index === s.chapterIndex) last.items.push(s)
    else out.push({ index: s.chapterIndex, title: s.intro ? 'Lời giới thiệu đầu sách' : s.chapter, items: [s] })
  }
  return out
})
const openCh = ref<Set<number>>(new Set())
watch(
  () => [sec.value?.chapterIndex, chapters.value.length] as const,
  ([ci]) => {
    if (ci !== undefined && !openCh.value.has(ci)) openCh.value = new Set([...openCh.value, ci])
  },
  { immediate: true },
)
function toggleCh(ci: number) {
  const s = new Set(openCh.value)
  if (s.has(ci)) s.delete(ci)
  else s.add(ci)
  openCh.value = s
}
const posLabel = computed(() => {
  const s = sec.value
  if (!s) return ''
  if (s.intro) return 'Lời giới thiệu đầu sách · Sano tự thêm, đọc trước chương đầu tiên'
  const ch = chapters.value.find((c) => c.index === s.chapterIndex)
  const n = ch?.items.findIndex((x) => x.index === s.index) ?? 0
  const chNo = chapters.value.filter((c) => !c.items[0]?.intro).findIndex((c) => c.index === s.chapterIndex) + 1
  return `Chương ${chNo} · Tiểu mục ${n + 1}/${ch?.items.length ?? 1}`
})

const title = computed({
  get: () => (sec.value ? current(sec.value).title : ''),
  set: (t: string) => sec.value && setDraft(sec.value, t, current(sec.value).text),
})
const text = computed({
  get: () => (sec.value ? current(sec.value).text : ''),
  set: (t: string) => sec.value && setDraft(sec.value, current(sec.value).title, t),
})
const dirty = computed(() => !!sec.value && edit.drafts[sec.value.index] !== undefined)
const busy = computed(() => !!sec.value && busySection(sec.value.index))
const readingThis = computed(() => busy.value && st.value?.current === sec.value?.index)
/** Mục đang chọn vừa lưu + đọc lại xong trong lượt gần nhất. */
const justSaved = computed(() => !!sec.value && sectionState(sec.value) === 'done')
/** Ước tính giây đọc lại một mục: nạp bộ đọc ~5 giây + đọc ~0,2 lần thời lượng (≈15 ký tự/giây nói). */
const estSec = computed(() => {
  const n = text.value.length + title.value.length
  return Math.max(5, Math.round((5 + (n / 15) * 0.2) / 5) * 5)
})

/** Tóm tắt chỗ đã sửa: "A" → "B" khi chỉ đổi một cụm; không thì "Đã sửa". */
const changeNote = computed(() => {
  const s = sec.value
  if (!s || !dirty.value) return ''
  const a = s.text
  const b = text.value
  if (a === b) return 'Đã sửa tiêu đề.'
  let i = 0
  while (i < a.length && i < b.length && a[i] === b[i]) i++
  let j = 0
  while (j < a.length - i && j < b.length - i && a[a.length - 1 - j] === b[b.length - 1 - j]) j++
  const from = a.slice(i, a.length - j).trim()
  const to = b.slice(i, b.length - j).trim()
  if (from && to && from.length <= 40 && to.length <= 40) return `Đã sửa "${from}" → "${to}".`
  return 'Đã sửa.'
})

function pick(i: number) {
  edit.index = i
  clip.stop()
}

// Nghe bản đang có / nghe thử đoạn bôi đen
const area = ref<HTMLTextAreaElement | null>(null)
const trying = ref(false)
const tryError = ref('')
function listenCurrent() {
  const s = sec.value
  if (!s || !v.value) return
  void clip.toggle('cur-' + s.stem, v.value.urls[s.stem])
}
async function trySelection() {
  const el = area.value
  const s = sec.value
  if (!el || !s) return
  if (clip.playing.value === 'try') return clip.stop()
  let piece = el.value.slice(el.selectionStart, el.selectionEnd).trim()
  if (!piece) piece = el.value.trim().slice(0, 300)
  tryError.value = ''
  trying.value = true
  try {
    await clip.toggle('try', await speakSample(voiceName.value, piece))
  } catch (e) {
    tryError.value = errText(e)
  } finally {
    trying.value = false
  }
}

function sectionState(s: EditSection): 'reading' | 'queued' | 'edited' | 'done' | '' {
  const x = st.value
  if (x?.running && x.kind === 'sections' && (x.queuedIdx ?? []).includes(s.index) && !x.doneIdx.includes(s.index)) {
    return x.current === s.index ? 'reading' : 'queued'
  }
  if (edit.drafts[s.index]) return 'edited'
  if (x && x.kind === 'sections' && x.doneIdx.includes(s.index) && Date.now() / 1000 - x.startedAt < 600) return 'done'
  return ''
}

const readingLabel = computed(() => {
  const x = st.value
  if (!x?.running || x.kind !== 'sections') return ''
  return `Đang lưu và đọc lại ${Math.min(x.finished + 1, x.total)}/${x.total} mục…`
})

// ── Tab Thông tin & bìa ───────────────────────────────────────────────────

const books = computed(() => state.library?.books ?? [])
const categories = computed(() => categoryCounts(books.value))
const seriesGroups = computed<[string, number][]>(() => {
  const m = new Map<string, [string, number]>()
  for (const b of books.value) {
    if (!b.series) continue
    const g = m.get(seriesKey(b.series)) ?? [b.series, 0]
    g[1]++
    m.set(seriesKey(b.series), g)
  }
  return [...m.values()]
})
const info = ref({ title: '', author: '', category: '', series: '', volume: 0 })
function resetInfo() {
  const b = v.value
  if (b) info.value = { title: b.title, author: b.author, category: b.category, series: b.series ?? '', volume: b.volume || 0 }
}
watch(() => v.value?.slug + '|' + v.value?.title + '|' + v.value?.author, resetInfo, { immediate: true })
const taken = computed(() =>
  books.value.filter((b) => b.slug !== edit.slug && b.series && seriesKey(b.series) === seriesKey(info.value.series)).map((b) => b.volume),
)
watch(() => info.value.series, (n, old) => {
  if (!n) info.value.volume = 0
  else if (seriesKey(n) !== seriesKey(old ?? '') && seriesKey(n) !== seriesKey(v.value?.series ?? '')) info.value.volume = Math.max(0, ...taken.value) + 1
})
const nameChanged = computed(() => !!v.value && (info.value.title.trim() !== v.value.title || info.value.author.trim() !== v.value.author))
const infoChanged = computed(() => {
  const b = v.value
  if (!b) return false
  return nameChanged.value || info.value.category !== b.category || (info.value.series || '') !== (b.series || '') || (info.value.volume || 0) !== (b.volume || 0)
})
const hasIntro = computed(() => !!secs.value[0]?.intro)
const saving = ref(false)
const infoMsg = ref('')
async function saveInfo() {
  if (!v.value || !info.value.title.trim() || saving.value) return
  saving.value = true
  edit.actionError = ''
  infoMsg.value = ''
  try {
    const r = await saveBookInfo(edit.slug, {
      title: info.value.title.trim(), author: info.value.author.trim(), category: info.value.category,
      series: info.value.series, volume: info.value.series ? Math.max(0, Math.floor(Number(info.value.volume) || 0)) : 0,
    })
    await load(false)
    void refreshLibrary()
    void refreshInfo(edit.slug)
    const parts = ['Đã lưu.']
    if (r.coverRedrawn) parts.push('Đã vẽ lại bìa theo tên mới.')
    infoMsg.value = parts.join(' ')
    if (r.introEdit) {
      await rereadWith(r.introEdit)
      if (!edit.actionError) infoMsg.value += ' Đã gửi đọc lại lời mở đầu, xong trong vài giây.'
      else infoMsg.value += ' Lời mở đầu mới chưa lưu, vào tab Nội dung bấm Lưu & đọc lại.'
    }
  } catch (e) {
    edit.actionError = errText(e)
  } finally {
    saving.value = false
  }
}
const coverBusy = ref(false)
async function pickCover() {
  edit.actionError = ''
  try {
    const f = await chooseCover()
    if (!f) return
    coverBusy.value = true
    await setBookCover(edit.slug, f.path)
    await load(false)
    void refreshLibrary()
    void refreshInfo(edit.slug)
  } catch (e) {
    edit.actionError = errText(e)
  } finally {
    coverBusy.value = false
  }
}
async function autoCover() {
  edit.actionError = ''
  coverBusy.value = true
  try {
    await useAutoCover(edit.slug)
    await load(false)
    void refreshLibrary()
    void refreshInfo(edit.slug)
  } catch (e) {
    edit.actionError = errText(e)
  } finally {
    coverBusy.value = false
  }
}

// ── Tab Giọng đọc ─────────────────────────────────────────────────────────

const REGIONS = ['Bắc', 'Trung', 'Nam']
const RECOMMENDED = ['Hải Đăng', 'Thiện Minh', 'Mỹ Duyên']
const region = ref('rec')
const newVoice = ref('')
watch(() => edit.tab, (t) => { if (t === 'voice') void loadVoices() }, { immediate: true })
const vparts = (desc: string) => {
  const [gender = '', reg = ''] = desc.split(' · ')
  return { gender, region: reg }
}
const shownVoices = computed(() => {
  const all = state.voices
  const list = region.value === 'rec' ? all.filter((x) => RECOMMENDED.includes(x.name)) : all.filter((x) => vparts(x.desc).region === region.value)
  return list.sort((a, b) => (RECOMMENDED.indexOf(a.name) + 1 || 99) - (RECOMMENDED.indexOf(b.name) + 1 || 99))
})
watch(() => [state.voices.length, v.value?.voice] as const, () => {
  if (newVoice.value) return
  const other = RECOMMENDED.find((n) => n !== voiceName.value && state.voices.some((x) => x.name === n))
  newVoice.value = v.value?.voiceJob?.voice || other || ''
}, { immediate: true })
const sampleText = computed(() => {
  const s = secs.value.find((x) => !x.intro && x.text.trim().length > 40) ?? secs.value[0]
  const t = (s?.text ?? '').replace(/\s+/g, ' ').trim()
  if (t.length <= 220) return t
  const cut = t.slice(0, 220)
  const end = Math.max(cut.lastIndexOf('. '), cut.lastIndexOf('? '), cut.lastIndexOf('! '))
  return end > 80 ? cut.slice(0, end + 1) : cut + '…'
})
const sampling = ref('')
const sampleError = ref('')
async function sample(voice: string, text: string, key: string) {
  if (clip.playing.value === key) return clip.stop()
  sampleError.value = ''
  sampling.value = key
  try {
    await clip.toggle(key, await speakSample(voice, text))
  } catch (e) {
    sampleError.value = errText(e)
  } finally {
    sampling.value = ''
  }
}
// Ước tính thời gian đọc cả cuốn: ~0,2 lần thời lượng (đo trên máy Apple Silicon).
const estMin = computed(() => Math.max(1, Math.round(((v.value?.durationSec ?? 0) * 0.2) / 60)))
const voiceRunning = computed(() => st.value?.running && st.value.kind === 'voice')
const voicePending = computed(() => !voiceRunning.value && v.value?.voiceJob ? v.value.voiceJob : null)
const voicePct = computed(() => (st.value && st.value.total ? Math.round((st.value.finished / st.value.total) * 100) : 0))

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape' && !(e.target instanceof HTMLTextAreaElement) && !(e.target instanceof HTMLInputElement)) closeEdit()
}
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <section class="flex-1 flex flex-col min-h-0">
    <div class="px-6 pt-5 border-b border-border">
      <div class="flex items-center gap-3">
        <button class="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1" @click="closeEdit"><ChevronLeft class="w-4 h-4" /> Quay lại</button>
        <template v-if="v">
          <div class="h-9 w-7 rounded overflow-hidden shrink-0"><BookCover :title="v.title" :author="v.author" :cover-image="v.coverUrl" class="h-full w-full" /></div>
          <div class="min-w-0">
            <h1 class="text-base font-semibold leading-tight truncate">Sửa sách · {{ v.title }}</h1>
            <p class="text-xs text-muted-foreground flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng {{ voiceName }} · {{ secs.length }} mục · {{ totalLabel }}</p>
          </div>
          <div v-if="pendingIdx.length || readingLabel" class="ml-auto flex items-center gap-2">
            <span class="text-xs text-muted-foreground">{{ readingLabel || `${pendingIdx.length} mục đã sửa, chưa lưu` }}</span>
            <Button size="sm" :disabled="!!edit.status?.running || !pendingIdx.length" @click="reread()"><RotateCcw class="w-4 h-4" /> Lưu & đọc lại {{ pendingIdx.length || '' }} mục</Button>
          </div>
        </template>
      </div>
      <div class="mt-4 flex gap-5 text-sm" role="tablist">
        <button v-for="t in ([['content', 'Nội dung'], ['info', 'Thông tin & bìa'], ['voice', 'Giọng đọc']] as const)" :key="t[0]" role="tab" :aria-selected="edit.tab === t[0]"
          class="pb-2.5 -mb-px border-b-2" :class="edit.tab === t[0] ? 'border-primary text-foreground font-medium' : 'border-transparent text-muted-foreground hover:text-foreground'"
          @click="setTab(t[0])">{{ t[1] }}</button>
      </div>
    </div>

    <div v-if="edit.loading && !v" class="flex-1 grid place-items-center text-sm text-muted-foreground"><Loader2 class="w-5 h-5 animate-spin" /></div>
    <div v-else-if="edit.error && !v" class="flex-1 grid place-items-center text-sm text-destructive">{{ edit.error }}</div>

    <template v-else-if="v">
      <p v-if="edit.actionError" class="mx-6 mt-3 text-xs text-destructive flex items-center gap-1.5"><AlertCircle class="w-3.5 h-3.5" /> {{ edit.actionError }}</p>

      <!-- ─── Tab Nội dung ─── -->
      <div v-if="edit.tab === 'content'" class="flex-1 flex min-h-0">
        <div class="w-72 shrink-0 border-r border-border overflow-auto py-2">
          <template v-for="c in chapters" :key="c.index">
            <button class="w-full flex items-center gap-1.5 px-3 py-2 text-xs font-semibold text-muted-foreground hover:text-foreground text-left" @click="toggleCh(c.index)">
              <component :is="openCh.has(c.index) ? ChevronDown : ChevronRight" class="w-3.5 h-3.5 shrink-0" /><span class="truncate">{{ c.title }}</span>
              <span v-if="!openCh.has(c.index) && c.items.some((s) => edit.drafts[s.index])" class="ml-auto h-2 w-2 rounded-full bg-rag-amber shrink-0"></span>
            </button>
            <template v-if="openCh.has(c.index)">
              <button v-for="s in c.items" :key="s.index" class="w-full flex items-center gap-2 pl-8 pr-3 py-2 text-sm text-left"
                :class="s.index === edit.index ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted'" @click="pick(s.index)">
                <span class="truncate flex-1">{{ current(s).title }}</span>
                <Loader2 v-if="sectionState(s) === 'reading'" class="w-3.5 h-3.5 shrink-0 animate-spin" />
                <span v-else-if="sectionState(s) === 'queued'" class="shrink-0 text-[11px] text-muted-foreground">chờ</span>
                <span v-else-if="sectionState(s) === 'edited'" class="shrink-0 h-2 w-2 rounded-full bg-rag-amber" title="Đã sửa, chưa lưu"></span>
                <Check v-else-if="sectionState(s) === 'done'" class="w-3.5 h-3.5 shrink-0 text-rag-green" />
                <span v-else class="text-xs tabular-nums text-muted-foreground shrink-0">{{ fmtDur(s.durationSec) }}</span>
              </button>
            </template>
          </template>
          <p class="px-3 pt-3 text-[11px] text-muted-foreground leading-relaxed flex gap-1.5"><span class="mt-1 h-2 w-2 rounded-full bg-rag-amber shrink-0"></span> Đã sửa, chưa lưu. Rời màn này vẫn giữ tạm bản sửa.</p>
        </div>

        <div v-if="sec" class="flex-1 flex flex-col min-w-0 p-5 gap-3">
          <div class="text-xs text-muted-foreground">{{ posLabel }}</div>
          <label class="block">
            <span class="text-xs font-medium text-muted-foreground">{{ sec.intro ? 'Tên sách (đổi ở tab Thông tin & bìa)' : 'Tiêu đề tiểu mục' }}</span>
            <input v-model="title" maxlength="200" :readonly="busy || sec.intro" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm read-only:text-muted-foreground" />
          </label>
          <label class="flex-1 flex flex-col min-h-0">
            <span class="text-xs font-medium text-muted-foreground flex items-center justify-between gap-3">
              Lời đọc
              <span class="font-normal truncate">Sửa chữ sai, cách đọc tên riêng, số, viết tắt. Xuống dòng = nghỉ dài.</span>
            </span>
            <textarea ref="area" v-model="text" :readonly="busy" spellcheck="false"
              class="mt-1 flex-1 w-full rounded-md border bg-background p-3 text-sm leading-relaxed resize-none focus:outline-none focus:ring-1"
              :class="dirty ? 'border-rag-amber focus:ring-rag-amber' : 'border-input focus:ring-ring'"></textarea>
          </label>
          <!-- Trạng thái lưu của mục đang chọn: chưa lưu → đang lưu, đọc lại → đã lưu xong -->
          <div v-if="readingThis" class="rounded-md bg-muted/60 px-3 py-2 text-xs flex items-center gap-2"><Loader2 class="w-3.5 h-3.5 animate-spin shrink-0" /> <span><b>Đang lưu và đọc lại mục này</b>{{ st?.remainSec ? ', ' + fmtRemain(st.remainSec) : `, khoảng ${estSec} giây` }}. Anh chọn mục khác sửa tiếp được.</span></div>
          <div v-else-if="busy" class="rounded-md bg-muted/60 px-3 py-2 text-xs flex items-center gap-2"><Loader2 class="w-3.5 h-3.5 shrink-0" /> <span><b>Đã xếp hàng</b>, lưu và đọc lại ngay sau mục đang làm.</span></div>
          <div v-else-if="dirty" class="rounded-md border border-rag-amber/50 bg-rag-amber/10 px-3 py-2 text-xs flex items-center gap-2"><AlertCircle class="w-3.5 h-3.5 text-rag-amber shrink-0" /> <span><b>Chưa lưu.</b> {{ changeNote }} Bấm <b>Lưu & đọc lại mục này</b> để đưa vào sách (khoảng {{ estSec }} giây). Chưa lưu thì khi nghe vẫn là bản cũ.</span></div>
          <div v-else-if="justSaved" class="rounded-md border border-rag-green/40 bg-rag-green/10 px-3 py-2 text-xs flex items-center gap-2"><Check class="w-3.5 h-3.5 text-rag-green shrink-0" /> <span><b>Đã lưu xong.</b> Mục này đã đọc lại bằng chữ mới, gói zip đã cập nhật. Bấm <b>Nghe bản đang có</b> để nghe lại.</span></div>
          <p v-if="tryError" class="text-xs text-destructive">{{ tryError }}</p>
          <div class="flex flex-wrap items-center gap-2 pt-1">
            <Button variant="outline" size="sm" @click="listenCurrent">
              <component :is="clip.playing.value === 'cur-' + sec.stem ? Pause : Headphones" class="w-4 h-4" /> Nghe bản đang có
            </Button>
            <Button variant="outline" size="sm" :disabled="busy || trying || !!edit.status?.running" @click="trySelection">
              <Loader2 v-if="trying" class="w-4 h-4 animate-spin" /><component :is="clip.playing.value === 'try' ? Pause : Play" v-else class="w-4 h-4" /> Nghe thử đoạn chọn
            </Button>
            <span class="text-[11px] text-muted-foreground">Bôi đen một câu rồi bấm để nghe thử trước khi đọc lại cả mục.</span>
            <div class="ml-auto flex gap-2">
              <Button v-if="dirty && !busy" variant="ghost" size="sm" title="Trả về chữ đang có trong sách" @click="dropDraft(sec.index)">Huỷ thay đổi</Button>
              <Button size="sm" :disabled="!dirty || busy || !!edit.status?.running" @click="reread([sec.index])"><Loader2 v-if="readingThis" class="w-4 h-4 animate-spin" /><Check v-else-if="justSaved && !dirty" class="w-4 h-4" /><RotateCcw v-else class="w-4 h-4" /> {{ readingThis ? 'Đang lưu…' : justSaved && !dirty ? 'Đã lưu' : 'Lưu & đọc lại mục này' }}</Button>
            </div>
          </div>
        </div>
      </div>

      <!-- ─── Tab Thông tin & bìa ─── -->
      <form v-else-if="edit.tab === 'info'" class="flex-1 flex min-h-0 p-6 gap-8 overflow-auto" @submit.prevent="saveInfo">
        <div class="w-[380px] shrink-0 space-y-3">
          <label class="block"><span class="text-xs font-medium text-muted-foreground">Tên sách</span>
            <input v-model="info.title" maxlength="200" class="mt-1 w-full h-9 rounded-md border bg-background px-3 text-sm" :class="info.title.trim() !== v.title ? 'border-rag-amber' : 'border-input'" /></label>
          <label class="block"><span class="text-xs font-medium text-muted-foreground">Tác giả</span>
            <input v-model="info.author" maxlength="200" placeholder="Không bắt buộc" class="mt-1 w-full h-9 rounded-md border bg-background px-3 text-sm" :class="info.author.trim() !== v.author ? 'border-rag-amber' : 'border-input'" /></label>
          <div><span class="text-xs font-medium text-muted-foreground">Danh mục</span>
            <CategoryPicker v-model="info.category" :categories="categories" class="mt-1" /></div>
          <div class="grid grid-cols-[1fr_90px] gap-2">
            <div class="min-w-0"><span class="text-xs font-medium text-muted-foreground">Bộ sách</span>
              <CategoryPicker v-model="info.series" kind="series" :categories="seriesGroups" class="mt-1" /></div>
            <label class="block" :class="!info.series && 'opacity-40'"><span class="text-xs font-medium text-muted-foreground">Tập số</span>
              <input :value="info.volume || ''" type="number" min="1" max="999" :disabled="!info.series" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm"
                @input="info.volume = Math.floor(Number(($event.target as HTMLInputElement).value)) || 0" /></label>
          </div>
          <p v-if="info.series && taken.length" class="text-xs text-muted-foreground">Bộ "{{ info.series }}" đang có tập {{ [...taken].sort((a, b) => a - b).join(', ') }}.</p>
          <div v-if="nameChanged && (v.coverAuto || hasIntro)" class="rounded-lg bg-muted/60 p-3 text-xs leading-relaxed space-y-1.5">
            <div class="font-medium text-foreground">Lưu xong, Sano tự làm thêm:</div>
            <div v-if="v.coverAuto" class="flex gap-1.5"><Wand2 class="w-3.5 h-3.5 shrink-0 mt-0.5" /> Vẽ lại bìa theo tên mới (bìa đang là bìa tự vẽ)</div>
            <div v-if="hasIntro" class="flex gap-1.5"><RotateCcw class="w-3.5 h-3.5 shrink-0 mt-0.5" /> Đọc lại lời giới thiệu: "Cuốn sách: {{ info.title.trim() }}." (vài giây)</div>
          </div>
          <p v-if="infoMsg" class="text-xs text-muted-foreground flex items-center gap-1.5"><Check class="w-3.5 h-3.5 text-rag-green" /> {{ infoMsg }}</p>
          <div class="flex gap-2 pt-1">
            <Button type="submit" size="sm" :disabled="saving || !info.title.trim() || !infoChanged"><Loader2 v-if="saving" class="w-4 h-4 animate-spin" /> Lưu</Button>
            <Button type="button" variant="ghost" size="sm" :disabled="saving || !infoChanged" @click="resetInfo(); infoMsg = ''">Huỷ</Button>
          </div>
        </div>
        <div>
          <div class="text-xs font-medium text-muted-foreground mb-2">Bìa</div>
          <div class="flex items-end gap-4">
            <template v-if="nameChanged && v.coverAuto">
              <div class="text-center">
                <div class="w-28 aspect-[3/4] rounded-lg overflow-hidden shadow opacity-60"><BookCover :title="v.title" :author="v.author" :cover-image="v.coverUrl" class="h-full w-full" /></div>
                <div class="mt-1 text-[11px] text-muted-foreground">Bìa cũ</div>
              </div>
              <ChevronRight class="w-5 h-5 mb-16 text-muted-foreground" />
              <div class="text-center">
                <div class="w-40 aspect-[3/4] rounded-lg overflow-hidden shadow-lg"><BookCover :title="info.title.trim() || v.title" :author="info.author.trim()" class="h-full w-full" /></div>
                <div class="mt-1 text-[11px] text-muted-foreground">Bìa mới, vẽ theo tên</div>
              </div>
            </template>
            <div v-else class="text-center">
              <div class="w-40 aspect-[3/4] rounded-lg overflow-hidden shadow-lg"><BookCover :title="v.title" :author="v.author" :cover-image="v.coverUrl" class="h-full w-full" /></div>
              <div class="mt-1 text-[11px] text-muted-foreground">{{ v.coverAuto ? 'Bìa tự vẽ theo tên' : 'Ảnh bìa đã chọn' }}</div>
            </div>
          </div>
          <div class="mt-4 flex gap-2">
            <Button type="button" variant="outline" size="sm" :disabled="coverBusy" @click="pickCover"><ImagePlus class="w-4 h-4" /> Chọn ảnh bìa…</Button>
            <Button type="button" variant="ghost" size="sm" :disabled="coverBusy || v.coverAuto" @click="autoCover"><Loader2 v-if="coverBusy" class="w-4 h-4 animate-spin" /><Wand2 v-else class="w-4 h-4" /> Dùng bìa tự vẽ</Button>
          </div>
          <p class="mt-2 text-[11px] text-muted-foreground max-w-xs leading-relaxed">Ảnh jpg, png, webp. Đã chọn ảnh thì đổi tên không vẽ lại bìa; bấm "Dùng bìa tự vẽ" để quay về bìa theo tên.</p>
        </div>
      </form>

      <!-- ─── Tab Giọng đọc ─── -->
      <div v-else class="flex-1 flex flex-col min-h-0 p-6 gap-4 overflow-auto">
        <template v-if="voiceRunning && st">
          <div class="max-w-xl rounded-lg border border-border p-4">
            <div class="flex items-center gap-2 font-medium"><Loader2 class="w-4 h-4 animate-spin" /> Đang đọc lại cả cuốn bằng giọng {{ st.voice }}</div>
            <div class="mt-3 h-2 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full transition-all" :style="{ width: voicePct + '%' }"></div></div>
            <div class="mt-1.5 flex justify-between text-xs text-muted-foreground tabular-nums"><span>{{ st.finished }}/{{ st.total }} mục</span><span>{{ fmtRemain(st.remainSec) }}</span></div>
            <p class="mt-3 text-xs text-muted-foreground leading-relaxed">Xong hết mới thay giọng cho cả cuốn, nên nghe lúc này vẫn là giọng {{ voiceName }}. Anh rời màn này hoặc đóng Sano: lần mở sau Sano đọc tiếp từ mục {{ st.finished + 1 }}, không đọc lại từ đầu.</p>
            <div class="mt-3 flex gap-2">
              <Button variant="outline" size="sm" @click="stopEdit"><Square class="w-4 h-4" /> Dừng, giữ giọng cũ</Button>
            </div>
          </div>
        </template>
        <template v-else>
          <div v-if="voicePending" class="max-w-3xl rounded-lg border border-rag-amber/60 bg-rag-amber/5 p-3 text-sm flex items-center gap-3">
            <AlertCircle class="w-4 h-4 text-rag-amber shrink-0" />
            <span class="flex-1">Đổi sang giọng <b>{{ voicePending.voice }}</b> đang dở: đã đọc {{ voicePending.done.length }}/{{ secs.length }} mục.</span>
            <Button size="sm" :disabled="!!edit.status?.running" @click="changeVoice(voicePending.voice)">Đọc tiếp</Button>
            <Button size="sm" variant="ghost" @click="discardVoice">Bỏ</Button>
          </div>
          <p class="text-sm">Đang đọc bằng giọng <b>{{ voiceName }}</b>. Chọn giọng mới, nghe thử một đoạn của chính cuốn này rồi đọc lại cả cuốn.</p>
          <div class="flex gap-2 text-xs">
            <button v-for="r in ['rec', ...REGIONS]" :key="r" class="h-7 px-3 rounded-full border flex items-center"
              :class="region === r ? 'bg-foreground text-background border-foreground' : 'border-border hover:bg-muted'" @click="region = r">{{ r === 'rec' ? 'Khuyên dùng' : 'Miền ' + r }}</button>
          </div>
          <div v-if="!state.voices.length" class="text-sm text-muted-foreground flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang tải danh sách giọng…</div>
          <div v-else class="grid grid-cols-3 gap-2 max-w-3xl">
            <div v-for="x in shownVoices" :key="x.name" role="radio" :aria-checked="x.name === newVoice" tabindex="0"
              class="rounded-lg border p-3 flex items-center gap-3 cursor-pointer" :class="x.name === newVoice ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50'"
              @click="newVoice = x.name" @keydown.enter.prevent="newVoice = x.name">
              <button type="button" class="h-8 w-8 rounded-full grid place-items-center border border-border shrink-0 hover:bg-muted" :aria-label="'Nghe mẫu giọng ' + x.name"
                @click.stop="sample(x.name, 'Bước thứ nhất, dừng việc đang làm và nhìn người nói.', 'v-' + x.name)">
                <Loader2 v-if="sampling === 'v-' + x.name" class="w-3.5 h-3.5 animate-spin" /><component :is="clip.playing.value === 'v-' + x.name ? Pause : Play" v-else class="w-3.5 h-3.5" />
              </button>
              <span class="min-w-0 flex-1">
                <span class="text-sm font-medium block truncate">{{ x.name }} <span v-if="x.name === voiceName" class="text-[11px] font-normal text-muted-foreground">· đang dùng</span></span>
                <span class="text-[11px] text-muted-foreground">{{ vparts(x.desc).gender.toLowerCase() }} · {{ vparts(x.desc).region }}<template v-if="RECOMMENDED.includes(x.name)"> · Khuyên dùng</template></span>
              </span>
              <Check v-if="x.name === newVoice" class="w-4 h-4 text-primary shrink-0" />
            </div>
          </div>
          <button v-if="newVoice && sampleText" type="button" class="rounded-lg border border-border p-3 max-w-3xl flex items-center gap-3 text-left hover:bg-muted/40"
            @click="sample(newVoice, sampleText, 'book-' + newVoice)">
            <span class="h-9 w-9 rounded-full bg-primary text-primary-foreground grid place-items-center shrink-0">
              <Loader2 v-if="sampling === 'book-' + newVoice" class="w-4 h-4 animate-spin" /><component :is="clip.playing.value === 'book-' + newVoice ? Pause : Play" v-else class="w-4 h-4" />
            </span>
            <span class="text-sm min-w-0">
              <span class="font-medium block">Nghe thử giọng {{ newVoice }} với đoạn trong sách</span>
              <span class="text-xs text-muted-foreground block truncate">"{{ sampleText }}"</span>
            </span>
          </button>
          <p v-if="sampleError" class="text-xs text-destructive">{{ sampleError }}</p>
          <div class="mt-auto flex items-center gap-3 border-t border-border pt-4">
            <Button :disabled="!newVoice || newVoice === voiceName || !!edit.status?.running" @click="changeVoice(newVoice)"><RotateCcw class="w-4 h-4" /> Đọc lại cả cuốn bằng giọng {{ newVoice || '…' }}</Button>
            <span class="text-xs text-muted-foreground leading-relaxed">{{ secs.length }} mục · ước tính khoảng {{ estMin }} phút. Chạy nền, trong lúc đó vẫn nghe bản cũ được.<br />Chỗ nghe dở, danh mục, bìa giữ nguyên. File đã chép sang điện thoại cần chép lại.</span>
          </div>
        </template>
      </div>
    </template>
  </section>
</template>
