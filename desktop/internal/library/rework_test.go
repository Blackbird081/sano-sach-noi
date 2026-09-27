package library

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sano/internal/bookmaker"
)

const reworkMeta = `{"title":"Sách thử","author":"Tác giả A","cover":"cover.png","image_x":"giữ",
"chapters":[
 {"title":"Sách thử","sections":[{"title":"Sách thử","file":"ch01-sec01.mp3","original_text":"Bạn đang nghe sách nói.\n\nCuốn sách: Sách thử.\n\nTác giả: Tác giả A.","reading_script":"x"}]},
 {"title":"Chương 1","sections":[
   {"title":"Mở đầu","file":"ch02-sec01.mp3","original_text":"Chữ cũ.","reading_script":"Mở đầu.\n\nChữ cũ.","image_description":"ảnh"},
   {"title":"Phần hai","file":"ch02-sec02.mp3","original_text":"Hai.","reading_script":"Phần hai.\n\nHai."}]}]}`

func makeReworkBook(t *testing.T) (*Library, string) {
	t.Helper()
	lib := New(t.TempDir())
	dir := filepath.Join(lib.BooksRoot(), "sach-thu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, s string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("metadata.json", reworkMeta)
	for _, f := range []string{"ch01-sec01.mp3", "ch02-sec01.mp3", "ch02-sec02.mp3"} {
		write(f, "cu-"+f)
	}
	if err := bookmaker.WriteAutoCover(filepath.Join(dir, "cover.png"), "Sách thử", "Tác giả A"); err != nil {
		t.Fatal(err)
	}
	if err := lib.Repack("sach-thu", "Trúc Ly"); err != nil {
		t.Fatal(err)
	}
	return lib, dir
}

func TestEdit_SectionsIntroAndCover(t *testing.T) {
	lib, _ := makeReworkBook(t)
	e, err := lib.Edit("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Sections) != 3 || !e.Sections[0].Intro || e.Sections[1].Intro {
		t.Fatalf("sections = %+v", e.Sections)
	}
	if s := e.Sections[1]; s.Text != "Chữ cũ." || s.Stem != "ch02-sec01" || s.Chapter != "Chương 1" || s.Index != 1 {
		t.Errorf("mục 2 = %+v", s)
	}
	if !e.CoverAuto {
		t.Error("bìa vẽ theo tên phải là bìa tự vẽ")
	}
	if e.Voice != "Trúc Ly" {
		t.Errorf("voice = %q", e.Voice)
	}
}

func TestApplySection_ReplacesAudioAndText(t *testing.T) {
	lib, dir := makeReworkBook(t)
	work, err := lib.SectionsWorkDir("sach-thu")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "ch02-sec01.mp3"), []byte("moi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := lib.ApplySection("sach-thu", 1, "Mở đầu mới", "Chữ mới.", "Mở đầu mới.\n\nChữ mới.", filepath.Join(work, "ch02-sec01.mp3")); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "ch02-sec01.mp3")); string(b) != "moi" {
		t.Errorf("mp3 = %q", b)
	}
	if err := lib.Repack("sach-thu", ""); err != nil {
		t.Fatal(err)
	}
	e, _ := lib.Edit("sach-thu")
	if s := e.Sections[1]; s.Title != "Mở đầu mới" || s.Text != "Chữ mới." || s.Script != "Mở đầu mới.\n\nChữ mới." {
		t.Errorf("sau sửa = %+v", s)
	}
	if e.Voice != "Trúc Ly" {
		t.Errorf("đóng gói lại phải giữ giọng: %q", e.Voice)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if !strings.Contains(string(raw), `"image_description": "ảnh"`) || !strings.Contains(string(raw), `"image_x": "giữ"`) {
		t.Errorf("mất trường khác trong metadata:\n%s", raw)
	}
	texts, _ := lib.Texts("sach-thu")
	if texts[1].Text != "Chữ mới." {
		t.Errorf("gói zip chưa có chữ mới: %+v", texts[1])
	}
}

func TestRetitleIntro(t *testing.T) {
	in := "Bạn đang nghe sách nói.\n\nCuốn sách: Cũ.\n\nTác giả: A."
	if got := RetitleIntro(in, "Mới", "B"); got != "Bạn đang nghe sách nói.\n\nCuốn sách: Mới.\n\nTác giả: B." {
		t.Errorf("đổi cả hai: %q", got)
	}
	if got := RetitleIntro(in, "Mới", ""); got != "Bạn đang nghe sách nói.\n\nCuốn sách: Mới." {
		t.Errorf("bỏ tác giả: %q", got)
	}
	if got := RetitleIntro("Bạn đang nghe sách nói.\n\nCuốn sách: Cũ.", "Mới.", "C"); got != "Bạn đang nghe sách nói.\n\nCuốn sách: Mới.\n\nTác giả: C." {
		t.Errorf("thêm tác giả: %q", got)
	}
}

func TestRedrawAndSetCover(t *testing.T) {
	lib, dir := makeReworkBook(t)
	if _, err := lib.UpdateInfo("sach-thu", Info{Title: "Tên mới", Author: "Tác giả A"}); err != nil {
		t.Fatal(err)
	}
	if err := lib.RedrawCover("sach-thu"); err != nil {
		t.Fatal(err)
	}
	if !bookmaker.IsAutoCover(filepath.Join(dir, "cover.png"), "Tên mới", "Tác giả A") {
		t.Error("bìa chưa vẽ lại theo tên mới")
	}
	img := filepath.Join(t.TempDir(), "anh.jpg")
	_ = os.WriteFile(img, []byte("jpg"), 0o644)
	if err := lib.SetCoverImage("sach-thu", img); err != nil {
		t.Fatal(err)
	}
	e, _ := lib.Edit("sach-thu")
	if e.CoverAuto || !strings.HasSuffix(e.Cover, "cover.jpg") || fileExists(filepath.Join(dir, "cover.png")) {
		t.Errorf("sau chọn ảnh: auto=%v cover=%q", e.CoverAuto, e.Cover)
	}
}

func TestVoiceJob_ResumeFinishCancel(t *testing.T) {
	lib, dir := makeReworkBook(t)
	j, err := lib.BeginVoice("sach-thu", "Mỹ Duyên")
	if err != nil || len(j.Done) != 0 {
		t.Fatal(j, err)
	}
	work, _ := lib.VoiceWorkDir("sach-thu")
	for _, stem := range []string{"ch01-sec01", "ch02-sec01"} {
		_ = os.WriteFile(filepath.Join(work, stem+".mp3"), []byte("giong-moi"), 0o644)
		if err := lib.MarkVoiceDone("sach-thu", stem); err != nil {
			t.Fatal(err)
		}
	}
	if err := lib.FinishVoice("sach-thu"); err == nil {
		t.Fatal("thiếu một mục: không được thay")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "ch01-sec01.mp3")); string(b) != "cu-ch01-sec01.mp3" {
		t.Fatal("thiếu mục mà vẫn thay file")
	}
	// Mở lại cùng giọng: nối tiếp, giữ mục đã xong.
	if j, _ := lib.BeginVoice("sach-thu", "Mỹ Duyên"); len(j.Done) != 2 {
		t.Fatalf("nối tiếp: %+v", j)
	}
	if p := lib.PendingVoiceJobs(); p["sach-thu"] == nil {
		t.Error("phải thấy đổi giọng dở")
	}
	// Sửa một mục đã đọc giọng mới → mục đó phải đọc lại.
	sw, _ := lib.SectionsWorkDir("sach-thu")
	_ = os.WriteFile(filepath.Join(sw, "x.mp3"), []byte("sua"), 0o644)
	if err := lib.ApplySection("sach-thu", 1, "", "Chữ mới.", "Chữ mới.", filepath.Join(sw, "x.mp3")); err != nil {
		t.Fatal(err)
	}
	if j, _ := lib.BeginVoice("sach-thu", "Mỹ Duyên"); len(j.Done) != 1 {
		t.Fatalf("mục sửa phải bị bỏ khỏi đã xong: %+v", j)
	}
	for _, stem := range []string{"ch02-sec01", "ch02-sec02"} {
		_ = os.WriteFile(filepath.Join(work, stem+".mp3"), []byte("giong-moi"), 0o644)
		_ = lib.MarkVoiceDone("sach-thu", stem)
	}
	if err := lib.FinishVoice("sach-thu"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "ch02-sec02.mp3")); string(b) != "giong-moi" {
		t.Error("chưa thay giọng")
	}
	e, _ := lib.Edit("sach-thu")
	if e.Voice != "Mỹ Duyên" || e.VoiceJob != nil {
		t.Errorf("voice=%q job=%v", e.Voice, e.VoiceJob)
	}
	// Đổi giọng khác rồi huỷ: giữ nguyên.
	_, _ = lib.BeginVoice("sach-thu", "Hải Đăng")
	if err := lib.CancelVoice("sach-thu"); err != nil {
		t.Fatal(err)
	}
	if e, _ := lib.Edit("sach-thu"); e.VoiceJob != nil || e.Voice != "Mỹ Duyên" {
		t.Error("huỷ phải giữ giọng cũ")
	}
}

func TestSetCoverImage_ShrinksBigPhoto(t *testing.T) {
	lib, dir := makeReworkBook(t)
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 3000, 4000)))
	src := filepath.Join(t.TempDir(), "anh.png")
	_ = os.WriteFile(src, b.Bytes(), 0o644)
	if err := lib.SetCoverImage("sach-thu", src); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(dir, "cover.jpg"))
	if err != nil {
		t.Fatal("ảnh lớn phải lưu thành cover.jpg đã thu nhỏ")
	}
	cfg, _, _ := image.DecodeConfig(f)
	_ = f.Close()
	if cfg.Width != 1200 || cfg.Height != 1600 || fileExists(filepath.Join(dir, "cover.png")) {
		t.Errorf("bìa %dx%d", cfg.Width, cfg.Height)
	}
}
