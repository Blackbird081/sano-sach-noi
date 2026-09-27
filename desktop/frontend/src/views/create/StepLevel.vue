<script setup lang="ts">
// B1 Cách đọc (wireframe D8): màn A chọn 1 trong 3 cấp, màn B (cấp 2/3) hướng dẫn
// nhờ AI theo 3 bước. Chọn AI đang dùng: Gemini bản miễn phí không tạo được file
// Word nên prompt của Gemini đòi khối mã để dán vào Sano (đã thử 27/09).
import { computed, onBeforeUnmount, ref } from 'vue'
import { Check, ChevronLeft, Clock, Copy, ExternalLink, Lightbulb, Pause, Play, ShieldCheck } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { copyText, openURL } from '../../lib/backend'
import { AI_TOOLS, promptFor, reviewPromptFor } from '../../lib/prompt'
import { saveAITool, saveLevel, state } from '../../lib/store'

const options = [
  { n: 1, title: 'Đọc nguyên văn', does: 'Sano đọc đúng từng chữ trong file Word.', fit: 'bài viết, ghi chép đã dễ đọc; hoặc bạn muốn nghe y nguyên', time: 'Không cần chuẩn bị', best: false },
  { n: 2, title: 'Làm mượt', does: 'Nhờ AI đổi bảng, hình, danh sách, chữ viết tắt thành lời. Giữ nguyên ý và giọng tác giả.', fit: 'tài liệu nhiều bảng biểu, gạch đầu dòng', time: 'Thêm khoảng 5 phút với AI', best: false },
  { n: 3, title: 'Viết lại thành văn sách nói', does: 'Nhờ AI viết lại như người kể: chuyện trước, lý thuyết sau, chương ngắn, cuối chương có ba ý cần nhớ.', fit: 'sách, giáo trình, tài liệu dài muốn nghe cuốn như sách nói', time: 'Thêm khoảng 10–15 phút với AI', best: true },
]

// Nghe mẫu: cùng một đoạn, ba cách viết, render sẵn bằng giọng mặc định. Thiếu
// file thì ẩn nút (bản dev chưa render mẫu).
const samples = import.meta.glob<string>('../../assets/levels/cap-*.mp3', { eager: true, import: 'default', query: '?url' })
const sampleFor = (n: number) => samples[`../../assets/levels/cap-${n}.mp3`]
const playing = ref(0)
let audio: HTMLAudioElement | null = null
function stop() {
  audio?.pause()
  audio = null
  playing.value = 0
}
function toggle(n: number) {
  const on = playing.value === n
  stop()
  const src = sampleFor(n)
  if (on || !src) return
  audio = new Audio(src)
  audio.onended = stop
  playing.value = n
  void audio.play().catch(stop)
}
onBeforeUnmount(stop)

function pick(n: number) {
  saveLevel(n)
}

const tool = computed(() => AI_TOOLS.find((a) => a.k === state.aiTool) ?? AI_TOOLS[0])
const gemini = computed(() => state.aiTool === 'gemini')
const copied = ref(0)
async function copy(which: 1 | 2) {
  const text = which === 1 ? promptFor(state.level, state.aiTool) : reviewPromptFor(state.aiTool)
  if (await copyText(text)) {
    copied.value = which
    setTimeout(() => (copied.value = 0), 1500)
  }
}
</script>

<template>
  <!-- ═══ Màn A: chọn cách làm ═══ -->
  <div v-if="state.levelScreen === 'choose'" class="max-w-3xl">
    <h1 class="text-xl font-semibold tracking-tight">Chọn cách làm sách nói</h1>
    <p class="text-sm text-muted-foreground">Văn viết để đọc bằng mắt, đọc to lên thường nghe chán.<template v-if="sampleFor(1)"> Bấm <b class="font-medium text-foreground">Nghe mẫu</b> ở từng cách để nghe khác biệt, rồi chọn một cách.</template></p>
    <div class="mt-4 space-y-2.5" role="radiogroup" aria-label="Cách làm sách nói">
      <div v-for="o in options" :key="o.n" role="radio" :aria-checked="state.level === o.n" tabindex="0"
        class="relative flex items-start gap-4 rounded-xl border p-4 cursor-pointer transition"
        :class="state.level === o.n ? 'border-primary ring-2 ring-primary/15 bg-primary/5' : 'border-border hover:border-primary/40 hover:bg-muted/30'"
        @click="pick(o.n)" @keydown.enter.prevent="pick(o.n)" @keydown.space.prevent="pick(o.n)">
        <span class="mt-0.5 h-5 w-5 rounded-full border-2 grid place-items-center shrink-0" :class="state.level === o.n ? 'border-primary' : 'border-muted-foreground/40'">
          <span v-if="state.level === o.n" class="h-2.5 w-2.5 rounded-full bg-primary"></span>
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
        <button v-if="sampleFor(o.n)" class="shrink-0 h-8 pl-2.5 pr-3 rounded-full border text-xs flex items-center gap-1.5"
          :class="playing === o.n ? 'border-chart bg-chart text-white' : 'border-border bg-background hover:bg-muted'"
          @click.stop="toggle(o.n)">
          <component :is="playing === o.n ? Pause : Play" class="w-3.5 h-3.5" /> {{ playing === o.n ? 'Dừng' : 'Nghe mẫu' }}
        </button>
      </div>
    </div>
    <p class="mt-4 text-sm text-muted-foreground flex items-start gap-2">
      <Lightbulb class="w-4 h-4 mt-0.5 text-rag-amber shrink-0" />
      <span>Không chắc? Sách hay tài liệu dài thì chọn <b class="font-medium text-foreground">cấp 3</b>. Bài ngắn, đọc đã trôi chảy thì chọn <b class="font-medium text-foreground">cấp 1</b>. Lần sau Sano nhớ lựa chọn của bạn.</span>
    </p>
  </div>

  <!-- ═══ Màn B: nhờ AI viết lại ═══ -->
  <div v-else class="max-w-3xl">
    <button class="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1" @click="state.levelScreen = 'choose'"><ChevronLeft class="w-4 h-4" /> Đổi cách làm</button>
    <h1 class="mt-2 text-xl font-semibold tracking-tight">Nhờ AI {{ state.level === 3 ? 'viết lại thành văn sách nói' : 'làm mượt tài liệu' }}</h1>
    <p class="text-sm text-muted-foreground">Dùng Claude, ChatGPT hoặc Gemini, bản miễn phí cũng được. Làm xong 3 bước là có bản để Sano đọc.</p>

    <div class="mt-4 flex items-center gap-3 text-sm">
      <span class="text-muted-foreground">Bạn dùng AI nào?</span>
      <div class="inline-flex rounded-lg border border-border p-0.5" role="radiogroup" aria-label="AI đang dùng">
        <button v-for="a in AI_TOOLS" :key="a.k" role="radio" :aria-checked="state.aiTool === a.k"
          class="h-8 px-3 rounded-md text-sm" :class="state.aiTool === a.k ? 'bg-primary text-primary-foreground font-medium' : 'text-muted-foreground hover:text-foreground'"
          @click="saveAITool(a.k)">{{ a.name }}</button>
      </div>
    </div>

    <ol class="mt-4 space-y-3">
      <li class="flex gap-4 rounded-xl border border-border p-4">
        <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">1</span>
        <div class="flex-1 min-w-0">
          <p class="font-medium">Sao chép prompt <span class="text-xs font-normal text-muted-foreground">(câu lệnh gửi cho AI)</span></p>
          <p class="text-sm text-muted-foreground">Prompt dặn AI {{ state.level === 3 ? 'tự lập danh sách ý, viết lại thành văn kể chuyện, rồi tự soát để không mất ý, không bịa' : 'giữ nguyên ý, chỉ đổi bảng, hình, danh sách thành lời' }}, và trả về {{ gemini ? 'văn bản đặt sẵn chương, mục để dán vào Sano' : 'file Word đặt sẵn chương, mục cho Sano' }}.</p>
        </div>
        <Button class="shrink-0" @click="copy(1)"><component :is="copied === 1 ? Check : Copy" class="w-4 h-4" /> {{ copied === 1 ? 'Đã sao chép' : 'Sao chép prompt' }}</Button>
      </li>
      <li class="flex gap-4 rounded-xl border border-border p-4">
        <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">2</span>
        <div class="flex-1 min-w-0">
          <p class="font-medium">Mở {{ tool.name }}, đính kèm file Word của bạn, dán prompt rồi gửi</p>
          <p class="text-sm text-muted-foreground">Tài liệu dài thì AI làm từng phần, gõ "tiếp" cho tới khi xong. AI từ chối hoặc dừng giữa chừng thì bấm tạo lại câu trả lời, hoặc mở cuộc trò chuyện mới.</p>
        </div>
        <Button variant="outline" class="shrink-0" @click="openURL(tool.url)">Mở {{ tool.name }} <ExternalLink class="w-3.5 h-3.5" /></Button>
      </li>
      <li class="flex gap-4 rounded-xl border border-border p-4">
        <span class="h-7 w-7 rounded-full bg-primary text-primary-foreground grid place-items-center text-sm font-semibold shrink-0">3</span>
        <div class="flex-1 min-w-0">
          <p v-if="gemini" class="font-medium">Sao chép văn bản AI trả về, dán vào Sano</p>
          <p v-else class="font-medium">Tải file Word AI tạo về, nạp vào Sano</p>
          <p v-if="gemini" class="text-sm text-muted-foreground">Bấm nút sao chép ở góc khung văn bản AI trả về, rồi bấm <b class="font-medium text-foreground">Tiếp: nạp file AI tạo</b> ở dưới và dán vào.</p>
          <p v-else class="text-sm text-muted-foreground">Bấm <b class="font-medium text-foreground">Tiếp: nạp file AI tạo</b> ở dưới. AI không tạo được file thì chép văn bản nó trả về, Sano cũng nhận.</p>
          <div v-if="state.level === 3" class="mt-2 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
            <span>Muốn chắc hơn: trước khi lấy kết quả, gửi thêm prompt soát lại, AI đối chiếu với file gốc và sửa chỗ mất ý.</span>
            <button class="h-7 px-2.5 rounded-md border border-border text-xs flex items-center gap-1 text-foreground hover:bg-muted" @click="copy(2)"><component :is="copied === 2 ? Check : Copy" class="w-3.5 h-3.5" /> {{ copied === 2 ? 'Đã sao chép' : 'Sao chép prompt soát lại' }}</button>
          </div>
        </div>
      </li>
    </ol>

    <p class="mt-4 text-xs text-muted-foreground flex items-start gap-1.5">
      <ShieldCheck class="w-3.5 h-3.5 mt-px shrink-0" />
      <span>Đọc lại bản AI viết, nhất là số liệu, tên riêng.{{ state.level === 3 ? ' Bản viết lại hợp với tài liệu của bạn hoặc để nghe riêng, không để phát hành lại sách của người khác.' : '' }} Tài liệu nhạy cảm thì cân nhắc trước khi gửi lên AI trên mạng.</span>
    </p>
  </div>
</template>
