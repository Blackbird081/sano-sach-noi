// Kết nối AI qua MCP (mcp.go, mcp_queue.go): sách AI nhờ tạo — TẠO TRƯỚC, CAM KẾT SAU (wireframe D22).
// Các cuốn chờ tạo / đang tạo / ở khu chờ cam kết; popup cam kết tự mở khi có cuốn mới tạo xong.
// Tách khỏi state.render của luồng Tạo sách nói trong app.
import { computed, reactive } from 'vue'
import {
  errText, mcpCommit, mcpDiscard, mcpDismiss, mcpInfo, mcpJobs, onEvent, type MCPInfo, type MCPJob, type RenderStatus,
} from './backend'

export const mcp = reactive({
  jobs: [] as MCPJob[],
  popup: false,
  // Lựa chọn trong popup, giữ qua các lần mở: cuốn đã tick / bỏ tick. seen: cuốn người dùng đã thấy trong
  // popup — tạo xong không tự mở popup lại vì nó (bỏ tick có chủ ý thì không làm phiền).
  picks: {} as Record<string, boolean>,
  seen: {} as Record<string, boolean>,
  error: '',
  info: null as MCPInfo | null, // màn MCP
})

export const activeJobs = computed(() => mcp.jobs.filter((j) => j.status === 'queued' || j.status === 'rendering'))
export const stagedJobs = computed(() => mcp.jobs.filter((j) => j.status === 'staged'))
/** Các cuốn hiện trong popup: đã tạo xong (khu chờ), đang tạo, chờ tạo mà chưa cam kết. */
export const pledgeJobs = computed(() => mcp.jobs.filter((j) => j.status === 'staged' || ((j.status === 'queued' || j.status === 'rendering') && !j.autoSave)))

export async function refreshMCPInfo() {
  try {
    mcp.info = await mcpInfo()
  } catch {
    /* giữ bản cũ */
  }
}

let onSaved: () => void = () => {}

function applyJobs(jobs: MCPJob[]) {
  const before = new Map(mcp.jobs.map((j) => [j.id, j.status]))
  mcp.jobs = jobs ?? []
  // Có cuốn vừa vào khu chờ → mở popup cam kết. Có cuốn vừa lưu → tải lại Thư viện.
  if (mcp.jobs.some((j) => j.status === 'staged' && before.get(j.id) !== 'staged' && !mcp.seen[j.id])) mcp.popup = true
  if (mcp.jobs.some((j) => j.status === 'saved' && before.get(j.id) !== 'saved')) onSaved()
  if (!pledgeJobs.value.length) mcp.popup = false
}

export async function initMCP(refreshLibrary: () => void) {
  onSaved = refreshLibrary
  onEvent<MCPJob[]>('mcp:jobs', applyJobs)
  onEvent('mcp:log', () => void refreshMCPInfo())
  // Tiến độ cuốn đang tạo: lấy từ sự kiện render (nhanh hơn chờ mcp:jobs).
  onEvent<RenderStatus>('render:progress', (st) => {
    if (st.source !== 'mcp') return
    const j = mcp.jobs.find((x) => x.status === 'rendering')
    const p = st.progress
    if (j && p?.totalChars) j.percent = Math.min(100, Math.round((p.doneChars / p.totalChars) * 1000) / 10)
  })
  applyJobs(await mcpJobs())
}

/** Đóng popup: mọi cuốn đang hiện coi như đã thấy. */
export function closePledge() {
  for (const j of pledgeJobs.value) mcp.seen[j.id] = true
  mcp.popup = false
}

async function run(fn: () => Promise<MCPJob[]>) {
  mcp.error = ''
  try {
    applyJobs(await fn())
    return true
  } catch (e) {
    mcp.error = errText(e)
    return false
  }
}

/** Cam kết các cuốn đã tick: cuốn đã tạo xong vào Thư viện ngay, cuốn đang / chờ tạo tự lưu khi xong. */
export const commitJobs = (ids: string[]) => run(() => mcpCommit(ids))
export const discardJob = (id: string) => run(() => mcpDiscard(id))
export const dismissJobs = () => run(mcpDismiss)
