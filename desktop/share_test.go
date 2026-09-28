package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"sano/desktop/internal/library"
)

func pngB64(t *testing.T, w, h int, c color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestShareFileName(t *testing.T) {
	cases := map[string]string{
		"Chánh niệm, nghệ thuật của sự có mặt": "Chánh niệm, nghệ thuật của sự có mặt - Sano.png",
		`a/b\c:d*e?"f"<g>|h`: "a b c d e f g h - Sano.png",
		"   ":                "Sách nói - Sano.png",
	}
	for in, want := range cases {
		if got := shareFileName(in, "png"); got != want {
			t.Errorf("shareFileName(%q) = %q, muốn %q", in, got, want)
		}
	}
	long := strings.Repeat("á", 200)
	if got := shareFileName(long, "mp4"); len([]rune(got)) != 80+len(" - Sano.mp4") {
		t.Errorf("tên dài không bị cắt còn 80 ký tự: %d", len([]rune(got)))
	}
}

func TestDecodePNG(t *testing.T) {
	if _, err := decodePNG(pngB64(t, 2, 2, color.White)); err != nil {
		t.Fatalf("PNG hợp lệ bị từ chối: %v", err)
	}
	if _, err := decodePNG("data:image/png;base64," + pngB64(t, 2, 2, color.White)); err != nil {
		t.Fatalf("PNG kèm tiền tố data: bị từ chối: %v", err)
	}
	if _, err := decodePNG(base64.StdEncoding.EncodeToString([]byte("GIF89a......"))); err == nil {
		t.Fatal("không phải PNG mà vẫn nhận")
	}
	if _, err := decodePNG("%%%"); err == nil {
		t.Fatal("base64 hỏng mà vẫn nhận")
	}
}

func shareTestApp(t *testing.T) (*App, string) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "Sach", "sach-thu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return &App{lib: library.New(root)}, dir
}

func TestMakeShareVideo_Validation(t *testing.T) {
	a, dir := shareTestApp(t)
	if err := os.WriteFile(filepath.Join(dir, "ch01-sec01.mp3"), []byte("ID3"), 0o644); err != nil {
		t.Fatal(err)
	}
	frame := []ShareFrame{{PNG: pngB64(t, 4, 4, color.Black), Sec: 1}}
	bad := []ShareVideoRequest{
		{Slug: "sach-thu", File: "ch01-sec01.mp3", Start: 0, End: 0, Frames: frame},    // quá ngắn
		{Slug: "sach-thu", File: "ch01-sec01.mp3", Start: 0, End: 120, Frames: frame},  // quá dài
		{Slug: "sach-thu", File: "ch01-sec01.mp3", Start: 0, End: 2},                   // không có ảnh
		{Slug: "sach-thu", File: "metadata.json", Start: 0, End: 2, Frames: frame},     // không phải mp3
		{Slug: "sach-thu", File: "khong-co.mp3", Start: 0, End: 2, Frames: frame},      // không có file
		{Slug: "../sach-thu", File: "ch01-sec01.mp3", Start: 0, End: 2, Frames: frame}, // slug lạ
		{Slug: "sach-thu", File: "Sach/sach-thu/..", Start: 0, End: 2, Frames: frame},  // không có tên file
	}
	for i, r := range bad {
		if _, err := a.MakeShareVideo(r); err == nil {
			t.Errorf("ca %d: phải báo lỗi", i)
		}
	}
}

func TestShareVideoArgs_Wave(t *testing.T) {
	req := ShareVideoRequest{Start: 1.5, WaveX: 10, WaveY: 20, WaveW: 300, WaveH: 60}
	args := strings.Join(shareVideoArgs("l.txt", "a.mp3", "w.png", "o.mp4", req, 8, "libx264"), " ")
	for _, want := range []string{"-ss 1.500", "-t 8.000", "-i w.png", "alphamerge", "overlay=10:20", "-c:v libx264", "+faststart"} {
		if !strings.Contains(args, want) {
			t.Errorf("thiếu %q trong: %s", want, args)
		}
	}
	noWave := strings.Join(shareVideoArgs("l.txt", "a.mp3", "", "o.mp4", req, 8, "mpeg4"), " ")
	if strings.Contains(noWave, "alphamerge") || !strings.Contains(noWave, "-q:v 3") {
		t.Errorf("không có sóng âm mà vẫn ghép mặt nạ, hoặc thiếu tham số mpeg4: %s", noWave)
	}
}

// Tạo video thật (cần ffmpeg trên máy): 2 ảnh thẻ + 3 giây tiếng + sóng âm.
func TestMakeShareVideo_FFmpeg(t *testing.T) {
	ff := findFFmpeg()
	if ff == "" {
		t.Skip("không có ffmpeg")
	}
	a, dir := shareTestApp(t)
	mp3 := filepath.Join(dir, "ch01-sec01.mp3")
	if out, err := exec.Command(ff, "-loglevel", "error", "-y", "-f", "lavfi", "-i", "sine=frequency=440:duration=5",
		"-c:a", "libmp3lame", mp3).CombinedOutput(); err != nil {
		t.Fatalf("tạo mp3 thử: %v %s", err, out)
	}
	t.Cleanup(func() {
		if share.video != "" {
			_ = os.Remove(share.video)
			share.video = ""
		}
	})
	dur, err := a.MakeShareVideo(ShareVideoRequest{
		Slug: "sach-thu", File: "Sach/sach-thu/ch01-sec01.mp3", Start: 1, End: 4, Title: "Sách thử",
		Frames:  []ShareFrame{{PNG: pngB64(t, 108, 192, color.RGBA{200, 50, 50, 255}), Sec: 1.2}, {PNG: pngB64(t, 108, 192, color.RGBA{50, 50, 200, 255}), Sec: 1.8}},
		WavePNG: pngB64(t, 60, 20, color.White), WaveX: 24, WaveY: 100, WaveW: 60, WaveH: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dur != 3 {
		t.Errorf("thời lượng = %v, muốn 3", dur)
	}
	src, name, err := currentShareVideo()
	if err != nil {
		t.Fatal(err)
	}
	if name != "Sách thử - Sano.mp4" {
		t.Errorf("tên file = %q", name)
	}
	if info, err := os.Stat(src); err != nil || info.Size() < 1000 {
		t.Fatalf("video không có hoặc quá nhỏ: %v", err)
	}
}
