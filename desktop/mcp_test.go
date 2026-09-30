package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	// Chưa xin tạo thì không có gì ở hàng đợi / khu chờ / Thư viện.
	if jobs := a.MCPJobs(); len(jobs) != 0 {
		t.Fatalf("chưa start_render mà đã có việc: %+v", jobs)
	}
}

// waitJobs chờ tới khi không còn cuốn nào chờ tạo / đang tạo.
func waitJobs(t *testing.T, a *App) []MCPJob {
	t.Helper()
	for i := 0; i < 1800; i++ {
		busy := false
		for _, j := range a.MCPJobs() {
			busy = busy || j.Status == jobQueued || j.Status == jobRendering
		}
		if !busy {
			return a.MCPJobs()
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("quá 3 phút vẫn đang tạo")
	return nil
}

func TestMCP_TaoTruocCamKetSau(t *testing.T) {
	a, srv, token := mcpTestServer(t)
	cs := mcpConnect(t, srv, token)
	draft := func(title string) string {
		var d mcpDraftOut
		callTool(t, cs, "create_book", map[string]any{"text": strings.Replace(mcpSampleText, "Sách thử MCP", title, 1), "no_intro": true}, &d)
		return d.DraftID
	}
	id1, id2, id3 := draft("Cuốn một"), draft("Cuốn hai"), draft("Cuốn ba")

	var r mcpRenderOut
	callTool(t, cs, "start_render", map[string]any{"draft_id": id1}, &r)
	callTool(t, cs, "start_render", map[string]any{"draft_id": id2}, &r)
	callTool(t, cs, "start_render", map[string]any{"draft_id": id3}, &r)
	if res, _ := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "start_render", Arguments: map[string]any{"draft_id": id1}}); !res.IsError {
		t.Fatal("bản nháp đã xếp hàng không xin tạo lần hai được")
	}
	jobs := a.MCPJobs()
	if len(jobs) != 3 || jobs[0].Status != jobRendering && jobs[0].Status != jobFailed {
		t.Fatalf("cuốn đầu phải đang tạo ngay: %+v", jobs)
	}
	if jobs[0].Status == jobFailed {
		t.Skip("máy này không có bộ đọc:", jobs[0].Error)
	}
	if jobs[1].Status != jobQueued || jobs[2].Status != jobQueued {
		t.Fatalf("các cuốn sau phải chờ tạo lần lượt: %+v", jobs)
	}
	// Cam kết trước cho cuốn ba khi nó còn chờ tạo → xong tự vào Thư viện.
	if _, err := a.MCPCommit([]string{id3}); err != nil {
		t.Fatal(err)
	}

	jobs = waitJobs(t, a)
	want := map[string]string{id1: jobStaged, id2: jobStaged, id3: jobSaved}
	for _, j := range jobs {
		if j.Status != want[j.ID] {
			t.Fatalf("%s: %s (%s), muốn %s", j.Title, j.Status, j.Error, want[j.ID])
		}
	}
	books, _ := a.lib.List()
	inLib := map[string]bool{}
	for _, b := range books {
		inLib[b.Title] = true
	}
	if inLib["Cuốn một"] || inLib["Cuốn hai"] || !inLib["Cuốn ba"] {
		t.Fatalf("chỉ cuốn đã cam kết mới vào Thư viện: %v", inLib)
	}
	callTool(t, cs, "get_render_status", map[string]any{}, &r)
	if !strings.Contains(r.Message, "2 đã tạo xong, chờ người dùng cam kết") || !strings.Contains(r.Message, "1 đã lưu") {
		t.Fatalf("AI phải thấy trạng thái: %s", r.Message)
	}

	// Cam kết cuốn một: vào Thư viện kèm thời điểm cam kết; bỏ cuốn hai: khu chờ sạch.
	if _, err := a.MCPCommit([]string{id1}); err != nil {
		t.Fatal(err)
	}
	var slug1 string
	for _, j := range a.MCPJobs() {
		if j.ID == id1 {
			slug1 = j.Slug
		}
	}
	dir, _ := a.lib.Dir(slug1)
	meta, _ := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if slug1 == "" || !strings.Contains(string(meta), `"rights_confirmed_at"`) || strings.Contains(string(meta), `"rights_confirmed_at": ""`) {
		t.Fatalf("sách đã lưu phải ghi thời điểm cam kết: %s", meta)
	}
	if _, err := a.MCPDiscard(id2); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(a.mcpStageRoot()); len(entries) != 0 {
		t.Fatalf("khu chờ phải trống sau khi lưu / bỏ: %v", entries)
	}
	if left := a.MCPDismiss(); len(left) != 0 {
		t.Fatalf("ẩn việc đã xong: %+v", left)
	}
}

func TestMCP_KhuChoNapLaiVaHetHan(t *testing.T) {
	a := &App{lib: library.New(t.TempDir())}
	if _, err := a.jobStagePath("../../thoat"); err == nil {
		t.Fatal("mã sách lạ phải bị từ chối")
	}
	mk := func(id string, staged time.Time) string {
		dir, _ := a.jobStagePath(id)
		_ = os.MkdirAll(filepath.Join(dir, "sach"), 0o700)
		b, _ := json.Marshal(MCPJob{ID: id, Title: "T " + id, Status: jobStaged, StagedAt: staged.Unix(), AutoSave: true, PledgedAt: "x"})
		_ = os.WriteFile(filepath.Join(dir, mcpJobFile), b, 0o600)
		return dir
	}
	fresh := mk("0123456789abcdef", time.Now().Add(-time.Hour))
	old := mk("fedcba9876543210", time.Now().Add(-8*24*time.Hour))
	a.loadStagedJobs()
	jobs := a.MCPJobs()
	if len(jobs) != 1 || jobs[0].ID != "0123456789abcdef" || jobs[0].AutoSave || jobs[0].PledgedAt != "" {
		t.Fatalf("nạp lại khu chờ (bỏ cam kết cũ, bỏ bản quá 7 ngày): %+v", jobs)
	}
	if _, err := os.Stat(old); err == nil {
		t.Fatal("bản quá 7 ngày phải bị xoá")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatal("bản còn hạn phải giữ")
	}
}
