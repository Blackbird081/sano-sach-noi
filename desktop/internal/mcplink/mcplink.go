// Package mcplink — chỗ hẹn giữa app Sano (máy chủ MCP trong máy) và cầu nối
// sano-mcp mà Claude Desktop / Claude Code / Codex chạy: địa chỉ và mã bí mật
// nằm trong <thư viện>/.mcp, chỉ người dùng của máy đọc được (0600).
package mcplink

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Dir — thư mục chứa file hẹn trong thư mục gốc thư viện (~/Sano/.mcp).
const Dir = ".mcp"

const (
	tokenFile = "token"
	localFile = "local.json"
)

// Local — app đang mở máy chủ MCP ở đâu (app ghi lúc mở, xoá lúc tắt).
type Local struct {
	URL string `json:"url"` // http://127.0.0.1:<cổng>/mcp
	PID int    `json:"pid"` // tiến trình app đã ghi file
}

// Token trả mã bí mật của máy (tạo lần đầu, giữ nguyên các lần sau để lệnh đã
// thêm vào Claude Code / Codex không phải sửa lại).
func Token(root string) (string, error) {
	p := filepath.Join(root, Dir, tokenFile)
	if b, err := os.ReadFile(p); err == nil {
		if t := strings.TrimSpace(string(b)); len(t) >= 32 {
			return t, nil
		}
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	t := hex.EncodeToString(buf)
	if err := writePrivate(root, tokenFile, []byte(t+"\n")); err != nil {
		return "", err
	}
	return t, nil
}

// ReadToken đọc mã bí mật đã có (cầu nối dùng; không tạo mới).
func ReadToken(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, Dir, tokenFile))
	if err != nil {
		return "", err
	}
	t := strings.TrimSpace(string(b))
	if t == "" {
		return "", errors.New("mã kết nối trống")
	}
	return t, nil
}

// WriteLocal ghi địa chỉ máy chủ MCP của app đang chạy.
func WriteLocal(root string, l Local) error {
	b, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return writePrivate(root, localFile, b)
}

// ReadLocal đọc địa chỉ máy chủ MCP; chưa có file = app chưa mở (hoặc đã tắt).
func ReadLocal(root string) (Local, error) {
	var l Local
	b, err := os.ReadFile(filepath.Join(root, Dir, localFile))
	if err != nil {
		return l, err
	}
	if err := json.Unmarshal(b, &l); err != nil {
		return l, fmt.Errorf("đọc %s: %w", localFile, err)
	}
	if !strings.HasPrefix(l.URL, "http://127.0.0.1:") {
		return l, fmt.Errorf("địa chỉ lạ trong %s", localFile)
	}
	return l, nil
}

// RemoveLocal xoá file địa chỉ nếu đúng là của tiến trình pid ghi (mở hai app
// thì app tắt sau không xoá nhầm file của app còn chạy).
func RemoveLocal(root string, pid int) {
	if l, err := ReadLocal(root); err == nil && l.PID == pid {
		_ = os.Remove(filepath.Join(root, Dir, localFile))
	}
}

func writePrivate(root, name string, data []byte) error {
	dir := filepath.Join(root, Dir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, name+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(dir, name))
}
