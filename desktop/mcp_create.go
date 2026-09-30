package main

// MCP · tạo sách (M2): AI gửi NỘI DUNG sách (văn bản "% Tên sách / # Chương / ## Mục",
// cùng quy ước với ô dán văn bản ở bước Cách đọc), không gửi đường dẫn file. Sano đổi thành
// .docx tạm trong ~/Sano/.tam/mcp, nạp mục lục → bản nháp. Bắt đầu đọc thì app hiện popup
// cam kết D19 cho người ngồi trước máy tick; tick đủ mới render, từ chối thì thôi. AI theo dõi
// bằng get_render_status. Không có công cụ xoá (D21).

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/internal/bookmaker"
)

const (
	eventMCPPledge = "mcp:pledge" // app hiện popup cam kết cho bản nháp AI gửi (nil = đóng popup)

	mcpDraftTTL     = 2 * time.Hour
	mcpMaxDrafts    = 5
	mcpPledgeTTL    = 10 * time.Minute // popup cam kết không ai tick → tự huỷ
	mcpDefaultVoice = "Hải Đăng"       // khớp DEFAULT_VOICE ở giao diện
)

// mcpDraft — bản nháp sách AI gửi, chờ bắt đầu đọc.
type mcpDraft struct {
	id       string
	docx     string
	settings BookSettings // chưa có RightsConfirmedAt: chỉ người dùng tick cam kết mới có
	outline  *bookmaker.Outline
	created  time.Time
}

// MCPPledge — yêu cầu cam kết đang chờ người dùng (gửi lên giao diện).
type MCPPledge struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Voice     string `json:"voice"`
	Chapters  int    `json:"chapters"`
	Sections  int    `json:"sections"`
	ListenMin int    `json:"listenMin"`
}

// mcpState — bản nháp + lượt cam kết của MCP (khoá riêng, không dùng chung a.mu).
type mcpState struct {
	mu      sync.Mutex
	drafts  map[string]*mcpDraft
	pledge  *MCPPledge
	pledged time.Time // lúc hiện popup
	// lastPledge — kết quả lượt cam kết gần nhất cho get_render_status: waiting | declined | expired | started | error
	lastID, lastState, lastErr string
}

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type mcpCreateInput struct {
	Text      string `json:"text" jsonschema:"toàn bộ nội dung sách: dòng '% Tên sách', mỗi chương một dòng '# Tên chương', mỗi mục một dòng '## Tên mục', dưới là các đoạn văn để đọc"`
	Title     string `json:"title,omitempty" jsonschema:"tên sách; bỏ trống thì lấy dòng %"`
	Author    string `json:"author,omitempty" jsonschema:"tác giả"`
	Category  string `json:"category,omitempty" jsonschema:"danh mục trong thư viện, ví dụ Kỹ năng"`
	Series    string `json:"series,omitempty" jsonschema:"tên bộ sách nếu cuốn này là một tập trong bộ"`
	Volume    int    `json:"volume,omitempty" jsonschema:"số tập trong bộ; bỏ trống = tập kế tiếp"`
	Voice     string `json:"voice,omitempty" jsonschema:"tên giọng (từ list_voices); bỏ trống = Hải Đăng"`
	IntroText string `json:"intro_text,omitempty" jsonschema:"lời giới thiệu đọc ở đầu sách; bỏ trống = tự tạo từ tên sách và tác giả"`
	NoIntro   bool   `json:"no_intro,omitempty" jsonschema:"true = không đọc lời giới thiệu ở đầu"`
}

type mcpDraftSection struct {
	Title string `json:"title"`
	Chars int    `json:"chars"`
	Skip  bool   `json:"skip,omitempty"` // trông như trang mục lục → không đọc
}

type mcpDraftChapter struct {
	Title    string            `json:"title"`
	Sections []mcpDraftSection `json:"sections"`
}

type mcpDraftOut struct {
	DraftID   string            `json:"draft_id"`
	Title     string            `json:"title"`
	Author    string            `json:"author,omitempty"`
	Voice     string            `json:"voice"`
	IntroText string            `json:"intro_text,omitempty"`
	Chapters  []mcpDraftChapter `json:"chapters"`
	Sections  int               `json:"sections"`
	Chars     int               `json:"chars"`
	ListenMin int               `json:"est_listen_min"`
	RenderMin int               `json:"est_render_min"`
	Warnings  []string          `json:"warnings,omitempty"`
	Next      string            `json:"next"`
}

func (a *App) mcpCreateBook(_ context.Context, _ *mcp.CallToolRequest, in mcpCreateInput) (*mcp.CallToolResult, mcpDraftOut, error) {
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil, mcpDraftOut{}, errors.New("text trống: gửi toàn bộ nội dung sách (% Tên sách, # Chương, ## Mục, đoạn văn)")
	}
	if len(text) > bookmaker.MaxPastedTextBytes {
		return nil, mcpDraftOut{}, fmt.Errorf("nội dung quá dài (%d MB), tối đa %d MB: chia thành nhiều tập", len(text)>>20, bookmaker.MaxPastedTextBytes>>20)
	}
	if !utf8.ValidString(text) {
		return nil, mcpDraftOut{}, errors.New("nội dung không phải chữ UTF-8")
	}
	voice := strings.TrimSpace(in.Voice)
	if voice == "" {
		voice = mcpDefaultVoice
	} else if vs, err := a.Voices(); err == nil {
		ok := false
		for _, v := range vs {
			ok = ok || v.Name == voice
		}
		if !ok {
			return nil, mcpDraftOut{}, fmt.Errorf("không có giọng %q: gọi list_voices để xem tên đúng", voice)
		}
	}

	data, err := bookmaker.TextToDocx(text)
	if err != nil {
		return nil, mcpDraftOut{}, err
	}
	dir := filepath.Join(a.lib.Root(), ".tam", "mcp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, mcpDraftOut{}, err
	}
	id := randomID()
	docx := filepath.Join(dir, id+".docx")
	if err := os.WriteFile(docx, data, 0o644); err != nil {
		return nil, mcpDraftOut{}, err
	}
	outline, err := bookmaker.Inspect(docx, bookmaker.InspectOptions{Pronunciations: []map[string]string{a.globalDict()}})
	if err != nil {
		_ = os.Remove(docx)
		return nil, mcpDraftOut{}, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = strings.TrimSpace(outline.Title)
	}
	if title == "" {
		_ = os.Remove(docx)
		return nil, mcpDraftOut{}, errors.New("chưa có tên sách: thêm dòng '% Tên sách' ở đầu hoặc truyền title")
	}
	author := strings.TrimSpace(in.Author)
	intro := ""
	if !in.NoIntro {
		intro = strings.TrimSpace(in.IntroText)
		if intro == "" {
			intro = defaultIntroText(title, author)
		}
	}

	s := BookSettings{
		Path: docx, Title: title, Author: author, Category: strings.TrimSpace(in.Category),
		Series: strings.TrimSpace(in.Series), Volume: in.Volume, Voice: voice, IntroText: intro,
	}
	out := mcpDraftOut{DraftID: id, Title: title, Author: author, Voice: voice, IntroText: intro, Chapters: []mcpDraftChapter{}}
	chars := len([]rune(intro))
	for _, ch := range outline.Chapters {
		c := mcpDraftChapter{Title: ch.Title, Sections: []mcpDraftSection{}}
		for _, sec := range ch.Sections {
			c.Sections = append(c.Sections, mcpDraftSection{Title: sec.Title, Chars: sec.Chars, Skip: sec.TOC})
			if sec.TOC {
				s.DropStems = append(s.DropStems, sec.Stem) // như bước Mục lục: trang mục lục bỏ sẵn
				continue
			}
			out.Sections++
			chars += sec.Chars
		}
		out.Chapters = append(out.Chapters, c)
	}
	if out.Sections == 0 {
		_ = os.Remove(docx)
		return nil, mcpDraftOut{}, errors.New("không có mục nào để đọc: mỗi chương cần ít nhất một đoạn văn")
	}
	out.Chars = chars
	out.ListenMin = max(1, chars/15/60) // ~15 ký tự/giây nghe, ~70 ký tự/giây render (như giao diện)
	out.RenderMin = max(1, chars/70/60)
	out.Warnings = outlineWarnings(outline.Warnings)
	out.Next = "Cho người dùng xem mục lục này. Họ đồng ý thì gọi start_render với draft_id."

	a.mcp.mu.Lock()
	a.pruneDraftsLocked()
	if a.mcp.drafts == nil {
		a.mcp.drafts = map[string]*mcpDraft{}
	}
	a.mcp.drafts[id] = &mcpDraft{id: id, docx: docx, settings: s, outline: outline, created: time.Now()}
	a.mcp.mu.Unlock()
	return nil, out, nil
}

// defaultIntroText — như defaultIntro() ở giao diện (lib/store.ts).
func defaultIntroText(title, author string) string {
	lines := []string{"Bạn đang nghe sách nói.", "Cuốn sách: " + title + "."}
	if author != "" {
		lines = append(lines, "Tác giả: "+author+".")
	}
	return strings.Join(lines, "\n\n")
}

func outlineWarnings(w bookmaker.LoadWarnings) []string {
	var out []string
	if len(w.FakeHeadings) > 0 {
		out = append(out, fmt.Sprintf("%d dòng trông như tiêu đề nhưng không có dấu #/##, sẽ đọc như đoạn văn: %s", len(w.FakeHeadings), strings.Join(first(w.FakeHeadings, 5), "; ")))
	}
	if len(w.UnknownAcronyms) > 0 {
		words := make([]string, 0, len(w.UnknownAcronyms))
		for _, a := range w.UnknownAcronyms {
			words = append(words, fmt.Sprintf("%s (%d lần)", a.Word, a.Count))
		}
		out = append(out, "Chữ viết tắt chưa có cách đọc, bộ đọc sẽ đánh vần từng chữ: "+strings.Join(first(words, 15), ", ")+". Nên viết thành lời trong nội dung.")
	}
	return out
}

func first(s []string, n int) []string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// pruneDraftsLocked bỏ bản nháp quá hạn và giữ tối đa mcpMaxDrafts bản mới nhất.
func (a *App) pruneDraftsLocked() {
	var oldest *mcpDraft
	for id, d := range a.mcp.drafts {
		if time.Since(d.created) > mcpDraftTTL {
			a.dropDraftLocked(id)
			continue
		}
		if oldest == nil || d.created.Before(oldest.created) {
			oldest = d
		}
	}
	if len(a.mcp.drafts) >= mcpMaxDrafts && oldest != nil {
		a.dropDraftLocked(oldest.id)
	}
}

func (a *App) dropDraftLocked(id string) {
	if d, ok := a.mcp.drafts[id]; ok {
		_ = os.Remove(d.docx)
		delete(a.mcp.drafts, id)
	}
}

type mcpDraftID struct {
	DraftID string `json:"draft_id" jsonschema:"draft_id trả về từ create_book"`
}

type mcpRenderOut struct {
	Status    string  `json:"status"` // waiting_pledge | declined | expired | rendering | done | error | cancelled | none
	Message   string  `json:"message"`
	Title     string  `json:"title,omitempty"`
	Slug      string  `json:"slug,omitempty"`
	Percent   float64 `json:"percent,omitempty"`
	Done      int     `json:"sections_done,omitempty"`
	Total     int     `json:"sections_total,omitempty"`
	RemainMin int     `json:"remain_min,omitempty"`
}

func (a *App) mcpStartRender(_ context.Context, _ *mcp.CallToolRequest, in mcpDraftID) (*mcp.CallToolResult, mcpRenderOut, error) {
	a.mu.Lock()
	busy := a.renderBusyLocked()
	a.mu.Unlock()
	if busy != nil {
		return nil, mcpRenderOut{}, fmt.Errorf("chưa đọc được: %v", busy)
	}
	a.mcp.mu.Lock()
	defer a.mcp.mu.Unlock()
	d, ok := a.mcp.drafts[strings.TrimSpace(in.DraftID)]
	if !ok {
		return nil, mcpRenderOut{}, errors.New("không thấy bản nháp (quá 2 giờ hoặc sai draft_id): gọi create_book lại")
	}
	if a.mcp.pledge != nil && time.Since(a.mcp.pledged) < mcpPledgeTTL {
		return nil, mcpRenderOut{}, errors.New("đang chờ người dùng cam kết cho một cuốn khác trên app Sano")
	}
	listen := 0
	for _, ch := range d.outline.Chapters {
		for _, s := range ch.Sections {
			if !s.TOC {
				listen += s.Chars
			}
		}
	}
	p := &MCPPledge{ID: d.id, Title: d.settings.Title, Author: d.settings.Author, Voice: d.settings.Voice,
		Chapters: len(d.outline.Chapters), Sections: d.outline.Sections, ListenMin: max(1, listen/15/60)}
	a.mcp.pledge, a.mcp.pledged = p, time.Now()
	a.mcp.lastID, a.mcp.lastState, a.mcp.lastErr = d.id, "waiting", ""
	a.emit(eventMCPPledge, p)
	if a.ctx != nil {
		wruntime.WindowUnminimise(a.ctx)
		wruntime.WindowShow(a.ctx)
	}
	return nil, mcpRenderOut{Status: "waiting_pledge", Title: p.Title,
		Message: "App Sano đang hiện popup cam kết. Nhờ người dùng tick đủ các ô trên máy tính rồi bấm \"Cam kết và đọc\". Gọi get_render_status để theo dõi."}, nil
}

// MCPPledgeAnswer — giao diện báo người dùng đã cam kết (ok) hay đóng popup.
func (a *App) MCPPledgeAnswer(id string, ok bool) error {
	a.mcp.mu.Lock()
	p := a.mcp.pledge
	if p == nil || p.ID != id {
		a.mcp.mu.Unlock()
		return errors.New("yêu cầu này không còn nữa")
	}
	a.mcp.pledge = nil
	d := a.mcp.drafts[id]
	if !ok || d == nil {
		a.mcp.lastState = "declined"
		a.mcp.mu.Unlock()
		return nil
	}
	s := d.settings
	s.RightsConfirmedAt = time.Now().UTC().Format(time.RFC3339)
	a.mcp.mu.Unlock()

	_, err := a.startRender(s, "mcp")
	a.mcp.mu.Lock()
	defer a.mcp.mu.Unlock()
	if err != nil {
		a.mcp.lastState, a.mcp.lastErr = "error", err.Error()
		return err
	}
	a.mcp.lastState = "started"
	delete(a.mcp.drafts, id) // file .docx giữ tới khi dọn .tam (render còn đọc)
	return nil
}

// MCPPendingPledge — popup cam kết đang chờ (mở lại cửa sổ / tải lại giao diện).
func (a *App) MCPPendingPledge() *MCPPledge {
	a.mcp.mu.Lock()
	defer a.mcp.mu.Unlock()
	a.expirePledgeLocked()
	return a.mcp.pledge
}

func (a *App) expirePledgeLocked() {
	if a.mcp.pledge != nil && time.Since(a.mcp.pledged) >= mcpPledgeTTL {
		a.mcp.pledge = nil
		a.mcp.lastState = "expired"
		a.emit(eventMCPPledge, nil)
	}
}

func (a *App) mcpRenderStatus(context.Context, *mcp.CallToolRequest, mcpNoInput) (*mcp.CallToolResult, mcpRenderOut, error) {
	a.mcp.mu.Lock()
	a.expirePledgeLocked()
	state, errText := a.mcp.lastState, a.mcp.lastErr
	var title string
	if p := a.mcp.pledge; p != nil {
		title = p.Title
	}
	a.mcp.mu.Unlock()
	switch state {
	case "waiting":
		return nil, mcpRenderOut{Status: "waiting_pledge", Title: title, Message: "Đang chờ người dùng tick cam kết trên app Sano."}, nil
	case "declined":
		return nil, mcpRenderOut{Status: "declined", Message: "Người dùng đã đóng popup cam kết, chưa đọc. Hỏi lại người dùng trước khi gọi start_render lần nữa."}, nil
	case "expired":
		return nil, mcpRenderOut{Status: "expired", Message: "Popup cam kết quá 10 phút không ai tick nên đã đóng. Gọi start_render lại khi người dùng ngồi trước máy."}, nil
	case "error":
		return nil, mcpRenderOut{Status: "error", Message: errText}, nil
	}
	r := a.RenderStatus()
	if r == nil || r.Source != "mcp" {
		return nil, mcpRenderOut{Status: "none", Message: "Chưa có cuốn nào AI nhờ đọc trong lần mở app này."}, nil
	}
	out := mcpRenderOut{Title: r.Title, Slug: r.Slug, Done: r.Progress.Done, Total: r.Progress.Total,
		Percent: progressPercent(r.Progress.DoneChars, r.Progress.TotalChars)}
	switch {
	case r.Running:
		out.Status = "rendering"
		out.Message = "Đang đọc thành sách trên máy. Hỏi lại sau vài phút."
		if r.Progress.DoneChars > 0 && r.Progress.ElapsedSec > 0 {
			left := float64(r.Progress.TotalChars-r.Progress.DoneChars) * float64(r.Progress.ElapsedSec) / float64(r.Progress.DoneChars)
			out.RemainMin = max(1, int(left/60+0.5))
		}
	case r.Done:
		out.Status, out.Percent = "done", 100
		out.Message = "Đã tạo xong, sách nằm trong thư viện Sano. Dùng get_book với slug để xem."
	case r.Cancelled:
		out.Status, out.Message = "cancelled", "Lượt đọc đã bị dừng, không lưu gì."
	default:
		out.Status, out.Message = "error", r.Error
	}
	return nil, out, nil
}

func (a *App) mcpCancelRender(context.Context, *mcp.CallToolRequest, mcpNoInput) (*mcp.CallToolResult, mcpRenderOut, error) {
	r := a.RenderStatus()
	if r == nil || !r.Running || r.Source != "mcp" {
		return nil, mcpRenderOut{}, errors.New("không có lượt đọc nào do AI bắt đầu đang chạy")
	}
	a.CancelRender()
	return nil, mcpRenderOut{Status: "cancelled", Title: r.Title, Message: "Đã dừng lượt đọc, không lưu gì vào thư viện."}, nil
}
