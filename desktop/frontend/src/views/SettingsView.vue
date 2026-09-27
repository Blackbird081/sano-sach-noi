<script setup lang="ts">
import GlobalDictSettings from '../components/GlobalDictSettings.vue'
import { onMounted, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Check, Loader2, RefreshCw, Trash2 } from 'lucide-vue-next'
import { clearListenLog, errText, isDesktop, librarySize, listenLog, openLibraryFolder } from '../lib/backend'
import { checkForUpdate, refreshTTS, setAutoUpdateCheck, state } from '../lib/store'
import TtsSettings from '../components/TtsSettings.vue'
import PauseCustomInputs from '../components/PauseCustomInputs.vue'
import { PAUSE_PRESETS, fmtGap, makeSetting, pauseState, setGlobalPause, type Gaps, type PauseLevel } from '../lib/pause'

// Quãng nghỉ chung (wireframe D8): bấm là lưu, báo đã lưu sau lần đổi đầu.
const pauseSaved = ref(false)
function pickPause(level: PauseLevel) {
  const g = pauseState.global
  setGlobalPause(level === 'custom' ? makeSetting('custom', { section: g.section, chapter: g.chapter }) : makeSetting(level))
  pauseSaved.value = true
}
function saveCustomPause(g: Gaps) {
  setGlobalPause(makeSetting('custom', g))
  pauseSaved.value = true
}

const size = ref<number | null>(null)
// Số liệu Hành trình nghe: số ngày đã ghi, xoá = chuyển file vào Thùng rác
const listenDays = ref<number | null>(null)
const listenMsg = ref('')
async function loadListenDays() {
  try {
    listenDays.value = Object.keys((await listenLog()).days).length
  } catch {
    listenDays.value = null
  }
}
async function clearListen() {
  listenMsg.value = ''
  try {
    const where = await clearListenLog()
    if (where) listenMsg.value = where === 'Thùng rác' ? 'Đã chuyển số liệu nghe vào Thùng rác.' : `Đã dời số liệu nghe tới ${where}.`
    await loadListenDays()
  } catch (e) {
    listenMsg.value = errText(e)
  }
}

onMounted(async () => {
  if (!state.tts && !state.ttsChecking) refreshTTS()
  void loadListenDays()
  try {
    size.value = await librarySize()
  } catch {
    size.value = null
  }
})

function humanSize(n: number) {
  if (n < 1 << 20) return `${Math.max(1, Math.round(n / 1024))} KB`
  if (n < 1 << 30) return `${Math.round(n / (1 << 20))} MB`
  return `${(n / (1 << 30)).toFixed(1).replace('.', ',')} GB`
}

const updateLine: Record<string, string> = {
  idle: '',
  checking: 'Đang kiểm tra…',
  latest: 'Đang dùng bản mới nhất',
  error: 'Chưa kiểm tra được bản mới',
}
</script>

<template>
  <section class="flex-1 overflow-auto p-6 max-w-2xl">
    <h1 class="text-xl font-semibold tracking-tight">Cài đặt</h1>
    <div class="mt-6 space-y-6">
      <div>
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Lưu trữ</h2>
        <div class="rounded-lg border border-border divide-y divide-border text-sm">
          <div class="flex items-center justify-between gap-4 px-4 py-3"><span class="shrink-0">Thư mục lưu sách</span><span class="flex items-center gap-2 text-muted-foreground min-w-0"><span class="truncate">{{ state.library?.dir || '~/Sano/Sach' }}</span> <Button variant="outline" size="sm" @click="openLibraryFolder()">Mở</Button></span></div>
          <div class="flex items-center justify-between px-4 py-3"><span>Dung lượng sách đã tạo</span><span class="text-muted-foreground tabular-nums">{{ size === null ? '—' : humanSize(size) }}</span></div>
        </div>
      </div>

      <div>
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Nghe và xuất M4B</h2>
        <div class="rounded-lg border border-border divide-y divide-border text-sm">
          <div class="flex items-center justify-between gap-4 px-4 py-3">
            <span>Quãng nghỉ giữa các phần
              <span class="block text-xs text-muted-foreground">Nghỉ {{ fmtGap(pauseState.global.section) }} giữa các tiểu mục, {{ fmtGap(pauseState.global.chapter) }} khi sang chương mới. Mức mặc định cho mọi cuốn, dùng cả khi nghe trong Sano và Xuất M4B. Muốn khác cho từng cuốn: chỉnh ở nút Nghỉ trong màn nghe.</span>
            </span>
            <div role="radiogroup" aria-label="Quãng nghỉ giữa các phần" class="shrink-0 inline-flex rounded-md bg-muted p-0.5">
              <button v-for="p in [...PAUSE_PRESETS, { level: 'custom' as const, label: 'Tuỳ chỉnh' }]" :key="p.level" role="radio" :aria-checked="pauseState.global.level === p.level"
                class="h-8 px-3 rounded text-sm"
                :class="pauseState.global.level === p.level ? 'bg-background text-foreground shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'"
                @click="pickPause(p.level)">{{ p.label }}</button>
            </div>
          </div>
          <PauseCustomInputs v-if="pauseState.global.level === 'custom'" class="px-4 py-3" :value="pauseState.global" @save="saveCustomPause" />
          <div v-if="pauseSaved" class="px-4 py-2.5 text-xs text-muted-foreground flex items-start gap-2">
            <Check class="w-3.5 h-3.5 mt-0.5 shrink-0" />
            <span>Đã lưu. Áp dụng ngay khi nghe trong Sano. Cuốn nào đã chọn riêng ở màn nghe thì giữ mức riêng. File M4B đã xuất trước đây giữ quãng nghỉ cũ, muốn đổi thì xuất lại.</span>
          </div>
        </div>
      </div>

      <div>
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Từ điển chung · dùng cho mọi sách</h2>
        <GlobalDictSettings />
        <p class="mt-2 text-xs text-muted-foreground">Đổi từ điển chung chỉ áp cho sách tạo sau. Sách đã tạo giữ cách đọc cũ; muốn đổi thì thêm từ đó ở Sửa sách → tab Từ điển của cuốn, rồi Lưu & đọc lại.</p>
      </div>

      <div>
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Dữ liệu nghe</h2>
        <div class="rounded-lg border border-border text-sm">
          <div class="flex items-center justify-between gap-4 px-4 py-3">
            <span>Số liệu Hành trình nghe
              <span class="block text-xs text-muted-foreground">{{ listenDays ? `Đã ghi ${listenDays} ngày nghe` : 'Chưa có số liệu' }} · chỉ lưu trên máy, trong ~/Sano/.nghe.json. Xoá thì sách và vị trí đang nghe vẫn giữ.</span>
              <span v-if="listenMsg" class="block text-xs text-muted-foreground mt-0.5">{{ listenMsg }}</span>
            </span>
            <Button variant="outline" size="sm" class="shrink-0" :disabled="!listenDays || !isDesktop()" @click="clearListen"><Trash2 class="w-3.5 h-3.5" /> Xoá số liệu</Button>
          </div>
        </div>
      </div>

      <div>
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Bộ đọc</h2>
        <TtsSettings />
      </div>

      <div>
        <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Ứng dụng</h2>
        <div class="rounded-lg border border-border divide-y divide-border text-sm">
          <div class="flex items-center justify-between px-4 py-3"><span>Giao diện</span><span class="text-muted-foreground">Theo hệ điều hành</span></div>
          <label class="flex items-center justify-between px-4 py-3 cursor-pointer">
            <span>Tự kiểm tra bản mới khi mở Sano <span class="block text-xs text-muted-foreground">Chỉ hỏi GitHub số phiên bản mới nhất, không gửi dữ liệu nào của bạn</span></span>
            <input type="checkbox" :checked="state.autoUpdateCheck" class="h-4 w-4 accent-[hsl(var(--primary))]" @change="setAutoUpdateCheck(($event.target as HTMLInputElement).checked)" />
          </label>
          <div class="flex items-center justify-between px-4 py-3">
            <span>Phiên bản {{ state.version }} <span v-if="updateLine[state.updateCheck]" class="block text-xs text-muted-foreground">{{ updateLine[state.updateCheck] }}</span></span>
            <span v-if="state.updateInfo" class="flex items-center gap-2"><Badge class="bg-rag-amber/15 text-rag-amber border-0">Có bản {{ state.updateInfo.version }}</Badge><Button size="sm" @click="state.update = 'info'">Xem</Button></span>
            <Button v-else variant="outline" size="sm" :disabled="state.updateCheck === 'checking'" @click="checkForUpdate()"><Loader2 v-if="state.updateCheck === 'checking'" class="w-3.5 h-3.5 animate-spin" />Kiểm tra bản mới</Button>
          </div>
        </div>
      </div>

    </div>
  </section>
</template>
