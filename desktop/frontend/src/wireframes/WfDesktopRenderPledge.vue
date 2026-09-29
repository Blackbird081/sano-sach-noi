<script setup lang="ts">
// D19 — Cam kết trước khi render. Thay ô tick "Tôi xác nhận có quyền dùng tài liệu" ở thanh
// dưới bước Nghe thử bằng popup: bấm "Nghe ổn, render cả cuốn" → popup hiện, phải tick TỪNG
// cam kết mới bật nút "Cam kết và render". Mỗi lần render một tài liệu mới là hỏi lại; lúc
// xác nhận + phiên bản cam kết ghi vào thông tin sách (bằng chứng người dùng đã cam kết).
// Wireframe tĩnh: KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import { ChevronLeft, ChevronRight, FileCheck2, Scale, UserX, Share2, ShieldAlert, BookOpen, Sun, Moon, X, Check } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const dark = ref(false)
const pledges = [
  {
    icon: FileCheck2,
    title: 'Tôi có quyền dùng tài liệu này',
    desc: 'Tài liệu do tôi viết, tác phẩm đã hết thời hạn bảo hộ quyền tác giả, hoặc tôi đã được tác giả / chủ sở hữu cho phép bằng văn bản chuyển thành sách nói.',
  },
  {
    icon: Scale,
    title: 'Nội dung không vi phạm pháp luật Việt Nam',
    desc: 'Không chống phá Nhà nước, xuyên tạc lịch sử; không kích động bạo lực, thù hằn dân tộc, tôn giáo; không đồi truỵ, mê tín, tin giả; không xúc phạm danh dự người khác; không chứa bí mật nhà nước hay thông tin cá nhân của người khác.',
  },
  {
    icon: UserX,
    title: 'Không mạo danh, không lừa đảo',
    desc: 'Không dùng giọng đọc để giả giọng hay mạo danh người thật, cơ quan, tổ chức; không dùng sách nói, video, ảnh tạo ra để lừa đảo.',
  },
  {
    icon: Share2,
    title: 'Không phát tán tác phẩm của người khác',
    desc: 'Không đăng công khai, chia sẻ cho người ngoài, bán hay dùng vào mục đích thương mại sách nói, video, ảnh làm từ tác phẩm của người khác khi chưa có sự đồng ý bằng văn bản của chủ sở hữu.',
  },
  {
    icon: ShieldAlert,
    title: 'Tôi tự chịu hoàn toàn trách nhiệm',
    desc: 'Tôi tự chịu trách nhiệm trước pháp luật về nội dung đưa vào và mọi thứ tạo ra từ Sano. Tác giả, người đóng góp và nhà tài trợ của Sano không liên quan và không chịu trách nhiệm; nếu họ bị khiếu nại vì việc làm của tôi, tôi chịu mọi chi phí và thiệt hại phát sinh.',
  },
]
const checked = ref<boolean[]>(pledges.map(() => false))
const readTerms = ref(false)
const allOk = computed(() => checked.value.every(Boolean) && readTerms.value)
const count = computed(() => checked.value.filter(Boolean).length + (readTerms.value ? 1 : 0))
function setAll(v: boolean) {
  checked.value = pledges.map(() => v)
  readTerms.value = v
}
const steps = ['Cách đọc', 'Nạp file', 'Mục lục', 'Giọng đọc', 'Lời mở đầu', 'Nghe thử', 'Render']
</script>

<template>
  <div :class="{ dark }" class="min-h-dvh w-full bg-muted/60 flex flex-col items-center justify-center gap-4 p-6">
    <div class="flex flex-wrap gap-2 justify-center">
      <button class="h-8 px-3 rounded-full border text-xs" :class="!allOk ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="setAll(false)">Vừa mở (chưa tick)</button>
      <button class="h-8 px-3 rounded-full border text-xs" :class="allOk ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background text-foreground'" @click="setAll(true)">Đã tick đủ</button>
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

      <!-- Nền: bước Nghe thử (giản lược). Thanh dưới KHÔNG còn ô tick, chỉ còn nút render. -->
      <div class="flex-1 min-h-0 flex">
        <aside class="w-52 shrink-0 border-r border-border bg-muted/30 p-4 text-sm text-muted-foreground">
          <div class="flex items-center gap-2 font-semibold text-foreground"><img src="@/assets/favicon.svg" alt="" class="h-7 w-7 rounded-md" /> Sano</div>
          <p class="mt-6">Thư viện</p><p class="mt-3 text-primary font-medium">Tạo sách nói</p><p class="mt-3">Hành trình nghe</p><p class="mt-3">Cài đặt</p>
        </aside>
        <section class="flex-1 min-w-0 flex flex-col">
          <div class="h-14 shrink-0 border-b border-border flex items-center gap-3 px-6 text-sm text-muted-foreground">
            <span v-for="(s, i) in steps" :key="s" :class="i === 5 ? 'text-foreground font-medium' : ''">{{ i + 1 }}. {{ s }}</span>
          </div>
          <div class="flex-1 p-6">
            <h1 class="text-xl font-semibold">Nghe thử trước khi render</h1>
            <div class="mt-4 h-64 rounded-lg border border-border bg-muted/30"></div>
          </div>
          <div class="shrink-0 border-t border-border px-6 h-16 flex items-center justify-between gap-4">
            <Button variant="ghost"><ChevronLeft class="w-4 h-4" /> Quay lại</Button>
            <Button>Nghe ổn, render cả cuốn <ChevronRight class="w-4 h-4" /></Button>
          </div>
        </section>
      </div>

      <!-- ═══ Popup cam kết ═══ -->
      <div class="absolute inset-0 top-9 z-20 grid place-items-center bg-background/70 backdrop-blur-sm">
        <div role="dialog" aria-modal="true" class="flex max-h-[660px] w-[640px] flex-col rounded-xl border border-border bg-card text-card-foreground shadow-2xl">
          <div class="flex items-start justify-between gap-4 border-b border-border px-5 py-4">
            <div>
              <h2 class="font-semibold">Cam kết trước khi tạo sách nói</h2>
              <p class="mt-0.5 text-xs text-muted-foreground">Tài liệu: <span class="font-medium text-foreground">Chánh niệm, nghệ thuật của sự có mặt.docx</span> · tick từng ô để tiếp tục</p>
            </div>
            <button class="text-muted-foreground hover:text-foreground" aria-label="Đóng"><X class="h-4 w-4" /></button>
          </div>

          <div class="flex-1 min-h-0 overflow-auto px-5 py-3 space-y-1.5">
            <label v-for="(p, i) in pledges" :key="p.title" class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2.5 transition-colors"
              :class="checked[i] ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/50'">
              <input v-model="checked[i]" type="checkbox" class="mt-0.5 h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
              <component :is="p.icon" class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
              <span class="min-w-0">
                <span class="block text-sm font-medium">{{ p.title }}</span>
                <span class="block text-xs leading-relaxed text-muted-foreground">{{ p.desc }}</span>
              </span>
            </label>
            <label class="flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-2.5 transition-colors" :class="readTerms ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/50'">
              <input v-model="readTerms" type="checkbox" class="mt-0.5 h-4 w-4 shrink-0 accent-[hsl(var(--primary))]" />
              <BookOpen class="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
              <span class="text-sm">Tôi đã đọc và đồng ý <a href="#" class="font-medium text-primary underline underline-offset-2" @click.prevent>Điều khoản sử dụng</a> của Sano.</span>
            </label>
          </div>

          <div class="border-t border-border px-5 py-3">
            <p class="text-[11px] leading-relaxed text-muted-foreground">Sano ghi lại thời điểm bạn cam kết vào thông tin cuốn sách. Vi phạm các cam kết trên, bạn tự chịu trách nhiệm trước pháp luật.</p>
            <div class="mt-3 flex items-center justify-between gap-3">
              <span class="text-xs" :class="allOk ? 'text-rag-green flex items-center gap-1' : 'text-muted-foreground'"><Check v-if="allOk" class="h-3.5 w-3.5" /> Đã tick {{ count }}/{{ pledges.length + 1 }}</span>
              <div class="flex gap-2">
                <Button variant="outline">Để sau</Button>
                <Button :disabled="!allOk">Cam kết và render <ChevronRight class="w-4 h-4" /></Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
