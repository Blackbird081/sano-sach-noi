<script setup lang="ts">
// D20 — Tìm và thay trong lời đọc cả cuốn (Sửa sách · tab Nội dung). Bấm "Tìm và thay" (hoặc
// ⌘F) → đầu cột mục lục có ô Tìm + Thay bằng; mục lục lọc còn các mục có chữ cần tìm, kèm số
// chỗ; ô lời đọc bên phải tô vàng chỗ khớp. "Thay tất cả" ghi chữ mới vào các mục thành bản
// CHƯA LƯU (chấm vàng) — chưa đọc lại. Rồi bấm nút có sẵn "Lưu & đọc lại N mục" ở đầu trang:
// chỉ đọc lại đúng các mục bị thay, không đọc lại cả cuốn. Chỉ thay trong lời đọc, tiêu đề
// mục giữ nguyên. Phân biệt hoa thường, khớp đúng nguyên cụm.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import { ChevronLeft, Mic, RotateCcw, Search, Replace, X, Sun, Moon, Loader2, AlertCircle, Headphones, Play, Check, Library, FilePlus2, BarChart3, Settings, Info } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'find' | 'replaced' | 'reading'
const mode = ref<Mode>('find')
const dark = ref(false)
const states: [Mode, string][] = [['find', '1. Đang tìm'], ['replaced', '2. Đã thay, chưa lưu'], ['reading', '3. Đang đọc lại 6 mục']]

const find = ref('R, O, A')
const repl = ref('R O A')
type Hit = { ch: string; t: string; d: string; n: number }
const hits: Hit[] = [
  { ch: 'Chương 6. R, O, A và R, O, E', t: 'Công ty nhỏ mà chỉ số đẹp nhất', d: '0:33', n: 2 },
  { ch: 'Chương 6. R, O, A và R, O, E', t: 'Hai công thức cơ bản', d: '0:39', n: 3 },
  { ch: 'Chương 6. R, O, A và R, O, E', t: 'Quan hệ giữa R, O, A và R, O, E', d: '0:41', n: 4 },
  { ch: 'Chương 7. Đọc chỉ số như người trong nghề', t: 'Công thức Đu-pông', d: '0:44', n: 2 },
  { ch: 'Chương 7. Đọc chỉ số như người trong nghề', t: 'Hai cảnh báo khi đọc chỉ số', d: '0:42', n: 2 },
  { ch: 'Chương 8. Tổng kết', t: 'Ba ý cần nhớ', d: '0:44', n: 1 },
]
const total = hits.reduce((n, h) => n + h.n, 0)
const picked = ref(3)

const before = `Có một công thức chặt nhỏ chỉ số ra để trả lời. Tên là Đu-pông, viết là D, U, P, O, N, T.

Đu-pông cho R, O, A: bằng biên lợi nhuận thuần, nhân vòng quay tài sản.

Biên lợi nhuận thuần là lợi nhuận sau thuế chia doanh thu. Vòng quay tài sản là doanh thu chia tổng tài sản.

Đu-pông cho R, O, E là bản mở rộng của R, O, A, nhân thêm đòn bẩy tài chính.`
// Tách đoạn để tô chỗ khớp (find) hoặc chỗ đã thay (replaced).
const parts = computed(() => {
  const src = mode.value === 'find' ? before : before.split(find.value).join(repl.value)
  const key = mode.value === 'find' ? find.value : repl.value
  const out: { s: string; hit: boolean }[] = []
  src.split(key).forEach((s, i, a) => {
    out.push({ s, hit: false })
    if (i < a.length - 1) out.push({ s: key, hit: true })
  })
  return out
})
const nav = [
  { label: 'Thư viện', icon: Library, on: true },
  { label: 'Tạo sách nói', icon: FilePlus2 },
  { label: 'Hành trình nghe', icon: BarChart3 },
  { label: 'Cài đặt', icon: Settings },
  { label: 'Giới thiệu', icon: Info },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="s in states" :key="s[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="mode === s[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="mode = s[0]">{{ s[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span><span class="h-3 w-3 rounded-full bg-rag-amber"></span><span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>
      <div class="flex-1 min-h-0 flex">
        <aside class="w-52 shrink-0 border-r border-border bg-muted/30 p-3 text-sm">
          <div class="flex items-center gap-2 px-2 py-1 font-semibold"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /> Sano</div>
          <nav class="mt-4 space-y-0.5">
            <span v-for="n in nav" :key="n.label" class="flex items-center gap-2.5 px-3 h-9 rounded-md" :class="n.on ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'"><component :is="n.icon" class="w-4 h-4" /> {{ n.label }}</span>
          </nav>
        </aside>

        <section class="flex-1 min-w-0 flex flex-col">
          <!-- Đầu trang Sửa sách: nút "Lưu & đọc lại N mục" có sẵn -->
          <div class="px-6 pt-5 border-b border-border">
            <div class="flex items-center gap-3">
              <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Quay lại</span>
              <div class="h-9 w-7 rounded overflow-hidden shrink-0"><WfBookCover title="Tài chính doanh nghiệp cơ bản" class="h-full w-full" /></div>
              <div class="min-w-0">
                <h1 class="text-base font-semibold leading-tight truncate">Sửa sách · 03. Tài chính doanh nghiệp cơ bản</h1>
                <p class="text-xs text-muted-foreground flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng Ngọc Linh · 99 mục · 1 giờ 1 phút</p>
              </div>
              <div v-if="mode !== 'find'" class="ml-auto flex items-center gap-2">
                <span class="text-xs text-muted-foreground">{{ mode === 'reading' ? 'Đang lưu và đọc lại 2/6 mục… còn khoảng 3 phút' : '6 mục đã sửa, chưa lưu' }}</span>
                <Button size="sm" :disabled="mode === 'reading'"><RotateCcw class="w-4 h-4" /> Lưu & đọc lại 6 mục</Button>
              </div>
            </div>
            <div class="mt-4 flex items-end gap-5 text-sm">
              <span class="pb-2.5 -mb-px border-b-2 border-primary font-medium">Nội dung</span>
              <span class="pb-2.5 text-muted-foreground">Thông tin & bìa</span>
              <span class="pb-2.5 text-muted-foreground">Giọng đọc</span>
              <span class="pb-2.5 text-muted-foreground">Từ điển</span>
            </div>
          </div>

          <div class="flex-1 min-h-0 flex">
            <!-- Trái: ô tìm / thay + mục lục đã lọc -->
            <div class="w-[300px] shrink-0 border-r border-border flex flex-col min-h-0">
              <div class="p-3 border-b border-border space-y-2 bg-muted/30">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-semibold flex items-center gap-1.5"><Replace class="w-3.5 h-3.5" /> Tìm và thay trong lời đọc</span>
                  <button class="text-muted-foreground" aria-label="Đóng tìm"><X class="w-3.5 h-3.5" /></button>
                </div>
                <div class="h-8 rounded-md border border-border bg-background px-2 flex items-center gap-1.5 text-sm">
                  <Search class="w-3.5 h-3.5 text-muted-foreground shrink-0" /><input v-model="find" class="flex-1 min-w-0 bg-transparent outline-none" placeholder="Tìm" />
                  <span class="text-[11px] text-muted-foreground shrink-0">{{ mode === 'find' ? `${total} chỗ` : '0 chỗ' }}</span>
                </div>
                <div class="h-8 rounded-md border border-border bg-background px-2 flex items-center gap-1.5 text-sm">
                  <Replace class="w-3.5 h-3.5 text-muted-foreground shrink-0" /><input v-model="repl" class="flex-1 min-w-0 bg-transparent outline-none" placeholder="Thay bằng" />
                </div>
                <Button v-if="mode === 'find'" size="sm" class="w-full">Thay tất cả · {{ total }} chỗ trong {{ hits.length }} mục</Button>
                <div v-else class="rounded-md border border-rag-green/40 bg-rag-green/10 px-2.5 py-1.5 text-[11px] leading-relaxed flex gap-1.5">
                  <Check class="w-3.5 h-3.5 text-rag-green shrink-0 mt-0.5" />
                  <span>Đã thay {{ total }} chỗ trong {{ hits.length }} mục. <template v-if="mode === 'replaced'">Chưa lưu: bấm <b>Lưu & đọc lại 6 mục</b> ở trên, hoặc <button class="underline">hoàn tác</button>.</template><template v-else>Đang đọc lại, các mục khác giữ nguyên.</template></span>
                </div>
              </div>
              <p class="px-3 pt-2.5 pb-1 text-[11px] text-muted-foreground">{{ hits.length }} mục có "{{ mode === 'find' ? find : repl }}" · bấm để xem</p>
              <div class="flex-1 overflow-auto pb-2">
                <template v-for="(h, i) in hits" :key="h.t">
                  <p v-if="i === 0 || hits[i - 1].ch !== h.ch" class="px-3 pt-2 pb-1 text-[11px] font-medium text-muted-foreground truncate">{{ h.ch }}</p>
                  <button class="w-full flex items-center gap-2 px-3 h-8 text-sm text-left" :class="picked === i ? 'bg-primary/10 text-primary' : 'hover:bg-muted/60'" @click="picked = i">
                    <span v-if="mode !== 'find'" class="h-1.5 w-1.5 rounded-full shrink-0" :class="mode === 'reading' && i < 1 ? 'bg-rag-green' : 'bg-rag-amber'"></span>
                    <span class="flex-1 truncate">{{ h.t }}</span>
                    <Loader2 v-if="mode === 'reading' && i === 1" class="w-3.5 h-3.5 animate-spin text-muted-foreground" />
                    <span v-else class="text-[11px] rounded-full bg-muted px-1.5 text-muted-foreground">{{ h.n }}</span>
                  </button>
                </template>
              </div>
            </div>

            <!-- Phải: mục đang xem, tô chỗ khớp -->
            <div class="flex-1 min-w-0 p-5 flex flex-col gap-3 overflow-auto">
              <p class="text-xs text-muted-foreground">Chương 7 · Tiểu mục 5/8</p>
              <div>
                <span class="text-xs font-medium">Tiêu đề tiểu mục</span>
                <div class="mt-1 h-9 rounded-md border border-border px-3 flex items-center text-sm">Công thức Đu-pông</div>
              </div>
              <div class="flex-1 flex flex-col min-h-0">
                <span class="text-xs font-medium">Lời đọc</span>
                <div class="mt-1 flex-1 rounded-md border border-border px-3 py-2.5 text-sm leading-relaxed whitespace-pre-wrap overflow-auto"><template v-for="(p, i) in parts" :key="i"><mark v-if="p.hit" class="rounded px-0.5" :class="mode === 'find' ? 'bg-rag-amber/40 text-foreground' : 'bg-rag-green/25 text-foreground'">{{ p.s }}</mark><template v-else>{{ p.s }}</template></template></div>
              </div>
              <div v-if="mode === 'replaced'" class="rounded-md border border-rag-amber/50 bg-rag-amber/10 px-3 py-2 text-xs flex items-center gap-2"><AlertCircle class="w-3.5 h-3.5 text-rag-amber shrink-0" /> <span><b>Chưa lưu.</b> Chữ đã thay, chưa đọc lại. Khi nghe vẫn là bản cũ.</span></div>
              <div class="flex items-center gap-2">
                <Button variant="outline" size="sm"><Headphones class="w-4 h-4" /> Nghe bản đang có</Button>
                <Button variant="outline" size="sm"><Play class="w-4 h-4" /> Nghe thử đoạn chọn</Button>
                <span class="text-[11px] text-muted-foreground">Nên nghe thử cách viết mới trước khi thay cả cuốn.</span>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>
