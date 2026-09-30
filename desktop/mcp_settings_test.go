package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCP_TatQuyenSuaVaNhatKy(t *testing.T) {
	a, srv, token := mcpTestServer(t)
	cs := mcpConnect(t, srv, token)

	var books mcpBooks
	callTool(t, cs, "list_books", map[string]any{}, &books)
	var d mcpBookDetail
	callTool(t, cs, "get_book", map[string]any{"slug": books.Books[0].Slug}, &d)

	if _, err := a.SetMCPAllowEdit(false); err != nil {
		t.Fatal(err)
	}
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "create_book", Arguments: map[string]any{"text": mcpSampleText}})
	if err != nil || !res.IsError {
		t.Fatal("tắt quyền Tạo và sửa thì create_book phải bị từ chối")
	}
	// Xem vẫn được.
	callTool(t, cs, "list_books", map[string]any{}, &books)

	info := a.MCPInfo()
	if info.AllowEdit || !info.LocalOn {
		t.Fatalf("cài đặt: %+v", info)
	}
	if len(info.Log) < 3 {
		t.Fatalf("nhật ký thiếu: %+v", info.Log)
	}
	if e := info.Log[1]; e.Tool != "create_book" || e.Error == "" || !e.Edit {
		t.Fatalf("dòng bị từ chối phải ghi lỗi: %+v", e)
	}
	if !strings.Contains(info.Log[2].Text, d.Title) {
		t.Fatalf("nhật ký xem mục lục phải có tên sách: %+v", info.Log[2])
	}
	if len(info.Active) != 1 || info.Active[0] != "AI" {
		t.Fatalf("phần mềm đang dùng: %+v", info.Active)
	}
	st, _ := os.Stat(a.mcpLogPath())
	if runtime.GOOS != "windows" && st.Mode().Perm() != 0o600 {
		t.Fatalf("quyền file nhật ký: %v", st.Mode().Perm())
	}
}

func TestMCPClientName(t *testing.T) {
	cases := map[[2]string]string{
		{"claude-code", ""}:      "Claude Code",
		{"claude-ai", ""}:        "Claude Desktop",
		{"codex-mcp-client", ""}: "Codex",
		{"", "Cursor/1.0"}:       "Cursor",
		{"", ""}:                 "AI",
		{"Zed\x00\n", ""}:        "Zed",
	}
	for in, want := range cases {
		if got := mcpClientName(in[0], in[1]); got != want {
			t.Errorf("%q → %q, muốn %q", in, got, want)
		}
	}
}

func TestAddToClaudeDesktop_GiuCauHinhCu(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", t.TempDir())
	} else {
		t.Setenv("HOME", t.TempDir())
	}
	p := claudeDesktopConfig()
	bridge, args := "/Ung dung/Sano.AppImage", []string{"mcp"}
	if claudeDesktopState(bridge, args) != "" {
		t.Fatal("chưa có thư mục Claude thì coi như chưa cài")
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	old := `{"theme":"dark","mcpServers":{"khac":{"command":"/bin/khac"}}}`
	if err := os.WriteFile(p, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if claudeDesktopState(bridge, args) != "found" {
		t.Fatal("có Claude Desktop mà chưa thêm Sano")
	}
	if err := addClaudeDesktopEntry(bridge, args); err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	b, _ := os.ReadFile(p)
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	servers := cfg["mcpServers"].(map[string]any)
	sano := servers["sano"].(map[string]any)
	if cfg["theme"] != "dark" || servers["khac"] == nil || sano["command"] != bridge || len(sano["args"].([]any)) != 1 {
		t.Fatalf("cấu hình sau khi thêm: %s", b)
	}
	if bak, _ := os.ReadFile(p + ".bak-sano"); string(bak) != old {
		t.Fatal("phải sao lưu file cũ")
	}
	if claudeDesktopState(bridge, args) != "added" {
		t.Fatal("phải nhận ra đã thêm")
	}
	if claudeDesktopState(bridge, nil) != "found" {
		t.Fatal("khác tham số thì chưa phải cấu hình đúng")
	}

	// File cấu hình hỏng: không dám sửa.
	_ = os.WriteFile(p, []byte("{hỏng"), 0o644)
	if err := addClaudeDesktopEntry(bridge, args); err == nil {
		t.Fatal("file hỏng phải báo lỗi")
	}
}
