<script setup lang="ts">
// D16 — Mở đầu video kiểu Fonos (bổ sung D15): không còn giây nào im lặng.
// Dòng thời gian: [Trích đoạn — tuỳ chọn] → giọng đọc "Bạn đang nghe sách nói, tạo bằng Sano."
// → nhạc hiệu ~4 giây (lên nhẹ, tắt dần) dưới màn tựa → giọng đọc giới thiệu sách (tên, tác giả,
// dịch giả, nhà xuất bản — trường nào trống thì bỏ) → nội dung; sang chương: thẻ chương 1,5 giây
// + tiếng chuông ngắn rồi đọc tên chương như cũ → màn kết: giọng đọc "Tạo sách nói của bạn tại
// sanobook.com" + nhạc hiệu tắt dần.
// - Câu đọc thêm: dùng chính bộ đọc VieNeu + đúng giọng của cuốn, đọc lúc tạo video (vài giây).
// - Tiểu mục mở đầu tự có của sách ("Bạn đang nghe sách nói. Cuốn sách: …") được thay bằng lời
//   giới thiệu mới khi bật, để không đọc trùng.
// - Nhạc hiệu: bản tổng hợp tạm của Sano (tự tạo, không vướng bản quyền) hoặc file nhạc của người
//   dùng; khi có nhạc hiệu Sano chính thức thì thay.
// - Sửa sách → Thông tin & bìa: thêm Dịch giả, Nhà xuất bản (không bắt buộc).
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API.
import { computed, ref } from 'vue'
import { Bell, BookOpen, Check, ChevronDown, Clapperboard, Mic, Music, Pencil, Play, Sparkles, Sun, Moon, Upload, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

const q = new URLSearchParams(window.location.search)
const view = ref<'video' | 'info'>((q.get('view') as never) || 'video')
const dark = ref(false)
const voiceIntro = ref(true)
const music = ref<'sano' | 'file' | 'none'>('sano')
const chime = ref(true)
const introOn = ref(true)
const title = 'Chánh niệm, nghệ thuật của sự có mặt'
const info = ref({ author: 'Bùi Tấn Việt', translator: '', publisher: 'Học Trò Thầy Thành' })
const introLine = computed(() =>
  [title + '.', info.value.author && `Tác giả ${info.value.author}.`, info.value.translator && `Dịch giả ${info.value.translator}.`, info.value.publisher && `Nhà xuất bản ${info.value.publisher}.`].filter(Boolean).join(' '),
)

// Dải dòng thời gian ~70 giây đầu (độ rộng theo số giây)
const strip = computed(() =>
  [
    introOn.value && { k: 'excerpt', t: 'Trích đoạn', s: 17, icon: Sparkles, cls: 'bg-amber-500/80' },
    voiceIntro.value && { k: 'brand', t: '“Bạn đang nghe sách nói, tạo bằng Sano”', s: 3, icon: Mic, cls: 'bg-primary' },
    music.value !== 'none' && { k: 'music', t: 'Nhạc hiệu', s: 4, icon: Music, cls: 'bg-violet-500' },
    !voiceIntro.value && music.value === 'none' && { k: 'silent', t: 'Màn tựa (lặng)', s: 3, icon: BookOpen, cls: 'bg-muted-foreground/40' },
    voiceIntro.value && { k: 'info', t: 'Giới thiệu sách', s: 6, icon: Mic, cls: 'bg-primary' },
    { k: 'ch0', t: 'Lời mở đầu', s: 76, icon: BookOpen, cls: 'bg-sky-600' },
    { k: 'ch1', t: 'Chương 1', s: 12, icon: chime.value ? Bell : BookOpen, cls: 'bg-sky-700' },
  ].filter(Boolean) as { k: string; t: string; s: number; icon: unknown; cls: string }[],
)
const stripTotal = computed(() => strip.value.reduce((n, x) => n + x.s, 0))
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button class="h-8 px-3 rounded-full border text-xs" :class="view === 'video' ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="view = 'video'">A. Tạo video cả cuốn — Mở đầu & chuyển cảnh</button>
      <button class="h-8 px-3 rounded-full border text-xs" :class="view === 'info' ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="view = 'info'">B. Sửa sách — Thông tin & bìa</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground shrink-0 w-[1100px] h-[720px]">
      <!-- A. Hộp Tạo video cả cuốn -->
      <template v-if="view === 'video'">
        <div class="absolute inset-0 bg-black/45 z-10"></div>
        <div class="absolute z-20 inset-0 m-auto w-[1060px] h-[690px] rounded-2xl bg-background border border-border shadow-2xl flex overflow-hidden">
          <div class="w-[540px] shrink-0 bg-muted/50 flex flex-col items-center justify-center gap-3 p-5">
            <!-- Màn tựa lúc nhạc hiệu: bìa + tên, nốt nhạc nhẹ -->
            <div class="relative w-[500px] h-[281px] rounded-xl overflow-hidden shadow-2xl text-white isolate">
              <div class="absolute inset-0 -z-10 bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800"><div class="absolute inset-0 bg-gradient-to-b from-black/15 to-black/45"></div></div>
              <div class="absolute top-3 right-3.5 flex items-center gap-1.5"><img src="@/assets/favicon.svg" alt="" class="h-5 w-5 rounded" /><span class="leading-tight"><span class="block text-[10px] font-bold">Sano · <span class="font-normal text-white/80">Tự tạo sách nói</span></span><span class="block text-[8.5px] font-semibold text-white/85">sanobook.com</span></span></div>
              <div class="absolute inset-0 flex items-center gap-6 px-12 pb-6">
                <div class="w-[120px] h-[160px] rounded-lg shadow-2xl overflow-hidden ring-1 ring-white/20 shrink-0"><WfBookCover :title="title" size="md" /></div>
                <div>
                  <span class="text-[9px] font-bold uppercase tracking-[0.2em] text-white/70">Sách nói đầy đủ</span>
                  <h3 class="mt-1 font-serif font-bold text-[24px] leading-tight">{{ title }}</h3>
                  <p class="mt-2 text-[11px] text-white/80">{{ info.author && `Tác giả ${info.author}` }}{{ info.translator && ` · Dịch giả ${info.translator}` }}</p>
                  <p class="text-[11px] text-white/70">{{ info.publisher && `NXB ${info.publisher}` }}</p>
                </div>
              </div>
              <span v-if="music !== 'none'" class="absolute left-4 bottom-4 flex items-center gap-1.5 rounded-full bg-white/15 px-2.5 py-1 text-[10px]"><Music class="w-3 h-3" /> nhạc hiệu lên nhẹ… tắt dần</span>
            </div>
            <div class="w-full max-w-[500px] flex items-center gap-2.5">
              <span class="h-9 w-9 shrink-0 rounded-full bg-primary text-primary-foreground grid place-items-center shadow"><Play class="w-4 h-4 ml-0.5" /></span>
              <div class="flex-1"><div class="h-1.5 rounded-full bg-muted-foreground/20"><div class="h-full w-[3%] rounded-full bg-primary"></div></div>
                <div class="mt-1 flex justify-between text-[11px] text-muted-foreground tabular-nums"><span>Nhạc hiệu</span><span>0:19 / 28:58</span></div></div>
            </div>

            <!-- Dải dòng thời gian mở đầu -->
            <div class="w-full max-w-[500px]">
              <p class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground mb-1.5">Mở đầu video (~70 giây đầu)</p>
              <div class="flex h-9 rounded-md overflow-hidden gap-px bg-border">
                <div v-for="x in strip" :key="x.k" class="relative text-white text-[10px] px-1.5 flex items-center gap-1 min-w-0" :class="x.cls" :style="{ flex: Math.max(x.s, 6) + ' 0 36px' }" :title="`${x.t} · ${x.s} giây`">
                  <component :is="x.icon" class="w-3 h-3 shrink-0" /><span class="truncate">{{ x.t }}</span>
                </div>
              </div>
              <div class="mt-1 flex justify-between text-[10px] text-muted-foreground tabular-nums"><span>0:00</span><span>{{ Math.floor(stripTotal / 60) }}:{{ String(stripTotal % 60).padStart(2, '0') }}</span></div>
              <ol class="mt-1.5 flex flex-wrap items-center gap-x-1 gap-y-0.5 text-[11px]">
                <li v-for="(x, i) in strip" :key="x.k" class="flex items-center gap-1"><span v-if="i" class="text-muted-foreground">→</span><span class="h-2 w-2 rounded-sm" :class="x.cls"></span>{{ x.t }} <span class="text-muted-foreground tabular-nums">{{ x.s }}s</span></li>
              </ol>
              <p class="mt-1 text-[11px] text-muted-foreground">Không còn đoạn im lặng: chỗ chuyển cảnh nào cũng có giọng đọc, nhạc hiệu hoặc tiếng chuông.</p>
            </div>
          </div>

          <div class="flex-1 min-w-0 flex flex-col">
            <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
              <div><h2 class="font-semibold flex items-center gap-2"><Clapperboard class="w-4 h-4" /> Tạo video cả cuốn</h2><p class="text-xs text-muted-foreground mt-0.5">Đăng YouTube, Facebook · {{ title }}</p></div>
              <X class="w-4 h-4 text-muted-foreground" />
            </div>
            <div class="flex-1 min-h-0 overflow-auto px-5 py-3 space-y-3 text-sm">
              <div class="rounded-lg border border-border px-3 py-2 flex items-center justify-between text-xs text-muted-foreground">
                <span>Phần sách: <b class="text-foreground font-medium">Cả cuốn</b> · Mở đầu bằng đoạn hay nhất: <b class="text-foreground font-medium">{{ introOn ? 'Bật · 4 câu, 17 giây' : 'Tắt' }}</b></span>
                <button class="text-primary hover:underline" @click="introOn = !introOn">(như D15)</button>
              </div>

              <!-- MỚI: Mở đầu & chuyển cảnh -->
              <div class="rounded-xl border-2 border-primary/30 bg-primary/[0.03] p-3.5 space-y-3">
                <p class="text-[11px] font-semibold uppercase tracking-wider text-primary">Mới · Mở đầu & chuyển cảnh</p>

                <div>
                  <div class="flex items-center justify-between">
                    <span class="font-medium flex items-center gap-1.5"><Mic class="w-4 h-4" /> Lời giới thiệu có giọng đọc</span>
                    <button class="h-5 w-9 rounded-full relative" :class="voiceIntro ? 'bg-primary' : 'bg-muted-foreground/30'" @click="voiceIntro = !voiceIntro"><span class="absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all" :class="voiceIntro ? 'left-[18px]' : 'left-0.5'"></span></button>
                  </div>
                  <div v-if="voiceIntro" class="mt-2 rounded-lg bg-background border border-border divide-y divide-border text-[13px]">
                    <div class="px-3 py-1.5 flex items-start gap-2"><span class="text-[10px] font-semibold text-muted-foreground w-14 shrink-0 pt-0.5">MỞ ĐẦU</span><span class="flex-1">“Bạn đang nghe sách nói, tạo bằng Sano.”</span><Play class="w-3.5 h-3.5 text-muted-foreground mt-0.5" /></div>
                    <div class="px-3 py-1.5 flex items-start gap-2"><span class="text-[10px] font-semibold text-muted-foreground w-14 shrink-0 pt-0.5">SAU NHẠC</span><span class="flex-1">“{{ introLine }}”</span><Play class="w-3.5 h-3.5 text-muted-foreground mt-0.5" /></div>
                    <div class="px-3 py-1.5 flex items-start gap-2"><span class="text-[10px] font-semibold text-muted-foreground w-14 shrink-0 pt-0.5">MÀN KẾT</span><span class="flex-1">“Tạo sách nói của bạn tại sanobook.com.”</span><Play class="w-3.5 h-3.5 text-muted-foreground mt-0.5" /></div>
                  </div>
                  <p v-if="voiceIntro" class="mt-1 text-[11px] text-muted-foreground flex items-center gap-1">Đọc bằng giọng Thiền Tâm Đức lúc tạo video (vài giây). Thiếu tác giả, NXB? <button class="text-primary hover:underline inline-flex items-center gap-0.5" @click="view = 'info'"><Pencil class="w-3 h-3" /> Sửa thông tin sách</button></p>
                </div>

                <div>
                  <div class="flex items-center justify-between">
                    <span class="font-medium flex items-center gap-1.5"><Music class="w-4 h-4" /> Nhạc hiệu</span>
                    <span class="flex items-center gap-2">
                      <span class="h-8 px-2.5 rounded-md border border-border bg-background text-xs flex items-center gap-1.5">{{ { sano: 'Nhạc hiệu Sano (bản tạm)', file: 'Nhac-hieu-cua-toi.mp3', none: 'Không nhạc' }[music] }}<ChevronDown class="w-3 h-3" /></span>
                      <span class="h-8 w-8 rounded-md border border-border grid place-items-center" title="Nghe thử nhạc hiệu"><Play class="w-3.5 h-3.5" /></span>
                    </span>
                  </div>
                  <div class="mt-1.5 flex gap-1.5 text-[11px]">
                    <button v-for="m in (['sano', 'file', 'none'] as const)" :key="m" class="h-6 px-2 rounded-full border" :class="music === m ? 'border-primary bg-primary/10 text-primary' : 'border-border text-muted-foreground'" @click="music = m">
                      <Upload v-if="m === 'file'" class="inline w-3 h-3 -mt-0.5" /> {{ { sano: 'Sano (tạm)', file: 'Chọn file nhạc của tôi…', none: 'Không nhạc' }[m] }}
                    </button>
                  </div>
                  <p class="mt-1 text-[11px] text-muted-foreground">Lên nhẹ ~4 giây dưới màn tựa rồi tắt dần; màn kết cũng có. Nhạc tự chọn: anh chị tự chịu trách nhiệm bản quyền.</p>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-medium flex items-center gap-1.5"><Bell class="w-4 h-4" /> Tiếng chuông khi sang chương</span>
                  <button class="h-5 w-9 rounded-full relative" :class="chime ? 'bg-primary' : 'bg-muted-foreground/30'" @click="chime = !chime"><span class="absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all" :class="chime ? 'left-[18px]' : 'left-0.5'"></span></button>
                </div>
              </div>

              <div class="rounded-lg border border-border px-3 py-2 text-xs text-muted-foreground">Khung · Nền · Kèm theo để đăng YouTube — như D15</div>
            </div>
            <div class="shrink-0 border-t border-border px-5 py-3.5">
              <Button class="w-full"><Clapperboard class="w-4 h-4" /> Tạo video 29 phút</Button>
              <p class="mt-2 text-[11px] text-muted-foreground text-center">MP4 1920×1080 · khoảng 70 MB · tạo trên máy, không cần mạng.</p>
            </div>
          </div>
        </div>
      </template>

      <!-- B. Sửa sách → Thông tin & bìa -->
      <template v-else>
        <div class="h-full flex flex-col">
          <div class="shrink-0 px-6 pt-5 pb-3 border-b border-border">
            <h1 class="text-lg font-semibold">Sửa sách · {{ title }}</h1>
            <div class="mt-3 flex gap-1 text-sm">
              <span v-for="t in ['Nội dung', 'Thông tin & bìa', 'Giọng đọc', 'Từ điển']" :key="t" class="h-8 px-3 rounded-md grid place-items-center" :class="t === 'Thông tin & bìa' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">{{ t }}</span>
            </div>
          </div>
          <div class="flex-1 flex p-6 gap-8">
            <div class="w-[380px] shrink-0 space-y-3 text-sm">
              <label class="block"><span class="text-xs font-medium text-muted-foreground">Tên sách</span><input :value="title" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" /></label>
              <label class="block"><span class="text-xs font-medium text-muted-foreground">Tác giả</span><input v-model="info.author" placeholder="Không bắt buộc" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" /></label>
              <!-- MỚI -->
              <div class="grid grid-cols-2 gap-2 rounded-lg ring-2 ring-primary/30 p-2 -m-2">
                <label class="block"><span class="text-xs font-medium text-muted-foreground">Dịch giả <span class="text-primary">· mới</span></span><input v-model="info.translator" placeholder="Không bắt buộc" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" /></label>
                <label class="block"><span class="text-xs font-medium text-muted-foreground">Nhà xuất bản <span class="text-primary">· mới</span></span><input v-model="info.publisher" placeholder="Không bắt buộc" class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3" /></label>
              </div>
              <p class="text-[11px] text-muted-foreground pt-1">Dùng cho lời giới thiệu có giọng đọc, màn tựa và mô tả YouTube khi tạo video.</p>
              <div><span class="text-xs font-medium text-muted-foreground">Danh mục</span><div class="mt-1 h-9 rounded-md border border-input px-3 flex items-center text-muted-foreground">Kỹ năng</div></div>
              <div class="flex gap-2 pt-1"><Button size="sm"><Check class="w-4 h-4" /> Lưu</Button><Button size="sm" variant="ghost">Huỷ</Button></div>
            </div>
            <div><div class="text-xs font-medium text-muted-foreground mb-2">Bìa</div><div class="w-40 aspect-[3/4] rounded-lg overflow-hidden shadow-lg"><WfBookCover :title="title" size="md" /></div></div>
          </div>
        </div>
      </template>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D16 · Mở đầu kiểu Fonos: giọng đọc thương hiệu → nhạc hiệu 4 giây → giọng đọc giới thiệu sách (tên, tác giả, dịch giả, NXB) → nội dung; chuông khi sang chương; màn kết có giọng đọc + nhạc.
      Câu đọc thêm dùng bộ đọc và giọng của cuốn. Sửa sách thêm Dịch giả, Nhà xuất bản.
    </p>
  </div>
</template>
