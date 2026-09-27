package bookmaker

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSectionScript_LikeCreate(t *testing.T) {
	got := SectionScript(nil, "1.2. Quản lý KPI", "Đo KPI mỗi tuần: đều đặn.")
	want := "Quản lý ca pê i.\n\nĐo ca pê i mỗi tuần. Đều đặn."
	if got != want {
		t.Errorf("SectionScript = %q, muốn %q", got, want)
	}
	if got := IntroScript(nil, "Bạn đang nghe sách nói.\n\nCuốn sách: KPI."); got != "Bạn đang nghe sách nói.\n\nCuốn sách. Ca pê i." {
		t.Errorf("IntroScript = %q", got)
	}
}

func TestKeepsHeadingNumbers(t *testing.T) {
	keep := &Normalizer{dict: defaultNormalizer.dict, keepHeadingNumbers: true}
	withNum := sectionReading(keep, Section{Title: "1.2. Quản lý", Text: "Nội dung."})
	if !KeepsHeadingNumbers("1.2. Quản lý", withNum) {
		t.Errorf("lời đọc có số %q phải nhận ra là giữ số", withNum)
	}
	if KeepsHeadingNumbers("1.2. Quản lý", SectionScript(nil, "1.2. Quản lý", "Nội dung.")) {
		t.Error("lời đọc bỏ số không được nhận là giữ số")
	}
	if KeepsHeadingNumbers("Quản lý", "Quản lý.\n\nNội dung.") {
		t.Error("tiêu đề không số: false")
	}
}

func TestAutoCover(t *testing.T) {
	p := filepath.Join(t.TempDir(), "cover.png")
	if err := WriteAutoCover(p, "Sách A", "Tác giả"); err != nil {
		t.Fatal(err)
	}
	if !IsAutoCover(p, "Sách A", "Tác giả") {
		t.Error("bìa vừa vẽ phải nhận ra là bìa tự vẽ")
	}
	if IsAutoCover(p, "Sách B", "Tác giả") {
		t.Error("khác tên không được coi là cùng bìa")
	}
}

func TestRenderScripts_Stub(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("không có ffmpeg")
	}
	dir := t.TempDir()
	cfg := TTSConfig{Mode: TTSModeStub, FFmpeg: "ffmpeg", Bitrate: "64k", StubSec: 1}
	var done []string
	err := RenderScripts(context.Background(), cfg, dir, []ScriptJob{{Stem: "a", Text: "Một."}, {Stem: "b", Text: "Hai."}}, func(s string) { done = append(done, s) })
	if err != nil {
		t.Fatal(err)
	}
	if len(done) != 2 {
		t.Fatalf("done = %v", done)
	}
	if _, err := os.Stat(filepath.Join(dir, "b.mp3")); err != nil {
		t.Error(err)
	}
	if err := RenderScripts(context.Background(), cfg, dir, []ScriptJob{{Stem: "c", Text: " "}}, nil); err == nil {
		t.Error("lời đọc rỗng phải báo lỗi")
	}
}
