package main

import (
	"encoding/base64"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDecodeFrame(t *testing.T) {
	if _, ext, err := decodeFrame(pngB64(t, 2, 2, color.White)); err != nil || ext != "png" {
		t.Fatalf("PNG: %v %s", err, ext)
	}
	jpg := base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0})
	if _, ext, err := decodeFrame("data:image/jpeg;base64," + jpg); err != nil || ext != "jpg" {
		t.Fatalf("JPEG: %v %s", err, ext)
	}
	if _, _, err := decodeFrame(base64.StdEncoding.EncodeToString([]byte("GIF89a"))); err == nil {
		t.Fatal("GIF mà vẫn nhận")
	}
}

func TestValidPlan(t *testing.T) {
	ok := BookVideoPlan{Frames: []float64{1, 2}, Segs: []BookVideoSeg{{Silence: true, Dur: 3}}, Width: 640, Height: 360}
	if _, err := validPlan(&ok, 2); err != nil {
		t.Fatalf("kế hoạch đúng bị từ chối: %v", err)
	}
	bad := []BookVideoPlan{
		{Frames: []float64{1}, Segs: ok.Segs, Width: 640, Height: 360},                                                          // thiếu khung hình
		{Frames: []float64{1, 1}, Segs: ok.Segs, Width: 640, Height: 360},                                                       // hình, tiếng lệch
		{Frames: []float64{1, 2}, Segs: ok.Segs, Width: 641, Height: 360},                                                       // cỡ lẻ
		{Frames: []float64{1, 2}, Segs: ok.Segs, Width: 640, Height: 360, Overlays: []BookVideoOverlay{{W: 700, H: 10, T1: 1}}}, // lớp phủ tràn
	}
	for i, p := range bad {
		if _, err := validPlan(&p, 2); err == nil {
			t.Errorf("ca %d: phải báo lỗi", i)
		}
	}
}

// Tạo video cả cuốn thật (cần ffmpeg): 2 tiểu mục + lặng + lớp tiến độ.
func TestBookVideo_FFmpeg(t *testing.T) {
	ff := findFFmpeg()
	if ff == "" {
		t.Skip("không có ffmpeg")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := os.MkdirAll(filepath.Join(home, "Downloads"), 0o755); err != nil {
		t.Fatal(err)
	}
	a, dir := shareTestApp(t)
	for _, n := range []string{"ch01-sec01.mp3", "ch01-sec02.mp3"} {
		if out, err := exec.Command(ff, "-loglevel", "error", "-y", "-f", "lavfi", "-i", "sine=frequency=330:duration=2",
			"-c:a", "libmp3lame", filepath.Join(dir, n)).CombinedOutput(); err != nil {
			t.Fatalf("tạo mp3 thử: %v %s", err, out)
		}
	}
	id, err := a.BookVideoBegin("sach-thu", "Sách thử")
	if err != nil {
		t.Fatal(err)
	}
	secs := []float64{1, 2, 2.5}
	for i := range secs {
		if err := a.BookVideoFrame(id, i, pngB64(t, 64, 36, color.RGBA{uint8(60 * i), 80, 160, 255})); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.BookVideoFrame(id, 7, pngB64(t, 64, 36, color.White)); err == nil {
		t.Fatal("khung hình sai thứ tự mà vẫn nhận")
	}
	plan := BookVideoPlan{
		Frames: secs, Width: 640, Height: 360, Title: "Sách thử", SRT: "1\n00:00:00,000 --> 00:00:01,000\nThử\n", Desc: "0:00 Thử",
		Segs:     []BookVideoSeg{{Silence: true, Dur: 1}, {File: "Sach/sach-thu/ch01-sec01.mp3", Dur: 2}, {Silence: true, Dur: 0.5}, {File: "ch01-sec02.mp3", Start: 0.5, Dur: 2}},
		Overlays: []BookVideoOverlay{{PNG: pngB64(t, 100, 8, color.White), X: 20, Y: 300, W: 100, H: 8, T0: 1, T1: 5.5}},
		Thumb:    pngB64(t, 128, 72, color.Black),
	}
	if err := a.BookVideoFinish(id, plan); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(60 * time.Second)
	var st *BookVideoStatus
	for time.Now().Before(deadline) {
		if st = a.BookVideoState(); st != nil && !st.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if st == nil || !st.Done || st.Error != "" {
		t.Fatalf("tạo video không xong: %+v", st)
	}
	out, err := exec.Command(filepath.Join(filepath.Dir(ff), "ffprobe"), "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", st.Video).Output()
	if err == nil && !strings.HasPrefix(strings.TrimSpace(string(out)), "5.") {
		t.Errorf("thời lượng video = %s, muốn ~5.5", out)
	}
	for _, n := range []string{"thumbnail.png", "phu-de.srt", "mo-ta-youtube.txt"} {
		if !fileExists(filepath.Join(st.Dir, n)) {
			t.Errorf("thiếu %s", n)
		}
	}
}
