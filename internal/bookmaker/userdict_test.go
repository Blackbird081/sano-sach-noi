package bookmaker

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewNormalizerWith_LayersOverride(t *testing.T) {
	n, err := NewNormalizerWith(false,
		map[string]string{"Nielsen": "Niu-sen", "KPI": "cây pi ai"},
		map[string]string{"KPI": "ca pê i nè", "CEO": ""},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := n.script("Nielsen đo KPI của CEO.")
	if got != "Niu-sen đo ca pê i nè của CEO." {
		t.Errorf("script = %q", got)
	}
	if def := normalizeReadingScript("KPI"); def != "ca pê i" {
		t.Errorf("bộ chuẩn không được đổi theo lớp người dùng: %q", def)
	}
}

func TestCountWord_WholeWordCaseSensitive(t *testing.T) {
	if n := CountWord("KPI, KPIs, kpi và KPI.", "KPI"); n != 2 {
		t.Errorf("CountWord = %d, muốn 2", n)
	}
	if !ValidPronunciationKey("R&D") || ValidPronunciationKey("Hồ Chí") || ValidPronunciationKey("") {
		t.Error("ValidPronunciationKey sai")
	}
	m, err := ParsePronunciations(FormatPronunciations(map[string]string{"B": "bê", "A": "a  a"}), "t")
	if err != nil || m["A"] != "a a" || m["B"] != "bê" {
		t.Errorf("TSV khứ hồi: %v %v", m, err)
	}
}

func TestRun_WritesBookDictAndZipEntry(t *testing.T) {
	needFFmpeg(t)
	o := stubOpts(t, writeSample(t))
	o.BookPronunciations = map[string]string{"Nielsen": "Niu-sen"}
	o.OutputZip = filepath.Join(o.OutputDir, "book.zip")
	if _, err := Run(o); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(o.OutputDir, BookPronunciationsFile)); err != nil || !strings.Contains(string(b), "Nielsen\tNiu-sen") {
		t.Fatalf("tu-dien.tsv: %q %v", b, err)
	}
	zr, err := zip.OpenReader(o.OutputZip)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	found := false
	for _, f := range zr.File {
		if f.Name == ZipPronunciationsEntry {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			found = strings.Contains(string(b), "Niu-sen")
		}
	}
	if !found {
		t.Error("gói zip thiếu pronunciations.tsv")
	}
}

func TestUserDict_SilencesAcronymWarning(t *testing.T) {
	texts := []string{"Đội XYZQ báo cáo, XYZQ tăng trưởng."}
	if len(unknownAcronyms(texts, defaultNormalizer.dict)) == 0 {
		t.Fatal("XYZQ phải bị cảnh báo khi chưa có trong từ điển")
	}
	n, _ := NewNormalizerWith(false, map[string]string{"XYZQ": "ích ây dét quy"})
	if got := unknownAcronyms(texts, n.dict); len(got) != 0 {
		t.Errorf("đã có trong từ điển người dùng mà vẫn cảnh báo: %v", got)
	}
}
