package main

import (
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sano/desktop/internal/library"
)

const m4bTestMeta = `{"title":"Sách: thử M4B","author":"Tác giả A","chapters":[
 {"title":"Chương 1","sections":[{"title":"Mở đầu","file":"ch01-sec01.mp3"},{"title":"Phần hai","file":"ch01-sec02.mp3"}]},
 {"title":"Chương 2","sections":[{"title":"Chương 2","file":"ch02-sec01.mp3"}]}]}`

// m4bTestApp tạo App với thư viện tạm có một cuốn "sach-thu" (MP3 im lặng thật).
func m4bTestApp(t *testing.T) *App {
	t.Helper()
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("không có ffmpeg")
	}
	lib := library.New(t.TempDir())
	dir := filepath.Join(lib.BooksRoot(), "sach-thu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(m4bTestMeta), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"ch01-sec01.mp3", "ch01-sec02.mp3", "ch02-sec01.mp3"} {
		out, err := exec.Command(ffmpeg, "-loglevel", "error", "-y", "-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
			"-t", "1", "-c:a", "libmp3lame", "-b:a", "64k", filepath.Join(dir, f)).CombinedOutput()
		if err != nil {
			t.Skipf("ffmpeg không tạo được MP3 thử: %v %s", err, out)
		}
	}
	return &App{lib: lib}
}

func waitM4B(t *testing.T, a *App) *M4BStatus {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if st := a.M4BStatus(); st != nil && !st.Running {
			return st
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("xuất M4B quá lâu")
	return nil
}

func TestExportM4BVaoThuMuc(t *testing.T) {
	a := m4bTestApp(t)
	outDir := t.TempDir()
	t.Setenv(EnvM4BOut, outDir)

	st, err := a.ExportM4B("sach-thu", 1.5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Running || st.Title != "Sách: thử M4B" {
		t.Fatalf("trạng thái đầu lạ: %+v", st)
	}
	st = waitM4B(t, a)
	if !st.Done || st.Error != "" || st.Chapters != 3 {
		t.Fatalf("xuất không xong: %+v", st)
	}
	want := filepath.Join(outDir, "Sách thử M4B.m4b")
	if st.Path != want || !fileExists(want) {
		t.Fatalf("file = %q, muốn %q", st.Path, want)
	}
	// 3 giây âm thanh + nghỉ 1,5 giây trước tiểu mục 2 + nghỉ 2 giây trước chương 2.
	if st.DurationSec < 6.4 || st.DurationSec > 6.6 {
		t.Errorf("thời lượng = %.2f, muốn ≈ 6,5", st.DurationSec)
	}
}

func TestExportM4BDuongDanFile(t *testing.T) {
	a := m4bTestApp(t)
	out := filepath.Join(t.TempDir(), "con", "ten-rieng.m4b")
	t.Setenv(EnvM4BOut, out)
	if _, err := a.ExportM4B("sach-thu", 1.5, 2); err != nil {
		t.Fatal(err)
	}
	if st := waitM4B(t, a); !st.Done || st.Path != out {
		t.Fatalf("trạng thái: %+v", st)
	}
}

func TestExportM4BTuChoiKhiDangXuat(t *testing.T) {
	a := m4bTestApp(t)
	t.Setenv(EnvM4BOut, t.TempDir())
	a.m4b = &m4bJob{cancel: func() {}, status: M4BStatus{Running: true, Slug: "khac"}}
	if _, err := a.ExportM4B("sach-thu", 1.5, 2); err == nil || !strings.Contains(err.Error(), "đang xuất") {
		t.Fatalf("err = %v, muốn từ chối vì đang xuất cuốn khác", err)
	}
}

func TestExportM4BSachKhongCo(t *testing.T) {
	a := m4bTestApp(t)
	t.Setenv(EnvM4BOut, t.TempDir())
	if _, err := a.ExportM4B("khong-co", 1.5, 2); err == nil {
		t.Fatal("sách không có phải báo lỗi")
	}
	if _, err := a.ExportM4B("../x", 1.5, 2); err == nil {
		t.Fatal("slug lạ phải báo lỗi")
	}
}

func TestRevealM4BChuaXuat(t *testing.T) {
	a := &App{lib: library.New(t.TempDir())}
	if err := a.RevealM4B(); err == nil {
		t.Fatal("chưa xuất mà mở được thư mục")
	}
}

func TestM4BGaps(t *testing.T) {
	g, err := m4bGaps(1.24, 2.5)
	if err != nil || g.Section != 1200*time.Millisecond || g.Chapter != 2500*time.Millisecond {
		t.Errorf("m4bGaps(1.24, 2.5) = %+v, %v", g, err)
	}
	if g, err := m4bGaps(0, 10); err != nil || g.Section != 0 || g.Chapter != 10*time.Second {
		t.Errorf("m4bGaps(0, 10) = %+v, %v", g, err)
	}
	for _, c := range [][2]float64{{-0.1, 2}, {1, 10.1}, {math.NaN(), 2}, {1, math.Inf(1)}} {
		if _, err := m4bGaps(c[0], c[1]); err == nil {
			t.Errorf("m4bGaps(%v, %v) phải báo lỗi", c[0], c[1])
		}
	}
	a := m4bTestApp(t)
	if _, err := a.ExportM4B("sach-thu", 11, 2); err == nil {
		t.Error("ExportM4B với quãng nghỉ 11 giây phải báo lỗi")
	}
}
