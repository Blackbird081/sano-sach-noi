// sano-mcp — cầu nối MCP cho Claude Desktop, Claude Code, Codex… trên cùng máy
// với app Sano. Phần mềm AI chạy chương trình này (stdio); mỗi dòng JSON-RPC nhận
// được chuyển nguyên sang máy chủ MCP trong app Sano (127.0.0.1, kèm mã bí mật
// đọc từ ~/Sano/.mcp), câu trả lời chuyển ngược lại. App chưa mở thì trả lỗi rõ
// để AI báo người dùng mở Sano.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"sano/desktop/internal/library"
	"sano/desktop/internal/mcplink"
)

// maxLine — một dòng JSON-RPC tối đa (lời cả cuốn gửi sang để tạo sách vẫn lọt).
const maxLine = 32 << 20

func main() {
	lib, err := library.Default()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sano-mcp:", err)
		os.Exit(1)
	}
	b := &bridge{root: lib.Root(), client: &http.Client{Timeout: 10 * time.Minute}, out: os.Stdout}
	if err := b.run(context.Background(), os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "sano-mcp:", err)
		os.Exit(1)
	}
}

type bridge struct {
	root   string
	client *http.Client
	out    io.Writer
	mu     sync.Mutex // ghi stdout từng dòng một
}

func (b *bridge) run(ctx context.Context, in io.Reader) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64<<10), maxLine)
	var wg sync.WaitGroup
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		msg := append([]byte(nil), line...)
		// Mỗi yêu cầu một luồng: gọi dài (đọc bộ đọc) không chặn yêu cầu sau.
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.forward(ctx, msg)
		}()
	}
	wg.Wait()
	return sc.Err()
}

// forward gửi một thông điệp sang app, ghi câu trả lời (nếu có) ra stdout.
func (b *bridge) forward(ctx context.Context, msg []byte) {
	id, isRequest := requestID(msg)
	resp, err := b.post(ctx, msg)
	if err != nil {
		if isRequest {
			b.writeError(id, err)
		}
		return
	}
	if len(bytes.TrimSpace(resp)) > 0 {
		b.writeLine(resp)
	}
}

func (b *bridge) post(ctx context.Context, msg []byte) ([]byte, error) {
	local, err := mcplink.ReadLocal(b.root)
	if err != nil {
		return nil, errAppClosed
	}
	token, err := mcplink.ReadToken(b.root)
	if err != nil {
		return nil, errAppClosed
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, local.URL, bytes.NewReader(msg))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := b.client.Do(req)
	if err != nil {
		return nil, errAppClosed
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(res.Body, maxLine))
	if err != nil {
		return nil, err
	}
	switch {
	case res.StatusCode == http.StatusAccepted || res.StatusCode == http.StatusNoContent:
		return nil, nil // thông báo (notification): không có câu trả lời
	case res.StatusCode == http.StatusUnauthorized:
		return nil, errors.New("sai mã kết nối Sano: mở lại app Sano rồi thử lại")
	case res.StatusCode >= 300:
		return nil, fmt.Errorf("app Sano trả lỗi %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

var errAppClosed = errors.New("app Sano chưa mở trên máy này. Mở Sano rồi thử lại")

// requestID lấy id của yêu cầu (thông báo không có id thì không cần trả lời).
func requestID(msg []byte) (json.RawMessage, bool) {
	var m struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
	}
	if json.Unmarshal(msg, &m) != nil || m.Method == "" || len(m.ID) == 0 || string(m.ID) == "null" {
		return nil, false
	}
	return m.ID, true
}

func (b *bridge) writeError(id json.RawMessage, err error) {
	line, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   map[string]any{"code": -32000, "message": err.Error()},
	})
	b.writeLine(line)
}

func (b *bridge) writeLine(line []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, _ = b.out.Write(append(bytes.TrimSpace(line), '\n'))
}
