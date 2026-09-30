package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSnapshot_SuaRoiHoanTac(t *testing.T) {
	lib, dir := makeReworkBook(t)
	s, err := lib.TakeSnapshot("sach-thu", "AI sửa lời 1 mục")
	if err != nil {
		t.Fatal(err)
	}
	if s.Title == "" || s.Label != "AI sửa lời 1 mục" {
		t.Fatalf("bản cũ: %+v", s)
	}
	snap := filepath.Join(lib.snapshotRoot("sach-thu"), s.ID)
	// MP3 là liên kết cứng (không tốn dung lượng), metadata chép thật.
	a, _ := os.Stat(filepath.Join(dir, "ch02-sec01.mp3"))
	b, _ := os.Stat(filepath.Join(snap, "ch02-sec01.mp3"))
	if !os.SameFile(a, b) {
		t.Log("hệ thống không hỗ trợ liên kết cứng, đã chép")
	}

	// Sửa như lượt đọc lại: file mp3 mới đổi tên đè lên, chữ mới trong metadata.
	newMP3 := filepath.Join(t.TempDir(), "moi.mp3")
	_ = os.WriteFile(newMP3, []byte("moi"), 0o644)
	if err := lib.ApplySection("sach-thu", 1, "Mục mới", "Chữ mới.", "Chữ mới.", newMP3); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(snap, "ch02-sec01.mp3")); string(got) != "cu-ch02-sec01.mp3" {
		t.Fatalf("bản cũ bị ghi đè theo: %q", got)
	}

	if err := lib.RestoreSnapshot("sach-thu", s.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "ch02-sec01.mp3")); string(got) != "cu-ch02-sec01.mp3" {
		t.Fatalf("hoàn tác phải trả file cũ: %q", got)
	}
	e, _ := lib.Edit("sach-thu")
	if e.Sections[1].Text != "Chữ cũ." {
		t.Fatalf("hoàn tác phải trả chữ cũ: %q", e.Sections[1].Text)
	}
	if _, err := os.Stat(filepath.Join(dir, snapshotMeta)); err == nil {
		t.Fatal("không để lại file ghi chú bản cũ trong thư mục sách")
	}
	if err := lib.RestoreSnapshot("sach-thu", s.ID); err != ErrSnapshotGone {
		t.Fatalf("bản cũ đã dùng thì hết: %v", err)
	}
	books, _ := lib.List()
	if len(books) != 1 {
		t.Fatalf("không để lại thư mục tạm trong thư viện: %d cuốn", len(books))
	}
	if entries, _ := os.ReadDir(lib.BooksRoot()); len(entries) != 1 {
		t.Fatalf("thư viện còn rác: %v", entries)
	}
}

func TestSnapshot_Giu3BanVaChanMaLa(t *testing.T) {
	lib, _ := makeReworkBook(t)
	var ids []string
	for i := 0; i < 5; i++ {
		s, err := lib.TakeSnapshot("sach-thu", "lần")
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, s.ID)
		time.Sleep(time.Millisecond)
	}
	got := lib.Snapshots("sach-thu")
	if len(got) != snapshotKeep || got[0].ID != ids[4] {
		t.Fatalf("giữ 3 bản mới nhất: %+v", got)
	}
	for _, bad := range []string{"../../x", "20260101-000000-000000000/../..", ""} {
		if err := lib.RestoreSnapshot("sach-thu", bad); err != ErrSnapshotGone {
			t.Errorf("mã lạ %q phải bị từ chối: %v", bad, err)
		}
	}
	if _, err := lib.TakeSnapshot("../thoat", "x"); err == nil || strings.Contains(err.Error(), "giữ") {
		t.Fatal("slug lạ phải bị từ chối")
	}
}
