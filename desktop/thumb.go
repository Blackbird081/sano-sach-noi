package main

// Ảnh bìa thu nhỏ cho Thư viện: bìa gốc 1200×1600 (giải nén ~7,7 MB mỗi cuốn) mà
// kệ sách chỉ hiện cỡ 70–180px. Nhiều cuốn cùng lúc làm WebKit hết bộ nhớ ảnh,
// thả bớt ảnh đã giải mã → bìa trắng. Bản thu nhỏ lưu đệm ở ~/Sano/.tam/bia.

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // giải mã bìa png
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // giải mã bìa webp
)

// thumbWidth — bề rộng bìa thu nhỏ (đủ nét cho bìa lưới ~180px trên màn Retina).
const thumbWidth = 480

// maxThumbSource — ảnh gốc lớn hơn thì không thu nhỏ (trả ảnh gốc).
const maxThumbSource = 10 << 20

// thumbnail trả đường dẫn bản thu nhỏ bề rộng w của ảnh full (tạo nếu chưa có).
// Tên file đệm gồm đường dẫn + giờ sửa + dung lượng: bìa đổi là ra bản mới.
func thumbnail(cacheDir, full string, w int) (string, error) {
	info, err := os.Stat(full)
	if err != nil {
		return "", err
	}
	if info.Size() > maxThumbSource {
		return "", fmt.Errorf("ảnh bìa quá lớn")
	}
	sum := sha1.Sum([]byte(full + "|" + strconv.FormatInt(info.ModTime().UnixNano(), 10) + "|" + strconv.FormatInt(info.Size(), 10) + "|" + strconv.Itoa(w)))
	out := filepath.Join(cacheDir, hex.EncodeToString(sum[:10])+".jpg")
	if _, err := os.Stat(out); err == nil {
		return out, nil
	}
	f, err := os.Open(full)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return "", err
	}
	if cfg.Width <= w || cfg.Width*cfg.Height > 40_000_000 {
		return "", fmt.Errorf("không cần / không thu nhỏ được")
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	src, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	h := src.Bounds().Dy() * w / src.Bounds().Dx()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(cacheDir, ".bia-*.jpg")
	if err != nil {
		return "", err
	}
	if err := jpeg.Encode(tmp, dst, &jpeg.Options{Quality: 86}); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	_ = os.Chmod(tmp.Name(), 0o644)
	if err := os.Rename(tmp.Name(), out); err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	return out, nil
}
