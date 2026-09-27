// Command sano-skill ghi skill làm sách nói (dựng từ docs/prompts) ra repo:
//
//	skills/sano-sach-noi/SKILL.md      skill cho Claude / ChatGPT Skills / Claude Code
//	skills/sano-huong-dan-ai.txt       file hướng dẫn cho Dự án ChatGPT
//	docs/public/skill/*.zip, *.txt     bản tải về trên trang hướng dẫn
//
// Chạy lại sau mỗi lần sửa prompt: go run ./cmd/sano-skill (từ gốc repo).
// Test ở docs/prompts báo lỗi nếu file trong repo lệch với prompt.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"sano/docs/prompts"
)

func main() {
	files, err := prompts.RepoFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, "dựng skill:", err)
		os.Exit(1)
	}
	for path, data := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("đã ghi", path)
	}
}
