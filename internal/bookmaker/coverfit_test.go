package bookmaker

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestFitCover(t *testing.T) {
	cases := []struct{ w, h, wantW, wantH int }{
		{4000, 3000, 1200, 900},  // ảnh ngang
		{3000, 4000, 1200, 1600}, // ảnh dọc cùng tỉ lệ bìa
		{1000, 4000, 400, 1600},  // ảnh dọc hẹp
	}
	for _, c := range cases {
		out, ext, err := FitCover(pngBytes(t, c.w, c.h), ".png")
		if err != nil || ext != ".jpg" {
			t.Fatalf("%dx%d: ext=%q err=%v", c.w, c.h, ext, err)
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
		if err != nil || format != "jpeg" || cfg.Width != c.wantW || cfg.Height != c.wantH {
			t.Errorf("%dx%d → %s %dx%d, muốn %dx%d", c.w, c.h, format, cfg.Width, cfg.Height, c.wantW, c.wantH)
		}
	}
	small := pngBytes(t, 800, 1000)
	out, ext, err := FitCover(small, ".PNG")
	if err != nil || ext != ".png" || !bytes.Equal(out, small) {
		t.Error("ảnh vừa khung phải giữ nguyên từng byte")
	}
	if out, ext, _ := FitCover([]byte("khong phai anh"), ".webp"); ext != ".webp" || string(out) != "khong phai anh" {
		t.Error("không giải mã được: giữ nguyên")
	}
}
