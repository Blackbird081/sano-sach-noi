// Kết nối AI qua MCP (mcp.go, mcp_create.go): popup cam kết khi AI nhờ đọc một cuốn, và
// tiến độ lượt đọc do AI bắt đầu (tách khỏi state.render của luồng Tạo sách nói trong app).
import { reactive } from 'vue'
import { errText, mcpPendingPledge, mcpPledgeAnswer, onEvent, type MCPPledge, type RenderStatus } from './backend'

export const mcp = reactive({
  pledge: null as MCPPledge | null,
  render: null as RenderStatus | null,
  error: '',
})

export async function initMCP() {
  onEvent<MCPPledge | null>('mcp:pledge', (p) => {
    mcp.pledge = p
    mcp.error = ''
  })
  mcp.pledge = await mcpPendingPledge()
}

/** Người dùng tick đủ cam kết (ok) hoặc đóng popup. */
export async function answerPledge(ok: boolean) {
  const p = mcp.pledge
  if (!p) return
  mcp.pledge = null
  try {
    await mcpPledgeAnswer(p.id, ok)
  } catch (e) {
    mcp.error = errText(e)
  }
}

export function mcpRenderPct(st: RenderStatus | null): number {
  const p = st?.progress
  if (!st || !p) return 0
  if (st.done) return 100
  return p.totalChars ? Math.min(100, Math.round((p.doneChars / p.totalChars) * 100)) : 0
}
