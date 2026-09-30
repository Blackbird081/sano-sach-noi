package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := c.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint: srv.URL + "/mcp", HTTPClient: &http.Client{Transport: bearer{token}}, DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()

	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, tl := range tools.Tools {
		names = append(names, tl.Name)
		if tl.Annotations == nil || !tl.Annotations.ReadOnlyHint {
			t.Errorf("%s phải đánh dấu chỉ đọc", tl.Name)
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
