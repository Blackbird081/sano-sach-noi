package main

// Tạo video cả cuốn (wireframe D15): giao diện dựng dòng thời gian (đoạn tiếng +
// khung hình từng câu), vẽ khung hình bằng canvas rồi gửi dần sang đây
// (BookVideoFrame), cuối cùng gọi BookVideoFinish. Ở đây: cắt/đệm từng đoạn tiếng
// thành WAV đúng số giây rồi nối (chính xác tới mẫu, không lệch dần như nối mp3,
// không vướng giới hạn độ dài dòng lệnh Windows), ghép khung hình + lớp sáng dần
// (sóng âm trích đoạn, thanh tiến độ cả cuốn), mã hoá MP4, ghi thumbnail, phụ đề,
// mô tả vào Tải về/Sano video/<Tên sách>. Chạy nền, tiến độ qua sự kiện, huỷ được.

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/tts"
)

const (
	eventBookVideoProgress = "bookvideo:progress"
	eventBookVideoFinished = "bookvideo:finished"
	maxBookVideoFrames     = 50000
	maxBookVideoSegs       = 5000
	maxBookVideoSec        = 12 * 3600
	maxBookVideoOverlays   = 4
	bookVideoFPS           = 24
	maxFrameImage          = 16 << 20
)

// BookVideoSeg — một đoạn tiếng: File (mp3 của tiểu mục, chỉ lấy tên file) từ
// Start, dài Dur giây; Silence = đoạn lặng Dur giây (màn tựa, thẻ chương, quãng nghỉ).
type BookVideoSeg struct {
	File    string  `json:"file"`
	Start   float64 `json:"start"`
	Dur     float64 `json:"dur"`
	Silence bool    `json:"silence"`
}

// BookVideoOverlay — ảnh sáng (PNG) đặt ở (X, Y), hiện dần từ trái sang trong
// [T0, T1] giây (sóng âm trích đoạn, thanh tiến độ cả cuốn); ngoài khoảng đó ẩn.
type BookVideoOverlay struct {
	PNG string  `json:"png"`
	X   int     `json:"x"`
	Y   int     `json:"y"`
	W   int     `json:"w"`
	H   int     `json:"h"`
	T0  float64 `json:"t0"`
	T1  float64 `json:"t1"`
}

// BookVideoPlan — phần còn lại sau khi đã gửi đủ khung hình.
type BookVideoPlan struct {
	Frames   []float64          `json:"frames"` // số giây hiện từng khung hình (đúng thứ tự đã gửi)
	Segs     []BookVideoSeg     `json:"segs"`
	Overlays []BookVideoOverlay `json:"overlays"`
	Width    int                `json:"width"`
	Height   int                `json:"height"`
	Title    string             `json:"title"`
	Thumb    string             `json:"thumb"` // PNG base64; trống = không ghi
	SRT      string             `json:"srt"`
	Desc     string             `json:"desc"`
}

// BookVideoStatus — trạng thái lượt tạo video (để mở lại hộp thoại giữa chừng).
type BookVideoStatus struct {
	Running bool    `json:"running"`
	Done    bool    `json:"done"`
	Error   string  `json:"error"`
	Slug    string  `json:"slug"`
	Title   string  `json:"title"`
	Phase   string  `json:"phase"` // frames | audio | video | files
	Pct     int     `json:"pct"`
	Dir     string  `json:"dir"`
	Video   string  `json:"video"`
	Bytes   int64   `json:"bytes"`
	DurSec  float64 `json:"durSec"`
	Started int64   `json:"started"`
}

type bookVideoJob struct {
	id     string
	slug   string
	dir    string // thư mục của cuốn sách
	work   string
	frames int // số khung hình đã nhận
	ext    map[int]string
	cancel context.CancelFunc
	status BookVideoStatus
}

var bookVideo struct {
	mu  sync.Mutex
	job *bookVideoJob
}

func (a *App) emitBookVideo(st BookVideoStatus) {
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, eventBookVideoProgress, st)
	}
}

// BookVideoBegin mở lượt tạo video cho một cuốn; trả mã lượt để gửi khung hình.
func (a *App) BookVideoBegin(slug, title string) (string, error) {
	dir, err := a.lib.Dir(slug)
	if err != nil {
		return "", err
	}
	bookVideo.mu.Lock()
	defer bookVideo.mu.Unlock()
	// Đang ghép / mã hoá (có cancel) thì chặn; còn đang ở bước vẽ khung hình (giao diện lo,
	// vd app tải lại giữa chừng) thì lượt mới thay lượt cũ.
	if j := bookVideo.job; j != nil && j.status.Running && j.cancel != nil {
		return "", errors.New("đang tạo video một cuốn khác — đợi xong hoặc huỷ trước")
	}
	if j := bookVideo.job; j != nil && j.work != "" {
		_ = os.RemoveAll(j.work)
	}
	root := filepath.Join(os.TempDir(), "sano-bookvideo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	work, err := os.MkdirTemp(root, "job-")
	if err != nil {
		return "", err
	}
	id := filepath.Base(work)
	bookVideo.job = &bookVideoJob{id: id, slug: slug, dir: dir, work: work, ext: map[int]string{}, status: BookVideoStatus{
		Running: true, Slug: slug, Title: title, Phase: "frames", Started: time.Now().Unix(),
	}}
	return id, nil
}

func (a *App) bookVideoJob(id string) (*bookVideoJob, error) {
	j := bookVideo.job
	if j == nil || j.id != id || !j.status.Running {
		return nil, errors.New("lượt tạo video đã kết thúc hoặc bị huỷ")
	}
	return j, nil
}

// decodeFrame nhận ảnh JPEG hoặc PNG (base64, có thể kèm tiền tố data:).
func decodeFrame(b64 string) ([]byte, string, error) {
	if i := strings.Index(b64, ","); i >= 0 && strings.HasPrefix(b64, "data:") {
		b64 = b64[i+1:]
	}
	if base64.StdEncoding.DecodedLen(len(b64)) > maxFrameImage {
		return nil, "", errors.New("ảnh quá lớn")
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, "", fmt.Errorf("ảnh không đọc được: %w", err)
	}
	switch {
	case len(data) > 3 && data[0] == 0xFF && data[1] == 0xD8:
		return data, "jpg", nil
	case len(data) > 8 && string(data[1:4]) == "PNG":
		return data, "png", nil
	}
	return nil, "", errors.New("ảnh phải là JPEG hoặc PNG")
}

// BookVideoFrame ghi khung hình thứ index (0, 1, 2… liên tiếp).
func (a *App) BookVideoFrame(id string, index int, img string) error {
	data, ext, err := decodeFrame(img)
	if err != nil {
		return err
	}
	bookVideo.mu.Lock()
	defer bookVideo.mu.Unlock()
	j, err := a.bookVideoJob(id)
	if err != nil {
		return err
	}
	if index != j.frames || index >= maxBookVideoFrames {
		return fmt.Errorf("khung hình %d không đúng thứ tự (đang chờ %d)", index, j.frames)
	}
	if err := os.WriteFile(filepath.Join(j.work, fmt.Sprintf("f%05d.%s", index, ext)), data, 0o644); err != nil {
		return err
	}
	j.ext[index] = ext
	j.frames++
	return nil
}

func validPlan(p *BookVideoPlan, frames int) (float64, error) {
	if len(p.Frames) != frames || frames == 0 {
		return 0, fmt.Errorf("số khung hình không khớp (%d/%d)", len(p.Frames), frames)
	}
	if len(p.Segs) == 0 || len(p.Segs) > maxBookVideoSegs {
		return 0, errors.New("dòng thời gian tiếng không hợp lệ")
	}
	if len(p.Overlays) > maxBookVideoOverlays {
		return 0, errors.New("quá nhiều lớp phủ")
	}
	if p.Width < 320 || p.Height < 320 || p.Width > 3840 || p.Height > 3840 || p.Width%2 != 0 || p.Height%2 != 0 {
		return 0, errors.New("cỡ khung hình không hợp lệ")
	}
	var vs, as float64
	for _, s := range p.Frames {
		if math.IsNaN(s) || s <= 0 {
			return 0, errors.New("thời lượng khung hình không hợp lệ")
		}
		vs += s
	}
	for _, s := range p.Segs {
		if math.IsNaN(s.Dur) || math.IsNaN(s.Start) || s.Dur <= 0 || s.Start < 0 {
			return 0, errors.New("đoạn tiếng không hợp lệ")
		}
		as += s.Dur
	}
	if as > maxBookVideoSec {
		return 0, errors.New("video quá dài")
	}
	if math.Abs(vs-as) > 0.5 {
		return 0, fmt.Errorf("hình (%.1f giây) và tiếng (%.1f giây) lệch nhau", vs, as)
	}
	for _, o := range p.Overlays {
		if o.W < 2 || o.H < 2 || o.X < 0 || o.Y < 0 || o.X+o.W > p.Width || o.Y+o.H > p.Height || o.T1 <= o.T0 {
			return 0, errors.New("lớp phủ không hợp lệ")
		}
	}
	return as, nil
}

// BookVideoFinish kiểm dòng thời gian rồi chạy ghép + mã hoá ở nền.
func (a *App) BookVideoFinish(id string, plan BookVideoPlan) error {
	ffmpeg := findFFmpeg()
	if ffmpeg == "" {
		return fmt.Errorf("không tìm thấy ffmpeg. %s", tts.FFmpegHint(runtime.GOOS))
	}
	bookVideo.mu.Lock()
	j, err := a.bookVideoJob(id)
	if err != nil {
		bookVideo.mu.Unlock()
		return err
	}
	dur, err := validPlan(&plan, j.frames)
	if err != nil {
		bookVideo.mu.Unlock()
		return err
	}
	for i := range plan.Segs {
		s := &plan.Segs[i]
		if s.Silence {
			continue
		}
		name := filepath.Base(filepath.FromSlash(s.File))
		if !strings.EqualFold(filepath.Ext(name), ".mp3") || !fileExists(filepath.Join(j.dir, name)) {
			bookVideo.mu.Unlock()
			return fmt.Errorf("không thấy file tiếng %s", name)
		}
		s.File = filepath.Join(j.dir, name)
	}
	ctx, cancel := context.WithCancel(a.context())
	j.cancel = cancel
	j.status.Phase, j.status.Pct, j.status.DurSec = "audio", 0, dur
	st := j.status
	bookVideo.mu.Unlock()
	a.emitBookVideo(st)
	go a.runBookVideo(ctx, j, plan, ffmpeg, dur)
	return nil
}

func (a *App) setBookVideo(j *bookVideoJob, f func(*BookVideoStatus)) {
	bookVideo.mu.Lock()
	f(&j.status)
	st := j.status
	bookVideo.mu.Unlock()
	a.emitBookVideo(st)
}

func (a *App) runBookVideo(ctx context.Context, j *bookVideoJob, p BookVideoPlan, ffmpeg string, dur float64) {
	err := a.buildBookVideo(ctx, j, p, ffmpeg, dur)
	if ctx.Err() != nil {
		err = errors.New("đã huỷ")
	}
	_ = os.RemoveAll(j.work)
	a.setBookVideo(j, func(s *BookVideoStatus) {
		s.Running = false
		if err != nil {
			s.Error = err.Error()
		} else {
			s.Done, s.Pct = true, 100
		}
	})
	bookVideo.mu.Lock()
	st := j.status
	bookVideo.mu.Unlock()
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, eventBookVideoFinished, st)
	}
}

func (a *App) buildBookVideo(ctx context.Context, j *bookVideoJob, p BookVideoPlan, ffmpeg string, dur float64) error {
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }
	// 1. Tiếng: mỗi đoạn → WAV đúng Dur giây (cắt thừa, đệm lặng nếu mp3 ngắn hơn), rồi nối.
	var list strings.Builder
	for i, s := range p.Segs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		out := filepath.Join(j.work, fmt.Sprintf("a%05d.wav", i))
		var args []string
		if s.Silence {
			args = []string{"-f", "lavfi", "-t", f(s.Dur), "-i", "anullsrc=r=44100:cl=mono"}
		} else {
			args = []string{"-ss", f(s.Start), "-t", f(s.Dur), "-i", s.File}
		}
		args = append([]string{"-hide_banner", "-loglevel", "error", "-y"}, args...)
		args = append(args, "-af", "aresample=44100,aformat=channel_layouts=mono,apad,atrim=0:"+f(s.Dur), "-c:a", "pcm_s16le", out)
		if err := runQuiet(ctx, ffmpeg, j.work, args); err != nil {
			return fmt.Errorf("ghép tiếng: %w", err)
		}
		fmt.Fprintf(&list, "file '%s'\n", filepath.Base(out))
		a.setBookVideo(j, func(st *BookVideoStatus) { st.Pct = (i + 1) * 100 / len(p.Segs) })
	}
	alist := filepath.Join(j.work, "audio.txt")
	if err := os.WriteFile(alist, []byte(list.String()), 0o644); err != nil {
		return err
	}
	audio := filepath.Join(j.work, "audio.m4a")
	if err := runQuiet(ctx, ffmpeg, j.work, []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", alist,
		"-c:a", "aac", "-b:a", "128k", audio}); err != nil {
		return fmt.Errorf("ghép tiếng: %w", err)
	}

	// 2. Hình: nối khung hình theo số giây, phủ lớp sáng dần, mã hoá cùng tiếng.
	a.setBookVideo(j, func(st *BookVideoStatus) { st.Phase, st.Pct = "video", 0 })
	var vl strings.Builder
	last := ""
	for i, sec := range p.Frames {
		name := fmt.Sprintf("f%05d.%s", i, j.ext[i])
		fmt.Fprintf(&vl, "file '%s'\nduration %.3f\n", name, sec)
		last = name
	}
	fmt.Fprintf(&vl, "file '%s'\n", last)
	vlist := filepath.Join(j.work, "video.txt")
	if err := os.WriteFile(vlist, []byte(vl.String()), 0o644); err != nil {
		return err
	}
	var ovs []string
	for i, o := range p.Overlays {
		data, err := decodePNG(o.PNG)
		if err != nil {
			return err
		}
		path := filepath.Join(j.work, fmt.Sprintf("ov%d.png", i))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
		ovs = append(ovs, path)
	}
	mp4 := filepath.Join(j.work, "out.mp4")
	var lastErr error
	for _, enc := range videoEncoders(ctx, ffmpeg) {
		args := bookVideoArgs(vlist, audio, mp4, ovs, p, dur, enc)
		if lastErr = a.runBookVideoEncode(ctx, j, ffmpeg, args, dur); lastErr == nil || ctx.Err() != nil {
			break
		}
	}
	if lastErr != nil {
		return lastErr
	}

	// 3. Ghi vào Tải về/Sano video/<Tên sách>.
	a.setBookVideo(j, func(st *BookVideoStatus) { st.Phase, st.Pct = "files", 100 })
	dest, err := bookVideoDir(p.Title)
	if err != nil {
		return err
	}
	video := filepath.Join(dest, shareFileName(p.Title, "mp4"))
	_ = os.Remove(video)
	if err := moveFile(mp4, video); err != nil {
		return err
	}
	if p.Thumb != "" {
		if data, err := decodePNG(p.Thumb); err == nil {
			_ = os.WriteFile(filepath.Join(dest, "thumbnail.png"), data, 0o644)
		}
	}
	if p.SRT != "" {
		_ = os.WriteFile(filepath.Join(dest, "phu-de.srt"), []byte(p.SRT), 0o644)
	}
	if p.Desc != "" {
		_ = os.WriteFile(filepath.Join(dest, "mo-ta-youtube.txt"), []byte(p.Desc), 0o644)
	}
	info, _ := os.Stat(video)
	a.setBookVideo(j, func(st *BookVideoStatus) {
		st.Dir, st.Video = dest, video
		if info != nil {
			st.Bytes = info.Size()
		}
	})
	return nil
}

// bookVideoDir — Tải về/Sano video/<Tên sách> (tạo nếu chưa có).
func bookVideoDir(title string) (string, error) {
	base := downloadsDir()
	if base == "" {
		return "", errors.New("không tìm thấy thư mục Tải về")
	}
	name := strings.TrimSuffix(shareFileName(title, "x"), " - Sano.x")
	d := filepath.Join(base, "Sano video", name)
	return d, os.MkdirAll(d, 0o755)
}

func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src) // khác ổ đĩa: chép rồi xoá
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}

func bookVideoArgs(vlist, audio, out string, ovs []string, p BookVideoPlan, dur float64, enc string) []string {
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }
	fps := strconv.Itoa(bookVideoFPS)
	args := []string{"-hide_banner", "-loglevel", "error", "-nostats", "-progress", "pipe:1", "-y",
		"-f", "concat", "-safe", "0", "-i", vlist, "-i", audio}
	for _, o := range ovs {
		args = append(args, "-loop", "1", "-framerate", fps, "-t", f(dur), "-i", o)
	}
	var fc strings.Builder
	fmt.Fprintf(&fc, "[0:v]fps=%s,scale=%d:%d,format=rgba[v0]", fps, p.Width, p.Height)
	cur := "v0"
	for i, o := range p.Overlays {
		// Mặt nạ trắng trượt từ trái sang trong [T0, T1] → phần sáng hiện dần; ngoài khoảng ẩn.
		span := f(o.T1 - o.T0)
		fmt.Fprintf(&fc, ";color=c=black:s=%[1]dx%[2]d:r=%[3]s:d=%[4]s,format=gray[m0_%[5]d];color=c=white:s=%[1]dx%[2]d:r=%[3]s:d=%[4]s,format=gray[m1_%[5]d]"+
			";[m0_%[5]d][m1_%[5]d]overlay=x='-w+w*clip((t-%[6]s)/%[7]s,0,1)':y=0:eval=frame[mk%[5]d]"+
			";[%[8]d:v]fps=%[3]s,format=rgba,scale=%[1]d:%[2]d,split[b%[5]d][c%[5]d];[c%[5]d]alphaextract[ba%[5]d];[ba%[5]d][mk%[5]d]blend=all_mode=multiply[am%[5]d];[b%[5]d][am%[5]d]alphamerge[o%[5]d]"+
			";[%[9]s][o%[5]d]overlay=%[10]d:%[11]d:enable='between(t,%[6]s,%[12]s)':eof_action=pass[v%[13]d]",
			o.W, o.H, fps, f(dur), i, f(o.T0), span, i+2, cur, o.X, o.Y, f(o.T1), i+1)
		cur = fmt.Sprintf("v%d", i+1)
	}
	fmt.Fprintf(&fc, ";[%s]format=yuv420p[vout]", cur)
	args = append(args, "-filter_complex", fc.String(), "-map", "[vout]", "-map", "1:a", "-c:v", enc)
	switch enc {
	case "libx264":
		args = append(args, "-preset", "veryfast", "-crf", "23", "-tune", "stillimage")
	case "mpeg4":
		args = append(args, "-q:v", "4")
	default:
		args = append(args, "-b:v", "4M")
	}
	return append(args, "-r", fps, "-c:a", "copy", "-t", f(dur), "-movflags", "+faststart", out)
}

func (a *App) runBookVideoEncode(ctx context.Context, j *bookVideoJob, ffmpeg string, args []string, dur float64) error {
	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	cmd.Dir = j.work
	tts.HideWindow(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("chạy ffmpeg: %w", err)
	}
	sc := bufio.NewScanner(stdout)
	lastPct := -1
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if !ok || k != "out_time_us" {
			continue
		}
		us, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		pct := int(math.Min(99, math.Max(0, us/1e6/dur*100)))
		if pct != lastPct {
			lastPct = pct
			a.setBookVideo(j, func(st *BookVideoStatus) { st.Pct = pct })
		}
	}
	if err := cmd.Wait(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if i := strings.LastIndex(msg, "\n"); i >= 0 {
			msg = msg[i+1:]
		}
		return fmt.Errorf("mã hoá video lỗi: %s", firstNonEmpty(msg, err.Error()))
	}
	return nil
}

func runQuiet(ctx context.Context, bin, dir string, args []string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	tts.HideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if i := strings.LastIndex(msg, "\n"); i >= 0 {
			msg = msg[i+1:]
		}
		return errors.New(firstNonEmpty(msg, err.Error()))
	}
	return nil
}

// BookVideoCancel huỷ lượt đang chạy (cả lúc giao diện còn đang vẽ khung hình).
func (a *App) BookVideoCancel() {
	bookVideo.mu.Lock()
	j := bookVideo.job
	if j == nil || !j.status.Running {
		bookVideo.mu.Unlock()
		return
	}
	if j.cancel != nil {
		j.cancel() // runBookVideo tự dọn và báo xong
		bookVideo.mu.Unlock()
		return
	}
	j.status.Running, j.status.Error = false, "đã huỷ"
	_ = os.RemoveAll(j.work)
	st := j.status
	bookVideo.mu.Unlock()
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, eventBookVideoFinished, st)
	}
}

// BookVideoState — lượt gần nhất (nil = chưa có).
func (a *App) BookVideoState() *BookVideoStatus {
	bookVideo.mu.Lock()
	defer bookVideo.mu.Unlock()
	if bookVideo.job == nil {
		return nil
	}
	st := bookVideo.job.status
	return &st
}

// OpenBookVideoFolder mở thư mục chứa video vừa tạo (chọn sẵn file video).
func (a *App) OpenBookVideoFolder() error {
	bookVideo.mu.Lock()
	var video string
	if bookVideo.job != nil {
		video = bookVideo.job.status.Video
	}
	bookVideo.mu.Unlock()
	if video == "" || !fileExists(video) {
		return errors.New("chưa có video nào vừa tạo")
	}
	return openPath(video, true)
}
