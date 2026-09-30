package mcplink

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestToken_GiuNguyenVaChiChuDoc(t *testing.T) {
	root := t.TempDir()
	a, err := Token(root)
	if err != nil || len(a) != 64 {
		t.Fatalf("%q %v", a, err)
	}
	b, _ := Token(root)
	if a != b {
		t.Fatal("mã phải giữ nguyên giữa các lần mở app")
	}
	if runtime.GOOS != "windows" {
		st, _ := os.Stat(filepath.Join(root, Dir, tokenFile))
		if st.Mode().Perm() != 0o600 {
			t.Fatalf("quyền file mã: %v", st.Mode().Perm())
		}
	}
}

func TestLocal_ChiNhan127VaXoaDungApp(t *testing.T) {
	root := t.TempDir()
	if err := WriteLocal(root, Local{URL: "http://evil.example/mcp", PID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadLocal(root); err == nil {
		t.Fatal("địa chỉ ngoài 127.0.0.1 phải bị từ chối")
	}
	_ = WriteLocal(root, Local{URL: "http://127.0.0.1:39390/mcp", PID: 7})
	RemoveLocal(root, 8)
	if _, err := ReadLocal(root); err != nil {
		t.Fatal("app khác không được xoá file hẹn")
	}
	RemoveLocal(root, 7)
	if _, err := ReadLocal(root); err == nil {
		t.Fatal("app ghi file phải xoá được")
	}
}

func TestLocalOff(t *testing.T) {
	root := t.TempDir()
	if LocalOff(root) {
		t.Fatal("mặc định là bật")
	}
	_ = os.MkdirAll(filepath.Join(root, Dir), 0o700)
	_ = os.WriteFile(filepath.Join(root, Dir, "settings.json"), []byte(`{"localOff":true}`), 0o600)
	if !LocalOff(root) {
		t.Fatal("đọc được cài đặt tắt")
	}
}
