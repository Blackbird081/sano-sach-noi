<script setup lang="ts">
// D14 — Chia sẻ đoạn hay trong sách nói (ý tưởng từ NCT: "Thẻ lời bài hát" + "Chia sẻ video").
// Vào: nút Chia sẻ ở chế độ Phóng to lời đọc (D13) và ở Xem cả lời; mặc định chọn câu đang đọc.
// Hộp thoại 2 cột: trái là bản xem trước đúng như file sẽ lưu; phải là tuỳ chọn.
// - Ảnh: chọn 1–4 câu liền nhau, khung Vuông 1:1 (bài Facebook) / Dọc 9:16 (Story, Reels),
//   nền Theo bìa (bìa làm mờ) hoặc vài dải màu. Lưu ảnh PNG / Sao chép ảnh / AirDrop (Mac).
// - Video: cùng thẻ, kèm tiếng đọc đúng các câu đã chọn (tối đa ~30 giây), chữ đổi theo câu,
//   thanh sóng âm chạy. "Tạo video" → tiến độ → Lưu video MP4 / AirDrop.
// - Thẻ luôn có: tên sách, giọng đọc, bìa nhỏ; dải thương hiệu "Sano · Tự tạo sách nói · sanobook.com".
// Vòng 2 (anh Việt): nền Hoàng hôn + khung Dọc 9:16 mặc định, hộp thoại to hơn để thấy hết tuỳ chọn,
// dải thương hiệu đầy đủ để người xem biết Sano làm gì và tìm tới (marketing lan truyền).
// Làm trên máy: ảnh vẽ bằng canvas; video ghép ảnh từng câu + đoạn mp3 bằng ffmpeg (có sẵn cho M4B).
// Bản làm thật: sóng âm là cột thật tính từ tiếng đọc, tô sáng dần từ trái sang theo thời gian
// (thay thanh tiến độ dưới câu); chữ dành chỗ theo câu dài nhất nên bìa không nhảy giữa các câu.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API.
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { Check, Copy, Download, Film, Image as ImageIcon, Loader2, Mic, Play, Share2, Sun, Moon, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

const q = new URLSearchParams(window.location.search)
const kind = ref<'image' | 'video'>((q.get('kind') as never) || 'image')
const ratio = ref<'square' | 'story'>((q.get('ratio') as never) || 'story') // vòng 2: Dọc 9:16 mặc định
const bg = ref(0)
const dark = ref(false)
type VState = 'idle' | 'busy' | 'done'
const vstate = ref<VState>((q.get('v') as VState) || 'idle')

const title = 'Chánh niệm, nghệ thuật của sự có mặt'
const voice = 'Thiền Tâm Đức'
const section = 'Có mặt cho người thương'
const sentences = [
  { t: 'Thiền sư Thích Nhất Hạnh có một ý rất đẹp.', d: 2.6 },
  { t: 'Món quà quý nhất mình tặng cho người thương, không phải là quà cáp hay tiền bạc.', d: 5.1 },
  { t: 'Mà là sự có mặt trọn vẹn của mình.', d: 2.4 },
  { t: 'Ngồi cạnh con mà đầu nghĩ chuyện công ty, thì thật ra mình đang vắng mặt.', d: 4.8 },
  { t: 'Con cảm nhận được điều đó, dù con không nói ra.', d: 3.0 },
  { t: 'Hãy tắt điện thoại, nhìn vào mắt người đối diện, và nghe.', d: 3.6 },
]
const picked = ref<number[]>([1, 2])
// Ảnh tối đa 4 câu; video tối đa 8 câu / 30 giây, mở ra chọn sẵn ~15 giây (mỗi câu ~3,3 giây).
const MAX = computed(() => (kind.value === 'video' ? 8 : 4))
function toggle(i: number) {
  const p = picked.value
  if (p.includes(i)) {
    // chỉ bỏ được câu ở hai đầu để đoạn luôn liền nhau
    if (p.length > 1 && (i === Math.min(...p) || i === Math.max(...p))) picked.value = p.filter((x) => x !== i)
    return
  }
  const lo = Math.min(...p)
  const hi = Math.max(...p)
  if ((i === lo - 1 || i === hi + 1) && p.length < MAX.value) picked.value = [...p, i].sort((a, b) => a - b)
  else picked.value = [i] // bấm câu không liền kề → bắt đầu đoạn mới
}
const canAdd = (i: number) => {
  const p = picked.value
  return p.length < MAX.value && (i === Math.min(...p) - 1 || i === Math.max(...p) + 1)
}
const dur = computed(() => picked.value.reduce((n, i) => n + sentences[i].d, 0))

// Nền (vòng 2): Hoàng hôn đứng đầu, mặc định; "Theo bìa" = bìa làm mờ, phủ tối.
const bgs: { label: string; cls: string; cover?: boolean }[] = [
  { label: 'Hoàng hôn', cls: 'bg-gradient-to-br from-orange-500 via-rose-600 to-purple-800' },
  { label: 'Theo bìa', cls: '', cover: true },
  { label: 'Biển', cls: 'bg-gradient-to-br from-cyan-500 via-teal-500 to-lime-500' },
  { label: 'Đêm', cls: 'bg-gradient-to-br from-indigo-700 via-blue-800 to-fuchsia-700' },
  { label: 'Trà', cls: 'bg-gradient-to-br from-stone-600 via-amber-800 to-stone-900' },
]

// Xem trước video: chữ đổi theo từng câu đã chọn, thanh sóng âm chạy.
const vIdx = ref(0)
const playing = ref(kind.value === 'video')
let timer: number | undefined
watch([playing, picked], () => {
  clearInterval(timer)
  vIdx.value = 0
  if (playing.value) timer = window.setInterval(() => (vIdx.value = (vIdx.value + 1) % picked.value.length), 2200)
}, { immediate: true })
watch(kind, (k) => (playing.value = k === 'video'))
onBeforeUnmount(() => clearInterval(timer))
const cardLines = computed(() => (kind.value === 'video' ? [sentences[picked.value[vIdx.value]].t] : picked.value.map((i) => sentences[i].t)))
const vPct = computed(() => Math.round(((vIdx.value + 1) / picked.value.length) * 100))

let busyTimer: number | undefined
const progress = ref(60)
function makeVideo() {
  vstate.value = 'busy'
  progress.value = 0
  clearInterval(busyTimer)
  busyTimer = window.setInterval(() => {
    progress.value += 12
    if (progress.value >= 100) {
      clearInterval(busyTimer)
      vstate.value = 'done'
    }
  }, 300)
}
watch([kind, picked, ratio, bg], () => vstate.value !== 'busy' && (vstate.value = 'idle'))
const toast = ref('')
function flash(t: string) {
  toast.value = t
  setTimeout(() => (toast.value = ''), 2500)
}
const waveBars = Array.from({ length: 40 }, (_, i) => 18 + Math.round(80 * Math.abs(Math.sin(i * 0.9) * Math.cos(i * 0.31))))
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="k in (['image', 'video'] as const)" :key="k" class="h-8 px-3 rounded-full border text-xs"
        :class="kind === k ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="kind = k">{{ k === 'image' ? '1. Ảnh có lời' : '2. Video có tiếng' }}</button>
      <button class="h-8 px-3 rounded-full border text-xs" :class="vstate === 'busy' ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="kind = 'video'; vstate = 'busy'; progress = 60">3. Đang tạo video</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <!-- Nền mờ: màn nghe phía sau -->
    <div class="relative rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground shrink-0 w-[1100px] h-[720px]">
      <div class="absolute inset-0 bg-black/45 backdrop-blur-[2px] z-10"></div>
      <div class="absolute inset-0 p-10 text-2xl font-bold text-foreground/30 leading-relaxed">Món quà quý nhất mình tặng cho người thương, không phải là quà cáp hay tiền bạc. Mà là sự có mặt trọn vẹn của mình.</div>

      <!-- Hộp thoại Chia sẻ -->
      <div class="absolute z-20 inset-x-0 top-1/2 -translate-y-1/2 mx-auto w-[980px] h-[684px] rounded-2xl bg-background border border-border shadow-2xl flex overflow-hidden">
        <!-- Trái: xem trước -->
        <div class="w-[460px] shrink-0 bg-muted/50 grid place-items-center p-6 relative">
          <div class="relative overflow-hidden rounded-2xl shadow-2xl text-white isolate" :class="ratio === 'square' ? 'w-[380px] h-[380px]' : 'w-[300px] h-[534px]'">
            <!-- Nền -->
            <template v-if="bgs[bg].cover">
              <div class="absolute -inset-10 -z-10 blur-2xl saturate-150 scale-110"><WfBookCover :title="title" size="lg" /></div>
              <div class="absolute inset-0 -z-10 bg-gradient-to-b from-black/35 via-black/30 to-black/60"></div>
            </template>
            <div v-else class="absolute inset-0 -z-10" :class="bgs[bg].cls"><div class="absolute inset-0 bg-gradient-to-b from-black/10 to-black/40"></div></div>

            <div class="h-full flex flex-col" :class="ratio === 'square' ? 'p-7' : 'px-6 py-8'">
              <!-- Video: bìa lớn ở giữa + sóng âm (thay đĩa than của NCT) -->
              <template v-if="kind === 'video'">
                <div class="flex-1 min-h-0 flex flex-col items-center justify-center gap-4">
                  <div class="rounded-lg shadow-2xl shadow-black/50 overflow-hidden ring-1 ring-white/20" :class="ratio === 'square' ? 'w-24 h-32' : 'w-36 h-48'"><WfBookCover :title="title" size="sm" /></div>
                  <!-- Sóng âm thật của đoạn đọc, sáng dần từ trái sang (như tin nhắn thoại) -->
                  <div class="flex items-center gap-[3px] h-8">
                    <span v-for="(h, i) in waveBars" :key="i" class="w-[3px] rounded-full transition-colors duration-300" :class="i / waveBars.length < vPct / 100 ? 'bg-white' : 'bg-white/35'" :style="{ height: h + '%' }"></span>
                  </div>
                </div>
                <p class="shrink-0 font-bold leading-snug min-h-[3.3em]" :class="ratio === 'square' ? 'text-[19px]' : 'text-[21px]'">{{ cardLines[0] }}</p>
              </template>
              <!-- Ảnh: dấu ngoặc kép + các câu đã chọn -->
              <template v-else>
                <span class="font-serif text-6xl leading-none text-white/50 -mb-2">“</span>
                <div class="flex-1 min-h-0 flex flex-col justify-center gap-2.5">
                  <p v-for="(l, i) in cardLines" :key="i" class="font-bold leading-snug" :class="[ratio === 'square' ? (cardLines.length > 2 ? 'text-[19px]' : 'text-[23px]') : 'text-[22px]']">{{ l }}</p>
                </div>
              </template>

              <!-- Chân thẻ: (bìa nhỏ) tên sách, giọng -->
              <div class="shrink-0 mt-5 flex items-end gap-3">
                <div v-if="kind === 'image'" class="w-9 h-12 rounded shadow-lg overflow-hidden ring-1 ring-white/20 shrink-0"><WfBookCover :title="title" size="sm" /></div>
                <div class="min-w-0 flex-1">
                  <p class="text-[13px] font-semibold leading-tight line-clamp-2">{{ title }}</p>
                  <p class="mt-0.5 text-[11px] text-white/70 flex items-center gap-1"><Mic class="w-3 h-3" /> Giọng {{ voice }}</p>
                </div>
              </div>
              <!-- Vòng 2 — dải thương hiệu (marketing viral): người xem biết ngay Sano để làm gì + địa chỉ tìm tới -->
              <div class="shrink-0 mt-3 pt-3 border-t border-white/20 flex items-center gap-2">
                <img src="@/assets/favicon.svg" alt="" class="h-6 w-6 rounded-md shadow shrink-0" />
                <span class="text-[12px] leading-tight"><b class="font-bold">Sano</b><span class="text-white/80"> · Tự tạo sách nói</span></span>
                <span class="ml-auto text-[11px] font-semibold tracking-wide text-white/90">sanobook.com</span>
              </div>
            </div>
          </div>
          <button v-if="kind === 'video'" class="absolute bottom-3 left-1/2 -translate-x-1/2 h-7 px-3 rounded-full bg-background border border-border text-xs flex items-center gap-1.5" @click="playing = !playing">
            <Play class="w-3 h-3" /> {{ playing ? 'Đang xem thử (không tiếng)' : 'Xem thử' }}
          </button>
        </div>

        <!-- Phải: tuỳ chọn -->
        <div class="flex-1 min-w-0 flex flex-col">
          <div class="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-border">
            <h2 class="font-semibold flex items-center gap-2"><Share2 class="w-4 h-4" /> Chia sẻ đoạn hay</h2>
            <X class="w-4 h-4 text-muted-foreground" />
          </div>

          <div class="flex-1 min-h-0 overflow-auto px-5 py-4 space-y-4 text-sm">
            <!-- Loại -->
            <div class="grid grid-cols-2 gap-2 p-1 rounded-lg bg-muted">
              <button v-for="k in (['image', 'video'] as const)" :key="k" class="h-9 rounded-md flex items-center justify-center gap-2 font-medium"
                :class="kind === k ? 'bg-background shadow-sm' : 'text-muted-foreground'" @click="kind = k">
                <component :is="k === 'image' ? ImageIcon : Film" class="w-4 h-4" /> {{ k === 'image' ? 'Ảnh có lời' : 'Video có tiếng đọc' }}
              </button>
            </div>

            <!-- Chọn câu -->
            <div>
              <div class="flex items-baseline justify-between">
                <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Chọn câu · {{ section }}</span>
                <span class="text-xs text-muted-foreground tabular-nums">{{ picked.length }}/{{ MAX }} câu<template v-if="kind === 'video'"> · {{ Math.round(dur) }} giây</template></span>
              </div>
              <div class="mt-2 max-h-[172px] overflow-auto rounded-lg border border-border divide-y divide-border">
                <button v-for="(s, i) in sentences" :key="i" class="w-full flex items-start gap-2.5 px-3 py-2 text-left"
                  :class="picked.includes(i) ? 'bg-primary/5' : 'hover:bg-muted/60'" @click="toggle(i)">
                  <span class="mt-0.5 h-4 w-4 rounded border grid place-items-center shrink-0"
                    :class="picked.includes(i) ? 'bg-primary border-primary text-primary-foreground' : canAdd(i) ? 'border-primary/50' : 'border-border'">
                    <Check v-if="picked.includes(i)" class="w-3 h-3" />
                  </span>
                  <span class="leading-snug" :class="picked.includes(i) ? 'text-foreground' : 'text-muted-foreground'">{{ s.t }}</span>
                </button>
              </div>
              <p class="mt-1.5 text-[11px] text-muted-foreground">Chọn các câu liền nhau, tối đa {{ MAX }} câu<template v-if="kind === 'video'">, 30 giây — video 15–30 giây dễ được xem hết nhất</template>. Bấm câu ở xa để chọn lại từ đầu.</p>
            </div>

            <!-- Khung -->
            <div>
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Khung</span>
              <div class="mt-2 grid grid-cols-2 gap-2">
                <button v-for="r in (['story', 'square'] as const)" :key="r" class="h-14 rounded-lg border flex items-center gap-3 px-3 text-left"
                  :class="ratio === r ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'" @click="ratio = r">
                  <span class="border-2 rounded-sm shrink-0" :class="[r === 'square' ? 'w-5 h-5' : 'w-4 h-7', ratio === r ? 'border-primary' : 'border-muted-foreground/50']"></span>
                  <span><span class="block font-medium">{{ r === 'square' ? 'Vuông 1:1' : 'Dọc 9:16' }}</span><span class="block text-[11px] text-muted-foreground">{{ r === 'square' ? 'Bài đăng Facebook, Zalo' : 'Story, Reels, TikTok' }}</span></span>
                </button>
              </div>
            </div>

            <!-- Nền -->
            <div>
              <span class="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Nền</span>
              <div class="mt-2 flex gap-2">
                <button v-for="(b, i) in bgs" :key="i" class="flex flex-col items-center gap-1 text-[11px]" :class="bg === i ? 'text-foreground font-medium' : 'text-muted-foreground'" @click="bg = i">
                  <span class="h-11 w-11 rounded-lg overflow-hidden ring-offset-2 ring-offset-background relative" :class="[bg === i ? 'ring-2 ring-primary' : 'ring-1 ring-border', b.cls]">
                    <span v-if="b.cover" class="absolute -inset-2 blur-md"><WfBookCover :title="title" size="sm" /></span>
                  </span>
                  {{ b.label }}
                </button>
              </div>
            </div>
          </div>

          <!-- Hành động -->
          <div class="shrink-0 border-t border-border px-5 py-3.5">
            <template v-if="kind === 'image'">
              <div class="flex items-center gap-2">
                <Button class="flex-1" @click="flash('Đã lưu ảnh vào thư mục Tải về')"><Download class="w-4 h-4" /> Lưu ảnh</Button>
                <Button variant="outline" @click="flash('Đã sao chép ảnh — dán thẳng vào Facebook, Zalo')"><Copy class="w-4 h-4" /> Sao chép ảnh</Button>
                <Button variant="outline"><Share2 class="w-4 h-4" /> AirDrop</Button>
              </div>
            </template>
            <template v-else>
              <div v-if="vstate === 'busy'" class="h-9 rounded-md bg-muted flex items-center justify-center gap-2 text-sm relative overflow-hidden">
                <div class="absolute inset-y-0 left-0 bg-primary/15 transition-all" :style="{ width: progress + '%' }"></div>
                <Loader2 class="w-4 h-4 animate-spin relative" /><span class="relative">Đang tạo video {{ progress }}%</span>
              </div>
              <div v-else-if="vstate === 'done'" class="flex items-center gap-2">
                <Button class="flex-1" @click="flash('Đã lưu video vào thư mục Tải về')"><Download class="w-4 h-4" /> Lưu video</Button>
                <Button variant="outline"><Share2 class="w-4 h-4" /> AirDrop</Button>
              </div>
              <Button v-else class="w-full" @click="makeVideo"><Film class="w-4 h-4" /> Tạo video {{ Math.round(dur) }} giây</Button>
            </template>
            <p class="mt-2 text-[11px] text-muted-foreground text-center">
              {{ kind === 'image' ? 'Ảnh PNG sắc nét 1080 px. Facebook trên máy tính: bấm Sao chép rồi dán (⌘V) vào ô đăng bài.' : 'Video MP4 1080 px kèm tiếng đọc đúng các câu đã chọn. Tạo ngay trên máy, không cần mạng.' }}
            </p>
          </div>
        </div>
      </div>

      <div v-if="toast" class="absolute z-30 bottom-6 left-1/2 -translate-x-1/2 rounded-full bg-foreground text-background text-xs px-3 py-1.5 shadow-lg flex items-center gap-1.5"><Check class="w-3.5 h-3.5" /> {{ toast }}</div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D14 · Chia sẻ đoạn hay: mở từ nút Chia sẻ ở chế độ Phóng to lời đọc hoặc Xem cả lời, mặc định chọn câu đang đọc. Ảnh có lời hoặc video kèm tiếng đọc.
      Vòng 2: mặc định Dọc 9:16 + nền Hoàng hôn; hộp thoại to hơn; chân thẻ có dải "Sano · Tự tạo sách nói · sanobook.com".
    </p>
  </div>
</template>
