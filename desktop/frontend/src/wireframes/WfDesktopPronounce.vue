<script setup lang="ts">
// D12 — Từ điển cách đọc + gọn bước Nghe thử. Bám khung D1 (1100×720).
// Từ điển: mỗi dòng "chữ trong sách → đọc là". Chỉ đổi LỜI ĐỌC, chữ hiện khi nghe giữ nguyên
// ("Nielsen" vẫn hiện Nielsen, máy đọc "Niu-sen"). Hai phạm vi: chỉ cuốn này / mọi sách.
// Dùng ở 3 nơi:
//   1. Tạo sách → Nạp file: bấm từ viết tắt đang bị cảnh báo → thêm cách đọc ngay.
//   2. Tạo sách → Nghe thử: bôi đen một từ → "Đọc từ này là…". Có khung Từ điển của cuốn.
//   3. Sửa sách → tab Từ điển: sửa xong, Sano chỉ ra các mục có từ đó → Lưu & đọc lại một lượt.
//   (+ Cài đặt → Từ điển chung: các từ dùng cho mọi sách.)
// Nghe thử gọn lại: ô sửa to hơn, nghe thử đoạn bôi đen, Huỷ thay đổi, nút "Lưu & đọc lại
// đoạn này" (cùng cách gọi với Sửa sách), nhắc sửa chi tiết cả đoạn ở Sửa sách sau khi tạo.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, Check, ChevronRight, ChevronDown,
  AlertTriangle, FileText, X, Play, Pause, Plus, Trash2, BookA, Headphones, RotateCcw, AlertCircle, Pencil, Search, Globe, BookOpen, Loader2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type Mode = 'file' | 'filePop' | 'preview' | 'select' | 'dict' | 'editDict' | 'settings'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'file')
const dark = ref(false)
const states: [Mode, string][] = [
  ['file', '1. Nạp file: từ chưa có cách đọc'],
  ['filePop', '2. Bấm "KPI" → thêm cách đọc'],
  ['preview', '3. Nghe thử (bản gọn)'],
  ['select', '4. Bôi đen → Đọc từ này là…'],
  ['dict', '5. Từ điển của cuốn'],
  ['editDict', '6. Sửa sách → tab Từ điển'],
  ['settings', '7. Cài đặt → Từ điển chung'],
]
const inCreate = computed(() => ['file', 'filePop', 'preview', 'select', 'dict'].includes(mode.value))
const navKey = computed(() => (inCreate.value ? 'create' : mode.value === 'settings' ? 'settings' : 'library'))
const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách nói', icon: FilePlus2 },
  { key: 'stats', label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const steps = ['Cách đọc', 'Nạp file', 'Mục lục', 'Giọng đọc', 'Lời mở đầu', 'Nghe thử', 'Render']
const curStep = computed(() => (mode.value === 'file' || mode.value === 'filePop' ? 2 : 6))

const acr = [
  { w: 'KPI', n: 14, added: mode.value === 'filePop' },
  { w: 'OKR', n: 6 },
  { w: 'SWOT', n: 3 },
]
const bookDict = [
  { w: 'KPI', r: 'ca pê i', n: 14 },
  { w: 'Nielsen', r: 'Niu-sen', n: 7 },
  { w: 'OKR', r: 'âu ca a', n: 6 },
]
const globalDict = [
  { w: 'CEO', r: 'xi i âu', builtin: true },
  { w: 'KPI', r: 'ca pê i', builtin: true },
  { w: 'SePay', r: 'Xi pay', builtin: false },
  { w: 'Shopee', r: 'Sốp pi', builtin: false },
  { w: 'Nielsen', r: 'Niu-sen', builtin: false },
]
const playing = ref(false)
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center max-w-[1100px]">
      <button v-for="m in states" :key="m[0]" class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="mode = m[0]">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span><span class="h-3 w-3 rounded-full bg-rag-amber"></span><span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>
      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /><span class="font-semibold tracking-tight">Sano</span></div>
          <nav class="px-2 space-y-0.5">
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
              :class="n.key === navKey ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'"><component :is="n.icon" class="w-4 h-4" /> {{ n.label }}</span>
          </nav>
          <div class="px-2 mt-2"><span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span></div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.17 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- ═══ Tạo sách ═══ -->
        <section v-if="inCreate" class="flex-1 flex flex-col min-w-0">
          <div class="px-6 pt-4 pb-3 border-b border-border flex items-center gap-1 text-sm overflow-hidden">
            <template v-for="(s, i) in steps" :key="s">
              <span class="flex items-center gap-1.5 px-1.5 h-8 whitespace-nowrap" :class="i + 1 === curStep ? 'font-medium' : 'text-muted-foreground'">
                <span class="h-5 w-5 rounded-full grid place-items-center text-[11px]" :class="i + 1 < curStep ? 'bg-primary/15 text-primary' : i + 1 === curStep ? 'bg-primary text-primary-foreground' : 'bg-muted'">
                  <Check v-if="i + 1 < curStep" class="w-3 h-3" /><template v-else>{{ i + 1 }}</template></span>{{ s }}
              </span>
              <ChevronRight v-if="i < steps.length - 1" class="w-3.5 h-3.5 text-muted-foreground/50 shrink-0" />
            </template>
          </div>

          <!-- 1–2. Nạp file -->
          <div v-if="mode === 'file' || mode === 'filePop'" class="flex-1 overflow-auto p-6">
            <div class="max-w-2xl">
              <h1 class="text-xl font-semibold tracking-tight">Nạp file Word</h1>
              <div class="mt-4 rounded-lg border border-border p-4 flex items-center gap-3">
                <FileText class="w-8 h-8 text-primary shrink-0" />
                <div class="flex-1"><p class="font-medium">Quan-tri-muc-tieu.docx</p><p class="text-xs text-muted-foreground">412 KB · 6 chương · 38 tiểu mục · 61.240 ký tự</p></div>
                <span class="text-sm text-muted-foreground flex items-center gap-1"><X class="w-4 h-4" /> Chọn file khác</span>
              </div>
              <div class="mt-4 rounded-lg border border-rag-amber/40 bg-rag-amber/10 p-4 text-sm">
                <p class="font-medium flex items-center gap-2 text-rag-amber"><AlertTriangle class="w-4 h-4" /> Có phần sẽ không được đọc trọn vẹn</p>
                <ul class="mt-2 space-y-1 text-foreground/80 list-disc pl-5">
                  <li>2 bảng: nội dung bảng được đọc phẳng từng ô, mất hàng/cột</li>
                </ul>
                <!-- MỚI: từ chưa có cách đọc thành từng nút bấm được -->
                <div class="mt-3 border-t border-rag-amber/30 pt-3">
                  <p class="text-foreground/80"><b>3 từ viết tắt chưa có cách đọc</b>, bộ đọc có thể đọc sai. Bấm từng từ để dạy Sano cách đọc:</p>
                  <div class="mt-2 flex flex-wrap gap-2 relative">
                    <span v-for="a in acr" :key="a.w" class="h-8 px-3 rounded-full border text-sm flex items-center gap-1.5"
                      :class="a.w === 'KPI' && mode === 'filePop' ? 'border-primary bg-background ring-2 ring-primary/30' : 'border-border bg-background'">
                      <b>{{ a.w }}</b><span class="text-xs text-muted-foreground">{{ a.n }} lần</span>
                    </span>
                    <span class="h-8 px-3 text-sm text-muted-foreground flex items-center">Bỏ qua cũng được, sửa sau ở bước Nghe thử</span>
                    <!-- Hộp thêm cách đọc -->
                    <div v-if="mode === 'filePop'" class="absolute top-10 left-0 w-80 rounded-lg border border-border bg-popover shadow-xl p-3 z-10">
                      <div class="text-sm font-medium">Dạy Sano đọc "KPI"</div>
                      <p class="mt-0.5 text-xs text-muted-foreground">Gõ đúng như tai nghe, cách nhau bằng dấu cách.</p>
                      <label class="mt-2 block text-xs text-muted-foreground">Đọc là
                        <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground" value="ca pê i" /></label>
                      <div class="mt-2 flex items-center gap-3 text-xs">
                        <label class="flex items-center gap-1.5"><input type="radio" checked class="accent-[hsl(var(--primary))]" /> Chỉ cuốn này</label>
                        <label class="flex items-center gap-1.5"><input type="radio" class="accent-[hsl(var(--primary))]" /> Mọi sách</label>
                      </div>
                      <div class="mt-3 flex items-center gap-2">
                        <Button variant="outline" size="sm"><Play class="w-3.5 h-3.5" /> Nghe thử</Button>
                        <span class="flex-1"></span>
                        <Button variant="ghost" size="sm">Huỷ</Button>
                        <Button size="sm">Thêm</Button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 3–5. Nghe thử (bản gọn) -->
          <div v-else class="flex-1 flex min-h-0">
            <div class="flex-1 overflow-auto p-6 min-w-0">
              <h1 class="text-xl font-semibold tracking-tight">Nghe thử trước khi render</h1>
              <p class="text-sm text-muted-foreground">Nghe vài đoạn để chốt giọng và cách đọc. Mỗi đoạn chỉ đọc khoảng 500 ký tự đầu.</p>
              <p class="mt-2 text-xs text-muted-foreground flex items-center gap-1.5"><BookOpen class="w-3.5 h-3.5" /> Sửa chi tiết cả đoạn sau khi tạo xong: Thư viện → ⋯ → <b>Sửa sách</b>.</p>

              <div class="mt-4 rounded-lg border border-border p-4">
                <div class="flex items-center gap-3">
                  <span class="h-9 w-9 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0"><Play class="w-4 h-4" /></span>
                  <span class="flex-1 font-medium text-sm">Giới thiệu sách</span>
                  <Check class="w-4 h-4 text-rag-green" /><span class="text-xs rounded-md bg-muted px-2 py-0.5">Đã render · 0:06</span>
                </div>
              </div>

              <div class="mt-3 rounded-lg border p-4" :class="mode === 'select' ? 'border-rag-amber' : 'border-border'">
                <div class="flex items-center gap-3">
                  <span class="h-9 w-9 grid place-items-center rounded-full bg-primary text-primary-foreground shrink-0" @click="playing = !playing"><component :is="playing ? Pause : Play" class="w-4 h-4" /></span>
                  <span class="flex-1 font-medium text-sm">Chương 1. Vì sao phải đo lường</span>
                  <span class="text-xs rounded-md bg-muted px-2 py-0.5">Đã render đoạn đầu · 0:38</span>
                </div>
                <p class="mt-3 text-xs font-medium text-muted-foreground flex items-center gap-1.5"><Pencil class="w-3 h-3" /> Lời đọc · sửa chữ, hoặc bôi đen một từ để dạy cách đọc</p>
                <div class="relative mt-1">
                  <div class="w-full h-40 overflow-auto rounded-md border bg-background p-3 text-sm leading-relaxed" :class="mode === 'select' ? 'border-rag-amber' : 'border-input'">
                    Vì sao phải đo lường.<br /><br />
                    Theo khảo sát của <span :class="mode === 'select' ? 'bg-primary/25 rounded px-0.5' : ''">Nielsen</span> năm 2024, doanh nghiệp đặt mục tiêu rõ ràng tăng trưởng nhanh gấp đôi. Công cụ phổ biến nhất là ca pê i, tức chỉ số đo hiệu quả công việc.<br /><br />
                    Nhưng đo cái gì mới là câu hỏi khó. Đo sai thì cả đội chạy sai hướng…
                  </div>
                  <!-- Thanh nổi khi bôi đen -->
                  <div v-if="mode === 'select'" class="absolute left-48 top-6 rounded-lg border border-border bg-popover shadow-xl p-1 flex items-center gap-1 text-sm z-10">
                    <span class="h-8 px-2.5 rounded-md bg-primary/10 text-primary font-medium flex items-center gap-1.5"><BookA class="w-4 h-4" /> Đọc từ này là…</span>
                    <span class="h-8 px-2.5 rounded-md flex items-center gap-1.5"><Play class="w-3.5 h-3.5" /> Nghe thử</span>
                  </div>
                  <div v-if="mode === 'select'" class="absolute left-48 top-[4.5rem] w-80 rounded-lg border border-border bg-popover shadow-xl p-3 z-10">
                    <div class="text-sm font-medium">Dạy Sano đọc "Nielsen"</div>
                    <p class="mt-0.5 text-xs text-muted-foreground">Có 7 chỗ trong sách. Chữ hiện khi nghe vẫn là Nielsen.</p>
                    <input class="mt-2 w-full h-9 rounded-md border border-input bg-background px-3 text-sm" value="Niu-sen" />
                    <div class="mt-2 flex items-center gap-3 text-xs">
                      <label class="flex items-center gap-1.5"><input type="radio" checked class="accent-[hsl(var(--primary))]" /> Chỉ cuốn này</label>
                      <label class="flex items-center gap-1.5"><input type="radio" class="accent-[hsl(var(--primary))]" /> Mọi sách</label>
                    </div>
                    <div class="mt-3 flex items-center gap-2">
                      <Button variant="outline" size="sm"><Play class="w-3.5 h-3.5" /> Nghe thử</Button><span class="flex-1"></span>
                      <Button variant="ghost" size="sm">Huỷ</Button><Button size="sm">Thêm và đọc lại</Button>
                    </div>
                  </div>
                </div>
                <div class="mt-2 flex flex-wrap items-center gap-2">
                  <Button variant="outline" size="sm"><Play class="w-3.5 h-3.5" /> Nghe thử đoạn chọn</Button>
                  <span class="text-[11px] text-muted-foreground">Bôi đen một câu rồi bấm để nghe thử.</span>
                  <div class="ml-auto flex gap-2">
                    <Button variant="ghost" size="sm" :disabled="mode !== 'select'">Huỷ thay đổi</Button>
                    <Button size="sm" :disabled="mode !== 'select'"><RotateCcw class="w-4 h-4" /> Lưu & đọc lại đoạn này</Button>
                  </div>
                </div>
              </div>

              <div class="mt-3 flex items-center gap-2">
                <span class="h-9 rounded-md border border-input bg-background px-3 text-sm flex items-center gap-2 text-muted-foreground w-96"><Plus class="w-4 h-4" /> Chọn thêm đoạn khác để nghe thử <ChevronDown class="w-4 h-4 ml-auto" /></span>
              </div>
            </div>

            <!-- Khung Từ điển của cuốn (bên phải bước Nghe thử) -->
            <aside class="w-80 shrink-0 border-l border-border flex flex-col" :class="mode === 'dict' ? 'bg-background' : 'bg-muted/20'">
              <div class="px-4 py-3 border-b border-border">
                <div class="text-sm font-semibold flex items-center gap-2"><BookA class="w-4 h-4" /> Từ điển của cuốn này · 3</div>
                <p class="text-[11px] text-muted-foreground mt-0.5">Chỉ đổi cách đọc, chữ hiện khi nghe giữ nguyên. Áp cho cả cuốn khi render.</p>
              </div>
              <div class="flex-1 overflow-auto divide-y divide-border">
                <div v-for="d in bookDict" :key="d.w" class="px-4 py-2.5 flex items-center gap-2 text-sm">
                  <div class="min-w-0 flex-1">
                    <div><b>{{ d.w }}</b> <span class="text-muted-foreground">→</span> {{ d.r }}</div>
                    <div class="text-[11px] text-muted-foreground">{{ d.n }} chỗ trong sách</div>
                  </div>
                  <Play class="w-3.5 h-3.5 text-muted-foreground" /><Pencil class="w-3.5 h-3.5 text-muted-foreground" /><Trash2 class="w-3.5 h-3.5 text-muted-foreground" />
                </div>
                <div v-if="mode === 'dict'" class="px-4 py-3 space-y-2 bg-muted/40">
                  <div class="text-xs font-medium">Thêm từ</div>
                  <div class="grid grid-cols-[1fr_auto_1fr] gap-1.5 items-center">
                    <input class="h-8 rounded-md border border-input bg-background px-2 text-sm" placeholder="Chữ trong sách" value="SWOT" />
                    <span class="text-muted-foreground text-sm">→</span>
                    <input class="h-8 rounded-md border border-input bg-background px-2 text-sm" placeholder="Đọc là" value="suốt" />
                  </div>
                  <div class="flex items-center gap-2 text-xs"><label class="flex items-center gap-1.5"><input type="checkbox" class="accent-[hsl(var(--primary))]" /> Dùng cho mọi sách</label><span class="flex-1"></span><Button size="sm">Thêm</Button></div>
                </div>
                <div v-else class="px-4 py-2.5 text-sm text-primary flex items-center gap-1.5"><Plus class="w-4 h-4" /> Thêm từ</div>
              </div>
              <div class="px-4 py-3 border-t border-border text-[11px] text-muted-foreground leading-relaxed flex gap-1.5">
                <Globe class="w-3.5 h-3.5 shrink-0 mt-0.5" /> Từ dùng cho mọi sách nằm ở Cài đặt → Từ điển chung (đang có 3 từ tự thêm).
              </div>
            </aside>
          </div>
        </section>

        <!-- ═══ 6. Sửa sách → tab Từ điển ═══ -->
        <section v-else-if="mode === 'editDict'" class="flex-1 flex flex-col min-w-0">
          <div class="px-6 pt-5 border-b border-border">
            <div class="flex items-center gap-3">
              <span class="text-sm text-muted-foreground">‹ Quay lại</span>
              <div class="min-w-0"><h1 class="text-base font-semibold">Sửa sách · Quản trị mục tiêu</h1><p class="text-xs text-muted-foreground">Giọng Hải Đăng · 38 mục · 1 giờ 5 phút</p></div>
              <div class="ml-auto flex items-center gap-2"><span class="text-xs text-muted-foreground">7 mục có từ vừa đổi, chưa lưu</span><Button size="sm"><RotateCcw class="w-4 h-4" /> Lưu & đọc lại 7 mục</Button></div>
            </div>
            <div class="mt-4 flex gap-5 text-sm">
              <span class="pb-2.5 text-muted-foreground">Nội dung</span><span class="pb-2.5 text-muted-foreground">Thông tin & bìa</span><span class="pb-2.5 text-muted-foreground">Giọng đọc</span>
              <span class="pb-2.5 -mb-px border-b-2 border-primary font-medium">Từ điển</span>
            </div>
          </div>
          <div class="flex-1 flex min-h-0">
            <div class="w-[420px] shrink-0 border-r border-border flex flex-col">
              <div class="p-4 flex gap-2">
                <span class="flex-1 h-9 rounded-md border border-input px-3 text-sm flex items-center gap-2 text-muted-foreground"><Search class="w-4 h-4" /> Tìm trong từ điển</span>
                <Button variant="outline" size="sm"><Plus class="w-4 h-4" /> Thêm từ</Button>
              </div>
              <div class="flex-1 overflow-auto divide-y divide-border border-t border-border">
                <div v-for="(d, i) in bookDict" :key="d.w" class="px-4 py-2.5 flex items-center gap-2 text-sm" :class="i === 1 ? 'bg-primary/5' : ''">
                  <div class="flex-1 min-w-0">
                    <div><b>{{ d.w }}</b> → <template v-if="i === 1"><input class="h-7 w-28 rounded border border-rag-amber bg-background px-2 text-sm" value="Ni-xần" /></template><template v-else>{{ d.r }}</template></div>
                    <div class="text-[11px] text-muted-foreground">{{ d.n }} chỗ · {{ i === 1 ? 'đã đổi từ "Niu-sen", chưa lưu' : 'đã đọc theo cách này' }}</div>
                  </div>
                  <span v-if="i === 1" class="h-2 w-2 rounded-full bg-rag-amber"></span>
                  <Play class="w-3.5 h-3.5 text-muted-foreground" /><Trash2 class="w-3.5 h-3.5 text-muted-foreground" />
                </div>
                <div class="px-4 py-3 text-[11px] text-muted-foreground">Từ dùng cho mọi sách (Cài đặt → Từ điển chung) cũng áp cho cuốn này, không liệt kê ở đây.</div>
              </div>
            </div>
            <div class="flex-1 p-5 min-w-0">
              <div class="text-sm font-medium">"Nielsen" có ở 7 mục</div>
              <p class="text-xs text-muted-foreground">Đổi cách đọc thì các mục này phải đọc lại. Bấm Lưu & đọc lại để làm một lượt.</p>
              <div class="mt-3 rounded-lg border border-border divide-y divide-border text-sm">
                <div v-for="m in ['Chương 1 · Vì sao phải đo lường', 'Chương 2 · Chọn chỉ số dẫn đường', 'Chương 2 · Đo cái gì trước', 'Chương 4 · Bài học từ khảo sát', 'Chương 5 · Đọc số liệu đúng cách']" :key="m" class="px-3 py-2 flex items-center gap-2">
                  <span class="h-2 w-2 rounded-full bg-rag-amber shrink-0"></span><span class="flex-1 truncate">{{ m }}</span><span class="text-xs text-muted-foreground">1–2 chỗ</span><Headphones class="w-3.5 h-3.5 text-muted-foreground" />
                </div>
                <div class="px-3 py-2 text-xs text-muted-foreground">+ 2 mục khác</div>
              </div>
              <div class="mt-4 rounded-md border border-rag-amber/50 bg-rag-amber/10 px-3 py-2 text-xs flex items-center gap-2"><AlertCircle class="w-3.5 h-3.5 text-rag-amber shrink-0" /> <span><b>Chưa lưu.</b> Bấm <b>Lưu & đọc lại 7 mục</b> (khoảng 2 phút). Chưa lưu thì khi nghe vẫn đọc "Niu-sen".</span></div>
              <p class="mt-3 text-[11px] text-muted-foreground flex items-center gap-1.5"><Loader2 class="w-3 h-3" /> Đang đọc lại thì mục đó hiện vòng xoay ở tab Nội dung, xong có dấu ✓ như khi sửa chữ.</p>
            </div>
          </div>
        </section>

        <!-- ═══ 7. Cài đặt → Từ điển chung ═══ -->
        <section v-else class="flex-1 overflow-auto p-6 min-w-0">
          <h1 class="text-xl font-semibold tracking-tight">Cài đặt</h1>
          <div class="mt-5 max-w-2xl space-y-6">
            <div class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Lưu trữ · Nghe và xuất M4B · Dữ liệu nghe · …</div>
            <div>
              <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Từ điển chung · dùng cho mọi sách</h2>
              <div class="rounded-lg border border-border">
                <div class="p-3 flex gap-2 border-b border-border">
                  <span class="flex-1 h-9 rounded-md border border-input px-3 text-sm flex items-center gap-2 text-muted-foreground"><Search class="w-4 h-4" /> Tìm từ</span>
                  <Button variant="outline" size="sm"><Plus class="w-4 h-4" /> Thêm từ</Button>
                </div>
                <div class="divide-y divide-border text-sm">
                  <div v-for="d in globalDict" :key="d.w" class="px-3 py-2 flex items-center gap-2">
                    <b class="w-28">{{ d.w }}</b><span class="text-muted-foreground">→</span><span class="flex-1">{{ d.r }}</span>
                    <span v-if="d.builtin" class="text-[11px] rounded bg-muted px-1.5 py-0.5 text-muted-foreground">có sẵn</span>
                    <Play class="w-3.5 h-3.5 text-muted-foreground" /><Pencil class="w-3.5 h-3.5 text-muted-foreground" /><Trash2 v-if="!d.builtin" class="w-3.5 h-3.5 text-muted-foreground" />
                  </div>
                  <div class="px-3 py-2 text-xs text-muted-foreground">+ 180 từ có sẵn (CEO, CFO, HR, B2B…). Sửa từ có sẵn = ghi đè cách đọc của anh; xoá chỉ xoá được từ tự thêm.</div>
                </div>
              </div>
              <p class="mt-2 text-xs text-muted-foreground">Đổi từ điển chung chỉ áp cho sách tạo sau. Sách đã tạo muốn đọc theo cách mới: Sửa sách → tab Từ điển → Lưu & đọc lại.</p>
            </div>
          </div>
        </section>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center leading-relaxed">
      D12 · Từ điển cách đọc: "chữ trong sách → đọc là", chỉ đổi lời đọc, chữ hiện khi nghe giữ nguyên. Thêm ở Nạp file (từ bị cảnh báo), Nghe thử (bôi đen), Sửa sách (tab Từ điển), Cài đặt (từ điển chung).
      Nghe thử gọn lại: ô sửa to, nghe thử đoạn bôi đen, Huỷ thay đổi, "Lưu & đọc lại đoạn này"; sửa chi tiết cả đoạn ở Sửa sách.
    </p>
  </div>
</template>
