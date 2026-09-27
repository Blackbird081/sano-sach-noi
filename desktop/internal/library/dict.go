package library

// Từ điển cách đọc của người dùng (wireframe D12): từ điển chung ~/Sano/.tu-dien.tsv
// (mọi sách) và từ điển riêng tu-dien.tsv trong thư mục từng cuốn. Cùng định dạng TSV
// với bộ chuẩn của bookmaker; lớp riêng của cuốn ghi đè lớp chung.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"sano/internal/bookmaker"
)

// globalDictFile — từ điển chung, cạnh thư mục Sach (ẩn, như .nghe.json).
const globalDictFile = ".tu-dien.tsv"

// maxDictEntries / maxReadingLen — giới hạn từ điển người dùng.
const (
	maxDictEntries = 2000
	maxReadingLen  = 120
	maxDictBytes   = 256 << 10
)

var globalDictMu sync.Mutex

// ErrBadWord — từ không tra được trong từ điển (có dấu cách, ký hiệu lạ...).
var ErrBadWord = errors.New("chỉ nhận một từ: chữ và số liền nhau, có thể nối bằng & (vd KPI, Nielsen, R&D)")

func readDictFile(path string) (map[string]string, error) {
	data, err := readJSONFile(path) // file thường, không theo symlink, có giới hạn dung lượng
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("đọc từ điển: %w", err)
	}
	if len(data) > maxDictBytes {
		return nil, errors.New("file từ điển quá lớn")
	}
	m, err := bookmaker.ParsePronunciations(string(data), filepath.Base(path))
	if err != nil {
		return nil, err
	}
	return m, nil
}

func writeDictFile(path string, m map[string]string) error {
	if len(m) == 0 {
		err := os.Remove(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return writeFileAtomic(path, []byte(bookmaker.FormatPronunciations(m)))
}

// cleanEntry kiểm tra một cặp từ → cách đọc.
func cleanEntry(word, reading string) (string, string, error) {
	word = strings.TrimSpace(word)
	if !bookmaker.ValidPronunciationKey(word) {
		return "", "", ErrBadWord
	}
	reading = strings.Join(strings.Fields(reading), " ")
	if reading == "" {
		return "", "", errors.New("chưa gõ cách đọc")
	}
	if utf8.RuneCountInString(reading) > maxReadingLen {
		return "", "", fmt.Errorf("cách đọc dài quá (tối đa %d ký tự)", maxReadingLen)
	}
	return word, reading, nil
}

// GlobalDict — từ điển chung (mọi sách).
func (l *Library) GlobalDict() (map[string]string, error) {
	globalDictMu.Lock()
	defer globalDictMu.Unlock()
	return readDictFile(filepath.Join(l.root, globalDictFile))
}

// SetGlobalWord thêm / sửa một từ trong từ điển chung.
func (l *Library) SetGlobalWord(word, reading string) error {
	word, reading, err := cleanEntry(word, reading)
	if err != nil {
		return err
	}
	globalDictMu.Lock()
	defer globalDictMu.Unlock()
	path := filepath.Join(l.root, globalDictFile)
	m, err := readDictFile(path)
	if err != nil {
		return err
	}
	if _, ok := m[word]; !ok && len(m) >= maxDictEntries {
		return fmt.Errorf("từ điển đã đủ %d từ", maxDictEntries)
	}
	m[word] = reading
	if err := os.MkdirAll(l.root, 0o755); err != nil {
		return err
	}
	return writeDictFile(path, m)
}

// DeleteGlobalWord xoá một từ khỏi từ điển chung (từ chuẩn trở lại cách đọc có sẵn).
func (l *Library) DeleteGlobalWord(word string) error {
	globalDictMu.Lock()
	defer globalDictMu.Unlock()
	path := filepath.Join(l.root, globalDictFile)
	m, err := readDictFile(path)
	if err != nil {
		return err
	}
	delete(m, strings.TrimSpace(word))
	return writeDictFile(path, m)
}

// BookDict — từ điển riêng của một cuốn.
func (l *Library) BookDict(slug string) (map[string]string, error) {
	dir, err := l.Dir(slug)
	if err != nil {
		return nil, err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	return readDictFile(filepath.Join(dir, bookmaker.BookPronunciationsFile))
}

// SetBookWord thêm / sửa một từ trong từ điển của cuốn. Gói zip đóng lại sau (Repack).
func (l *Library) SetBookWord(slug, word, reading string) error {
	word, reading, err := cleanEntry(word, reading)
	if err != nil {
		return err
	}
	return l.editBookDict(slug, func(m map[string]string) error {
		if _, ok := m[word]; !ok && len(m) >= maxDictEntries {
			return fmt.Errorf("từ điển đã đủ %d từ", maxDictEntries)
		}
		m[word] = reading
		return nil
	})
}

// DeleteBookWord xoá một từ khỏi từ điển của cuốn.
func (l *Library) DeleteBookWord(slug, word string) error {
	return l.editBookDict(slug, func(m map[string]string) error {
		delete(m, strings.TrimSpace(word))
		return nil
	})
}

func (l *Library) editBookDict(slug string, fn func(map[string]string) error) error {
	dir, err := l.Dir(slug)
	if err != nil {
		return err
	}
	bookFilesMu.Lock()
	defer bookFilesMu.Unlock()
	path := filepath.Join(dir, bookmaker.BookPronunciationsFile)
	m, err := readDictFile(path)
	if err != nil {
		return err
	}
	if err := fn(m); err != nil {
		return err
	}
	mtime := dirModTime(dir)
	err = writeDictFile(path, m)
	if !mtime.IsZero() {
		_ = os.Chtimes(dir, mtime, mtime) // giữ thứ tự "Mới tạo nhất"
	}
	return err
}
