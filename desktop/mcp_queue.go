package main

// Sách AI tạo qua MCP (wireframe D22): TẠO TRƯỚC, CAM KẾT SAU.
//   - start_render xếp bản nháp vào hàng đợi; Sano render lần lượt từng cuốn khi máy rảnh (không
//     chen vào lượt tạo sách / sửa sách người dùng đang chạy).
//   - Render xong: sách vào KHU CHỜ ~/Sano/.cho-cam-ket/<id>/sach (chưa vào Thư viện, chưa nghe /
//     xuất / chia sẻ được); app mở popup cam kết.
//   - Người dùng tick cuốn muốn lưu + 6 ô cam kết D19 một lần (MCPCommit): cuốn đã xong vào Thư viện
//     ngay, kèm thời điểm cam kết trong metadata.json; cuốn đang tạo / chờ tạo tự lưu khi xong.
//   - Không cam kết: khu chờ giữ 7 ngày rồi tự xoá. "Bỏ" (MCPDiscard) bỏ hẳn một cuốn.
// Luồng Tạo sách nói trong app giữ nguyên (cam kết trước khi render).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	eventMCPJobs = "mcp:jobs" // danh sách sách AI nhờ tạo đổi → giao diện tải lại (popup, thẻ thanh bên)

	mcpStageDir = ".cho-cam-ket" // trong thư mục gốc ~/Sano
	mcpStageTTL = 7 * 24 * time.Hour
	mcpJobFile  = "job.json"
)

// Trạng thái một cuốn AI nhờ tạo.
const (
	jobQueued    = "queued"    // chờ tạo
	jobRendering = "rendering" // đang tạo
	jobStaged    = "staged"    // tạo xong, ở khu chờ cam kết
	jobSaved     = "saved"     // đã cam kết, đã vào Thư viện
	jobFailed    = "failed"
	jobCancelled = "cancelled"
)

// MCPJob — một cuốn AI nhờ tạo (giao diện + job.json ở khu chờ).
type MCPJob struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Author    string  `json:"author,omitempty"`
	Voice     string  `json:"voice"`
	Client    string  `json:"client"`
	Sections  int     `json:"sections"`
	ListenMin int     `json:"listenMin"`
	Status    string  `json:"status"`
	Percent   float64 `json:"percent"`
	Slug      string  `json:"slug,omitempty"`  // đã lưu: slug trong Thư viện
	Error     string  `json:"error,omitempty"` // lỗi tạo / lưu
	AutoSave  bool    `json:"autoSave"`        // đã cam kết khi còn đang tạo / chờ tạo → xong tự lưu
	PledgedAt string  `json:"pledgedAt,omitempty"`
	CreatedAt int64   `json:"createdAt"`
	StagedAt  int64   `json:"stagedAt,omitempty"`
	SlugBase  string  `json:"slugBase,omitempty"` // slug dự kiến khi vào Thư viện

	settings BookSettings
}

var jobIDRe = regexp.MustCompile(`^[0-9a-f]{16}$`)

func (a *App) mcpStageRoot() string { return filepath.Join(a.lib.Root(), mcpStageDir) }

// jobStagePath — thư mục khu chờ của một cuốn; id luôn là 16 ký tự hex (không thoát ra ngoài được).
func (a *App) jobStagePath(id string) (string, error) {
	if !jobIDRe.MatchString(id) {
		return "", errors.New("mã sách không hợp lệ")
	}
	return filepath.Join(a.mcpStageRoot(), id), nil
}

// ── Công cụ MCP ──────────────────────────────────────────────────────────

type mcpDraftID struct {
	DraftID string `json:"draft_id" jsonschema:"draft_id trả về từ create_book"`
}

type mcpJobOut struct {
	DraftID   string  `json:"draft_id"`
	Title     string  `json:"title"`
	Status    string  `json:"status"` // queued | rendering | waiting_pledge | saved | failed | cancelled
	Percent   float64 `json:"percent,omitempty"`
	RemainMin int     `json:"remain_min,omitempty"`
	Slug      string  `json:"slug,omitempty"`
	Error     string  `json:"error,omitempty"`
}

type mcpRenderOut struct {
	Message string      `json:"message"`
	Books   []mcpJobOut `json:"books"`
}

func (a *App) mcpStartRender(ctx context.Context, _ *mcp.CallToolRequest, in mcpDraftID) (*mcp.CallToolResult, mcpRenderOut, error) {
	id := strings.TrimSpace(in.DraftID)
	a.mcp.mu.Lock()
	d, ok := a.mcp.drafts[id]
	if !ok {
		a.mcp.mu.Unlock()
		return nil, mcpRenderOut{}, errors.New("không thấy bản nháp (quá 2 giờ, sai draft_id hoặc đã xếp hàng rồi): gọi create_book lại")
	}
	delete(a.mcp.drafts, id)
	listen := 0
	for _, ch := range d.outline.Chapters {
		for _, s := range ch.Sections {
			if !s.TOC {
				listen += s.Chars
			}
		}
	}
	listen += len([]rune(d.settings.IntroText))
	j := &MCPJob{ID: id, Title: d.settings.Title, Author: d.settings.Author, Voice: d.settings.Voice, Client: clientFrom(ctx),
		Sections: d.outline.Sections, ListenMin: max(1, listen/15/60), Status: jobQueued, CreatedAt: time.Now().Unix(),
		settings: d.settings}
	a.mcp.jobs = append(a.mcp.jobs, j)
	a.mcp.mu.Unlock()
	a.emitJobs()
	a.pumpMCP()
	out := a.jobsOut(id)
	out.Message = "Đã xếp hàng. Sano tạo sách ngay khi máy rảnh (lần lượt từng cuốn). Tạo xong, sách nằm ở khu chờ; người dùng tick cam kết trên app Sano thì mới vào Thư viện. Theo dõi bằng get_render_status."
	return nil, out, nil
}

type mcpStatusInput struct {
	DraftID string `json:"draft_id,omitempty" jsonschema:"chỉ xem một cuốn; bỏ trống = mọi cuốn AI nhờ tạo trong lần mở app này và các cuốn còn ở khu chờ"`
}

func (a *App) mcpRenderStatus(_ context.Context, _ *mcp.CallToolRequest, in mcpStatusInput) (*mcp.CallToolResult, mcpRenderOut, error) {
	out := a.jobsOut(strings.TrimSpace(in.DraftID))
	switch {
	case len(out.Books) == 0:
		out.Message = "Chưa có cuốn nào AI nhờ tạo."
	default:
		n := map[string]int{}
		for _, b := range out.Books {
			n[b.Status]++
		}
		var parts []string
		for _, k := range []string{"rendering", "queued", "waiting_pledge", "saved", "failed", "cancelled"} {
			if n[k] > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", n[k], map[string]string{
					"rendering": "đang tạo", "queued": "chờ tạo", "waiting_pledge": "đã tạo xong, chờ người dùng cam kết trên app",
					"saved": "đã lưu vào Thư viện", "failed": "lỗi", "cancelled": "đã dừng"}[k]))
			}
		}
		out.Message = strings.Join(parts, "; ") + "."
	}
	return nil, out, nil
}

func (a *App) mcpCancelRender(_ context.Context, _ *mcp.CallToolRequest, in mcpDraftID) (*mcp.CallToolResult, mcpRenderOut, error) {
	id := strings.TrimSpace(in.DraftID)
	a.mcp.mu.Lock()
	j := a.findJobLocked(id)
	if j == nil || (j.Status != jobQueued && j.Status != jobRendering) {
		a.mcp.mu.Unlock()
		return nil, mcpRenderOut{}, errors.New("cuốn này không ở trạng thái chờ tạo / đang tạo")
	}
	rendering := j.Status == jobRendering
	if !rendering {
		j.Status = jobCancelled
	}
	a.mcp.mu.Unlock()
	if rendering {
		a.CancelRender() // afterRender đánh dấu cancelled rồi tạo cuốn kế
	}
	a.emitJobs()
	out := a.jobsOut(id)
	out.Message = "Đã dừng, không lưu gì."
	return nil, out, nil
}

// jobsOut — tình trạng các cuốn cho AI (id rỗng = tất cả).
func (a *App) jobsOut(id string) mcpRenderOut {
	out := mcpRenderOut{Books: []mcpJobOut{}}
	for _, j := range a.MCPJobs() {
		if id != "" && j.ID != id {
			continue
		}
		st := j.Status
		if st == jobStaged {
			st = "waiting_pledge"
		}
		o := mcpJobOut{DraftID: j.ID, Title: j.Title, Status: st, Percent: j.Percent, Slug: j.Slug, Error: j.Error}
		if j.Status == jobRendering {
			if r := a.RenderStatus(); r != nil && r.Running && r.Progress.DoneChars > 0 && r.Progress.ElapsedSec > 0 {
				left := float64(r.Progress.TotalChars-r.Progress.DoneChars) * float64(r.Progress.ElapsedSec) / float64(r.Progress.DoneChars)
				o.RemainMin = max(1, int(left/60+0.5))
			}
		}
		out.Books = append(out.Books, o)
	}
	return out
}

// ── Hàng đợi ─────────────────────────────────────────────────────────────

func (a *App) findJobLocked(id string) *MCPJob {
	for _, j := range a.mcp.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}

// pumpMCP bắt đầu tạo cuốn kế tiếp nếu máy rảnh; bận (người dùng đang tạo / sửa sách) thì hẹn thử lại.
func (a *App) pumpMCP() {
	a.mcp.mu.Lock()
	var next *MCPJob
	for _, j := range a.mcp.jobs {
		if j.Status == jobRendering {
			a.mcp.mu.Unlock()
			return
		}
		if next == nil && j.Status == jobQueued {
			next = j
		}
	}
	if next == nil {
		a.mcp.mu.Unlock()
		return
	}
	a.mu.Lock()
	busy := a.renderBusyLocked()
	a.mu.Unlock()
	if busy != nil {
		if !a.mcp.pumping {
			a.mcp.pumping = true
			go func() {
				time.Sleep(5 * time.Second)
				a.mcp.mu.Lock()
				a.mcp.pumping = false
				a.mcp.mu.Unlock()
				a.pumpMCP()
			}()
		}
		a.mcp.mu.Unlock()
		return
	}
	next.Status = jobRendering
	next.SlugBase = ""
	id, s := next.ID, next.settings
	a.mcp.mu.Unlock()

	_, err := a.startRender(s, "mcp", func(work, slug string) (string, error) { return a.stageJob(id, work, slug) })
	if err != nil {
		a.mcp.mu.Lock()
		if j := a.findJobLocked(id); j != nil {
			j.Status, j.Error = jobFailed, err.Error()
		}
		a.mcp.mu.Unlock()
		a.emitJobs()
		go a.pumpMCP()
		return
	}
	a.emitJobs()
}

// afterRender — một lượt render xong (mọi nguồn): cập nhật cuốn AI vừa tạo (lỗi / dừng), tạo cuốn kế.
func (a *App) afterRender(st RenderStatus) {
	if st.Source == "mcp" {
		a.mcp.mu.Lock()
		for _, j := range a.mcp.jobs {
			if j.Status != jobRendering {
				continue
			}
			switch {
			case st.Cancelled:
				j.Status = jobCancelled
			case st.Error != "":
				j.Status, j.Error = jobFailed, st.Error
			}
		}
		a.mcp.mu.Unlock()
		a.emitJobs()
	}
	a.pumpMCP()
}

// stageJob — finalize của lượt render MCP: chuyển sách vào khu chờ; đã cam kết trước thì lưu luôn.
func (a *App) stageJob(id, work, slug string) (string, error) {
	dir, err := a.jobStagePath(id)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("tạo khu chờ: %w", err)
	}
	if err := os.Rename(work, filepath.Join(dir, "sach")); err != nil {
		return "", fmt.Errorf("chuyển sách vào khu chờ: %w", err)
	}
	a.mcp.mu.Lock()
	j := a.findJobLocked(id)
	if j == nil {
		a.mcp.mu.Unlock()
		_ = os.RemoveAll(dir)
		return "", errors.New("cuốn này đã bị bỏ")
	}
	j.Status, j.Percent, j.SlugBase, j.StagedAt = jobStaged, 100, slug, time.Now().Unix()
	auto := j.AutoSave
	snapshot := *j
	a.mcp.mu.Unlock()
	if b, err := json.Marshal(snapshot); err == nil {
		_ = os.WriteFile(filepath.Join(dir, mcpJobFile), b, 0o600)
	}
	if auto {
		if saved, err := a.commitJob(id); err == nil {
			a.emitJobs()
			return saved, nil
		}
	}
	a.emitJobs()
	return "", nil
}

// commitJob — cuốn ở khu chờ đã được cam kết: ghi thời điểm cam kết, chuyển vào Thư viện.
func (a *App) commitJob(id string) (string, error) {
	a.mcp.mu.Lock()
	j := a.findJobLocked(id)
	if j == nil || j.Status != jobStaged {
		a.mcp.mu.Unlock()
		return "", errors.New("cuốn này không ở khu chờ")
	}
	pledged, base := j.PledgedAt, j.SlugBase
	a.mcp.mu.Unlock()
	if pledged == "" {
		return "", errors.New("chưa cam kết")
	}
	dir, err := a.jobStagePath(id)
	if err != nil {
		return "", err
	}
	book := filepath.Join(dir, "sach")
	if err := setRightsConfirmed(filepath.Join(book, "metadata.json"), pledged); err != nil {
		return "", a.failJob(id, fmt.Errorf("ghi thời điểm cam kết: %w", err))
	}
	if base == "" {
		base = "sach-ai"
	}
	slug, err := a.lib.Commit(book, base)
	if err != nil {
		return "", a.failJob(id, err)
	}
	_ = os.RemoveAll(dir)
	a.mcp.mu.Lock()
	if j := a.findJobLocked(id); j != nil {
		j.Status, j.Slug, j.Error = jobSaved, slug, ""
	}
	a.mcp.mu.Unlock()
	return slug, nil
}

func (a *App) failJob(id string, err error) error {
	a.mcp.mu.Lock()
	if j := a.findJobLocked(id); j != nil {
		j.Error = err.Error() // giữ ở khu chờ, người dùng thử lại được
	}
	a.mcp.mu.Unlock()
	return err
}

// setRightsConfirmed ghi rights_confirmed_at vào metadata.json, giữ nguyên các trường khác.
func setRightsConfirmed(path, at string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	v, _ := json.Marshal(at)
	m["rights_confirmed_at"] = v
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// ── Giao diện ────────────────────────────────────────────────────────────

// MCPJobs — các cuốn AI nhờ tạo (cũ trước), tiến độ cuốn đang tạo lấy từ lượt render.
func (a *App) MCPJobs() []MCPJob {
	a.expireStaged()
	r := a.RenderStatus()
	a.mcp.mu.Lock()
	defer a.mcp.mu.Unlock()
	out := make([]MCPJob, 0, len(a.mcp.jobs))
	for _, j := range a.mcp.jobs {
		c := *j
		if c.Status == jobRendering && r != nil && r.Source == "mcp" && r.Running {
			c.Percent = progressPercent(r.Progress.DoneChars, r.Progress.TotalChars)
		}
		out = append(out, c)
	}
	return out
}

// MCPCommit — người dùng tick các cuốn + đủ cam kết: cuốn đã tạo xong vào Thư viện ngay, cuốn đang
// tạo / chờ tạo tự lưu khi xong. Mỗi cuốn ghi thời điểm cam kết riêng.
func (a *App) MCPCommit(ids []string) ([]MCPJob, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var staged []string
	a.mcp.mu.Lock()
	for _, id := range ids {
		j := a.findJobLocked(id)
		if j == nil {
			continue
		}
		switch j.Status {
		case jobStaged:
			j.PledgedAt = now
			staged = append(staged, id)
		case jobQueued, jobRendering:
			j.PledgedAt, j.AutoSave = now, true
		}
	}
	a.mcp.mu.Unlock()
	var errs []string
	for _, id := range staged {
		if _, err := a.commitJob(id); err != nil {
			errs = append(errs, err.Error())
		}
	}
	a.emitJobs()
	if len(errs) > 0 {
		return a.MCPJobs(), fmt.Errorf("chưa lưu được: %s", strings.Join(errs, "; "))
	}
	return a.MCPJobs(), nil
}

// MCPDiscard — bỏ hẳn một cuốn AI tạo: ở khu chờ thì xoá bản ở khu chờ (không đụng Thư viện),
// chờ tạo thì bỏ khỏi hàng, đang tạo thì dừng.
func (a *App) MCPDiscard(id string) ([]MCPJob, error) {
	a.mcp.mu.Lock()
	j := a.findJobLocked(id)
	if j == nil {
		a.mcp.mu.Unlock()
		return a.MCPJobs(), nil
	}
	status := j.Status
	a.removeJobLocked(id)
	a.mcp.mu.Unlock()
	switch status {
	case jobStaged:
		if dir, err := a.jobStagePath(id); err == nil {
			_ = os.RemoveAll(dir)
		}
	case jobRendering:
		a.CancelRender()
	}
	a.emitJobs()
	return a.MCPJobs(), nil
}

// MCPDismiss — ẩn các cuốn đã xong việc (đã lưu, lỗi, đã dừng) khỏi thẻ thanh bên.
func (a *App) MCPDismiss() []MCPJob {
	a.mcp.mu.Lock()
	keep := a.mcp.jobs[:0]
	for _, j := range a.mcp.jobs {
		if j.Status == jobQueued || j.Status == jobRendering || j.Status == jobStaged {
			keep = append(keep, j)
		}
	}
	a.mcp.jobs = keep
	a.mcp.mu.Unlock()
	a.emitJobs()
	return a.MCPJobs()
}

func (a *App) removeJobLocked(id string) {
	for i, j := range a.mcp.jobs {
		if j.ID == id {
			a.mcp.jobs = append(a.mcp.jobs[:i], a.mcp.jobs[i+1:]...)
			return
		}
	}
}

func (a *App) emitJobs() {
	a.emit(eventMCPJobs, a.MCPJobs())
}

// loadStagedJobs — mở app: nạp lại các cuốn còn ở khu chờ (giữ 7 ngày), dọn phần dở dang.
func (a *App) loadStagedJobs() {
	entries, err := os.ReadDir(a.mcpStageRoot())
	if err != nil {
		return
	}
	for _, e := range entries {
		dir := filepath.Join(a.mcpStageRoot(), e.Name())
		if !e.IsDir() || !jobIDRe.MatchString(e.Name()) {
			continue
		}
		var j MCPJob
		b, err := os.ReadFile(filepath.Join(dir, mcpJobFile))
		if err != nil || json.Unmarshal(b, &j) != nil || j.ID != e.Name() {
			if info, err := e.Info(); err == nil && time.Since(info.ModTime()) > 24*time.Hour {
				_ = os.RemoveAll(dir)
			}
			continue
		}
		if time.Since(time.Unix(j.StagedAt, 0)) > mcpStageTTL {
			_ = os.RemoveAll(dir)
			continue
		}
		j.Status, j.AutoSave, j.PledgedAt, j.Error = jobStaged, false, "", ""
		a.mcp.mu.Lock()
		if a.findJobLocked(j.ID) == nil {
			jj := j
			a.mcp.jobs = append(a.mcp.jobs, &jj)
		}
		a.mcp.mu.Unlock()
	}
}

// expireStaged xoá các cuốn ở khu chờ quá 7 ngày chưa cam kết.
func (a *App) expireStaged() {
	a.mcp.mu.Lock()
	var old []string
	for _, j := range a.mcp.jobs {
		if j.Status == jobStaged && j.StagedAt > 0 && time.Since(time.Unix(j.StagedAt, 0)) > mcpStageTTL {
			old = append(old, j.ID)
		}
	}
	for _, id := range old {
		a.removeJobLocked(id)
	}
	a.mcp.mu.Unlock()
	for _, id := range old {
		if dir, err := a.jobStagePath(id); err == nil {
			_ = os.RemoveAll(dir)
		}
	}
}
