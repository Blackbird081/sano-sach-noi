package main

import (
	"path/filepath"
	"testing"
)

func TestWritePastedDocx(t *testing.T) {
	p := filepath.Join(t.TempDir(), pastedDocxName)
	f, err := writePastedDocx(p, "% Sách\n# Chương 1. Mở đầu\n## Mục một\nChữ.")
	if err != nil {
		t.Fatal(err)
	}
	if f.Path != p || f.Name != PastedDocxLabel || f.Size == 0 {
		t.Errorf("DocxFile = %+v", f)
	}
	if _, err := writePastedDocx(p, "không có chương"); err == nil {
		t.Error("văn bản không có dòng # phải báo lỗi")
	}
}
