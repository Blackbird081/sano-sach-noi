<script setup lang="ts">
// Màn MCP · Kết nối AI (wireframe D21): kết nối trong máy (bật sẵn), kết nối từ xa (chưa có, "Sắp có"),
// quyền AI (Xem luôn bật, Tạo và sửa bật sẵn, không có quyền xoá), việc AI làm gần đây (7 ngày, chỉ trên máy).
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Ban, Bot, Check, Copy, Eye, Globe, History, Laptop, Loader2, PenLine, Plug, TriangleAlert } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { addToClaudeDesktop, copyText, errText, setMCPAllowEdit, setMCPLocal, type MCPInfo } from '../lib/backend'
import { mcp, refreshMCPInfo } from '../lib/mcp'

type Client = 'claude-desktop' | 'claude-code' | 'codex' | 'other'
const clients: { key: Client; label: string }[] = [
  { key: 'claude-desktop', label: 'Claude Desktop' },
  { key: 'claude-code', label: 'Claude Code' },
  { key: 'codex', label: 'Codex' },
  { key: 'other', label: 'Phần mềm khác' },
]
const client = ref<Client>('claude-desktop')
const info = computed(() => mcp.info)
const error = ref('')
const busy = ref('')
const copied = ref('')

const command = computed(() => {
  const i = info.value
  if (!i) return ''
  return client.value === 'claude-code' ? i.claudeCode : client.value === 'codex' ? i.codex : i.configJson
})

async function act(key: string, fn: () => Promise<MCPInfo>) {
  error.value = ''
  busy.value = key
  try {
    mcp.info = await fn()
  } catch (e) {
    error.value = errText(e)
  } finally {
    busy.value = ''
  }
}

async function copy(key: string, text: string) {
  if (await copyText(text)) {
    copied.value = key
    setTimeout(() => copied.value === key && (copied.value = ''), 1500)
  }
}

function fmtWhen(at: number) {
  const d = new Date(at * 1000)
  const now = new Date()
  const hm = d.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })
  if (d.toDateString() === now.toDateString()) return hm
  const y = new Date(now)
  y.setDate(now.getDate() - 1)
  if (d.toDateString() === y.toDateString()) return 'hôm qua ' + hm
  return d.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' }) + ' ' + hm
}

// Chữ "1 phần mềm đang dùng" cập nhật theo giờ, không chỉ khi có dòng nhật ký mới.
let timer = 0
onMounted(() => {
  void refreshMCPInfo()
  timer = window.setInterval(() => void refreshMCPInfo(), 60_000)
})
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <section class="flex-1 overflow-auto p-6">
    <div class="max-w-2xl">
      <h1 class="text-xl font-semibold tracking-tight">MCP · Kết nối AI</h1>
      <p class="mt-2 mb-4 text-sm text-muted-foreground">Cho AI xem, tạo, sửa sách trong Sano bằng lời nói thường (chuẩn MCP). Sano phải đang mở thì AI mới làm được. Việc tạo sách (render) vẫn chạy trên máy này.</p>
      <p v-if="error" class="mb-3 text-sm text-destructive">{{ error }}</p>

      <div v-if="!info" class="text-sm text-muted-foreground flex items-center gap-2"><Loader2 class="w-4 h-4 animate-spin" /> Đang tải…</div>
      <div v-else class="rounded-lg border border-border text-sm divide-y divide-border">
        <!-- Trong máy -->
        <div class="px-4 py-3">
          <div class="flex items-center justify-between gap-4">
            <span class="flex items-start gap-3">
              <Laptop class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
              <span>Kết nối trong máy
                <span class="block text-xs text-muted-foreground">Phần mềm AI cài trên máy này: Claude Desktop, Claude Code, Codex, Cursor… Không ra internet.</span>
              </span>
            </span>
            <span class="shrink-0 flex items-center gap-2">
              <span v-if="info.localOn && info.running" class="text-xs text-emerald-600 dark:text-emerald-400 flex items-center gap-1">
                <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>{{ info.active.length ? `${info.active.join(', ')} đang dùng` : 'Sẵn sàng' }}</span>
              <span v-else-if="info.localOn" class="text-xs text-destructive">Không mở được cổng</span>
              <button role="switch" :aria-checked="info.localOn" aria-label="Kết nối trong máy" :disabled="busy === 'local'"
                class="h-5 w-9 rounded-full p-0.5 transition-colors disabled:opacity-60" :class="info.localOn ? 'bg-primary' : 'bg-muted-foreground/30'"
                @click="act('local', () => setMCPLocal(!info!.localOn))">
                <span class="block h-4 w-4 rounded-full bg-white shadow transition-transform" :class="info.localOn ? 'translate-x-4' : ''"></span>
              </button>
            </span>
          </div>
          <div v-if="info.localOn" class="mt-3 ml-7">
            <div class="inline-flex rounded-md bg-muted p-0.5">
              <button v-for="c in clients" :key="c.key" class="h-7 px-2.5 rounded text-xs"
                :class="client === c.key ? 'bg-background text-foreground shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'"
                @click="client = c.key">{{ c.label }}</button>
            </div>
            <p v-if="!info.bridgeFound" class="mt-2.5 flex items-start gap-1.5 text-xs text-amber-700 dark:text-amber-400">
              <TriangleAlert class="w-3.5 h-3.5 mt-0.5 shrink-0" /> Bản Sano đang chạy chưa kèm cầu nối sano-mcp (bản dev). Bản phát hành sẽ có sẵn.</p>
            <div v-if="client === 'claude-desktop'" class="mt-2.5 flex items-center justify-between gap-3 rounded-md border border-border bg-muted/30 px-3 py-2.5">
              <span class="text-xs text-muted-foreground">
                <template v-if="info.claudeDesktop === 'added'"><span class="text-emerald-600 dark:text-emerald-400 font-medium">Đã thêm vào Claude Desktop.</span> Thoát hẳn Claude Desktop (⌘Q) rồi mở lại là dùng được.</template>
                <template v-else-if="info.claudeDesktop === 'found'">Bấm nút, Sano thêm mình vào cấu hình Claude Desktop (giữ nguyên cấu hình khác, có sao lưu). Xong thoát hẳn Claude rồi mở lại.</template>
                <template v-else>Chưa thấy Claude Desktop trên máy này. Cài Claude Desktop rồi quay lại đây.</template>
              </span>
              <Button size="sm" class="shrink-0" :disabled="!info.claudeDesktop || !info.bridgeFound || busy === 'claude'" @click="act('claude', addToClaudeDesktop)">
                <Loader2 v-if="busy === 'claude'" class="w-3.5 h-3.5 animate-spin" /><Plug v-else class="w-3.5 h-3.5" />
                {{ info.claudeDesktop === 'added' ? 'Thêm lại' : 'Thêm vào Claude Desktop' }}</Button>
            </div>
            <div v-else class="mt-2.5">
              <div class="text-xs text-muted-foreground mb-1.5">{{ client === 'other' ? 'Dán vào phần cấu hình MCP của phần mềm (chạy qua stdio):' : 'Dán lệnh này vào Terminal, chạy một lần:' }}</div>
              <div class="flex items-start gap-2 rounded-md border border-border bg-muted/40 px-3 py-2">
                <pre class="flex-1 min-w-0 whitespace-pre-wrap break-all font-mono text-xs leading-relaxed select-all">{{ command }}</pre>
                <button class="shrink-0 h-7 px-2 rounded text-xs flex items-center gap-1 text-muted-foreground hover:bg-background hover:text-foreground" @click="copy(client, command)">
                  <component :is="copied === client ? Check : Copy" class="w-3.5 h-3.5" />{{ copied === client ? 'Đã chép' : 'Chép' }}</button>
              </div>
            </div>
            <p class="mt-2 text-xs text-muted-foreground">Xong thử hỏi AI: <span class="text-foreground">"Trong Sano có những sách nào?"</span></p>
          </div>
          <p v-else class="mt-2 ml-7 text-xs text-muted-foreground">Đang tắt. Phần mềm AI trên máy không kết nối được Sano.</p>
        </div>

        <!-- Từ xa: chưa có máy chủ trung chuyển -->
        <div class="px-4 py-3">
          <div class="flex items-center justify-between gap-4">
            <span class="flex items-start gap-3">
              <Globe class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
              <span>Kết nối từ xa <span class="ml-1 rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground align-middle">Sắp có</span>
                <span class="block text-xs text-muted-foreground">Dùng Claude trên điện thoại, claude.ai, ChatGPT hay máy khác, qua trạm sanobook.com. Máy này phải bật và mở Sano.</span>
              </span>
            </span>
            <button role="switch" aria-checked="false" aria-label="Kết nối từ xa" disabled class="shrink-0 h-5 w-9 rounded-full p-0.5 bg-muted-foreground/30 opacity-50 cursor-not-allowed">
              <span class="block h-4 w-4 rounded-full bg-white shadow"></span>
            </button>
          </div>
        </div>

        <!-- AI được làm gì -->
        <div class="px-4 py-3">
          <span class="flex items-start gap-3">
            <Bot class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
            <span class="flex-1">AI được làm gì
              <span class="block text-xs text-muted-foreground">Áp cho cả kết nối trong máy và từ xa. Trước khi tạo sách, bạn phải tick cam kết trên app.</span>
            </span>
          </span>
          <div class="mt-2 ml-7 grid grid-cols-3 gap-2 text-xs">
            <label class="rounded-md border border-border px-3 py-2 flex items-start gap-2 opacity-80">
              <input type="checkbox" checked disabled class="mt-0.5 accent-[hsl(var(--primary))]" />
              <span><span class="flex items-center gap-1 font-medium"><Eye class="w-3.5 h-3.5" /> Xem sách</span><span class="text-muted-foreground">Mục lục, lời đọc, giọng. Luôn bật</span></span>
            </label>
            <label class="rounded-md border border-border px-3 py-2 flex items-start gap-2 cursor-pointer">
              <input type="checkbox" :checked="info.allowEdit" :disabled="busy === 'edit'" class="mt-0.5 accent-[hsl(var(--primary))]"
                @change="act('edit', () => setMCPAllowEdit(!info!.allowEdit))" />
              <span><span class="flex items-center gap-1 font-medium"><PenLine class="w-3.5 h-3.5" /> Tạo và sửa sách</span><span class="text-muted-foreground">Tạo sách nói (sửa lời, đổi giọng sẽ có sau)</span></span>
            </label>
            <div class="rounded-md border border-dashed border-border px-3 py-2 flex items-start gap-2 text-muted-foreground">
              <Ban class="w-3.5 h-3.5 mt-0.5 shrink-0" />
              <span><span class="font-medium text-foreground">Không xoá được</span><br />AI không có lệnh xoá sách hay file nào. Xoá sách chỉ làm trong app</span>
            </div>
          </div>
          <p class="mt-2 ml-7 text-xs text-muted-foreground">AI chỉ chạm được sách trong thư viện Sano, không đọc hay ghi file nào khác trên máy.</p>
        </div>

        <!-- Nhật ký -->
        <div class="px-4 py-3">
          <span class="flex items-start gap-3">
            <History class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
            <span class="flex-1">Việc AI làm gần đây
              <span class="block text-xs text-muted-foreground">Chỉ lưu trên máy, giữ 7 ngày.</span>
            </span>
          </span>
          <div v-if="info.log.length" class="mt-2 ml-7 space-y-1.5 text-xs">
            <div v-for="(l, i) in info.log" :key="i" class="flex items-baseline gap-2">
              <span class="tabular-nums text-muted-foreground w-[4.5rem] shrink-0">{{ fmtWhen(l.at) }}</span>
              <component :is="l.edit ? PenLine : Eye" class="w-3 h-3 text-muted-foreground shrink-0 self-center" />
              <span class="font-medium shrink-0">{{ l.client }}</span>
              <span class="min-w-0 truncate" :class="l.error ? 'text-destructive' : 'text-muted-foreground'" :title="l.error ? `${l.text}: ${l.error}` : l.text">{{ l.text }}<template v-if="l.error"> · lỗi: {{ l.error }}</template></span>
            </div>
          </div>
          <p v-else class="mt-2 ml-7 text-xs text-muted-foreground">Chưa có phần mềm AI nào kết nối.</p>
        </div>
      </div>
    </div>
  </section>
</template>
