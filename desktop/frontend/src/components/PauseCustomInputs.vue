<script setup lang="ts">
// Hai ô số giây của mức nghỉ "Tuỳ chỉnh" (Cài đặt và màn nghe). Gõ số hợp lệ
// (0–10, nhận cả "1,2" lẫn "1.2") là lưu ngay; sai thì viền đỏ, báo lỗi, không lưu.
import { ref, watch } from 'vue'
import { MAX_GAP_SEC, parseGapSec, type Gaps } from '../lib/pause'

const props = defineProps<{ value: Gaps; compact?: boolean }>()
const emit = defineEmits<{ save: [Gaps] }>()

const show = (n: number) => n.toLocaleString('vi-VN', { maximumFractionDigits: 1 })
const section = ref(show(props.value.section))
const chapter = ref(show(props.value.chapter))
const bad = ref({ section: false, chapter: false })

watch(
  () => [props.value.section, props.value.chapter],
  ([s, c]) => {
    if (parseGapSec(section.value) !== s) section.value = show(s)
    if (parseGapSec(chapter.value) !== c) chapter.value = show(c)
  },
)

function input() {
  const s = parseGapSec(section.value)
  const c = parseGapSec(chapter.value)
  bad.value = { section: s === null, chapter: c === null }
  if (s !== null && c !== null) emit('save', { section: s, chapter: c })
}

// Rời ô: số hợp lệ viết lại gọn ("1.20" → "1,2"); số sai trả về giá trị đã lưu.
function blur() {
  section.value = show(parseGapSec(section.value) ?? props.value.section)
  chapter.value = show(parseGapSec(chapter.value) ?? props.value.chapter)
  bad.value = { section: false, chapter: false }
}

const field = 'w-16 h-8 rounded-md border bg-background px-2 text-right text-foreground tabular-nums outline-none focus-visible:ring-2 focus-visible:ring-ring'
</script>

<template>
  <div :class="compact ? 'space-y-2' : 'flex flex-wrap items-center gap-x-6 gap-y-2'">
    <label class="flex items-center gap-2" :class="compact && 'justify-between'">Giữa tiểu mục
      <span class="flex items-center gap-1.5 text-muted-foreground">
        <input v-model="section" inputmode="decimal" aria-label="Số giây nghỉ giữa tiểu mục" :aria-invalid="bad.section"
          :class="[field, bad.section ? 'border-destructive' : 'border-input']" @input="input" @blur="blur" /> giây
      </span>
    </label>
    <label class="flex items-center gap-2" :class="compact && 'justify-between'">Sang chương mới
      <span class="flex items-center gap-1.5 text-muted-foreground">
        <input v-model="chapter" inputmode="decimal" aria-label="Số giây nghỉ khi sang chương mới" :aria-invalid="bad.chapter"
          :class="[field, bad.chapter ? 'border-destructive' : 'border-input']" @input="input" @blur="blur" /> giây
      </span>
    </label>
    <p v-if="bad.section || bad.chapter" class="text-xs text-destructive" role="alert">Nhập số từ 0 đến {{ MAX_GAP_SEC }} giây, ví dụ 1,2. Chưa lưu.</p>
    <p v-else class="text-xs text-muted-foreground">Từ 0 đến {{ MAX_GAP_SEC }} giây, bước 0,1.{{ compact ? ' Gõ xong là lưu.' : '' }}</p>
  </div>
</template>
