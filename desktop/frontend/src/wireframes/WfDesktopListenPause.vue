<script setup lang="ts">
// D9 — Màn nghe: chỉnh quãng nghỉ riêng cho từng cuốn. Bám khung D1 (1100×720) và màn nghe thật.
// Cạnh nút tốc độ thêm nút "Nghỉ": mở menu chọn Theo cài đặt chung / Ngắn / Vừa / Dài.
// - Chọn là áp dụng ngay (lần chuyển tiểu mục kế tiếp) và lưu riêng cho cuốn này.
// - Xuất M4B cuốn này dùng đúng mức đã chọn.
// - "Theo cài đặt chung" = mức ở Cài đặt → Nghe và xuất M4B (wireframe D8).
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, ChevronLeft, ChevronDown, Check, Gauge,
  Pause, SkipBack, SkipForward, RotateCcw, RotateCw, Download, Package, FolderOpen, Trash2, Volume2, Mic, Timer,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Level = 'global' | 'short' | 'medium' | 'long' | 'custom'
type Mode = 'closed' | 'menu' | 'picked' | 'custom'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'closed')
const dark = ref(false)
const level = ref<Level>(mode.value === 'picked' ? 'long' : mode.value === 'custom' ? 'custom' : 'global')
const customSection = ref(1.2)
const customChapter = ref(2.5)
const fmtSec = (n: number) => n.toLocaleString('vi-VN') + ' giây'
const globalLevel = 'Vừa'

const levels: { key: Level; label: string; sub: string }[] = [
  { key: 'global', label: `Theo cài đặt chung (${globalLevel})`, sub: '1,5 giây giữa tiểu mục · 2 giây sang chương' },
  { key: 'short', label: 'Ngắn', sub: '0,5 giây · 1 giây' },
  { key: 'medium', label: 'Vừa', sub: '1,5 giây · 2 giây' },
  { key: 'long', label: 'Dài', sub: '3 giây · 4 giây' },
]
const customSub = computed(() => `${fmtSec(customSection.value)} · ${fmtSec(customChapter.value)}`)
const pillLabel = computed(() =>
  level.value === 'global' ? 'Nghỉ: theo chung'
    : level.value === 'custom' ? `Nghỉ: ${customSection.value.toLocaleString('vi-VN')} giây`
      : `Nghỉ: ${levels.find((l) => l.key === level.value)!.label}`)

function setMode(m: Mode) {
  mode.value = m
  level.value = m === 'picked' ? 'long' : m === 'custom' ? 'custom' : 'global'
}
function pick(l: Level) {
  level.value = l
  mode.value = l === 'global' ? 'closed' : 'picked'
}

const toc = [
  { t: 'Lời mở đầu', d: '0:54', heard: true },
  { t: 'Chai Sunlight và chai nước rửa chén', d: '0:35', heard: true },
  { t: 'Bốn áp lực buộc phải có thương hiệu', d: '0:48', cur: true },
  { t: 'Ba thời kỳ của marketing', d: '0:30' },
  { t: 'Thương hiệu là gì', d: '1:12' },
  { t: 'Tài sản thương hiệu', d: '0:58' },
  { t: 'Định vị trong tâm trí', d: '1:05' },
  { t: 'Câu chuyện thương hiệu', d: '0:47' },
  { t: 'Kết chương', d: '0:22' },
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
    <!-- Nút duyệt trạng thái (ngoài khung) -->
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([['closed', '1. Đang nghe'], ['menu', '2. Mở menu Nghỉ'], ['picked', '3. Đã chọn Dài cho cuốn này'], ['custom', '4. Tuỳ chỉnh số giây']] as [Mode, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="setMode(m[0])">{{ m[1] }}</button>
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
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.16 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <section class="flex-1 flex min-h-0">
          <div class="flex-1 flex flex-col p-6 min-w-0">
            <span class="text-sm text-muted-foreground flex items-center gap-1 w-fit"><ChevronLeft class="w-4 h-4" /> Thư viện</span>
            <div class="flex-1 flex flex-col items-center justify-center">
              <div class="aspect-[3/4] w-40 rounded-xl shadow-2xl overflow-hidden"><WfBookCover title="Brand Marketing" author="Tác giả mẫu" size="lg" /></div>
              <h2 class="mt-5 text-lg font-semibold">Brand Marketing</h2>
              <p class="text-sm text-muted-foreground flex items-center gap-1">Tác giả mẫu · <Mic class="w-3.5 h-3.5" /> Giọng Hải Đăng</p>
              <p class="text-sm text-muted-foreground">Bốn áp lực buộc phải có thương hiệu</p>
              <div class="mt-5 w-full max-w-md">
                <div class="h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[38%] bg-primary rounded-full"></div></div>
                <div class="flex justify-between text-[11px] text-muted-foreground mt-1 tabular-nums"><span>0:18</span><span>-0:30</span></div>
              </div>
              <div class="mt-4 flex items-center gap-4">
                <span class="h-10 w-10 grid place-items-center rounded-full"><SkipBack class="w-5 h-5" /></span>
                <span class="h-10 w-10 grid place-items-center rounded-full"><RotateCcw class="w-5 h-5" /></span>
                <span class="h-14 w-14 grid place-items-center rounded-full bg-primary text-primary-foreground shadow"><Pause class="w-6 h-6" /></span>
                <span class="h-10 w-10 grid place-items-center rounded-full"><RotateCw class="w-5 h-5" /></span>
                <span class="h-10 w-10 grid place-items-center rounded-full"><SkipForward class="w-5 h-5" /></span>
              </div>
              <div class="mt-4 flex items-center gap-2 text-xs">
                <span class="h-8 px-3 rounded-full border border-border flex items-center gap-1.5"><Gauge class="w-3.5 h-3.5" />1,25×<ChevronDown class="w-3 h-3 text-muted-foreground" /></span>
                <!-- ─── MỚI: nút Nghỉ ─── -->
                <div class="relative">
                  <button class="h-8 px-3 rounded-full border flex items-center gap-1.5 hover:bg-muted"
                    :class="mode === 'menu' || mode === 'custom' ? 'border-foreground/40 bg-muted' : 'border-border'"
                    @click="mode = mode === 'menu' || mode === 'custom' ? (level === 'global' ? 'closed' : 'picked') : 'menu'">
                    <Timer class="w-3.5 h-3.5" />{{ pillLabel }}<ChevronDown class="w-3 h-3 text-muted-foreground" />
                  </button>
                  <div v-if="mode === 'menu' || mode === 'custom'" role="listbox" class="absolute bottom-full left-0 mb-2 w-72 rounded-lg border border-border bg-popover text-popover-foreground shadow-lg py-1 z-20">
                    <div class="px-3 pt-1.5 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Quãng nghỉ cho cuốn này</div>
                    <button v-for="l in levels" :key="l.key" role="option" :aria-selected="level === l.key"
                      class="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-muted"
                      :class="level === l.key && 'font-medium'" @click="pick(l.key)">
                      <span class="text-sm">{{ l.label }}<span class="block text-[11px] font-normal text-muted-foreground">{{ l.sub }}</span></span>
                      <Check v-if="level === l.key" class="w-4 h-4 shrink-0" />
                    </button>
                    <button role="option" :aria-selected="level === 'custom'"
                      class="w-full flex items-center justify-between gap-2 px-3 py-1.5 text-left hover:bg-muted"
                      :class="level === 'custom' && 'font-medium'" @click="level = 'custom'; mode = 'custom'">
                      <span class="text-sm">Tuỳ chỉnh<span class="block text-[11px] font-normal text-muted-foreground">{{ level === 'custom' ? customSub : 'Tự đặt số giây' }}</span></span>
                      <Check v-if="level === 'custom'" class="w-4 h-4 shrink-0" />
                    </button>
                    <div v-if="mode === 'custom'" class="mx-3 mb-1.5 mt-0.5 rounded-md bg-muted/60 p-2.5 space-y-2 text-sm">
                      <label class="flex items-center justify-between gap-2">Giữa tiểu mục
                        <span class="flex items-center gap-1.5 text-muted-foreground"><input v-model.number="customSection" type="number" min="0" max="10" step="0.1" class="w-16 h-8 rounded-md border border-input bg-background px-2 text-right text-foreground tabular-nums" /> giây</span>
                      </label>
                      <label class="flex items-center justify-between gap-2">Sang chương mới
                        <span class="flex items-center gap-1.5 text-muted-foreground"><input v-model.number="customChapter" type="number" min="0" max="10" step="0.1" class="w-16 h-8 rounded-md border border-input bg-background px-2 text-right text-foreground tabular-nums" /> giây</span>
                      </label>
                      <p class="text-[11px] text-muted-foreground">Từ 0 đến 10 giây, bước 0,1. Gõ xong là lưu.</p>
                    </div>
                    <div class="mt-1 border-t border-border px-3 pt-2 pb-1.5 text-[11px] text-muted-foreground leading-relaxed">Lưu riêng cho cuốn này, dùng cả khi Xuất M4B. Sách truyện hợp mức Dài, sách kiến thức hợp Vừa hoặc Ngắn.</div>
                  </div>
                </div>
                <span class="text-muted-foreground">Tua −15s / +30s · nhớ vị trí nghe</span>
              </div>
              <p v-if="mode === 'picked'" class="mt-3 text-xs text-muted-foreground flex items-center gap-1.5"><Check class="w-3.5 h-3.5" /> Cuốn này nghỉ 3 giây giữa tiểu mục, 4 giây sang chương. Áp dụng từ tiểu mục kế tiếp.</p>
            </div>
            <div class="flex flex-wrap items-center gap-2 border-t border-border pt-4">
              <Button variant="outline" size="sm"><Download class="w-4 h-4" /> Xuất M4B</Button>
              <Button variant="outline" size="sm"><Package class="w-4 h-4" /> Xuất gói zip</Button>
              <Button variant="ghost" size="sm"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
              <Button variant="ghost" size="sm" class="ml-auto text-destructive hover:text-destructive"><Trash2 class="w-4 h-4" /> Xoá</Button>
            </div>
          </div>
          <div class="w-72 shrink-0 border-l border-border overflow-auto">
            <div class="px-4 py-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Mục lục · 116 mục · 1 giờ 15 phút</div>
            <div v-for="c in toc" :key="c.t" class="w-full flex items-center justify-between gap-2 px-4 py-2.5 text-sm"
              :class="c.cur ? 'bg-primary/10 text-primary font-medium' : c.heard ? 'text-muted-foreground' : ''">
              <span class="truncate flex items-center gap-2">
                <Volume2 v-if="c.cur" class="w-3.5 h-3.5 shrink-0" /><Check v-else-if="c.heard" class="w-3.5 h-3.5 shrink-0" /><span v-else class="w-3.5 shrink-0"></span>{{ c.t }}
              </span>
              <span class="text-xs tabular-nums shrink-0">{{ c.d }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D9 · Màn nghe → nút "Nghỉ" cạnh tốc độ. Chọn mức riêng cho từng cuốn, lưu theo sách, dùng cả khi Xuất M4B. Chưa chọn thì theo Cài đặt chung (D8).
    </p>
  </div>
</template>
