<script setup lang="ts">
// D13 — Lời đọc phóng to NGAY TRONG khung màn nghe (để chụp ảnh khoe: thấy cả player,
// chữ chạy và mục lục). Thanh bên trái, mục lục bên phải giữ nguyên.
// Vòng 3: vòng 2 (khung tối kiểu NCT) đẹp riêng nhưng lạc tông với app → giữ nền sáng / tối của
// app, chỉ phủ màu bìa thật nhẹ lên cả cột giữa; chữ to đậm kiểu lời bài hát, màu chữ của app.
// - Hàng trên: bìa nhỏ + tên sách, giọng, tiểu mục; góc phải nút Thu lại / Xem cả lời.
// - Giữa: mỗi câu một khối như lời bài hát; câu đang đọc đậm, câu khác nhạt, mép trên/dưới
//   mờ dần; tự cuộn giữ câu đang đọc ở ~1/3 trên, bấm câu → nghe từ câu đó.
// - Dưới: thanh thời gian (giữa có dấu "Sano · sách nói"), hàng nút y như màn nghe thường.
// - Khung hẹp (cửa sổ nhỏ nhất): nút Thu lại, Nghỉ chỉ còn biểu tượng.
// - Vào: khung Lời đọc ở màn nghe có 2 nút "Phóng to" và "Xem cả lời". Ra: nút "Thu lại".
//   App nhớ chế độ đã chọn cho lần mở sau.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Câu tự chạy ~2,5 s/câu để xem hiệu ứng.
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, ChevronLeft, ChevronDown, Check, Gauge,
  Play, Pause, SkipBack, SkipForward, RotateCcw, RotateCw, Volume2, Mic, Timer, Smartphone, Maximize2, Minimize2, Expand,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

const sizes = [
  { key: 'big', label: '1280×860', w: 1280, h: 860 },
  { key: 'default', label: '1100×720 (mặc định)', w: 1100, h: 720 },
  { key: 'min', label: '960×640 (nhỏ nhất)', w: 960, h: 640 },
] as const
const q = new URLSearchParams(window.location.search)
const size = ref<(typeof sizes)[number]['key']>((q.get('size') as never) || 'default')
const frame = computed(() => sizes.find((s) => s.key === size.value)!)
const expanded = ref(q.get('mode') !== 'normal')
const dark = ref(false)
const playing = ref(true)

const title = 'Chánh niệm, nghệ thuật của sự có mặt'
const section = 'Điều thứ ba: vô ngã'
const paragraphs = [
  ['Khi cơn giận nổi lên, người mới tập thường ôm chặt lấy nó.', 'Như thể cơn giận chính là mình.'],
  ['Người tập lâu thấy khác.', 'Họ thấy cơn giận đến, ở lại một lúc, rồi đi.', 'Nó không phải là mình, nó chỉ là một thứ đang có mặt.'],
  ['Vô ngã không có nghĩa là mình biến mất.', 'Mình vẫn đây, vẫn thở, vẫn nghe, vẫn thương.', 'Chỉ là mình không còn đồng nhất với từng cảm xúc đi qua.'],
  ['Thử lần tới, khi thấy bực, hãy nói thầm: à, cơn bực đang có mặt.', 'Chỉ một câu thôi, khoảng cách giữa mình và cơn bực đã rộng ra.', 'Trong khoảng cách đó, mình có quyền chọn cách đáp lại.'],
]
const sentences = paragraphs.flatMap((p) => p)
let n = 0
const para = paragraphs.map((p) => p.map((t) => ({ t, i: n++ })))
const cur = ref(1)
let timer: number | undefined
watch(playing, () => {
  clearInterval(timer)
  if (playing.value) timer = window.setInterval(() => (cur.value = (cur.value + 1) % sentences.length), 2500)
}, { immediate: true })
onBeforeUnmount(() => clearInterval(timer))

const box = ref<HTMLElement | null>(null)
watch([cur, expanded], async () => {
  await nextTick()
  const el = box.value?.querySelector<HTMLElement>(`[data-i="${cur.value}"]`)
  if (el && box.value) box.value.scrollTo({ top: el.offsetTop - box.value.clientHeight / 3, behavior: 'smooth' })
})
const pct = computed(() => Math.round(((cur.value + 0.5) / sentences.length) * 100))
const clock = (s: number) => `${Math.floor(s / 60)}:${String(Math.round(s) % 60).padStart(2, '0')}`

// Bìa hàng trên theo cỡ cửa sổ (bản thật tính ~22% chiều cao cột giữa, 80–150 px).
// Khung hẹp (< ~520 px, cửa sổ nhỏ nhất): nút Nghỉ chỉ còn biểu tượng, lề nhỏ hơn.
const compact = computed(() => size.value === 'min')
const coverH = computed(() => ({ big: 150, default: 120, min: 96 })[size.value])

// Mục lục (áp cho mọi lúc, không riêng chế độ phóng to): đầu có thanh tiến độ cả cuốn;
// tên tiểu mục xuống tối đa 2 dòng thay vì cắt cụt; tên chương chữ nhỏ in hoa (đã có từ 0eed810).
const toc: { t: string; d: string; heard?: boolean; cur?: boolean; ch?: string; curCh?: boolean }[] = [
  { t: 'Ôm ấp cảm xúc', d: '0:58', heard: true, ch: 'Chương 4 · Làm bạn với cảm xúc' },
  { t: 'Năm chướng ngại trong chậu nước', d: '0:45', heard: true },
  { t: 'Ba chướng ngại làm tâm mờ và nặng', d: '0:38', heard: true },
  { t: 'Biết khi có, biết khi không', d: '0:40', heard: true },
  { t: 'Chấp nhận không phải là cam chịu', d: '0:56', heard: true },
  { t: 'Ba ý cần nhớ', d: '0:46', heard: true },
  { t: 'Không phải chuyện thần bí', d: '0:34', heard: true, ch: 'Chương 5 · Ba sự thật của đời sống', curCh: true },
  { t: 'Điều thứ nhất: vô thường', d: '0:32', heard: true },
  { t: 'Điều thứ hai: khổ đến từ bám chấp', d: '0:38', heard: true },
  { t: 'Điều thứ ba: vô ngã', d: '0:38', cur: true },
  { t: 'Ba ý cần nhớ', d: '0:36' },
  { t: 'Người ở cạnh thấy mình được nghe', d: '0:38', ch: 'Chương 6 · Quả ngọt của sự tập' },
  { t: 'Vững và phục hồi nhanh', d: '0:28' },
  { t: 'Hiểu người, sống giản dị, thấy nhẹ', d: '0:30' },
  { t: 'Hai lời dặn trên đường tập', d: '0:45' },
  { t: 'Ôn lại cả hành trình', d: '1:27' },
  { t: 'Lời dặn cuối', d: '0:28' },
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
      <button class="h-8 px-3 rounded-full border text-xs" :class="!expanded ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="expanded = false">A. Màn nghe thường</button>
      <button class="h-8 px-3 rounded-full border text-xs" :class="expanded ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="expanded = true">B. MỚI: Lời đọc phóng to trong khung</button>
      <span class="w-px bg-border mx-1"></span>
      <button v-for="s in sizes" :key="s.key" class="h-8 px-3 rounded-full border text-xs"
        :class="size === s.key ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="size = s.key">{{ s.label }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col shrink-0" :style="{ width: frame.w + 'px', height: frame.h + 'px' }">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span><span class="h-3 w-3 rounded-full bg-rag-amber"></span><span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>
      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /><span class="font-semibold tracking-tight">Sano</span></div>
          <nav class="px-2 space-y-0.5">
            <span v-for="it in nav" :key="it.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm" :class="it.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="it.icon" class="w-4 h-4" /> {{ it.label }}
            </span>
          </nav>
          <div class="px-2 mt-2"><span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span></div>
        </aside>

        <section class="flex-1 flex min-h-0 min-w-0">
          <div class="relative isolate overflow-hidden flex-1 flex flex-col p-6 min-w-0 min-h-0">
            <!-- B: lớp màu bìa thật nhẹ phủ cả cột giữa (bìa phóng to, làm mờ), tan dần về nền app -->
            <template v-if="expanded">
              <div class="absolute -inset-10 -z-10 blur-3xl saturate-150 opacity-[0.18] dark:opacity-[0.22]"><WfBookCover :title="title" size="lg" /></div>
              <div class="absolute inset-0 -z-10 bg-gradient-to-b from-transparent via-background/40 to-background"></div>
            </template>
            <div class="shrink-0 flex items-center justify-between">
              <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Thư viện</span>
              <span class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground"><Settings class="w-4 h-4" /></span>
            </div>

            <!-- B. MỚI (vòng 3): cùng tông app — nền sáng / tối của app, chỉ phủ nhẹ màu bìa; chữ lời đọc
                 to đậm kiểu thẻ lời bài hát nhưng dùng màu chữ của app; hàng nút y như màn nghe thường. -->
            <div v-if="expanded" class="flex-1 min-h-0 flex flex-col pt-4">
              <!-- Hàng trên: bìa nhỏ + tên sách, giọng, tiểu mục; góc phải nút Thu lại / Xem cả lời -->
              <div class="shrink-0 w-full max-w-2xl mx-auto flex items-center gap-4">
                <div class="rounded-lg shadow-lg overflow-hidden shrink-0" :style="{ height: coverH + 'px', width: Math.round(coverH * 0.75) + 'px' }"><WfBookCover :title="title" size="sm" /></div>
                <div class="min-w-0 flex-1">
                  <h2 class="text-lg font-semibold leading-snug line-clamp-2">{{ title }}</h2>
                  <p class="mt-0.5 text-sm text-muted-foreground flex items-center gap-1 min-w-0"><Mic class="w-3.5 h-3.5 shrink-0" /><span class="truncate">Giọng Thiền Tâm Đức</span></p>
                  <p class="text-sm text-primary font-medium truncate">{{ section }}</p>
                </div>
                <div class="self-start flex items-center gap-1 text-xs text-muted-foreground">
                  <span class="h-8 rounded-md border border-border bg-background/70 hover:bg-muted flex items-center gap-1.5" :class="compact ? 'w-8 justify-center' : 'px-2.5'" title="Thu lại"><Minimize2 class="w-3.5 h-3.5" /><template v-if="!compact">Thu lại</template></span>
                  <span class="h-8 w-8 rounded-md border border-border bg-background/70 hover:bg-muted grid place-items-center" title="Xem cả lời"><Expand class="w-3.5 h-3.5" /></span>
                </div>
              </div>

              <!-- Lời đọc: mỗi câu một khối, câu đang đọc đậm màu chữ app, câu khác nhạt; mép trên / dưới mờ dần -->
              <div ref="box" class="relative mt-3 flex-1 min-h-0 w-full max-w-2xl mx-auto overflow-auto [scrollbar-width:none] [mask-image:linear-gradient(to_bottom,transparent,black_14%,black_80%,transparent)]">
                <div class="py-[16%] space-y-2.5">
                  <template v-for="(p, pi) in para" :key="pi">
                    <p v-for="(s, si) in p" :key="s.i" :data-i="s.i" class="text-[23px] leading-[1.35] font-bold tracking-[-0.01em] cursor-pointer transition-colors duration-500"
                      :class="[s.i === cur ? 'text-foreground' : s.i < cur ? 'text-foreground/20' : 'text-foreground/35 hover:text-foreground/60', si === p.length - 1 && pi < para.length - 1 ? 'pb-4' : '']"
                      @click="cur = s.i">{{ s.t }}</p>
                  </template>
                  <p class="pt-4 text-sm text-muted-foreground">Hết tiểu mục · tiếp theo: Ba ý cần nhớ</p>
                </div>
              </div>

              <!-- Thanh thời gian + hàng nút: y như màn nghe thường -->
              <div class="mt-2 w-full max-w-md mx-auto shrink-0">
                <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full transition-all" :style="{ width: pct + '%' }"></div></div>
                <div class="flex items-center justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>{{ clock(38 * pct / 100) }}</span>
                  <!-- Dấu Sano nhỏ (sau này ảnh chia sẻ có sẵn thương hiệu) -->
                  <span class="flex items-center gap-1 font-medium"><img src="@/assets/favicon.svg" alt="" class="h-3.5 w-3.5 rounded-[3px]" /> Sano · sách nói</span>
                  <span>-{{ clock(38 - 38 * pct / 100) }}</span></div>
              </div>
              <div class="mt-3 w-full max-w-md mx-auto grid grid-cols-[1fr_auto_1fr] items-center gap-2 shrink-0">
                <span class="justify-self-start h-8 px-2.5 rounded-full border border-border bg-background/70 flex items-center gap-1 text-xs whitespace-nowrap"><Gauge class="w-3.5 h-3.5" />1×<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
                <div class="flex items-center gap-0.5">
                  <span class="h-9 w-9 grid place-items-center rounded-full"><SkipBack class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center rounded-full"><RotateCcw class="w-5 h-5" /></span>
                  <button class="mx-1 h-12 w-12 grid place-items-center rounded-full bg-primary text-primary-foreground shadow" @click="playing = !playing"><Pause v-if="playing" class="w-5 h-5" /><Play v-else class="w-5 h-5 ml-0.5" /></button>
                  <span class="h-9 w-9 grid place-items-center rounded-full"><RotateCw class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center rounded-full"><SkipForward class="w-5 h-5" /></span>
                </div>
                <span class="justify-self-end h-8 px-2.5 rounded-full border border-border bg-background/70 flex items-center gap-1 text-xs whitespace-nowrap" title="Nghỉ: theo chung"><Timer class="w-3.5 h-3.5" /><template v-if="!compact">Nghỉ: theo chung</template><ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
              </div>
            </div>

            <!-- A. Màn nghe thường: khung Lời đọc có thêm nút Phóng to -->
            <div v-else class="flex-1 min-h-0 flex flex-col items-center justify-center py-2">
              <div class="rounded-xl shadow-2xl overflow-hidden shrink-0" :style="{ height: '214px', width: '160px' }"><WfBookCover :title="title" size="md" /></div>
              <h2 class="mt-4 text-lg font-semibold text-center">{{ title }}</h2>
              <p class="text-sm text-muted-foreground flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng Thiền Tâm Đức</p>
              <p class="text-sm text-muted-foreground">{{ section }}</p>
              <div class="mt-4 w-full max-w-md rounded-lg border border-border bg-muted/30 px-4 py-3 shrink-0">
                <span class="flex items-center justify-between text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                  Lời đọc
                  <span class="flex items-center gap-3 normal-case font-normal tracking-normal text-primary">
                    <button class="flex items-center gap-1 hover:underline" @click="expanded = true"><Maximize2 class="w-3 h-3" /> Phóng to</button>
                    <span class="flex items-center gap-1">Xem cả lời →</span>
                  </span>
                </span>
                <span class="mt-1 block text-sm font-medium leading-snug h-[2.75em] line-clamp-2">{{ sentences[cur] }}</span>
                <span class="mt-0.5 block h-[1.375em] text-sm text-muted-foreground leading-snug line-clamp-1">{{ sentences[cur + 1] }}</span>
              </div>
              <div class="mt-4 w-full max-w-md shrink-0">
                <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full" :style="{ width: pct + '%' }"></div></div>
                <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>{{ clock(38 * pct / 100) }}</span><span>-{{ clock(38 - 38 * pct / 100) }}</span></div>
              </div>
              <div class="mt-3 w-full max-w-md grid grid-cols-[1fr_auto_1fr] items-center gap-2 shrink-0">
                <span class="justify-self-start h-8 px-2.5 rounded-full border border-border flex items-center gap-1 text-xs"><Gauge class="w-3.5 h-3.5" />1×<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
                <div class="flex items-center gap-0.5">
                  <span class="h-9 w-9 grid place-items-center"><SkipBack class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center"><RotateCcw class="w-5 h-5" /></span>
                  <span class="mx-1 h-12 w-12 grid place-items-center rounded-full bg-primary text-primary-foreground shadow"><Pause class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center"><RotateCw class="w-5 h-5" /></span>
                  <span class="h-9 w-9 grid place-items-center"><SkipForward class="w-5 h-5" /></span>
                </div>
                <span class="justify-self-end h-8 px-2.5 rounded-full border border-border flex items-center gap-1 text-xs whitespace-nowrap"><Timer class="w-3.5 h-3.5" />Nghỉ: theo chung<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
              </div>
            </div>

            <div class="shrink-0 flex items-center gap-2 border-t border-border pt-4">
              <Button size="sm" variant="outline" class="border-primary/30 bg-primary/10 text-primary"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</Button>
            </div>
          </div>
          <div class="w-72 shrink-0 border-l border-border overflow-hidden">
            <!-- Đầu mục lục: tiến độ cả cuốn -->
            <div class="px-4 pt-3 pb-2.5 border-b border-border/60">
              <div class="flex items-baseline justify-between text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục <span class="normal-case font-normal tracking-normal tabular-nums">39 mục · 24 phút</span></div>
              <div class="mt-2 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full w-[24%] bg-primary rounded-full"></div></div>
              <div class="mt-1.5 text-[11px] text-muted-foreground tabular-nums">Đã nghe <b class="font-semibold text-foreground">9/39</b> mục · còn 14 phút</div>
            </div>
            <template v-for="(c, i) in toc" :key="i">
              <div v-if="c.ch" class="px-4 pt-4 pb-1 text-[11px] font-semibold uppercase tracking-wider truncate" :class="c.curCh ? 'text-primary' : 'text-muted-foreground'">{{ c.ch }}</div>
              <div class="flex items-baseline justify-between gap-2 px-4 py-2 text-sm leading-snug" :class="c.cur ? 'bg-primary/10 text-primary font-medium' : c.heard ? 'text-muted-foreground' : ''">
                <span class="flex items-baseline gap-2 min-w-0">
                  <span class="self-start w-3.5 h-[1.375em] shrink-0 grid place-items-center"><Volume2 v-if="c.cur" class="w-3.5 h-3.5" /><Check v-else-if="c.heard" class="w-3.5 h-3.5" /></span>
                  <span class="line-clamp-2">{{ c.t }}</span>
                </span>
                <span class="text-xs tabular-nums shrink-0">{{ c.d }}</span>
              </div>
            </template>
          </div>
        </section>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D13 · Lời đọc phóng to trong khung, cùng tông app: phủ nhẹ màu bìa, chữ to đậm kiểu lời bài hát, câu đang đọc đậm; bìa và tên sách thu lên hàng trên;
      hàng nút nằm trong cùng khung. Thanh bên và mục lục giữ nguyên để một tấm ảnh chụp thấy đủ. "Xem cả lời" (toàn vùng nội dung) vẫn giữ như cũ.
    </p>
  </div>
</template>
