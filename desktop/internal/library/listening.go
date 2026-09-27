package library

// Nhật ký nghe: số giây nghe mỗi ngày (theo cuốn, theo giờ trong ngày) và ngày nghe
// xong từng cuốn, làm dữ liệu cho tính năng thống kê. Một file nhỏ ở gốc ~/Sano,
// cạnh .thu-tu.json. Chỉ cộng dồn, không lưu nội dung sách.

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// listenFile — nhật ký nghe ở gốc ~/Sano (ẩn).
const listenFile = ".nghe.json"

// maxListenChunk — số giây tối đa một lần ghi (app ghi mỗi ~30 giây); lớn hơn là
// dữ liệu lạ (đồng hồ máy nhảy, gọi sai) → cắt bớt.
const maxListenChunk = 300

// maxListenDays — số ngày giữ tối đa (~30 năm), chặn file phình vô hạn.
const maxListenDays = 11000

// BookListen — số giây nghe một cuốn trong ngày: Listen là thời gian thật, Audio là
// số giây nội dung đã nghe (nghe tốc độ 1.5× thì Audio > Listen).
type BookListen struct {
	Listen float64 `json:"listen"`
	Audio  float64 `json:"audio"`
}

// DayListen — một ngày nghe: theo cuốn + theo giờ (24 ô, giây thời gian thật).
type DayListen struct {
	Books map[string]*BookListen `json:"books"`
	Hours [24]float64            `json:"hours"`
}

// ListenLog — toàn bộ nhật ký. Days khoá "YYYY-MM-DD" theo giờ máy; Finished là
// ngày nghe xong lần đầu của từng cuốn.
type ListenLog struct {
	Version  int                   `json:"version"`
	Days     map[string]*DayListen `json:"days"`
	Finished map[string]string     `json:"finished"`
}

var listenMu sync.Mutex

func (l *Library) listenPath() string { return filepath.Join(l.root, listenFile) }

// ListenLog đọc nhật ký nghe. Chưa có file hoặc file hỏng → nhật ký rỗng.
func (l *Library) ListenLog() (*ListenLog, error) {
	listenMu.Lock()
	defer listenMu.Unlock()
	return l.readListen()
}

func (l *Library) readListen() (*ListenLog, error) {
	d := &ListenLog{Version: 1, Days: map[string]*DayListen{}, Finished: map[string]string{}}
	data, err := readJSONFile(l.listenPath())
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return nil, fmt.Errorf("đọc nhật ký nghe: %w", err)
	}
	if json.Unmarshal(data, d) != nil {
		// file hỏng: giữ lại bản hỏng để không mất hẳn, bắt đầu nhật ký mới
		_ = os.Rename(l.listenPath(), l.listenPath()+".hong")
		return &ListenLog{Version: 1, Days: map[string]*DayListen{}, Finished: map[string]string{}}, nil
	}
	if d.Days == nil {
		d.Days = map[string]*DayListen{}
	}
	if d.Finished == nil {
		d.Finished = map[string]string{}
	}
	return d, nil
}

func (l *Library) writeListen(d *ListenLog) error {
	if len(d.Days) > maxListenDays {
		return fmt.Errorf("nhật ký nghe quá lớn (%d ngày)", len(d.Days))
	}
	if err := os.MkdirAll(l.root, 0o755); err != nil {
		return err
	}
	out, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return writeFileAtomic(l.listenPath(), out)
}

func clampSec(v float64, max float64) float64 {
	if math.IsNaN(v) || v <= 0 {
		return 0
	}
	return math.Min(v, max)
}

// RecordListening cộng số giây vừa nghe cuốn slug vào ngày + giờ của lúc now.
func (l *Library) RecordListening(slug string, listenSec, audioSec float64, now time.Time) error {
	if !validSlug(slug) {
		return fmt.Errorf("slug không hợp lệ: %q", slug)
	}
	listenSec = clampSec(listenSec, maxListenChunk)
	audioSec = clampSec(audioSec, maxListenChunk*4)
	if listenSec == 0 && audioSec == 0 {
		return nil
	}
	listenMu.Lock()
	defer listenMu.Unlock()
	d, err := l.readListen()
	if err != nil {
		return err
	}
	key := now.Format("2006-01-02")
	day := d.Days[key]
	if day == nil {
		day = &DayListen{Books: map[string]*BookListen{}}
		d.Days[key] = day
	}
	if day.Books == nil {
		day.Books = map[string]*BookListen{}
	}
	b := day.Books[slug]
	if b == nil {
		b = &BookListen{}
		day.Books[slug] = b
	}
	b.Listen += listenSec
	b.Audio += audioSec
	day.Hours[now.Hour()] += listenSec
	return l.writeListen(d)
}

// MarkFinished ghi ngày nghe xong cuốn slug (chỉ lần đầu; nghe lại không ghi đè).
func (l *Library) MarkFinished(slug string, now time.Time) error {
	if !validSlug(slug) {
		return fmt.Errorf("slug không hợp lệ: %q", slug)
	}
	listenMu.Lock()
	defer listenMu.Unlock()
	d, err := l.readListen()
	if err != nil {
		return err
	}
	if _, ok := d.Finished[slug]; ok {
		return nil
	}
	d.Finished[slug] = now.Format("2006-01-02")
	return l.writeListen(d)
}
