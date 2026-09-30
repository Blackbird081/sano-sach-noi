package main

// MCP · bản nháp sách: AI gửi NỘI DUNG sách (văn bản "% Tên sách / # Chương / ## Mục", cùng quy
// ước với ô dán văn bản ở bước Cách đọc), không gửi đường dẫn file. Sano đổi thành .docx tạm trong
// ~/Sano/.tam/mcp, nạp mục lục → bản nháp. Tạo sách, khu chờ cam kết: mcp_queue.go (D22).
// Không có công cụ xoá (D21).

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

	"sano/internal/bookmaker"
)

const (
	mcpDraftTTL     = 2 * time.Hour
	mcpMaxDrafts    = 10
	mcpDefaultVoice = "Hải Đăng" // khớp DEFAULT_VOICE ở giao diện
)

// mcpDraft — bản nháp sách AI gửi, chờ bắt đầu tạo.
type mcpDraft struct {
	id       string
	docx     string
	settings BookSettings // chưa có RightsConfirmedAt: chỉ người dùng tick cam kết mới có
	outline  *bookmaker.Outline
	created  time.Time
}

// mcpState — bản nháp + lượt cam kết của MCP (khoá riêng, không dùng chung a.mu).
type mcpState struct {
	mu      sync.Mutex
	drafts  map[string]*mcpDraft
	jobs    []*MCPJob // sách AI nhờ tạo: chờ tạo → đang tạo → khu chờ cam kết → đã lưu (mcp_queue.go)
	pumping bool
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
	out.Next = "Cho người dùng xem mục lục này. Họ đồng ý thì gọi start_render với draft_id. Có thể tạo nhiều cuốn: gọi create_book + start_render cho từng cuốn, Sano tạo lần lượt."

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
