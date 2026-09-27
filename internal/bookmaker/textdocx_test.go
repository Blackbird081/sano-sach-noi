package bookmaker

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseText(t *testing.T, text string) *Book {
	t.Helper()
	data, err := TextToDocx(text)
	if err != nil {
		t.Fatalf("TextToDocx: %v", err)
	}
	p := filepath.Join(t.TempDir(), "dan.docx")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	book, err := ParseDocx(p)
	if err != nil {
		t.Fatalf("ParseDocx: %v", err)
	}
	return book
}

func TestTextToDocx(t *testing.T) {
	text := "Dưới đây là nội dung sách của bạn:\n\n```markdown\n% Quản lý tiền\n\n# Chương 1. Ngân sách\n\n## Chuyện của Lan\n\nLan hai mươi ba tuổi.\n\n- Ghi lại **mọi** khoản chi.\n\n## Ba ý cần nhớ\nMột. Ghi chép.\n\n# Chương 2. Quỹ dự phòng\n### Bốn tiêu chí\nThanh khoản cao.\n```\n"
	book := parseText(t, text)
	if book.Title != "Quản lý tiền" {
		t.Errorf("Title = %q", book.Title)
	}
	if len(book.Chapters) != 2 {
		t.Fatalf("chapters = %d, muốn 2", len(book.Chapters))
	}
	if n := len(book.Chapters[0].Sections); n != 2 {
		t.Errorf("chương 1 có %d mục, muốn 2", n)
	}
	s := book.Chapters[0].Sections[0]
	if s.Title != "Chuyện của Lan" {
		t.Errorf("mục đầu = %q", s.Title)
	}
	body := s.Text
	if strings.Contains(body, "Dưới đây") || strings.Contains(body, "**") || strings.Contains(body, "- ") {
		t.Errorf("còn lời chào hoặc dấu markdown: %q", body)
	}
	if !strings.Contains(body, "Ghi lại mọi khoản chi.") {
		t.Errorf("mất đoạn văn: %q", body)
	}
	if got := book.Chapters[1].Sections[0].Title; got != "Bốn tiêu chí" {
		t.Errorf("### phải thành mục, được %q", got)
	}
}

func TestTextToDocxNoChapter(t *testing.T) {
	if _, err := TextToDocx("% Chỉ có tên sách\n## Một mục\nChữ."); !errors.Is(err, ErrNoChapterLine) {
		t.Errorf("err = %v, muốn ErrNoChapterLine", err)
	}
	if _, err := TextToDocx(strings.Repeat("a", MaxPastedTextBytes+1)); err == nil {
		t.Error("văn bản quá dài phải báo lỗi")
	}
}

func TestTextToDocxEscapes(t *testing.T) {
	book := parseText(t, "# Chương <1> & \"hai\"\nA < B & C")
	if got := book.Chapters[0].Title; got != `Chương <1> & "hai"` {
		t.Errorf("tiêu đề = %q", got)
	}
}
