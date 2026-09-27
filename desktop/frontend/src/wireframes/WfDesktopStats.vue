<script setup lang="ts">
// D7 — Hành trình nghe (thống kê). Bám khung D1 (1100×720). Mục mới trên thanh bên,
// dưới "Thư viện". Dữ liệu từ nhật ký nghe ~/Sano/.nghe.json (số giây nghe mỗi ngày,
// theo giờ, theo cuốn; ngày nghe xong).
// - Chọn khoảng: Tuần này / Tháng này / Năm nay / Tất cả.
// - 4 ô số: thời gian nghe (so với kỳ trước), chuỗi ngày nghe, sách nghe xong,
//   thời gian tiết kiệm nhờ nghe nhanh (nội dung − thời gian thật).
// - Biểu đồ cột phút nghe mỗi ngày (năm: mỗi tháng) + vạch mục tiêu mỗi ngày.
// - Mục tiêu hôm nay (vòng tiến độ, đặt được số phút), giờ hay nghe (24 cột).
// - Lịch nghe 6 tháng (ô đậm nhạt), sách nghe nhiều nhất, nghe xong gần đây.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, Clock, Flame,
  BookCheck, Gauge, Target, Pencil, Headphones, Check, X,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'week' | 'month' | 'year' | 'goal' | 'empty'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'week')
const dark = ref(false)
const range = computed(() => (mode.value === 'goal' || mode.value === 'empty' ? 'week' : mode.value))

const goal = ref(30)
const goalDraft = ref(30)
const today = 18

// Phút nghe mỗi ngày / mỗi tháng (giả)
const week = [
  { l: 'T2', d: '22/9', m: 42 }, { l: 'T3', d: '23/9', m: 35 }, { l: 'T4', d: '24/9', m: 0 },
  { l: 'T5', d: '25/9', m: 64 }, { l: 'T6', d: '26/9', m: 51 }, { l: 'T7', d: '27/9', m: 18 }, { l: 'CN', d: '28/9', m: 0, future: true },
]
const month = Array.from({ length: 30 }, (_, i) => ({ l: i % 5 === 0 ? String(i + 1) : '', d: `${i + 1}/9`, m: i > 26 ? 0 : [0, 35, 48, 20, 0, 62, 41, 30, 55, 12, 0, 38, 44, 27, 70, 33, 0, 25, 46, 52, 38, 0, 29, 61, 42, 35, 18, 0, 0, 0][i], future: i > 26 }))
const year = ['T1', 'T2', 'T3', 'T4', 'T5', 'T6', 'T7', 'T8', 'T9', 'T10', 'T11', 'T12'].map((l, i) => ({ l, d: `Tháng ${i + 1}`, m: i < 6 ? 0 : [0, 0, 0, 0, 0, 0, 380, 720, 910, 0, 0, 0][i], future: i > 8 }))
const bars = computed(() => (range.value === 'month' ? month : range.value === 'year' ? year : week))
const maxM = computed(() => Math.max(...bars.value.map((b) => b.m), range.value === 'year' ? 1 : goal.value) * 1.15)
const hover = ref<number | null>(3)
// Tuần này: cột hôm nay (T7) tô đỏ

const tiles = computed(() => {
  const r = range.value
  return [
    { icon: Clock, label: 'Thời gian nghe', value: r === 'year' ? '33 giờ 30 phút' : r === 'month' ? '15 giờ 10 phút' : '3 giờ 30 phút', sub: r === 'year' ? 'từ tháng 7, lúc bắt đầu ghi' : r === 'month' ? 'tăng 25% so với tháng 8' : 'tăng 20% so với tuần trước' },
    { icon: Flame, label: 'Chuỗi ngày nghe', value: '3 ngày', sub: 'kỷ lục 12 ngày' },
    { icon: BookCheck, label: 'Sách nghe xong', value: r === 'year' ? '9 cuốn' : r === 'month' ? '4 cuốn' : '1 cuốn', sub: r === 'year' ? 'trung bình 3 cuốn / tháng' : 'Tâm lý tích cực mỗi ngày' },
    { icon: Gauge, label: 'Tiết kiệm nhờ nghe nhanh', value: r === 'year' ? '9 giờ 40 phút' : r === 'month' ? '4 giờ 15 phút' : '55 phút', sub: 'hay nghe ở tốc độ 1,3×' },
  ]
})

const hours = [0, 0, 0, 0, 0, 2, 14, 38, 30, 6, 3, 2, 10, 8, 2, 1, 3, 6, 12, 18, 26, 34, 16, 4]
const maxH = Math.max(...hours)

// Lịch nghe 26 tuần × 7 ngày: 0..4 mức đậm nhạt
const heat = Array.from({ length: 26 }, (_, w) => Array.from({ length: 7 }, (_, d) => {
  if (w < 10) return 0
  if (w === 25 && d > 5) return -1 // tương lai
  const v = (w * 7 + d * 3 + (w % 3) * 5) % 11
  return v < 3 ? 0 : v < 5 ? 1 : v < 8 ? 2 : v < 10 ? 3 : 4
}))
// Một sắc xám nhạt → đậm (đỏ để dành cho hành động chính và "hôm nay")
const heatClass = (v: number) => ['bg-muted', 'bg-slate-300 dark:bg-slate-700', 'bg-slate-400 dark:bg-slate-500', 'bg-slate-500 dark:bg-slate-400', 'bg-slate-700 dark:bg-slate-200'][v] ?? 'bg-transparent'
const barIdle = 'bg-slate-400 dark:bg-slate-500'
const barHover = 'bg-slate-600 dark:bg-slate-300'

const top = [
  { t: 'Kinh doanh cho người mới', a: 'Nguyễn Văn A', m: 380, series: true },
  { t: 'Tâm lý tích cực mỗi ngày', a: 'Phạm Thị D', m: 255 },
  { t: 'Lãnh đạo cho quản lý mới', a: 'Nguyễn Văn A', m: 190 },
  { t: 'Quản lý thời gian hiệu quả', a: 'Nguyễn Văn A', m: 120 },
]
const done = [
  { t: 'Tâm lý tích cực mỗi ngày', d: 'Hôm qua' },
  { t: 'Kinh doanh cho người mới · Tập 1', d: '21/9' },
  { t: 'Tài chính cá nhân cơ bản', d: '14/9' },
]
const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'stats', label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'create', label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const ringLen = 2 * Math.PI * 26
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <!-- Nút duyệt trạng thái (ngoài khung) -->
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['week', '1. Tuần này'], ['month', '2. Tháng này'], ['year', '3. Năm nay'], ['goal', '4. Đặt mục tiêu mỗi ngày'], ['empty', '5. Chưa có dữ liệu'],
      ] as [Mode, string][])" :key="m[0]"
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
              :class="n.key === 'stats' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.13 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <!-- ─── Chưa có dữ liệu ─── -->
        <main v-if="mode === 'empty'" class="flex-1 min-w-0 grid place-items-center p-6">
          <div class="max-w-sm text-center">
            <div class="mx-auto h-14 w-14 rounded-full bg-muted grid place-items-center"><Headphones class="w-7 h-7 text-muted-foreground" /></div>
            <h1 class="mt-4 text-lg font-semibold">Chưa có số liệu nghe</h1>
            <p class="mt-1 text-sm text-muted-foreground">Sano bắt đầu ghi từ bản 0.1.13. Nghe vài phút, quay lại đây sẽ thấy thời gian nghe, chuỗi ngày và sách đã nghe xong.</p>
            <Button class="mt-4"><Library class="w-4 h-4" /> Mở thư viện</Button>
          </div>
        </main>

        <main v-else class="flex-1 min-w-0 overflow-auto px-6 pt-6 pb-6">
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-xl font-semibold tracking-tight">Hành trình nghe</h1>
              <p class="text-sm text-muted-foreground">Số liệu ghi từ 12/7/2026 · chỉ lưu trên máy bạn</p>
            </div>
            <div class="flex rounded-md border border-border p-0.5 text-sm" role="radiogroup" aria-label="Khoảng thời gian">
              <button v-for="r in ([['week', 'Tuần này'], ['month', 'Tháng này'], ['year', 'Năm nay'], ['all', 'Tất cả']] as [string, string][])" :key="r[0]"
                class="h-7 px-3 rounded" :class="range === r[0] ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground'"
                @click="r[0] !== 'all' && (mode = r[0] as Mode)">{{ r[1] }}</button>
            </div>
          </div>

          <!-- 4 ô số -->
          <div class="mt-4 grid grid-cols-4 gap-3">
            <div v-for="t in tiles" :key="t.label" class="rounded-lg border border-border p-3">
              <p class="text-xs text-muted-foreground flex items-center gap-1.5"><component :is="t.icon" class="w-3.5 h-3.5" /> {{ t.label }}</p>
              <p class="mt-1 text-xl font-semibold tracking-tight tabular-nums">{{ t.value }}</p>
              <p class="text-[11px] text-muted-foreground truncate">{{ t.sub }}</p>
            </div>
          </div>

          <div class="mt-3 grid grid-cols-3 gap-3">
            <!-- Biểu đồ cột -->
            <div class="col-span-2 rounded-lg border border-border p-4">
              <div class="flex items-center justify-between">
                <h2 class="text-sm font-medium">Phút nghe {{ range === 'year' ? 'mỗi tháng' : 'mỗi ngày' }}</h2>
                <span v-if="range !== 'year'" class="text-[11px] text-muted-foreground flex items-center gap-1.5"><span class="w-4 border-t border-dashed border-foreground/50"></span> Mục tiêu {{ goal }} phút</span>
              </div>
              <div class="relative mt-3 h-40">
                <!-- lưới ngang nhạt -->
                <div v-for="g in 3" :key="g" class="absolute inset-x-0 border-t border-border/60" :style="{ bottom: `${(g / 3) * 100}%` }"></div>
                <!-- vạch mục tiêu -->
                <div v-if="range !== 'year'" class="absolute inset-x-0 border-t border-dashed border-foreground/50 z-10" :style="{ bottom: `${(goal / maxM) * 100}%` }"></div>
                <div class="absolute inset-0 flex items-end" :class="range === 'month' ? 'gap-[2px]' : 'gap-3'">
                  <div v-for="(b, i) in bars" :key="i" class="relative flex-1 h-full flex items-end justify-center" @mouseenter="hover = i" @mouseleave="hover = null">
                    <div class="w-full rounded-t-[4px] transition-colors"
                      :class="[b.future ? 'bg-transparent' : b.m === 0 ? 'bg-muted' : range === 'week' && i === 5 ? 'bg-primary' : hover === i ? barHover : barIdle, range === 'week' && 'max-w-10']"
                      :style="{ height: b.m === 0 ? '2px' : `${(b.m / maxM) * 100}%` }"></div>
                    <!-- tooltip -->
                    <div v-if="hover === i && !b.future" class="absolute bottom-full mb-1 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-md border border-border bg-popover shadow-md px-2 py-1 text-[11px] z-20" :style="{ bottom: `${(b.m / maxM) * 100}%` }">
                      <span class="text-muted-foreground">{{ b.d }}</span> · <b class="tabular-nums">{{ range === 'year' ? `${Math.floor(b.m / 60)} giờ ${b.m % 60} phút` : `${b.m} phút` }}</b>
                    </div>
                  </div>
                </div>
              </div>
              <div class="mt-1.5 flex text-[11px] text-muted-foreground" :class="range === 'month' ? 'gap-[2px]' : 'gap-3'">
                <span v-for="(b, i) in bars" :key="i" class="flex-1 text-center" :class="range === 'week' && i === 5 && 'font-medium text-foreground'">{{ b.l }}</span>
              </div>
            </div>

            <!-- Mục tiêu hôm nay -->
            <div class="rounded-lg border border-border p-4 flex flex-col">
              <div class="flex items-center justify-between">
                <h2 class="text-sm font-medium flex items-center gap-1.5"><Target class="w-4 h-4 text-muted-foreground" /> Mục tiêu hôm nay</h2>
                <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-muted" title="Đổi mục tiêu" @click="mode = 'goal'"><Pencil class="w-3.5 h-3.5" /></button>
              </div>
              <div class="flex-1 flex items-center gap-4 mt-2">
                <svg viewBox="0 0 64 64" class="h-24 w-24 -rotate-90 shrink-0" aria-hidden="true">
                  <circle cx="32" cy="32" r="26" fill="none" stroke-width="7" class="stroke-muted" />
                  <circle cx="32" cy="32" r="26" fill="none" stroke-width="7" stroke-linecap="round" class="stroke-primary"
                    :stroke-dasharray="`${(today / goal) * ringLen} ${ringLen}`" />
                </svg>
                <div>
                  <p class="text-2xl font-semibold tabular-nums">{{ today }}<span class="text-base text-muted-foreground font-normal"> / {{ goal }} phút</span></p>
                  <p class="text-xs text-muted-foreground mt-0.5">Còn {{ goal - today }} phút là đạt. Tuần này đạt 3 / 6 ngày.</p>
                </div>
              </div>
            </div>
          </div>

          <div class="mt-3 grid grid-cols-3 gap-3">
            <!-- Lịch nghe -->
            <div class="col-span-2 rounded-lg border border-border p-4">
              <div class="flex items-center justify-between">
                <h2 class="text-sm font-medium">Lịch nghe 6 tháng</h2>
                <span class="flex items-center gap-1 text-[11px] text-muted-foreground">Ít
                  <span v-for="v in 5" :key="v" class="h-2.5 w-2.5 rounded-[2px]" :class="heatClass(v - 1)"></span> Nhiều</span>
              </div>
              <div class="mt-3 flex gap-2">
                <div class="flex flex-col gap-[3px] text-[10px] text-muted-foreground pt-4">
                  <span v-for="(d, i) in ['T2', '', 'T4', '', 'T6', '', 'CN']" :key="i" class="h-3 leading-3">{{ d }}</span>
                </div>
                <div class="flex-1">
                  <div class="flex text-[10px] text-muted-foreground h-4">
                    <span v-for="m in ['T4', 'T5', 'T6', 'T7', 'T8', 'T9']" :key="m" class="flex-1">{{ m }}</span>
                  </div>
                  <div class="flex gap-[3px]">
                    <div v-for="(w, wi) in heat" :key="wi" class="flex flex-col gap-[3px] flex-1">
                      <span v-for="(v, di) in w" :key="di" class="h-3 rounded-[2px]" :class="heatClass(v)" :title="v >= 0 ? `${['Không nghe', 'Dưới 15 phút', '15–30 phút', '30–60 phút', 'Trên 1 giờ'][v]}` : ''"></span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Giờ hay nghe -->
            <div class="rounded-lg border border-border p-4">
              <h2 class="text-sm font-medium">Giờ hay nghe</h2>
              <p class="text-xs text-muted-foreground">Bạn nghe nhiều nhất lúc 7–8 giờ sáng và 21 giờ tối</p>
              <div class="mt-3 h-16 flex items-end gap-[2px]">
                <div v-for="(h, i) in hours" :key="i" class="flex-1 rounded-t-[2px]" :class="h === 0 ? 'bg-muted' : barIdle" :style="{ height: h === 0 ? '2px' : `${(h / maxH) * 100}%` }" :title="`${i} giờ: ${h} phút`"></div>
              </div>
              <div class="mt-1 flex justify-between text-[10px] text-muted-foreground"><span>0h</span><span>6h</span><span>12h</span><span>18h</span><span>23h</span></div>
            </div>
          </div>

          <div class="mt-3 grid grid-cols-2 gap-3">
            <!-- Nghe nhiều nhất -->
            <div class="rounded-lg border border-border p-4">
              <h2 class="text-sm font-medium">Nghe nhiều nhất</h2>
              <div class="mt-2 space-y-2">
                <div v-for="(b, i) in top" :key="b.t" class="flex items-center gap-3">
                  <span class="w-4 text-xs text-muted-foreground tabular-nums">{{ i + 1 }}</span>
                  <div class="h-10 w-[30px] shrink-0 rounded overflow-hidden shadow-sm"><WfBookCover :title="b.t" :author="b.a" size="sm" /></div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm truncate">{{ b.t }}</p>
                    <div class="mt-1 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full rounded-full" :class="barIdle" :style="{ width: `${(b.m / top[0].m) * 100}%` }"></div></div>
                  </div>
                  <span class="w-20 text-right text-xs text-muted-foreground tabular-nums">{{ Math.floor(b.m / 60) }} giờ {{ b.m % 60 }} phút</span>
                </div>
              </div>
            </div>
            <!-- Nghe xong gần đây -->
            <div class="rounded-lg border border-border p-4">
              <h2 class="text-sm font-medium">Nghe xong gần đây</h2>
              <div class="mt-2 divide-y divide-border">
                <div v-for="b in done" :key="b.t" class="flex items-center gap-3 py-2">
                  <span class="h-6 w-6 grid place-items-center rounded-full bg-muted text-muted-foreground"><Check class="w-3.5 h-3.5" /></span>
                  <span class="flex-1 text-sm truncate">{{ b.t }}</span>
                  <span class="text-xs text-muted-foreground">{{ b.d }}</span>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>

      <!-- ═══ Đặt mục tiêu mỗi ngày ═══ -->
      <div v-if="mode === 'goal'" class="absolute inset-0 bg-black/40 grid place-items-center z-30">
        <div class="w-[400px] rounded-xl border border-border bg-background shadow-2xl p-5">
          <div class="flex items-start justify-between">
            <div>
              <h2 class="font-semibold">Mục tiêu nghe mỗi ngày</h2>
              <p class="text-xs text-muted-foreground mt-0.5">Chỉ để tự nhắc mình, không có thông báo làm phiền.</p>
            </div>
            <button class="text-muted-foreground" @click="mode = 'week'"><X class="w-4 h-4" /></button>
          </div>
          <div class="mt-4 grid grid-cols-5 gap-2">
            <button v-for="m in [15, 30, 45, 60, 90]" :key="m" class="h-10 rounded-md border text-sm tabular-nums"
              :class="goalDraft === m ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border hover:bg-muted'" @click="goalDraft = m">{{ m }} phút</button>
          </div>
          <label class="mt-3 flex items-center gap-2 text-sm text-muted-foreground">Hoặc tự nhập
            <input v-model.number="goalDraft" type="number" min="5" max="600" class="w-20 h-8 rounded-md border border-input bg-background px-2 text-foreground" /> phút
          </label>
          <div class="mt-5 flex justify-between gap-2">
            <Button variant="ghost" class="text-muted-foreground" @click="mode = 'week'">Tắt mục tiêu</Button>
            <div class="flex gap-2">
              <Button variant="outline" @click="mode = 'week'">Huỷ</Button>
              <Button @click="goal = goalDraft; mode = 'week'">Lưu</Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D7 · Hành trình nghe — mục mới trên thanh bên. Dữ liệu từ nhật ký nghe trên máy (không gửi đi đâu). Rê chuột lên cột để xem số phút.
      "Tiết kiệm nhờ nghe nhanh" = thời lượng nội dung đã nghe trừ thời gian thật (nghe 1,5× thì 60 phút nội dung mất 40 phút).
    </p>
  </div>
</template>
