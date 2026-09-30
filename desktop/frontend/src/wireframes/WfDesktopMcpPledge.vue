<script setup lang="ts">
// D22 — Sách AI tạo qua MCP: TẠO TRƯỚC, CAM KẾT SAU, cam kết một lần cho nhiều cuốn.
// - AI xin tạo là Sano render ngay (hàng đợi, lần lượt từng cuốn), không chờ người dùng.
// - Tạo xong: sách nằm ở KHU CHỜ (chưa vào Thư viện, chưa nghe / xuất / chia sẻ được).
// - Popup "cam kết để lưu" tự mở khi có cuốn đầu tiên tạo xong; gom mọi cuốn của AI (đã xong, đang tạo,
//   chờ tạo). Tick cuốn nào thì cuốn đó được lưu; cuốn đang tạo / chờ tạo mà đã tick thì tự lưu khi xong.
// - 6 ô cam kết D19 tick MỘT lần, áp cho các cuốn đã tick; mỗi cuốn ghi thời điểm cam kết riêng.
// - "Để sau" / bấm ra ngoài: chỉ đóng popup, sách vẫn ở khu chờ (thẻ thanh bên mở lại). Không tick sau 7 ngày
//   thì khu chờ tự xoá. "Bỏ" ở từng dòng: bỏ hẳn cuốn đó (hỏi lại một lần).
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, Check, X, Loader2, Clock, BookOpen, Bot, Cable,
  CheckCircle2, Hourglass,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { PLEDGES } from '@/lib/pledge'

type Mode = 'rendering' | 'popup' | 'popupDone' | 'saved'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'popup')
const dark = ref(false)

type St = 'done' | 'reading' | 'queued'
interface Book { title: string; sections: number; min: number; st: St; pct?: number; pick: boolean }
const books = ref<Book[]>([])
function reset(m: Mode) {
  mode.value = m
  const allDone = m === 'popupDone' || m === 'saved'
  books.value = [
    { title: 'Ba thói quen buổi sáng', sections: 12, min: 18, st: 'done', pick: true },
    { title: 'Quản lý dòng tiền cho chủ shop', sections: 20, min: 35, st: allDone ? 'done' : 'reading', pct: 45, pick: true },
    { title: 'Kể chuyện thương hiệu', sections: 15, min: 22, st: allDone ? 'done' : 'queued', pick: m !== 'popupDone' },
  ]
  checked.value = PLEDGES.map(() => m === 'popupDone')
  readTerms.value = m === 'popupDone'
}
const checked = ref<boolean[]>([])
const readTerms = ref(false)
reset(mode.value)

const picked = computed(() => books.value.filter((b) => b.pick))
const pickedDone = computed(() => picked.value.filter((b) => b.st === 'done').length)
const pledgeOk = computed(() => checked.value.every(Boolean) && readTerms.value)
const count = computed(() => checked.value.filter(Boolean).length + (readTerms.value ? 1 : 0))
const popup = computed(() => mode.value === 'popup' || mode.value === 'popupDone')
const action = computed(() => {
  const n = picked.value.length
  if (!n) return 'Chọn ít nhất một cuốn'
  const later = n - pickedDone.value
  return later ? `Cam kết · lưu ${pickedDone.value} cuốn, ${later} cuốn lưu khi tạo xong` : `Cam kết và lưu ${n} cuốn`
})

const nav = [
  { label: 'Thư viện', icon: Library, on: true },
  { label: 'Tạo sách nói', icon: FilePlus2 },
  { label: 'Hành trình nghe', icon: BarChart3 },
  { label: 'MCP', icon: Cable },
  { label: 'Cài đặt', icon: Settings },
  { label: 'Giới thiệu', icon: Info },
]
const shelf = ['Thấy ra chính mình', 'Chánh niệm', 'Tiệm Cà Phê Thứ Hai', 'Nghe Để Nhớ', 'Giới thiệu Sano']
const colors = ['bg-red-800', 'bg-orange-700', 'bg-teal-800', 'bg-violet-800', 'bg-indigo-800']
const states: [Mode, string][] = [
  ['rendering', '1. AI xin tạo 3 cuốn (đang tạo)'],
  ['popup', '2. Cuốn đầu xong → popup tự mở'],
  ['popupDone', '3. Tạo xong cả 3, bỏ tick 1 cuốn'],
  ['saved', '4. Đã lưu, 1 cuốn còn ở khu chờ'],
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in states" :key="m[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="reset(m[0])">{{ m[1] }}</button>
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
            <span v-for="n in nav" :key="n.label" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.on ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>

          <!-- MỚI: thẻ sách AI tạo (thay thẻ "AI đang tạo sách" cũ) -->
          <div v-if="mode !== 'saved'" class="mx-3 mb-3 rounded-lg border border-border bg-background p-2.5 text-xs">
            <div class="flex items-center gap-1.5 font-medium">
              <Loader2 v-if="mode === 'rendering' || mode === 'popup'" class="w-3.5 h-3.5 animate-spin text-primary shrink-0" />
              <Bot v-else class="w-3.5 h-3.5 text-primary shrink-0" />
              <span class="truncate">{{ mode === 'popupDone' ? 'Claude Desktop tạo xong 3 cuốn' : 'Claude Desktop nhờ tạo 3 cuốn' }}</span>
            </div>
            <template v-if="mode !== 'popupDone'">
              <p class="mt-1 text-muted-foreground truncate">Đang tạo {{ mode === 'rendering' ? 1 : 2 }}/3 · {{ mode === 'rendering' ? 'Ba thói quen buổi sáng' : 'Quản lý dòng tiền…' }}</p>
              <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: (mode === 'rendering' ? 62 : 45) + '%' }"></div></div>
            </template>
            <button v-if="mode !== 'rendering'" class="mt-2 w-full h-7 rounded-md bg-primary/10 text-primary font-medium hover:bg-primary/15">
              {{ mode === 'popup' ? '1 cuốn' : '3 cuốn' }} chờ cam kết · Mở</button>
          </div>
          <div v-else class="mx-3 mb-3 rounded-lg border border-border bg-background p-2.5 text-xs">
            <div class="flex items-center gap-1.5 font-medium"><CheckCircle2 class="w-3.5 h-3.5 text-emerald-600 shrink-0" /> Đã lưu 2 cuốn vào Thư viện <X class="w-3.5 h-3.5 ml-auto text-muted-foreground" /></div>
            <p class="mt-1 text-muted-foreground">1 cuốn vẫn ở khu chờ, tự xoá sau 7 ngày nếu không cam kết.</p>
            <button class="mt-2 w-full h-7 rounded-md bg-muted text-foreground font-medium">1 cuốn chờ cam kết · Mở</button>
          </div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.21 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- Nền: Thư viện giản lược -->
        <main class="flex-1 min-w-0 p-6">
          <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
          <p class="text-sm text-muted-foreground">{{ mode === 'saved' ? 29 : 27 }} cuốn</p>
          <div class="mt-6 grid grid-cols-5 gap-4">
            <template v-if="mode === 'saved'">
              <div v-for="t in ['Ba thói quen buổi sáng', 'Quản lý dòng tiền cho chủ shop']" :key="t" class="aspect-[3/4] rounded-md bg-amber-700 p-3 flex items-end text-white font-serif text-sm ring-2 ring-primary/50">{{ t }}</div>
            </template>
            <div v-for="(t, i) in shelf.slice(0, mode === 'saved' ? 3 : 5)" :key="t" class="aspect-[3/4] rounded-md p-3 flex items-end text-white font-serif text-sm" :class="colors[i]">{{ t }}</div>
          </div>
        </main>
      </div>

      <!-- ═══ Popup cam kết để lưu (nhiều cuốn) ═══ -->
      <div v-if="popup" class="absolute inset-0 top-9 z-20 flex items-center justify-center bg-background/70 backdrop-blur-sm p-4">
        <div role="dialog" aria-modal="true" class="flex max-h-full min-h-0 w-[680px] flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl">
          <div class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
            <div>
              <h2 class="font-semibold">Claude Desktop đã tạo sách giúp bạn · cam kết để lưu vào Thư viện</h2>
              <p class="mt-0.5 text-xs text-muted-foreground">Sách đang ở khu chờ, chưa nghe, xuất hay chia sẻ được. Tick cuốn muốn lưu, rồi tick các cam kết.</p>
            </div>
            <button class="text-muted-foreground hover:text-foreground" aria-label="Để sau"><X class="h-4 w-4" /></button>
          </div>

          <div class="flex-1 min-h-0 overflow-auto px-5 py-3">
            <!-- Danh sách cuốn -->
            <div class="rounded-lg border border-border divide-y divide-border">
              <label v-for="b in books" :key="b.title" class="flex items-center gap-3 px-3 py-2 cursor-pointer" :class="b.pick ? '' : 'opacity-60'">
                <input v-model="b.pick" type="checkbox" class="h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
                <span class="flex-1 min-w-0">
                  <span class="block text-sm font-medium truncate">{{ b.title }}</span>
                  <span class="block text-xs text-muted-foreground">{{ b.sections }} mục · {{ b.min }} phút · giọng Hải Đăng</span>
                </span>
                <span class="shrink-0 text-xs flex items-center gap-1.5">
                  <template v-if="b.st === 'done'"><CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" /> <span class="text-emerald-700 dark:text-emerald-400">Đã tạo xong</span></template>
                  <template v-else-if="b.st === 'reading'"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" /> Đang tạo {{ b.pct }}%</template>
                  <template v-else><Hourglass class="w-3.5 h-3.5 text-muted-foreground" /> <span class="text-muted-foreground">Chờ tạo</span></template>
                </span>
                <button class="shrink-0 text-xs text-muted-foreground hover:text-destructive" @click.prevent>Bỏ</button>
              </label>
            </div>
            <p v-if="picked.length - pickedDone > 0" class="mt-1.5 text-xs text-muted-foreground flex items-center gap-1"><Clock class="w-3 h-3" /> Cuốn đang tạo / chờ tạo đã tick sẽ tự lưu khi tạo xong.</p>
            <p v-if="books.some((b) => !b.pick)" class="mt-1.5 text-xs text-muted-foreground">Cuốn không tick vẫn ở khu chờ 7 ngày, muốn bỏ hẳn bấm "Bỏ".</p>

            <!-- Cam kết (D19), gọn hơn: mô tả nhỏ -->
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
                <span class="text-sm">Tôi đã đọc và đồng ý <a href="#" class="font-medium text-primary underline underline-offset-2" @click.prevent>Điều khoản sử dụng</a> của Sano.</span>
              </label>
            </div>
          </div>

          <div class="border-t border-border px-5 py-3">
            <p class="text-[11px] leading-relaxed text-muted-foreground">Sano ghi thời điểm cam kết vào thông tin từng cuốn. Vi phạm các cam kết trên, bạn tự chịu trách nhiệm trước pháp luật.</p>
            <div class="mt-3 flex items-center justify-between gap-3">
              <span class="text-xs" :class="pledgeOk ? 'text-rag-green flex items-center gap-1' : 'text-muted-foreground'"><Check v-if="pledgeOk" class="h-3.5 w-3.5" /> Đã tick {{ count }}/{{ PLEDGES.length + 1 }}</span>
              <div class="flex gap-2">
                <Button variant="outline">Để sau</Button>
                <Button :disabled="!pledgeOk || !picked.length">{{ action }}</Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D22 · Sách AI tạo: tạo trước, cam kết sau. Một popup cho mọi cuốn của AI, tick cuốn muốn lưu + 6 ô cam kết một lần. "Để sau" chỉ đóng popup, sách vẫn ở khu chờ 7 ngày.
      Luồng Tạo sách nói trong app giữ nguyên (cam kết trước khi render).
    </p>
  </div>
</template>
