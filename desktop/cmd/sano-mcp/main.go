// sano-mcp — cầu nối MCP cho Claude Desktop, Claude Code, Codex… (xem internal/mcpbridge).
// Nằm cạnh file chạy của app: Sano.app/Contents/MacOS, thư mục cài Windows.
package main

import (
	"os"

	"sano/desktop/internal/mcpbridge"
)

func main() {
	os.Exit(mcpbridge.Main())
}
