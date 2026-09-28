package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Nhạc hiệu tạm + tiếng chuông (không cần bộ đọc): tạo được, đúng độ dài, dùng lại lần sau.
func TestBookVideoExtras_MusicChime(t *testing.T) {
	if findFFmpeg() == "" {
		t.Skip("không có ffmpeg")
	}
	a, dir := shareTestApp(t)
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte(`{"title":"Sách thử","chapters":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := a.BookVideoExtras(VideoExtrasRequest{Slug: "sach-thu", Music: "sano", Chime: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Key != "music" || got[1].Key != "chime" {
		t.Fatalf("kết quả = %+v", got)
	}
	if got[0].DurSec < 4.3 || got[0].DurSec > 4.8 || got[1].DurSec < 1.1 || got[1].DurSec > 1.5 {
		t.Errorf("thời lượng nhạc hiệu %.2f, chuông %.2f", got[0].DurSec, got[1].DurSec)
	}
	if p, ok := extraPath(got[0].ID); !ok || !fileExists(p) {
		t.Errorf("không tra được file theo mã")
	}
	again, err := a.BookVideoExtras(VideoExtrasRequest{Slug: "sach-thu", Music: "sano"})
	if err != nil || len(again) != 1 || again[0].ID != got[0].ID {
		t.Errorf("lần hai phải dùng lại cùng file: %+v %v", again, err)
	}
	if _, err := a.BookVideoExtras(VideoExtrasRequest{Slug: "sach-thu", Music: "file", MusicPath: "/khong/co.mp3"}); err == nil {
		t.Error("file nhạc không có mà vẫn nhận")
	}
	if _, ok := extraPath("khong-co-ma-nay"); ok {
		t.Error("mã lạ mà vẫn ra file")
	}
}
