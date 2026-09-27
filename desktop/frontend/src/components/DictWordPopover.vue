<script setup lang="ts">
// Hộp "Dạy Sano đọc <từ>" (wireframe D12): gõ cách đọc, nghe thử, chọn phạm vi
// (chỉ cuốn này / mọi sách). Dùng ở Nạp file, Nghe thử và Sửa sách.
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { Loader2, Pause, Play } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { errText, speakSample } from '../lib/backend'
import { useClipPlayer } from '../lib/audio'

const props = withDefaults(
  defineProps<{
    word: string
    voice: string
    reading?: string
    count?: number | null // số chỗ trong sách; null = chưa đếm xong
    confirmLabel?: string
    scopes?: boolean // hiện chọn Chỉ cuốn này / Mọi sách
    busy?: boolean
    error?: string
  }>(),
  { reading: '', count: null, confirmLabel: 'Thêm', scopes: true, busy: false, error: '' },
)
const emit = defineEmits<{ add: [reading: string, scope: 'book' | 'global']; cancel: [] }>()

const reading = ref(props.reading)
const scope = ref<'book' | 'global'>('book')
const input = ref<HTMLInputElement | null>(null)
const clip = useClipPlayer()
const loading = ref(false)
const listenError = ref('')

async function listen() {
  if (clip.playing.value === 'w') return clip.stop()
  if (!reading.value.trim()) return
  listenError.value = ''
  loading.value = true
  try {
    await clip.toggle('w', await speakSample(props.voice, reading.value.trim()))
  } catch (e) {
    listenError.value = errText(e)
  } finally {
    loading.value = false
  }
}
function submit() {
  if (reading.value.trim() && !props.busy) emit('add', reading.value.trim(), scope.value)
}
function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('cancel')
}
const root = ref<HTMLElement | null>(null)
onMounted(() => {
  document.addEventListener('keydown', onKey)
  void nextTick(() => {
    root.value?.scrollIntoView({ block: 'nearest', behavior: 'smooth' }) // hộp mở sát mép dưới khung cuộn
    input.value?.focus()
    input.value?.select()
  })
})
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <form ref="root" class="w-80 rounded-lg border border-border bg-popover text-popover-foreground shadow-xl p-3 z-30 text-left" @submit.prevent="submit" @mousedown.stop>
    <div class="text-sm font-medium">Dạy Sano đọc "{{ word }}"</div>
    <p class="mt-0.5 text-xs text-muted-foreground">
      <template v-if="count === null">Đang đếm số chỗ trong sách…</template>
      <template v-else-if="count > 0">Có {{ count }} chỗ trong sách. Chữ hiện khi nghe vẫn là {{ word }}.</template>
      <template v-else>Gõ đúng như tai nghe, cách nhau bằng dấu cách.</template>
    </p>
    <label class="mt-2 block text-xs text-muted-foreground">Đọc là
      <input ref="input" v-model="reading" maxlength="120" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground" placeholder="vd ca pê i" />
    </label>
    <div v-if="scopes" class="mt-2 flex items-center gap-3 text-xs">
      <label class="flex items-center gap-1.5"><input v-model="scope" type="radio" value="book" class="accent-[hsl(var(--primary))]" /> Chỉ cuốn này</label>
      <label class="flex items-center gap-1.5"><input v-model="scope" type="radio" value="global" class="accent-[hsl(var(--primary))]" /> Mọi sách</label>
    </div>
    <p v-if="error || listenError" class="mt-2 text-xs text-destructive">{{ error || listenError }}</p>
    <div class="mt-3 flex items-center gap-2">
      <Button type="button" variant="outline" size="sm" :disabled="!reading.trim() || loading" @click="listen">
        <Loader2 v-if="loading" class="w-3.5 h-3.5 animate-spin" /><component :is="clip.playing.value === 'w' ? Pause : Play" v-else class="w-3.5 h-3.5" /> Nghe thử
      </Button>
      <span class="flex-1"></span>
      <Button type="button" variant="ghost" size="sm" @click="emit('cancel')">Huỷ</Button>
      <Button type="submit" size="sm" :disabled="!reading.trim() || busy"><Loader2 v-if="busy" class="w-3.5 h-3.5 animate-spin" /> {{ confirmLabel }}</Button>
    </div>
  </form>
</template>
