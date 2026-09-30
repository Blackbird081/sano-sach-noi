<script setup lang="ts">
// Popup cam kết trước khi render (wireframe D19), cũng dùng trước khi tạo video. Tick từng cam kết + đồng ý Điều khoản mới bật
// nút "Cam kết và render". Esc, bấm nền hoặc "Để sau" để đóng.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { BookOpen, Check, ChevronRight, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import TermsDialog from './TermsDialog.vue'
import { PLEDGES } from '../lib/pledge'

withDefaults(defineProps<{ fileName: string; title?: string; label?: string; action?: string }>(), {
  title: 'Cam kết trước khi tạo sách nói',
  label: 'Tài liệu',
  action: 'Cam kết và render',
})
const emit = defineEmits<{ close: []; confirm: [] }>()

const checked = ref<boolean[]>(PLEDGES.map(() => false))
const readTerms = ref(false)
const showTerms = ref(false)
const allOk = computed(() => checked.value.every(Boolean) && readTerms.value)
const count = computed(() => checked.value.filter(Boolean).length + (readTerms.value ? 1 : 0))

const onKey = (e: KeyboardEvent) => e.key === 'Escape' && !showTerms.value && emit('close')
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="absolute inset-0 z-20 flex items-center justify-center bg-background/70 backdrop-blur-sm p-4" @click.self="emit('close')">
    <div role="dialog" aria-modal="true" aria-labelledby="pledge-title" class="flex max-h-full w-[640px] max-w-full flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl">
      <div class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
        <div class="min-w-0">
          <h2 id="pledge-title" class="font-semibold">{{ title }}</h2>
          <p class="mt-0.5 truncate text-xs text-muted-foreground">{{ label }}: <span class="font-medium text-foreground">{{ fileName }}</span> · tick từng ô để tiếp tục</p>
        </div>
        <button class="text-muted-foreground hover:text-foreground" aria-label="Đóng" @click="emit('close')"><X class="h-4 w-4" /></button>
      </div>

      <div class="flex-1 min-h-0 overflow-auto px-5 py-3 space-y-1.5">
        <label v-for="(p, i) in PLEDGES" :key="p.title" class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2.5 transition-colors"
          :class="checked[i] ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/50'">
          <input v-model="checked[i]" type="checkbox" class="mt-0.5 h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
          <component :is="p.icon" class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
          <span class="min-w-0">
            <span class="block text-sm font-medium">{{ p.title }}</span>
            <span class="block text-xs leading-relaxed text-muted-foreground">{{ p.desc }}</span>
          </span>
        </label>
        <label class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2.5 transition-colors" :class="readTerms ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/50'">
          <input v-model="readTerms" type="checkbox" class="mt-0.5 h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
          <BookOpen class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
          <span class="text-sm">Tôi đã đọc và đồng ý <button type="button" class="font-medium text-primary underline underline-offset-2" @click.prevent="showTerms = true">Điều khoản sử dụng</button> của Sano.</span>
        </label>
      </div>

      <div class="border-t border-border px-5 py-3">
        <p class="text-[11px] leading-relaxed text-muted-foreground">Sano ghi lại thời điểm bạn cam kết vào thông tin cuốn sách. Vi phạm các cam kết trên, bạn tự chịu trách nhiệm trước pháp luật.</p>
        <div class="mt-3 flex items-center justify-between gap-3">
          <span class="text-xs" :class="allOk ? 'text-rag-green flex items-center gap-1' : 'text-muted-foreground'"><Check v-if="allOk" class="h-3.5 w-3.5" /> Đã tick {{ count }}/{{ PLEDGES.length + 1 }}</span>
          <div class="flex gap-2">
            <Button variant="outline" @click="emit('close')">Để sau</Button>
            <Button :disabled="!allOk" @click="emit('confirm')">{{ action }} <ChevronRight class="w-4 h-4" /></Button>
          </div>
        </div>
      </div>
    </div>
    <TermsDialog v-if="showTerms" @close="showTerms = false" />
  </div>
</template>
