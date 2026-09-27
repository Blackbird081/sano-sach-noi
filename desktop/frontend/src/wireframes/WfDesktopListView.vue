<script setup lang="ts">
// D6 — Xem thư viện dạng danh sách (bổ sung cho D2/D5). Bám khung D1 (1100×720).
// - Cạnh nút "Sắp xếp" ở hàng "Tất cả sách" có cặp nút Lưới / Danh sách. Bấm là đổi ngay.
// - Sano nhớ kiểu xem đã chọn, lần sau mở vẫn giữ.
// - Mỗi hàng: bìa nhỏ, tên sách, tác giả · giọng, danh mục, thời lượng, tiến độ, nút nghe, menu ⋯.
//   Bộ sách là một hàng, bìa xếp chồng + nhãn "Bộ N tập", bấm vào mở trang bộ sách như dạng lưới.
// - Sắp xếp ở dạng danh sách: kéo cả hàng lên xuống.
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { ref } from 'vue'
import {
  Library, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Search, ChevronDown, Play, Upload,
  MoreHorizontal, Pencil, Trash2, FolderOpen, Sun, Moon, Layers, GripVertical, ArrowUpDown,
  LayoutGrid, List, Mic, Settings2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import WfBookCover from './WfBookCover.vue'

type Mode = 'grid' | 'list' | 'arrange' | 'menu'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'list')
const dark = ref(false)

interface Item {
  id: string; kind: 'book' | 'series'; title: string; author: string; voice: string; cat: string
  dur: string; progress: number; note: string; vols?: number; listening?: boolean
}
const initial: Item[] = [
  { id: 's1', kind: 'series', title: 'Kinh doanh cho người mới', author: 'Nguyễn Văn A', voice: 'Hải Đăng', cat: 'Kinh doanh', dur: '4 giờ 10 phút', progress: 47, note: 'Đang nghe tập 2', vols: 3, listening: true },
  { id: 'b1', kind: 'book', title: 'Khởi nghiệp từ số 0', author: 'Trần Thị B', voice: 'Hải Đăng', cat: 'Kinh doanh', dur: '55 phút', progress: 0, note: 'Chưa nghe' },
  { id: 'b2', kind: 'book', title: 'Lãnh đạo cho quản lý mới', author: 'Nguyễn Văn A', voice: 'Ngọc Huyền', cat: 'Kỹ năng', dur: '1 giờ 14 phút', progress: 5, note: 'Đã nghe 5%' },
  { id: 'b3', kind: 'book', title: 'Tài chính cá nhân cơ bản', author: 'Lê Văn C', voice: 'Hải Đăng', cat: 'Tài chính', dur: '1 giờ 03 phút', progress: 100, note: 'Đã nghe xong' },
  { id: 'b4', kind: 'book', title: 'Tâm lý tích cực mỗi ngày: thói quen nhỏ, thay đổi lớn', author: 'Phạm Thị D', voice: 'Mỹ Duyên', cat: 'Kỹ năng', dur: '1 giờ 07 phút', progress: 21, note: 'Đã nghe 21%' },
  { id: 's2', kind: 'series', title: 'Tiếng Anh giao tiếp', author: 'Đỗ Thị G', voice: 'Ngọc Huyền', cat: 'Ngoại ngữ', dur: '2 giờ 30 phút', progress: 0, note: 'Chưa nghe', vols: 2 },
  { id: 'b5', kind: 'book', title: 'Quản lý thời gian hiệu quả', author: 'Nguyễn Văn A', voice: 'Thiện Minh', cat: 'Kỹ năng', dur: '2 giờ 10 phút', progress: 64, note: 'Đã nghe 64%' },
  { id: 'b6', kind: 'book', title: 'Ghi chép cuộc họp quý 3', author: '', voice: 'Hải Đăng', cat: '', dur: '34 phút', progress: 0, note: 'Chưa nghe' },
  { id: 'b7', kind: 'book', title: 'Đầu tư cho người mới', author: 'Hoàng Văn E', voice: 'Hải Đăng', cat: 'Tài chính', dur: '2 giờ 02 phút', progress: 9, note: 'Đã nghe 9%' },
  { id: 'b8', kind: 'book', title: 'Ngủ ngon mỗi đêm', author: 'Đỗ Thị G', voice: 'Mỹ Duyên', cat: 'Sức khoẻ', dur: '52 phút', progress: 0, note: 'Chưa nghe' },
]
const items = ref<Item[]>([...initial])
const filter = ref('all')

// Kéo hàng (chạy thật trong wireframe để cảm được thao tác)
const dragId = ref<string | null>(null)
const overId = ref<string | null>(null)
function onDrop(target: string) {
  const from = items.value.findIndex((i) => i.id === dragId.value)
  const to = items.value.findIndex((i) => i.id === target)
  if (from < 0 || to < 0 || from === to) return
  const list = [...items.value]
  const [m] = list.splice(from, 1)
  list.splice(to, 0, m)
  items.value = list
  dragId.value = overId.value = null
}

const view = ref<'grid' | 'list'>(mode.value === 'grid' ? 'grid' : 'list')
function setMode(m: Mode) {
  mode.value = m
  view.value = m === 'grid' ? 'grid' : 'list'
}
const arranging = () => mode.value === 'arrange'
const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách mới', icon: FilePlus2 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const continueList = initial.filter((i) => i.progress > 0 && i.progress < 100).slice(0, 3)
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <!-- Nút duyệt trạng thái (ngoài khung) -->
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['grid', '1. Dạng lưới (hiện tại) + nút chuyển'], ['list', '2. Dạng danh sách'],
        ['menu', '3. Danh sách · menu ⋯'], ['arrange', '4. Danh sách · Sắp xếp (kéo hàng)'],
      ] as [Mode, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="setMode(m[0])">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <!-- Khung cửa sổ -->
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
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.12 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <main class="flex-1 min-w-0 overflow-auto px-6 pt-6 pb-6">
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-xl font-semibold tracking-tight">Thư viện</h1>
              <p class="text-sm text-muted-foreground">13 cuốn · 2 bộ sách · 17 giờ 20 phút · lưu ở ~/Sano/Sach</p>
            </div>
            <div class="flex gap-2">
              <Button variant="outline"><Upload class="w-4 h-4" /> Nhập sách</Button>
              <Button><FilePlus2 class="w-4 h-4" /> Tạo sách mới</Button>
            </div>
          </div>

          <div class="mt-4 flex items-center gap-2">
            <div class="flex-1 flex items-center h-9 rounded-md border border-input bg-background px-3">
              <Search class="w-4 h-4 text-muted-foreground mr-2" />
              <input class="flex-1 bg-transparent outline-none text-sm" placeholder="Tìm theo tên sách, tác giả hoặc giọng đọc…" />
            </div>
            <button class="h-9 px-3 rounded-md border border-input bg-background text-sm flex items-center gap-1.5">
              <span class="text-muted-foreground">Sắp xếp:</span> {{ mode === 'arrange' ? 'Tự sắp xếp' : 'Mới tạo nhất' }} <ChevronDown class="w-4 h-4 opacity-60" />
            </button>
          </div>

          <div class="mt-3 flex flex-wrap items-center gap-1.5">
            <button v-for="c in ([['all', 'Tất cả · 13'], ['listening', 'Đang nghe · 5'], ['Kinh doanh', 'Kinh doanh · 4'], ['Kỹ năng', 'Kỹ năng · 3'], ['Tài chính', 'Tài chính · 2']] as string[][])" :key="c[0]"
              class="h-7 px-3 rounded-full text-xs border"
              :class="filter === c[0] ? 'border-primary bg-primary/10 text-primary font-medium' : 'border-border text-muted-foreground hover:text-foreground'"
              @click="filter = c[0]">{{ c[1] }}</button>
            <span class="h-7 px-2.5 text-xs text-muted-foreground flex items-center gap-1"><Settings2 class="w-3.5 h-3.5" /> Quản lý danh mục, bộ sách</span>
          </div>

          <!-- Nghe tiếp (giữ nguyên ở cả hai kiểu xem) -->
          <div v-if="mode !== 'arrange'" class="mt-5">
            <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Nghe tiếp</h2>
            <div class="mt-2 grid grid-cols-3 gap-3">
              <div v-for="b in continueList" :key="b.id" class="flex items-center gap-3 rounded-lg border border-border p-2.5">
                <div class="h-12 w-9 shrink-0 rounded shadow-sm overflow-hidden"><WfBookCover :title="b.title" :author="b.author" size="sm" /></div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium truncate">{{ b.title }}</p>
                  <p class="text-xs text-muted-foreground">{{ b.note }}</p>
                  <div class="mt-1 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
                </div>
                <span class="h-8 w-8 grid place-items-center rounded-full bg-primary text-primary-foreground"><Play class="w-4 h-4 ml-0.5" /></span>
              </div>
            </div>
          </div>

          <!-- Hàng tiêu đề: Sắp xếp + chuyển Lưới / Danh sách -->
          <div class="flex items-center justify-between" :class="mode === 'arrange' ? 'mt-4' : 'mt-6'">
            <h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Tất cả sách</h2>
            <div class="flex items-center gap-2">
              <button v-if="mode !== 'arrange'" class="h-7 px-2.5 rounded-md text-xs flex items-center gap-1.5 text-muted-foreground hover:bg-muted hover:text-foreground" @click="setMode('arrange')"><ArrowUpDown class="w-3.5 h-3.5" /> Sắp xếp</button>
              <Button v-else size="sm" @click="setMode('list')">Xong</Button>
              <div class="flex rounded-md border border-border p-0.5" role="radiogroup" aria-label="Kiểu xem">
                <button role="radio" :aria-checked="view === 'grid'" title="Xem dạng lưới" class="h-6 w-7 grid place-items-center rounded"
                  :class="view === 'grid' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'" @click="setMode('grid')"><LayoutGrid class="w-3.5 h-3.5" /></button>
                <button role="radio" :aria-checked="view === 'list'" title="Xem dạng danh sách" class="h-6 w-7 grid place-items-center rounded"
                  :class="view === 'list' ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'" @click="setMode('list')"><List class="w-3.5 h-3.5" /></button>
              </div>
            </div>
          </div>

          <div v-if="mode === 'arrange'" class="mt-3 flex items-center gap-2 rounded-lg bg-primary/5 border border-primary/20 px-3 py-2 text-sm">
            <GripVertical class="w-4 h-4 text-primary" />
            <span class="flex-1">Kéo hàng lên xuống để đổi chỗ. Thứ tự dùng chung cho cả dạng lưới.</span>
            <button class="text-xs text-muted-foreground hover:text-foreground" @click="items = [...initial]">Về thứ tự cũ</button>
          </div>

          <!-- ─── Dạng lưới (hiện tại) ─── -->
          <div v-if="view === 'grid'" class="mt-6 grid grid-cols-5 gap-x-5 gap-y-7">
            <div v-for="b in items" :key="b.id">
              <div class="relative">
                <template v-if="b.kind === 'series'">
                  <div class="absolute inset-0 translate-x-[10px] -translate-y-[10px] rounded-lg bg-slate-400 dark:bg-slate-600 border-2 border-background"></div>
                  <div class="absolute inset-0 translate-x-[5px] -translate-y-[5px] rounded-lg bg-slate-600 dark:bg-slate-400 border-2 border-background"></div>
                </template>
                <div class="relative aspect-[3/4] rounded-lg shadow-md overflow-hidden"><WfBookCover :title="b.title" :author="b.author" size="sm" />
                  <span v-if="b.kind === 'series'" class="absolute top-2 right-2 flex items-center gap-1 rounded-full bg-primary text-primary-foreground text-[11px] font-semibold px-2 py-0.5"><Layers class="w-3 h-3" /> Bộ {{ b.vols }} tập</span>
                </div>
              </div>
              <p class="mt-2 text-sm font-medium truncate">{{ b.title }}</p>
              <p class="text-xs text-muted-foreground truncate">{{ b.kind === 'series' ? `Bộ sách · ${b.vols} tập · ` : b.author ? b.author + ' · ' : '' }}{{ b.dur }}</p>
              <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
              <p class="mt-1 text-[11px] text-muted-foreground">{{ b.note }}</p>
            </div>
          </div>

          <!-- ─── Dạng danh sách ─── -->
          <div v-else class="mt-3 rounded-lg border border-border">
            <!-- tiêu đề cột -->
            <div class="flex items-center gap-4 px-3 h-8 text-[11px] font-medium uppercase tracking-wider text-muted-foreground border-b border-border bg-muted/30 rounded-t-lg">
              <span v-if="arranging()" class="w-4"></span>
              <span class="w-9"></span>
              <span class="flex-1">Tên sách</span>
              <span class="w-28">Danh mục</span>
              <span class="w-28 text-right">Thời lượng</span>
              <span class="w-36">Tiến độ</span>
              <span class="w-[68px]"></span>
            </div>
            <div v-for="(b, i) in items" :key="b.id"
              class="group relative flex items-center gap-4 px-3 py-2 hover:bg-muted/40"
              :class="[i < items.length - 1 && 'border-b border-border', arranging() && 'cursor-grab', dragId === b.id && 'opacity-40', overId === b.id && dragId !== b.id && 'bg-primary/5 shadow-[inset_0_2px_0_hsl(var(--primary))]']"
              :draggable="arranging()"
              @dragstart="dragId = b.id" @dragover.prevent="overId = b.id" @dragleave="overId = null" @drop="onDrop(b.id)" @dragend="dragId = overId = null">
              <GripVertical v-if="arranging()" class="w-4 h-4 text-muted-foreground shrink-0" />
              <!-- bìa nhỏ; bộ sách có bìa xếp chồng -->
              <div class="relative h-12 w-9 shrink-0">
                <template v-if="b.kind === 'series'">
                  <div class="absolute inset-0 translate-x-[4px] -translate-y-[4px] rounded bg-slate-400 dark:bg-slate-600 border border-background"></div>
                  <div class="absolute inset-0 translate-x-[2px] -translate-y-[2px] rounded bg-slate-600 dark:bg-slate-400 border border-background"></div>
                </template>
                <div class="relative h-full w-full rounded shadow-sm overflow-hidden"><WfBookCover :title="b.title" :author="b.author" size="sm" /></div>
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate flex items-center gap-1.5">
                  <span class="truncate" :title="b.title">{{ b.title }}</span>
                  <span v-if="b.kind === 'series'" class="shrink-0 flex items-center gap-1 rounded-full bg-primary/10 text-primary text-[11px] font-semibold px-1.5 py-px"><Layers class="w-3 h-3" /> Bộ {{ b.vols }} tập</span>
                </p>
                <p class="text-xs text-muted-foreground truncate flex items-center gap-1">
                  <template v-if="b.author">{{ b.author }} ·</template>
                  <Mic class="w-3 h-3 shrink-0" /> {{ b.voice }}
                </p>
              </div>
              <span class="w-28 text-xs text-muted-foreground truncate">{{ b.cat || '—' }}</span>
              <span class="w-28 text-right text-xs text-muted-foreground tabular-nums">{{ b.dur }}</span>
              <div class="w-36">
                <div class="h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: b.progress + '%' }"></div></div>
                <p class="mt-1 text-[11px] text-muted-foreground truncate">{{ b.note }}</p>
              </div>
              <div class="w-[68px] flex items-center justify-end gap-1">
                <template v-if="!arranging()">
                  <span class="h-8 w-8 grid place-items-center rounded-full"
                    :class="b.listening ? 'bg-primary text-primary-foreground' : 'border border-border text-foreground opacity-0 group-hover:opacity-100'"
                    :title="b.kind === 'series' ? 'Nghe tiếp bộ sách' : 'Nghe'"><Play class="w-4 h-4 ml-0.5" /></span>
                  <button v-if="b.kind === 'book'" class="h-7 w-7 grid place-items-center rounded-md text-muted-foreground hover:bg-muted"
                    :class="mode === 'menu' && b.id === 'b2' ? 'bg-muted text-foreground' : 'opacity-0 group-hover:opacity-100'"><MoreHorizontal class="w-4 h-4" /></button>
                  <span v-else class="w-7"></span>
                </template>
              </div>
              <!-- menu ⋯ (giống menu trên bìa ở dạng lưới) -->
              <div v-if="mode === 'menu' && b.id === 'b2'" class="absolute top-11 right-3 w-48 rounded-lg border border-border bg-popover shadow-lg py-1 z-20 text-sm">
                <span class="w-full flex items-center gap-2 px-3 py-1.5 hover:bg-muted"><Pencil class="w-4 h-4" /> Sửa thông tin</span>
                <span class="w-full flex items-center gap-2 px-3 py-1.5 hover:bg-muted"><FolderOpen class="w-4 h-4" /> Mở thư mục</span>
                <span class="w-full flex items-center gap-2 px-3 py-1.5 hover:bg-muted text-destructive"><Trash2 class="w-4 h-4" /> Chuyển vào Thùng rác</span>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D6 · Xem thư viện dạng danh sách. Cặp nút Lưới / Danh sách nằm cạnh "Sắp xếp" ở hàng "Tất cả sách", Sano nhớ kiểu xem đã chọn.
      Bấm một hàng để nghe (bộ sách thì mở trang bộ). Nút nghe và ⋯ hiện khi rê chuột, cuốn đang nghe luôn hiện nút đỏ. Thử kéo hàng ở trạng thái 4.
    </p>
  </div>
</template>
