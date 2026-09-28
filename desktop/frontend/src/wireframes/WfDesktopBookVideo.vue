<script setup lang="ts">
// D15 — Tạo video cả cuốn (đăng YouTube, video thường trên Facebook).
// Vào: menu bánh răng ở màn nghe → "Tạo video cả cuốn…" (cạnh Xuất gói zip, Lưu file M4B).
// Video gồm 4 cảnh: (1) Intro đoạn hay nhất 15–30 giây — tuỳ chọn, người dùng tự chọn câu;
// (2) Màn tựa ~3 giây; (3) Cả cuốn (hoặc từ chương … đến chương …) — mỗi chương mở bằng thẻ
// tên chương, trong lúc nghe hiện câu đang đọc chữ to (thế mạnh chữ chạy mà Fonos không có),
// câu kế nhạt, bìa + tên sách + chương, thanh tiến độ cả cuốn có vạch chương; (4) Màn kết 5 giây
// "Tạo sách nói của bạn · sanobook.com".
// Khung mặc định Ngang 16:9 1920×1080 (YouTube); có Dọc 9:16. Chữ chính nằm giữa khung, chừa
// ~15% đáy (YouTube phủ thanh điều khiển, phụ đề). Logo góc trên phải như Fonos.
// Kèm theo (bật sẵn): ảnh thumbnail 1280×720, phụ đề .srt từ lời đọc (lên tìm kiếm tốt hơn),
// mô tả YouTube có mốc chương (YouTube tự chia chương) + link sanobook.com, chép bằng một nút.
// Tạo trên máy bằng ffmpeg: ghép tiếng như M4B, vẽ khung hình từng câu ghi dần ra đĩa;
// ~1–3 phút cho cuốn 27 phút, có tiến độ theo bước, huỷ được. Lưu vào một thư mục chọn sẵn.
// Bản làm thật: chỉ vài chương thì màn tựa / thumbnail ghi "NGHE THỬ SÁCH NÓI", lưu thư mục riêng
// "<Tên sách> - <Chương>" để không đè video cả cuốn.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API.
import { computed, ref } from 'vue'
import { Check, ChevronDown, Clapperboard, Copy, FileText, FolderOpen, Image as ImageIcon, Loader2, Mic, Sparkles, Subtitles, Sun, Moon, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

const q = new URLSearchParams(window.location.search)
type Scene = 'intro' | 'title' | 'main' | 'end' | 'thumb'
type Phase = 'setup' | 'busy' | 'done'
const scene = ref<Scene>((q.get('scene') as Scene) || 'main')
const phase = ref<Phase>((q.get('phase') as Phase) || 'setup')
const ratio = ref<'wide' | 'tall'>((q.get('ratio') as never) || 'wide')
const dark = ref(false)
const introOn = ref(true)
const range = ref<'all' | 'some'>('all')
const bg = ref(0)
const extras = ref({ thumb: true, srt: true, desc: true })

const title = 'Chánh niệm, nghệ thuật của sự có mặt'
const voice = 'Thiền Tâm Đức'
const chapters = [
  { t: 'Lời mở đầu', start: 0, dur: 76 },
  { t: 'Chương 1. Có mặt, món quà quý nhất', start: 76, dur: 250 },
  { t: 'Chương 2. Tập chánh niệm để được gì', start: 326, dur: 240 },
  { t: 'Chương 3. Tập thế nào', start: 566, dur: 260 },
  { t: 'Chương 4. Khi gặp vướng', start: 826, dur: 300 },
  { t: 'Chương 5. Điều được khai mở bên trong', start: 1126, dur: 190 },
  { t: 'Chương 6. Chân dung người tập lâu năm', start: 1316, dur: 304 },
]
const fromCh = ref(1)
const toCh = ref(6)
const total = 1620
const clock = (s: number) => `${Math.floor(s / 60)}:${String(Math.round(s) % 60).padStart(2, '0')}`
const bgs = [
  { label: 'Hoàng hôn', cls: 'bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800' },
  { label: 'Theo bìa', cls: 'bg-gradient-to-br from-amber-800 via-orange-900 to-stone-900' },
  { label: 'Biển', cls: 'bg-gradient-to-br from-cyan-500 via-teal-500 to-lime-500' },
  { label: 'Đêm', cls: 'bg-gradient-to-br from-indigo-700 via-blue-800 to-fuchsia-700' },
  { label: 'Trà', cls: 'bg-gradient-to-br from-stone-600 via-amber-800 to-stone-900' },
]
const sel = computed(() => (range.value === 'all' ? chapters : chapters.slice(fromCh.value, toCh.value + 1)))
const selDur = computed(() => sel.value.reduce((n, c) => n + c.dur, 0) + (introOn.value ? 18 : 0) + 8)
const sizeMB = computed(() => Math.round((selDur.value / 60) * 7))
const desc = computed(() =>
  [
    `${title} — Sách nói đầy đủ, giọng ${voice}.`,
    '',
    ...(introOn.value ? ['0:00 Trích đoạn'] : []),
    ...sel.value.map((c, i) => `${clock((introOn.value ? 21 : 3) + sel.value.slice(0, i).reduce((n, x) => n + x.dur, 0))} ${c.t}`),
    '',
    'Sách nói tạo bằng Sano — tự tạo sách nói miễn phí từ file Word: https://sanobook.com',
  ].join('\n'),
)
const introSentences = [
  { t: 'Món quà quý nhất mình tặng cho người thương, không phải là quà cáp hay tiền bạc.', on: true },
  { t: 'Mà là sự có mặt trọn vẹn của mình.', on: true },
  { t: 'Ngồi cạnh con mà đầu nghĩ chuyện công ty, thì thật ra mình đang vắng mặt.', on: true },
  { t: 'Con cảm nhận được điều đó, dù con không nói ra.', on: true },
  { t: 'Hãy tắt điện thoại, nhìn vào mắt người đối diện, và nghe.', on: false },
]
const waveBars = Array.from({ length: 36 }, (_, i) => 18 + Math.round(80 * Math.abs(Math.sin(i * 0.9) * Math.cos(i * 0.31))))
const scenes: { key: Scene; label: string }[] = [
  { key: 'intro', label: '1. Trích đoạn' },
  { key: 'title', label: '2. Màn tựa' },
  { key: 'main', label: '3. Trong sách' },
  { key: 'end', label: '4. Màn kết' },
  { key: 'thumb', label: 'Thumbnail' },
]
const steps = [
  { t: 'Ghép tiếng cả cuốn', state: 'done' },
  { t: 'Vẽ khung hình · 312/486 câu', state: 'run' },
  { t: 'Mã hoá video', state: 'wait' },
  { t: 'Thumbnail, phụ đề, mô tả', state: 'wait' },
]
const copied = ref(false)
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="p in (['setup', 'busy', 'done'] as const)" :key="p" class="h-8 px-3 rounded-full border text-xs"
        :class="phase === p ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="phase = p">
        {{ { setup: 'A. Tuỳ chọn', busy: 'B. Đang tạo', done: 'C. Xong' }[p] }}</button>
      <span class="w-px bg-border mx-1"></span>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs text-foreground" @click="ratio = ratio === 'wide' ? 'tall' : 'wide'">Khung: {{ ratio === 'wide' ? 'Ngang 16:9' : 'Dọc 9:16' }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground shrink-0 w-[1100px] h-[720px]">
      <div class="absolute inset-0 bg-black/45 backdrop-blur-[2px] z-10"></div>

      <div class="absolute z-20 inset-0 m-auto w-[1060px] h-[690px] rounded-2xl bg-background border border-border shadow-2xl flex overflow-hidden">
        <!-- Trái: xem trước từng cảnh -->
        <div class="w-[540px] shrink-0 bg-muted/50 flex flex-col items-center justify-center gap-3 p-5">
          <div class="flex gap-1 p-1 rounded-lg bg-muted text-xs">
            <button v-for="s in scenes" :key="s.key" class="h-7 px-2.5 rounded-md" :class="[scene === s.key ? 'bg-background shadow-sm font-medium' : 'text-muted-foreground', s.key === 'intro' && !introOn && 'line-through opacity-50']" @click="scene = s.key">{{ s.label }}</button>
          </div>

          <!-- Khung video (tỉ lệ thật) -->
          <div class="relative overflow-hidden rounded-xl shadow-2xl text-white isolate shrink-0" :class="ratio === 'wide' || scene === 'thumb' ? 'w-[500px] h-[281px]' : 'w-[264px] h-[470px]'">
            <div class="absolute inset-0 -z-10" :class="bgs[bg].cls"><div class="absolute inset-0 bg-gradient-to-b from-black/15 to-black/45"></div></div>
            <!-- Vùng YouTube phủ điều khiển / phụ đề (chỉ để xem, không có trong video) -->
            <div v-if="scene !== 'thumb'" class="absolute inset-x-0 bottom-0 h-[15%] border-t border-dashed border-white/30 bg-white/5 pointer-events-none"><span class="absolute right-2 top-0.5 text-[8px] text-white/50">vùng YouTube phủ điều khiển</span></div>

            <!-- Thương hiệu góc trên phải (mọi cảnh trừ màn kết) -->
            <div v-if="scene !== 'end'" class="absolute right-3.5 flex items-center gap-1.5" :class="scene === 'thumb' ? 'bottom-3' : 'top-3'">
              <img src="@/assets/favicon.svg" alt="" class="h-5 w-5 rounded" />
              <span class="leading-tight"><span class="block text-[10px] font-bold">Sano · <span class="font-normal text-white/80">Tự tạo sách nói</span></span><span class="block text-[8.5px] font-semibold text-white/85">sanobook.com</span></span>
            </div>

            <!-- 1. Trích đoạn -->
            <template v-if="scene === 'intro'">
              <div class="absolute inset-0 flex flex-col justify-center" :class="ratio === 'wide' ? 'px-10 pb-8' : 'px-6 pb-14'">
                <span class="w-fit rounded-full bg-white/20 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.15em] flex items-center gap-1"><Sparkles class="w-2.5 h-2.5" /> Trích đoạn</span>
                <p class="mt-3 font-bold leading-snug" :class="ratio === 'wide' ? 'text-[22px] max-w-[420px]' : 'text-[19px]'">Món quà quý nhất mình tặng cho người thương, không phải là quà cáp hay tiền bạc.</p>
                <div class="mt-3 flex items-center gap-[2px] h-5">
                  <span v-for="(h, i) in waveBars" :key="i" class="w-[2px] rounded-full" :class="i < 14 ? 'bg-white' : 'bg-white/35'" :style="{ height: h + '%' }"></span>
                </div>
                <p class="mt-3 text-[10px] text-white/75">{{ title }} · Giọng {{ voice }}</p>
              </div>
            </template>

            <!-- 2. Màn tựa -->
            <template v-else-if="scene === 'title'">
              <div class="absolute inset-0 flex items-center" :class="ratio === 'wide' ? 'gap-6 px-12 pb-6' : 'flex-col justify-center gap-4 px-6 pb-12 text-center'">
                <div class="rounded-lg shadow-2xl shadow-black/50 overflow-hidden ring-1 ring-white/20 shrink-0" :class="ratio === 'wide' ? 'w-[120px] h-[160px]' : 'w-[130px] h-[173px]'"><WfBookCover :title="title" size="md" /></div>
                <div>
                  <span class="text-[9px] font-bold uppercase tracking-[0.2em] text-white/70">Sách nói đầy đủ</span>
                  <h3 class="mt-1 font-serif font-bold leading-tight" :class="ratio === 'wide' ? 'text-[26px]' : 'text-[22px]'">{{ title }}</h3>
                  <p class="mt-2 text-[11px] text-white/80"><Mic class="inline w-3 h-3 -mt-0.5 mr-0.5" />Giọng {{ voice }} · {{ Math.round(selDur / 60) }} phút · {{ sel.length }} chương</p>
                </div>
              </div>
            </template>

            <!-- 3. Trong sách -->
            <template v-else-if="scene === 'main'">
              <div v-if="ratio === 'wide'" class="absolute inset-0 flex items-center gap-7 pl-10 pr-10 pt-12 pb-14">
                <div class="w-[104px] shrink-0">
                  <div class="w-[104px] h-[139px] rounded-lg shadow-2xl shadow-black/50 overflow-hidden ring-1 ring-white/20"><WfBookCover :title="title" size="sm" /></div>
                  <p class="mt-2 text-[10px] font-semibold leading-tight line-clamp-2">{{ title }}</p>
                  <p class="text-[8.5px] text-white/70">Giọng {{ voice }}</p>
                </div>
                <div class="min-w-0">
                  <p class="text-[8.5px] font-bold uppercase tracking-[0.15em] text-white/70 truncate">Chương 1 · Có mặt, món quà quý nhất</p>
                  <p class="text-[9.5px] text-white/60">Có mặt cho người thương</p>
                  <p class="mt-3 text-[20px] font-bold leading-snug">Ngồi cạnh con mà đầu nghĩ chuyện công ty, thì thật ra mình đang vắng mặt.</p>
                  <p class="mt-2 text-[15px] font-bold leading-snug text-white/35">Con cảm nhận được điều đó, dù con không nói ra.</p>
                </div>
              </div>
              <div v-else class="absolute inset-0 flex flex-col px-6 pt-14 pb-16">
                <div class="mx-auto w-[96px] h-[128px] rounded-lg shadow-2xl shadow-black/50 overflow-hidden ring-1 ring-white/20"><WfBookCover :title="title" size="sm" /></div>
                <p class="mt-4 text-[8.5px] font-bold uppercase tracking-[0.15em] text-white/70 truncate">Chương 1 · Có mặt, món quà quý nhất</p>
                <p class="mt-2 text-[18px] font-bold leading-snug">Ngồi cạnh con mà đầu nghĩ chuyện công ty, thì thật ra mình đang vắng mặt.</p>
                <p class="mt-2 text-[14px] font-bold leading-snug text-white/35">Con cảm nhận được điều đó, dù con không nói ra.</p>
                <p class="mt-auto text-[10px] font-semibold line-clamp-1">{{ title }}</p>
              </div>
              <!-- Thanh tiến độ cả cuốn + vạch chương -->
              <div class="absolute inset-x-6 bottom-[4%]">
                <div class="relative h-[3px] rounded-full bg-white/25">
                  <div class="absolute inset-y-0 left-0 w-[9%] rounded-full bg-white"></div>
                  <span v-for="c in chapters.slice(1)" :key="c.t" class="absolute -top-[2px] h-[7px] w-px bg-white/60" :style="{ left: (c.start / total) * 100 + '%' }"></span>
                </div>
                <div class="mt-1 flex justify-between text-[8px] text-white/70 tabular-nums"><span>2:31</span><span>27:00</span></div>
              </div>
            </template>

            <!-- Thumbnail 1280×720 (ảnh riêng, không nằm trong video): chữ to đọc được khi thu nhỏ -->
            <template v-else-if="scene === 'thumb'">
              <div class="absolute inset-0 flex items-center gap-6 pl-9 pr-8">
                <div class="w-[150px] h-[200px] rounded-lg shadow-2xl shadow-black/50 overflow-hidden ring-1 ring-white/25 shrink-0 -rotate-2"><WfBookCover :title="title" size="md" /></div>
                <div class="min-w-0">
                  <span class="inline-block rounded bg-white text-rose-700 px-2 py-0.5 text-[12px] font-black uppercase tracking-wide">Sách nói đầy đủ</span>
                  <h3 class="mt-2 text-[30px] font-black leading-[1.05] drop-shadow">{{ title }}</h3>
                  <p class="mt-2 text-[12px] font-semibold text-white/90">{{ Math.round(selDur / 60) }} phút · Giọng {{ voice }}</p>
                </div>
              </div>
            </template>

            <!-- 4. Màn kết -->
            <template v-else>
              <div class="absolute inset-0 flex flex-col items-center justify-center text-center gap-2 pb-8">
                <img src="@/assets/favicon.svg" alt="" class="h-14 w-14 rounded-xl shadow-xl" />
                <p class="mt-1 text-[22px] font-bold">Tạo sách nói của bạn</p>
                <p class="text-[11px] text-white/80">Miễn phí · chạy ngay trên máy · từ file Word</p>
                <span class="mt-2 rounded-full bg-white text-neutral-900 px-4 py-1.5 text-[13px] font-bold tracking-wide">sanobook.com</span>
              </div>
            </template>
          </div>
          <p class="text-[11px] text-muted-foreground text-center max-w-[520px]">
            {{ { intro: 'Trích đoạn 15–30 giây: câu đổi theo lời đọc, sóng âm sáng dần — người xem nghe thử trước khi vào sách.', title: 'Màn tựa ~3 giây.', main: 'Suốt cuốn: câu đang đọc chữ to, câu kế nhạt; tên chương; thanh tiến độ cả cuốn có vạch chương. Mỗi chương mở bằng thẻ tên chương.', end: 'Màn kết 5 giây, kêu gọi dùng Sano.', thumb: 'Ảnh thumbnail 1280×720 (file riêng, luôn khung ngang dù video dọc): chữ thật to, đọc được cả khi YouTube thu nhỏ trong danh sách gợi ý.' }[scene] }}
          </p>
        </div>

        <!-- Phải -->
        <div class="flex-1 min-w-0 flex flex-col">
          <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
            <div>
              <h2 class="font-semibold flex items-center gap-2"><Clapperboard class="w-4 h-4" /> Tạo video cả cuốn</h2>
              <p class="text-xs text-muted-foreground mt-0.5">Đăng YouTube, Facebook · {{ title }}</p>
            </div>
            <X class="w-4 h-4 text-muted-foreground" />
          </div>

          <!-- A. Tuỳ chọn -->
          <div v-if="phase === 'setup'" class="flex-1 min-h-0 overflow-auto px-5 py-3.5 space-y-3.5 text-sm">
            <div>
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Phần sách</span>
              <div class="mt-2 grid grid-cols-2 gap-2">
                <button class="h-11 rounded-lg border px-3 text-left" :class="range === 'all' ? 'border-primary bg-primary/5' : 'border-border'" @click="range = 'all'"><span class="block font-medium">Cả cuốn</span><span class="block text-[11px] text-muted-foreground">7 chương · 27 phút</span></button>
                <button class="h-11 rounded-lg border px-3 text-left" :class="range === 'some' ? 'border-primary bg-primary/5' : 'border-border'" @click="range = 'some'"><span class="block font-medium">Một vài chương</span><span class="block text-[11px] text-muted-foreground">Ví dụ chương 1 làm mồi</span></button>
              </div>
              <div v-if="range === 'some'" class="mt-2 flex items-center gap-2 text-xs">
                Từ <span class="h-8 px-2.5 rounded-md border border-border flex items-center gap-1 min-w-0 max-w-[150px]"><span class="truncate">{{ chapters[fromCh].t }}</span><ChevronDown class="w-3 h-3 shrink-0" /></span>
                đến <span class="h-8 px-2.5 rounded-md border border-border flex items-center gap-1 min-w-0 max-w-[150px]"><span class="truncate">{{ chapters[toCh].t }}</span><ChevronDown class="w-3 h-3 shrink-0" /></span>
              </div>
            </div>

            <div>
              <label class="flex items-center justify-between">
                <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Mở đầu bằng đoạn hay nhất</span>
                <button class="h-5 w-9 rounded-full relative transition-colors" :class="introOn ? 'bg-primary' : 'bg-muted-foreground/30'" @click="introOn = !introOn"><span class="absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all" :class="introOn ? 'left-[18px]' : 'left-0.5'"></span></button>
              </label>
              <template v-if="introOn">
                <div class="mt-2 flex items-center justify-between text-xs">
                  <span class="h-8 px-2.5 rounded-md border border-border flex items-center gap-1">Tiểu mục: Có mặt cho người thương <ChevronDown class="w-3 h-3" /></span>
                  <span class="text-muted-foreground tabular-nums">4/8 câu · 16 giây</span>
                </div>
                <div class="mt-2 max-h-[84px] overflow-auto rounded-lg border border-border divide-y divide-border">
                  <div v-for="(s, i) in introSentences" :key="i" class="flex items-start gap-2.5 px-3 py-1.5" :class="s.on && 'bg-primary/5'">
                    <span class="mt-0.5 h-4 w-4 rounded border grid place-items-center shrink-0" :class="s.on ? 'bg-primary border-primary text-primary-foreground' : 'border-border'"><Check v-if="s.on" class="w-3 h-3" /></span>
                    <span class="leading-snug text-[13px]" :class="s.on ? '' : 'text-muted-foreground'">{{ s.t }}</span>
                  </div>
                </div>
                <p class="mt-1 text-[11px] text-muted-foreground">Mặc định lấy đoạn đang nghe · người xem nghe thử trước khi vào sách.</p>
              </template>
              <p v-else class="mt-1 text-[11px] text-muted-foreground">Tắt: video phát từ đầu đến cuối như thường.</p>
            </div>

            <div class="flex items-start gap-5">
              <div class="flex-1">
                <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Khung</span>
                <div class="mt-2 grid grid-cols-2 gap-1.5">
                  <button v-for="r in (['wide', 'tall'] as const)" :key="r" class="h-11 rounded-lg border flex items-center gap-2 px-2.5 text-left" :class="ratio === r ? 'border-primary bg-primary/5' : 'border-border'" @click="ratio = r">
                    <span class="border-2 rounded-sm shrink-0" :class="[r === 'wide' ? 'w-6 h-4' : 'w-3.5 h-6', ratio === r ? 'border-primary' : 'border-muted-foreground/50']"></span>
                    <span class="leading-tight min-w-0"><span class="block text-[13px] font-medium">{{ r === 'wide' ? 'Ngang 16:9' : 'Dọc 9:16' }}</span><span class="block text-[10px] text-muted-foreground truncate">{{ r === 'wide' ? 'YouTube · 1920×1080' : 'Điện thoại · 1080×1920' }}</span></span>
                  </button>
                </div>
              </div>
              <div class="shrink-0">
                <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Nền · {{ bgs[bg].label }}</span>
                <div class="mt-2 flex gap-1.5">
                  <button v-for="(b, i) in bgs" :key="i" class="h-11 w-7 rounded-lg ring-offset-2 ring-offset-background" :class="[b.cls, bg === i ? 'ring-2 ring-primary' : 'ring-1 ring-border']" :title="b.label" @click="bg = i"></button>
                </div>
              </div>
            </div>

            <div>
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Kèm theo để đăng YouTube</span>
              <div class="mt-2 rounded-lg border border-border divide-y divide-border">
                <label v-for="x in ([{ k: 'thumb', i: ImageIcon, t: 'Ảnh thumbnail 1280×720', d: 'bìa + “Sách nói đầy đủ”' }, { k: 'srt', i: Subtitles, t: 'Phụ đề .srt từ lời đọc', d: 'dễ lên tìm kiếm YouTube' }, { k: 'desc', i: FileText, t: 'Mô tả có mốc chương', d: 'YouTube tự chia chương, kèm link' }] as const)" :key="x.k"
                  class="flex items-center gap-2.5 px-3 h-9 cursor-pointer" @click="extras[x.k] = !extras[x.k]">
                  <span class="h-4 w-4 rounded border grid place-items-center shrink-0" :class="extras[x.k] ? 'bg-primary border-primary text-primary-foreground' : 'border-border'"><Check v-if="extras[x.k]" class="w-3 h-3" /></span>
                  <component :is="x.i" class="w-4 h-4 shrink-0 text-muted-foreground" />
                  <span class="text-[13px] font-medium whitespace-nowrap">{{ x.t }}</span><span class="text-[11px] text-muted-foreground truncate">· {{ x.d }}</span>
                </label>
              </div>
            </div>
          </div>

          <!-- B. Đang tạo -->
          <div v-else-if="phase === 'busy'" class="flex-1 min-h-0 px-5 py-6 text-sm">
            <p class="font-medium">Đang tạo video {{ Math.round(selDur / 60) }} phút…</p>
            <p class="text-xs text-muted-foreground mt-0.5">Vẫn nghe sách, dùng app bình thường được. Đóng hộp này thì video vẫn tạo tiếp ở nền.</p>
            <div class="mt-4 h-2 rounded-full bg-muted overflow-hidden"><div class="h-full w-[48%] bg-primary rounded-full"></div></div>
            <p class="mt-1 text-xs text-muted-foreground tabular-nums">48% · còn khoảng 1 phút</p>
            <ol class="mt-5 space-y-2.5">
              <li v-for="s in steps" :key="s.t" class="flex items-center gap-2.5" :class="s.state === 'wait' && 'text-muted-foreground'">
                <span class="h-5 w-5 rounded-full grid place-items-center shrink-0" :class="s.state === 'done' ? 'bg-primary text-primary-foreground' : s.state === 'run' ? 'border-2 border-primary' : 'border border-border'">
                  <Check v-if="s.state === 'done'" class="w-3 h-3" /><Loader2 v-else-if="s.state === 'run'" class="w-3 h-3 animate-spin text-primary" />
                </span>{{ s.t }}
              </li>
            </ol>
          </div>

          <!-- C. Xong -->
          <div v-else class="flex-1 min-h-0 overflow-auto px-5 py-4 text-sm space-y-3">
            <p class="flex items-center gap-2 font-medium"><span class="h-5 w-5 rounded-full bg-primary text-primary-foreground grid place-items-center"><Check class="w-3 h-3" /></span> Đã tạo xong · {{ Math.round(selDur / 60) }} phút · {{ sizeMB }} MB</p>
            <div class="rounded-lg border border-border p-3 text-xs space-y-1 font-mono">
              <p class="font-sans text-[11px] font-semibold uppercase tracking-wider text-muted-foreground mb-1">Tải về / Sano video / {{ title }}</p>
              <p class="flex items-center gap-1.5"><Clapperboard class="w-3.5 h-3.5" /> {{ title }} - Sano.mp4</p>
              <p v-if="extras.thumb" class="flex items-center gap-1.5"><ImageIcon class="w-3.5 h-3.5" /> thumbnail.png</p>
              <p v-if="extras.srt" class="flex items-center gap-1.5"><Subtitles class="w-3.5 h-3.5" /> phu-de.srt</p>
              <p v-if="extras.desc" class="flex items-center gap-1.5"><FileText class="w-3.5 h-3.5" /> mo-ta-youtube.txt</p>
            </div>
            <div v-if="extras.desc">
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Mô tả YouTube</span>
                <button class="h-7 px-2.5 rounded-md border border-border text-xs flex items-center gap-1.5 hover:bg-muted" @click="copied = true"><component :is="copied ? Check : Copy" class="w-3.5 h-3.5" /> {{ copied ? 'Đã chép' : 'Chép mô tả' }}</button>
              </div>
              <pre class="mt-2 max-h-[190px] overflow-auto rounded-lg bg-muted/60 p-3 text-[11.5px] leading-relaxed whitespace-pre-wrap font-sans">{{ desc }}</pre>
            </div>
            <p class="text-[11px] text-muted-foreground">Đăng YouTube: tải video lên, chọn thumbnail.png làm hình thu nhỏ, dán mô tả, vào Phụ đề → Tải tệp lên → phu-de.srt.</p>
          </div>

          <div class="shrink-0 border-t border-border px-5 py-3.5">
            <Button v-if="phase === 'setup'" class="w-full" @click="phase = 'busy'"><Clapperboard class="w-4 h-4" /> Tạo video {{ Math.round(selDur / 60) }} phút</Button>
            <Button v-else-if="phase === 'busy'" variant="outline" class="w-full">Huỷ</Button>
            <div v-else class="flex gap-2">
              <Button class="flex-1"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
              <Button variant="outline" @click="phase = 'setup'">Tạo video khác</Button>
            </div>
            <p v-if="phase === 'setup'" class="mt-2 text-[11px] text-muted-foreground text-center">MP4 {{ ratio === 'wide' ? '1920×1080' : '1080×1920' }} · khoảng {{ sizeMB }} MB · tạo trên máy mất 1–3 phút, không cần mạng.</p>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D15 · Tạo video cả cuốn: menu bánh răng màn nghe → "Tạo video cả cuốn…". Trích đoạn (tuỳ chọn) → màn tựa → cả cuốn hoặc vài chương, chữ chạy theo câu → màn kết sanobook.com.
      Mặc định Ngang 16:9 1920×1080; kèm thumbnail, phụ đề .srt, mô tả có mốc chương.
    </p>
  </div>
</template>
