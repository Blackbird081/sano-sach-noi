<script setup lang="ts">
// D17 — Đưa ảnh và video ra ngoài (sửa D14 + D15, làm trong 0.1.18 trước khi ra mắt).
// Vấn đề: nút Chia sẻ là icon nhỏ trong khung lời đọc (ít ai thấy), video ngắn lại nằm trong đó;
// "Tạo video cả cuốn…" giấu trong menu bánh răng.
// Đổi:
// - Hàng nút dưới màn nghe: [Nghe trên điện thoại] [Chia sẻ ảnh] [Tạo video]. Nút điện thoại giữ
//   màu nhấn, hai nút mới viền mảnh. Sách chưa có lời đọc thì hai nút mờ, di chuột hiện lý do.
// - "Chia sẻ ảnh": hộp D14 chỉ còn phần Ảnh có lời (bỏ lựa chọn Ảnh / Video).
// - "Tạo video": một hộp, đầu cột phải chọn Video ngắn (Reels, TikTok, Shorts · dọc 9:16 · 15–30 giây)
//   hoặc Video cả cuốn (YouTube, Facebook · ngang 16:9). Phần tuỳ chọn mỗi loại giữ nguyên như
//   D14 (video) và D15/D16. Nhớ loại dùng lần trước; lần đầu mặc định Video ngắn.
// - Menu bánh răng bỏ mục "Tạo video cả cuốn…". Nút Chia sẻ trong khung lời đọc (Phóng to, Xem cả
//   lời) giữ làm lối tắt: mở Chia sẻ ảnh, chọn sẵn câu đang đọc. Video ngắn mở từ nút Tạo video
//   cũng chọn sẵn câu đang đọc.
// - Chưa có menu Video riêng ở thanh bên: làm khi có danh sách video đã tạo / hàng đợi nhiều cuốn.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API.
import { computed, ref } from 'vue'
import {
  BarChart3, Check, ChevronLeft, Clapperboard, Copy, Download, FilePlus2, FolderOpen, Image as ImageIcon, Info, Library, Mic, Moon,
  Package, Pause, Pencil, Settings, Share2, SkipBack, SkipForward, Smartphone, Sparkles, Sun, Trash2, Film, MonitorPlay, X,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

const q = new URLSearchParams(window.location.search)
type View = 'listen' | 'image' | 'short' | 'full'
const view = ref<View>((q.get('view') as View) || 'listen')
const dark = ref(false)
const menu = ref(false)
const noLyrics = ref(false)
const hover = ref<'' | 'share' | 'video'>('')

const title = 'Chánh niệm, nghệ thuật của sự có mặt'
const voice = 'Thiền Tâm Đức'
const toc = ['Lời mở đầu', 'Chương 1. Có mặt, món quà quý nhất', 'Chương 2. Tập chánh niệm để được gì', 'Chương 3. Tập thế nào', 'Chương 4. Khi gặp vướng', 'Chương 5. Điều được khai mở bên trong', 'Chương 6. Chân dung người tập lâu năm']
const sentences = [
  'Thiền sư Thích Nhất Hạnh có một ý rất đẹp.',
  'Món quà quý nhất mình tặng cho người thương, không phải là quà cáp hay tiền bạc.',
  'Mà là sự có mặt trọn vẹn của mình.',
  'Ngồi cạnh con mà đầu nghĩ chuyện công ty, thì thật ra mình đang vắng mặt.',
  'Con cảm nhận được điều đó, dù con không nói ra.',
]
const picked = ref([1, 2])
const pickedShort = ref([1, 2, 3, 4])
const waveBars = Array.from({ length: 40 }, (_, i) => 18 + Math.round(80 * Math.abs(Math.sin(i * 0.9) * Math.cos(i * 0.31))))

const videoKind = computed({
  get: () => (view.value === 'full' ? 'full' : 'short'),
  set: (k: 'short' | 'full') => (view.value = k),
})
const views: { k: View; t: string }[] = [
  { k: 'listen', t: 'A. Màn nghe · hàng nút mới' },
  { k: 'image', t: 'B. Chia sẻ ảnh' },
  { k: 'short', t: 'C. Tạo video · Video ngắn' },
  { k: 'full', t: 'D. Tạo video · Video cả cuốn' },
]
const nav = [
  { t: 'Thư viện', i: Library, on: true },
  { t: 'Tạo sách nói', i: FilePlus2 },
  { t: 'Hành trình nghe', i: BarChart3 },
  { t: 'Cài đặt', i: Settings },
  { t: 'Giới thiệu', i: Info },
]
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="v in views" :key="v.k" class="h-8 px-3 rounded-full border text-xs"
        :class="view === v.k ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="view = v.k; menu = false">{{ v.t }}</button>
      <span class="w-px bg-border mx-1"></span>
      <button v-if="view === 'listen'" class="h-8 px-3 rounded-full border text-xs" :class="noLyrics ? 'border-primary bg-primary/10 text-primary' : 'border-border bg-background text-foreground'" @click="noLyrics = !noLyrics">Sách chưa có lời đọc</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground shrink-0 w-[1100px] h-[720px] flex">
      <!-- Thanh bên (giữ nguyên, chưa thêm menu Video) -->
      <aside class="w-52 shrink-0 border-r border-border bg-muted/30 p-3 flex flex-col gap-0.5 text-sm">
        <div class="flex items-center gap-2 px-2 pb-4 pt-1"><img src="@/assets/favicon.svg" alt="" class="h-6 w-6 rounded-md" /><b>Sano</b></div>
        <div v-for="n in nav" :key="n.t" class="h-8 px-2 rounded-md flex items-center gap-2" :class="n.on ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'"><component :is="n.i" class="w-4 h-4" /> {{ n.t }}</div>
      </aside>

      <!-- Màn nghe -->
      <main class="flex-1 min-w-0 flex flex-col p-6">
        <div class="shrink-0 flex items-center justify-between">
          <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Thư viện</span>
          <div class="relative">
            <button class="h-8 w-8 grid place-items-center rounded-md text-muted-foreground hover:bg-muted" :class="menu && 'bg-muted text-foreground'" @click="menu = !menu"><Settings class="w-4 h-4" /></button>
            <div v-if="menu" class="absolute top-full right-0 mt-1 w-64 rounded-lg border border-border bg-popover shadow-lg py-1 z-30 text-sm">
              <div class="px-3 py-1.5 flex items-center gap-2"><Pencil class="w-4 h-4" /> Sửa mục đang nghe</div>
              <div class="px-3 py-1.5 flex items-center gap-2"><Package class="w-4 h-4" /> Xuất gói zip</div>
              <div class="px-3 py-1.5 flex items-center gap-2"><Download class="w-4 h-4" /> Lưu file M4B vào chỗ khác…</div>
              <div class="px-3 py-1.5 flex items-center gap-2 text-muted-foreground/60 line-through"><Clapperboard class="w-4 h-4" /> Tạo video cả cuốn…</div>
              <div class="px-3 py-1.5 flex items-center gap-2"><FolderOpen class="w-4 h-4" /> Mở thư mục sách</div>
              <div class="my-1 border-t border-border"></div>
              <div class="px-3 py-1.5 flex items-center gap-2 text-destructive"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</div>
              <p class="mx-2 mt-1 mb-1 rounded bg-amber-500/10 text-amber-700 dark:text-amber-400 text-[11px] px-2 py-1">D17: bỏ mục Tạo video cả cuốn, chuyển ra nút Tạo video.</p>
            </div>
          </div>
        </div>

        <div class="flex-1 min-h-0 flex flex-col items-center justify-center">
          <div class="w-[180px] h-[240px] rounded-xl shadow-2xl overflow-hidden"><WfBookCover :title="title" size="lg" /></div>
          <h2 class="mt-4 text-lg font-semibold text-center max-w-md">{{ title }}</h2>
          <p class="text-sm text-muted-foreground flex items-center gap-1">Bùi Tấn Việt · <Mic class="w-3.5 h-3.5" /> Giọng {{ voice }}</p>
          <p class="text-sm text-primary font-medium">Có mặt cho người thương</p>
          <div class="mt-4 w-full max-w-md">
            <div class="h-1.5 rounded-full bg-muted"><div class="h-full w-[38%] rounded-full bg-primary"></div></div>
            <div class="mt-1 flex justify-between text-[11px] text-muted-foreground tabular-nums"><span>1:36</span><span>4:12</span></div>
          </div>
          <div class="mt-2 flex items-center gap-5">
            <SkipBack class="w-5 h-5 text-muted-foreground" />
            <span class="h-12 w-12 rounded-full bg-primary text-primary-foreground grid place-items-center shadow"><Pause class="w-5 h-5" /></span>
            <SkipForward class="w-5 h-5 text-muted-foreground" />
          </div>
          <p class="mt-5 max-w-md text-center text-[15px] leading-relaxed text-muted-foreground">… <b class="text-foreground">Mà là sự có mặt trọn vẹn của mình.</b> Ngồi cạnh con mà…</p>
        </div>

        <!-- MỚI: hàng nút -->
        <div class="shrink-0 relative flex flex-wrap items-center gap-2 border-t border-border pt-4">
          <Button size="sm" variant="outline" class="border-primary/30 bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</Button>
          <Button size="sm" variant="outline" :disabled="noLyrics" @mouseenter="hover = 'share'" @mouseleave="hover = ''" @click="view = 'image'"><ImageIcon class="w-4 h-4" /> Chia sẻ ảnh</Button>
          <Button size="sm" variant="outline" :disabled="noLyrics" @mouseenter="hover = 'video'" @mouseleave="hover = ''" @click="view = 'short'"><Clapperboard class="w-4 h-4" /> Tạo video</Button>
          <span v-if="view === 'listen'" class="ml-1 rounded-full bg-amber-500/10 text-amber-700 dark:text-amber-400 text-[11px] px-2 py-0.5">D17 · 2 nút mới</span>
          <div v-if="hover && view === 'listen'" class="absolute bottom-full mb-2 rounded-md bg-foreground text-background text-xs px-2.5 py-1.5 shadow-lg max-w-[320px]" :class="hover === 'share' ? 'left-[170px]' : 'left-[280px]'">
            <template v-if="noLyrics">Cuốn này chưa có lời đọc theo câu. Tạo lại sách bằng Sano bản mới để dùng.</template>
            <template v-else-if="hover === 'share'">Ảnh có lời từ đoạn hay, đăng Facebook, Zalo, Story</template>
            <template v-else>Video ngắn đăng Reels, TikTok · Video cả cuốn đăng YouTube</template>
          </div>
        </div>
      </main>

      <!-- Mục lục -->
      <div class="w-64 shrink-0 border-l border-border overflow-hidden">
        <div class="px-4 pt-3 pb-2 border-b border-border/60 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục</div>
        <div v-for="(t, i) in toc" :key="t" class="px-4 py-2 text-sm border-b border-border/40" :class="i === 1 ? 'text-primary font-medium bg-primary/5' : 'text-muted-foreground'">{{ t }}</div>
      </div>

      <!-- Hộp thoại -->
      <template v-if="view !== 'listen'">
        <div class="absolute inset-0 bg-black/45 backdrop-blur-[2px] z-10"></div>
        <div class="absolute z-20 inset-0 m-auto rounded-2xl bg-background border border-border shadow-2xl flex overflow-hidden" :class="view === 'image' ? 'w-[980px] h-[684px]' : 'w-[1060px] h-[690px]'">
          <!-- Trái: xem trước -->
          <div class="shrink-0 bg-muted/50 grid place-items-center p-6" :class="view === 'image' ? 'w-[460px]' : 'w-[540px]'">
            <!-- Thẻ dọc 9:16 (ảnh / video ngắn) -->
            <div v-if="view !== 'full'" class="relative w-[300px] h-[534px] rounded-2xl overflow-hidden shadow-2xl text-white isolate">
              <div class="absolute inset-0 -z-10 bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800"><div class="absolute inset-0 bg-gradient-to-b from-black/10 to-black/40"></div></div>
              <div class="h-full flex flex-col px-6 py-8">
                <template v-if="view === 'short'">
                  <div class="flex-1 flex flex-col items-center justify-center gap-4">
                    <div class="w-36 h-48 rounded-lg shadow-2xl overflow-hidden ring-1 ring-white/20"><WfBookCover :title="title" size="sm" /></div>
                    <div class="flex items-center gap-[3px] h-8"><span v-for="(h, i) in waveBars" :key="i" class="w-[3px] rounded-full" :class="i < 16 ? 'bg-white' : 'bg-white/35'" :style="{ height: h + '%' }"></span></div>
                  </div>
                  <p class="font-bold text-[21px] leading-snug min-h-[3.3em]">{{ sentences[1] }}</p>
                </template>
                <template v-else>
                  <span class="font-serif text-6xl leading-none text-white/50 -mb-2">“</span>
                  <div class="flex-1 flex flex-col justify-center gap-2.5"><p v-for="i in picked" :key="i" class="font-bold text-[22px] leading-snug">{{ sentences[i] }}</p></div>
                </template>
                <div class="mt-5 flex items-end gap-3">
                  <div v-if="view === 'image'" class="w-9 h-12 rounded shadow-lg overflow-hidden ring-1 ring-white/20 shrink-0"><WfBookCover :title="title" size="sm" /></div>
                  <div><p class="text-[13px] font-semibold leading-tight">{{ title }}</p><p class="mt-0.5 text-[11px] text-white/70 flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng {{ voice }}</p></div>
                </div>
                <div class="mt-3 pt-3 border-t border-white/20 flex items-center gap-2">
                  <img src="@/assets/favicon.svg" alt="" class="h-6 w-6 rounded-md shrink-0" />
                  <span class="text-[12px]"><b>Sano</b><span class="text-white/80"> · Tự tạo sách nói</span></span>
                  <span class="ml-auto text-[11px] font-semibold text-white/90">sanobook.com</span>
                </div>
              </div>
            </div>
            <!-- Khung ngang 16:9 (video cả cuốn, như D15) -->
            <div v-else class="flex flex-col items-center gap-3">
              <div class="relative w-[500px] h-[281px] rounded-xl overflow-hidden shadow-2xl text-white isolate">
                <div class="absolute inset-0 -z-10 bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800"><div class="absolute inset-0 bg-gradient-to-b from-black/15 to-black/45"></div></div>
                <div class="absolute top-3 right-3.5 flex items-center gap-1.5"><img src="@/assets/favicon.svg" alt="" class="h-5 w-5 rounded" /><span class="leading-tight"><span class="block text-[10px] font-bold">Sano · <span class="font-normal text-white/80">Tự tạo sách nói</span></span><span class="block text-[8.5px] font-semibold text-white/85">sanobook.com</span></span></div>
                <div class="absolute inset-0 flex items-center gap-6 px-10 pb-10">
                  <div class="w-[96px] h-[128px] rounded-lg shadow-2xl overflow-hidden ring-1 ring-white/20 shrink-0"><WfBookCover :title="title" size="md" /></div>
                  <div class="min-w-0">
                    <p class="text-[10px] text-white/70">Chương 1 · Có mặt cho người thương</p>
                    <p class="mt-1.5 font-bold text-[20px] leading-snug">{{ sentences[2] }}</p>
                    <p class="mt-1.5 text-[13px] text-white/55 leading-snug">{{ sentences[3] }}</p>
                  </div>
                </div>
                <div class="absolute inset-x-6 bottom-5 h-1 rounded-full bg-white/25"><div class="h-full w-[12%] rounded-full bg-white"></div></div>
              </div>
              <p class="text-[11px] text-muted-foreground">Xem trước, nghe thử như video, màn tựa, màn kết: giữ nguyên D15 / D16</p>
            </div>
          </div>

          <!-- Phải: tuỳ chọn -->
          <div class="flex-1 min-w-0 flex flex-col">
            <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
              <h2 class="font-semibold flex items-center gap-2">
                <template v-if="view === 'image'"><ImageIcon class="w-4 h-4" /> Chia sẻ ảnh</template>
                <template v-else><Clapperboard class="w-4 h-4" /> Tạo video</template>
                <span class="font-normal text-xs text-muted-foreground truncate">· {{ title }}</span>
              </h2>
              <X class="w-4 h-4 text-muted-foreground shrink-0" />
            </div>

            <div class="flex-1 min-h-0 overflow-auto px-5 py-4 space-y-4 text-sm">
              <!-- MỚI: chọn loại video -->
              <div v-if="view !== 'image'" class="grid grid-cols-2 gap-2">
                <button v-for="k in (['short', 'full'] as const)" :key="k" class="rounded-xl border-2 p-3 text-left flex gap-3 items-start"
                  :class="videoKind === k ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="videoKind = k">
                  <span class="mt-0.5 border-2 rounded-sm shrink-0" :class="[k === 'short' ? 'w-4 h-7' : 'w-7 h-4 mt-2', videoKind === k ? 'border-primary' : 'border-muted-foreground/50']"></span>
                  <span class="min-w-0">
                    <span class="font-semibold flex items-center gap-1.5"><component :is="k === 'short' ? Film : MonitorPlay" class="w-4 h-4" /> {{ k === 'short' ? 'Video ngắn' : 'Video cả cuốn' }}</span>
                    <span class="block text-[11px] text-muted-foreground mt-0.5">{{ k === 'short' ? 'Reels, TikTok, Shorts · dọc 9:16 · 15–30 giây' : 'YouTube, Facebook · ngang 16:9 · cả cuốn hoặc vài chương' }}</span>
                  </span>
                  <Check v-if="videoKind === k" class="w-4 h-4 text-primary ml-auto shrink-0" />
                </button>
              </div>

              <!-- Ảnh + video ngắn: chọn câu (như D14) -->
              <template v-if="view !== 'full'">
                <div>
                  <div class="flex items-baseline justify-between">
                    <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Chọn câu · Có mặt cho người thương</span>
                    <span class="text-xs text-muted-foreground tabular-nums">{{ view === 'image' ? `${picked.length}/4 câu` : `${pickedShort.length}/8 câu · 15 giây` }}</span>
                  </div>
                  <div class="mt-2 rounded-lg border border-border divide-y divide-border">
                    <div v-for="(s, i) in sentences" :key="i" class="flex items-start gap-2.5 px-3 py-2" :class="(view === 'image' ? picked : pickedShort).includes(i) ? 'bg-primary/5' : ''">
                      <span class="mt-0.5 h-4 w-4 rounded border grid place-items-center shrink-0" :class="(view === 'image' ? picked : pickedShort).includes(i) ? 'bg-primary border-primary text-primary-foreground' : 'border-border'"><Check v-if="(view === 'image' ? picked : pickedShort).includes(i)" class="w-3 h-3" /></span>
                      <span class="leading-snug" :class="(view === 'image' ? picked : pickedShort).includes(i) ? '' : 'text-muted-foreground'">{{ s }}</span>
                    </div>
                  </div>
                  <p class="mt-1.5 text-[11px] text-muted-foreground">Chọn sẵn câu đang đọc. Nghe thử từng câu, khung, nền: giữ nguyên D14.</p>
                </div>
                <div v-if="view === 'image'">
                  <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Khung · Nền</span>
                  <p class="mt-1 text-xs text-muted-foreground">Dọc 9:16 / Vuông 1:1 · Hoàng hôn, Theo bìa, Biển, Đêm, Trà (như D14)</p>
                </div>
                <div v-else>
                  <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Nền</span>
                  <p class="mt-1 text-xs text-muted-foreground">Hoàng hôn, Theo bìa, Biển, Đêm, Trà (như D14). Khung cố định dọc 9:16.</p>
                </div>
              </template>

              <!-- Video cả cuốn: tóm tắt các khối D15/D16 -->
              <template v-else>
                <div class="rounded-lg border border-border divide-y divide-border">
                  <div class="px-3 py-2.5 flex items-center justify-between"><span>Phần sách</span><span class="text-muted-foreground">Cả cuốn · 27 phút</span></div>
                  <div class="px-3 py-2.5 flex items-center justify-between"><span class="flex items-center gap-1.5"><Sparkles class="w-4 h-4" /> Mở đầu bằng đoạn hay nhất</span><span class="text-muted-foreground">Bật · 4 câu, 17 giây</span></div>
                  <div class="px-3 py-2.5 flex items-center justify-between"><span class="flex items-center gap-1.5"><Mic class="w-4 h-4" /> Lời giới thiệu, nhạc hiệu, chuông</span><span class="text-muted-foreground">Như D16</span></div>
                  <div class="px-3 py-2.5 flex items-center justify-between"><span>Khung · Nền</span><span class="text-muted-foreground">Ngang 16:9 · Hoàng hôn</span></div>
                  <div class="px-3 py-2.5 flex items-center justify-between"><span>Kèm theo</span><span class="text-muted-foreground">Thumbnail · phụ đề .srt · mô tả YouTube</span></div>
                </div>
                <p class="text-[11px] text-muted-foreground">Các khối giữ nguyên nội dung D15 / D16, chỉ đổi chỗ vào: từ menu bánh răng sang nút Tạo video.</p>
              </template>
            </div>

            <div class="shrink-0 border-t border-border px-5 py-3.5">
              <div v-if="view === 'image'" class="flex items-center gap-2">
                <Button class="flex-1"><Download class="w-4 h-4" /> Lưu ảnh</Button>
                <Button variant="outline"><Copy class="w-4 h-4" /> Sao chép ảnh</Button>
                <Button variant="outline"><Share2 class="w-4 h-4" /> AirDrop</Button>
              </div>
              <Button v-else-if="view === 'short'" class="w-full"><Film class="w-4 h-4" /> Tạo video 15 giây</Button>
              <Button v-else class="w-full"><Clapperboard class="w-4 h-4" /> Tạo video cả cuốn · khoảng 2 phút</Button>
              <p class="mt-2 text-[11px] text-muted-foreground text-center">
                {{ view === 'image' ? 'Ảnh PNG 1080 px. Facebook trên máy tính: bấm Sao chép ảnh rồi dán vào ô đăng bài.' : view === 'short' ? 'Video MP4 1080×1920 kèm tiếng đọc đúng các câu đã chọn. Tạo ngay trên máy.' : 'Video MP4 1920×1080, lưu vào thư mục riêng kèm thumbnail, phụ đề, mô tả.' }}
              </p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D17 · Hàng nút dưới màn nghe: Nghe trên điện thoại · Chia sẻ ảnh · Tạo video. Hộp Chia sẻ chỉ còn ảnh; hộp Tạo video gộp Video ngắn và Video cả cuốn.
      Menu bánh răng bỏ "Tạo video cả cuốn…". Nút Chia sẻ trong khung lời đọc giữ làm lối tắt tới Chia sẻ ảnh.
    </p>
  </div>
</template>
