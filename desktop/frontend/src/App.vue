<script setup lang="ts">
// Khung cửa sổ Sano (bám wireframe WfDesktop): thanh tiêu đề, màn cài bộ đọc
// hoặc thanh bên + nội dung, hộp cập nhật phủ trên cùng.
import { onMounted, ref } from 'vue'
import { platform } from './lib/backend'
import { init, refreshLibrary, state } from './lib/store'
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
import McpPledgeDialog from './components/McpPledgeDialog.vue'
import { initMCP, mcp } from './lib/mcp'
import MiniPlayer from './components/MiniPlayer.vue'
import { player } from './lib/player'

const os = ref('')
onMounted(async () => {
  os.value = await platform()
  await init()
  void initEdit()
  void initMCP(() => void refreshLibrary())
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
    <!-- Sách AI tạo qua MCP: tạo trước, cam kết sau, một popup cho nhiều cuốn (D22) -->
    <McpPledgeDialog v-if="mcp.popup" />
  </div>
</template>
