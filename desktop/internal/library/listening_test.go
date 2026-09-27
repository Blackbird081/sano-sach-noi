package library

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRecordListening(t *testing.T) {
	l := New(t.TempDir())
	at := time.Date(2026, 9, 27, 21, 15, 0, 0, time.Local)
	if err := l.RecordListening("sach-a", 30, 45, at); err != nil {
		t.Fatal(err)
	}
	if err := l.RecordListening("sach-a", 30, 45, at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := l.RecordListening("sach-b", 9999, 9999, at.Add(3*time.Hour)); err != nil { // sang ngày sau, bị cắt
		t.Fatal(err)
	}
	if err := l.RecordListening("../x", 10, 10, at); err == nil {
		t.Error("slug lạ phải báo lỗi")
	}
	if err := l.RecordListening("sach-a", math.NaN(), -5, at); err != nil {
		t.Fatal(err)
	}
	d, err := l.ListenLog()
	if err != nil {
		t.Fatal(err)
	}
	day := d.Days["2026-09-27"]
	if day == nil || day.Books["sach-a"].Listen != 60 || day.Books["sach-a"].Audio != 90 || day.Hours[21] != 60 {
		t.Fatalf("ngày 27 sai: %+v", day)
	}
	next := d.Days["2026-09-28"]
	if next == nil || next.Books["sach-b"].Listen != maxListenChunk || next.Hours[0] != maxListenChunk {
		t.Fatalf("ngày 28 sai: %+v", next)
	}
}

func TestMarkFinished_ChiLanDau(t *testing.T) {
	l := New(t.TempDir())
	d1 := time.Date(2026, 9, 1, 8, 0, 0, 0, time.Local)
	if err := l.MarkFinished("sach-a", d1); err != nil {
		t.Fatal(err)
	}
	if err := l.MarkFinished("sach-a", d1.AddDate(0, 0, 5)); err != nil {
		t.Fatal(err)
	}
	d, _ := l.ListenLog()
	if d.Finished["sach-a"] != "2026-09-01" {
		t.Errorf("Finished = %q", d.Finished["sach-a"])
	}
}

func TestListenLog_FileHong(t *testing.T) {
	dir := t.TempDir()
	l := New(dir)
	if err := os.WriteFile(filepath.Join(dir, listenFile), []byte("{hỏng"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := l.ListenLog()
	if err != nil || len(d.Days) != 0 {
		t.Fatalf("d=%+v err=%v", d, err)
	}
	if _, err := os.Stat(filepath.Join(dir, listenFile+".hong")); err != nil {
		t.Error("phải giữ bản hỏng")
	}
}

func TestDetachListenLog(t *testing.T) {
	dir := t.TempDir()
	l := New(dir)
	at := time.Date(2026, 9, 27, 10, 0, 0, 0, time.Local)
	if p, err := l.DetachListenLog(at); p != "" || err != nil {
		t.Fatalf("chưa có nhật ký: p=%q err=%v", p, err)
	}
	if err := l.RecordListening("sach-a", 30, 30, at); err != nil {
		t.Fatal(err)
	}
	if n := l.ListenDays(); n != 1 {
		t.Errorf("ListenDays = %d", n)
	}
	p, err := l.DetachListenLog(at)
	if err != nil || filepath.Base(p) != "Sano so lieu nghe 2026-09-27 10.00.00.json" {
		t.Fatalf("p=%q err=%v", p, err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("file tách ra phải còn")
	}
	if n := l.ListenDays(); n != 0 {
		t.Errorf("sau khi tách ListenDays = %d, muốn 0", n)
	}
}
