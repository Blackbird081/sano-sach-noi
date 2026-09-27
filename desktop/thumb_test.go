package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestThumbnail_ResizesAndCaches(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "cover.png")
	f, _ := os.Create(src)
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, 1200, 1600))); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	cache := filepath.Join(dir, "bia")
	out, err := thumbnail(cache, src, thumbWidth)
	if err != nil {
		t.Fatal(err)
	}
	g, _ := os.Open(out)
	cfg, format, err := image.DecodeConfig(g)
	_ = g.Close()
	if err != nil || format != "jpeg" || cfg.Width != 480 || cfg.Height != 640 {
		t.Fatalf("thu nhỏ sai: %v %s %+v", err, format, cfg)
	}
	again, _ := thumbnail(cache, src, thumbWidth)
	if again != out {
		t.Error("lần hai phải dùng bản đệm")
	}
	small := filepath.Join(dir, "nho.png")
	f, _ = os.Create(small)
	_ = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 300, 400)))
	_ = f.Close()
	if _, err := thumbnail(cache, small, thumbWidth); err == nil {
		t.Error("ảnh nhỏ hơn bản thu nhỏ: trả ảnh gốc")
	}
}
