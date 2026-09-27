<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2, RefreshCw, Trash2 } from 'lucide-vue-next'
import { clearListenLog, errText, isDesktop, librarySize, listenLog, openLibraryFolder } from '../lib/backend'
import { checkForUpdate, refreshTTS, setAutoUpdateCheck, state } from '../lib/store'
import TtsSettings from '../components/TtsSettings.vue'

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
