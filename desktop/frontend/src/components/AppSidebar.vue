<script setup lang="ts">
import { Library, BarChart3, FilePlus2, Settings, Info, Loader2, ArrowUpCircle, LifeBuoy, ExternalLink, RefreshCw, Bot, X, Cable, CheckCircle2 } from 'lucide-vue-next'
import faviconUrl from '@/assets/favicon.svg'
import sepayLogo from '../assets/sponsors/sepay.svg'
import hostLogo from '../assets/sponsors/123host.svg'
import { AUTHOR_FB, DOCS } from '../lib/mock'
import { go, openBook, remainMin, renderPct, rendering, state, type View } from '../lib/store'
import { computed } from 'vue'
import { activeJobs, dismissJobs, mcp, stagedJobs } from '../lib/mcp'
import { edit, fmtRemain, openEdit } from '../lib/edit'

// Thẻ sách AI tạo (D22)
const currentJob = computed(() => mcp.jobs.find((j) => j.status === 'rendering'))
const savedJobs = computed(() => mcp.jobs.filter((j) => j.status === 'saved'))
const failedJobs = computed(() => mcp.jobs.filter((j) => j.status === 'failed'))
const doneCount = computed(() => mcp.jobs.filter((j) => j.status !== 'queued' && j.status !== 'rendering').length)
const jobClient = computed(() => [...new Set(mcp.jobs.map((j) => j.client || 'AI'))].join(', '))

const nav = [
  { key: 'library' as View, label: 'Thư viện', icon: Library },
  { key: 'create' as View, label: 'Tạo sách nói', icon: FilePlus2 },
  { key: 'stats' as View, label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'mcp' as View, label: 'MCP', icon: Cable },
  { key: 'settings' as View, label: 'Cài đặt', icon: Settings },
  { key: 'about' as View, label: 'Giới thiệu', icon: Info },
]

// Phần mềm miễn phí, tài trợ phát triển bởi (nhỏ, cuối thanh bên). utm để biết lượt ghé đến từ app.
const utm = '?utm_source=sano&utm_medium=app-sidebar&utm_campaign=tai-tro'
const sponsors = [
  { name: 'SePay', logo: sepayLogo, url: 'https://sepay.vn' + utm },
  { name: '123HOST', logo: hostLogo, url: 'https://123host.vn' + utm },
]
</script>

<template>
  <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
    <div class="px-4 py-4 flex items-center gap-2">
      <img :src="faviconUrl" alt="" class="h-7 w-7 rounded-md" />
      <span class="font-semibold tracking-tight">Sano</span>
    </div>
    <nav class="px-2 space-y-0.5">
      <button
        v-for="n in nav" :key="n.key"
        class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm"
        :class="state.view === n.key || (n.key === 'library' && (state.view === 'player' || state.view === 'edit')) ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
        @click="go(n.key)"
      >
        <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
      </button>
    </nav>

    <div class="px-2 mt-2">
      <a :href="DOCS" target="_blank" rel="noopener" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground hover:bg-muted hover:text-foreground">
        <LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" />
      </a>
    </div>

    <div class="flex-1"></div>

    <!-- Thẻ tiến độ render (hiện ở mọi màn khi đang render) -->
    <button v-if="rendering && state.view !== 'create'" class="m-3 rounded-lg border border-border bg-background p-3 text-left hover:border-primary/50" @click="go('create')">
      <div class="flex items-center gap-1.5 text-xs font-medium"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" /> Đang render</div>
      <p class="mt-1 text-xs text-muted-foreground truncate">{{ state.render?.title }}</p>
      <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full" :style="{ width: renderPct + '%' }"></div></div>
      <div class="mt-1 flex justify-between text-[11px] text-muted-foreground tabular-nums"><span>{{ renderPct }}%</span><span>còn ~{{ remainMin }} phút</span></div>
    </button>
    <!-- Thẻ sách AI nhờ tạo qua MCP (D22): đang tạo / chờ cam kết / đã lưu -->
    <div v-if="mcp.jobs.length" class="mx-3 mb-3 rounded-lg border border-border bg-background p-2.5 text-xs">
      <template v-if="activeJobs.length || stagedJobs.length">
        <div class="flex items-center gap-1.5 font-medium">
          <Loader2 v-if="activeJobs.length" class="w-3.5 h-3.5 animate-spin text-primary shrink-0" />
          <Bot v-else class="w-3.5 h-3.5 text-primary shrink-0" />
          <span class="truncate">{{ jobClient }} {{ activeJobs.length ? 'nhờ tạo' : 'tạo xong' }} {{ activeJobs.length ? mcp.jobs.length : stagedJobs.length }} cuốn</span>
        </div>
        <template v-if="currentJob">
          <p class="mt-1 text-muted-foreground truncate" :title="currentJob.title">Đang tạo {{ doneCount + 1 }}/{{ mcp.jobs.length }} · {{ currentJob.title }}</p>
          <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: currentJob.percent + '%' }"></div></div>
        </template>
        <button v-if="stagedJobs.length" class="mt-2 w-full h-7 rounded-md bg-primary/10 text-primary font-medium hover:bg-primary/15" @click="mcp.popup = true">
          {{ stagedJobs.length }} cuốn chờ cam kết · Mở</button>
      </template>
      <template v-if="savedJobs.length && !activeJobs.length">
        <div class="flex items-center gap-1.5 font-medium" :class="stagedJobs.length ? 'mt-2.5' : ''">
          <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600 shrink-0" /> <span class="truncate">Đã lưu {{ savedJobs.length }} cuốn vào Thư viện</span>
          <button class="ml-auto text-muted-foreground hover:text-foreground" aria-label="Ẩn" @click="dismissJobs()"><X class="w-3.5 h-3.5" /></button>
        </div>
        <button v-if="savedJobs.length === 1 && savedJobs[0].slug" class="mt-1 font-medium text-primary hover:underline truncate max-w-full" @click="openBook(savedJobs[0].slug!)">Nghe {{ savedJobs[0].title }}</button>
        <button v-else class="mt-1 font-medium text-primary hover:underline" @click="go('library')">Xem trong Thư viện</button>
        <p v-if="stagedJobs.length" class="mt-1 text-muted-foreground">{{ stagedJobs.length }} cuốn vẫn ở khu chờ, tự xoá sau 7 ngày nếu không cam kết.</p>
      </template>
      <p v-for="j in failedJobs" :key="j.id" class="mt-1.5 text-destructive line-clamp-2" :title="j.error">Lỗi tạo {{ j.title }}: {{ j.error }}</p>
      <button v-if="failedJobs.length && !activeJobs.length && !savedJobs.length" class="mt-1 text-muted-foreground hover:text-foreground" @click="dismissJobs()">Ẩn</button>
    </div>
    <p v-if="mcp.error" class="mx-3 mb-3 text-xs text-destructive">{{ mcp.error }}</p>
    <!-- Thẻ tiến độ sửa sách: đổi giọng / đọc lại chạy nền (wireframe D11) -->
    <button v-if="edit.status?.running && !(state.view === 'edit' && edit.slug === edit.status.slug)" class="mx-3 mb-3 rounded-lg border border-border bg-background p-2.5 text-left text-xs hover:border-primary/50"
      @click="openEdit(edit.status.slug, { tab: edit.status.kind === 'voice' ? 'voice' : 'content' })">
      <div class="flex items-center gap-1.5 font-medium"><Loader2 class="w-3.5 h-3.5 animate-spin text-primary" /> <span class="truncate">{{ edit.status.kind === 'voice' ? 'Đổi giọng' : 'Đọc lại' }} {{ edit.status.bookTitle }}</span></div>
      <div class="mt-1.5 h-1 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary" :style="{ width: (edit.status.total ? Math.round((edit.status.finished / edit.status.total) * 100) : 0) + '%' }"></div></div>
      <div class="mt-1 text-muted-foreground tabular-nums">{{ edit.status.finished }}/{{ edit.status.total }} mục<template v-if="edit.status.remainSec"> · {{ fmtRemain(edit.status.remainSec) }}</template></div>
    </button>
    <!-- Có bản mới: thẻ nổi bật, nền đỏ nhạt -->
    <button v-if="state.updateInfo" class="mx-3 mb-3 rounded-lg border border-primary/30 bg-primary/10 text-primary p-3 text-left hover:bg-primary/15 transition" @click="state.update = 'info'">
      <span class="flex items-center gap-2 text-sm font-semibold">
        <RefreshCw v-if="state.upd.applyOnQuit" class="w-5 h-5 shrink-0" /><ArrowUpCircle v-else class="w-5 h-5 shrink-0" />
        {{ state.upd.applyOnQuit ? 'Khởi động lại để cập nhật' : `Có bản mới ${state.updateInfo.version}` }}
        <span v-if="!state.upd.applyOnQuit" class="ml-auto h-2 w-2 rounded-full bg-primary animate-pulse" aria-hidden="true"></span>
      </span>
      <span class="mt-1 block text-xs text-foreground/70">{{ state.upd.applyOnQuit ? 'Bản mới đã tải xong, mở lại Sano là dùng được' : 'Bấm để xem có gì mới và cập nhật' }}</span>
    </button>
    <div class="px-4 pb-3">
      <p class="text-center text-[10px] font-medium uppercase tracking-wider leading-snug text-muted-foreground/80">Phần mềm miễn phí,<br />tài trợ phát triển bởi</p>
      <div class="mt-1.5 grid grid-cols-2 gap-1.5">
        <a v-for="sp in sponsors" :key="sp.name" :href="sp.url" target="_blank" rel="noopener" :title="sp.name" :aria-label="`Nhà tài trợ ${sp.name}`"
          class="h-8 rounded-md bg-white border border-border grid place-items-center px-2 opacity-80 hover:opacity-100 hover:border-primary/40 transition">
          <img :src="sp.logo" :alt="sp.name" class="max-h-5 max-w-full object-contain" />
        </a>
      </div>
    </div>
    <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">
      Phiên bản {{ state.version || '…' }} · mã nguồn mở<br />
      Tác giả <a :href="AUTHOR_FB" target="_blank" rel="noopener" class="hover:text-foreground underline-offset-2 hover:underline">Bùi Tấn Việt</a>
    </div>
  </aside>
</template>
