<script setup lang="ts">
// Hành trình nghe (wireframe D7): thống kê từ nhật ký nghe ~/Sano/.nghe.json — thời gian
// nghe, chuỗi ngày, sách nghe xong, thời gian tiết kiệm nhờ nghe nhanh, phút nghe mỗi
// ngày + mục tiêu, lịch nghe 6 tháng, giờ hay nghe, nghe nhiều nhất, nghe xong gần đây.
import { computed, onMounted, ref } from 'vue'
import { BarChart3, BookCheck, Check, ChevronDown, Clock, Flame, Gauge, Headphones, Library, Pencil, Target, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import BookCover from '@/components/sano/BookCover.vue'
import { errText, listenLog, type ListenLog } from '../lib/backend'
import { flushListening } from '../lib/listenlog'
import { go, openBook, refreshLibrary, state } from '../lib/store'
import {
  bars as makeBars, dayKey, finishedIn, firstDay, fmtDay, fmtDur, fmtRecent, heatmap, heatmapMonths, hours as makeHours,
  loadGoal, loadRange, peakHours, rangeBounds, saveGoal, saveRange, streaks, sumRange, topBooks, years as makeYears, type StatsRange,
} from '../lib/stats'

const log = ref<ListenLog | null>(null)
const error = ref('')
const today = ref(new Date())
onMounted(async () => {
  try {
    await flushListening() // phút vừa nghe chưa kịp ghi
    const [l] = await Promise.all([listenLog(), state.library ? Promise.resolve() : refreshLibrary()])
    log.value = l
    today.value = new Date()
  } catch (e) {
    error.value = errText(e)
  }
})

const range = ref<StatsRange>(loadRange())
const year = ref(new Date().getFullYear())
const yearOpen = ref(false)
const yearList = computed(() => makeYears(L.value, today.value))
const thisYear = computed(() => year.value === today.value.getFullYear())
function setRange(r: StatsRange, y?: number) {
  range.value = r
  if (y) year.value = y
  saveRange(r)
  hover.value = null
  yearOpen.value = false
}
const PREV = computed<Record<StatsRange, string>>(() => ({ week: 'tuần trước', month: 'tháng trước', year: thisYear.value ? 'năm trước' : `năm ${year.value - 1}`, all: '' }))

const L = computed<ListenLog>(() => log.value ?? { version: 1, days: {}, finished: {} })
const first = computed(() => firstDay(L.value))
const bounds = computed(() => rangeBounds(range.value, today.value, L.value, year.value))
const sum = computed(() => sumRange(L.value, bounds.value.from, bounds.value.to))
const prev = computed(() => (bounds.value.prevFrom ? sumRange(L.value, bounds.value.prevFrom, bounds.value.prevTo) : null))
const streak = computed(() => streaks(L.value, today.value))
const finished = computed(() => finishedIn(L.value, bounds.value.from, bounds.value.to))
const bookBySlug = computed(() => new Map((state.library?.books ?? []).map((b) => [b.slug, b])))
const titleOf = (slug: string) => {
  const b = bookBySlug.value.get(slug)
  if (!b) return 'Sách đã xoá'
  return b.series ? `${b.series} · Tập ${b.volume}` : b.title
}

const tiles = computed(() => {
  const s = sum.value
  const p = prev.value
  let timeSub = first.value ? `từ ${fmtDay(first.value)}, lúc bắt đầu ghi` : ''
  if (p && range.value !== 'all') {
    if (first.value && bounds.value.prevTo < first.value) timeSub = `bắt đầu ghi từ ${fmtDay(first.value)}`
    else if (p.listen < 60) timeSub = `${PREV.value[range.value]} chưa nghe`
    else {
      const pct = Math.round(((s.listen - p.listen) / p.listen) * 100)
      timeSub = pct >= 200 ? `gấp ${Math.round(s.listen / p.listen)} lần ${PREV.value[range.value]}` : pct === 0 ? `bằng ${PREV.value[range.value]}` : `${pct > 0 ? 'tăng' : 'giảm'} ${Math.abs(pct)}% so với ${PREV.value[range.value]}`
    }
  }
  const saved = s.audio - s.listen
  const speed = s.listen > 0 ? s.audio / s.listen : 1
  const lastDone = finished.value[0]
  return [
    { icon: Clock, tint: 'bg-chart/10 text-chart', label: 'Thời gian nghe', value: fmtDur(s.listen), sub: timeSub },
    { icon: Flame, tint: 'bg-rag-amber/15 text-rag-amber', label: 'Chuỗi ngày nghe', value: `${streak.value.current} ngày`, sub: streak.value.best > streak.value.current ? `kỷ lục ${streak.value.best} ngày` : streak.value.current ? 'đang là kỷ lục' : 'nghe hôm nay để bắt đầu' },
    { icon: BookCheck, tint: 'bg-rag-green/15 text-rag-green', label: 'Sách nghe xong', value: `${finished.value.length} cuốn`, sub: lastDone ? titleOf(lastDone.slug) : 'chưa có cuốn nào' },
    {
      icon: Gauge, tint: 'bg-chart/10 text-chart', label: 'Tiết kiệm nhờ nghe nhanh', value: fmtDur(Math.max(0, saved)),
      sub: saved >= 60 ? `nghe trung bình ở tốc độ ${speed.toFixed(1).replace('.', ',')}×` : 'nghe tốc độ bình thường',
    },
  ]
})

// ── Biểu đồ cột ──
const goal = ref(loadGoal())
const dayRange = computed(() => range.value === 'week' || range.value === 'month')
const chart = computed(() => makeBars(L.value, range.value, today.value, year.value))
const maxMin = computed(() => Math.max(1, ...chart.value.map((b) => b.min), dayRange.value ? goal.value : 0) * 1.15)
const hover = ref<number | null>(null)
const fmtMin = (m: number) => fmtDur(m * 60)

// ── Mục tiêu ──
const todayMin = computed(() => (L.value.days[dayKey(today.value)] ? Object.values(L.value.days[dayKey(today.value)].books ?? {}).reduce((n, b) => n + (b.listen || 0), 0) / 60 : 0))
const weekGoal = computed(() => {
  const w = makeBars(L.value, 'week', today.value).filter((b) => !b.future)
  return { hit: w.filter((b) => b.min >= goal.value).length, days: w.length }
})
const ringLen = 2 * Math.PI * 26
const editingGoal = ref(false)
const goalDraft = ref(goal.value || 30)
function saveGoalDraft(v: number) {
  goal.value = Math.max(0, Math.min(600, Math.round(v || 0)))
  saveGoal(goal.value)
  editingGoal.value = false
}

// ── Lịch nghe, giờ hay nghe, nghe nhiều nhất, nghe xong gần đây ──
const heat = computed(() => heatmap(L.value, today.value, 26))
const heatMonths = computed(() => heatmapMonths(heat.value))
// Một sắc xanh ngọc (--chart) nhạt → đậm; đỏ để dành cho hành động chính
const heatClass = (v: number) => ['bg-muted', 'bg-chart/25', 'bg-chart/50', 'bg-chart/75', 'bg-chart'][v] ?? 'bg-transparent'
const HEAT_TEXT = ['Không nghe', 'Dưới 15 phút', '15–30 phút', '30–60 phút', 'Trên 1 giờ']
const barIdle = 'bg-chart/55'
const hourList = computed(() => makeHours(L.value, bounds.value.from, bounds.value.to))
const maxHour = computed(() => Math.max(1, ...hourList.value))
const peaks = computed(() => peakHours(hourList.value))
const peakText = computed(() => {
  const p = peaks.value
  if (!p.length) return 'Chưa có số liệu trong khoảng này'
  const part = (h: number) => (h < 11 ? 'sáng' : h < 14 ? 'trưa' : h < 18 ? 'chiều' : 'tối')
  return `Bạn nghe nhiều nhất lúc ${p.map((h) => `${h}–${h + 1} giờ ${part(h)}`).join(' và ')}`
})
const top = computed(() => topBooks(L.value, bounds.value.from, bounds.value.to, 5))
const recentDone = computed(() => finishedIn(L.value, '0000-01-01', '9999-12-31').slice(0, 5))
const empty = computed(() => !!log.value && !first.value)
</script>

<template>
  <section class="flex-1 min-w-0 overflow-auto">
    <!-- Chưa có số liệu -->
    <div v-if="empty" class="h-full grid place-items-center p-6">
      <div class="max-w-sm text-center">
        <div class="mx-auto h-14 w-14 rounded-full bg-muted grid place-items-center"><Headphones class="w-7 h-7 text-muted-foreground" /></div>
        <h1 class="mt-4 text-lg font-semibold">Chưa có số liệu nghe</h1>
        <p class="mt-1 text-sm text-muted-foreground">Sano bắt đầu ghi từ bản 0.1.14. Nghe vài phút, quay lại đây sẽ thấy thời gian nghe, chuỗi ngày và sách đã nghe xong.</p>
        <Button class="mt-4" @click="go('library')"><Library class="w-4 h-4" /> Mở thư viện</Button>
      </div>
    </div>

    <div v-else class="px-6 pt-6 pb-6">
      <div class="flex items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-semibold tracking-tight">Hành trình nghe</h1>
          <p class="text-sm text-muted-foreground">{{ first ? `Số liệu ghi từ ${fmtDay(first)}` : 'Đang đọc số liệu…' }} · chỉ lưu trên máy bạn</p>
        </div>
        <div class="relative flex rounded-md border border-border p-0.5 text-sm shrink-0" role="radiogroup" aria-label="Khoảng thời gian">
          <button v-for="r in ([['week', 'Tuần này'], ['month', 'Tháng này']] as [StatsRange, string][])" :key="r[0]" role="radio" :aria-checked="range === r[0]" class="h-7 px-3 rounded"
            :class="range === r[0] ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground'"
            @click="setRange(r[0])">{{ r[1] }}</button>
          <!-- Năm: chọn năm để xem lại năm cũ -->
          <button role="radio" :aria-checked="range === 'year'" aria-haspopup="listbox" :aria-expanded="yearOpen" class="h-7 pl-3 pr-2 rounded flex items-center gap-1"
            :class="range === 'year' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground'"
            @click="yearList.length > 1 ? (yearOpen = !yearOpen) : setRange('year', yearList[0])">
            {{ range === 'year' && !thisYear ? `Năm ${year}` : 'Năm nay' }}<ChevronDown v-if="yearList.length > 1" class="w-3.5 h-3.5 opacity-60" />
          </button>
          <button role="radio" :aria-checked="range === 'all'" class="h-7 px-3 rounded"
            :class="range === 'all' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground'"
            @click="setRange('all')">Tất cả</button>
          <div v-if="yearOpen" role="listbox" aria-label="Chọn năm" class="absolute right-16 top-full mt-1 w-36 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-30">
            <button v-for="y in yearList" :key="y" role="option" :aria-selected="range === 'year' && year === y" class="w-full flex items-center justify-between px-3 py-1.5 text-left hover:bg-muted"
              :class="range === 'year' && year === y && 'text-primary font-medium'" @click="setRange('year', y)">
              {{ y === today.getFullYear() ? `Năm nay (${y})` : `Năm ${y}` }}<Check v-if="range === 'year' && year === y" class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
      <p v-if="error" class="mt-3 text-sm text-destructive">Không đọc được nhật ký nghe: {{ error }}</p>

      <!-- 4 ô số -->
      <div class="mt-4 grid grid-cols-4 gap-3">
        <div v-for="t in tiles" :key="t.label" class="rounded-lg border border-border p-3 min-w-0">
          <p class="text-xs text-muted-foreground flex items-center gap-2">
            <span class="h-6 w-6 grid place-items-center rounded-md shrink-0" :class="t.tint"><component :is="t.icon" class="w-3.5 h-3.5" /></span> {{ t.label }}
          </p>
          <p class="mt-1.5 text-xl font-semibold tracking-tight tabular-nums">{{ t.value }}</p>
          <p class="text-[11px] text-muted-foreground truncate" :title="t.sub">{{ t.sub }}</p>
        </div>
      </div>

      <div class="mt-3 grid grid-cols-3 gap-3">
        <!-- Phút nghe mỗi ngày / tháng -->
        <div class="col-span-2 rounded-lg border border-border p-4">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-medium">Phút nghe {{ dayRange ? 'mỗi ngày' : 'mỗi tháng' }}</h2>
            <span v-if="dayRange && goal" class="text-[11px] text-muted-foreground flex items-center gap-1.5"><span class="w-4 border-t border-dashed border-foreground/50"></span> Mục tiêu {{ goal }} phút</span>
          </div>
          <div class="relative mt-3 h-40">
            <div v-for="g in 3" :key="g" class="absolute inset-x-0 border-t border-border/60" :style="{ bottom: `${(g / 3) * 100}%` }"></div>
            <div v-if="dayRange && goal" class="absolute inset-x-0 border-t border-dashed border-foreground/50 z-10 pointer-events-none" :style="{ bottom: `${(goal / maxMin) * 100}%` }"></div>
            <div class="absolute inset-0 flex items-end" :class="chart.length > 12 ? 'gap-[2px]' : 'gap-3'">
              <div v-for="(b, i) in chart" :key="i" class="relative flex-1 h-full flex items-end justify-center" @mouseenter="hover = i" @mouseleave="hover = null">
                <div class="w-full rounded-t-[4px] transition-colors"
                  :class="[b.future ? 'bg-transparent' : b.min < 0.5 ? 'bg-muted' : b.today || hover === i ? 'bg-chart' : barIdle, chart.length <= 7 && 'max-w-10']"
                  :style="{ height: b.min < 0.5 ? '2px' : `${(b.min / maxMin) * 100}%` }"></div>
                <div v-if="hover === i && !b.future" class="absolute mb-1 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-md border border-border bg-popover text-popover-foreground shadow-md px-2 py-1 text-[11px] z-20 pointer-events-none"
                  :style="{ bottom: `calc(${Math.max(b.min, 0) / maxMin * 100}% + 4px)` }">
                  <span class="text-muted-foreground">{{ b.title }}</span> · <b class="tabular-nums">{{ fmtMin(b.min) }}</b>
                </div>
              </div>
            </div>
          </div>
          <div class="mt-1.5 flex text-[11px] text-muted-foreground" :class="chart.length > 12 ? 'gap-[2px]' : 'gap-3'">
            <span v-for="(b, i) in chart" :key="i" class="flex-1 text-center whitespace-nowrap" :class="b.today && 'font-medium text-foreground'">{{ b.label }}</span>
          </div>
        </div>

        <!-- Mục tiêu hôm nay -->
        <div class="rounded-lg border border-border p-4 flex flex-col">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-medium flex items-center gap-1.5"><Target class="w-4 h-4 text-muted-foreground" /> Mục tiêu hôm nay</h2>
            <button class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground" title="Đổi mục tiêu" aria-label="Đổi mục tiêu" @click="goalDraft = goal || 30; editingGoal = true"><Pencil class="w-3.5 h-3.5" /></button>
          </div>
          <div v-if="goal" class="flex-1 flex items-center gap-4 mt-2">
            <svg viewBox="0 0 64 64" class="h-24 w-24 -rotate-90 shrink-0" aria-hidden="true">
              <circle cx="32" cy="32" r="26" fill="none" stroke-width="7" class="stroke-muted" />
              <circle v-if="todayMin > 0" cx="32" cy="32" r="26" fill="none" stroke-width="7" stroke-linecap="round" class="stroke-chart"
                :stroke-dasharray="`${Math.min(1, todayMin / goal) * ringLen} ${ringLen}`" />
            </svg>
            <div>
              <p class="text-2xl font-semibold tabular-nums">{{ Math.floor(todayMin) }}<span class="text-base text-muted-foreground font-normal"> / {{ goal }} phút</span></p>
              <p class="text-xs text-muted-foreground mt-0.5">
                <template v-if="todayMin >= goal"><Check class="w-3.5 h-3.5 inline -mt-0.5" /> Đã đạt mục tiêu hôm nay.</template>
                <template v-else>Còn {{ Math.ceil(goal - todayMin) }} phút là đạt.</template>
                Tuần này đạt {{ weekGoal.hit }} / {{ weekGoal.days }} ngày.
              </p>
            </div>
          </div>
          <div v-else class="flex-1 flex flex-col items-start justify-center gap-2 mt-2">
            <p class="text-sm text-muted-foreground">Chưa đặt mục tiêu. Đặt một con số nhỏ, ví dụ 15 phút mỗi ngày, để giữ thói quen nghe.</p>
            <Button size="sm" variant="outline" @click="goalDraft = 15; editingGoal = true"><Target class="w-4 h-4" /> Đặt mục tiêu</Button>
          </div>
        </div>
      </div>

      <div class="mt-3 grid grid-cols-3 gap-3">
        <!-- Lịch nghe 6 tháng -->
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
            <div class="flex-1 min-w-0">
              <div class="flex gap-[3px] text-[10px] text-muted-foreground h-4">
                <span v-for="(m, i) in heatMonths" :key="i" class="flex-1 whitespace-nowrap overflow-visible">{{ m }}</span>
              </div>
              <div class="flex gap-[3px]">
                <div v-for="(w, wi) in heat" :key="wi" class="flex flex-col gap-[3px] flex-1">
                  <span v-for="c in w" :key="c.key" class="h-3 rounded-[2px]" :class="heatClass(c.level)"
                    :title="c.level >= 0 ? `${fmtDay(c.key)}: ${c.min >= 1 ? fmtMin(c.min) : HEAT_TEXT[0]}` : ''"></span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Giờ hay nghe -->
        <div class="rounded-lg border border-border p-4">
          <h2 class="text-sm font-medium">Giờ hay nghe</h2>
          <p class="text-xs text-muted-foreground">{{ peakText }}</p>
          <div class="mt-3 h-16 flex items-end gap-[2px]">
            <div v-for="(h, i) in hourList" :key="i" class="flex-1 rounded-t-[2px]" :class="h < 0.5 ? 'bg-muted' : peaks.includes(i) ? 'bg-chart' : barIdle"
              :style="{ height: h < 0.5 ? '2px' : `${(h / maxHour) * 100}%` }" :title="`${i} giờ: ${fmtMin(h)}`"></div>
          </div>
          <div class="mt-1 flex justify-between text-[10px] text-muted-foreground"><span>0h</span><span>6h</span><span>12h</span><span>18h</span><span>23h</span></div>
        </div>
      </div>

      <div class="mt-3 grid grid-cols-2 gap-3">
        <!-- Nghe nhiều nhất -->
        <div class="rounded-lg border border-border p-4">
          <h2 class="text-sm font-medium">Nghe nhiều nhất</h2>
          <p v-if="!top.length" class="mt-2 text-sm text-muted-foreground">Chưa nghe cuốn nào trong khoảng này.</p>
          <div v-else class="mt-2 space-y-1">
            <button v-for="(b, i) in top" :key="b.slug" class="w-full flex items-center gap-3 rounded-md px-1 py-1 text-left hover:bg-muted/50 disabled:hover:bg-transparent"
              :disabled="!bookBySlug.get(b.slug)" @click="openBook(b.slug)">
              <span class="w-4 text-xs text-muted-foreground tabular-nums">{{ i + 1 }}</span>
              <div class="h-10 w-[30px] shrink-0 rounded overflow-hidden shadow-sm bg-muted">
                <img v-if="bookBySlug.get(b.slug)?.coverUrl" :src="bookBySlug.get(b.slug)!.coverUrl" alt="" class="h-full w-full object-cover" />
                <BookCover v-else-if="bookBySlug.get(b.slug)" :title="bookBySlug.get(b.slug)!.title" :author="bookBySlug.get(b.slug)!.author" class="h-full w-full rounded shadow-none" />
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm truncate" :class="!bookBySlug.get(b.slug) && 'text-muted-foreground italic'">{{ titleOf(b.slug) }}</p>
                <div class="mt-1 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full rounded-full" :class="barIdle" :style="{ width: `${(b.sec / top[0].sec) * 100}%` }"></div></div>
              </div>
              <span class="w-24 text-right text-xs text-muted-foreground tabular-nums">{{ fmtDur(b.sec) }}</span>
            </button>
          </div>
        </div>
        <!-- Nghe xong gần đây -->
        <div class="rounded-lg border border-border p-4">
          <h2 class="text-sm font-medium">Nghe xong gần đây</h2>
          <p v-if="!recentDone.length" class="mt-2 text-sm text-muted-foreground">Nghe hết một cuốn, cuốn đó sẽ hiện ở đây.</p>
          <div v-else class="mt-2 divide-y divide-border">
            <div v-for="b in recentDone" :key="b.slug" class="flex items-center gap-3 py-2">
              <span class="h-6 w-6 grid place-items-center rounded-full bg-rag-green/15 text-rag-green shrink-0"><Check class="w-3.5 h-3.5" /></span>
              <span class="flex-1 text-sm truncate" :class="!bookBySlug.get(b.slug) && 'text-muted-foreground italic'">{{ titleOf(b.slug) }}</span>
              <span class="text-xs text-muted-foreground shrink-0">{{ fmtRecent(b.day, today) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Đặt mục tiêu mỗi ngày -->
    <div v-if="editingGoal" class="fixed inset-0 bg-black/40 grid place-items-center z-40" @click.self="editingGoal = false">
      <div role="dialog" aria-label="Mục tiêu nghe mỗi ngày" class="w-[400px] rounded-xl border border-border bg-background shadow-2xl p-5">
        <div class="flex items-start justify-between">
          <div>
            <h2 class="font-semibold flex items-center gap-2"><BarChart3 class="w-4 h-4 text-muted-foreground" /> Mục tiêu nghe mỗi ngày</h2>
            <p class="text-xs text-muted-foreground mt-0.5">Chỉ để tự nhắc mình, không có thông báo làm phiền.</p>
          </div>
          <button class="text-muted-foreground hover:text-foreground" aria-label="Đóng" @click="editingGoal = false"><X class="w-4 h-4" /></button>
        </div>
        <div class="mt-4 grid grid-cols-5 gap-2">
          <button v-for="m in [15, 30, 45, 60, 90]" :key="m" class="h-10 rounded-md border text-sm tabular-nums"
            :class="goalDraft === m ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border hover:bg-muted'" @click="goalDraft = m">{{ m }} phút</button>
        </div>
        <label class="mt-3 flex items-center gap-2 text-sm text-muted-foreground">Hoặc tự nhập
          <input v-model.number="goalDraft" type="number" min="5" max="600" class="w-20 h-8 rounded-md border border-input bg-background px-2 text-foreground" /> phút
        </label>
        <div class="mt-5 flex justify-between gap-2">
          <Button variant="ghost" class="text-muted-foreground" @click="saveGoalDraft(0)">Tắt mục tiêu</Button>
          <div class="flex gap-2">
            <Button variant="outline" @click="editingGoal = false">Huỷ</Button>
            <Button :disabled="!(goalDraft >= 5 && goalDraft <= 600)" @click="saveGoalDraft(goalDraft)">Lưu</Button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
