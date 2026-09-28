<script setup lang="ts">
// D18 — Làm nút chia sẻ hút mắt hơn để người nghe đăng lên mạng (Sano lan truyền).
// Hàng nút cũ (D17) ba nút viền xám ngang nhau: không có gì kéo mắt, không gợi "câu này đáng khoe".
// Ba phương án cho màn nghe Phóng to (mặc định):
// A. Nút màu nổi: Chia sẻ ảnh / Tạo video tô màu Hoàng hôn (đúng nền thẻ chia sẻ), có chữ mời;
//    Nghe trên điện thoại lùi thành nút nhẹ bên phải.
// B. Mời ngay dưới câu đang đọc: dưới câu đang sáng có nút nhỏ "Chia sẻ câu này" · "Tạo video
//    15 giây" (kiểu Spotify / NCT). Người nghe thấy câu hay là bấm luôn, mở hộp đã chọn sẵn câu đó.
// C. Thẻ xem trước: góc dưới có thẻ 9:16 nhỏ vẽ sẵn câu đang đọc (đổi theo câu), kèm lời mời
//    "Câu này hay? Đăng lên Facebook, TikTok". Nhìn thấy trước thành phẩm nên dễ muốn bấm.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API.
import { ref } from 'vue'
import { Clapperboard, Image as ImageIcon, Mic, Minimize2, Moon, Pause, RotateCcw, RotateCw, Settings, SkipBack, SkipForward, Smartphone, Sparkles, Sun, Timer, Gauge } from 'lucide-vue-next'
import WfBookCover from './WfBookCover.vue'

const q = new URLSearchParams(window.location.search)
type V = 'old' | 'a' | 'b' | 'c'
const v = ref<V>((q.get('v') as V) || 'a')
const dark = ref(false)
const title = 'Chánh niệm, nghệ thuật của sự có mặt'
const lines = [
  { t: 'Thiền sư Thích Nhất Hạnh có một ý rất đẹp.', s: 'past' },
  { t: 'Món quà quý nhất mình tặng cho người thương, không phải là quà cáp hay tiền bạc.', s: 'past' },
  { t: 'Mà là sự có mặt trọn vẹn của mình.', s: 'now' },
  { t: 'Ngồi cạnh con mà đầu nghĩ chuyện công ty, thì thật ra mình đang vắng mặt.', s: 'next' },
]
const opts: { k: V; t: string }[] = [
  { k: 'old', t: 'Hiện tại (D17)' },
  { k: 'a', t: 'A. Nút màu nổi' },
  { k: 'b', t: 'B. Mời dưới câu đang đọc' },
  { k: 'c', t: 'C. Thẻ xem trước' },
]
const sunset = 'bg-gradient-to-r from-orange-500 via-rose-600 to-purple-700'
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="o in opts" :key="o.k" class="h-8 px-3 rounded-full border text-xs" :class="v === o.k ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="v = o.k">{{ o.t }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark"><component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}</button>
    </div>

    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground shrink-0 w-[900px] h-[720px] flex flex-col p-6 isolate">
      <div class="absolute inset-0 -z-10 bg-gradient-to-b from-orange-100/60 via-background to-background dark:from-orange-950/30"></div>
      <div class="flex items-center justify-between text-sm text-muted-foreground"><span>‹ Thư viện</span><span class="flex items-center gap-2 text-foreground"><img src="@/assets/favicon.svg" alt="" class="h-6 w-6 rounded-md" /><b>Sano</b> <span class="text-[11px] tracking-[0.18em] text-muted-foreground">SÁCH NÓI</span></span><Settings class="w-4 h-4" /></div>

      <div class="mt-4 flex items-center gap-4">
        <div class="w-[84px] h-[112px] rounded-lg shadow-lg overflow-hidden shrink-0"><WfBookCover :title="title" size="sm" /></div>
        <div class="flex-1 min-w-0"><h2 class="text-lg font-semibold">{{ title }}</h2><p class="text-sm text-muted-foreground flex items-center gap-1"><Mic class="w-3.5 h-3.5" /> Giọng Thiền Tâm Đức</p><p class="text-sm text-primary font-medium">Có mặt cho người thương</p></div>
        <Minimize2 class="self-start w-4 h-4 text-muted-foreground/70" />
      </div>

      <!-- Lời đọc -->
      <div class="relative mt-5 flex-1 min-h-0 flex gap-6">
        <div class="flex-1 min-w-0 space-y-3">
          <template v-for="(l, i) in lines" :key="i">
            <p class="text-[26px] font-bold leading-snug" :class="l.s === 'now' ? 'text-foreground' : l.s === 'past' ? 'text-foreground/25' : 'text-foreground/40'">{{ l.t }}</p>
            <!-- B: mời ngay dưới câu đang đọc -->
            <div v-if="v === 'b' && l.s === 'now'" class="-mt-1 flex items-center gap-2">
              <button class="h-8 pl-2.5 pr-3 rounded-full text-white text-xs font-semibold flex items-center gap-1.5 shadow-md shadow-rose-500/30" :class="sunset"><Sparkles class="w-3.5 h-3.5" /> Chia sẻ câu này</button>
              <button class="h-8 px-3 rounded-full border border-rose-500/40 text-rose-600 dark:text-rose-400 bg-background/70 text-xs font-semibold flex items-center gap-1.5"><Clapperboard class="w-3.5 h-3.5" /> Video 15 giây</button>
              <span class="text-[11px] text-muted-foreground">hiện khi di chuột vào lời đọc hoặc tạm dừng</span>
            </div>
          </template>
        </div>
        <!-- C: thẻ xem trước -->
        <div v-if="v === 'c'" class="w-[190px] shrink-0 self-end flex flex-col items-center gap-2">
          <div class="relative w-[150px] h-[267px] rounded-xl overflow-hidden shadow-2xl text-white rotate-[3deg] ring-1 ring-black/5" :class="'bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800'">
            <div class="h-full flex flex-col p-3.5">
              <span class="font-serif text-3xl leading-none text-white/50">“</span>
              <p class="flex-1 flex items-center font-bold text-[14px] leading-snug">Mà là sự có mặt trọn vẹn của mình.</p>
              <p class="text-[8px] font-semibold leading-tight">{{ title }}</p>
              <div class="mt-1.5 pt-1.5 border-t border-white/25 flex items-center gap-1 text-[7px]"><img src="@/assets/favicon.svg" alt="" class="h-3 w-3 rounded" /><b>Sano</b> · Tự tạo sách nói<span class="ml-auto">sanobook.com</span></div>
            </div>
          </div>
          <p class="text-xs text-center font-medium">Câu này hay?<br /><span class="text-muted-foreground font-normal">Đăng lên Facebook, TikTok</span></p>
        </div>
      </div>

      <!-- Điều khiển -->
      <div class="shrink-0 mt-3 mx-auto w-full max-w-md">
        <div class="h-1.5 rounded-full bg-muted"><div class="h-full w-[45%] rounded-full bg-primary"></div></div>
        <div class="mt-1 flex justify-between text-[11px] text-muted-foreground"><span>0:16</span><span>-0:20</span></div>
        <div class="mt-2 flex items-center justify-center gap-4 text-muted-foreground">
          <span class="h-8 px-2.5 rounded-full border border-border flex items-center gap-1 text-xs"><Gauge class="w-3.5 h-3.5" /> 1x</span>
          <SkipBack class="w-5 h-5" /><RotateCcw class="w-5 h-5" />
          <span class="h-12 w-12 rounded-full bg-primary text-primary-foreground grid place-items-center"><Pause class="w-5 h-5" /></span>
          <RotateCw class="w-5 h-5" /><SkipForward class="w-5 h-5" />
          <span class="h-8 px-2.5 rounded-full border border-border flex items-center gap-1 text-xs"><Timer class="w-3.5 h-3.5" /> Nghỉ</span>
        </div>
      </div>

      <!-- Hàng nút -->
      <div class="shrink-0 mt-4 border-t border-border pt-4 flex items-center gap-2">
        <template v-if="v === 'old'">
          <span class="h-8 px-3 rounded-md border border-primary/30 bg-primary/10 text-primary text-sm flex items-center gap-1.5"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</span>
          <span class="h-8 px-3 rounded-md border border-border text-sm flex items-center gap-1.5"><ImageIcon class="w-4 h-4" /> Chia sẻ ảnh</span>
          <span class="h-8 px-3 rounded-md border border-border text-sm flex items-center gap-1.5"><Clapperboard class="w-4 h-4" /> Tạo video</span>
        </template>
        <template v-else-if="v === 'a' || v === 'c'">
          <button class="h-10 pl-3 pr-4 rounded-full text-white font-semibold text-sm flex items-center gap-2 shadow-lg shadow-rose-500/30 hover:brightness-110" :class="sunset">
            <ImageIcon class="w-4 h-4" /> Chia sẻ câu hay <span class="font-normal text-white/80 text-xs">· ảnh</span>
          </button>
          <button class="h-10 pl-3 pr-4 rounded-full text-white font-semibold text-sm flex items-center gap-2 shadow-lg shadow-indigo-500/30 hover:brightness-110 bg-gradient-to-r from-indigo-600 via-blue-700 to-fuchsia-700">
            <Clapperboard class="w-4 h-4" /> Tạo video <span class="font-normal text-white/80 text-xs">· Reels, YouTube</span>
          </button>
          <span class="ml-auto h-9 px-3 rounded-full text-sm text-muted-foreground hover:bg-muted flex items-center gap-1.5"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</span>
        </template>
        <template v-else>
          <span class="h-8 px-3 rounded-md border border-primary/30 bg-primary/10 text-primary text-sm flex items-center gap-1.5"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</span>
          <span class="h-8 px-3 rounded-md border border-border text-sm flex items-center gap-1.5"><ImageIcon class="w-4 h-4" /> Chia sẻ ảnh</span>
          <span class="h-8 px-3 rounded-md border border-border text-sm flex items-center gap-1.5"><Clapperboard class="w-4 h-4" /> Tạo video</span>
        </template>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[900px] text-center">
      D18 · A: hai nút tô màu nền thẻ chia sẻ, điện thoại lùi sang phải. B: mời chia sẻ ngay dưới câu đang đọc, bấm là mở hộp chọn sẵn câu đó.
      C: thẻ nhỏ vẽ sẵn câu đang đọc (đổi theo câu) để thấy trước thành phẩm. Có thể ghép A + B.
    </p>
  </div>
</template>
