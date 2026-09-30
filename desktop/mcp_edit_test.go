package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func waitEdit(t *testing.T, a *App) {
	t.Helper()
	for i := 0; i < 1800 && a.editing() != nil; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if a.editing() != nil {
		t.Fatal("quá 3 phút vẫn đang sửa")
	}
}

func TestMCP_SuaSachVaHoanTac(t *testing.T) {
	a, srv, token := mcpTestServer(t)
	cs := mcpConnect(t, srv, token)
	var books mcpBooks
	callTool(t, cs, "list_books", map[string]any{}, &books)
	slug := books.Books[0].Slug
	info, _ := a.lib.Edit(slug)
	var sec = info.Sections[len(info.Sections)-1]
	word := strings.Fields(sec.Text)[0]

	// Tìm (xem trước): không thay gì, không giữ bản cũ.
	var fr mcpFindReplaceOut
	callTool(t, cs, "find_replace", map[string]any{"slug": slug, "find": word, "replace": word + "X", "preview": true}, &fr)
	if fr.Matches == 0 || len(fr.Hits) == 0 || fr.Hits[0].Excerpt == "" {
		t.Fatalf("xem trước phải thấy chỗ khớp: %+v", fr)
	}
	if n := len(a.lib.Snapshots(slug)); n != 0 {
		t.Fatalf("xem trước không được giữ bản cũ: %d", n)
	}

	// Sửa thông tin (không cần bộ đọc nếu sách không có lời giới thiệu tự tạo) + hoàn tác.
	var eo mcpEditOut
	callTool(t, cs, "update_book_info", map[string]any{"slug": slug, "author": "Tác giả mới"}, &eo)
	waitEdit(t, a)
	if d, _ := a.lib.Get(slug); d.Author != "Tác giả mới" {
		t.Fatalf("tác giả chưa đổi: %q", d.Author)
	}
	mi := a.MCPInfo()
	if !mi.Log[0].CanUndo || mi.Log[0].Slug != slug {
		t.Fatalf("dòng nhật ký sửa phải có Hoàn tác: %+v", mi.Log[0])
	}
	if _, err := a.MCPUndo(slug, mi.Log[0].Undo); err != nil {
		t.Fatal(err)
	}
	if d, _ := a.lib.Get(slug); d.Author == "Tác giả mới" {
		t.Fatal("hoàn tác phải trả tác giả cũ")
	}
	mi = a.MCPInfo()
	if mi.Log[0].Tool != "undo" || !mi.Log[1].Undone || mi.Log[1].CanUndo {
		t.Fatalf("sau hoàn tác: %+v", mi.Log[:2])
	}

	// Đổi bìa bằng ảnh base64.
	img := image.NewRGBA(image.Rect(0, 0, 30, 40))
	img.Set(1, 1, color.RGBA{200, 0, 0, 255})
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	callTool(t, cs, "set_cover", map[string]any{"slug": slug, "image_base64": "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())}, &eo)
	if res, _ := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "set_cover", Arguments: map[string]any{"slug": slug, "image_base64": base64.StdEncoding.EncodeToString([]byte("không phải ảnh"))}}); !res.IsError {
		t.Fatal("dữ liệu không phải ảnh phải bị từ chối")
	}

	// Sửa lời một mục rồi đọc lại (cần bộ đọc thật).
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "update_sections", Arguments: map[string]any{
		"slug": slug, "sections": []map[string]any{{"index": sec.Index + 1, "text": "Lời mới ngắn gọn để thử."}}}})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Skip("máy này không có bộ đọc:", res.Content)
	}
	waitEdit(t, a)
	after, _ := a.lib.Edit(slug)
	if got := after.Sections[sec.Index].Text; got != "Lời mới ngắn gọn để thử." {
		t.Fatalf("lời chưa đổi: %q (%+v)", got, a.EditStatus())
	}
	mi = a.MCPInfo()
	if _, err := a.MCPUndo(slug, mi.Log[0].Undo); err != nil {
		t.Fatal(err)
	}
	if back, _ := a.lib.Edit(slug); back.Sections[sec.Index].Text != sec.Text {
		t.Fatal("hoàn tác phải trả lời cũ")
	}
}
