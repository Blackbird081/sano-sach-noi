package prompts

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestSkillZip(t *testing.T) {
	data, err := SkillZip()
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) != 1 || zr.File[0].Name != "sano-sach-noi/SKILL.md" {
		t.Fatalf("zip phải chỉ có sano-sach-noi/SKILL.md, được %v", zr.File)
	}
	rc, _ := zr.File[0].Open()
	md, _ := io.ReadAll(rc)
	s := string(md)
	if !strings.HasPrefix(s, "---\nname: sano-sach-noi\ndescription: ") {
		t.Errorf("thiếu frontmatter: %q", s[:60])
	}
	for _, want := range []string{"VIỆC A", "VIỆC B", "VIỆC C", "Ba ý cần nhớ", "khối mã"} {
		if !strings.Contains(s, want) {
			t.Errorf("SKILL.md thiếu %q", want)
		}
	}
}

func TestGuideText(t *testing.T) {
	g := GuideText()
	if strings.Contains(g, "name: sano-sach-noi") {
		t.Error("file hướng dẫn không cần frontmatter")
	}
	if !strings.Contains(g, strings.TrimSpace(rewrite)[:40]) || !strings.Contains(g, strings.TrimSpace(smooth)[:40]) {
		t.Error("file hướng dẫn phải chứa nguyên prompt cấp 2 và cấp 3")
	}
}
