package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sano/internal/bookmaker"
)

func TestGlobalDict_SetDelete(t *testing.T) {
	lib := New(t.TempDir())
	if m, err := lib.GlobalDict(); err != nil || len(m) != 0 {
		t.Fatalf("chưa có file: %v %v", m, err)
	}
	if err := lib.SetGlobalWord("Nielsen", "  Niu-sen "); err != nil {
		t.Fatal(err)
	}
	if err := lib.SetGlobalWord("Hồ Chí", "x"); err != ErrBadWord {
		t.Errorf("cụm nhiều từ phải bị từ chối: %v", err)
	}
	if err := lib.SetGlobalWord("KPI", " "); err == nil {
		t.Error("cách đọc rỗng phải báo lỗi")
	}
	m, _ := lib.GlobalDict()
	if m["Nielsen"] != "Niu-sen" {
		t.Errorf("m = %v", m)
	}
	if err := lib.DeleteGlobalWord("Nielsen"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(lib.Root(), globalDictFile)); !os.IsNotExist(err) {
		t.Error("xoá hết từ thì bỏ file")
	}
}

func TestBookDict_SavedAndRepacked(t *testing.T) {
	lib, dir := makeReworkBook(t)
	if err := lib.SetBookWord("sach-thu", "KPI", "cây pi ai"); err != nil {
		t.Fatal(err)
	}
	m, err := lib.BookDict("sach-thu")
	if err != nil || m["KPI"] != "cây pi ai" {
		t.Fatalf("%v %v", m, err)
	}
	if err := lib.Repack("sach-thu", ""); err != nil {
		t.Fatal(err)
	}
	names, contents := readZip(t, filepath.Join(dir, "book-sach-thu.zip"))
	if !strings.Contains(strings.Join(names, ","), bookmaker.ZipPronunciationsEntry) || !strings.Contains(contents[bookmaker.ZipPronunciationsEntry], "cây pi ai") {
		t.Errorf("gói zip thiếu từ điển: %v", names)
	}
	if err := lib.DeleteBookWord("sach-thu", "KPI"); err != nil {
		t.Fatal(err)
	}
	if m, _ := lib.BookDict("sach-thu"); len(m) != 0 {
		t.Errorf("còn %v", m)
	}
}
