<script setup lang="ts">
// D11 — Màn nghe vừa mọi cỡ cửa sổ, KHÔNG cuộn, KHÔNG bị cắt.
// Đo chiều cao thật của cột giữa (ResizeObserver) rồi chọn cách xếp:
//   - Rộng  (≥ 540 px): bìa to ở giữa như hiện tại
//   - Vừa   (≥ 380 px): bìa nhỏ bên trái, tên sách / giọng / tiểu mục bên phải
//   - Hẹp   (<  380 px): ẩn bìa, khung Lời đọc 1 dòng
// Gộp hàng: tốc độ (trái) · nút phát (giữa) · Nghỉ (phải) trên cùng một hàng;
// bỏ dòng "Tua −15s / +30s…"; xác nhận đổi quãng nghỉ thành thông báo nhỏ tự tắt.
// Cửa sổ nhỏ nhất 960×640 rơi vào cách xếp Vừa. Hẹp chỉ gặp khi chữ hệ thống to
// hoặc thu phóng; vẫn phải vừa.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để đổi cỡ cửa sổ.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, ChevronLeft, ChevronDown, Check, Gauge,
  Play, SkipBack, SkipForward, RotateCcw, RotateCw, Volume2, Mic, Timer, Smartphone,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

const sizes = [
  { key: 'big', label: '1. 1280×860', w: 1280, h: 860 },
  { key: 'default', label: '2. 1100×720 (mặc định)', w: 1100, h: 720 },
  { key: 'min', label: '3. 960×640 (nhỏ nhất)', w: 960, h: 640 },
  { key: 'tiny', label: '4. Chữ to / thu phóng (960×540)', w: 960, h: 540 },
] as const
const q = new URLSearchParams(window.location.search)
const size = ref<(typeof sizes)[number]['key']>((q.get('size') as never) || 'default')
const frame = computed(() => sizes.find((s) => s.key === size.value)!)
const dark = ref(false)
const toast = ref(q.get('toast') === '1')

// Đo chỗ trống thật của cột giữa (giữa hàng "Thư viện" và hàng nút dưới).
const mid = ref<HTMLElement | null>(null)
const midH = ref(0)
let ro: ResizeObserver | null = null
onMounted(() => {
  ro = new ResizeObserver(([e]) => (midH.value = Math.round(e.contentRect.height)))
  if (mid.value) ro.observe(mid.value)
})
onBeforeUnmount(() => ro?.disconnect())
const tier = computed(() => (midH.value >= 540 ? 'wide' : midH.value >= 380 ? 'medium' : 'narrow'))
const tierLabel = computed(() => ({ wide: 'Rộng', medium: 'Vừa', narrow: 'Hẹp' })[tier.value])
watch(size, () => (toast.value = false))

const toc = [
  { t: 'Ứng phó kinh tế suy thoái', d: '0:04', heard: true },
  { t: 'Lời mở đầu', d: '0:51', heard: true },
  { t: 'Một nửa lớp đang bị ảnh hưởng nặng', d: '0:39', cur: true },
  { t: 'Bốn việc của buổi học', d: '0:27' },
  { t: 'Đọc từ đâu', d: '0:25' },
  { t: 'Dịch từng dòng thành hệ quả', d: '1:37' },
  { t: 'Chuỗi domino thất nghiệp', d: '1:00' },
  { t: 'Ba ý cần nhớ', d: '0:31' },
  { t: 'Hậu Giang và Bắc Ninh', d: '0:37' },
  { t: 'Điểm sáng thứ nhất: tỉnh nào tăng', d: '1:25' },
  { t: 'Điểm sáng thứ hai: nông nghiệp', d: '0:54' },
  { t: 'Điểm sáng thứ ba: công nghiệp', d: '0:31' },
  { t: 'Điểm sáng thứ tư: bán lẻ và du lịch', d: '0:33' },
  { t: 'Ba ý cần nhớ', d: '0:35' },
]
const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách nói', icon: FilePlus2 },
  { key: 'stats', label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="s in sizes" :key="s.key" class="h-8 px-3 rounded-full border text-xs"
        :class="size === s.key ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="size = s.key">{{ s.label }}</button>
      <button class="h-8 px-3 rounded-full border text-xs" :class="toast ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="toast = !toast">5. Vừa đổi quãng nghỉ</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>
    <p class="text-xs text-muted-foreground">Cột giữa còn <b class="text-foreground tabular-nums">{{ midH }} px</b> → cách xếp <b class="text-foreground">{{ tierLabel }}</b></p>

    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col shrink-0" :style="{ width: frame.w + 'px', height: frame.h + 'px' }">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span><span class="h-3 w-3 rounded-full bg-rag-amber"></span><span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>
      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /><span class="font-semibold tracking-tight">Sano</span></div>
          <nav class="px-2 space-y-0.5">
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm" :class="n.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2"><span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span></div>
        </aside>

        <section class="flex-1 flex min-h-0 min-w-0">
          <div class="flex-1 flex flex-col p-6 min-w-0 min-h-0">
            <div class="shrink-0 flex items-center justify-between">
              <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Thư viện</span>
              <span class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground"><Settings class="w-4 h-4" /></span>
            </div>

            <!-- Cột giữa: KHÔNG cuộn; cách xếp theo chỗ trống đo được -->
            <div ref="mid" class="relative flex-1 min-h-0 overflow-hidden flex flex-col items-center justify-center">
              <!-- Đầu: Rộng = bìa to giữa; Vừa = bìa nhỏ trái + chữ phải; Hẹp = chỉ chữ -->
              <template v-if="tier === 'wide'">
                <div class="w-36 aspect-[3/4] rounded-xl shadow-2xl overflow-hidden shrink-0"><WfBookCover title="Ứng phó kinh tế suy thoái" size="md" /></div>
                <h2 class="mt-4 text-lg font-semibold text-center">Ứng phó kinh tế suy thoái</h2>
                <p class="text-sm text-muted-foreground flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng Hải Đăng</p>
                <p class="text-sm text-muted-foreground">Một nửa lớp đang bị ảnh hưởng nặng</p>
              </template>
              <div v-else class="w-full max-w-md flex items-center gap-4">
                <div v-if="tier === 'medium'" class="w-20 aspect-[3/4] rounded-lg shadow-lg overflow-hidden shrink-0"><WfBookCover title="Ứng phó kinh tế suy thoái" size="sm" /></div>
                <div class="min-w-0">
                  <h2 class="text-lg font-semibold truncate">Ứng phó kinh tế suy thoái</h2>
                  <p class="text-sm text-muted-foreground flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng Hải Đăng</p>
                  <p class="text-sm text-muted-foreground truncate">Một nửa lớp đang bị ảnh hưởng nặng</p>
                </div>
              </div>

              <!-- Lời đọc: Hẹp chỉ 1 dòng -->
              <div class="mt-4 w-full max-w-md rounded-lg border border-border bg-muted/30 px-4 py-3 shrink-0">
                <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Lời đọc</span>
                <span class="mt-1 text-sm font-medium leading-snug" :class="tier === 'narrow' ? 'line-clamp-1' : 'line-clamp-2 h-[2.75em]'">Trước buổi học, ban điều hành lớp làm một cuộc khảo sát.</span>
                <span v-if="tier !== 'narrow'" class="mt-0.5 block text-sm text-muted-foreground line-clamp-1">Có 103 anh chị trả lời.</span>
              </div>

              <div class="mt-4 w-full max-w-md shrink-0">
                <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[6%] bg-primary rounded-full"></div></div>
                <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>0:01</span><span>-0:37</span></div>
              </div>

              <!-- MỚI: một hàng: tốc độ · nút phát · Nghỉ -->
              <div class="mt-3 w-full max-w-md grid grid-cols-[1fr_auto_1fr] items-center gap-2 shrink-0">
                <span class="justify-self-start h-8 px-2.5 rounded-full border border-border flex items-center gap-1 text-xs whitespace-nowrap"><Gauge class="w-3.5 h-3.5" />1×<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
                <div class="flex items-center gap-0.5">
                  <span class="h-9 w-9 grid place-items-center rounded-full"><SkipBack class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center rounded-full"><RotateCcw class="w-5 h-5" /></span>
                  <span class="h-12 w-12 grid place-items-center rounded-full bg-primary text-primary-foreground shadow"><Play class="w-5 h-5 ml-0.5" /></span>
                  <span class="h-9 w-9 grid place-items-center rounded-full"><RotateCw class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center rounded-full"><SkipForward class="w-5 h-5" /></span>
                </div>
                <span class="justify-self-end h-8 px-2.5 rounded-full border border-border flex items-center gap-1 text-xs whitespace-nowrap"><Timer class="w-3.5 h-3.5" />Nghỉ: Dài<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
              </div>

              <!-- Thông báo nhỏ tự tắt sau 4 giây (thay dòng xác nhận cố định) -->
              <div v-if="toast" class="absolute bottom-2 left-1/2 -translate-x-1/2 rounded-full bg-foreground text-background text-xs px-3 py-1.5 shadow-lg flex items-center gap-1.5 whitespace-nowrap">
                <Check class="w-3.5 h-3.5" /> Cuốn này nghỉ 3 giây giữa tiểu mục, 4 giây sang chương
              </div>
            </div>

            <div class="shrink-0 flex items-center gap-2 border-t border-border pt-4">
              <Button size="sm" variant="outline" class="border-primary/30 bg-primary/10 text-primary"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</Button>
            </div>
          </div>
          <div class="w-72 shrink-0 border-l border-border overflow-hidden">
            <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · 65 mục · 42 phút</div>
            <div v-for="(c, i) in toc" :key="i" class="flex items-center justify-between gap-2 px-4 py-2.5 text-sm" :class="c.cur ? 'bg-primary/10 text-primary font-medium' : c.heard ? 'text-muted-foreground' : ''">
              <span class="truncate flex items-center gap-2"><Volume2 v-if="c.cur" class="w-3.5 h-3.5 shrink-0" /><Check v-else-if="c.heard" class="w-3.5 h-3.5 shrink-0" /><span v-else class="w-3.5 shrink-0"></span>{{ c.t }}</span>
              <span class="text-xs tabular-nums shrink-0">{{ c.d }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D11 · Màn nghe tự chọn cách xếp theo chỗ trống thật của cột giữa, không cuộn, không cắt. Tốc độ và Nghỉ lên cùng hàng nút phát; bỏ dòng mô tả tua.
      Kèm kiểm tra tự động ở các cỡ 960×640 → 1440×900: mọi nút phải nằm trọn trong khung.
    </p>
  </div>
</template>
