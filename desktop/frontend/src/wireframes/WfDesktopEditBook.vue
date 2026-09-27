<script setup lang="ts">
// D11 — Sửa sách: sửa sách đã tạo mà không phải xoá rồi tạo lại. Bám khung D1 (1100×720).
// Một màn riêng, 3 tab:
// - Nội dung: mục lục bên trái, ô sửa lời đọc của tiểu mục bên phải. "Đọc lại mục này"
//   chỉ đọc lại đúng tiểu mục đó (vài chục giây), các mục khác giữ nguyên.
// - Thông tin & bìa: tên, tác giả, danh mục, bộ sách. Đổi tên → bìa tự vẽ vẽ lại theo tên
//   mới, lời giới thiệu "Cuốn sách: …" tự đọc lại. Bìa là ảnh anh chọn thì giữ nguyên.
// - Giọng đọc: đổi giọng cả cuốn. Đọc lại chạy nền, bản cũ vẫn nghe được đến khi xong.
// Mở từ: menu ⋯ trên bìa ở Thư viện ("Sửa sách") và bánh răng ở màn nghe ("Sửa mục đang nghe").
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, ChevronLeft, ChevronDown, ChevronRight,
  Check, Mic, Play, Pause, RotateCcw, Loader2, Pencil, ImagePlus, Wand2, AlertCircle, Headphones, Square, Package, FolderOpen,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'entry' | 'content' | 'dirty' | 'rendering' | 'info' | 'voice' | 'voiceRunning'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'content')
const dark = ref(false)
const tab = computed(() => (mode.value === 'info' ? 'info' : mode.value === 'voice' || mode.value === 'voiceRunning' ? 'voice' : 'content'))
const states: [Mode, string][] = [
  ['entry', '0. Lối vào'],
  ['content', '1. Nội dung'],
  ['dirty', '2. Đã sửa, chờ đọc lại'],
  ['rendering', '3. Đang đọc lại'],
  ['info', '4. Thông tin & bìa'],
  ['voice', '5. Đổi giọng'],
  ['voiceRunning', '6. Đang đổi giọng'],
]

const bookTitle = computed(() => (mode.value === 'info' ? 'Brand Marketing căn bản' : 'Brand Marketing'))
const newVoice = ref('Mỹ Duyên')

type Sec = { t: string; d: string; state?: 'edited' | 'rendering' | 'done' | 'queued' }
const chapters = computed<{ t: string; open: boolean; secs: Sec[] }[]>(() => {
  const s2: Sec['state'] = mode.value === 'dirty' ? 'edited' : mode.value === 'rendering' ? 'rendering' : undefined
  const s3: Sec['state'] = mode.value === 'dirty' ? 'edited' : mode.value === 'rendering' ? 'queued' : undefined
  return [
    { t: 'Giới thiệu', open: false, secs: [{ t: 'Brand Marketing', d: '0:06' }] },
    { t: 'Lời mở đầu', open: false, secs: [{ t: 'Lời mở đầu', d: '0:54' }] },
    {
      t: 'Chương 1. Vì sao phải làm brand marketing', open: true, secs: [
        { t: 'Chai Sunlight và chai nước rửa chén', d: '0:35', state: mode.value === 'rendering' ? 'done' : undefined },
        { t: 'Bốn áp lực buộc phải có thương hiệu', d: '0:48', state: s2 },
        { t: 'Ba thời kỳ của marketing', d: '0:30', state: s3 },
        { t: 'Thương hiệu là gì', d: '1:12' },
      ],
    },
    { t: 'Chương 2. Định vị trong tâm trí', open: false, secs: [] },
    { t: 'Chương 3. Thiết kế thương hiệu', open: false, secs: [] },
    { t: 'Chương 4. Tài sản thương hiệu', open: false, secs: [] },
  ]
})
const pending = computed(() => (mode.value === 'dirty' ? 2 : mode.value === 'rendering' ? 2 : 0))

const original = `Bốn áp lực buộc phải có thương hiệu.

Áp lực thứ nhất: hàng hoá ngày càng giống nhau. Ra chợ, mười chai nước rửa chén thì chín chai rửa sạch như nhau.

Áp lực thứ hai là cạnh tranh về giá. Không có thương hiệu thì chỉ còn cách giảm giá, mà giảm mãi thì lỗ.

Áp lực thứ ba, khách hàng có quá nhiều lựa chọn. Theo khảo sát của Niu-sen năm 2024, trung bình một người thấy hơn năm nghìn quảng cáo mỗi ngày.`
const edited = original.replace('Niu-sen', 'Nielsen')
const text = computed(() => (mode.value === 'dirty' || mode.value === 'rendering' ? edited : original))

const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách nói', icon: FilePlus2 },
  { key: 'stats', label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const voices = [
  { n: 'Hải Đăng', s: 'nam · Bắc', rec: true },
  { n: 'Thiện Minh', s: 'nam · Bắc', rec: true },
  { n: 'Mỹ Duyên', s: 'nữ · Nam', rec: true },
  { n: 'Trúc Ly', s: 'nữ · Bắc' },
  { n: 'Thái Sơn', s: 'nam · Nam' },
  { n: 'Ngọc Trân', s: 'nữ · Trung' },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <!-- Nút duyệt trạng thái (ngoài khung) -->
    <div class="flex flex-wrap gap-2 justify-center max-w-[1100px]">
      <button v-for="m in states" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="mode = m[0]">{{ m[1] }}</button>
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
              :class="n.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <!-- Đổi giọng chạy nền: thẻ tiến độ ở thanh bên, đi đâu cũng thấy -->
          <div v-if="mode === 'voiceRunning'" class="mx-3 mb-3 rounded-lg border border-border bg-background p-2.5 text-xs">
            <div class="flex items-center gap-1.5 font-medium"><Loader2 class="w-3.5 h-3.5 animate-spin" /> Đổi giọng Brand Marketing</div>
            <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full w-[29%] bg-primary"></div></div>
            <div class="mt-1 text-muted-foreground tabular-nums">34/116 mục · còn ~8 phút</div>
          </div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.17 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- ─── 0. Lối vào: menu ⋯ ở Thư viện + bánh răng ở màn nghe ─── -->
        <section v-if="mode === 'entry'" class="flex-1 grid grid-cols-2 gap-8 p-8 min-w-0">
          <div>
            <div class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">Thư viện · menu ⋯ trên bìa</div>
            <div class="flex gap-4">
              <div class="w-36">
                <div class="aspect-[3/4] rounded-lg overflow-hidden shadow"><WfBookCover title="Brand Marketing" author="Tác giả mẫu" /></div>
                <div class="mt-2 text-sm font-medium truncate">Brand Marketing</div>
              </div>
              <div class="w-56 h-fit rounded-lg border border-border bg-popover shadow-lg py-1 text-sm">
                <span class="flex items-center gap-2 px-3 py-1.5 bg-primary/10 text-primary font-medium"><Pencil class="w-4 h-4" /> Sửa sách</span>
                <span class="flex items-center gap-2 px-3 py-1.5"><Package class="w-4 h-4" /> Xuất gói zip</span>
                <span class="flex items-center gap-2 px-3 py-1.5"><FolderOpen class="w-4 h-4" /> Mở thư mục</span>
                <span class="flex items-center gap-2 px-3 py-1.5 text-destructive">Xoá</span>
              </div>
            </div>
            <p class="mt-3 text-xs text-muted-foreground leading-relaxed">"Sửa thông tin" cũ đổi thành "Sửa sách", mở màn Sửa sách ở tab Nội dung. Hộp Sửa thông tin nhỏ bỏ, gộp vào tab Thông tin & bìa.</p>
          </div>
          <div>
            <div class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">Màn nghe · bánh răng góc trên phải</div>
            <div class="w-64 rounded-lg border border-border bg-popover shadow-lg py-1 text-sm">
              <span class="flex items-center gap-2 px-3 py-1.5 bg-primary/10 text-primary font-medium"><Pencil class="w-4 h-4" /> Sửa mục đang nghe</span>
              <span class="flex items-center gap-2 px-3 py-1.5"><Package class="w-4 h-4" /> Xuất gói zip</span>
              <span class="flex items-center gap-2 px-3 py-1.5"><FolderOpen class="w-4 h-4" /> Mở thư mục sách</span>
              <span class="flex items-center gap-2 px-3 py-1.5 text-destructive">Xoá sách</span>
            </div>
            <p class="mt-3 text-xs text-muted-foreground leading-relaxed">Mở màn Sửa sách, chọn sẵn đúng tiểu mục đang phát, sách tự dừng. Nghe thấy lỗi là sửa được ngay.</p>
          </div>
        </section>

        <section v-else class="flex-1 flex flex-col min-w-0">
          <!-- Đầu màn: quay lại + tên sách + tab -->
          <div class="px-6 pt-5 pb-0 border-b border-border">
            <div class="flex items-center gap-3">
              <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Quay lại</span>
              <div class="h-9 w-7 rounded overflow-hidden shrink-0"><WfBookCover :title="bookTitle" author="Tác giả mẫu" size="sm" /></div>
              <div class="min-w-0">
                <h1 class="text-base font-semibold leading-tight truncate">Sửa sách · Brand Marketing</h1>
                <p class="text-xs text-muted-foreground flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng Hải Đăng · 116 mục · 1 giờ 15 phút</p>
              </div>
              <div v-if="pending" class="ml-auto flex items-center gap-2">
                <span class="text-xs text-muted-foreground">{{ mode === 'rendering' ? 'Đang đọc lại 1/2 mục…' : `${pending} mục đã sửa chưa đọc lại` }}</span>
                <Button size="sm" :disabled="mode === 'rendering'"><RotateCcw class="w-4 h-4" /> Đọc lại {{ pending }} mục</Button>
              </div>
            </div>
            <div class="mt-4 flex gap-5 text-sm">
              <span v-for="t in ([['content', 'Nội dung'], ['info', 'Thông tin & bìa'], ['voice', 'Giọng đọc']] as const)" :key="t[0]"
                class="pb-2.5 -mb-px border-b-2" :class="tab === t[0] ? 'border-primary text-foreground font-medium' : 'border-transparent text-muted-foreground'">{{ t[1] }}</span>
            </div>
          </div>

          <!-- ─── Tab Nội dung ─── -->
          <div v-if="tab === 'content'" class="flex-1 flex min-h-0">
            <div class="w-72 shrink-0 border-r border-border overflow-auto py-2">
              <template v-for="c in chapters" :key="c.t">
                <div class="flex items-center gap-1.5 px-3 py-2 text-xs font-semibold text-muted-foreground">
                  <component :is="c.open ? ChevronDown : ChevronRight" class="w-3.5 h-3.5 shrink-0" /><span class="truncate">{{ c.t }}</span>
                </div>
                <template v-if="c.open">
                  <div v-for="(s, i) in c.secs" :key="s.t" class="flex items-center gap-2 pl-8 pr-3 py-2 text-sm"
                    :class="i === 1 ? 'bg-primary/10 text-primary font-medium' : ''">
                    <span class="truncate flex-1">{{ s.t }}</span>
                    <span v-if="s.state === 'edited'" class="shrink-0 h-2 w-2 rounded-full bg-rag-amber" title="Đã sửa, chưa đọc lại"></span>
                    <Loader2 v-else-if="s.state === 'rendering'" class="w-3.5 h-3.5 shrink-0 animate-spin" />
                    <span v-else-if="s.state === 'queued'" class="shrink-0 text-[11px] text-muted-foreground">chờ</span>
                    <Check v-else-if="s.state === 'done'" class="w-3.5 h-3.5 shrink-0 text-rag-green" />
                    <span v-else class="text-xs tabular-nums text-muted-foreground shrink-0">{{ s.d }}</span>
                  </div>
                </template>
              </template>
              <p class="px-3 pt-3 text-[11px] text-muted-foreground leading-relaxed flex gap-1.5"><span class="mt-1 h-2 w-2 rounded-full bg-rag-amber shrink-0"></span> Đã sửa, chưa đọc lại. Rời màn này vẫn giữ bản sửa.</p>
            </div>

            <div class="flex-1 flex flex-col min-w-0 p-5 gap-3">
              <div class="text-xs text-muted-foreground">Chương 1 · Tiểu mục 2/4</div>
              <label class="block">
                <span class="text-xs font-medium text-muted-foreground">Tiêu đề tiểu mục</span>
                <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm" value="Bốn áp lực buộc phải có thương hiệu" />
              </label>
              <label class="flex-1 flex flex-col min-h-0">
                <span class="text-xs font-medium text-muted-foreground flex items-center justify-between">
                  Lời đọc
                  <span class="font-normal">Sửa chữ sai, cách đọc tên riêng, số, viết tắt. Xuống dòng = nghỉ dài.</span>
                </span>
                <textarea class="mt-1 flex-1 w-full rounded-md border bg-background p-3 text-sm leading-relaxed resize-none"
                  :class="mode === 'dirty' ? 'border-rag-amber' : 'border-input'" :value="text" :readonly="mode === 'rendering'"></textarea>
              </label>
              <p v-if="mode === 'dirty'" class="text-xs text-muted-foreground flex items-center gap-1.5"><AlertCircle class="w-3.5 h-3.5 text-rag-amber" /> Đã sửa "Niu-sen" → "Nielsen". Chưa đọc lại: khi nghe vẫn là bản cũ.</p>
              <p v-else-if="mode === 'rendering'" class="text-xs text-muted-foreground flex items-center gap-1.5"><Loader2 class="w-3.5 h-3.5 animate-spin" /> Đang đọc lại mục này, khoảng 20 giây. Anh chọn mục khác sửa tiếp được.</p>
              <div class="flex flex-wrap items-center gap-2 pt-1">
                <Button variant="outline" size="sm"><Headphones class="w-4 h-4" /> Nghe bản đang có</Button>
                <Button variant="outline" size="sm" :disabled="mode === 'rendering'"><Play class="w-4 h-4" /> Nghe thử đoạn chọn</Button>
                <span class="text-[11px] text-muted-foreground">Bôi đen một câu rồi bấm để nghe thử trước khi đọc lại cả mục.</span>
                <div class="ml-auto flex gap-2">
                  <Button v-if="mode === 'dirty'" variant="ghost" size="sm">Bỏ sửa</Button>
                  <Button size="sm" :disabled="mode !== 'dirty'"><RotateCcw class="w-4 h-4" /> Đọc lại mục này</Button>
                </div>
              </div>
            </div>
          </div>

          <!-- ─── Tab Thông tin & bìa ─── -->
          <div v-else-if="tab === 'info'" class="flex-1 flex min-h-0 p-6 gap-8">
            <div class="w-[380px] space-y-3">
              <label class="block"><span class="text-xs font-medium text-muted-foreground">Tên sách</span>
                <input class="mt-1 w-full h-9 rounded-md border border-rag-amber bg-background px-3 text-sm" value="Brand Marketing căn bản" /></label>
              <label class="block"><span class="text-xs font-medium text-muted-foreground">Tác giả</span>
                <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm" value="Tác giả mẫu" /></label>
              <label class="block"><span class="text-xs font-medium text-muted-foreground">Danh mục</span>
                <span class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm flex items-center justify-between">Kinh doanh <ChevronDown class="w-4 h-4 text-muted-foreground" /></span></label>
              <div class="grid grid-cols-[1fr_90px] gap-2">
                <label class="block"><span class="text-xs font-medium text-muted-foreground">Bộ sách</span>
                  <span class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm flex items-center justify-between text-muted-foreground">Không thuộc bộ nào <ChevronDown class="w-4 h-4" /></span></label>
                <label class="block"><span class="text-xs font-medium text-muted-foreground">Tập số</span>
                  <input class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm" disabled /></label>
              </div>
              <div class="rounded-lg bg-muted/60 p-3 text-xs leading-relaxed space-y-1.5">
                <div class="font-medium text-foreground">Lưu xong, Sano tự làm thêm:</div>
                <div class="flex gap-1.5"><Wand2 class="w-3.5 h-3.5 shrink-0 mt-0.5" /> Vẽ lại bìa theo tên mới (bìa đang là bìa tự vẽ)</div>
                <div class="flex gap-1.5"><RotateCcw class="w-3.5 h-3.5 shrink-0 mt-0.5" /> Đọc lại lời giới thiệu: "Cuốn sách: Brand Marketing căn bản." (vài giây)</div>
              </div>
              <div class="flex gap-2 pt-1">
                <Button size="sm">Lưu</Button>
                <Button variant="ghost" size="sm">Huỷ</Button>
              </div>
            </div>
            <div>
              <div class="text-xs font-medium text-muted-foreground mb-2">Bìa</div>
              <div class="flex items-end gap-4">
                <div class="text-center">
                  <div class="w-28 aspect-[3/4] rounded-lg overflow-hidden shadow opacity-60"><WfBookCover title="Brand Marketing" author="Tác giả mẫu" /></div>
                  <div class="mt-1 text-[11px] text-muted-foreground">Bìa cũ</div>
                </div>
                <ChevronRight class="w-5 h-5 mb-16 text-muted-foreground" />
                <div class="text-center">
                  <div class="w-40 aspect-[3/4] rounded-lg overflow-hidden shadow-lg"><WfBookCover title="Brand Marketing căn bản" author="Tác giả mẫu" /></div>
                  <div class="mt-1 text-[11px] text-muted-foreground">Bìa mới, vẽ theo tên</div>
                </div>
              </div>
              <div class="mt-4 flex gap-2">
                <Button variant="outline" size="sm"><ImagePlus class="w-4 h-4" /> Chọn ảnh bìa…</Button>
                <Button variant="ghost" size="sm" disabled><Wand2 class="w-4 h-4" /> Dùng bìa tự vẽ</Button>
              </div>
              <p class="mt-2 text-[11px] text-muted-foreground max-w-xs leading-relaxed">Ảnh jpg, png, webp. Đã chọn ảnh thì đổi tên không vẽ lại bìa; bấm "Dùng bìa tự vẽ" để quay về bìa theo tên.</p>
            </div>
          </div>

          <!-- ─── Tab Giọng đọc ─── -->
          <div v-else class="flex-1 flex flex-col min-h-0 p-6 gap-4">
            <template v-if="mode === 'voice'">
              <p class="text-sm">Đang đọc bằng giọng <b>Hải Đăng</b>. Chọn giọng mới, nghe thử một đoạn của chính cuốn này rồi đọc lại cả cuốn.</p>
              <div class="flex gap-2 text-xs">
                <span class="h-7 px-3 rounded-full bg-foreground text-background flex items-center">Khuyên dùng</span>
                <span class="h-7 px-3 rounded-full border border-border flex items-center">Miền Bắc</span>
                <span class="h-7 px-3 rounded-full border border-border flex items-center">Miền Trung</span>
                <span class="h-7 px-3 rounded-full border border-border flex items-center">Miền Nam</span>
              </div>
              <div class="grid grid-cols-3 gap-2 max-w-3xl">
                <div v-for="v in voices" :key="v.n" class="rounded-lg border p-3 flex items-center gap-3"
                  :class="v.n === newVoice ? 'border-primary bg-primary/5' : 'border-border'" @click="newVoice = v.n">
                  <span class="h-8 w-8 rounded-full grid place-items-center border border-border shrink-0"><Play class="w-3.5 h-3.5" /></span>
                  <span class="min-w-0 flex-1">
                    <span class="text-sm font-medium block">{{ v.n }} <span v-if="v.n === 'Hải Đăng'" class="text-[11px] font-normal text-muted-foreground">· đang dùng</span></span>
                    <span class="text-[11px] text-muted-foreground">{{ v.s }}<template v-if="v.rec"> · Khuyên dùng</template></span>
                  </span>
                  <Check v-if="v.n === newVoice" class="w-4 h-4 text-primary shrink-0" />
                </div>
              </div>
              <div class="rounded-lg border border-border p-3 max-w-3xl flex items-center gap-3">
                <span class="h-9 w-9 rounded-full bg-primary text-primary-foreground grid place-items-center shrink-0"><Play class="w-4 h-4" /></span>
                <div class="text-sm min-w-0">
                  <div class="font-medium">Nghe thử giọng {{ newVoice }} với đoạn trong sách</div>
                  <div class="text-xs text-muted-foreground truncate">"Chuyên đề này mở bằng một ví dụ rất đời. Vì sao mẹ không bảo ra mua chai nước rửa chén…"</div>
                </div>
              </div>
              <div class="mt-auto flex items-center gap-3 border-t border-border pt-4">
                <Button><RotateCcw class="w-4 h-4" /> Đọc lại cả cuốn bằng giọng {{ newVoice }}</Button>
                <span class="text-xs text-muted-foreground leading-relaxed">116 mục · khoảng 11 phút trên máy này. Chạy nền, trong lúc đó vẫn nghe bản cũ được.<br />Chỗ nghe dở, danh mục, bìa giữ nguyên. File đã chép sang điện thoại cần chép lại.</span>
              </div>
            </template>
            <template v-else>
              <div class="max-w-xl rounded-lg border border-border p-4">
                <div class="flex items-center gap-2 font-medium"><Loader2 class="w-4 h-4 animate-spin" /> Đang đọc lại cả cuốn bằng giọng Mỹ Duyên</div>
                <div class="mt-3 h-2 rounded-full bg-muted overflow-hidden"><div class="h-full w-[29%] bg-primary rounded-full"></div></div>
                <div class="mt-1.5 flex justify-between text-xs text-muted-foreground tabular-nums"><span>34/116 mục</span><span>còn khoảng 8 phút</span></div>
                <p class="mt-3 text-xs text-muted-foreground leading-relaxed">Xong hết mới thay giọng cho cả cuốn, nên nghe lúc này vẫn là giọng Hải Đăng. Anh rời màn này hoặc đóng Sano: lần mở sau Sano đọc tiếp từ mục 35, không đọc lại từ đầu.</p>
                <div class="mt-3 flex gap-2">
                  <Button variant="outline" size="sm"><Square class="w-4 h-4" /> Dừng, giữ giọng cũ</Button>
                </div>
              </div>
            </template>
          </div>
        </section>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center leading-relaxed">
      D11 · Sửa sách: không phải xoá rồi tạo lại. Sửa lời đọc từng tiểu mục, chỉ đọc lại mục đã sửa. Đổi tên thì vẽ lại bìa tự vẽ và đọc lại lời giới thiệu.
      Đổi giọng đọc lại cả cuốn, chạy nền, xong mới thay. Sau mỗi lần đọc lại, gói zip tự cập nhật; M4B làm lại khi bấm Nghe trên điện thoại.
    </p>
  </div>
</template>
