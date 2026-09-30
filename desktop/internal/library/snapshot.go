package library

// Bản cũ để hoàn tác (MCP · M3): trước khi AI sửa một cuốn (sửa lời, tìm và thay, đổi giọng,
// thông tin, bìa), Sano giữ bản hiện tại trong ~/Sano/.ban-cu/<slug>/<id>. MP3, gói zip, ảnh chỉ
// làm liên kết cứng (hardlink): mọi lần sửa đều ghi file mới rồi đổi tên nên bản cũ không bị ghi
// đè, gần như không tốn thêm dung lượng. Giữ 3 bản mới nhất mỗi cuốn, tối đa 7 ngày.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	snapshotDir     = ".ban-cu"
	snapshotMeta    = "ban-cu.json"
	snapshotKeep    = 3
	snapshotMaxAge  = 7 * 24 * time.Hour
	restoreTempMark = ".hoan-tac-"
)

var snapshotIDRe = regexp.MustCompile(`^\d{8}-\d{6}-\d{9}$`)

// Snapshot — một bản cũ của cuốn sách.
type Snapshot struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Label string `json:"label"` // việc sắp làm, vd "AI sửa lời 2 mục"
	At    int64  `json:"at"`
}

func (l *Library) snapshotRoot(slug string) string {
	return filepath.Join(l.root, snapshotDir, slug)
}

// TakeSnapshot giữ bản hiện tại của cuốn slug trước khi sửa. Trả mã bản cũ.
func (l *Library) TakeSnapshot(slug, label string) (*Snapshot, error) {
	dir, err := l.Dir(slug)
	if err != nil {
		return nil, err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	rb, err := readRawBook(dir)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	s := &Snapshot{ID: now.Format("20060102-150405") + fmt.Sprintf("-%09d", now.Nanosecond()), Slug: slug,
		Title: str(rb.meta, "title"), Label: label, At: now.Unix()}
	dst := filepath.Join(l.snapshotRoot(slug), s.ID)
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return nil, err
	}
	if err := copyBookTree(dir, dst); err != nil {
		_ = os.RemoveAll(dst)
		return nil, fmt.Errorf("giữ bản cũ: %w", err)
	}
	b, _ := json.Marshal(s)
	if err := os.WriteFile(filepath.Join(dst, snapshotMeta), b, 0o600); err != nil {
		_ = os.RemoveAll(dst)
		return nil, err
	}
	l.pruneSnapshots(slug)
	return s, nil
}

// copyBookTree chép thư mục sách (bỏ thư mục làm .doc-lai): MP3 / zip / ảnh làm liên kết cứng,
// file khác (metadata.json, chữ, bìa) chép thật. Liên kết cứng lỗi (ổ khác, hệ thống không hỗ trợ) thì chép.
func copyBookTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			if d.Name() == reworkDir || (rel != "." && strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() || strings.HasPrefix(d.Name(), ".") {
			return nil // bỏ liên kết mềm, file tạm ẩn
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		inImages := strings.HasPrefix(filepath.ToSlash(rel), "images/")
		if ext == ".mp3" || ext == ".zip" || inImages {
			if os.Link(p, target) == nil {
				return nil
			}
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// Snapshots — các bản cũ còn giữ của cuốn slug (mới nhất trước).
func (l *Library) Snapshots(slug string) []Snapshot {
	if !validSlug(slug) {
		return nil
	}
	entries, err := os.ReadDir(l.snapshotRoot(slug))
	if err != nil {
		return nil
	}
	var out []Snapshot
	for _, e := range entries {
		if !e.IsDir() || !snapshotIDRe.MatchString(e.Name()) {
			continue
		}
		var s Snapshot
		b, err := os.ReadFile(filepath.Join(l.snapshotRoot(slug), e.Name(), snapshotMeta))
		if err != nil || json.Unmarshal(b, &s) != nil || s.ID != e.Name() || s.Slug != slug {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out
}

// pruneSnapshots giữ snapshotKeep bản mới nhất trong 7 ngày của cuốn slug.
func (l *Library) pruneSnapshots(slug string) {
	for i, s := range l.Snapshots(slug) {
		if i >= snapshotKeep || time.Since(time.Unix(s.At, 0)) > snapshotMaxAge {
			_ = os.RemoveAll(filepath.Join(l.snapshotRoot(slug), s.ID))
		}
	}
}

// PruneAllSnapshots dọn bản cũ quá hạn của mọi cuốn (mở app).
func (l *Library) PruneAllSnapshots() {
	entries, err := os.ReadDir(filepath.Join(l.root, snapshotDir))
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || !validSlug(e.Name()) {
			continue
		}
		l.pruneSnapshots(e.Name())
		if left, _ := os.ReadDir(l.snapshotRoot(e.Name())); len(left) == 0 {
			_ = os.Remove(l.snapshotRoot(e.Name()))
		}
	}
}

// ErrSnapshotGone — bản cũ không còn (đã quá hạn / đã dùng để hoàn tác).
var ErrSnapshotGone = errors.New("bản cũ không còn nữa (quá 7 ngày hoặc đã hoàn tác)")

// RestoreSnapshot đưa cuốn slug về bản cũ id. Bản hiện tại bị thay (không giữ lại).
// Bên gọi phải chắc không có lượt sửa nào đang chạy trên cuốn này.
func (l *Library) RestoreSnapshot(slug, id string) error {
	if !snapshotIDRe.MatchString(id) {
		return ErrSnapshotGone
	}
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	src := filepath.Join(l.snapshotRoot(slug), id)
	if _, err := os.Stat(filepath.Join(src, snapshotMeta)); err != nil {
		return ErrSnapshotGone
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	mtime := dirModTime(dir)
	old := filepath.Join(l.BooksRoot(), restoreTempMark+slug+"-"+id)
	if err := os.Rename(dir, old); err != nil {
		return fmt.Errorf("hoàn tác: %w", err)
	}
	if err := os.Rename(src, dir); err != nil {
		_ = os.Rename(old, dir) // trả lại như cũ
		return fmt.Errorf("hoàn tác: %w", err)
	}
	_ = os.Remove(filepath.Join(dir, snapshotMeta))
	_ = os.Chmod(dir, 0o755)
	if !mtime.IsZero() {
		_ = os.Chtimes(dir, mtime, mtime) // giữ thứ tự "Mới tạo nhất"
	}
	_ = os.RemoveAll(old)
	return nil
}
