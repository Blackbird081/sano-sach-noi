<script setup lang="ts">
// D8b — Màn "Nhờ AI" (cấp 2/3) làm lại theo góp ý anh Việt 27/09:
// - Không chọn sẵn AI: lần đầu phải bấm 1 trong 3 thẻ Claude / ChatGPT / Gemini,
//   chưa chọn thì chưa hiện các bước và khoá nút Tiếp. Thẻ đã chọn luôn nhìn thấy.
// - Bỏ nút "prompt soát lại" (prompt chính đã dặn AI tự soát; soát lại vẫn có trong skill + trang hướng dẫn).
// - Mỗi bước một dòng + một nút; mẹo gom thành một dòng nhỏ; "Nạp skill" là một dòng cuối, không bị chân trang che.
// - Nút ghi tên AI ("Sao chép prompt cho Gemini", "Mở Gemini"); Gemini đổi bước 3 + nút chân thành "dán văn bản".
// - Hộp nạp skill mở theo AI đã chọn (không tab): Claude tải zip SKILL.md; ChatGPT đường chính là
//   Dự án + file hướng dẫn (ai cũng dùng được), dòng phụ cho gói Business/Enterprise tải zip như Claude
//   (ChatGPT Skills dùng cùng định dạng SKILL.md). Gemini: không hướng dẫn tạo Gem (tin rò rỉ Google bỏ
//   Gem từ 13/10/2026, Skills của Gemini chỉ gói trả phí) → dòng cuối ghi "mỗi lần dán prompt, không cần skill".
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, FilePlus2, BarChart3, Settings, Info, LifeBuoy, ExternalLink, Check, ChevronRight, ChevronLeft,
  Copy, FileText, ClipboardPaste, Sparkles, Lightbulb, Moon, Sun, Download, X, CheckCircle2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type Tool = '' | 'claude' | 'chatgpt' | 'gemini'
const q = new URLSearchParams(window.location.search)
const tool = ref<Tool>((q.get('ai') as Tool | null) || '')
const level = ref(Number(q.get('level')) === 2 ? 2 : 3)
const dark = ref(false)
const copied = ref(false)
const skill = ref(q.get('skill') === '1')
const saved = ref(false)
// ChatGPT: chọn gói trong hộp nạp skill (không chọn sẵn)
const plan = ref<'' | 'skills' | 'project'>((q.get('plan') as '' | 'skills' | 'project' | null) || '')
const plans = [
  { k: 'skills' as const, title: 'Có mục Skills', sub: 'Gói Business, Enterprise, Edu' },
  { k: 'project' as const, title: 'Chưa có mục Skills', sub: 'Gói Free, Plus, Pro · dùng Dự án' },
]
type DStep = { text: string; act?: 'file' | 'open' | 'copy'; file?: string; label?: string }
const dlgSteps = computed<DStep[]>(() => {
  if (tool.value === 'claude') return [
    { text: 'Tải file skill về, không cần giải nén.', act: 'file', file: 'sano-sach-noi.zip' },
    { text: 'Mở Customize → Skills → nút + → Create skill → Upload a skill, chọn file vừa tải.', act: 'open', label: 'Mở Skills của Claude' },
  ]
  if (plan.value === 'skills') return [
    { text: 'Tải file skill về, không cần giải nén.', act: 'file', file: 'sano-sach-noi.zip' },
    { text: 'Mở ChatGPT → Skills → Tạo → Tải lên, chọn file vừa tải.', act: 'open', label: 'Mở ChatGPT' },
  ]
  return [
    { text: 'Tải file hướng dẫn về.', act: 'file', file: 'sano-huong-dan-ai.txt' },
    { text: 'Tạo Dự án (Project) tên "Sano – sách nói", thêm file vừa tải vào phần Tệp.', act: 'open', label: 'Mở ChatGPT' },
    { text: 'Dán câu Sano đưa vào ô Hướng dẫn (Instructions) của dự án. Từ nay làm sách trong dự án đó.', act: 'copy' },
  ]
})

const tools = [
  { k: 'claude' as const, name: 'Claude', gives: 'Trả về file Word', icon: FileText },
  { k: 'chatgpt' as const, name: 'ChatGPT', gives: 'Trả về file Word', icon: FileText },
  { k: 'gemini' as const, name: 'Gemini', gives: 'Trả về văn bản để dán', icon: ClipboardPaste },
]
const t = computed(() => tools.find((x) => x.k === tool.value))
const gemini = computed(() => tool.value === 'gemini')
function copy() {
  copied.value = true
  setTimeout(() => (copied.value = false), 1500)
}

const steps = ['Cách đọc', 'Nạp file', 'Mục lục', 'Giọng đọc', 'Lời mở đầu', 'Nghe thử', 'Render']
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
    <div class="flex flex-wrap gap-2 justify-center">
      <button v-for="m in ([['', '1. Chưa chọn AI'], ['claude', '2. Đã chọn Claude'], ['chatgpt', '3. Đã chọn ChatGPT'], ['gemini', '4. Đã chọn Gemini']] as [Tool, string][])" :key="m[0]"
        class="h-8 px-3 rounded-full border text-xs"
        :class="tool === m[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'"
        @click="tool = m[0]">{{ m[1] }}</button>
      <button class="h-8 px-3 rounded-full border text-xs" :class="skill ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="skill = !skill; if (!tool || tool === 'gemini') tool = 'claude'">5. Hộp nạp skill</button>
      <button v-if="skill && tool === 'chatgpt'" v-for="p in ([['', 'ChatGPT: chưa chọn gói'], ['skills', 'Có Skills'], ['project', 'Dự án']] as ['' | 'skills' | 'project', string][])" :key="p[0]" class="h-8 px-3 rounded-full border text-xs" :class="plan === p[0] ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="plan = p[0]">{{ p[1] }}</button>
      <button class="h-8 px-3 rounded-full border border-border bg-background text-xs text-foreground" @click="level = level === 3 ? 2 : 3">Cấp {{ level }} (đổi)</button>
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
              <span class="flex items-center gap-1.5 px-1.5 h-8 rounded-md text-sm whitespace-nowrap" :class="i === 0 ? 'text-foreground font-medium' : 'text-muted-foreground'">
                <span class="h-6 w-6 rounded-full grid place-items-center text-xs" :class="i === 0 ? 'bg-primary text-primary-foreground' : 'bg-muted'">{{ i + 1 }}</span>{{ s }}
              </span>
              <ChevronRight v-if="i < steps.length - 1" class="w-3.5 h-3.5 text-muted-foreground/50 shrink-0" />
            </template>
          </div>

          <div class="flex-1 overflow-auto p-6">
            <div class="max-w-3xl">
              <button class="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Đổi cách làm</button>
              <h1 class="mt-2 text-xl font-semibold tracking-tight">Nhờ AI {{ level === 3 ? 'viết lại thành văn sách nói' : 'làm mượt tài liệu' }}</h1>
              <p class="text-sm text-muted-foreground">AI bản miễn phí cũng được. {{ level === 3 ? 'Mất khoảng 10–15 phút.' : 'Mất khoảng 5 phút.' }}</p>

              <!-- Chọn AI: không chọn sẵn -->
              <p class="mt-5 text-sm font-medium">Bạn dùng AI nào?</p>
              <div class="mt-2 grid grid-cols-3 gap-3" role="radiogroup" aria-label="AI đang dùng">
                <button v-for="x in tools" :key="x.k" role="radio" :aria-checked="tool === x.k"
                  class="relative rounded-xl border p-3.5 text-left transition"
                  :class="tool === x.k ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40 hover:bg-muted/30'"
                  @click="tool = x.k">
                  <span class="absolute top-3 right-3 h-5 w-5 rounded-full border-2 grid place-items-center" :class="tool === x.k ? 'border-primary bg-primary' : 'border-muted-foreground/30'">
                    <Check v-if="tool === x.k" class="w-3 h-3 text-primary-foreground" />
                  </span>
                  <span class="block font-semibold">{{ x.name }}</span>
                  <span class="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground"><component :is="x.icon" class="w-3.5 h-3.5" /> {{ x.gives }}</span>
                </button>
              </div>

              <!-- Chưa chọn: chưa hiện bước -->
              <div v-if="!t" class="mt-4 rounded-xl border border-dashed border-border py-10 text-center text-sm text-muted-foreground">
                Chọn AI bạn dùng để Sano đưa đúng prompt và cách làm.
              </div>

              <!-- Đã chọn: 3 bước, mỗi bước một dòng + một nút -->
              <template v-else>
                <ol class="mt-4 rounded-xl border border-border divide-y divide-border">
                  <li class="flex items-center gap-4 px-4 py-3.5">
                    <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">1</span>
                    <p class="flex-1 font-medium">Sao chép prompt <span class="font-normal text-muted-foreground">(câu lệnh gửi cho AI)</span></p>
                    <Button class="shrink-0 w-60" @click="copy"><component :is="copied ? Check : Copy" class="w-4 h-4" /> {{ copied ? 'Đã sao chép' : `Sao chép prompt cho ${t.name}` }}</Button>
                  </li>
                  <li class="flex items-center gap-4 px-4 py-3.5">
                    <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">2</span>
                    <p class="flex-1 font-medium">Mở {{ t.name }}, đính kèm file Word của bạn, dán prompt rồi gửi</p>
                    <Button variant="outline" class="shrink-0 w-60">Mở {{ t.name }} <ExternalLink class="w-3.5 h-3.5" /></Button>
                  </li>
                  <li class="flex items-center gap-4 px-4 py-3.5">
                    <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">3</span>
                    <p class="flex-1 font-medium">
                      <template v-if="gemini">Gemini viết xong: bấm nút sao chép ở góc khung văn bản, rồi bấm <span class="text-primary">Tiếp</span></template>
                      <template v-else>{{ t.name }} viết xong: tải file Word về máy, rồi bấm <span class="text-primary">Tiếp</span></template>
                    </p>
                  </li>
                </ol>
                <p class="mt-3 text-xs text-muted-foreground flex items-start gap-1.5">
                  <Lightbulb class="w-3.5 h-3.5 mt-px text-rag-amber shrink-0" />
                  <span>Tài liệu dài thì gõ "tiếp" để AI làm phần sau. AI từ chối hoặc dừng giữa chừng thì bấm tạo lại câu trả lời. Đọc lại bản AI viết, nhất là số liệu, tên riêng.</span>
                </p>
                <div v-if="!gemini" class="mt-5 flex items-center gap-2 text-sm">
                  <Sparkles class="w-4 h-4 text-chart shrink-0" />
                  <span class="text-muted-foreground">Làm sách thường xuyên?</span>
                  <button class="text-primary hover:underline" @click="skill = true; saved = false">Nạp skill cho {{ t.name }} một lần, lần sau chỉ cần gửi file</button>
                </div>
                <p v-else class="mt-5 flex items-center gap-2 text-sm text-muted-foreground">
                  <Sparkles class="w-4 h-4 text-chart shrink-0" /> Với Gemini, mỗi lần làm sách chỉ cần dán prompt, Gemini chưa hỗ trợ skill.
                </p>
              </template>
            </div>
          </div>

          <div class="shrink-0 border-t border-border px-6 h-14 flex items-center justify-between">
            <Button variant="ghost"><ChevronLeft class="w-4 h-4" /> Quay lại</Button>
            <div class="flex items-center gap-3">
              <span v-if="!t" class="text-xs text-muted-foreground">Chọn AI để tiếp tục</span>
              <Button :disabled="!t">{{ gemini ? 'Tiếp: dán văn bản AI trả về' : 'Tiếp: nạp file AI tạo' }} <ChevronRight class="w-4 h-4" /></Button>
            </div>
          </div>
        </section>
      </div>

      <!-- ═══ Hộp nạp skill: theo AI đã chọn, mỗi bước có nút làm ngay tại dòng ═══ -->
      <div v-if="skill && t && !gemini" class="absolute inset-0 bg-black/40 grid place-items-center z-30" @click.self="skill = false">
        <div class="w-[560px] rounded-xl border border-border bg-background shadow-2xl">
          <div class="flex items-start justify-between p-5 pb-0">
            <div>
              <h2 class="font-semibold">Nạp skill làm sách nói cho {{ t.name }}</h2>
              <p class="text-xs text-muted-foreground mt-0.5">Làm một lần. Lần sau chỉ cần đính kèm file Word và gõ "cấp 3" (hoặc "cấp 2").</p>
            </div>
            <button class="text-muted-foreground" @click="skill = false"><X class="w-4 h-4" /></button>
          </div>

          <div class="p-5">
            <!-- ChatGPT: chọn gói trước, hai cách ngang nhau -->
            <template v-if="tool === 'chatgpt'">
              <p class="text-sm font-medium">ChatGPT của bạn có mục Skills không?</p>
              <div class="mt-2 grid grid-cols-2 gap-3">
                <button v-for="p in plans" :key="p.k" class="relative rounded-xl border p-3 text-left transition"
                  :class="plan === p.k ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40'" @click="plan = p.k; saved = false">
                  <span class="absolute top-3 right-3 h-4 w-4 rounded-full border-2 grid place-items-center" :class="plan === p.k ? 'border-primary bg-primary' : 'border-muted-foreground/30'"><Check v-if="plan === p.k" class="w-2.5 h-2.5 text-primary-foreground" /></span>
                  <span class="block text-sm font-semibold">{{ p.title }}</span>
                  <span class="mt-0.5 block text-xs text-muted-foreground">{{ p.sub }}</span>
                </button>
              </div>
            </template>

            <ol v-if="tool === 'claude' || plan" class="mt-4 rounded-xl border border-border divide-y divide-border text-sm" :class="tool === 'claude' && 'mt-0'">
              <li v-for="(st, i) in dlgSteps" :key="i" class="flex items-center gap-3 px-4 py-3">
                <span class="h-6 w-6 rounded-full bg-primary text-primary-foreground grid place-items-center text-xs font-semibold shrink-0">{{ i + 1 }}</span>
                <span class="flex-1">{{ st.text }}</span>
                <Button v-if="st.act === 'file'" size="sm" class="shrink-0 w-44" @click="saved = true"><component :is="saved ? CheckCircle2 : Download" class="w-4 h-4" /> {{ saved ? 'Đã lưu vào Tải về' : `Tải ${st.file}` }}</Button>
                <Button v-else-if="st.act === 'open'" size="sm" variant="outline" class="shrink-0 w-44">{{ st.label }} <ExternalLink class="w-3.5 h-3.5" /></Button>
                <Button v-else-if="st.act === 'copy'" size="sm" variant="outline" class="shrink-0 w-44" @click="copy"><component :is="copied ? Check : Copy" class="w-4 h-4" /> {{ copied ? 'Đã sao chép' : 'Sao chép câu dán' }}</Button>
              </li>
            </ol>
            <p v-if="tool === 'chatgpt' && plan === 'project'" class="mt-2 text-xs text-muted-foreground">Câu dán: "Mỗi khi tôi gửi tài liệu để làm sách nói, làm đúng theo file sano-huong-dan-ai.txt đã đính kèm…"</p>
            <p class="mt-3 text-xs text-muted-foreground">
              <template v-if="tool === 'claude'">Mọi gói Claude, kể cả miễn phí, đều nạp được skill.</template>
              <template v-else-if="plan === 'skills'">Mục Skills của ChatGPT dùng cùng định dạng skill với Claude, nên dùng chung một file.</template>
              <template v-else-if="plan === 'project'">Ô Hướng dẫn của Dự án tối đa 8.000 ký tự, nên hướng dẫn đầy đủ nằm trong file đính kèm.</template>
            </p>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D8b · Màn Nhờ AI làm lại: không chọn sẵn AI, bỏ prompt soát lại khỏi app, mỗi bước một dòng một nút, nạp skill là một dòng cuối nhìn thấy được.
    </p>
  </div>
</template>
