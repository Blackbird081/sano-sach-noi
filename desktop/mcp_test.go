package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sano/desktop/internal/library"
)

type bearer struct{ token string }

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(r)
}

func mcpTestServer(t *testing.T) (*App, *httptest.Server, string) {
	t.Helper()
	a := &App{lib: library.New(t.TempDir())}
	if err := seedSampleBooks(a.lib); err != nil {
		t.Fatal(err)
	}
	const token = "0123456789abcdef0123456789abcdef"
	srv := httptest.NewServer(a.mcpHandler(token))
	t.Cleanup(srv.Close)
	return a, srv, token
}

func mcpConnect(t *testing.T, srv *httptest.Server, token string) *mcp.ClientSession {
	t.Helper()
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := c.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: srv.URL + "/mcp", HTTPClient: &http.Client{Transport: bearer{token}}, DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args any, out any) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s lỗi: %+v", name, res.Content)
	}
	b, _ := json.Marshal(res.StructuredContent)
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("%s: %v (%s)", name, err, b)
	}
}

func TestMCP_XemSach(t *testing.T) {
	_, srv, token := mcpTestServer(t)
	cs := mcpConnect(t, srv, token)

	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	readOnly := map[string]bool{"list_books": true, "get_book": true, "get_section_texts": true, "list_voices": true, "get_status": true, "get_render_status": true}
	for _, tl := range tools.Tools {
		names = append(names, tl.Name)
		if tl.Annotations == nil || tl.Annotations.ReadOnlyHint != readOnly[tl.Name] {
			t.Errorf("%s: đánh dấu chỉ đọc sai", tl.Name)
		}
		if strings.Contains(tl.Name, "delete") || strings.Contains(tl.Name, "remove") {
			t.Errorf("MCP không được có công cụ xoá: %s", tl.Name)
		}
		if tl.Annotations != nil && tl.Annotations.DestructiveHint != nil && *tl.Annotations.DestructiveHint {
			t.Errorf("%s không được là công cụ phá dữ liệu", tl.Name)
		}
	}
	if got := strings.Join(names, ","); !strings.Contains(got, "list_books") || !strings.Contains(got, "get_section_texts") {
		t.Fatalf("thiếu công cụ: %s", got)
	}

	var books mcpBooks
	callTool(t, cs, "list_books", map[string]any{}, &books)
	if len(books.Books) == 0 {
		t.Fatal("thư viện mẫu phải có sách")
	}
	slug := books.Books[0].Slug

	var d mcpBookDetail
	callTool(t, cs, "get_book", map[string]any{"slug": slug}, &d)
	if len(d.Sections) == 0 || d.Sections[0].Index != 1 {
		t.Fatalf("mục lục sai: %+v", d.Sections)
	}

	var tx mcpTexts
	callTool(t, cs, "get_section_texts", map[string]any{"slug": slug, "from": 1, "to": 1}, &tx)
	if len(tx.Sections) != 1 || tx.Sections[0].Text == "" {
		t.Fatalf("lời mục 1 trống: %+v", tx)
	}

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_book", Arguments: map[string]any{"slug": "../khong-co"}})
	if err != nil || !res.IsError {
		t.Fatalf("slug lạ phải báo lỗi công cụ: %v %+v", err, res)
	}
}

func TestMCP_ChanKhongMaVaTrangWeb(t *testing.T) {
	_, srv, token := mcpTestServer(t)
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	do := func(auth, origin string) int {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}
	if c := do("", ""); c != http.StatusUnauthorized {
		t.Errorf("không mã: %d", c)
	}
	if c := do("Bearer sai", ""); c != http.StatusUnauthorized {
		t.Errorf("sai mã: %d", c)
	}
	if c := do("Bearer "+token, "https://trang-la.example"); c != http.StatusForbidden {
		t.Errorf("trang web: %d", c)
	}
	if c := do("Bearer "+token, ""); c != http.StatusOK {
		t.Errorf("đúng mã: %d", c)
	}
}

const mcpSampleText = `Đây là lời chào của AI, phải bị bỏ.

% Sách thử MCP

# Chương 1. Mở đầu

## Mục một

Đây là đoạn văn đầu tiên để đọc thử.

## Mục hai

Đoạn văn thứ hai, câu ngắn, dễ nghe.

# Chương 2. Kết

## Ba ý cần nhớ

Một. Hai. Ba.
`

func TestMCP_TaoBanNhapVaCamKet(t *testing.T) {
	a, srv, token := mcpTestServer(t)
	cs := mcpConnect(t, srv, token)

	var d mcpDraftOut
	callTool(t, cs, "create_book", map[string]any{"text": mcpSampleText, "author": "Người thử"}, &d)
	if d.DraftID == "" || d.Title != "Sách thử MCP" || len(d.Chapters) != 2 || d.Sections != 3 || d.Voice != mcpDefaultVoice {
		t.Fatalf("bản nháp sai: %+v", d)
	}
	if !strings.Contains(d.IntroText, "Tác giả: Người thử.") {
		t.Fatalf("lời giới thiệu tự tạo sai: %q", d.IntroText)
	}
	a.mcp.mu.Lock()
	draft := a.mcp.drafts[d.DraftID]
	a.mcp.mu.Unlock()
	if draft == nil || draft.settings.RightsConfirmedAt != "" {
		t.Fatal("bản nháp không được có sẵn cam kết")
	}
	if !strings.HasPrefix(draft.docx, a.lib.Root()) {
		t.Fatalf("file tạm phải nằm trong thư viện: %s", draft.docx)
	}

	for _, bad := range []map[string]any{{"text": ""}, {"text": "không có dòng chương nào"}, {"text": "# Chương\n\n## Mục\n\nChữ"}} {
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "create_book", Arguments: bad})
		if err != nil || !res.IsError {
			t.Errorf("nội dung sai phải báo lỗi: %v %v", bad, err)
		}
	}

	var r mcpRenderOut
	callTool(t, cs, "start_render", map[string]any{"draft_id": d.DraftID}, &r)
	if r.Status != "waiting_pledge" {
		t.Fatalf("phải chờ cam kết: %+v", r)
	}
	if p := a.MCPPendingPledge(); p == nil || p.ID != d.DraftID || p.Sections != 3 {
		t.Fatalf("popup cam kết: %+v", p)
	}
	res, _ := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "start_render", Arguments: map[string]any{"draft_id": d.DraftID}})
	if !res.IsError {
		t.Fatal("đang chờ cam kết thì không mở popup thứ hai")
	}

	// Người dùng đóng popup → từ chối, bản nháp vẫn còn để hỏi lại.
	if err := a.MCPPledgeAnswer(d.DraftID, false); err != nil {
		t.Fatal(err)
	}
	callTool(t, cs, "get_render_status", map[string]any{}, &r)
	if r.Status != "declined" {
		t.Fatalf("sau khi từ chối: %+v", r)
	}
	if err := a.MCPPledgeAnswer(d.DraftID, true); err == nil {
		t.Fatal("trả lời popup đã đóng phải báo lỗi")
	}

	// Cam kết: máy có bộ đọc thì bắt đầu đọc (dừng ngay), không có thì báo lỗi rõ, không treo.
	callTool(t, cs, "start_render", map[string]any{"draft_id": d.DraftID}, &r)
	if err := a.MCPPledgeAnswer(d.DraftID, true); err == nil {
		st := a.RenderStatus()
		if st == nil || st.Source != "mcp" || st.Title != "Sách thử MCP" {
			t.Fatalf("lượt đọc phải ghi nguồn mcp: %+v", st)
		}
		callTool(t, cs, "cancel_render", map[string]any{}, &r)
		for i := 0; i < 300 && a.Rendering(); i++ {
			time.Sleep(100 * time.Millisecond)
		}
		callTool(t, cs, "get_render_status", map[string]any{}, &r)
		if r.Status != "cancelled" {
			t.Fatalf("sau khi dừng: %+v", r)
		}
		return
	}
	callTool(t, cs, "get_render_status", map[string]any{}, &r)
	if r.Status != "error" || r.Message == "" {
		t.Fatalf("lỗi render phải báo lại: %+v", r)
	}
}
