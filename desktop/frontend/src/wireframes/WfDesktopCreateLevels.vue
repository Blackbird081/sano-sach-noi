<script setup lang="ts">
// D8 (bản 2) — Tạo sách nói: bước đầu "Cách đọc", mỗi màn một việc.
// Bám khung D1 (1100×720). Thêm bước 1 "Cách đọc" trước "Nạp file" (7 bước).
// Màn A — Chọn cách làm: 3 lựa chọn xếp dọc, có nút chọn tròn. Mỗi dòng: Sano làm gì,
//   "Hợp khi", thời gian chuẩn bị thêm, nút "Nghe mẫu" (kết quả của cấp đó, âm thanh
//   render sẵn nhúng trong app). Gợi ý "Không chắc thì chọn…". Sano nhớ lựa chọn.
// Màn B — Nhờ AI viết lại (cấp 2/3): 3 bước đánh số: sao chép prompt → mở AI, gửi
//   file + prompt → tải file Word AI tạo về, nạp vào Sano. Prompt dặn AI ưu tiên trả
//   file .docx (Heading 1 chương, Heading 2 mục); không tạo được file thì trả văn bản
//   # / ## để dán vào Sano (dự phòng). Cấp 3 có prompt soát lại (không bắt buộc).
//   Góc dưới: "Làm thường xuyên? Nạp skill cho AI" (Claude / ChatGPT / Gemini).
// Cấp 1: bỏ qua màn B, sang thẳng Nạp file. Nạp file: file Word là chính; link nhỏ "AI không tạo được file? Dán văn bản".
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, FilePlus2, BarChart3, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, Check, ChevronRight, ChevronLeft,
  Copy, Download, Play, Pause, Upload, ClipboardPaste, FileText, X, ShieldCheck, Lightbulb, Clock, Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type Mode = 'choose' | 'chosen' | 'ai3' | 'ai2' | 'skill' | 'paste' | 'pastetext' | 'file1'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'choose')
const dark = ref(false)
const picked = ref<number>(mode.value === 'choose' ? 0 : mode.value === 'ai2' ? 2 : 3)
const step = computed(() => (mode.value === 'paste' || mode.value === 'pastetext' || mode.value === 'file1' ? 2 : 1))
const copied = ref<number | null>(null)
const playing = ref<number | null>(null)
const skillTab = ref<'claude' | 'chatgpt' | 'gemini'>('claude')
const pasteTab = ref<'file' | 'paste'>('file')

function setMode(m: Mode) {
  if (m === 'file1') {
    picked.value = 1
    pasteTab.value = 'file'
    mode.value = 'paste'
    return
  }
  if (m === 'pastetext') {
    mode.value = 'paste'
    pasteTab.value = 'paste'
    return
  }
  if (m === 'paste') {
    pasteTab.value = 'file'
    if (picked.value < 2) picked.value = 3
  }
  mode.value = m
  if (m === 'choose') picked.value = 0
  else if (m === 'ai2') picked.value = 2
  else if (m !== 'paste' && m !== 'skill' && m !== 'chosen') picked.value = 3
  else if (m === 'chosen' && !picked.value) picked.value = 3
}
function copy(n: number) {
  copied.value = n
  setTimeout(() => (copied.value = null), 1500)
}

const steps = ['Cách đọc', 'Nạp file', 'Mục lục', 'Giọng đọc', 'Lời mở đầu', 'Nghe thử', 'Render']
const options = [
  { n: 1, title: 'Đọc nguyên văn', does: 'Sano đọc đúng từng chữ trong file Word.', fit: 'bài viết, ghi chép đã dễ đọc; hoặc bạn muốn nghe y nguyên', time: 'Không cần chuẩn bị', ai: false, dur: '0:09', best: false },
  { n: 2, title: 'Làm mượt', does: 'Nhờ AI đổi bảng, hình, danh sách, chữ viết tắt thành lời. Giữ nguyên ý và giọng tác giả.', fit: 'tài liệu nhiều bảng biểu, gạch đầu dòng', time: 'Thêm khoảng 5 phút với AI', ai: true, dur: '0:10', best: false },
  { n: 3, title: 'Viết lại thành văn sách nói', does: 'Nhờ AI viết lại như người kể: chuyện trước, lý thuyết sau, chương ngắn, cuối chương có ba ý cần nhớ.', fit: 'sách, giáo trình, tài liệu dài muốn nghe cuốn như sách nói', time: 'Thêm khoảng 10–15 phút với AI', ai: true, dur: '0:12', best: true },
]
const opt = computed(() => options.find((o) => o.n === picked.value))
const ais = [
  { k: 'claude' as const, name: 'Claude', file: 'sano-sach-noi.zip', steps: ['Tải file sano-sach-noi.zip (không cần giải nén).', 'Mở claude.ai → Cài đặt → Capabilities → Skills → Tải lên.', 'Từ nay chỉ cần đính kèm file Word và gõ: "Làm sách nói cấp 3 cho file này".'] },
  { k: 'chatgpt' as const, name: 'ChatGPT', file: 'sano-huong-dan-ai.txt', steps: ['Tải file hướng dẫn sano-huong-dan-ai.txt.', 'Tạo một Dự án (Project) hoặc GPT riêng tên "Sano – sách nói", dán nội dung file vào ô Hướng dẫn (Instructions).', 'Mỗi lần làm sách: mở dự án đó, đính kèm file Word, gõ "cấp 2" hoặc "cấp 3".'] },
  { k: 'gemini' as const, name: 'Gemini', file: 'sano-huong-dan-ai.txt', steps: ['Tải file hướng dẫn sano-huong-dan-ai.txt.', 'Mở Gemini → Gem → Tạo Gem mới tên "Sano – sách nói", dán nội dung file vào ô Hướng dẫn.', 'Mỗi lần làm sách: mở Gem đó, đính kèm file Word, gõ "cấp 2" hoặc "cấp 3".'] },
]
const ai = computed(() => ais.find((a) => a.k === skillTab.value)!)
const nav = [
  { key: 'library', label: 'Thư viện', icon: Library },
  { key: 'create', label: 'Tạo sách nói', icon: FilePlus2 },
  { key: 'stats', label: 'Hành trình nghe', icon: BarChart3 },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
  { key: 'about', label: 'Giới thiệu', icon: Info },
]
const onAI = computed(() => mode.value === 'ai3' || mode.value === 'ai2' || mode.value === 'skill')
function next() {
  if (mode.value === 'chosen' && picked.value === 1) {
    pasteTab.value = 'file'
    mode.value = 'paste'
  } else if (mode.value === 'chosen') setMode(picked.value === 2 ? 'ai2' : 'ai3')
  else if (onAI.value) mode.value = 'paste'
}
function back() {
  if (mode.value === 'paste') mode.value = picked.value === 1 ? 'chosen' : picked.value === 2 ? 'ai2' : 'ai3'
  else mode.value = 'chosen'
}
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([
        ['choose', '1. Chọn cách làm'], ['chosen', '2. Đã chọn cấp 3'], ['ai3', '3. Nhờ AI (cấp 3)'], ['ai2', '4. Nhờ AI (cấp 2)'],
        ['skill', '5. Nạp skill cho AI'], ['paste', '6. Nạp file AI tạo'], ['pastetext', '7. Dự phòng: dán văn bản'], ['file1', '8. Cấp 1: nạp file Word'],
      ] as [Mode, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="mode === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="setMode(m[0])">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs flex items-center gap-1.5 text-foreground" @click="dark = !dark">
        <component :is="dark ? Sun : Moon" class="w-3.5 h-3.5" /> {{ dark ? 'Sáng' : 'Tối' }}
      </button>
    </div>

    <div class="relative w-[1100px] h-[720px] rounded-xl overflow-hidden border border-border shadow-2xl bg-background text-foreground flex flex-col">
      <div class="h-9 shrink-0 flex items-center gap-2 px-3 border-b border-border bg-muted/50">
        <span class="h-3 w-3 rounded-full bg-rag-red"></span><span class="h-3 w-3 rounded-full bg-rag-amber"></span><span class="h-3 w-3 rounded-full bg-rag-green"></span>
        <span class="flex-1 text-center text-xs text-muted-foreground">Sano</span>
      </div>
      <div class="flex-1 flex min-h-0">
        <aside class="w-56 shrink-0 border-r border-border bg-muted/30 flex flex-col">
          <div class="px-4 py-4 flex items-center gap-2"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /><span class="font-semibold tracking-tight">Sano</span></div>
          <nav class="px-2 space-y-0.5">
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm" :class="n.key === 'create' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2"><span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span></div>
          <div class="flex-1"></div>
          <div class="px-4 pb-3 text-[11px] text-muted-foreground leading-relaxed">Phiên bản 0.1.14 · mã nguồn mở<br />Tác giả Bùi Tấn Việt</div>
        </aside>

        <section class="flex-1 min-w-0 flex flex-col">
          <div class="shrink-0 border-b border-border px-5 h-14 flex items-center gap-0.5">
            <template v-for="(s, i) in steps" :key="s">
              <span class="flex items-center gap-1.5 px-1.5 h-8 rounded-md text-sm whitespace-nowrap" :class="step === i + 1 ? 'text-foreground font-medium' : 'text-muted-foreground'">
                <span class="h-6 w-6 rounded-full grid place-items-center text-xs" :class="step > i + 1 ? 'bg-primary/15 text-primary' : step === i + 1 ? 'bg-primary text-primary-foreground' : 'bg-muted'">
                  <Check v-if="step > i + 1" class="w-3.5 h-3.5" /><template v-else>{{ i + 1 }}</template>
                </span>{{ s }}
              </span>
              <ChevronRight v-if="i < steps.length - 1" class="w-3.5 h-3.5 text-muted-foreground/50 shrink-0" />
            </template>
          </div>

          <div class="flex-1 overflow-auto p-6">
            <!-- ═══ Màn A: chọn cách làm ═══ -->
            <div v-if="mode === 'choose' || mode === 'chosen'" class="max-w-3xl">
              <h1 class="text-xl font-semibold tracking-tight">Chọn cách làm sách nói</h1>
              <p class="text-sm text-muted-foreground">Văn viết để đọc bằng mắt, đọc to lên thường nghe chán. Bấm <b class="font-medium text-foreground">Nghe mẫu</b> ở từng cách để nghe khác biệt, rồi chọn một cách.</p>

              <div class="mt-4 space-y-2.5" role="radiogroup" aria-label="Cách làm sách nói">
                <div v-for="o in options" :key="o.n" role="radio" :aria-checked="picked === o.n"
                  class="relative flex items-start gap-4 rounded-xl border p-4 cursor-pointer transition"
                  :class="picked === o.n ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40 hover:bg-muted/30'"
                  @click="picked = o.n; mode = 'chosen'">
                  <span class="mt-0.5 h-5 w-5 rounded-full border-2 grid place-items-center shrink-0" :class="picked === o.n ? 'border-primary' : 'border-muted-foreground/40'">
                    <span v-if="picked === o.n" class="h-2.5 w-2.5 rounded-full bg-primary"></span>
                  </span>
                  <div class="flex-1 min-w-0">
                    <p class="font-medium flex items-center gap-2">
                      <span class="text-xs font-normal text-muted-foreground">Cấp {{ o.n }}</span> {{ o.title }}
                      <span v-if="o.best" class="rounded-full bg-chart/15 text-chart text-[11px] font-semibold px-2 py-0.5">Nghe hấp dẫn nhất</span>
                    </p>
                    <p class="mt-0.5 text-sm text-muted-foreground">{{ o.does }}</p>
                    <p class="mt-1.5 text-xs text-muted-foreground flex flex-wrap gap-x-4 gap-y-1">
                      <span><span class="text-foreground">Hợp khi:</span> {{ o.fit }}</span>
                      <span class="flex items-center gap-1"><Clock class="w-3 h-3" /> {{ o.time }}</span>
                    </p>
                  </div>
                  <button class="shrink-0 h-8 pl-2.5 pr-3 rounded-full border text-xs flex items-center gap-1.5"
                    :class="playing === o.n ? 'border-chart bg-chart text-white' : 'border-border bg-background hover:bg-muted'"
                    @click.stop="playing = playing === o.n ? null : o.n">
                    <component :is="playing === o.n ? Pause : Play" class="w-3.5 h-3.5" /> {{ playing === o.n ? 'Dừng' : 'Nghe mẫu' }} <span :class="playing === o.n ? 'opacity-80' : 'text-muted-foreground'">· {{ o.dur }}</span>
                  </button>
                </div>
              </div>

              <p class="mt-4 text-sm text-muted-foreground flex items-start gap-2">
                <Lightbulb class="w-4 h-4 mt-0.5 text-rag-amber shrink-0" />
                <span>Không chắc? Sách hay tài liệu dài thì chọn <b class="font-medium text-foreground">cấp 3</b>. Bài ngắn, đọc đã trôi chảy thì chọn <b class="font-medium text-foreground">cấp 1</b>. Lần sau Sano nhớ lựa chọn của bạn.</span>
              </p>
            </div>

            <!-- ═══ Màn B: nhờ AI viết lại ═══ -->
            <div v-else-if="onAI" class="max-w-3xl">
              <button class="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1" @click="mode = 'chosen'"><ChevronLeft class="w-4 h-4" /> Đổi cách làm</button>
              <h1 class="mt-2 text-xl font-semibold tracking-tight">Nhờ AI {{ picked === 3 ? 'viết lại thành văn sách nói' : 'làm mượt tài liệu' }}</h1>
              <p class="text-sm text-muted-foreground">Dùng ChatGPT, Gemini hoặc Claude, bản miễn phí cũng được. Làm xong 3 bước là có bản để Sano đọc.</p>

              <ol class="mt-5 space-y-3">
                <li class="flex gap-4 rounded-xl border border-border p-4">
                  <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">1</span>
                  <div class="flex-1 min-w-0">
                    <p class="font-medium">Sao chép prompt <span class="text-xs font-normal text-muted-foreground">(câu lệnh gửi cho AI)</span></p>
                    <p class="text-sm text-muted-foreground">Prompt dặn AI {{ picked === 3 ? 'tự lập danh sách ý, viết lại thành văn kể chuyện, rồi tự soát để không mất ý, không bịa' : 'giữ nguyên ý, chỉ đổi bảng, hình, danh sách thành lời' }}, và trả về file Word đặt sẵn chương, mục cho Sano.</p>
                  </div>
                  <Button class="shrink-0" @click="copy(1)"><component :is="copied === 1 ? Check : Copy" class="w-4 h-4" /> {{ copied === 1 ? 'Đã sao chép' : 'Sao chép prompt' }}</Button>
                </li>
                <li class="flex gap-4 rounded-xl border border-border p-4">
                  <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">2</span>
                  <div class="flex-1 min-w-0">
                    <p class="font-medium">Mở AI, đính kèm file Word của bạn, dán prompt rồi gửi</p>
                    <p class="text-sm text-muted-foreground">Tài liệu dài thì AI làm từng phần. Gõ "tiếp" cho tới khi xong.</p>
                    <div class="mt-2 flex gap-2">
                      <span v-for="a in ais" :key="a.k" class="h-8 px-3 rounded-md border border-border text-sm flex items-center gap-1.5 cursor-pointer hover:bg-muted">{{ a.name }} <ExternalLink class="w-3 h-3 text-muted-foreground" /></span>
                    </div>
                  </div>
                </li>
                <li class="flex gap-4 rounded-xl border border-border p-4">
                  <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">3</span>
                  <div class="flex-1 min-w-0">
                    <p class="font-medium">Tải file Word AI tạo về, nạp vào Sano</p>
                    <p class="text-sm text-muted-foreground">Bấm <b class="font-medium text-foreground">Tiếp: nạp file AI tạo</b> ở dưới. AI không tạo được file thì chép văn bản nó trả về, Sano cũng nhận.</p>
                    <div v-if="picked === 3" class="mt-2 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                      <span>Muốn chắc hơn: trước khi tải file, gửi thêm prompt soát lại, AI đối chiếu với file gốc và sửa chỗ mất ý.</span>
                      <button class="h-7 px-2.5 rounded-md border border-border text-xs flex items-center gap-1 text-foreground hover:bg-muted" @click="copy(2)"><component :is="copied === 2 ? Check : Copy" class="w-3.5 h-3.5" /> {{ copied === 2 ? 'Đã sao chép' : 'Sao chép prompt soát lại' }}</button>
                    </div>
                  </div>
                </li>
              </ol>

              <div class="mt-4 flex items-center gap-3 rounded-lg bg-muted/50 px-4 py-3 text-sm">
                <Sparkles class="w-4 h-4 text-chart shrink-0" />
                <span class="flex-1">Làm sách thường xuyên? Nạp skill Sano cho AI một lần, lần sau chỉ cần gửi file.</span>
                <Button size="sm" variant="outline" @click="mode = 'skill'">Nạp skill cho AI</Button>
              </div>
              <p class="mt-3 text-xs text-muted-foreground flex items-start gap-1.5">
                <ShieldCheck class="w-3.5 h-3.5 mt-px shrink-0" />
                <span>Đọc lại bản AI viết, nhất là số liệu, tên riêng.{{ picked === 3 ? ' Bản viết lại hợp với tài liệu của bạn hoặc để nghe riêng, không để phát hành lại sách của người khác.' : '' }} Tài liệu nhạy cảm thì cân nhắc trước khi gửi lên AI trên mạng.</span>
              </p>
            </div>

            <!-- ═══ Bước 2: Nạp file · Dán kết quả ═══ -->
            <div v-else class="max-w-2xl">
              <h1 class="text-xl font-semibold tracking-tight">{{ picked === 1 ? 'Nạp file Word' : 'Nạp file AI tạo' }}</h1>
              <p class="text-sm text-muted-foreground">Cấp {{ picked }} · {{ opt?.title }} <button class="text-primary hover:underline ml-1" @click="mode = 'chosen'">Đổi</button></p>
              <template v-if="pasteTab === 'file'">
                <div class="mt-4 w-full h-52 rounded-xl border-2 border-dashed border-border grid place-items-center">
                  <span class="text-center"><Upload class="w-8 h-8 mx-auto text-muted-foreground" /><span class="block mt-3 font-medium">{{ picked === 1 ? 'Kéo file .docx vào đây' : 'Kéo file Word AI tạo vào đây' }}</span><span class="block text-sm text-muted-foreground">hoặc bấm để chọn file .docx</span></span>
                </div>
                <button v-if="picked !== 1" class="mt-3 text-sm text-muted-foreground hover:text-foreground flex items-center gap-1.5" @click="pasteTab = 'paste'">
                  <ClipboardPaste class="w-4 h-4" /> AI không tạo được file Word? <span class="text-primary">Dán văn bản AI trả về</span>
                </button>
              </template>
              <template v-else>
                <button class="mt-4 text-sm text-muted-foreground hover:text-foreground flex items-center gap-1" @click="pasteTab = 'file'"><ChevronLeft class="w-4 h-4" /> Nạp file Word thay vì dán</button>
                <textarea class="mt-3 w-full h-48 rounded-lg border border-input bg-background p-3 text-sm font-mono leading-relaxed" spellcheck="false">% Quản lý thời gian cho người bận rộn
# Chương 1. Vì sao cả ngày bận mà không xong việc
## Chín giờ tối của Lan
Chín giờ tối. Lan vẫn chưa xong việc, dù cả ngày chẳng nghỉ phút nào.
Vì sao?
## Việc gấp và việc quan trọng
…</textarea>
                <p class="mt-2 text-xs text-muted-foreground">Dán nguyên kết quả AI trả về. Dòng <code class="font-mono">#</code> là chương, <code class="font-mono">##</code> là mục, <code class="font-mono">%</code> là tên sách (nếu có).</p>
                <div class="mt-3 rounded-lg border border-border p-3 flex items-center gap-3 text-sm">
                  <FileText class="w-6 h-6 text-primary shrink-0" />
                  <span class="flex-1">Nhận ra <b>6 chương</b> · 24 mục · khoảng 58 phút nghe</span>
                  <X class="w-4 h-4 text-muted-foreground" />
                </div>
              </template>
            </div>
          </div>

          <!-- Chân: nút chính nói rõ bước kế tiếp -->
          <div class="shrink-0 border-t border-border px-6 h-14 flex items-center justify-between">
            <Button variant="ghost" :disabled="mode === 'choose' || mode === 'chosen'" @click="back"><ChevronLeft class="w-4 h-4" /> Quay lại</Button>
            <div class="flex items-center gap-3">
              <span v-if="mode === 'choose'" class="text-xs text-muted-foreground">Chọn một cách để tiếp tục</span>
              <Button :disabled="mode === 'choose'" @click="next">
                {{ mode === 'chosen' ? (opt?.ai ? 'Tiếp: nhờ AI viết lại' : 'Tiếp: nạp file Word') : onAI ? 'Tiếp: nạp file AI tạo' : mode === 'choose' ? 'Tiếp' : 'Tiếp: mục lục' }} <ChevronRight class="w-4 h-4" />
              </Button>
            </div>
          </div>
        </section>
      </div>

      <!-- ═══ Nạp skill cho AI ═══ -->
      <div v-if="mode === 'skill'" class="absolute inset-0 bg-black/40 grid place-items-center z-30">
        <div class="w-[520px] rounded-xl border border-border bg-background shadow-2xl">
          <div class="flex items-start justify-between p-5 pb-0">
            <div>
              <h2 class="font-semibold">Nạp skill làm sách nói cho AI</h2>
              <p class="text-xs text-muted-foreground mt-0.5">Làm một lần, dùng mãi. Skill gồm cả cấp 2 (làm mượt) và cấp 3 (viết lại thành văn sách nói).</p>
            </div>
            <button class="text-muted-foreground" @click="mode = 'ai3'"><X class="w-4 h-4" /></button>
          </div>
          <div class="px-5 mt-4 flex gap-1 border-b border-border">
            <button v-for="a in ais" :key="a.k" class="h-9 px-3 text-sm -mb-px border-b-2" :class="skillTab === a.k ? 'border-primary font-medium' : 'border-transparent text-muted-foreground'" @click="skillTab = a.k">{{ a.name }}</button>
          </div>
          <div class="p-5">
            <ol class="space-y-2.5 text-sm">
              <li v-for="(s, i) in ai.steps" :key="i" class="flex gap-3"><span class="h-5 w-5 rounded-full bg-muted grid place-items-center text-[11px] font-medium shrink-0 mt-px">{{ i + 1 }}</span><span>{{ s }}</span></li>
            </ol>
            <div class="mt-4 flex items-center gap-3 rounded-lg bg-muted/40 p-3">
              <FileText class="w-5 h-5 text-muted-foreground shrink-0" />
              <span class="flex-1 text-sm font-mono">{{ ai.file }}</span>
              <Button size="sm"><Download class="w-4 h-4" /> Tải về</Button>
            </div>
            <p class="mt-3 text-xs text-muted-foreground">{{ skillTab === 'claude' ? 'Skill cũng dùng được trong Claude Code.' : 'Tên các mục trong ChatGPT, Gemini có thể đổi theo phiên bản; trang hướng dẫn luôn cập nhật cách làm mới nhất.' }}</p>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D8 bản 2 · Mỗi màn một việc. Màn A chọn cách làm (nút chọn tròn, "Hợp khi", thời gian thêm, Nghe mẫu kết quả từng cấp).
      Màn B (cấp 2/3) hướng dẫn nhờ AI theo 3 bước đánh số. Cấp 1 sang thẳng Nạp file. Nút chính ở chân luôn ghi rõ bước kế tiếp.
    </p>
  </div>
</template>
