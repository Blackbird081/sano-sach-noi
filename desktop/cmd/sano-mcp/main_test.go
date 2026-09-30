package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sano/desktop/internal/mcplink"
)

func TestBridge_ChuyenYeuCauVaBaoAppChuaMo(t *testing.T) {
	root := t.TempDir()
	token, err := mcplink.Token(root)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), `"notifications/`) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		var m struct{ ID json.RawMessage }
		_ = json.Unmarshal(body, &m)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(m.ID) + `,"result":{}}`))
	}))
	defer srv.Close()

	in := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n" + `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n"

	// App chưa mở (chưa có file hẹn): trả lỗi cho yêu cầu, bỏ qua thông báo.
	var out bytes.Buffer
	b := &bridge{root: root, client: srv.Client(), out: &out}
	if err := b.run(context.Background(), strings.NewReader(in)); err != nil {
		t.Fatal(err)
	}
	if lines := strings.Count(out.String(), "\n"); lines != 1 || !strings.Contains(out.String(), "chưa mở") {
		t.Fatalf("app chưa mở: %q", out.String())
	}

	if err := mcplink.WriteLocal(root, mcplink.Local{URL: srv.URL + "/mcp", PID: 1}); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := b.run(context.Background(), strings.NewReader(in)); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != `{"jsonrpc":"2.0","id":1,"result":{}}` {
		t.Fatalf("chuyển yêu cầu: %q", got)
	}
}

func TestBridge_GuiTenPhanMemAI(t *testing.T) {
	root := t.TempDir()
	token, _ := mcplink.Token(root)
	got := make(chan string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.Header.Get("X-Sano-Client")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()
	_ = mcplink.WriteLocal(root, mcplink.Local{URL: srv.URL + "/mcp", PID: 1})
	b := &bridge{root: root, client: srv.Client(), out: io.Discard}
	_ = token
	in := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"claude-code","version":"2"}}}` + "\n"
	if err := b.run(context.Background(), strings.NewReader(in)); err != nil {
		t.Fatal(err)
	}
	if h := <-got; h != "claude-code" {
		t.Fatalf("initialize phải kèm tên phần mềm: %q", h)
	}
}
