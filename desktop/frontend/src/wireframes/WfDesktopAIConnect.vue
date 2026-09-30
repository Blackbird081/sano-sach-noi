<script setup lang="ts">
// D21 — Màn MCP (kết nối AI). Mục riêng "MCP" ở thanh bên, ngay trên Cài đặt.
// - Kết nối trong máy (local): BẬT mặc định. Claude Desktop, Claude Code, Codex, Cursor… cài
//   trên cùng máy nối vào Sano đang mở. Thẻ chọn phần mềm → nút thêm một lần / lệnh sao chép sẵn.
// - Kết nối từ xa (remote, qua sanobook.com): TẮT mặc định. Bật → popup cam kết → địa chỉ kết nối
//   + mã ghép dài (40 ký tự, chỉ chép-dán được, hết hạn sau 2 phút, dùng một lần) để thêm vào
//   Claude (web, điện thoại) / ChatGPT. Danh sách nơi đã ghép, gỡ từng nơi. Đang bật thì thanh
//   bên có dấu "Đang mở kết nối từ xa".
// - AI được làm gì: Xem luôn được; Tạo & sửa bật sẵn. KHÔNG có quyền xoá qua MCP (không có công
//   cụ xoá nào): xoá sách chỉ làm trong app. AI chỉ chạm được sách trong thư viện theo slug.
// - Việc AI làm gần đây: nhật ký ngắn (chỉ trên máy).
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, Check, Copy, Plug, Globe, Laptop,
  Bot, Eye, PenLine, Ban, X, Cable, ShieldAlert, Smartphone, RefreshCw, Radio, Coffee, History,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type Mode = 'default' | 'connected' | 'pledge' | 'remote' | 'off'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'default')
const dark = ref(false)

const localOn = computed(() => mode.value !== 'off')
const remoteOn = computed(() => mode.value === 'remote')
const hasLog = computed(() => mode.value === 'connected' || mode.value === 'remote')

type Client = 'claude-desktop' | 'claude-code' | 'codex' | 'other'
const client = ref<Client>('claude-desktop')
const clients: { key: Client; label: string }[] = [
  { key: 'claude-desktop', label: 'Claude Desktop' },
  { key: 'claude-code', label: 'Claude Code' },
  { key: 'codex', label: 'Codex' },
  { key: 'other', label: 'Phần mềm khác' },
]
const bridge = '/Applications/Sano.app/Contents/MacOS/sano-mcp'
const cmd: Record<Exclude<Client, 'claude-desktop'>, string> = {
  'claude-code': `claude mcp add sano -- ${bridge}`,
  codex: `codex mcp add sano -- ${bridge}`,
  other: `{\n  "mcpServers": {\n    "sano": { "command": "${bridge}" }\n  }\n}`,
}
const copied = ref('')
function copy(k: string) {
  copied.value = k
  setTimeout(() => (copied.value === k ? (copied.value = '') : null), 1500)
}

const allow = ref({ edit: true })
const pairCode = 'sano_pair_7F3KQ2MXR4TB8WNJC6HDP9LAE5VYG2ZU'
const keepAwake = ref(true)

const pledges = [
  'Tôi hiểu: ai có địa chỉ và đã ghép bằng mã sẽ điều khiển được Sano trên máy này (xem, tạo, sửa sách; không xoá được). Tôi không đưa mã ghép cho người khác.',
  'Tôi hiểu: nội dung sách đi qua trạm trung chuyển sanobook.com (mã hoá khi truyền, trạm không lưu lại) để tới máy này.',
  'Sách AI tạo giúp tôi vẫn theo đúng các cam kết khi render (quyền dùng tài liệu, không vi phạm pháp luật, không mạo danh, không phát tán, tự chịu trách nhiệm).',
]
const ticked = ref<boolean[]>(pledges.map(() => false))
const pledgeOk = computed(() => ticked.value.every(Boolean))

const paired = [
  { icon: Smartphone, name: 'Claude · iPhone của Việt', when: 'Ghép hôm nay 10:42 · dùng lần cuối 11:05' },
  { icon: Globe, name: 'ChatGPT · web', when: 'Ghép hôm qua · dùng lần cuối hôm qua 21:30' },
]
const log = computed(() => [
  ...(remoteOn.value ? [{ who: 'Claude · iPhone', what: 'Tạo sách "Kênh phân phối tập 4" từ lời AI viết (6 chương, 24 mục) · đang đọc 8/24', when: '11:05', kind: 'edit' }] : []),
  { who: 'Claude Code', what: 'Đọc lời mục 2–7 cuốn "Thấy ra chính mình"', when: '10:58', kind: 'view' },
  { who: 'Claude Code', what: 'Sửa lời mục "Cái bẫy ngược", đọc lại 1 mục', when: '10:51', kind: 'edit' },
  { who: 'Claude Code', what: 'Xem danh sách sách', when: '10:50', kind: 'view' },
])

function setMode(m: Mode) {
  mode.value = m
  ticked.value = pledges.map(() => m === 'remote')
}

const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách nói', icon: FilePlus2 },
  { key: 'stats', label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'mcp', label: 'MCP', icon: Cable },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const states: [Mode, string][] = [
  ['default', '1. Mặc định'],
  ['connected', '2. Claude Code đã kết nối'],
  ['pledge', '3. Bật kết nối từ xa (cam kết)'],
  ['remote', '4. Từ xa đang bật'],
  ['off', '5. Tắt cả hai'],
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in states" :key="m[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="setMode(m[0])">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span>
        <span class="h-3 w-3 rounded-full bg-rag-amber"></span>
        <span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>

      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2">
            <img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" />
            <span class="font-semibold tracking-tight">Sano</span>
          </div>
          <nav class="px-2 space-y-0.5">
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.key === 'mcp' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <!-- MỚI: dấu nhắc khi đang mở kết nối từ xa (bấm → về nhóm Kết nối AI) -->
          <div v-if="remoteOn" class="mx-3 mb-3 rounded-lg border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs">
            <span class="flex items-center gap-1.5 font-medium text-amber-700 dark:text-amber-400"><Radio class="w-3.5 h-3.5" /> Đang mở kết nối từ xa</span>
            <span class="block mt-0.5 text-muted-foreground">2 nơi đã ghép · bấm để xem</span>
          </div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.21 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <main class="flex-1 min-w-0 overflow-auto p-6">
          <div class="max-w-2xl">
            <h1 class="text-xl font-semibold tracking-tight">MCP · Kết nối AI</h1>
            <div class="mt-2 space-y-6">
              <div>
                <p class="text-sm text-muted-foreground mb-4">Cho AI xem, tạo, sửa sách trong Sano bằng lời nói thường (chuẩn MCP). Sano phải đang mở thì AI mới làm được. Việc tạo sách (render) vẫn chạy trên máy này.</p>

                <div class="rounded-lg border border-border text-sm divide-y divide-border">
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
                        <span v-if="localOn" class="text-xs text-emerald-600 dark:text-emerald-400 flex items-center gap-1"><span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>{{ hasLog ? '1 phần mềm đang dùng' : 'Sẵn sàng' }}</span>
                        <button role="switch" :aria-checked="localOn" class="h-5 w-9 rounded-full p-0.5 transition-colors" :class="localOn ? 'bg-primary' : 'bg-muted-foreground/30'"
                          @click="setMode(localOn ? 'off' : 'default')"><span class="block h-4 w-4 rounded-full bg-white shadow transition-transform" :class="localOn ? 'translate-x-4' : ''"></span></button>
                      </span>
                    </div>
                    <div v-if="localOn" class="mt-3 ml-7">
                      <div class="inline-flex rounded-md bg-muted p-0.5">
                        <button v-for="c in clients" :key="c.key" class="h-7 px-2.5 rounded text-xs"
                          :class="client === c.key ? 'bg-background text-foreground shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'"
                          @click="client = c.key">{{ c.label }}</button>
                      </div>
                      <div v-if="client === 'claude-desktop'" class="mt-2.5 flex items-center justify-between gap-3 rounded-md border border-border bg-muted/30 px-3 py-2.5">
                        <span class="text-xs text-muted-foreground">Bấm nút, Claude Desktop hỏi cài phần mở rộng "Sano" → Cài. Mở lại Claude là dùng được.</span>
                        <Button size="sm" class="shrink-0"><Plug class="w-3.5 h-3.5" /> Thêm vào Claude Desktop</Button>
                      </div>
                      <div v-else class="mt-2.5">
                        <div class="text-xs text-muted-foreground mb-1.5">{{ client === 'other' ? 'Dán vào phần cấu hình MCP của phần mềm (chạy qua stdio):' : 'Dán lệnh này vào Terminal, chạy một lần:' }}</div>
                        <div class="flex items-start gap-2 rounded-md border border-border bg-muted/40 px-3 py-2">
                          <pre class="flex-1 min-w-0 whitespace-pre-wrap break-all font-mono text-xs leading-relaxed">{{ cmd[client] }}</pre>
                          <button class="shrink-0 h-7 px-2 rounded text-xs flex items-center gap-1 text-muted-foreground hover:bg-background hover:text-foreground" @click="copy(client)">
                            <component :is="copied === client ? Check : Copy" class="w-3.5 h-3.5" />{{ copied === client ? 'Đã chép' : 'Chép' }}</button>
                        </div>
                      </div>
                      <p class="mt-2 text-xs text-muted-foreground">Xong thử hỏi AI: <span class="text-foreground">"Trong Sano có những sách nào?"</span></p>
                    </div>
                  </div>

                  <!-- Từ xa -->
                  <div class="px-4 py-3">
                    <div class="flex items-center justify-between gap-4">
                      <span class="flex items-start gap-3">
                        <Globe class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
                        <span>Kết nối từ xa
                          <span class="block text-xs text-muted-foreground">Dùng Claude trên điện thoại, claude.ai, ChatGPT hay máy khác, qua trạm sanobook.com. Máy này phải bật và mở Sano.</span>
                        </span>
                      </span>
                      <span class="shrink-0 flex items-center gap-2">
                        <span v-if="remoteOn" class="text-xs text-amber-700 dark:text-amber-400 flex items-center gap-1"><span class="h-1.5 w-1.5 rounded-full bg-amber-500"></span>Đang mở</span>
                        <button role="switch" :aria-checked="remoteOn" class="h-5 w-9 rounded-full p-0.5 transition-colors" :class="remoteOn ? 'bg-primary' : 'bg-muted-foreground/30'"
                          @click="setMode(remoteOn ? 'default' : 'pledge')"><span class="block h-4 w-4 rounded-full bg-white shadow transition-transform" :class="remoteOn ? 'translate-x-4' : ''"></span></button>
                      </span>
                    </div>
                    <div v-if="remoteOn" class="mt-3 ml-7 space-y-3">
                      <div class="rounded-md border border-border bg-muted/30 px-3 py-2.5">
                        <div class="text-xs text-muted-foreground">1. Địa chỉ kết nối (cố định cho máy này)</div>
                        <div class="mt-0.5 flex items-center gap-2">
                          <span class="font-mono text-xs flex-1 truncate">https://mcp.sanobook.com/k/7f3kq2mxr4tb</span>
                          <button class="h-7 px-2 rounded text-xs flex items-center gap-1 text-muted-foreground hover:bg-background hover:text-foreground" @click="copy('url')">
                            <component :is="copied === 'url' ? Check : Copy" class="w-3.5 h-3.5" />{{ copied === 'url' ? 'Đã chép' : 'Chép' }}</button>
                        </div>
                      </div>
                      <div class="rounded-md border border-primary/30 bg-primary/5 px-3 py-2.5">
                        <div class="flex items-center justify-between text-xs text-muted-foreground">
                          <span>2. Mã ghép, dán vào trang ghép khi Claude / ChatGPT hỏi</span>
                          <span class="tabular-nums">dùng một lần · còn <b class="text-foreground">1:48</b></span>
                        </div>
                        <div class="mt-1 flex items-center gap-2">
                          <span class="font-mono text-[13px] flex-1 break-all select-all text-primary font-medium">{{ pairCode }}</span>
                          <Button size="sm" class="shrink-0" @click="copy('code')"><component :is="copied === 'code' ? Check : Copy" class="w-3.5 h-3.5" />{{ copied === 'code' ? 'Đã chép' : 'Chép mã' }}</Button>
                        </div>
                        <div class="mt-1 text-[11px] text-muted-foreground flex items-center gap-1">Hết 2 phút thì mã tự huỷ, bấm <span class="text-primary flex items-center gap-0.5"><RefreshCw class="w-3 h-3" /> Tạo mã mới</span></div>
                      </div>
                      <div class="text-xs text-muted-foreground leading-relaxed">
                        <b class="text-foreground font-medium">Claude</b> (web, điện thoại): Settings → Connectors → Add custom connector → dán địa chỉ → trang ghép hỏi mã → dán mã ghép.
                        <b class="text-foreground font-medium">ChatGPT</b>: Settings → Connectors → Developer mode → Create → dán địa chỉ, rồi dán mã ghép.
                        <a class="text-primary underline underline-offset-2">Xem hướng dẫn có ảnh</a>
                      </div>
                      <div>
                        <div class="text-xs font-medium mb-1">Nơi đã ghép</div>
                        <div class="rounded-md border border-border divide-y divide-border">
                          <div v-for="p in paired" :key="p.name" class="flex items-center gap-3 px-3 py-2">
                            <component :is="p.icon" class="w-4 h-4 text-muted-foreground shrink-0" />
                            <span class="flex-1 min-w-0">{{ p.name }}<span class="block text-xs text-muted-foreground">{{ p.when }}</span></span>
                            <Button variant="outline" size="sm">Gỡ</Button>
                          </div>
                        </div>
                      </div>
                      <label class="flex items-center gap-2 text-xs"><input v-model="keepAwake" type="checkbox" class="accent-primary" /><Coffee class="w-3.5 h-3.5 text-muted-foreground" /> Giữ máy không ngủ khi đang mở kết nối từ xa</label>
                    </div>
                    <p v-else class="mt-2 ml-7 text-xs text-muted-foreground">Đang tắt. Bật lên thì các nơi ghép trước đây phải nhập mã ghép lại.</p>
                  </div>

                  <!-- AI được làm gì -->
                  <div class="px-4 py-3">
                    <span class="flex items-start gap-3">
                      <Bot class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
                      <span class="flex-1">AI được làm gì
                        <span class="block text-xs text-muted-foreground">Áp cho cả kết nối trong máy và từ xa. Trước khi tạo sách, AI luôn hỏi bạn trong khung trò chuyện.</span>
                      </span>
                    </span>
                    <div class="mt-2 ml-7 grid grid-cols-3 gap-2 text-xs">
                      <label class="rounded-md border border-border px-3 py-2 flex items-start gap-2 opacity-80">
                        <input type="checkbox" checked disabled class="mt-0.5 accent-primary" />
                        <span><span class="flex items-center gap-1 font-medium"><Eye class="w-3.5 h-3.5" /> Xem sách</span><span class="text-muted-foreground">Mục lục, lời đọc, giọng. Luôn bật</span></span>
                      </label>
                      <label class="rounded-md border border-border px-3 py-2 flex items-start gap-2">
                        <input v-model="allow.edit" type="checkbox" class="mt-0.5 accent-primary" />
                        <span><span class="flex items-center gap-1 font-medium"><PenLine class="w-3.5 h-3.5" /> Tạo và sửa sách</span><span class="text-muted-foreground">Tạo, sửa lời, đọc lại, đổi giọng, bìa, xuất M4B</span></span>
                      </label>
                      <div class="rounded-md border border-dashed border-border px-3 py-2 flex items-start gap-2 text-muted-foreground">
                        <Ban class="w-3.5 h-3.5 mt-0.5 shrink-0" />
                        <span><span class="font-medium text-foreground">Không xoá được</span><br />AI không có lệnh xoá sách hay file nào. Xoá sách chỉ làm trong app</span>
                      </div>
                    </div>
                    <p class="mt-2 ml-7 text-xs text-muted-foreground">AI chỉ chạm được sách trong thư viện Sano, không đọc hay ghi file nào khác trên máy. Sửa lời, đổi giọng thì Sano giữ bản trước để hoàn tác.</p>
                  </div>

                  <!-- Nhật ký -->
                  <div v-if="localOn || remoteOn" class="px-4 py-3">
                    <span class="flex items-start gap-3">
                      <History class="w-4 h-4 mt-0.5 text-muted-foreground shrink-0" />
                      <span class="flex-1">Việc AI làm gần đây
                        <span class="block text-xs text-muted-foreground">Chỉ lưu trên máy, giữ 7 ngày.</span>
                      </span>
                    </span>
                    <div v-if="hasLog" class="mt-2 ml-7 space-y-1.5 text-xs">
                      <div v-for="(l, i) in log" :key="i" class="flex items-baseline gap-2">
                        <span class="tabular-nums text-muted-foreground w-10 shrink-0">{{ l.when }}</span>
                        <component :is="l.kind === 'edit' ? PenLine : Eye" class="w-3 h-3 text-muted-foreground shrink-0 self-center" />
                        <span class="font-medium shrink-0">{{ l.who }}</span>
                        <span class="text-muted-foreground min-w-0 truncate">{{ l.what }}</span>
                      </div>
                    </div>
                    <p v-else class="mt-2 ml-7 text-xs text-muted-foreground">Chưa có phần mềm AI nào kết nối.</p>
                  </div>
                </div>
              </div>

              <div>
              </div>
            </div>
          </div>
        </main>
      </div>

      <!-- Popup bật kết nối từ xa -->
      <div v-if="mode === 'pledge'" class="absolute inset-0 top-9 z-20 grid place-items-center bg-background/70 backdrop-blur-sm">
        <div role="dialog" aria-modal="true" class="w-[520px] rounded-xl border border-border bg-background shadow-2xl">
          <div class="flex items-start gap-3 px-5 pt-5">
            <span class="h-9 w-9 rounded-full bg-amber-500/15 grid place-items-center shrink-0"><ShieldAlert class="w-5 h-5 text-amber-600" /></span>
            <span class="flex-1">
              <span class="block font-semibold">Bật kết nối từ xa</span>
              <span class="block text-sm text-muted-foreground mt-0.5">AI trên điện thoại, web sẽ điều khiển được Sano trên máy này qua internet. Tick từng ý để bật.</span>
            </span>
            <button class="text-muted-foreground hover:text-foreground" aria-label="Đóng" @click="setMode('default')"><X class="w-4 h-4" /></button>
          </div>
          <div class="px-5 py-4 space-y-2">
            <label v-for="(p, i) in pledges" :key="i" class="flex items-start gap-3 rounded-lg border px-3 py-2.5 text-sm cursor-pointer"
              :class="ticked[i] ? 'border-primary/40 bg-primary/5' : 'border-border'">
              <input v-model="ticked[i]" type="checkbox" class="mt-1 accent-primary" />
              <span>{{ p }}</span>
            </label>
          </div>
          <div class="flex items-center justify-between gap-3 px-5 pb-5">
            <span class="text-xs text-muted-foreground">Tắt bất cứ lúc nào trong Cài đặt</span>
            <span class="flex gap-2">
              <Button variant="outline" @click="setMode('default')">Để sau</Button>
              <Button :disabled="!pledgeOk" @click="setMode('remote')">Bật kết nối từ xa</Button>
            </span>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D21 · Mục riêng "MCP" trên Cài đặt. Trong máy bật sẵn, từ xa tắt sẵn. Bật từ xa phải tick đủ 3 ý. Mã ghép 40 ký tự, chỉ chép-dán, dùng một lần, hết hạn sau 2 phút. Không có quyền xoá qua MCP.
    </p>
  </div>
</template>
