<script setup lang="ts">
// Khung cửa sổ Sano (bám wireframe WfDesktop): thanh tiêu đề, màn cài bộ đọc
// hoặc thanh bên + nội dung, hộp cập nhật phủ trên cùng.
import { onMounted, ref } from 'vue'
import { platform } from './lib/backend'
import { init, state } from './lib/store'
import TitleBar from './components/TitleBar.vue'
import AppSidebar from './components/AppSidebar.vue'
import UpdateDialog from './components/UpdateDialog.vue'
import PhoneDialog from './components/PhoneDialog.vue'
import ShareDialog from './components/ShareDialog.vue'
import BookVideoDialog from './components/BookVideoDialog.vue'
import SetupView from './views/SetupView.vue'
import TermsView from './views/TermsView.vue'
import LibraryView from './views/LibraryView.vue'
import PlayerView from './views/PlayerView.vue'
import SettingsView from './views/SettingsView.vue'
import McpView from './views/McpView.vue'
import AboutView from './views/AboutView.vue'
import StatsView from './views/StatsView.vue'
import CreateView from './views/CreateView.vue'
import EditBookView from './views/EditBookView.vue'
import { initEdit } from './lib/edit'
import RenderPledgeDialog from './components/RenderPledgeDialog.vue'
import { answerPledge, initMCP, mcp } from './lib/mcp'
import MiniPlayer from './components/MiniPlayer.vue'
import { player } from './lib/player'

const os = ref('')
onMounted(async () => {
  os.value = await platform()
  await init()
  void initEdit()
  void initMCP()
})
</script>

<template>
  <div class="relative h-dvh w-full overflow-hidden bg-background text-foreground flex flex-col">
    <TitleBar :os="os" />

    <SetupView v-if="state.view === 'setup'" />
    <TermsView v-else-if="state.view === 'terms'" />

    <div v-else class="flex-1 flex min-h-0">
      <AppSidebar />
      <main class="flex-1 min-w-0 flex flex-col">
        <LibraryView v-if="state.view === 'library'" />
        <PlayerView v-else-if="state.view === 'player'" />
        <StatsView v-else-if="state.view === 'stats'" />
        <SettingsView v-else-if="state.view === 'settings'" />
        <McpView v-else-if="state.view === 'mcp'" />
        <AboutView v-else-if="state.view === 'about'" />
        <EditBookView v-else-if="state.view === 'edit'" />
        <CreateView v-else />
        <MiniPlayer v-if="player.slug && state.view !== 'player'" />
      </main>
    </div>

    <UpdateDialog v-if="state.update !== 'closed'" />
    <PhoneDialog />
    <ShareDialog />
    <BookVideoDialog />
    <!-- AI nhờ đọc một cuốn qua MCP: người ngồi trước máy tick cam kết D19 mới bắt đầu -->
    <RenderPledgeDialog v-if="mcp.pledge" :key="mcp.pledge.id" :title="`${mcp.pledge.client || 'AI'} nhờ tạo sách nói`" label="Sách AI gửi" action="Cam kết và đọc" cancel="Không đọc" strict
      :file-name="`${mcp.pledge.title} · ${mcp.pledge.sections} mục · khoảng ${mcp.pledge.listenMin} phút nghe · giọng ${mcp.pledge.voice}`"
      @confirm="answerPledge(true)" @close="answerPledge(false)" />
  </div>
</template>
