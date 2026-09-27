<script setup lang="ts">
// B6 Nghe thử (không bắt buộc, wireframe D12): render thật lời mở đầu + 2 tiểu mục đầu
// (mỗi đoạn ~500 ký tự đầu), phát trong app, hiện đúng lời đã đọc.
// - Sửa lời đọc rồi "Lưu & đọc lại đoạn này" — bản cuối dùng lời đã sửa.
// - Bôi đen một từ → "Đọc từ này là…" thêm vào từ điển cách đọc (áp cả cuốn).
// - Khung bên phải: Từ điển của cuốn này. Sửa chi tiết cả đoạn để sau, ở Sửa sách.
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { AlertCircle, BookA, BookOpen, Check, Globe, Loader2, Pause, Pencil, Play, Plus, RotateCcw, Trash2 } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import DictWordPopover from '../../components/DictWordPopover.vue'
import { errText, speakSample } from '../../lib/backend'
import { useClipPlayer } from '../../lib/audio'
import { WORD_HINT, globalDict, globalReading, isDictWord, loadGlobalDict } from '../../lib/dict'
import {
  INTRO_LABEL, addBookWord, addPreview, editClip, ensurePreview, estRender, markHeard, refreshWordCounts, removeBookWord,
  selectedStems, state, stemTitle,
} from '../../lib/store'

const player = useClipPlayer(markHeard)
const drafts = reactive<Record<string, string>>({})
const more = ref('')

onMounted(() => {
  void ensurePreview()
  void loadGlobalDict()
  void refreshWordCounts()
})

// Mỗi lần có đoạn mới render → bản nháp = đúng lời đã đọc.
watch(
  () => state.clips.map((c) => c.stem + '\u0000' + c.text + '\u0000' + c.url),
  () => {
    for (const c of state.clips) drafts[c.stem] = c.text
  },
  { immediate: true },
)

const others = computed(() => selectedStems.value.filter((s) => !state.clips.some((c) => c.stem === s)))
const fmtDur = (sec: number) => `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, '0')}`

async function addMore() {
  if (!more.value) return
  const stem = more.value
  more.value = ''
  await addPreview([stem])
}

const rightsOk = computed(() => !!state.rightsConfirmedAt)
function toggleRights(e: Event) {
  state.rightsConfirmedAt = (e.target as HTMLInputElement).checked ? new Date().toISOString() : ''
}

// ── Bôi đen trong ô lời đọc ───────────────────────────────────────────────
const areas = new Map<string, HTMLTextAreaElement>()
function setArea(stem: string, el: unknown) {
  if (el instanceof HTMLTextAreaElement) areas.set(stem, el)
  else areas.delete(stem)
}
const sel = reactive({ stem: '', text: '' })
function onSelect(stem: string) {
  const el = areas.get(stem)
  if (!el) return
  const t = el.value.slice(el.selectionStart, el.selectionEnd).trim()
  if (t) {
    sel.stem = stem
    sel.text = t
  } else if (sel.stem === stem && !teach.word) {
    sel.stem = ''
    sel.text = ''
  }
}
const selWord = computed(() => (isDictWord(sel.text) ? sel.text : ''))

// Nghe thử: đoạn bôi đen (hoặc ~300 ký tự đầu) bằng giọng đang chọn.
const trying = ref('')
const tryError = ref('')
async function trySelection(stem: string) {
  const key = 'try-' + stem
  if (player.playing.value === key) return player.stop()
  const el = areas.get(stem)
  let piece = sel.stem === stem ? sel.text : ''
  if (!piece && el) piece = el.value.trim().slice(0, 300)
  if (!piece) return
  tryError.value = ''
  trying.value = stem
  try {
    await player.toggle(key, await speakSample(state.voice, piece))
  } catch (e) {
    tryError.value = errText(e)
  } finally {
    trying.value = ''
  }
}

// ── "Đọc từ này là…" ─────────────────────────────────────────────────────
const teach = reactive({ word: '', stem: '', busy: false, error: '' })
function openTeach(word: string, stem = '') {
  teach.word = word
  teach.stem = stem
  teach.error = ''
  if (state.wordCounts[word] === undefined) void refreshWordCounts([word])
}
async function onTeach(reading: string, scope: 'book' | 'global') {
  teach.busy = true
  teach.error = ''
  try {
    await addBookWord(teach.word, reading, scope)
    teach.word = ''
    sel.stem = ''
    sel.text = ''
  } catch (e) {
    teach.error = errText(e)
  } finally {
    teach.busy = false
  }
}
function closePopups(e: MouseEvent) {
  const t = e.target as HTMLElement
  if (teach.word && !t.closest('[data-teach]')) teach.word = ''
}
onMounted(() => document.addEventListener('mousedown', closePopups))
onBeforeUnmount(() => document.removeEventListener('mousedown', closePopups))

// ── Khung Từ điển của cuốn ────────────────────────────────────────────────
const dictRows = computed(() => Object.entries(state.bookDict).sort((a, b) => a[0].localeCompare(b[0], 'vi')))
const mineGlobal = computed(() => globalDict.entries.filter((e) => e.mine).length)
const addForm = reactive({ open: false, word: '', reading: '', global: false, busy: false, error: '' })
async function addFromForm() {
  const w = addForm.word.trim()
  if (!isDictWord(w)) return void (addForm.error = WORD_HINT)
  if (!addForm.reading.trim()) return void (addForm.error = 'Chưa gõ cách đọc.')
  addForm.busy = true
  addForm.error = ''
  try {
    await addBookWord(w, addForm.reading.trim(), addForm.global ? 'global' : 'book')
    Object.assign(addForm, { open: false, word: '', reading: '', global: false })
  } catch (e) {
    addForm.error = errText(e)
  } finally {
    addForm.busy = false
  }
}
const sayKey = ref('')
async function say(word: string, reading: string) {
  const key = 'dict-' + word
  if (player.playing.value === key) return player.stop()
  sayKey.value = key
  try {
    await player.toggle(key, await speakSample(state.voice, reading))
  } catch (e) {
    tryError.value = errText(e)
  } finally {
    sayKey.value = ''
  }
}
</script>

<template>
  <div class="flex gap-6 items-start">
    <div class="max-w-3xl flex-1 min-w-0">
      <h1 class="text-xl font-semibold tracking-tight">Nghe thử trước khi render</h1>
      <p class="text-sm text-muted-foreground">
        Nghe vài đoạn để chốt giọng và cách đọc. Mỗi đoạn chỉ đọc khoảng 500 ký tự đầu. Render cả cuốn mất khoảng {{ estRender }} phút.
      </p>
      <p class="mt-2 text-xs text-muted-foreground flex items-center gap-1.5"><BookOpen class="w-3.5 h-3.5" /> Sửa chi tiết cả đoạn sau khi tạo xong: Thư viện → ⋯ → <b>Sửa sách</b>.</p>
      <p v-if="state.previewing" class="mt-4 text-sm text-muted-foreground flex items-center gap-2">
        <Loader2 class="w-4 h-4 animate-spin" /> Đang đọc thử{{ state.clips.length ? ' đoạn mới' : ' các đoạn đầu' }}… lần đầu mất khoảng nửa phút để nạp bộ đọc.
      </p>
      <p v-if="state.previewError" class="mt-4 text-sm text-destructive">{{ state.previewError }}</p>
      <p v-if="state.previewNote && !state.previewing" class="mt-4 rounded-md border border-rag-amber/50 bg-rag-amber/10 px-3 py-2 text-sm">{{ state.previewNote }}</p>
      <p v-if="player.error.value || tryError" class="mt-2 text-sm text-destructive">{{ player.error.value || tryError }}</p>

      <div class="mt-5 space-y-3">
        <div v-for="s in state.clips" :key="s.stem" class="rounded-lg border p-4" :class="drafts[s.stem] !== s.text ? 'border-rag-amber' : 'border-border'">
          <div class="flex items-center gap-3">
            <button :aria-label="player.playing.value === s.stem ? 'Dừng' : 'Nghe'" class="h-9 w-9 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0" @click="player.toggle(s.stem, s.url)">
              <component :is="player.playing.value === s.stem ? Pause : Play" class="w-4 h-4" />
            </button>
            <span class="flex-1 font-medium text-sm">{{ s.stem === 'intro' ? INTRO_LABEL : s.title }}</span>
            <Check v-if="state.heard.includes(s.stem)" class="w-4 h-4 text-rag-green" aria-label="Đã nghe" />
            <Badge variant="secondary">{{ s.full ? 'Đã render' : 'Đã render đoạn đầu' }} · {{ fmtDur(s.durationSec) }}</Badge>
          </div>
          <p class="mt-3 text-xs font-medium text-muted-foreground flex items-center gap-1.5"><Pencil class="w-3 h-3" /> Lời đọc · sửa chữ, hoặc bôi đen một từ để dạy cách đọc</p>
          <div class="relative mt-1" data-teach>
            <textarea :ref="(el) => setArea(s.stem, el)" v-model="drafts[s.stem]" spellcheck="false"
              class="w-full h-40 rounded-md border bg-background p-3 text-sm leading-relaxed resize-y focus:outline-none focus:ring-1"
              :class="drafts[s.stem] !== s.text ? 'border-rag-amber focus:ring-rag-amber' : 'border-input focus:ring-ring'"
              @select="onSelect(s.stem)" @mouseup="onSelect(s.stem)" @keyup="onSelect(s.stem)"></textarea>
            <!-- Thanh nổi khi bôi đen -->
            <div v-if="sel.stem === s.stem && sel.text && !teach.word" class="absolute right-2 top-2 rounded-lg border border-border bg-popover shadow-xl p-1 flex items-center gap-1 text-sm z-20">
              <button v-if="selWord" type="button" class="h-8 px-2.5 rounded-md bg-primary/10 text-primary font-medium flex items-center gap-1.5 hover:bg-primary/15" @click="openTeach(selWord, s.stem)">
                <BookA class="w-4 h-4" /> Đọc "{{ selWord }}" là…
              </button>
              <span v-else class="px-2 text-xs text-muted-foreground max-w-56">Bôi đen đúng một từ để dạy cách đọc</span>
              <button type="button" class="h-8 px-2.5 rounded-md flex items-center gap-1.5 hover:bg-muted" @click="trySelection(s.stem)">
                <Loader2 v-if="trying === s.stem" class="w-3.5 h-3.5 animate-spin" /><component :is="player.playing.value === 'try-' + s.stem ? Pause : Play" v-else class="w-3.5 h-3.5" /> Nghe thử
              </button>
            </div>
            <DictWordPopover v-if="teach.word && teach.stem === s.stem" :key="teach.word" class="absolute right-2 top-12" :word="teach.word" :voice="state.voice"
              :reading="state.bookDict[teach.word] || globalReading(teach.word)" :count="state.wordCounts[teach.word] ?? null"
              confirm-label="Thêm và đọc lại" :busy="teach.busy" :error="teach.error" @add="onTeach" @cancel="teach.word = ''" />
          </div>
          <div v-if="drafts[s.stem] !== s.text" class="mt-2 rounded-md border border-rag-amber/50 bg-rag-amber/10 px-3 py-2 text-xs flex items-center gap-2">
            <AlertCircle class="w-3.5 h-3.5 text-rag-amber shrink-0" /> <span><b>Chưa lưu.</b> Bấm <b>Lưu & đọc lại đoạn này</b> để nghe bản sửa. Render cả cuốn sẽ dùng bản đã lưu.</span>
          </div>
          <div class="mt-2 flex flex-wrap items-center gap-2">
            <Button variant="outline" size="sm" :disabled="!!trying" @click="trySelection(s.stem)">
              <Loader2 v-if="trying === s.stem" class="w-3.5 h-3.5 animate-spin" /><component :is="player.playing.value === 'try-' + s.stem ? Pause : Play" v-else class="w-3.5 h-3.5" /> Nghe thử đoạn chọn
            </Button>
            <span class="text-[11px] text-muted-foreground">Bôi đen một câu rồi bấm để nghe thử.</span>
            <div class="ml-auto flex gap-2">
              <Button v-if="drafts[s.stem] !== s.text" variant="ghost" size="sm" @click="drafts[s.stem] = s.text">Huỷ thay đổi</Button>
              <Button size="sm" :disabled="drafts[s.stem] === s.text || state.previewing" @click="editClip(s.stem, drafts[s.stem])">
                <RotateCcw class="w-4 h-4" /> Lưu & đọc lại đoạn này
              </Button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="others.length && state.clips.length" class="mt-3 flex items-center gap-2">
        <select v-model="more" class="h-9 rounded-md border border-input bg-background px-2 text-sm max-w-md" :disabled="state.previewing" aria-label="Chọn thêm đoạn khác để nghe thử">
          <option value="">+ Chọn thêm đoạn khác để nghe thử</option>
          <option v-for="st in others" :key="st" :value="st">{{ stemTitle(st) }}</option>
        </select>
        <Button variant="outline" size="sm" :disabled="!more || state.previewing" @click="addMore">Nghe thử đoạn này</Button>
      </div>

      <!-- Xác nhận quyền dùng tài liệu cho từng cuốn — bắt buộc trước khi render cả cuốn -->
      <label v-if="state.clips.length" class="mt-6 flex cursor-pointer items-start gap-2.5 rounded-lg border border-border bg-muted/30 p-4 text-sm"
        :class="rightsOk && 'border-primary/40 bg-primary/5'">
        <input :checked="rightsOk" type="checkbox" class="mt-0.5 h-4 w-4 accent-[hsl(var(--primary))]" @change="toggleRights" />
        <span>
          Tôi xác nhận có quyền dùng tài liệu này để làm sách nói
          <span class="text-muted-foreground">(tài liệu của tôi, tác phẩm đã hết thời hạn bảo hộ, hoặc được tác giả cho phép).</span>
        </span>
      </label>
    </div>

    <!-- Khung Từ điển của cuốn này -->
    <aside class="w-80 shrink-0 sticky top-0 rounded-lg border border-border flex flex-col max-h-[calc(100dvh-14rem)]">
      <div class="px-4 py-3 border-b border-border">
        <div class="text-sm font-semibold flex items-center gap-2"><BookA class="w-4 h-4" /> Từ điển của cuốn này · {{ dictRows.length }}</div>
        <p class="text-[11px] text-muted-foreground mt-0.5">Chỉ đổi cách đọc, chữ hiện khi nghe giữ nguyên. Áp cho cả cuốn khi render.</p>
      </div>
      <div class="flex-1 overflow-auto divide-y divide-border">
        <div v-for="[w, r] in dictRows" :key="w" class="relative px-4 py-2.5 flex items-center gap-2 text-sm" data-teach>
          <div class="min-w-0 flex-1">
            <div class="truncate"><b>{{ w }}</b> <span class="text-muted-foreground">→</span> {{ r }}</div>
            <div class="text-[11px] text-muted-foreground">{{ state.wordCounts[w] === undefined ? '…' : state.wordCounts[w] + ' chỗ trong sách' }}</div>
          </div>
          <button type="button" class="text-muted-foreground hover:text-foreground" :aria-label="'Nghe ' + w" @click="say(w, r)">
            <Loader2 v-if="sayKey === 'dict-' + w" class="w-3.5 h-3.5 animate-spin" /><component :is="player.playing.value === 'dict-' + w ? Pause : Play" v-else class="w-3.5 h-3.5" />
          </button>
          <button type="button" class="text-muted-foreground hover:text-foreground" :aria-label="'Sửa ' + w" @click="openTeach(w, '__dict')"><Pencil class="w-3.5 h-3.5" /></button>
          <button type="button" class="text-muted-foreground hover:text-destructive" :aria-label="'Xoá ' + w" :disabled="state.previewing" @click="removeBookWord(w)"><Trash2 class="w-3.5 h-3.5" /></button>
          <DictWordPopover v-if="teach.word === w && teach.stem === '__dict'" :key="'e' + w" class="absolute right-2 top-full mt-1" :word="w" :voice="state.voice" :reading="r"
            :count="state.wordCounts[w] ?? null" :scopes="false" confirm-label="Lưu và đọc lại" :busy="teach.busy" :error="teach.error"
            @add="(rd) => onTeach(rd, 'book')" @cancel="teach.word = ''" />
        </div>
        <p v-if="!dictRows.length && !addForm.open" class="px-4 py-3 text-xs text-muted-foreground">Chưa có từ nào. Bôi đen một từ trong ô lời đọc, hoặc bấm Thêm từ.</p>
        <form v-if="addForm.open" class="px-4 py-3 space-y-2 bg-muted/40" @submit.prevent="addFromForm">
          <div class="text-xs font-medium">Thêm từ</div>
          <div class="grid grid-cols-[1fr_auto_1fr] gap-1.5 items-center">
            <input v-model="addForm.word" class="h-8 min-w-0 rounded-md border border-input bg-background px-2 text-sm" placeholder="Chữ trong sách" />
            <span class="text-muted-foreground text-sm">→</span>
            <input v-model="addForm.reading" class="h-8 min-w-0 rounded-md border border-input bg-background px-2 text-sm" placeholder="Đọc là" />
          </div>
          <p v-if="addForm.error" class="text-xs text-destructive">{{ addForm.error }}</p>
          <div class="flex items-center gap-2 text-xs">
            <label class="flex items-center gap-1.5"><input v-model="addForm.global" type="checkbox" class="accent-[hsl(var(--primary))]" /> Dùng cho mọi sách</label>
            <span class="flex-1"></span>
            <Button type="button" variant="ghost" size="sm" @click="addForm.open = false">Huỷ</Button>
            <Button type="submit" size="sm" :disabled="addForm.busy"><Loader2 v-if="addForm.busy" class="w-3.5 h-3.5 animate-spin" /> Thêm</Button>
          </div>
        </form>
        <button v-else type="button" class="w-full px-4 py-2.5 text-sm text-primary flex items-center gap-1.5 hover:bg-muted/50" @click="addForm.open = true"><Plus class="w-4 h-4" /> Thêm từ</button>
      </div>
      <div class="px-4 py-3 border-t border-border text-[11px] text-muted-foreground leading-relaxed flex gap-1.5">
        <Globe class="w-3.5 h-3.5 shrink-0 mt-0.5" /> Từ dùng cho mọi sách nằm ở Cài đặt → Từ điển chung{{ mineGlobal ? ` (đang có ${mineGlobal} từ tự thêm)` : '' }}.
      </div>
    </aside>
  </div>
</template>
