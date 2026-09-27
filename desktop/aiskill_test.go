package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAIGuide(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvSampleOut, dir)
	a := &App{}
	for kind, name := range map[string]string{"claude": "sano-sach-noi.zip", "chatgpt": "sano-huong-dan-ai.txt"} {
		p, err := a.SaveAIGuide(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if p != filepath.Join(dir, name) {
			t.Errorf("%s lưu vào %q", kind, p)
		}
		data, _ := os.ReadFile(p)
		if len(data) == 0 {
			t.Errorf("%s: file rỗng", kind)
		}
		if kind != "claude" && !strings.Contains(string(data), "VIỆC A") {
			t.Error("file hướng dẫn thiếu nội dung")
		}
	}
}
