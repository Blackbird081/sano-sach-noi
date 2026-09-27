package library

// Sửa nội dung một cuốn đã tạo (màn Sửa sách, wireframe D11): đọc lời từng tiểu
// mục để sửa, thay MP3 của tiểu mục vừa đọc lại, vẽ lại / đổi bìa, đổi giọng cả
// cuốn (đọc vào thư mục làm riêng, xong hết mới thay), rồi đóng gói lại zip.
//
// Thư mục làm nằm TRONG thư mục sách (.doc-lai/) để đổi tên file là thay ngay,
// không chép qua ổ khác, và đổi giọng dở dang còn đó cho lần mở sau đọc tiếp.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"sano/internal/bookmaker"
	"sano/internal/safepath"
)

// bookFilesMu — một lượt ghi metadata.json / gói zip của sách tại một thời điểm
// (đọc lại chạy nền trong khi người dùng sửa tên, đổi bìa).
var bookFilesMu sync.Mutex

const (
	reworkDir    = ".doc-lai"   // thư mục làm trong thư mục sách
	sectionsWork = "muc"        // đọc lại từng tiểu mục
	voiceWork    = "giong"      // đổi giọng cả cuốn
	voiceJobFile = "giong.json" // trạng thái đổi giọng (giọng mới, các mục đã xong)
)

// EditSection — một tiểu mục để sửa.
type EditSection struct {
	Index        int    `json:"index"` // thứ tự trong cả cuốn (khớp Detail.Tracks)
	ChapterIndex int    `json:"chapterIndex"`
	Chapter      string `json:"chapter"`
	Title        string `json:"title"`
	Text         string `json:"text"`   // chữ gốc (hiện khi nghe, sửa ở màn Sửa sách)
	Script       string `json:"script"` // lời đọc đang dùng
	Intro        bool   `json:"intro"`  // lời mở đầu "Bạn đang nghe sách nói…"
	File         string `json:"file"`   // MP3, tương đối với Root()
	Stem         string `json:"stem"`   // tên file không đuôi (ch01-sec01)
	DurationSec  int    `json:"durationSec"`
}

// VoiceJob — đổi giọng đang dở của một cuốn.
type VoiceJob struct {
	Voice string   `json:"voice"`
	Done  []string `json:"done"` // stem đã đọc xong bằng giọng mới
}

// EditInfo — đủ thông tin cho màn Sửa sách.
type EditInfo struct {
	Book
	CoverAuto bool          `json:"coverAuto"` // bìa đang là bìa tự vẽ theo tên
	Sections  []EditSection `json:"sections"`
	VoiceJob  *VoiceJob     `json:"voiceJob"` // nil = không có đổi giọng dở
}

// rawBook — metadata.json đọc dạng thô để ghi lại không mất trường nào.
type rawBook struct {
	meta     map[string]json.RawMessage
	chapters []rawChapter
}

type rawChapter struct {
	m    map[string]json.RawMessage
	secs []map[string]json.RawMessage
}

func readRawBook(dir string) (*rawBook, error) {
	data, err := readJSONFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return nil, fmt.Errorf("đọc metadata.json: %w", err)
	}
	rb := &rawBook{}
	if err := json.Unmarshal(data, &rb.meta); err != nil {
		return nil, fmt.Errorf("đọc metadata.json: %w", err)
	}
	var chs []map[string]json.RawMessage
	if raw, ok := rb.meta["chapters"]; ok {
		if err := json.Unmarshal(raw, &chs); err != nil {
			return nil, fmt.Errorf("đọc mục lục trong metadata.json: %w", err)
		}
	}
	for _, c := range chs {
		rc := rawChapter{m: c}
		if raw, ok := c["sections"]; ok {
			if err := json.Unmarshal(raw, &rc.secs); err != nil {
				return nil, fmt.Errorf("đọc mục lục trong metadata.json: %w", err)
			}
		}
		rb.chapters = append(rb.chapters, rc)
	}
	return rb, nil
}

func (rb *rawBook) write(dir string) error {
	chs := make([]map[string]json.RawMessage, 0, len(rb.chapters))
	for _, c := range rb.chapters {
		b, err := json.Marshal(c.secs)
		if err != nil {
			return err
		}
		c.m["sections"] = b
		chs = append(chs, c.m)
	}
	b, err := json.Marshal(chs)
	if err != nil {
		return err
	}
	rb.meta["chapters"] = b
	out, err := marshalIndent(rb.meta)
	if err != nil {
		return err
	}
	mtime := dirModTime(dir)
	if err := writeFileAtomic(filepath.Join(dir, "metadata.json"), out); err != nil {
		return fmt.Errorf("ghi metadata.json: %w", err)
	}
	if !mtime.IsZero() {
		_ = os.Chtimes(dir, mtime, mtime) // giữ thứ tự "Mới tạo nhất"
	}
	return nil
}

func str(m map[string]json.RawMessage, k string) string {
	var s string
	_ = json.Unmarshal(m[k], &s)
	return s
}

// section trả tiểu mục thứ index (đếm cả cuốn) cùng chỉ số chương.
func (rb *rawBook) section(index int) (map[string]json.RawMessage, int, error) {
	i := 0
	for ci, c := range rb.chapters {
		for _, s := range c.secs {
			if i == index {
				return s, ci, nil
			}
			i++
		}
	}
	return nil, 0, fmt.Errorf("không có tiểu mục số %d", index+1)
}

// introMarkerRe — dòng tên sách trong lời mở đầu mặc định ("Cuốn sách: X.").
var introMarkerRe = regexp.MustCompile(`(?m)^Cuốn sách:[^\n]*$`)

// isIntro: chương đầu chỉ có một mục ch01-sec01, chữ có dòng "Cuốn sách: …"
// (lời mở đầu do phần mềm dựng lúc tạo sách).
func (rb *rawBook) isIntro(ci int) bool {
	if ci != 0 || len(rb.chapters) < 2 || len(rb.chapters[0].secs) != 1 {
		return false
	}
	s := rb.chapters[0].secs[0]
	return str(s, "file") == "ch01-sec01.mp3" && introMarkerRe.MatchString(str(s, "original_text"))
}

// RetitleIntro sửa lời mở đầu theo tên sách / tác giả mới: dòng "Cuốn sách: …"
// và "Tác giả: …" (thêm nếu chưa có, bỏ nếu tác giả trống). Các dòng khác giữ nguyên.
func RetitleIntro(text, title, author string) string {
	text = introMarkerRe.ReplaceAllLiteralString(text, "Cuốn sách: "+strings.TrimRight(title, ". ")+".")
	authorRe := regexp.MustCompile(`(?m)\n*^Tác giả:[^\n]*$`)
	text = authorRe.ReplaceAllLiteralString(text, "")
	if a := strings.TrimRight(strings.TrimSpace(author), ". "); a != "" {
		loc := introMarkerRe.FindStringIndex(text)
		if loc != nil {
			text = text[:loc[1]] + "\n\nTác giả: " + a + "." + text[loc[1]:]
		}
	}
	return strings.TrimSpace(text)
}

// Edit trả thông tin sửa của một cuốn: các tiểu mục kèm chữ, bìa tự vẽ hay
// không, đổi giọng dở dang.
func (l *Library) Edit(slug string) (*EditInfo, error) {
	d, err := l.Get(slug)
	if err != nil {
		return nil, err
	}
	dir, err := l.Dir(slug)
	if err != nil {
		return nil, err
	}
	rb, err := readRawBook(dir)
	if err != nil {
		return nil, err
	}
	out := &EditInfo{Book: d.Book, Sections: []EditSection{}}
	i := 0
	for ci, c := range rb.chapters {
		for _, s := range c.secs {
			file := str(s, "file")
			out.Sections = append(out.Sections, EditSection{
				Index: i, ChapterIndex: ci, Chapter: str(c.m, "title"), Title: str(s, "title"),
				Text: strings.TrimSpace(str(s, "original_text")), Script: strings.TrimSpace(str(s, "reading_script")),
				Intro: rb.isIntro(ci), File: filepath.ToSlash(filepath.Join(BooksDir, slug, file)),
				Stem: strings.TrimSuffix(file, filepath.Ext(file)),
			})
			if i < len(d.Tracks) {
				out.Sections[i].DurationSec = d.Tracks[i].DurationSec
			}
			i++
		}
	}
	out.CoverAuto = l.coverAuto(dir, rb)
	if j, err := readVoiceJob(dir); err == nil {
		out.VoiceJob = j
	}
	return out, nil
}

// coverAuto: bìa hiện tại là bìa tự vẽ. Ưu tiên cờ cover_auto (ghi khi Sano vẽ
// lại / đổi bìa); sách cũ chưa có cờ thì so từng byte với bìa vẽ theo tên hiện
// tại hoặc tên lúc tạo (tên chương lời mở đầu).
func (l *Library) coverAuto(dir string, rb *rawBook) bool {
	if raw, ok := rb.meta["cover_auto"]; ok {
		var v bool
		_ = json.Unmarshal(raw, &v)
		return v
	}
	name := str(rb.meta, "cover")
	if name == "" {
		return true // chưa có bìa: phần mềm hiện bìa tự vẽ
	}
	if name != "cover.png" || !safepath.IsPlainName(name) {
		return false
	}
	p := filepath.Join(dir, name)
	author := str(rb.meta, "author")
	titles := []string{str(rb.meta, "title")}
	if len(rb.chapters) > 0 {
		titles = append(titles, str(rb.chapters[0].m, "title"))
	}
	for _, t := range titles {
		if t != "" && bookmaker.IsAutoCover(p, t, author) {
			return true
		}
	}
	return false
}

// ApplySection thay MP3 của tiểu mục index bằng file mp3 vừa đọc lại (cùng thư
// mục sách, đổi tên là xong) và ghi tiêu đề, chữ, lời đọc mới vào metadata.json.
// Lời mở đầu: tên chương đi theo tên mục. Gói zip đóng lại sau (Repack).
func (l *Library) ApplySection(slug string, index int, title, text, script, mp3 string) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	rb, err := readRawBook(dir)
	if err != nil {
		return err
	}
	sec, ci, err := rb.section(index)
	if err != nil {
		return err
	}
	file := str(sec, "file")
	if !safepath.IsPlainName(file) || !strings.EqualFold(filepath.Ext(file), ".mp3") {
		return fmt.Errorf("tên file tiểu mục không hợp lệ: %q", file)
	}
	if err := os.Rename(mp3, filepath.Join(dir, file)); err != nil {
		return fmt.Errorf("thay file âm thanh: %w", err)
	}
	if txt := strings.TrimSuffix(mp3, ".mp3") + ".txt"; fileExists(txt) {
		_ = os.Rename(txt, filepath.Join(dir, strings.TrimSuffix(file, ".mp3")+".txt"))
	}
	title = clip(strings.TrimSpace(title), maxTitleLen)
	if title != "" {
		setJSON(sec, "title", title)
		if rb.isIntro(ci) {
			setJSON(rb.chapters[ci].m, "title", title)
		}
	}
	setJSON(sec, "original_text", strings.TrimSpace(text))
	setJSON(sec, "reading_script", strings.TrimSpace(script))
	// Đổi giọng đang dở đã đọc mục này bằng chữ cũ → bỏ để lượt sau đọc lại chữ mới.
	_ = dropVoiceDone(dir, strings.TrimSuffix(file, ".mp3"))
	return rb.write(dir)
}

// Repack đóng gói lại zip của cuốn từ thư mục sách (thời lượng, chữ, bìa mới).
// voice trống = giữ giọng đang ghi trong gói zip.
func (l *Library) Repack(slug, voice string) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	return l.repackLocked(slug, dir, voice)
}

func (l *Library) repackLocked(slug, dir, voice string) error {
	zipName := findZip(dir)
	if zipName == "" {
		zipName = "book-" + slug + ".zip"
	} else if voice == "" {
		_, voice = zipInfo(filepath.Join(dir, zipName))
	}
	if voice == "" {
		voice = bookmaker.DefaultVoice
	}
	mtime := dirModTime(dir)
	defer func() {
		if !mtime.IsZero() {
			_ = os.Chtimes(dir, mtime, mtime)
		}
	}()
	return bookmaker.RepackZipAtomic(dir, filepath.Join(dir, zipName), voice, nil)
}

// RedrawCover vẽ lại bìa theo tên, tác giả hiện tại (cover.png) và đánh dấu bìa
// tự vẽ. Ảnh bìa riêng (nếu có) bị thay. Gói zip đóng lại luôn.
func (l *Library) RedrawCover(slug string) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	rb, err := readRawBook(dir)
	if err != nil {
		return err
	}
	if err := bookmaker.WriteAutoCover(filepath.Join(dir, "cover.png"), str(rb.meta, "title"), str(rb.meta, "author")); err != nil {
		return fmt.Errorf("vẽ bìa: %w", err)
	}
	return l.setCoverLocked(slug, dir, rb, "cover.png", true)
}

// SetCoverImage dùng ảnh src (jpg/png/webp, đã kiểm dung lượng) làm bìa.
func (l *Library) SetCoverImage(slug, src string) error {
	ext := strings.ToLower(filepath.Ext(src))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return errors.New("ảnh bìa phải là jpg, png hoặc webp")
	}
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("đọc ảnh bìa: %w", err)
	}
	if data, ext, err = bookmaker.FitCover(data, ext); err != nil { // ảnh chụp lớn → thu về 1200×1600
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	rb, err := readRawBook(dir)
	if err != nil {
		return err
	}
	name := "cover" + ext
	if err := writeFileAtomic(filepath.Join(dir, name), data); err != nil {
		return fmt.Errorf("ghi bìa: %w", err)
	}
	return l.setCoverLocked(slug, dir, rb, name, false)
}

func (l *Library) setCoverLocked(slug, dir string, rb *rawBook, name string, auto bool) error {
	if old := str(rb.meta, "cover"); old != "" && old != name && safepath.IsPlainName(old) && strings.HasPrefix(old, "cover") {
		_ = os.Remove(filepath.Join(dir, old)) // bìa cũ là bản chép trong thư mục sách
	}
	setJSON(rb.meta, "cover", name)
	b, _ := json.Marshal(auto)
	rb.meta["cover_auto"] = b
	if err := rb.write(dir); err != nil {
		return err
	}
	return l.repackLocked(slug, dir, "")
}

// ── Đổi giọng cả cuốn ───────────────────────────────────────────────────────

// VoiceWorkDir — thư mục đọc giọng mới của cuốn (tạo nếu chưa có).
func (l *Library) VoiceWorkDir(slug string) (string, error) {
	dir, err := l.Dir(slug)
	if err != nil {
		return "", err
	}
	w := filepath.Join(dir, reworkDir, voiceWork)
	if err := os.MkdirAll(w, 0o755); err != nil {
		return "", fmt.Errorf("tạo thư mục đổi giọng: %w", err)
	}
	return w, nil
}

// SectionsWorkDir — thư mục đọc lại từng tiểu mục (làm mới mỗi lượt).
func (l *Library) SectionsWorkDir(slug string) (string, error) {
	dir, err := l.Dir(slug)
	if err != nil {
		return "", err
	}
	w := filepath.Join(dir, reworkDir, sectionsWork)
	_ = os.RemoveAll(w)
	if err := os.MkdirAll(w, 0o755); err != nil {
		return "", fmt.Errorf("tạo thư mục đọc lại: %w", err)
	}
	return w, nil
}

func voiceJobPath(dir string) string { return filepath.Join(dir, reworkDir, voiceJobFile) }

func readVoiceJob(dir string) (*VoiceJob, error) {
	data, err := readJSONFile(voiceJobPath(dir))
	if err != nil {
		return nil, err
	}
	var j VoiceJob
	if err := json.Unmarshal(data, &j); err != nil || strings.TrimSpace(j.Voice) == "" {
		return nil, errors.New("trạng thái đổi giọng hỏng")
	}
	return &j, nil
}

func writeVoiceJob(dir string, j *VoiceJob) error {
	b, err := marshalIndent(j)
	if err != nil {
		return err
	}
	return writeFileAtomic(voiceJobPath(dir), b)
}

// BeginVoice bắt đầu (hoặc nối tiếp) đổi giọng sang voice. Đang dở một giọng
// khác thì bỏ phần cũ. Trả trạng thái (các mục đã xong được bỏ qua).
func (l *Library) BeginVoice(slug, voice string) (*VoiceJob, error) {
	dir, err := l.Dir(slug)
	if err != nil {
		return nil, err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	if j, err := readVoiceJob(dir); err == nil && j.Voice == voice {
		return j, nil
	}
	_ = os.RemoveAll(filepath.Join(dir, reworkDir, voiceWork))
	if err := os.MkdirAll(filepath.Join(dir, reworkDir, voiceWork), 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục đổi giọng: %w", err)
	}
	j := &VoiceJob{Voice: voice, Done: []string{}}
	return j, writeVoiceJob(dir, j)
}

// MarkVoiceDone ghi nhận mục stem đã đọc xong bằng giọng mới.
func (l *Library) MarkVoiceDone(slug, stem string) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	j, err := readVoiceJob(dir)
	if err != nil {
		return err
	}
	j.Done = append(j.Done, stem)
	return writeVoiceJob(dir, j)
}

// dropVoiceDone bỏ stem khỏi danh sách đã đọc giọng mới (gọi khi đang giữ khoá).
func dropVoiceDone(dir, stem string) error {
	j, err := readVoiceJob(dir)
	if err != nil {
		return nil
	}
	kept := j.Done[:0]
	for _, s := range j.Done {
		if s != stem {
			kept = append(kept, s)
		}
	}
	j.Done = kept
	_ = os.Remove(filepath.Join(dir, reworkDir, voiceWork, stem+".mp3"))
	return writeVoiceJob(dir, j)
}

// FinishVoice thay toàn bộ MP3 bằng bản giọng mới, ghi giọng vào metadata và
// gói zip, dọn thư mục làm. Thiếu mục nào thì báo lỗi, không thay gì.
func (l *Library) FinishVoice(slug string) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	j, err := readVoiceJob(dir)
	if err != nil {
		return err
	}
	rb, err := readRawBook(dir)
	if err != nil {
		return err
	}
	work := filepath.Join(dir, reworkDir, voiceWork)
	done := map[string]bool{}
	for _, s := range j.Done {
		done[s] = true
	}
	var files []string
	for _, c := range rb.chapters {
		for _, s := range c.secs {
			f := str(s, "file")
			stem := strings.TrimSuffix(f, ".mp3")
			if !safepath.IsPlainName(f) || !done[stem] || !fileExists(filepath.Join(work, f)) {
				return fmt.Errorf("chưa đọc xong tiểu mục %s bằng giọng mới", stem)
			}
			files = append(files, f)
		}
	}
	for _, f := range files {
		if err := os.Rename(filepath.Join(work, f), filepath.Join(dir, f)); err != nil {
			return fmt.Errorf("thay file âm thanh: %w", err)
		}
		txt := strings.TrimSuffix(f, ".mp3") + ".txt"
		if fileExists(filepath.Join(work, txt)) {
			_ = os.Rename(filepath.Join(work, txt), filepath.Join(dir, txt))
		}
	}
	setJSON(rb.meta, "narrator", "VieNeu-TTS ("+j.Voice+")")
	if err := rb.write(dir); err != nil {
		return err
	}
	if err := l.repackLocked(slug, dir, j.Voice); err != nil {
		return err
	}
	_ = os.Remove(voiceJobPath(dir))
	err = os.RemoveAll(work)
	_ = os.Remove(filepath.Join(dir, reworkDir))
	return err
}

// CancelVoice bỏ đổi giọng đang dở (giữ giọng cũ).
func (l *Library) CancelVoice(slug string) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	_ = os.Remove(voiceJobPath(dir))
	err = os.RemoveAll(filepath.Join(dir, reworkDir, voiceWork))
	_ = os.Remove(filepath.Join(dir, reworkDir))
	return err
}

// PendingVoiceJobs — các cuốn đang đổi giọng dở (lần mở trước tắt giữa chừng).
func (l *Library) PendingVoiceJobs() map[string]*VoiceJob {
	out := map[string]*VoiceJob{}
	entries, err := os.ReadDir(l.BooksRoot())
	if err != nil {
		return out
	}
	for _, e := range entries {
		if !e.IsDir() || !validSlug(e.Name()) {
			continue
		}
		if j, err := readVoiceJob(filepath.Join(l.BooksRoot(), e.Name())); err == nil {
			out[e.Name()] = j
		}
	}
	return out
}

// StemIndex — "ch03-sec02" → chỉ số tiểu mục trong cả cuốn theo sections (-1 nếu không có).
func StemIndex(secs []EditSection, stem string) int {
	for _, s := range secs {
		if s.Stem == stem {
			return s.Index
		}
	}
	return -1
}
