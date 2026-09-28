package main

// Chia sẻ đoạn hay (wireframe D14): giao diện vẽ thẻ bằng canvas rồi gửi PNG
// (base64) sang đây để lưu, chép vào clipboard, AirDrop; video = các ảnh thẻ
// từng câu nối lại + đúng đoạn tiếng đọc của tiểu mục; sóng âm (giao diện tính từ
// tiếng đọc thật, vẽ sẵn cột mờ trên thẻ + ảnh cột sáng riêng) được ffmpeg tô sáng
// dần từ trái sang theo thời gian, như tin nhắn thoại. File tạm nằm trong thư mục tạm của app; giao diện
// không gửi đường dẫn file nào ngoài tên file mp3 của tiểu mục (kiểm lại ở đây).

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
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/desktop/internal/tts"
)

const (
	eventShareProgress = "share:progress"
	maxShareVideoSec   = 60 // giao diện giới hạn ~30 giây; chặn thêm ở đây
	maxShareFrames     = 8
	maxSharePNG        = 12 << 20 // một ảnh thẻ PNG sau khi giải base64
)

// ShareFrame — một ảnh thẻ của video và số giây hiện ảnh đó.
type ShareFrame struct {
	PNG string  `json:"png"` // base64, không kèm tiền tố data:
	Sec float64 `json:"sec"`
}

// ShareVideoRequest — video từ đoạn [Start, End) giây của một tiểu mục.
type ShareVideoRequest struct {
	Slug   string       `json:"slug"`
	File   string       `json:"file"` // file mp3 của tiểu mục (Track.File; chỉ dùng tên file)
	Start  float64      `json:"start"`
	End    float64      `json:"end"`
	Frames []ShareFrame `json:"frames"`
	// Ảnh cột sóng âm sáng (base64 PNG, cỡ WaveW×WaveH) đặt ở (WaveX, WaveY) trên thẻ;
	// trống thì không tô.
	WavePNG string `json:"wavePNG"`
	WaveX   int    `json:"waveX"`
	WaveY   int    `json:"waveY"`
	WaveW   int    `json:"waveW"`
	WaveH   int    `json:"waveH"`
	Title   string `json:"title"` // tên sách, để đặt tên file khi lưu
}

type shareState struct {
	mu        sync.Mutex
	video     string // file mp4 vừa tạo (trong thư mục tạm)
	videoName string
	cancel    context.CancelFunc
}

var share shareState

func shareTempDir() (string, error) {
	d := filepath.Join(os.TempDir(), "sano-share")
	return d, os.MkdirAll(d, 0o755)
}

func decodePNG(b64 string) ([]byte, error) {
	b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	if base64.StdEncoding.DecodedLen(len(b64)) > maxSharePNG {
		return nil, errors.New("ảnh quá lớn")
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("ảnh không đọc được: %w", err)
	}
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		return nil, errors.New("không phải ảnh PNG")
	}
	return data, nil
}

var badName = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]+`)

// shareFileName — "<Tên sách> - Sano.<ext>", bỏ ký tự hệ điều hành cấm.
func shareFileName(title, ext string) string {
	t := strings.TrimSpace(badName.ReplaceAllString(title, " "))
	if r := []rune(t); len(r) > 80 {
		t = strings.TrimSpace(string(r[:80]))
	}
	if t == "" {
		t = "Sách nói"
	}
	return t + " - Sano." + ext
}

// saveAs hỏi nơi lưu (mặc định Tải về). Huỷ → "".
func (a *App) saveAs(title, name, ext, label string) (string, error) {
	if a.ctx == nil {
		return "", errors.New("ứng dụng chưa khởi động xong")
	}
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:                title,
		DefaultDirectory:     downloadsDir(),
		DefaultFilename:      name,
		Filters:              []wruntime.FileFilter{{DisplayName: label, Pattern: "*." + ext}},
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", fmt.Errorf("mở hộp lưu file: %w", err)
	}
	if path != "" && !strings.EqualFold(filepath.Ext(path), "."+ext) {
		path += "." + ext
	}
	return path, nil
}

// SaveShareImage lưu ảnh thẻ; trả đường dẫn đã lưu ("" nếu người dùng huỷ).
func (a *App) SaveShareImage(png, title string) (string, error) {
	data, err := decodePNG(png)
	if err != nil {
		return "", err
	}
	path, err := a.saveAs("Lưu ảnh chia sẻ", shareFileName(title, "png"), "png", "Ảnh PNG (*.png)")
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("lưu ảnh: %w", err)
	}
	return path, nil
}

// CopyShareImage chép ảnh thẻ vào clipboard (dán thẳng vào Facebook, Zalo).
// Chỉ có trên Mac; máy khác giao diện tự chép bằng API của trình duyệt.
func (a *App) CopyShareImage(png string) error {
	data, err := decodePNG(png)
	if err != nil {
		return err
	}
	return copyImage(data)
}

// CanCopyImage — máy này chép ảnh qua phần Go được không.
func (a *App) CanCopyImage() bool { return copyImageSupported }

// AirDropShareImage ghi ảnh ra file tạm rồi mở bảng AirDrop.
func (a *App) AirDropShareImage(png, title string) error {
	data, err := decodePNG(png)
	if err != nil {
		return err
	}
	dir, err := shareTempDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, shareFileName(title, "png"))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return airDrop(path)
}

// MakeShareVideo tạo video MP4 (chạy xong mới trả về; tiến độ qua sự kiện
// share:progress 0–100). Trả thời lượng video (giây).
func (a *App) MakeShareVideo(req ShareVideoRequest) (float64, error) {
	dur := req.End - req.Start
	if math.IsNaN(dur) || req.Start < 0 || dur <= 0.2 || dur > maxShareVideoSec {
		return 0, fmt.Errorf("đoạn video phải dài từ 0,2 đến %d giây", maxShareVideoSec)
	}
	if len(req.Frames) == 0 || len(req.Frames) > maxShareFrames {
		return 0, fmt.Errorf("cần 1–%d ảnh thẻ", maxShareFrames)
	}
	dir, err := a.lib.Dir(req.Slug)
	if err != nil {
		return 0, err
	}
	// Track.File tính từ gốc thư viện (Sach/<cuốn>/x.mp3): chỉ lấy tên file và luôn
	// tìm trong thư mục của đúng cuốn này, không theo đường dẫn giao diện gửi.
	name := filepath.Base(filepath.FromSlash(req.File))
	if name == "." || name == string(filepath.Separator) || !strings.EqualFold(filepath.Ext(name), ".mp3") {
		return 0, errors.New("file tiếng không hợp lệ")
	}
	audio := filepath.Join(dir, name)
	if !fileExists(audio) {
		return 0, fmt.Errorf("không thấy file tiếng %s", name)
	}
	ffmpeg := findFFmpeg()
	if ffmpeg == "" {
		return 0, fmt.Errorf("không tìm thấy ffmpeg. %s", tts.FFmpegHint(runtime.GOOS))
	}

	tmp, err := shareTempDir()
	if err != nil {
		return 0, err
	}
	work, err := os.MkdirTemp(tmp, "video-")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(work)

	// Danh sách ảnh cho concat demuxer: ảnh cuối lặp lại (quy ước của ffmpeg).
	var list strings.Builder
	total := 0.0
	last := ""
	for i, f := range req.Frames {
		data, err := decodePNG(f.PNG)
		if err != nil {
			return 0, err
		}
		p := filepath.Join(work, fmt.Sprintf("f%02d.png", i))
		if err := os.WriteFile(p, data, 0o644); err != nil {
			return 0, err
		}
		sec := f.Sec
		if i == len(req.Frames)-1 {
			sec = dur - total // câu cuối kéo tới hết đoạn
		}
		if math.IsNaN(sec) || sec < 0.05 {
			sec = 0.05
		}
		total += sec
		fmt.Fprintf(&list, "file '%s'\nduration %.3f\n", filepath.Base(p), sec)
		last = filepath.Base(p)
	}
	fmt.Fprintf(&list, "file '%s'\n", last)
	listPath := filepath.Join(work, "list.txt")
	if err := os.WriteFile(listPath, []byte(list.String()), 0o644); err != nil {
		return 0, err
	}
	wave := ""
	if req.WavePNG != "" && req.WaveW > 1 && req.WaveH > 1 {
		data, err := decodePNG(req.WavePNG)
		if err != nil {
			return 0, err
		}
		wave = filepath.Join(work, "wave.png")
		if err := os.WriteFile(wave, data, 0o644); err != nil {
			return 0, err
		}
	}
	out := filepath.Join(work, "out.mp4")

	ctx, cancel := context.WithCancel(a.context())
	share.mu.Lock()
	if share.cancel != nil {
		share.mu.Unlock()
		cancel()
		return 0, errors.New("đang tạo một video khác")
	}
	share.cancel = cancel
	share.mu.Unlock()
	defer func() {
		cancel()
		share.mu.Lock()
		share.cancel = nil
		share.mu.Unlock()
	}()

	var lastErr error
	for _, enc := range videoEncoders(ctx, ffmpeg) {
		args := shareVideoArgs(listPath, audio, wave, out, req, dur, enc)
		if lastErr = a.runShareFFmpeg(ctx, ffmpeg, work, args, dur); lastErr == nil {
			break
		}
		if ctx.Err() != nil {
			return 0, errors.New("đã huỷ")
		}
	}
	if lastErr != nil {
		return 0, lastErr
	}

	final := filepath.Join(tmp, fmt.Sprintf("video-%d.mp4", time.Now().UnixNano()))
	if err := os.Rename(out, final); err != nil {
		return 0, err
	}
	share.mu.Lock()
	if share.video != "" {
		_ = os.Remove(share.video)
	}
	share.video = final
	share.videoName = shareFileName(req.Title, "mp4")
	share.mu.Unlock()
	return dur, nil
}

// CancelShareVideo huỷ lượt tạo video đang chạy (đóng hộp thoại).
func (a *App) CancelShareVideo() {
	share.mu.Lock()
	defer share.mu.Unlock()
	if share.cancel != nil {
		share.cancel()
	}
}

// videoEncoders — bộ mã H.264 theo thứ tự ưu tiên có trong ffmpeg này.
func videoEncoders(ctx context.Context, ffmpeg string) []string {
	cmd := exec.CommandContext(ctx, ffmpeg, "-hide_banner", "-encoders")
	tts.HideWindow(cmd)
	b, _ := cmd.Output()
	s := string(b)
	var out []string
	for _, e := range []string{"libx264", "h264_videotoolbox", "h264_mf"} {
		if strings.Contains(s, " "+e+" ") {
			out = append(out, e)
		}
	}
	return append(out, "mpeg4") // bộ mã có sẵn trong mọi bản ffmpeg
}

func shareVideoArgs(list, audio, wave, out string, req ShareVideoRequest, dur float64, enc string) []string {
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }
	args := []string{"-hide_banner", "-loglevel", "error", "-nostats", "-progress", "pipe:1", "-y",
		"-f", "concat", "-safe", "0", "-i", list,
		"-ss", f(req.Start), "-t", f(dur), "-i", audio}
	filter := "[0:v]fps=30,format=yuv420p[v]"
	if wave != "" {
		// Mặt nạ trắng trượt từ trái sang (hết đoạn thì phủ kín) → cột sáng hiện dần.
		w, h := req.WaveW, req.WaveH
		args = append(args, "-loop", "1", "-framerate", "30", "-t", f(dur), "-i", wave)
		filter = fmt.Sprintf("[0:v]fps=30,format=rgba[bg];"+
			"color=c=black:s=%[1]dx%[2]d:r=30:d=%[3]s,format=gray[m0];color=c=white:s=%[1]dx%[2]d:r=30:d=%[3]s,format=gray[m1];"+
			"[m0][m1]overlay=x='-w+w*t/%[3]s':y=0:eval=frame[mask];"+
			"[2:v]fps=30,format=rgba,scale=%[1]d:%[2]d,split[b][b2];[b2]alphaextract[ba];[ba][mask]blend=all_mode=multiply[am];[b][am]alphamerge[bm];"+
			"[bg][bm]overlay=%[4]d:%[5]d:shortest=1,format=yuv420p[v]", w, h, f(dur), req.WaveX, req.WaveY)
	}
	args = append(args, "-filter_complex", filter, "-map", "[v]", "-map", "1:a", "-c:v", enc)
	switch enc {
	case "libx264":
		args = append(args, "-preset", "medium", "-crf", "20", "-tune", "stillimage")
	case "mpeg4":
		args = append(args, "-q:v", "3")
	default:
		args = append(args, "-b:v", "6M")
	}
	return append(args, "-r", "30", "-c:a", "aac", "-b:a", "160k", "-t", f(dur), "-movflags", "+faststart", out)
}

// runShareFFmpeg chạy ffmpeg, đọc -progress để báo phần trăm.
func (a *App) runShareFFmpeg(ctx context.Context, ffmpeg, dir string, args []string, dur float64) error {
	cmd := exec.CommandContext(ctx, ffmpeg, args...)
	cmd.Dir = dir
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
	for sc.Scan() {
		k, v, ok := strings.Cut(sc.Text(), "=")
		if !ok || k != "out_time_us" {
			continue
		}
		us, err := strconv.ParseFloat(v, 64)
		if err != nil || a.ctx == nil {
			continue
		}
		pct := math.Min(99, math.Max(0, us/1e6/dur*100))
		wruntime.EventsEmit(a.ctx, eventShareProgress, int(pct))
	}
	if err := cmd.Wait(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if i := strings.LastIndex(msg, "\n"); i >= 0 {
			msg = msg[i+1:]
		}
		return fmt.Errorf("tạo video lỗi: %s", firstNonEmpty(msg, err.Error()))
	}
	return nil
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func currentShareVideo() (string, string, error) {
	share.mu.Lock()
	defer share.mu.Unlock()
	if share.video == "" || !fileExists(share.video) {
		return "", "", errors.New("chưa có video nào vừa tạo")
	}
	return share.video, share.videoName, nil
}

// SaveShareVideo chép video vừa tạo tới nơi người dùng chọn ("" nếu huỷ).
func (a *App) SaveShareVideo() (string, error) {
	src, name, err := currentShareVideo()
	if err != nil {
		return "", err
	}
	path, err := a.saveAs("Lưu video chia sẻ", name, "mp4", "Video MP4 (*.mp4)")
	if err != nil || path == "" {
		return "", err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("lưu video: %w", err)
	}
	return path, nil
}

// AirDropShareVideo mở bảng AirDrop cho video vừa tạo (đặt lại tên dễ hiểu).
func (a *App) AirDropShareVideo() error {
	src, name, err := currentShareVideo()
	if err != nil {
		return err
	}
	dst := filepath.Join(filepath.Dir(src), name)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	return airDrop(dst)
}
