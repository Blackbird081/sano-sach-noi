<script setup lang="ts">
// Cài đặt → Từ điển chung (wireframe D12): từ dùng cho mọi sách. Từ có sẵn (bộ chuẩn)
// sửa được = ghi đè cách đọc; chỉ xoá được từ tự thêm (xoá bản ghi đè = về cách đọc chuẩn).
import { computed, onMounted, reactive, ref } from 'vue'
import { Loader2, Pause, Pencil, Play, Plus, RotateCcw, Search, Trash2 } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { deleteGlobalPronunciation, errText, isDesktop, setGlobalPronunciation, speakSample } from '../lib/backend'
import { useClipPlayer } from '../lib/audio'
import { WORD_HINT, globalDict, isDictWord, loadGlobalDict } from '../lib/dict'

const clip = useClipPlayer()
onMounted(() => void loadGlobalDict(true))

const search = ref('')
const showAll = ref(false)
const rows = computed(() => {
  const q = search.value.trim().toLowerCase()
  const list = globalDict.entries.filter((e) => !q || (e.word + ' ' + e.reading).toLowerCase().includes(q))
  return q || showAll.value ? list : list.filter((e) => e.mine)
})
const builtinCount = computed(() => globalDict.entries.filter((e) => e.builtin && !e.mine).length)
const hiddenBuiltin = computed(() => (search.value.trim() || showAll.value ? 0 : builtinCount.value))

const editing = reactive<Record<string, string>>({})
const busy = ref('')
const error = ref('')
async function save(word: string, reading: string) {
  if (!reading.trim()) return void (error.value = 'Chưa gõ cách đọc.')
  busy.value = word
  error.value = ''
  try {
    await setGlobalPronunciation(word, reading)
    delete editing[word]
    await loadGlobalDict(true)
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = ''
  }
}
async function remove(word: string) {
  busy.value = word
  error.value = ''
  try {
    await deleteGlobalPronunciation(word)
    await loadGlobalDict(true)
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = ''
  }
}
function focusSelect({ el }: { el: unknown }) {
  // v-model gán giá trị sau khi mount (làm mất vùng chọn) → chọn ở nhịp vẽ kế tiếp.
  if (el instanceof HTMLInputElement) requestAnimationFrame(() => {
    el.focus()
    el.select()
  })
}
const add = reactive({ open: false, word: '', reading: '' })
async function addWord() {
  const w = add.word.trim()
  if (!isDictWord(w)) return void (error.value = WORD_HINT)
  await save(w, add.reading)
  if (!error.value) Object.assign(add, { open: false, word: '', reading: '' })
}
const sayBusy = ref('')
async function say(word: string, reading: string) {
  const key = 'g-' + word
  if (clip.playing.value === key) return clip.stop()
  sayBusy.value = key
  try {
    await clip.toggle(key, await speakSample('Hải Đăng', reading))
  } catch (e) {
    error.value = errText(e)
  } finally {
    sayBusy.value = ''
  }
}
</script>

<template>
  <div class="rounded-lg border border-border text-sm">
    <div class="p-3 flex gap-2 border-b border-border">
      <label class="flex-1 h-9 rounded-md border border-input px-3 flex items-center gap-2 text-muted-foreground focus-within:ring-1 focus-within:ring-ring">
        <Search class="w-4 h-4 shrink-0" /><input v-model="search" class="flex-1 min-w-0 bg-transparent outline-none text-foreground" placeholder="Tìm từ" />
      </label>
      <Button variant="outline" size="sm" :disabled="!isDesktop()" @click="add.open = true; error = ''"><Plus class="w-4 h-4" /> Thêm từ</Button>
    </div>
    <p v-if="error" class="px-3 pt-2 text-xs text-destructive">{{ error }}</p>
    <div class="divide-y divide-border">
      <form v-if="add.open" class="px-3 py-2 flex items-center gap-2 bg-muted/40" @submit.prevent="addWord">
        <input v-model="add.word" class="h-8 w-32 rounded-md border border-input bg-background px-2" placeholder="Chữ trong sách" />
        <span class="text-muted-foreground">→</span>
        <input v-model="add.reading" class="h-8 flex-1 min-w-0 rounded-md border border-input bg-background px-2" placeholder="Đọc là" />
        <Button type="button" variant="ghost" size="sm" @click="add.open = false">Huỷ</Button>
        <Button type="submit" size="sm" :disabled="!!busy">Thêm</Button>
      </form>
      <div v-for="e in rows" :key="e.word" class="px-3 py-2 flex items-center gap-2">
        <b class="w-28 truncate shrink-0">{{ e.word }}</b><span class="text-muted-foreground">→</span>
        <form v-if="editing[e.word] !== undefined" class="flex-1" @submit.prevent="save(e.word, editing[e.word])">
          <input v-model="editing[e.word]" class="h-7 w-full rounded border border-rag-amber bg-background px-2" @vue:mounted="focusSelect" @keydown.esc="delete editing[e.word]" />
        </form>
        <span v-else class="flex-1 truncate">{{ e.reading }} <span v-if="e.mine && e.default" class="text-[11px] text-muted-foreground">· chuẩn: {{ e.default }}</span></span>
        <span v-if="e.builtin && !e.mine" class="text-[11px] rounded bg-muted px-1.5 py-0.5 text-muted-foreground">có sẵn</span>
        <span v-else-if="e.builtin" class="text-[11px] rounded bg-primary/10 text-primary px-1.5 py-0.5">đã sửa</span>
        <Loader2 v-if="busy === e.word" class="w-3.5 h-3.5 animate-spin" />
        <button type="button" class="text-muted-foreground hover:text-foreground" :aria-label="'Nghe ' + e.word" @click="say(e.word, editing[e.word] ?? e.reading)">
          <Loader2 v-if="sayBusy === 'g-' + e.word" class="w-3.5 h-3.5 animate-spin" /><component :is="clip.playing.value === 'g-' + e.word ? Pause : Play" v-else class="w-3.5 h-3.5" />
        </button>
        <button type="button" class="text-muted-foreground hover:text-foreground" :aria-label="'Sửa ' + e.word" @click="editing[e.word] = e.reading"><Pencil class="w-3.5 h-3.5" /></button>
        <button v-if="e.mine" type="button" class="text-muted-foreground hover:text-destructive" :aria-label="e.builtin ? 'Về cách đọc chuẩn' : 'Xoá ' + e.word" :title="e.builtin ? 'Về cách đọc chuẩn' : 'Xoá'" @click="remove(e.word)">
          <component :is="e.builtin ? RotateCcw : Trash2" class="w-3.5 h-3.5" />
        </button>
      </div>
      <p v-if="!rows.length && !add.open" class="px-3 py-2 text-xs text-muted-foreground">{{ search ? 'Không thấy từ nào.' : 'Chưa có từ tự thêm.' }}</p>
      <button v-if="hiddenBuiltin" type="button" class="w-full px-3 py-2 text-xs text-muted-foreground text-left hover:bg-muted/50" @click="showAll = true">
        + {{ hiddenBuiltin }} từ có sẵn (CEO, HR, KPI, B2B…). Bấm để xem. Sửa từ có sẵn = ghi đè cách đọc; chỉ xoá được từ tự thêm.
      </button>
    </div>
  </div>
</template>
