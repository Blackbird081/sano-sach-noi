package main

// Máy chủ MCP trong máy (đường local): Claude Desktop, Claude Code, Codex… trên
// cùng máy nói chuyện với app Sano đang mở để xem, tạo, sửa sách. Chỉ nghe ở
// 127.0.0.1, mỗi yêu cầu phải kèm mã bí mật (~/Sano/.mcp/token), chặn yêu cầu
// từ trang web (Origin lạ, Host lạ). Không phiên (stateless) + trả JSON: cầu nối
// sano-mcp chỉ việc chuyển từng dòng JSON-RPC, app mở lại cũng nối tiếp được.
//
// M0: chỉ các việc Xem. Tạo, sửa, xuất làm ở các bước sau.

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sano/desktop/internal/mcplink"
)

// mcpPort — cổng ưu tiên (bận thì lấy cổng bất kỳ; cầu nối đọc cổng thật trong file hẹn).
const mcpPort = 39390

const mcpInstructions = `Sano là phần mềm làm sách nói tiếng Việt chạy trên máy người dùng: biến tài liệu (Tiêu đề, Chương, Mục) thành sách nói bằng giọng đọc VieNeu ngay trên máy.
Thư viện gồm các cuốn đã tạo; mỗi cuốn có nhiều chương, mỗi chương nhiều mục (tiểu mục), mỗi mục là một file tiếng.
Bắt đầu bằng list_books để lấy slug của cuốn, rồi get_book xem mục lục, get_section_texts đọc lời từng mục.
Trả lời người dùng bằng tiếng Việt có dấu. Bản này chỉ xem được; tạo sách, sửa lời, xuất M4B sẽ có ở bản sau.`

// startMCP mở máy chủ MCP trong máy và ghi file hẹn cho cầu nối. Lỗi thì chỉ ghi
// log: app vẫn chạy bình thường, chỉ không kết nối AI được.
func (a *App) startMCP() {
	root := a.lib.Root()
	token, err := mcplink.Token(root)
	if err != nil {
		log.Printf("mcp: tạo mã kết nối: %v", err)
		return
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", mcpPort))
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		log.Printf("mcp: mở cổng: %v", err)
		return
	}
	srv := &http.Server{Handler: a.mcpHandler(token), ReadHeaderTimeout: 10 * time.Second}
	url := fmt.Sprintf("http://%s/mcp", ln.Addr().String())
	if err := mcplink.WriteLocal(root, mcplink.Local{URL: url, PID: os.Getpid()}); err != nil {
		log.Printf("mcp: ghi file hẹn: %v", err)
	}
	a.mcpSrv = srv
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("mcp: %v", err)
		}
	}()
}

// stopMCP đóng máy chủ lúc tắt app, xoá file hẹn (nếu là của app này).
func (a *App) stopMCP() {
	if a.mcpSrv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = a.mcpSrv.Shutdown(ctx)
	mcplink.RemoveLocal(a.lib.Root(), os.Getpid())
}

// mcpHandler: /mcp, kiểm mã bí mật, chặn gọi từ trình duyệt. Máy chủ MCP của SDK
// tự chặn Host lạ khi nối qua 127.0.0.1 (chống DNS rebinding).
func (a *App) mcpHandler(token string) http.Handler {
	server := a.newMCPServer()
	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})
	protect := http.NewCrossOriginProtection()
	want := []byte("Bearer " + token)
	mux := http.NewServeMux()
	mux.Handle("/mcp", protect.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trang web mở trên máy luôn gửi Origin; phần mềm AI thì không.
		if r.Header.Get("Origin") != "" {
			http.Error(w, "không nhận yêu cầu từ trang web", http.StatusForbidden)
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), want) != 1 {
			http.Error(w, "sai mã kết nối Sano", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})))
	return mux
}

func (a *App) newMCPServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "sano", Title: "Sano · Sách nói", Version: a.Version()},
		&mcp.ServerOptions{Instructions: mcpInstructions})
	ro := &mcp.ToolAnnotations{ReadOnlyHint: true}
	mcp.AddTool(s, &mcp.Tool{Name: "list_books", Title: "Danh sách sách",
		Description: "Liệt kê các cuốn sách nói trong thư viện Sano (mới tạo trước): slug, tên, tác giả, giọng đọc, số chương, số mục, thời lượng.",
		Annotations: ro}, a.mcpListBooks)
	mcp.AddTool(s, &mcp.Tool{Name: "get_book", Title: "Xem một cuốn",
		Description: "Thông tin một cuốn và mục lục: từng mục có số thứ tự (index, bắt đầu từ 1), tên chương, tên mục, thời lượng.",
		Annotations: ro}, a.mcpGetBook)
	mcp.AddTool(s, &mcp.Tool{Name: "get_section_texts", Title: "Đọc lời các mục",
		Description: "Lời của các mục từ `from` đến `to` (index như get_book, tính cả hai đầu). Dài quá thì cắt, xem next_from để đọc tiếp.",
		Annotations: ro}, a.mcpSectionTexts)
	mcp.AddTool(s, &mcp.Tool{Name: "list_voices", Title: "Danh sách giọng đọc",
		Description: "Các giọng đọc bộ đọc trên máy có (tên, mô tả). Lần đầu mất vài giây.",
		Annotations: ro}, a.mcpListVoices)
	mcp.AddTool(s, &mcp.Tool{Name: "get_status", Title: "Tình trạng Sano",
		Description: "Phiên bản app, bộ đọc đã sẵn sàng chưa, có đang đọc (render) hay sửa sách nào không và tiến độ.",
		Annotations: ro}, a.mcpStatus)
	return s
}

type mcpNoInput struct{}

type mcpBook struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Author      string `json:"author,omitempty"`
	Translator  string `json:"translator,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	Category    string `json:"category,omitempty"`
	Series      string `json:"series,omitempty"`
	Volume      int    `json:"volume,omitempty"`
	Voice       string `json:"voice,omitempty"`
	Chapters    int    `json:"chapters"`
	Sections    int    `json:"sections"`
	DurationSec int    `json:"duration_sec"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type mcpBooks struct {
	Books []mcpBook `json:"books"`
}

func (a *App) mcpListBooks(context.Context, *mcp.CallToolRequest, mcpNoInput) (*mcp.CallToolResult, mcpBooks, error) {
	books, err := a.lib.List()
	if err != nil {
		return nil, mcpBooks{}, err
	}
	out := mcpBooks{Books: make([]mcpBook, 0, len(books))}
	for _, b := range books {
		out.Books = append(out.Books, mcpBook{
			Slug: b.Slug, Title: b.Title, Author: b.Author, Translator: b.Translator, Publisher: b.Publisher,
			Category: b.Category, Series: b.Series, Volume: b.Volume, Voice: b.Voice,
			Chapters: b.Chapters, Sections: b.Sections, DurationSec: b.DurationSec, CreatedAt: b.CreatedAt,
		})
	}
	return nil, out, nil
}

type mcpSlug struct {
	Slug string `json:"slug" jsonschema:"slug của cuốn sách (từ list_books)"`
}

type mcpSection struct {
	Index       int    `json:"index"`
	Chapter     string `json:"chapter,omitempty"`
	Title       string `json:"title"`
	DurationSec int    `json:"duration_sec"`
}

type mcpBookDetail struct {
	mcpBook
	Sections []mcpSection `json:"sections"`
}

func (a *App) mcpGetBook(_ context.Context, _ *mcp.CallToolRequest, in mcpSlug) (*mcp.CallToolResult, mcpBookDetail, error) {
	d, err := a.lib.Get(strings.TrimSpace(in.Slug))
	if err != nil {
		return nil, mcpBookDetail{}, mcpNotFound(in.Slug, err)
	}
	b := d.Book
	out := mcpBookDetail{mcpBook: mcpBook{
		Slug: b.Slug, Title: b.Title, Author: b.Author, Translator: b.Translator, Publisher: b.Publisher,
		Category: b.Category, Series: b.Series, Volume: b.Volume, Voice: b.Voice,
		Chapters: b.Chapters, Sections: b.Sections, DurationSec: b.DurationSec, CreatedAt: b.CreatedAt,
	}, Sections: make([]mcpSection, 0, len(d.Tracks))}
	for i, t := range d.Tracks {
		out.Sections = append(out.Sections, mcpSection{Index: i + 1, Chapter: t.Chapter, Title: t.Title, DurationSec: t.DurationSec})
	}
	return nil, out, nil
}

type mcpTextsInput struct {
	Slug string `json:"slug" jsonschema:"slug của cuốn sách"`
	From int    `json:"from,omitempty" jsonschema:"mục đầu (index từ 1), bỏ trống = 1"`
	To   int    `json:"to,omitempty" jsonschema:"mục cuối (tính cả mục này), bỏ trống = đến hết"`
}

type mcpSectionText struct {
	mcpSection
	Text string `json:"text"`
}

type mcpTexts struct {
	Sections []mcpSectionText `json:"sections"`
	NextFrom int              `json:"next_from,omitempty"` // còn mục chưa trả (cắt vì dài) → gọi lại với from = next_from
}

// mcpTextLimit — số ký tự lời tối đa một lần trả (khoảng 25–30 phút nghe), tránh
// tràn khung trò chuyện của AI.
const mcpTextLimit = 40000

func (a *App) mcpSectionTexts(_ context.Context, _ *mcp.CallToolRequest, in mcpTextsInput) (*mcp.CallToolResult, mcpTexts, error) {
	slug := strings.TrimSpace(in.Slug)
	d, err := a.lib.Get(slug)
	if err != nil {
		return nil, mcpTexts{}, mcpNotFound(slug, err)
	}
	texts, err := a.lib.Texts(slug)
	if err != nil {
		return nil, mcpTexts{}, err
	}
	n := len(d.Tracks)
	from, to := max(in.From, 1), in.To
	if to <= 0 || to > n {
		to = n
	}
	if from > to {
		return nil, mcpTexts{}, fmt.Errorf("cuốn này có %d mục, from/to không hợp lệ", n)
	}
	out := mcpTexts{Sections: []mcpSectionText{}}
	size := 0
	for i := from; i <= to; i++ {
		t, txt := d.Tracks[i-1], ""
		if i-1 < len(texts) {
			txt = texts[i-1].Text
		}
		if size > 0 && size+len(txt) > mcpTextLimit {
			out.NextFrom = i
			break
		}
		size += len(txt)
		out.Sections = append(out.Sections, mcpSectionText{
			mcpSection: mcpSection{Index: i, Chapter: t.Chapter, Title: t.Title, DurationSec: t.DurationSec}, Text: txt,
		})
	}
	return nil, out, nil
}

type mcpVoice struct {
	Name     string `json:"name"`
	Desc     string `json:"desc,omitempty"`
	Featured bool   `json:"featured,omitempty"`
}

type mcpVoices struct {
	Voices []mcpVoice `json:"voices"`
}

func (a *App) mcpListVoices(context.Context, *mcp.CallToolRequest, mcpNoInput) (*mcp.CallToolResult, mcpVoices, error) {
	vs, err := a.Voices()
	if err != nil {
		return nil, mcpVoices{}, fmt.Errorf("chưa hỏi được bộ đọc (%v). Kiểm tra bộ đọc trong app Sano", err)
	}
	out := mcpVoices{Voices: make([]mcpVoice, 0, len(vs))}
	for _, v := range vs {
		out.Voices = append(out.Voices, mcpVoice{Name: v.Name, Desc: v.Desc, Featured: v.Featured})
	}
	return nil, out, nil
}

type mcpJob struct {
	Kind    string  `json:"kind"` // render | edit
	Title   string  `json:"title,omitempty"`
	Slug    string  `json:"slug,omitempty"`
	Running bool    `json:"running"`
	Done    bool    `json:"done"`
	Error   string  `json:"error,omitempty"`
	Percent float64 `json:"percent"`
}

type mcpStatusOut struct {
	Version  string   `json:"version"`
	Library  string   `json:"library"`
	TTSReady bool     `json:"tts_ready"`
	TTSNote  string   `json:"tts_note,omitempty"`
	Jobs     []mcpJob `json:"jobs"`
}

func (a *App) mcpStatus(context.Context, *mcp.CallToolRequest, mcpNoInput) (*mcp.CallToolResult, mcpStatusOut, error) {
	out := mcpStatusOut{Version: a.Version(), Library: a.lib.BooksRoot(), Jobs: []mcpJob{}}
	st := a.CheckTTS()
	out.TTSReady = st.Ready
	if !st.Ready {
		out.TTSNote = strings.TrimSpace(st.Message + " " + st.Detail)
	}
	if r := a.RenderStatus(); r != nil && (r.Running || r.Done || r.Error != "") {
		out.Jobs = append(out.Jobs, mcpJob{Kind: "render", Title: r.Title, Slug: r.Slug, Running: r.Running, Done: r.Done,
			Error: r.Error, Percent: progressPercent(r.Progress.Done, r.Progress.Total)})
	}
	if e := a.EditStatus(); e != nil && (e.Running || e.Done || e.Error != "") {
		out.Jobs = append(out.Jobs, mcpJob{Kind: "edit", Title: e.BookTitle, Slug: e.Slug, Running: e.Running, Done: e.Done,
			Error: e.Error, Percent: progressPercent(e.Finished, e.Total)})
	}
	return nil, out, nil
}

func progressPercent(done, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(int(float64(done)/float64(total)*1000)) / 10
}

func mcpNotFound(slug string, err error) error {
	return fmt.Errorf("không mở được cuốn %q (%v). Gọi list_books để lấy đúng slug", slug, err)
}
