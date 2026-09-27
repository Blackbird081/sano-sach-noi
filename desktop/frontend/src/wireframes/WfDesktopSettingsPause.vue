<script setup lang="ts">
// D8 — Cài đặt: quãng nghỉ giữa các phần. Bám khung D1 (1100×720) và màn Cài đặt thật.
// Nhóm mới "Nghe và xuất M4B" (dưới Lưu trữ): 3 mức Ngắn / Vừa / Dài, dùng chung cho
// trình phát trong Sano và file M4B.
// - Ngắn: 0,5 giây giữa tiểu mục, 1 giây sang chương
// - Vừa (mặc định): 1,5 giây / 2 giây
// - Dài: 3 giây / 4 giây
// Đổi mức: có hiệu lực ngay khi nghe; M4B đã xuất trước đó không đổi → dòng nhắc xuất lại.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, Trash2, Check,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type Level = 'short' | 'medium' | 'long' | 'custom'
type Mode = 'default' | 'changed' | 'custom'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'default')
const dark = ref(false)
const level = ref<Level>(mode.value === 'changed' ? 'long' : mode.value === 'custom' ? 'custom' : 'medium')
const changed = computed(() => mode.value !== 'default')
const customSection = ref(1.2)
const customChapter = ref(2.5)
const fmtSec = (n: number) => n.toLocaleString('vi-VN') + ' giây'

const levels: { key: Level; label: string; section: string; chapter: string }[] = [
  { key: 'short', label: 'Ngắn', section: '0,5 giây', chapter: '1 giây' },
  { key: 'medium', label: 'Vừa', section: '1,5 giây', chapter: '2 giây' },
  { key: 'long', label: 'Dài', section: '3 giây', chapter: '4 giây' },
]
const cur = computed(() => levels.find((l) => l.key === level.value) ?? { key: 'custom', label: 'Tuỳ chỉnh', section: fmtSec(customSection.value), chapter: fmtSec(customChapter.value) })

function pick(l: Level) {
  level.value = l
  mode.value = l === 'medium' ? 'default' : l === 'custom' ? 'custom' : 'changed'
}
function setMode(m: Mode) {
  mode.value = m
  level.value = m === 'changed' ? 'long' : m === 'custom' ? 'custom' : 'medium'
}

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
      <button v-for="m in ([['default', '1. Mặc định (Vừa)'], ['changed', '2. Vừa đổi sang Dài'], ['custom', '3. Tuỳ chỉnh']] as [Mode, string][])" :key="m[0]"
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
              :class="n.key === 'settings' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2">
            <span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span>
          </div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.16 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <main class="flex-1 min-w-0 overflow-auto p-6">
          <div class="max-w-2xl">
            <h1 class="text-xl font-semibold tracking-tight">Cài đặt</h1>
            <div class="mt-6 space-y-6">
              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Lưu trữ</h2>
                <div class="rounded-lg border border-border divide-y divide-border text-sm">
                  <div class="flex items-center justify-between gap-4 px-4 py-3"><span>Thư mục lưu sách</span><span class="flex items-center gap-2 text-muted-foreground">~/Sano/Sach <Button variant="outline" size="sm">Mở</Button></span></div>
                  <div class="flex items-center justify-between px-4 py-3"><span>Dung lượng sách đã tạo</span><span class="text-muted-foreground tabular-nums">2,4 GB</span></div>
                </div>
              </div>

              <!-- ─── MỚI: quãng nghỉ ─── -->
              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Nghe và xuất M4B</h2>
                <div class="rounded-lg border border-border text-sm">
                  <div class="flex items-center justify-between gap-4 px-4 py-3">
                    <span>Quãng nghỉ giữa các phần
                      <span class="block text-xs text-muted-foreground">Nghỉ {{ cur.section }} giữa các tiểu mục, {{ cur.chapter }} khi sang chương mới. Mức mặc định cho mọi cuốn, dùng cả khi nghe trong Sano và Xuất M4B. Muốn khác cho từng cuốn: chỉnh ở nút Nghỉ trong màn nghe.</span>
                    </span>
                    <div role="radiogroup" class="shrink-0 inline-flex rounded-md bg-muted p-0.5">
                      <button v-for="l in levels" :key="l.key" role="radio" :aria-checked="level === l.key"
                        class="h-8 px-3 rounded text-sm"
                        :class="level === l.key ? 'bg-background text-foreground shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'"
                        @click="pick(l.key)">{{ l.label }}</button>
                      <button role="radio" :aria-checked="level === 'custom'" class="h-8 px-3 rounded text-sm"
                        :class="level === 'custom' ? 'bg-background text-foreground shadow-sm font-medium' : 'text-muted-foreground hover:text-foreground'"
                        @click="pick('custom')">Tuỳ chỉnh</button>
                    </div>
                  </div>
                  <div v-if="level === 'custom'" class="border-t border-border px-4 py-3 flex flex-wrap items-center gap-x-6 gap-y-2">
                    <label class="flex items-center gap-2">Giữa tiểu mục
                      <input v-model.number="customSection" type="number" min="0" max="10" step="0.1" class="w-16 h-8 rounded-md border border-input bg-background px-2 text-right tabular-nums" /><span class="text-muted-foreground">giây</span>
                    </label>
                    <label class="flex items-center gap-2">Sang chương mới
                      <input v-model.number="customChapter" type="number" min="0" max="10" step="0.1" class="w-16 h-8 rounded-md border border-input bg-background px-2 text-right tabular-nums" /><span class="text-muted-foreground">giây</span>
                    </label>
                    <span class="text-xs text-muted-foreground">Từ 0 đến 10 giây, bước 0,1</span>
                  </div>
                  <div v-if="changed" class="border-t border-border px-4 py-2.5 text-xs text-muted-foreground flex items-start gap-2">
                    <Check class="w-3.5 h-3.5 mt-0.5 shrink-0" />
                    <span>Đã lưu. Áp dụng ngay khi nghe trong Sano. Cuốn nào đã chọn riêng ở màn nghe thì giữ mức riêng. File M4B đã xuất trước đây giữ quãng nghỉ cũ, muốn đổi thì xuất lại.</span>
                  </div>
                </div>
              </div>

              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Dữ liệu nghe</h2>
                <div class="rounded-lg border border-border text-sm">
                  <div class="flex items-center justify-between gap-4 px-4 py-3">
                    <span>Số liệu Hành trình nghe
                      <span class="block text-xs text-muted-foreground">Đã ghi 12 ngày nghe · chỉ lưu trên máy, trong ~/Sano/.nghe.json. Xoá thì sách và vị trí đang nghe vẫn giữ.</span>
                    </span>
                    <Button variant="outline" size="sm" class="shrink-0"><Trash2 class="w-3.5 h-3.5" /> Xoá số liệu</Button>
                  </div>
                </div>
              </div>

              <div>
                <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-2">Bộ đọc</h2>
                <div class="rounded-lg border border-border px-4 py-3 text-sm text-muted-foreground">(giữ nguyên như hiện tại)</div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D8 · Cài đặt → nhóm mới "Nghe và xuất M4B". Bấm Ngắn / Vừa / Dài là lưu ngay, dòng mô tả đổi số giây theo mức. Mặc định Vừa (như bản 0.1.15).
    </p>
  </div>
</template>
