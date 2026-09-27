package bookmaker

// Sửa sách đã tạo (desktop "Sửa sách", wireframe D11): dựng lại lời đọc cho tiểu
// mục đã sửa, đọc lại một nhóm tiểu mục, vẽ lại bìa, đóng gói lại zip. Không cần
// file Word gốc: mọi thứ lấy từ thư mục sách đã render.

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"sano/internal/cover"
)

// SectionScript dựng lời đọc của một tiểu mục như lúc tạo sách: tiêu đề đã
// chuẩn hóa thành một đoạn riêng rồi tới nội dung (xem sectionReading).
func SectionScript(norm *Normalizer, title, text string) string {
	if norm == nil {
		norm = defaultNormalizer
	}
	return sectionReading(norm, Section{Title: title, Text: text})
}

// IntroScript dựng lời đọc của lời mở đầu: nguyên văn đã chuẩn hóa, không ghép
// tiêu đề (xem buildIntroChapter).
func IntroScript(norm *Normalizer, text string) string {
	if norm == nil {
		norm = defaultNormalizer
	}
	return norm.script(strings.TrimSpace(text))
}

// KeepsHeadingNumbers đoán cuốn sách lúc tạo có chọn đọc số mục không, từ lời
// đọc cũ của một tiểu mục: lời đọc bắt đầu bằng tiêu đề có số ("một chấm hai.")
// mà bản bỏ số thì không. Không nhận ra → false (mặc định lúc tạo sách).
func KeepsHeadingNumbers(title, oldScript string) bool {
	keep := &Normalizer{dict: defaultNormalizer.dict, keepHeadingNumbers: true}
	withNum := strings.TrimRight(keep.script(keep.spokenTitle(title)), " .")
	noNum := strings.TrimRight(defaultNormalizer.script(defaultNormalizer.spokenTitle(title)), " .")
	return withNum != noNum && withNum != "" && strings.HasPrefix(oldScript, withNum)
}

// ScriptJob — một tiểu mục cần đọc: Stem là tên file (không đuôi) trong thư mục làm.
type ScriptJob struct {
	Stem string
	Text string
}

// RenderScripts đọc các lời đọc đã dựng sẵn ra workDir/<stem>.mp3. onDone báo
// từng file xong (có thể nil).
func RenderScripts(ctx context.Context, tts TTSConfig, workDir string, jobs []ScriptJob, onDone func(stem string)) error {
	if len(jobs) == 0 {
		return nil
	}
	if onDone == nil {
		onDone = func(string) {}
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("tạo thư mục đọc lại: %w", err)
	}
	tts.OnlyStems, tts.PreviewChars = nil, 0
	if err := tts.preflight(); err != nil {
		return err
	}
	tj := make([]ttsJob, 0, len(jobs))
	for _, j := range jobs {
		if strings.TrimSpace(j.Text) == "" {
			return fmt.Errorf("tiểu mục %s không có lời đọc", j.Stem)
		}
		tj = append(tj, ttsJob(j))
	}
	return tts.RenderContext(ctx, workDir, tj, onDone)
}

// WriteAutoCover vẽ bìa mặc định theo tên sách vào path (PNG, cùng kích thước
// lúc tạo sách).
func WriteAutoCover(path, title, author string) error {
	return writeFileVia(path, func(w io.Writer) error {
		return cover.WritePNG(w, title, author, coverWidth, coverHeight)
	})
}

// IsAutoCover báo ảnh ở path có đúng là bìa tự vẽ theo title/author không (bìa
// vẽ ra luôn giống hệt nhau từng byte với cùng tên và tác giả).
func IsAutoCover(path, title, author string) bool {
	have, err := os.ReadFile(path)
	if err != nil || len(have) == 0 {
		return false
	}
	var b strings.Builder
	if err := cover.WritePNG(&b, title, author, coverWidth, coverHeight); err != nil {
		return false
	}
	return b.String() == string(have)
}

// RepackZipAtomic như RepackZip nhưng ghi ra file tạm cạnh zipPath rồi mới thay:
// lỗi giữa chừng thì gói zip cũ còn nguyên, trình phát đang đọc không thấy file dở.
func RepackZipAtomic(dir, zipPath, voice string, logf func(string, ...any)) error {
	tmp, err := os.CreateTemp(filepath.Dir(zipPath), ".dong-goi-*.zip")
	if err != nil {
		return fmt.Errorf("tạo file tạm: %w", err)
	}
	name := tmp.Name()
	_ = tmp.Close()
	if _, err := RepackZip(dir, name, voice, logf); err != nil {
		_ = os.Remove(name)
		return err
	}
	_ = os.Chmod(name, 0o644) // CreateTemp tạo 0600; gói sách là file thường như lúc tạo
	if err := os.Rename(name, zipPath); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("thay gói zip: %w", err)
	}
	return nil
}

// writeFileVia ghi file qua fn vào file tạm cùng thư mục rồi đổi tên.
func writeFileVia(path string, fn func(io.Writer) error) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ghi-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if err := fn(tmp); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	_ = os.Chmod(name, 0o644)
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}
