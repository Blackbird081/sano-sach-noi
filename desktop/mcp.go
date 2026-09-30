package main

// Máy chủ MCP trong máy (đường local): Claude Desktop, Claude Code, Codex… trên
// cùng máy nói chuyện với app Sano đang mở để xem, tạo, sửa sách. Chỉ nghe ở
// 127.0.0.1, mỗi yêu cầu phải kèm mã bí mật (~/Sano/.mcp/token), chặn yêu cầu
// từ trang web (Origin lạ, Host lạ). Không phiên (stateless) + trả JSON: cầu nối
// sano-mcp chỉ việc chuyển từng dòng JSON-RPC, app mở lại cũng nối tiếp được.
//
// M0: các việc Xem. M2: tạo sách (mcp_create.go). Không có công cụ xoá.

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
Trả lời người dùng bằng tiếng Việt có dấu. Tạo sách: create_book (gửi nội dung, xem lại mục lục cùng người dùng) → start_render (Sano tạo ngay, lần lượt nếu nhiều cuốn) → get_render_status. Tạo xong sách ở khu chờ, người dùng tick cam kết trên app thì mới vào Thư viện. Sửa sách: update_sections, find_replace (preview trước), update_book_info, set_cover, change_voice, set_pronunciation; theo dõi bằng get_edit_status. Mỗi lần sửa Sano giữ bản cũ để người dùng hoàn tác trong app. Không xoá được sách qua đây.`

// startMCP mở máy chủ MCP trong máy và ghi file hẹn cho cầu nối. Lỗi thì chỉ ghi
// log: app vẫn chạy bình thường, chỉ không kết nối AI được.
func (a *App) startMCP() {
	if a.loadMCPSettings().LocalOff {
		return
	}
	a.mu.Lock()
	running := a.mcpSrv != nil
	a.mu.Unlock()
	if running {
		return
	}
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
	a.mu.Lock()
	a.mcpSrv = srv
	a.mu.Unlock()
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("mcp: %v", err)
		}
	}()
}

// stopMCP đóng máy chủ lúc tắt app, xoá file hẹn (nếu là của app này).
func (a *App) stopMCP() {
	a.mu.Lock()
	srv := a.mcpSrv
	a.mcpSrv = nil
	a.mu.Unlock()
	if srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
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
	addTool(a, s, &mcp.Tool{Name: "list_books", Title: "Danh sách sách",
		Description: "Liệt kê các cuốn sách nói trong thư viện Sano (mới tạo trước): slug, tên, tác giả, giọng đọc, số chương, số mục, thời lượng.",
		Annotations: ro}, false, func(_ mcpNoInput, _ mcpBooks) string { return "Xem danh sách sách" }, a.mcpListBooks)
	addTool(a, s, &mcp.Tool{Name: "get_book", Title: "Xem một cuốn",
		Description: "Thông tin một cuốn và mục lục: từng mục có số thứ tự (index, bắt đầu từ 1), tên chương, tên mục, thời lượng.",
		Annotations: ro}, false, func(in mcpSlug, o mcpBookDetail) string { return "Xem mục lục " + quoteTitle(o.Title, in.Slug) }, a.mcpGetBook)
	addTool(a, s, &mcp.Tool{Name: "get_section_texts", Title: "Đọc lời các mục",
		Description: "Lời của các mục từ `from` đến `to` (index như get_book, tính cả hai đầu). Dài quá thì cắt, xem next_from để đọc tiếp.",
		Annotations: ro}, false, func(in mcpTextsInput, o mcpTexts) string { return textsLog(in, o, a.bookTitle(in.Slug)) }, a.mcpSectionTexts)
	addTool(a, s, &mcp.Tool{Name: "list_voices", Title: "Danh sách giọng đọc",
		Description: "Các giọng đọc bộ đọc trên máy có (tên, mô tả). Lần đầu mất vài giây.",
		Annotations: ro}, false, func(_ mcpNoInput, _ mcpVoices) string { return "Xem danh sách giọng đọc" }, a.mcpListVoices)
	addTool(a, s, &mcp.Tool{Name: "get_status", Title: "Tình trạng Sano",
		Description: "Phiên bản app, bộ đọc đã sẵn sàng chưa, có đang đọc (render) hay sửa sách nào không và tiến độ.",
		Annotations: ro}, false, nil, a.mcpStatus)
	f := false
	write := &mcp.ToolAnnotations{DestructiveHint: &f, OpenWorldHint: &f}
	// Việc tạo / sửa: cần quyền "Tạo và sửa sách" (màn MCP), ghi nhật ký.
	addTool(a, s, &mcp.Tool{Name: "create_book", Title: "Tạo bản nháp sách",
		Description: "Gửi toàn bộ nội dung một cuốn sách nói để Sano dựng bản nháp (chưa tạo sách). Định dạng văn bản: dòng đầu '% Tên sách'; mỗi chương '# Tên chương'; mỗi mục '## Tên mục'; dưới mỗi mục là các đoạn văn viết để nghe (câu ngắn, số và chữ viết tắt viết thành lời, không bảng, không gạch đầu dòng). Trả về draft_id, mục lục, thời lượng ước tính. Không nhận đường dẫn file: tự đọc tài liệu rồi gửi nội dung.",
		Annotations: write}, true, func(_ mcpCreateInput, o mcpDraftOut) string {
		return fmt.Sprintf("Tạo bản nháp %s (%d chương, %d mục, khoảng %d phút nghe)", quoteTitle(o.Title, "?"), len(o.Chapters), o.Sections, o.ListenMin)
	}, a.mcpCreateBook)
	addTool(a, s, &mcp.Tool{Name: "start_render", Title: "Tạo sách nói",
		Description: "Tạo sách nói (render) từ bản nháp trên máy. Hỏi người dùng trước khi gọi. Sano xếp hàng và tạo lần lượt từng cuốn khi máy rảnh, không cần chờ ai. Tạo xong sách nằm ở khu chờ: người dùng tick cam kết trên app Sano (một lần cho nhiều cuốn) thì mới vào Thư viện. Theo dõi bằng get_render_status.",
		Annotations: write}, true, func(in mcpDraftID, o mcpRenderOut) string {
		return "Xin tạo sách " + quoteTitle(firstTitle(o), in.DraftID)
	}, a.mcpStartRender)
	addTool(a, s, &mcp.Tool{Name: "get_render_status", Title: "Tiến độ tạo sách",
		Description: "Tình trạng các cuốn AI nhờ tạo: queued (chờ tạo), rendering (đang tạo, phần trăm, phút còn lại), waiting_pledge (đã tạo xong, chờ người dùng cam kết trên app), saved (đã vào Thư viện, có slug), failed, cancelled.",
		Annotations: ro}, false, nil, a.mcpRenderStatus)
	addTool(a, s, &mcp.Tool{Name: "cancel_render", Title: "Dừng tạo sách",
		Description: "Dừng một cuốn đang tạo hoặc bỏ khỏi hàng chờ tạo (không lưu gì). Chỉ dùng khi người dùng yêu cầu.",
		Annotations: write}, true, func(in mcpDraftID, o mcpRenderOut) string {
		return "Dừng tạo sách " + quoteTitle(firstTitle(o), in.DraftID)
	}, a.mcpCancelRender)
	// M3: sửa sách. Trước mỗi việc sửa một cuốn, Sano giữ bản cũ để người dùng hoàn tác.
	editTitle := func(o mcpEditOut, what string) string { return what + " " + quoteTitle(o.Title, o.Slug) }
	addTool(a, s, &mcp.Tool{Name: "update_sections", Title: "Sửa lời mục",
		Description: "Sửa lời (và tên) một hay nhiều mục của một cuốn rồi đọc lại đúng các mục đó bằng giọng hiện tại. Gửi TOÀN BỘ lời mới của mỗi mục (lấy lời hiện tại bằng get_section_texts). Sano giữ bản cũ để người dùng hoàn tác trong app.",
		Annotations: write}, true, func(in mcpUpdateSectionsInput, o mcpEditOut) string {
		return editTitle(o, fmt.Sprintf("Sửa lời %d mục cuốn", len(o.Sections)))
	}, a.mcpUpdateSections)
	addTool(a, s, &mcp.Tool{Name: "find_replace", Title: "Tìm và thay",
		Description: "Tìm một cụm chữ trong lời đọc cả cuốn (khớp nguyên cụm, phân biệt hoa thường) và thay, rồi đọc lại các mục bị thay. Gọi preview=true trước để xem chỗ khớp và cho người dùng duyệt.",
		Annotations: write}, true, func(in mcpFindReplaceInput, o mcpFindReplaceOut) string {
		if in.Preview {
			return fmt.Sprintf("Tìm %q trong %s: %d chỗ", in.Find, quoteTitle(o.Title, o.Slug), o.Matches)
		}
		return fmt.Sprintf("Thay %q → %q trong %s: %d chỗ, %d mục", in.Find, in.Replace, quoteTitle(o.Title, o.Slug), o.Matches, len(o.Sections))
	}, a.mcpFindReplace)
	addTool(a, s, &mcp.Tool{Name: "update_book_info", Title: "Sửa thông tin sách",
		Description: "Sửa tên sách, tác giả, dịch giả, nhà xuất bản, danh mục, bộ sách, số tập (chỉ gửi trường cần đổi). Đổi tên / tác giả thì bìa tự vẽ vẽ lại và lời giới thiệu đầu sách đọc lại.",
		Annotations: write}, true, func(_ mcpInfoInput, o mcpEditOut) string { return editTitle(o, "Sửa thông tin cuốn") }, a.mcpUpdateInfo)
	addTool(a, s, &mcp.Tool{Name: "set_cover", Title: "Đổi bìa",
		Description: "Đổi ảnh bìa: gửi ảnh jpg / png / webp dạng base64 (tối đa 10 MB), hoặc auto=true để dùng bìa Sano tự vẽ theo tên sách.",
		Annotations: write}, true, func(_ mcpCoverInput, o mcpEditOut) string { return editTitle(o, "Đổi bìa cuốn") }, a.mcpSetCover)
	addTool(a, s, &mcp.Tool{Name: "change_voice", Title: "Đổi giọng",
		Description: "Đọc lại cả cuốn bằng giọng khác (tên từ list_voices). Mất thời gian như tạo lại cả cuốn; xong hết mới thay giọng cũ. Hỏi người dùng trước khi gọi.",
		Annotations: write}, true, func(in mcpVoiceInput, o mcpEditOut) string {
		return editTitle(o, "Đổi giọng sang "+in.Voice+" cho cuốn")
	}, a.mcpChangeVoice)
	addTool(a, s, &mcp.Tool{Name: "set_pronunciation", Title: "Cách đọc một từ",
		Description: "Dạy Sano cách đọc một từ / chữ viết tắt (ví dụ SePay → Xi Pây). Có slug = chỉ cho cuốn đó; không có = từ điển chung cho sách tạo sau. Âm thanh đã có không đổi cho tới khi đọc lại mục chứa từ đó.",
		Annotations: write}, true, func(in mcpPronounceInput, _ mcpPronounceOut) string {
		return fmt.Sprintf("Cách đọc %s → %q", in.Word, in.Reading)
	}, a.mcpSetPronunciation)
	addTool(a, s, &mcp.Tool{Name: "get_edit_status", Title: "Tiến độ sửa sách",
		Description: "Tiến độ lượt đọc lại / đổi giọng đang chạy hoặc vừa xong: số mục, phần trăm, phút còn lại, lỗi.",
		Annotations: ro}, false, nil, a.mcpEditStatus)
	return s
}

// addTool đăng ký công cụ kèm: kiểm quyền "Tạo và sửa sách" (edit), ghi nhật ký (describe;
// nil = không ghi, cho các lệnh AI hỏi lặp như get_status, get_render_status).
func addTool[In, Out any](a *App, s *mcp.Server, t *mcp.Tool, edit bool, describe func(In, Out) string, h mcp.ToolHandlerFor[In, Out]) {
	mcp.AddTool(s, t, func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		client := "AI"
		if req != nil && req.Extra != nil && req.Extra.Header != nil {
			client = mcpClientName(req.Extra.Header.Get("X-Sano-Client"), req.Extra.Header.Get("User-Agent"))
		}
		if edit && a.loadMCPSettings().NoEdit {
			var zero Out
			a.logMCP(MCPLogEntry{Client: client, Tool: t.Name, Text: t.Title, Edit: true, Error: "không có quyền Tạo và sửa sách"})
			return nil, zero, errMCPNoEdit
		}
		res, out, err := h(context.WithValue(ctx, mcpClientKey{}, client), req, in)
		if describe != nil {
			e := MCPLogEntry{Client: client, Tool: t.Name, Edit: edit}
			if err != nil {
				e.Text, e.Error = t.Title, err.Error()
			} else {
				e.Text = describe(in, out)
				if u, ok := any(out).(interface{ undoRef() (string, string) }); ok {
					e.Slug, e.Undo = u.undoRef()
				}
			}
			a.logMCP(e)
		}
		return res, out, err
	})
}

type mcpClientKey struct{}

// clientFrom — tên phần mềm AI đang gọi (addTool gắn vào ctx).
func clientFrom(ctx context.Context) string {
	if c, ok := ctx.Value(mcpClientKey{}).(string); ok && c != "" {
		return c
	}
	return "AI"
}

func (a *App) bookTitle(slug string) string {
	if d, err := a.lib.Get(strings.TrimSpace(slug)); err == nil {
		return d.Title
	}
	return ""
}

func firstTitle(o mcpRenderOut) string {
	if len(o.Books) > 0 {
		return o.Books[0].Title
	}
	return ""
}

func quoteTitle(title, fallback string) string {
	if title == "" {
		title = fallback
	}
	return "\u201c" + title + "\u201d"
}

func textsLog(in mcpTextsInput, o mcpTexts, title string) string {
	if len(o.Sections) == 0 {
		return "Đọc lời " + quoteTitle(title, in.Slug)
	}
	a, b := o.Sections[0].Index, o.Sections[len(o.Sections)-1].Index
	if a == b {
		return fmt.Sprintf("Đọc lời mục %d cuốn %s", a, quoteTitle(title, in.Slug))
	}
	return fmt.Sprintf("Đọc lời mục %d–%d cuốn %s", a, b, quoteTitle(title, in.Slug))
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
