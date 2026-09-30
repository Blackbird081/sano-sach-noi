package main

// Màn MCP (wireframe D21): bật/tắt kết nối trong máy, quyền "Tạo và sửa sách", nhật ký việc
// AI làm (7 ngày, chỉ trên máy), lệnh thêm Sano vào Claude Code / Codex, thêm vào Claude Desktop.
// Cài đặt nằm trong ~/Sano/.mcp/settings.json, nhật ký ~/Sano/.mcp/log.jsonl.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"sano/desktop/internal/mcplink"
)

const (
	eventMCPLog = "mcp:log" // có dòng nhật ký mới → màn MCP tải lại

	mcpLogKeep     = 7 * 24 * time.Hour
	mcpLogMax      = 300
	mcpActiveSince = 15 * time.Minute // "đang dùng" = có gọi trong 15 phút gần đây
)

// mcpSettings — lựa chọn của người dùng ở màn MCP. Mặc định: trong máy bật, được tạo & sửa.
type mcpSettings struct {
	LocalOff bool `json:"localOff"` // tắt kết nối trong máy
	NoEdit   bool `json:"noEdit"`   // bỏ quyền Tạo và sửa sách
}

// MCPLogEntry — một việc AI đã làm.
type MCPLogEntry struct {
	At     int64  `json:"at"`     // unix giây
	Client string `json:"client"` // Claude Code, Claude Desktop, Codex… (không rõ = "AI")
	Tool   string `json:"tool"`
	Text   string `json:"text"` // mô tả tiếng Việt
	Edit   bool   `json:"edit"` // việc tạo / sửa (không phải chỉ xem)
	Error  string `json:"error,omitempty"`
	Slug   string `json:"slug,omitempty"` // cuốn bị sửa
	Undo   string `json:"undo,omitempty"` // mã bản cũ giữ trước khi sửa (library.Snapshot)
	// Chỉ có trong MCPInfo (không ghi file): còn hoàn tác được / đã hoàn tác.
	CanUndo bool `json:"canUndo,omitempty"`
	Undone  bool `json:"undone,omitempty"`
}

// MCPInfo — dữ liệu cho màn MCP.
type MCPInfo struct {
	LocalOn       bool          `json:"localOn"`
	Running       bool          `json:"running"` // máy chủ đang nghe (bật mà mở cổng lỗi thì false)
	AllowEdit     bool          `json:"allowEdit"`
	Bridge        string        `json:"bridge"`      // đường dẫn sano-mcp
	BridgeFound   bool          `json:"bridgeFound"` // có file sano-mcp cạnh app (bản dev thường chưa có)
	ClaudeCode    string        `json:"claudeCode"`  // lệnh thêm vào Claude Code
	Codex         string        `json:"codex"`
	ConfigJSON    string        `json:"configJson"`    // khối cấu hình cho phần mềm khác
	ClaudeDesktop string        `json:"claudeDesktop"` // "" = chưa thấy Claude Desktop | "found" | "added"
	Active        []string      `json:"active"`        // phần mềm AI dùng trong 15 phút gần đây
	Log           []MCPLogEntry `json:"log"`           // mới nhất trước
}

var mcpFileMu sync.Mutex // settings.json + log.jsonl

func (a *App) mcpSettingsPath() string {
	return filepath.Join(a.lib.Root(), mcplink.Dir, "settings.json")
}
func (a *App) mcpLogPath() string { return filepath.Join(a.lib.Root(), mcplink.Dir, "log.jsonl") }

func (a *App) loadMCPSettings() mcpSettings {
	mcpFileMu.Lock()
	defer mcpFileMu.Unlock()
	var s mcpSettings
	if b, err := os.ReadFile(a.mcpSettingsPath()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	return s
}

func (a *App) saveMCPSettings(s mcpSettings) error {
	mcpFileMu.Lock()
	defer mcpFileMu.Unlock()
	b, _ := json.Marshal(s)
	if err := os.MkdirAll(filepath.Dir(a.mcpSettingsPath()), 0o700); err != nil {
		return err
	}
	return os.WriteFile(a.mcpSettingsPath(), b, 0o600)
}

// SetMCPLocal bật / tắt kết nối trong máy (tắt là đóng cổng ngay, AI đang nối nhận lỗi "chưa mở").
func (a *App) SetMCPLocal(on bool) (*MCPInfo, error) {
	s := a.loadMCPSettings()
	s.LocalOff = !on
	if err := a.saveMCPSettings(s); err != nil {
		return nil, err
	}
	if on {
		a.startMCP()
	} else {
		a.stopMCP()
	}
	return a.MCPInfo(), nil
}

// SetMCPAllowEdit bật / tắt quyền Tạo và sửa sách của AI.
func (a *App) SetMCPAllowEdit(on bool) (*MCPInfo, error) {
	s := a.loadMCPSettings()
	s.NoEdit = !on
	if err := a.saveMCPSettings(s); err != nil {
		return nil, err
	}
	return a.MCPInfo(), nil
}

var errMCPNoEdit = errors.New("người dùng đã tắt quyền \"Tạo và sửa sách\" trong Sano (mục MCP). Nhờ họ bật lại nếu muốn tiếp tục")

// mcpBridge — lệnh phần mềm AI chạy để nối vào Sano: sano-mcp cạnh file chạy của app
// (Sano.app/Contents/MacOS, thư mục cài Windows); Linux AppImage thì chính file .AppImage với
// tham số "mcp" (đường dẫn bên trong AppImage đổi mỗi lần chạy). Không có sano-mcp (bản dev) thì
// macOS / Linux dùng "Sano mcp"; Windows cần sano-mcp.exe (app giao diện không nhận stdio ổn định).
func mcpBridge() (cmd string, args []string, found bool) {
	if runtime.GOOS == "linux" {
		if p := os.Getenv("APPIMAGE"); p != "" {
			if _, err := os.Stat(p); err == nil {
				return p, []string{"mcp"}, true
			}
		}
	}
	name := "sano-mcp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe, err := os.Executable()
	if err != nil {
		return name, nil, false
	}
	if r, err := filepath.EvalSymlinks(exe); err == nil {
		exe = r
	}
	p := filepath.Join(filepath.Dir(exe), name)
	if _, err := os.Stat(p); err == nil {
		return p, nil, true
	}
	if runtime.GOOS == "windows" {
		return p, nil, false
	}
	return exe, []string{"mcp"}, true
}

func shellQuote(p string) string {
	if strings.ContainsAny(p, " '\"()&;") {
		if runtime.GOOS == "windows" {
			return `"` + p + `"`
		}
		return "'" + strings.ReplaceAll(p, "'", `'\''`) + "'"
	}
	return p
}

// MCPInfo trả trạng thái cho màn MCP.
func (a *App) MCPInfo() *MCPInfo {
	s := a.loadMCPSettings()
	bridge, args, found := mcpBridge()
	cfg, _ := json.MarshalIndent(map[string]any{"mcpServers": map[string]any{"sano": bridgeEntry(bridge, args)}}, "", "  ")
	cmdLine := shellQuote(bridge)
	for _, a := range args {
		cmdLine += " " + a
	}
	info := &MCPInfo{
		LocalOn: !s.LocalOff, AllowEdit: !s.NoEdit, Bridge: bridge, BridgeFound: found,
		ClaudeCode: "claude mcp add sano -- " + cmdLine,
		Codex:      "codex mcp add sano -- " + cmdLine,
		ConfigJSON: string(cfg),
		Active:     []string{},
	}
	a.mu.Lock()
	info.Running = a.mcpSrv != nil
	a.mu.Unlock()
	info.ClaudeDesktop = claudeDesktopState(bridge, args)
	info.Log = a.readMCPLog()
	a.markUndo(info.Log)
	seen := map[string]bool{}
	for _, e := range info.Log {
		if time.Since(time.Unix(e.At, 0)) > mcpActiveSince {
			break
		}
		if !seen[e.Client] {
			seen[e.Client] = true
			info.Active = append(info.Active, e.Client)
		}
	}
	if len(info.Log) > 50 {
		info.Log = info.Log[:50]
	}
	return info
}

// markUndo đánh dấu dòng nhật ký nào còn hoàn tác được (bản cũ còn, chưa hoàn tác, là bản cũ
// mới nhất của cuốn đó — hoàn tác bản cũ hơn sẽ mất các lần sửa sau nên không cho).
func (a *App) markUndo(log []MCPLogEntry) {
	undone := map[string]bool{}
	for _, e := range log {
		if e.Tool == "undo" && e.Undo != "" {
			undone[e.Undo] = true
		}
	}
	latest := map[string]string{}
	for i := range log {
		e := &log[i]
		if e.Undo == "" || e.Tool == "undo" {
			continue
		}
		if undone[e.Undo] {
			e.Undone = true
			continue
		}
		if _, seen := latest[e.Slug]; seen {
			continue // log mới nhất trước: chỉ dòng đầu tiên của mỗi cuốn
		}
		latest[e.Slug] = e.Undo
		for _, s := range a.lib.Snapshots(e.Slug) {
			if s.ID == e.Undo {
				e.CanUndo = true
			}
		}
	}
}

// ── Nhật ký ──────────────────────────────────────────────────────────────

func (a *App) logMCP(e MCPLogEntry) {
	e.At = time.Now().Unix()
	e.Text = truncRunes(e.Text, 200)
	e.Error = truncRunes(e.Error, 200)
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	mcpFileMu.Lock()
	path := a.mcpLogPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err == nil {
		if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
			_, _ = f.Write(append(b, '\n'))
			_ = f.Close()
		}
	}
	mcpFileMu.Unlock()
	a.pruneMCPLog()
	a.emit(eventMCPLog, e)
}

// readMCPLog — các dòng còn hạn, mới nhất trước.
func (a *App) readMCPLog() []MCPLogEntry {
	mcpFileMu.Lock()
	defer mcpFileMu.Unlock()
	return readLogFile(a.mcpLogPath())
}

func readLogFile(path string) []MCPLogEntry {
	out := []MCPLogEntry{}
	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 4096), 64<<10)
	cutoff := time.Now().Add(-mcpLogKeep).Unix()
	for sc.Scan() {
		var e MCPLogEntry
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.At >= cutoff {
			out = append(out, e)
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// pruneMCPLog giữ tối đa mcpLogMax dòng trong 7 ngày (viết lại file khi dài gấp đôi).
func (a *App) pruneMCPLog() {
	mcpFileMu.Lock()
	defer mcpFileMu.Unlock()
	path := a.mcpLogPath()
	st, err := os.Stat(path)
	if err != nil || st.Size() < 64<<10 {
		return
	}
	entries := readLogFile(path)
	if len(entries) > mcpLogMax {
		entries = entries[:mcpLogMax]
	}
	var buf bytes.Buffer
	for i := len(entries) - 1; i >= 0; i-- {
		b, _ := json.Marshal(entries[i])
		buf.Write(append(b, '\n'))
	}
	_ = os.WriteFile(path, buf.Bytes(), 0o600)
}

func truncRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

// mcpClientName — tên phần mềm AI (cầu nối gửi X-Sano-Client lấy từ clientInfo lúc initialize).
func mcpClientName(raw, userAgent string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	if n == "" {
		n = strings.ToLower(userAgent)
	}
	switch {
	case strings.Contains(n, "claude-code") || strings.Contains(n, "claude code"):
		return "Claude Code"
	case strings.Contains(n, "claude"):
		return "Claude Desktop"
	case strings.Contains(n, "codex"):
		return "Codex"
	case strings.Contains(n, "cursor"):
		return "Cursor"
	case strings.Contains(n, "chatgpt") || strings.Contains(n, "openai"):
		return "ChatGPT"
	}
	name := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, strings.TrimSpace(raw))
	if name == "" {
		return "AI"
	}
	return truncRunes(name, 40)
}

// ── Claude Desktop ───────────────────────────────────────────────────────

// claudeDesktopConfig — file cấu hình của Claude Desktop trên máy này.
func claudeDesktopConfig() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		if d := os.Getenv("APPDATA"); d != "" {
			return filepath.Join(d, "Claude", "claude_desktop_config.json")
		}
		return ""
	default:
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}
}

// bridgeEntry — mục "sano" trong cấu hình MCP của phần mềm AI.
func bridgeEntry(cmd string, args []string) map[string]any {
	m := map[string]any{"command": cmd}
	if len(args) > 0 {
		m["args"] = args
	}
	return m
}

// claudeDesktopState: "" chưa thấy Claude Desktop, "found" có mà chưa thêm Sano, "added" đã thêm đúng cầu nối.
func claudeDesktopState(bridge string, args []string) string {
	p := claudeDesktopConfig()
	if p == "" {
		return ""
	}
	if _, err := os.Stat(filepath.Dir(p)); err != nil {
		return ""
	}
	cfg, err := readJSONObject(p)
	if err != nil {
		return "found"
	}
	if servers, ok := cfg["mcpServers"].(map[string]any); ok {
		if s, ok := servers["sano"].(map[string]any); ok && s["command"] == bridge {
			got, _ := json.Marshal(s["args"])
			want, _ := json.Marshal(args)
			if len(args) == 0 && s["args"] == nil || string(got) == string(want) {
				return "added"
			}
		}
	}
	return "found"
}

func readJSONObject(p string) (map[string]any, error) {
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// AddToClaudeDesktop thêm "sano" vào mcpServers của Claude Desktop, giữ nguyên mọi cấu hình khác;
// sao lưu file cũ thành claude_desktop_config.json.bak-sano trước khi ghi.
func (a *App) AddToClaudeDesktop() (*MCPInfo, error) {
	bridge, args, found := mcpBridge()
	if !found {
		return nil, errors.New("không thấy sano-mcp cạnh app Sano (bản đang chạy chưa kèm cầu nối MCP)")
	}
	if err := addClaudeDesktopEntry(bridge, args); err != nil {
		return nil, err
	}
	return a.MCPInfo(), nil
}

func addClaudeDesktopEntry(bridge string, args []string) error {
	p := claudeDesktopConfig()
	if p == "" {
		return errors.New("không tìm thấy thư mục cấu hình Claude Desktop")
	}
	if _, err := os.Stat(filepath.Dir(p)); err != nil {
		return errors.New("chưa thấy Claude Desktop trên máy này: cài Claude Desktop rồi thử lại")
	}
	cfg, err := readJSONObject(p)
	if err != nil {
		return fmt.Errorf("file cấu hình Claude Desktop đang lỗi, không dám sửa (%v): %s", err, p)
	}
	if old, err := os.ReadFile(p); err == nil {
		if err := os.WriteFile(p+".bak-sano", old, 0o600); err != nil {
			return fmt.Errorf("sao lưu cấu hình Claude Desktop: %w", err)
		}
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers["sano"] = bridgeEntry(bridge, args)
	cfg["mcpServers"] = servers
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp-sano"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
