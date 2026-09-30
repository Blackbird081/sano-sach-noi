<script setup lang="ts">
// Popup cam kết để lưu sách AI tạo qua MCP (wireframe D22): mọi cuốn của AI trong một popup (đã tạo
// xong ở khu chờ, đang tạo, chờ tạo). Tick cuốn muốn lưu + 6 ô cam kết D19 một lần. Cuốn đang / chờ tạo
// đã tick thì tự lưu khi xong. "Để sau", X, Esc, bấm ra ngoài: chỉ đóng, sách vẫn ở khu chờ 7 ngày.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { BookOpen, Check, CheckCircle2, Clock, Hourglass, Loader2, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import TermsDialog from './TermsDialog.vue'
import { PLEDGES } from '../lib/pledge'
import { closePledge, commitJobs, discardJob, mcp, pledgeJobs } from '../lib/mcp'

const picks = mcp.picks
// Cuốn mới vào danh sách: tick sẵn (cuốn đã bỏ tick trước đó giữ nguyên).
watch(pledgeJobs, (jobs) => { for (const j of jobs) if (!(j.id in picks)) picks[j.id] = true }, { immediate: true })

const checked = ref<boolean[]>(PLEDGES.map(() => false))
const readTerms = ref(false)
const showTerms = ref(false)
const confirmDrop = ref('')
const saving = ref(false)

const picked = computed(() => pledgeJobs.value.filter((j) => picks[j.id]))
const pickedDone = computed(() => picked.value.filter((j) => j.status === 'staged').length)
const pledgeOk = computed(() => checked.value.every(Boolean) && readTerms.value)
const count = computed(() => checked.value.filter(Boolean).length + (readTerms.value ? 1 : 0))
const clients = computed(() => [...new Set(pledgeJobs.value.map((j) => j.client || 'AI'))].join(', '))
const action = computed(() => {
  const n = picked.value.length
  if (!n) return 'Chọn ít nhất một cuốn'
  const later = n - pickedDone.value
  if (!later) return `Cam kết và lưu ${n} cuốn`
  return pickedDone.value ? `Cam kết · lưu ${pickedDone.value} cuốn, ${later} cuốn lưu khi tạo xong` : `Cam kết · ${later} cuốn lưu khi tạo xong`
})

const close = closePledge

async function save() {
  if (!pledgeOk.value || !picked.value.length || saving.value) return
  saving.value = true
  const ok = await commitJobs(picked.value.map((j) => j.id))
  saving.value = false
  if (ok) close()
}

async function drop(id: string) {
  if (confirmDrop.value !== id) {
    confirmDrop.value = id
    return
  }
  confirmDrop.value = ''
  delete picks[id]
  await discardJob(id)
}

const onKey = (e: KeyboardEvent) => e.key === 'Escape' && !showTerms.value && close()
onMounted(() => document.addEventListener('keydown', onKey))
onBeforeUnmount(() => document.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="absolute inset-0 z-20 flex items-center justify-center bg-background/70 backdrop-blur-sm p-4" @click.self="close">
    <div role="dialog" aria-modal="true" aria-labelledby="mcp-pledge-title" class="flex max-h-full min-h-0 w-[680px] max-w-full flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl">
      <div class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
        <div class="min-w-0">
          <h2 id="mcp-pledge-title" class="font-semibold">{{ clients }} đã tạo sách giúp bạn · cam kết để lưu vào Thư viện</h2>
          <p class="mt-0.5 text-xs text-muted-foreground">Sách đang ở khu chờ, chưa nghe, xuất hay chia sẻ được. Tick cuốn muốn lưu, rồi tick các cam kết.</p>
        </div>
        <button class="text-muted-foreground hover:text-foreground" aria-label="Để sau" @click="close"><X class="h-4 w-4" /></button>
      </div>

      <div class="flex-1 min-h-0 overflow-auto px-5 py-3">
        <div class="rounded-lg border border-border divide-y divide-border">
          <label v-for="j in pledgeJobs" :key="j.id" class="flex items-center gap-3 px-3 py-2 cursor-pointer" :class="picks[j.id] ? '' : 'opacity-60'">
            <input v-model="picks[j.id]" type="checkbox" class="h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
            <span class="flex-1 min-w-0">
              <span class="block text-sm font-medium truncate" :title="j.title">{{ j.title }}</span>
              <span class="block text-xs text-muted-foreground">{{ j.sections }} mục · {{ j.listenMin }} phút · giọng {{ j.voice }}</span>
              <span v-if="j.error" class="block text-xs text-destructive">{{ j.error }}</span>
            </span>
            <span class="shrink-0 text-xs flex items-center gap-1.5">
              <template v-if="j.status === 'staged'"><CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" /> <span class="text-emerald-700 dark:text-emerald-400">Đã tạo xong</span></template>
              <template v-else-if="j.status === 'rendering'"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" /> <span class="tabular-nums">Đang tạo {{ Math.round(j.percent) }}%</span></template>
              <template v-else><Hourglass class="w-3.5 h-3.5 text-muted-foreground" /> <span class="text-muted-foreground">Chờ tạo</span></template>
            </span>
            <button class="shrink-0 text-xs hover:text-destructive" :class="confirmDrop === j.id ? 'text-destructive font-medium' : 'text-muted-foreground'"
              @click.prevent="drop(j.id)" @blur="confirmDrop === j.id && (confirmDrop = '')">{{ confirmDrop === j.id ? 'Bỏ hẳn?' : 'Bỏ' }}</button>
          </label>
        </div>
        <p v-if="picked.length - pickedDone > 0" class="mt-1.5 text-xs text-muted-foreground flex items-center gap-1"><Clock class="w-3 h-3" /> Cuốn đang tạo / chờ tạo đã tick sẽ tự lưu khi tạo xong.</p>
        <p v-if="pledgeJobs.some((j) => !picks[j.id])" class="mt-1.5 text-xs text-muted-foreground">Cuốn không tick vẫn ở khu chờ 7 ngày, muốn bỏ hẳn bấm "Bỏ".</p>

        <div class="mt-3 text-xs font-medium text-muted-foreground">Cam kết cho {{ picked.length }} cuốn đã tick</div>
        <div class="mt-1.5 space-y-1.5">
          <label v-for="(p, i) in PLEDGES" :key="p.title" class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2 transition-colors"
            :class="checked[i] ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/50'">
            <input v-model="checked[i]" type="checkbox" class="mt-0.5 h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
            <component :is="p.icon" class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
            <span class="min-w-0">
              <span class="block text-sm font-medium">{{ p.title }}</span>
              <span class="block text-[11px] leading-relaxed text-muted-foreground">{{ p.desc }}</span>
            </span>
          </label>
          <label class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2 transition-colors" :class="readTerms ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/50'">
            <input v-model="readTerms" type="checkbox" class="mt-0.5 h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
            <BookOpen class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
            <span class="text-sm">Tôi đã đọc và đồng ý <button type="button" class="font-medium text-primary underline underline-offset-2" @click.prevent="showTerms = true">Điều khoản sử dụng</button> của Sano.</span>
          </label>
        </div>
      </div>

      <div class="border-t border-border px-5 py-3">
        <p class="text-[11px] leading-relaxed text-muted-foreground">Sano ghi thời điểm cam kết vào thông tin từng cuốn. Vi phạm các cam kết trên, bạn tự chịu trách nhiệm trước pháp luật.</p>
        <p v-if="mcp.error" class="mt-1 text-xs text-destructive">{{ mcp.error }}</p>
        <div class="mt-3 flex items-center justify-between gap-3">
          <span class="text-xs" :class="pledgeOk ? 'text-rag-green flex items-center gap-1' : 'text-muted-foreground'"><Check v-if="pledgeOk" class="h-3.5 w-3.5" /> Đã tick {{ count }}/{{ PLEDGES.length + 1 }}</span>
          <div class="flex gap-2">
            <Button variant="outline" @click="close">Để sau</Button>
            <Button :disabled="!pledgeOk || !picked.length || saving" @click="save"><Loader2 v-if="saving" class="w-4 h-4 animate-spin" />{{ action }}</Button>
          </div>
        </div>
      </div>
    </div>
    <TermsDialog v-if="showTerms" @close="showTerms = false" />
  </div>
</template>
