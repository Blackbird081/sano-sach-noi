<script setup lang="ts">
// D10 — Nghe trên điện thoại (thay nút "Xuất M4B"). Một hộp dẫn từng bước để người
// chưa biết M4B là gì vẫn làm được:
//   1. Chọn điện thoại: iPhone / Android (nhớ lần sau)
//   2. Cài app nghe sách BookPlayer (miễn phí): mã QR quét bằng điện thoại + link
//   3. Tạo file cho điện thoại (file M4B): tiến độ ngay trong hộp, dùng quãng nghỉ của cuốn
//   4. Chép sang điện thoại: cách theo máy (iPhone từ Mac: AirDrop; iPhone từ Windows,
//      Android: Zalo / Google Drive / cáp USB) rồi mở trong BookPlayer
// Lần nào cũng đi đủ 4 bước (anh Việt chốt 27/09: ai cài rồi thì bấm Tiếp); máy đã chọn được chọn sẵn.
// (Trạng thái 5 "Lần sau" bỏ, giữ để tham khảo.)
// Wireframe tĩnh: dữ liệu giả, KHÔNG gọi API. Nút ngoài khung để chuyển trạng thái duyệt.
import { computed, ref } from 'vue'
import {
  Library, BarChart3, FilePlus2, Settings, Info, LifeBuoy, ExternalLink, Sun, Moon, ChevronLeft, Smartphone, Package, FolderOpen,
  Trash2, X, Check, Apple, Loader2, QrCode, Download, Share2, Usb, MessageCircle, HardDrive, ArrowRight, ArrowLeft, Timer,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

type Mode = 'pick' | 'app' | 'making' | 'copy-ios' | 'copy-android' | 'again'
const q = new URLSearchParams(window.location.search)
const mode = ref<Mode>((q.get('mode') as Mode | null) || 'pick')
const dark = ref(false)
const phone = ref<'ios' | 'android'>(mode.value === 'copy-android' ? 'android' : 'ios')
const step = computed(() => ({ pick: 1, app: 2, making: 3, again: 3, 'copy-ios': 4, 'copy-android': 4 })[mode.value])
const steps = ['Chọn máy', 'Cài app nghe', 'Tạo file', 'Chép sang máy']

function setMode(m: Mode) {
  mode.value = m
  if (m === 'copy-android') phone.value = 'android'
  if (m === 'copy-ios') phone.value = 'ios'
}
function pickPhone(p: 'ios' | 'android') {
  phone.value = p
  mode.value = 'app'
}

const appName = 'BookPlayer'
const store = computed(() => (phone.value === 'ios' ? 'App Store' : 'Google Play'))
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
    <div class="flex flex-wrap gap-2 justify-center max-w-[1100px]">
      <button v-for="m in ([
        ['pick', '1. Chọn máy'], ['app', '2. Cài app (iPhone)'], ['making', '3. Đang tạo file'],
        ['copy-ios', '4a. Chép sang iPhone'], ['copy-android', '4b. Chép sang Android'], ['again', '5. Lần sau: vào thẳng tạo file'],
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
            <span v-for="n in nav" :key="n.key" class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm" :class="n.key === 'library' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'">
              <component :is="n.icon" class="w-4 h-4" /> {{ n.label }}
            </span>
          </nav>
          <div class="px-2 mt-2"><span class="w-full flex items-center gap-2.5 px-3 h-9 rounded-md text-sm text-muted-foreground"><LifeBuoy class="w-4 h-4" /> Hướng dẫn <ExternalLink class="w-3 h-3 ml-auto opacity-60" /></span></div>
        </aside>

        <!-- Màn nghe (nền, rút gọn) -->
        <section class="flex-1 flex flex-col p-6 min-w-0">
          <span class="text-sm text-muted-foreground flex items-center gap-1"><ChevronLeft class="w-4 h-4" /> Thư viện</span>
          <div class="flex-1 grid place-items-center text-sm text-muted-foreground">(màn nghe như hiện tại)</div>
          <!-- ─── ĐỔI: nút chính "Nghe trên điện thoại" thay "Xuất M4B" ─── -->
          <div class="flex flex-wrap items-center gap-2 border-t border-border pt-4">
            <Button size="sm"><Smartphone class="w-4 h-4" /> Nghe trên điện thoại</Button>
            <Button variant="outline" size="sm"><Package class="w-4 h-4" /> Xuất gói zip</Button>
            <Button variant="ghost" size="sm"><FolderOpen class="w-4 h-4" /> Mở thư mục</Button>
            <Button variant="ghost" size="sm" class="ml-auto text-destructive"><Trash2 class="w-4 h-4" /> Xoá</Button>
          </div>
        </section>
      </div>

      <!-- ─── HỘP MỚI ─── -->
      <div class="absolute inset-0 top-9 bg-black/40 grid place-items-center">
        <div class="w-[600px] rounded-xl border border-border bg-background shadow-2xl">
          <div class="flex items-start justify-between px-6 pt-5">
            <div>
              <h2 class="font-semibold text-lg flex items-center gap-2"><Smartphone class="w-5 h-5" /> Nghe trên điện thoại</h2>
              <p class="text-sm text-muted-foreground mt-0.5">Ứng phó kinh tế suy thoái · nghe không cần mạng, nhớ chỗ đang nghe, có mục lục chương</p>
            </div>
            <X class="w-5 h-5 text-muted-foreground" />
          </div>

          <!-- Các bước -->
          <ol class="mx-6 mt-4 grid grid-cols-4 gap-2 text-xs">
            <li v-for="(s, i) in steps" :key="s" class="flex items-center gap-1.5" :class="i + 1 === step ? 'text-foreground font-medium' : i + 1 < step ? 'text-muted-foreground' : 'text-muted-foreground/60'">
              <span class="h-5 w-5 shrink-0 rounded-full grid place-items-center text-[11px]"
                :class="i + 1 < step ? 'bg-foreground/80 text-background' : i + 1 === step ? 'border-2 border-foreground' : 'border border-border'">
                <Check v-if="i + 1 < step" class="w-3 h-3" /><template v-else>{{ i + 1 }}</template>
              </span>{{ s }}
            </li>
          </ol>

          <div class="px-6 py-5 min-h-[300px]">
            <!-- 1. Chọn máy -->
            <template v-if="mode === 'pick'">
              <p class="text-sm">Bạn dùng điện thoại nào?</p>
              <div class="mt-3 grid grid-cols-2 gap-3">
                <button class="rounded-lg border border-border p-4 text-left hover:border-foreground/40 hover:bg-muted/50" @click="pickPhone('ios')">
                  <Apple class="w-7 h-7" /><span class="mt-2 block font-medium">iPhone, iPad</span><span class="block text-xs text-muted-foreground">iOS</span>
                </button>
                <button class="rounded-lg border border-border p-4 text-left hover:border-foreground/40 hover:bg-muted/50" @click="pickPhone('android')">
                  <Smartphone class="w-7 h-7" /><span class="mt-2 block font-medium">Android</span><span class="block text-xs text-muted-foreground">Samsung, Xiaomi, Oppo…</span>
                </button>
              </div>
              <p class="mt-4 text-xs text-muted-foreground">Sano tạo một file sách nói (định dạng M4B) cho cả cuốn: có bìa, mục lục chương. Bạn chép file này sang điện thoại rồi nghe bằng app miễn phí. Làm lần đầu mất khoảng 3 phút.</p>
            </template>

            <!-- 2. Cài app -->
            <template v-else-if="mode === 'app'">
              <p class="text-sm">Cài app nghe sách <b>{{ appName }}</b> trên {{ phone === 'ios' ? 'iPhone' : 'điện thoại Android' }}. Miễn phí, không quảng cáo, mã nguồn mở.</p>
              <div class="mt-4 flex gap-5 items-center">
                <div class="h-36 w-36 shrink-0 rounded-lg border border-border grid place-items-center bg-white text-slate-400"><QrCode class="w-24 h-24" /></div>
                <div class="text-sm space-y-2">
                  <p>Mở camera điện thoại, quét mã này để vào <b>{{ store }}</b>, bấm <b>Nhận</b> / <b>Cài đặt</b>.</p>
                  <p class="text-muted-foreground text-xs">Hoặc tìm "{{ appName }}" trên {{ store }}.</p>
                  <a class="inline-flex items-center gap-1 text-xs text-primary"><ExternalLink class="w-3 h-3" /> Mở trang {{ appName }} trên {{ store }}</a>
                </div>
              </div>
              <p class="mt-5 text-xs text-muted-foreground">BookPlayer nhớ chỗ đang nghe, chỉnh tốc độ, hẹn giờ tắt{{ phone === 'ios' ? ', chạy trên CarPlay' : ', chạy trên Android Auto' }}.</p>
            </template>

            <!-- 3. Tạo file -->
            <template v-else-if="mode === 'making' || mode === 'again'">
              <div class="rounded-lg border border-border p-4">
                <div class="flex items-center gap-2 text-sm"><Loader2 class="w-4 h-4 animate-spin text-primary" /> Đang tạo file cho điện thoại… 38%</div>
                <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full w-[38%] bg-primary rounded-full"></div></div>
                <div class="mt-1.5 flex justify-between text-xs text-muted-foreground"><span>tiểu mục 25/65 · còn khoảng 20 giây</span><button class="text-foreground">Huỷ</button></div>
              </div>
              <ul class="mt-4 space-y-1.5 text-xs text-muted-foreground">
                <li>Lưu vào: <b class="text-foreground">Tải về / Ứng phó kinh tế suy thoái.m4b</b> · khoảng 20 MB <a class="text-primary ml-1">Đổi chỗ lưu</a></li>
                <li class="flex items-center gap-1"><Timer class="w-3.5 h-3.5" /> Quãng nghỉ: theo cài đặt chung (Vừa: 1,5 giây giữa tiểu mục, 2 giây sang chương)</li>
              </ul>
              <p v-if="mode === 'again'" class="mt-4 text-xs text-muted-foreground">Đã cài BookPlayer rồi nên Sano tạo file luôn. <a class="text-primary">Xem lại hướng dẫn từ đầu</a></p>
            </template>

            <!-- 4a. Chép sang iPhone -->
            <template v-else-if="mode === 'copy-ios'">
              <div class="flex items-center gap-2 text-sm"><Check class="w-4 h-4 text-rag-green" /> Đã tạo <b>Ứng phó kinh tế suy thoái.m4b</b> (42 phút, 19,8 MB) trong Tải về.</div>
              <p class="mt-4 text-sm font-medium">Chép sang iPhone, chọn một cách:</p>
              <div class="mt-2 grid grid-cols-2 gap-3 text-sm">
                <div class="rounded-lg border border-border p-3">
                  <p class="font-medium flex items-center gap-1.5"><Share2 class="w-4 h-4" /> AirDrop (máy Mac)</p>
                  <ol class="mt-1.5 list-decimal pl-4 space-y-0.5 text-xs text-muted-foreground">
                    <li>Bấm <b class="text-foreground">Gửi bằng AirDrop</b> dưới đây, chọn iPhone</li>
                    <li>Trên iPhone chọn <b class="text-foreground">Lưu vào Tệp</b></li>
                  </ol>
                  <Button size="sm" variant="outline" class="mt-2 w-full"><Share2 class="w-3.5 h-3.5" /> Gửi bằng AirDrop</Button>
                </div>
                <div class="rounded-lg border border-border p-3">
                  <p class="font-medium flex items-center gap-1.5"><MessageCircle class="w-4 h-4" /> Zalo, iCloud, Google Drive</p>
                  <ol class="mt-1.5 list-decimal pl-4 space-y-0.5 text-xs text-muted-foreground">
                    <li>Gửi file cho chính bạn (Zalo "Cloud của tôi") hoặc tải lên Drive</li>
                    <li>Trên iPhone mở file → <b class="text-foreground">Chia sẻ</b> → <b class="text-foreground">Lưu vào Tệp</b></li>
                  </ol>
                  <Button size="sm" variant="outline" class="mt-2 w-full"><FolderOpen class="w-3.5 h-3.5" /> Mở thư mục chứa file</Button>
                </div>
              </div>
              <p class="mt-4 text-sm"><b>Cuối cùng:</b> mở BookPlayer → bấm <b>+</b> → <b>Nhập tệp</b> → chọn file vừa lưu. Xong, bấm nghe.</p>
            </template>

            <!-- 4b. Chép sang Android -->
            <template v-else>
              <div class="flex items-center gap-2 text-sm"><Check class="w-4 h-4 text-rag-green" /> Đã tạo <b>Ứng phó kinh tế suy thoái.m4b</b> (42 phút, 19,8 MB) trong Tải về.</div>
              <p class="mt-4 text-sm font-medium">Chép sang điện thoại Android, chọn một cách:</p>
              <div class="mt-2 grid grid-cols-3 gap-3 text-xs text-muted-foreground">
                <div class="rounded-lg border border-border p-3"><p class="text-sm font-medium text-foreground flex items-center gap-1.5"><MessageCircle class="w-4 h-4" /> Zalo</p>Gửi file vào "Cloud của tôi", trên điện thoại bấm file → <b class="text-foreground">Tải về</b>.</div>
                <div class="rounded-lg border border-border p-3"><p class="text-sm font-medium text-foreground flex items-center gap-1.5"><HardDrive class="w-4 h-4" /> Google Drive</p>Tải file lên Drive, trên điện thoại mở Drive → <b class="text-foreground">Tải xuống</b>.</div>
                <div class="rounded-lg border border-border p-3"><p class="text-sm font-medium text-foreground flex items-center gap-1.5"><Usb class="w-4 h-4" /> Cáp USB</p>Cắm máy, chọn <b class="text-foreground">Truyền tệp</b>, chép vào thư mục <b class="text-foreground">Audiobooks</b>.</div>
              </div>
              <Button size="sm" variant="outline" class="mt-3"><FolderOpen class="w-3.5 h-3.5" /> Mở thư mục chứa file</Button>
              <p class="mt-4 text-sm"><b>Cuối cùng:</b> mở BookPlayer → thêm sách → chọn file vừa tải (thường trong <b>Download</b>). Xong, bấm nghe.</p>
            </template>
          </div>

          <div class="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
            <a class="text-xs text-muted-foreground inline-flex items-center gap-1"><ExternalLink class="w-3 h-3" /> Hướng dẫn có ảnh</a>
            <div class="flex gap-2">
              <Button v-if="mode === 'app'" variant="outline" size="sm" @click="mode = 'pick'"><ArrowLeft class="w-4 h-4" /> Quay lại</Button>
              <Button v-if="mode === 'app'" size="sm" @click="mode = 'making'"><Download class="w-4 h-4" /> Đã cài, tạo file</Button>
              <Button v-if="mode === 'making' || mode === 'again'" variant="outline" size="sm">Để chạy nền</Button>
              <Button v-if="mode === 'copy-ios' || mode === 'copy-android'" size="sm">Xong <ArrowRight class="w-4 h-4" /></Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p class="text-xs text-muted-foreground max-w-[1100px] text-center">
      D10 · Nút "Nghe trên điện thoại" thay "Xuất M4B" (màn nghe và màn render xong). Chữ "M4B" chỉ nhắc phụ. Lần nào cũng đi đủ 4 bước, máy đã chọn được chọn sẵn, có nút Tiếp.
      "Để chạy nền": đóng hộp, tiến độ vẫn hiện dưới hàng nút như hiện tại; xong thì hộp tự mở lại ở bước 4.
    </p>
  </div>
</template>
