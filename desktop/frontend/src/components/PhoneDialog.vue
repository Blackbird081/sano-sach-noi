<script setup lang="ts">
// Hộp "Nghe trên điện thoại" (wireframe D10). Luồng + trạng thái ở lib/phone.ts;
// tạo file dùng lượt xuất M4B chung (lib/m4b.ts) nên đóng hộp vẫn chạy nền,
// tiến độ vẫn hiện dưới hàng nút ở màn nghe.
import { computed, onMounted, ref, watch } from 'vue'
import QRCode from 'qrcode'
import {
  Apple, ArrowLeft, ArrowRight, Check, Download, ExternalLink, FolderOpen, HardDrive, Loader2, MessageCircle, Share2, Smartphone, Timer, Usb, X,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { airDropM4B, canAirDrop, errText, openURL } from '../lib/backend'
import { DOCS } from '../lib/mock'
import { cancelM4B, fmtSize, showM4B, useM4B } from '../lib/m4b'
import { fmtGap, gapsFor, bookPause, levelLabel, pauseState } from '../lib/pause'
import { fmtLong } from '../lib/position'
import { STORE_URL, changeLocation, closePhone, make, phone, pickDevice, runInBackground, type Phone } from '../lib/phone'

const m4b = useM4B()
const st = computed(() => (m4b.status?.slug === phone.slug ? m4b.status : null))
const pct = computed(() => st.value?.progress?.percent ?? 0)
const fileName = computed(() => st.value?.path.split(/[\\/]/).pop() ?? '')
// Nơi lưu gọn: trong Tải về thì ghi "Tải về / tên file", nơi khác chỉ ghi tên file (rê chuột xem đủ).
const whereShort = computed(() => {
  const p = st.value?.running ? st.value.path : ''
  if (!p) return `Tải về / ${phone.title}.m4b`
  const name = p.split(/[\\/]/).pop()
  return /[\\/]Downloads[\\/][^\\/]+$/.test(p) ? `Tải về / ${name}` : name
})
const otherRunning = computed(() => (m4b.status?.running && m4b.status.slug !== phone.slug ? m4b.status : null))
const clickError = computed(() => (m4b.error && m4b.errorSlug === phone.slug ? m4b.error : ''))
const steps = ['Chọn máy', 'Cài app nghe', 'Tạo file', 'Chép sang máy']
const stepNo = computed(() => ({ pick: 1, app: 2, make: 3, copy: 4 })[phone.step])
const isIOS = computed(() => phone.device === 'ios')
const store = computed(() => (isIOS.value ? 'App Store' : 'Google Play'))
const sizeGuess = computed(() => (phone.totalSec > 0 ? fmtSize((phone.totalSec / 3600) * 29 * (1 << 20)) : ''))

// Quãng nghỉ sẽ dùng cho file này (mức riêng của cuốn hoặc cài đặt chung).
const pauseLine = computed(() => {
  const mine = bookPause(phone.slug)
  const g = gapsFor(phone.slug)
  const which = mine ? `riêng cuốn này (${levelLabel(mine.level)})` : `theo cài đặt chung (${levelLabel(pauseState.global.level)})`
  return `Quãng nghỉ: ${which}: ${fmtGap(g.section)} giữa tiểu mục, ${fmtGap(g.chapter)} sang chương`
})

// Mã QR tới trang BookPlayer trên kho ứng dụng của máy đã chọn.
const qr = ref('')
watch(
  () => phone.device,
  async (d) => {
    qr.value = d ? await QRCode.toDataURL(STORE_URL[d], { margin: 1, width: 288 }).catch(() => '') : ''
  },
  { immediate: true },
)

const airdrop = ref(false)
onMounted(async () => {
  airdrop.value = await canAirDrop().catch(() => false)
})
const actionError = ref('')
watch(() => [phone.open, phone.step], () => (actionError.value = ''))
async function sendAirDrop() {
  actionError.value = ''
  try {
    await airDropM4B()
  } catch (e) {
    actionError.value = errText(e)
  }
}
async function reveal() {
  actionError.value = ''
  await showM4B()
  if (m4b.error) actionError.value = m4b.error
}

function close() {
  // Đang tạo file mà đóng hộp: coi như để chạy nền (xong thì mở lại ở bước chép).
  if (st.value?.running) runInBackground()
  else closePhone()
}
const pick = (d: Phone) => pickDevice(d)
</script>

<template>
  <div v-if="phone.open" class="absolute inset-0 bg-background/70 backdrop-blur-sm grid place-items-center z-20" @keydown.esc="close">
    <div role="dialog" aria-modal="true" aria-labelledby="phone-title" class="w-[600px] max-w-[calc(100%-2rem)] max-h-[calc(100%-2rem)] overflow-auto rounded-xl border border-border bg-card text-card-foreground shadow-2xl flex flex-col">
      <div class="flex items-start justify-between gap-4 px-6 pt-5">
        <div class="min-w-0">
          <h2 id="phone-title" class="font-semibold text-lg flex items-center gap-2"><Smartphone class="w-5 h-5" /> Nghe trên điện thoại</h2>
          <p class="text-sm text-muted-foreground mt-0.5">{{ phone.title }} · nghe không cần mạng, nhớ chỗ đang nghe, có mục lục chương</p>
        </div>
        <button aria-label="Đóng" class="h-8 w-8 shrink-0 grid place-items-center rounded-md hover:bg-muted" @click="close"><X class="w-4 h-4" /></button>
      </div>

      <ol class="mx-6 mt-4 grid grid-cols-4 gap-2 text-xs">
        <li v-for="(s, i) in steps" :key="s" class="flex items-center gap-1.5" :class="i + 1 === stepNo ? 'text-foreground font-medium' : i + 1 < stepNo ? 'text-muted-foreground' : 'text-muted-foreground/60'">
          <span class="h-5 w-5 shrink-0 rounded-full grid place-items-center text-[11px]"
            :class="i + 1 < stepNo ? 'bg-foreground/80 text-background' : i + 1 === stepNo ? 'border-2 border-foreground' : 'border border-border'">
            <Check v-if="i + 1 < stepNo" class="w-3 h-3" /><template v-else>{{ i + 1 }}</template>
          </span>{{ s }}
        </li>
      </ol>

      <div class="px-6 py-5 min-h-[280px]">
        <!-- 1. Chọn máy -->
        <template v-if="phone.step === 'pick'">
          <p class="text-sm">Bạn dùng điện thoại nào?</p>
          <div class="mt-3 grid grid-cols-2 gap-3">
            <button class="rounded-lg border p-4 text-left hover:border-foreground/40 hover:bg-muted/50" :class="phone.device === 'ios' ? 'border-foreground/40' : 'border-border'" @click="pick('ios')">
              <Apple class="w-7 h-7" /><span class="mt-2 block font-medium">iPhone, iPad</span><span class="block text-xs text-muted-foreground">iOS</span>
            </button>
            <button class="rounded-lg border p-4 text-left hover:border-foreground/40 hover:bg-muted/50" :class="phone.device === 'android' ? 'border-foreground/40' : 'border-border'" @click="pick('android')">
              <Smartphone class="w-7 h-7" /><span class="mt-2 block font-medium">Android</span><span class="block text-xs text-muted-foreground">Samsung, Xiaomi, Oppo…</span>
            </button>
          </div>
          <p class="mt-4 text-xs text-muted-foreground">Sano tạo một file sách nói (định dạng M4B) cho cả cuốn: có bìa, mục lục chương. Bạn chép file này sang điện thoại rồi nghe bằng app miễn phí. Làm lần đầu mất khoảng 3 phút.</p>
        </template>

        <!-- 2. Cài app -->
        <template v-else-if="phone.step === 'app'">
          <p class="text-sm">Cài app nghe sách <b>BookPlayer</b> trên {{ isIOS ? 'iPhone' : 'điện thoại Android' }}. Miễn phí, mã nguồn mở.</p>
          <div class="mt-4 flex gap-5 items-center">
            <div class="h-36 w-36 shrink-0 rounded-lg border border-border grid place-items-center bg-white p-1.5">
              <img v-if="qr" :src="qr" :alt="`Mã QR tới BookPlayer trên ${store}`" class="h-full w-full" />
            </div>
            <div class="text-sm space-y-2">
              <p>Mở camera điện thoại, quét mã này để vào <b>{{ store }}</b>, rồi bấm <b>{{ isIOS ? 'Nhận' : 'Cài đặt' }}</b>.</p>
              <p class="text-muted-foreground text-xs">Hoặc tìm "BookPlayer" trên {{ store }}.</p>
              <button class="inline-flex items-center gap-1 text-xs text-primary hover:underline" @click="openURL(STORE_URL[phone.device ?? 'ios'])"><ExternalLink class="w-3 h-3" /> Mở trang BookPlayer trên {{ store }}</button>
            </div>
          </div>
          <p class="mt-5 text-xs text-muted-foreground">BookPlayer nhớ chỗ đang nghe, chỉnh tốc độ, hẹn giờ tắt, chạy trên {{ isIOS ? 'CarPlay' : 'Android Auto' }}.</p>
        </template>

        <!-- 3. Tạo file -->
        <template v-else-if="phone.step === 'make'">
          <div v-if="otherRunning" class="rounded-lg border border-border p-4 text-sm">
            Đang tạo file cho cuốn "{{ otherRunning.title }}". Đợi xong hoặc huỷ rồi thử lại.
            <div class="mt-2 flex gap-2"><Button size="sm" variant="outline" @click="cancelM4B">Huỷ cuốn kia</Button><Button size="sm" @click="make()">Thử lại</Button></div>
          </div>
          <div v-else-if="st?.running" class="rounded-lg border border-border p-4" role="status" aria-live="polite">
            <div class="flex items-center gap-2 text-sm"><Loader2 class="w-4 h-4 animate-spin text-primary" /> Đang tạo file cho điện thoại… {{ pct }}%</div>
            <div class="mt-2 h-1.5 rounded-full bg-muted overflow-hidden"><div class="h-full bg-primary rounded-full transition-[width]" :style="{ width: pct + '%' }"></div></div>
            <div class="mt-1.5 flex justify-between text-xs text-muted-foreground">
              <span>{{ st.progress.phase === 'mux' ? 'ghép mục lục, bìa…' : `tiểu mục ${st.progress.track}/${st.progress.tracks}` }}</span>
              <button class="text-foreground hover:underline" @click="cancelM4B">Huỷ</button>
            </div>
          </div>
          <div v-else class="rounded-lg border border-border p-4 text-sm">
            <p v-if="st?.error" class="text-destructive whitespace-pre-wrap break-words max-h-32 overflow-auto">Tạo file không thành công: {{ st.error }}</p>
            <p v-else-if="clickError" class="text-destructive">{{ clickError }}</p>
            <p v-else-if="st?.cancelled" class="text-muted-foreground">Đã huỷ tạo file.</p>
            <p v-else class="text-muted-foreground">Chưa tạo file.</p>
            <Button size="sm" class="mt-3" :disabled="m4b.asking" @click="make()"><Download class="w-4 h-4" /> Tạo file</Button>
          </div>
          <ul class="mt-4 space-y-1.5 text-xs text-muted-foreground">
            <li>
              Lưu vào: <b class="text-foreground" :title="st?.running ? st.path : ''">{{ whereShort }}</b><template v-if="sizeGuess"> · khoảng {{ sizeGuess }}</template>
              <button class="text-primary ml-1 hover:underline disabled:opacity-50" :disabled="m4b.asking" @click="changeLocation">Đổi chỗ lưu</button>
            </li>
            <li class="flex items-center gap-1"><Timer class="w-3.5 h-3.5 shrink-0" /> {{ pauseLine }}</li>
          </ul>
        </template>

        <!-- 4. Chép sang máy -->
        <template v-else>
          <div class="flex items-start gap-2 text-sm">
            <Check class="w-4 h-4 text-rag-green shrink-0 mt-0.5" />
            <span>Đã tạo <b>{{ fileName }}</b><template v-if="st"> ({{ fmtLong(st.durationSec) }}, {{ fmtSize(st.size) }})</template>.</span>
          </div>
          <template v-if="isIOS">
            <p class="mt-4 text-sm font-medium">Chép sang iPhone, chọn một cách:</p>
            <div class="mt-2 grid gap-3 text-sm" :class="airdrop ? 'grid-cols-2' : 'grid-cols-1'">
              <div v-if="airdrop" class="rounded-lg border border-border p-3 flex flex-col">
                <p class="font-medium flex items-center gap-1.5"><Share2 class="w-4 h-4" /> AirDrop</p>
                <ol class="mt-1.5 list-decimal pl-4 space-y-0.5 text-xs text-muted-foreground flex-1">
                  <li>Bấm <b class="text-foreground">Gửi bằng AirDrop</b> dưới đây, chọn iPhone</li>
                  <li>Trên iPhone chọn <b class="text-foreground">Lưu vào Tệp</b></li>
                </ol>
                <Button size="sm" variant="outline" class="mt-2 w-full" @click="sendAirDrop"><Share2 class="w-3.5 h-3.5" /> Gửi bằng AirDrop</Button>
              </div>
              <div class="rounded-lg border border-border p-3 flex flex-col">
                <p class="font-medium flex items-center gap-1.5"><MessageCircle class="w-4 h-4" /> Zalo, iCloud, Google Drive</p>
                <ol class="mt-1.5 list-decimal pl-4 space-y-0.5 text-xs text-muted-foreground flex-1">
                  <li>Gửi file cho chính bạn (Zalo "Cloud của tôi") hoặc tải lên iCloud Drive, Google Drive</li>
                  <li>Trên iPhone mở file → <b class="text-foreground">Chia sẻ</b> → <b class="text-foreground">Lưu vào Tệp</b></li>
                </ol>
                <Button size="sm" variant="outline" class="mt-2 w-full" @click="reveal"><FolderOpen class="w-3.5 h-3.5" /> Mở thư mục chứa file</Button>
              </div>
            </div>
            <p class="mt-4 text-sm"><b>Cuối cùng:</b> mở BookPlayer → bấm <b>+</b> → <b>Nhập tệp</b> (Import files) → chọn file vừa lưu. Xong, bấm nghe.</p>
          </template>
          <template v-else>
            <p class="mt-4 text-sm font-medium">Chép sang điện thoại Android, chọn một cách:</p>
            <div class="mt-2 grid grid-cols-3 gap-3 text-xs text-muted-foreground">
              <div class="rounded-lg border border-border p-3"><p class="text-sm font-medium text-foreground flex items-center gap-1.5"><MessageCircle class="w-4 h-4" /> Zalo</p>Gửi file vào "Cloud của tôi", trên điện thoại bấm file → <b class="text-foreground">Tải về</b>.</div>
              <div class="rounded-lg border border-border p-3"><p class="text-sm font-medium text-foreground flex items-center gap-1.5"><HardDrive class="w-4 h-4" /> Google Drive</p>Tải file lên Drive, trên điện thoại mở Drive → <b class="text-foreground">Tải xuống</b>.</div>
              <div class="rounded-lg border border-border p-3"><p class="text-sm font-medium text-foreground flex items-center gap-1.5"><Usb class="w-4 h-4" /> Cáp USB</p>Cắm máy, chọn <b class="text-foreground">Truyền tệp</b>, chép vào thư mục <b class="text-foreground">Audiobooks</b>.</div>
            </div>
            <Button size="sm" variant="outline" class="mt-3" @click="reveal"><FolderOpen class="w-3.5 h-3.5" /> Mở thư mục chứa file</Button>
            <p class="mt-4 text-sm"><b>Cuối cùng:</b> mở BookPlayer → thêm sách → chọn file vừa tải (thường trong <b>Download</b>). Xong, bấm nghe.</p>
          </template>
          <p v-if="actionError" class="mt-3 text-sm text-destructive">{{ actionError }}</p>
        </template>
      </div>

      <div class="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
        <button class="text-xs text-muted-foreground inline-flex items-center gap-1 hover:text-foreground" @click="openURL(`${DOCS}/nghe-tren-dien-thoai`)"><ExternalLink class="w-3 h-3" /> Hướng dẫn có ảnh</button>
        <div class="flex gap-2">
          <Button v-if="phone.step === 'pick' && phone.device" size="sm" @click="pick(phone.device)">Tiếp <ArrowRight class="w-4 h-4" /></Button>
          <Button v-if="phone.step === 'app'" variant="outline" size="sm" @click="phone.step = 'pick'"><ArrowLeft class="w-4 h-4" /> Quay lại</Button>
          <Button v-if="phone.step === 'app'" size="sm" @click="make()"><Download class="w-4 h-4" /> Đã cài, tạo file</Button>
          <Button v-if="phone.step === 'make' && st?.running" variant="outline" size="sm" @click="runInBackground">Để chạy nền</Button>
          <Button v-if="phone.step === 'copy'" size="sm" @click="closePhone">Xong <ArrowRight class="w-4 h-4" /></Button>
        </div>
      </div>
    </div>
  </div>
</template>
