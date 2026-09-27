package bookmaker

// Ảnh bìa người dùng chọn: ảnh chụp điện thoại 4000×3000 chép nguyên vào sách, gói
// zip và M4B thì nặng vô ích (giải mã ~48 MB mỗi lần hiện). Thu về tối đa bằng bìa
// tự vẽ (1200×1600), giữ tỉ lệ; ảnh nhỏ hơn giữ nguyên từng byte.

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // giải mã bìa png
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // giải mã bìa webp
)

// maxCoverPixels — ảnh lớn hơn thì từ chối giải mã (chống ảnh "bom" nén).
const maxCoverPixels = 60_000_000

// FitCover thu ảnh bìa data (đuôi ext: .jpg/.jpeg/.png/.webp) về trong khung
// coverWidth×coverHeight. Trả dữ liệu + đuôi file mới (".jpg" khi đã thu nhỏ;
// nguyên bản khi ảnh vừa khung hoặc không giải mã được).
func FitCover(data []byte, ext string) ([]byte, string, error) {
	ext = strings.ToLower(ext)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return data, ext, nil // định dạng lạ: giữ nguyên, để chỗ hiện ảnh tự xử lý
	}
	if cfg.Width <= coverWidth && cfg.Height <= coverHeight {
		return data, ext, nil
	}
	if cfg.Width*cfg.Height > maxCoverPixels {
		return nil, "", fmt.Errorf("ảnh bìa quá lớn (%d×%d điểm ảnh)", cfg.Width, cfg.Height)
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("đọc ảnh bìa: %w", err)
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	// Vừa khung 1200×1600, giữ tỉ lệ.
	if w*coverHeight > h*coverWidth {
		h, w = h*coverWidth/w, coverWidth
	} else {
		w, h = w*coverHeight/h, coverHeight
	}
	dst := image.NewRGBA(image.Rect(0, 0, max(w, 1), max(h, 1)))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src) // nền trắng cho png trong suốt
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 88}); err != nil {
		return nil, "", fmt.Errorf("ghi ảnh bìa: %w", err)
	}
	return out.Bytes(), ".jpg", nil
}
