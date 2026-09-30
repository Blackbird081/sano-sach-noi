package main

// MCP · sửa sách (M3): sửa lời từng mục rồi đọc lại, tìm và thay cả cuốn, sửa thông tin, bìa,
// đổi giọng, cách đọc một từ. Trước mỗi việc sửa một cuốn, Sano giữ bản cũ (library.TakeSnapshot)
// để người dùng hoàn tác ở màn MCP. Không có lệnh xoá sách, không có lệnh hoàn tác cho AI.

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"sano/desktop/internal/library"
)

// mcpEditOut — kết quả một việc sửa (kèm bản cũ để nhật ký có nút Hoàn tác).
type mcpEditOut struct {
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Sections []int  `json:"sections,omitempty"` // các mục sẽ đọc lại (index từ 1)
	Reread   bool   `json:"rereading,omitempty"`
	snapshot string // mã bản cũ (không gửi AI)
}

func (o mcpEditOut) undoRef() (string, string) { return o.Slug, o.snapshot }

// mcpBook — mở cuốn để sửa + kiểm không có lượt sửa nào đang chạy trên máy.
func (a *App) mcpBook(slug string) (*library.EditInfo, error) {
	slug = strings.TrimSpace(slug)
	info, err := a.lib.Edit(slug)
	if err != nil {
		return nil, mcpNotFound(slug, err)
	}
	if st := a.editing(); st != nil {
		return nil, fmt.Errorf("Sano đang sửa cuốn %q (%s), đợi xong rồi thử lại (xem get_edit_status)", st.BookTitle, map[bool]string{true: "đổi giọng", false: "đọc lại"}[st.Kind == editKindVoice])
	}
	if info.VoiceJob != nil {
		return nil, fmt.Errorf("cuốn này đang đổi giọng dở sang %s, đợi xong rồi sửa", info.VoiceJob.Voice)
	}
	return info, nil
}

func (a *App) snapshot(slug, label string) (string, error) {
	s, err := a.lib.TakeSnapshot(slug, label)
	if err != nil {
		return "", err
	}
	return s.ID, nil
}

// ── Sửa lời ──────────────────────────────────────────────────────────────

type mcpSectionEdit struct {
	Index int    `json:"index" jsonschema:"số thứ tự mục (từ 1, như get_book)"`
	Title string `json:"title,omitempty" jsonschema:"tên mục mới; bỏ trống = giữ tên cũ"`
	Text  string `json:"text" jsonschema:"toàn bộ lời mới của mục (viết để nghe)"`
}

type mcpUpdateSectionsInput struct {
	Slug     string           `json:"slug" jsonschema:"slug của cuốn"`
	Sections []mcpSectionEdit `json:"sections" jsonschema:"các mục cần sửa"`
}

const mcpMaxSectionText = 60000

func (a *App) mcpUpdateSections(_ context.Context, _ *mcp.CallToolRequest, in mcpUpdateSectionsInput) (*mcp.CallToolResult, mcpEditOut, error) {
	info, err := a.mcpBook(in.Slug)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	if len(in.Sections) == 0 {
		return nil, mcpEditOut{}, errors.New("chưa có mục nào để sửa")
	}
	var edits []SectionEdit
	var idx []int
	for _, e := range in.Sections {
		if e.Index < 1 || e.Index > len(info.Sections) {
			return nil, mcpEditOut{}, fmt.Errorf("cuốn này có %d mục, không có mục %d", len(info.Sections), e.Index)
		}
		if len(e.Text) > mcpMaxSectionText || !utf8.ValidString(e.Text+e.Title) {
			return nil, mcpEditOut{}, fmt.Errorf("lời mục %d quá dài hoặc không phải chữ UTF-8", e.Index)
		}
		edits = append(edits, SectionEdit{Index: e.Index - 1, Title: e.Title, Text: e.Text})
		idx = append(idx, e.Index)
	}
	return a.rereadWithSnapshot(info, edits, idx, fmt.Sprintf("AI sửa lời %d mục", len(edits)))
}

// rereadWithSnapshot giữ bản cũ rồi đọc lại các mục đã sửa.
func (a *App) rereadWithSnapshot(info *library.EditInfo, edits []SectionEdit, idx []int, label string) (*mcp.CallToolResult, mcpEditOut, error) {
	snap, err := a.snapshot(info.Slug, label)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	if _, err := a.StartReread(info.Slug, edits); err != nil {
		return nil, mcpEditOut{}, err
	}
	return nil, mcpEditOut{Slug: info.Slug, Title: info.Title, Sections: idx, Reread: true, snapshot: snap,
		Message: fmt.Sprintf("Đang đọc lại %d mục bằng giọng %s. Theo dõi bằng get_edit_status. Người dùng hoàn tác được ở mục MCP trong app.", len(edits), info.Voice)}, nil
}

// ── Tìm và thay ──────────────────────────────────────────────────────────

type mcpFindReplaceInput struct {
	Slug          string `json:"slug" jsonschema:"slug của cuốn"`
	Find          string `json:"find" jsonschema:"cụm chữ cần tìm (khớp nguyên cụm, phân biệt hoa thường)"`
	Replace       string `json:"replace" jsonschema:"thay bằng"`
	IncludeTitles bool   `json:"include_titles,omitempty" jsonschema:"thay cả trong tên mục"`
	Preview       bool   `json:"preview,omitempty" jsonschema:"true = chỉ liệt kê chỗ khớp, chưa thay"`
}

type mcpHit struct {
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Count   int    `json:"count"`
	Excerpt string `json:"excerpt"`
}

type mcpFindReplaceOut struct {
	mcpEditOut
	Matches int      `json:"matches"`
	Hits    []mcpHit `json:"hits"`
}

func (o mcpFindReplaceOut) undoRef() (string, string) { return o.mcpEditOut.undoRef() }

// excerpt — đoạn quanh chỗ khớp đầu tiên (để AI / người dùng xem ngữ cảnh).
func excerpt(text, needle string) string {
	i := strings.Index(text, needle)
	if i < 0 {
		return ""
	}
	r := []rune(text)
	start := utf8.RuneCountInString(text[:i])
	from, to := max(0, start-40), min(len(r), start+utf8.RuneCountInString(needle)+40)
	s := string(r[from:to])
	if from > 0 {
		s = "…" + s
	}
	if to < len(r) {
		s += "…"
	}
	return strings.Join(strings.Fields(s), " ")
}

func (a *App) mcpFindReplace(_ context.Context, _ *mcp.CallToolRequest, in mcpFindReplaceInput) (*mcp.CallToolResult, mcpFindReplaceOut, error) {
	var info *library.EditInfo
	var err error
	if in.Preview {
		info, err = a.lib.Edit(strings.TrimSpace(in.Slug))
		if err != nil {
			err = mcpNotFound(in.Slug, err)
		}
	} else {
		info, err = a.mcpBook(in.Slug)
	}
	if err != nil {
		return nil, mcpFindReplaceOut{}, err
	}
	if in.Find == "" || utf8.RuneCountInString(in.Find) > 200 || utf8.RuneCountInString(in.Replace) > 1000 {
		return nil, mcpFindReplaceOut{}, errors.New("find phải có chữ (tối đa 200 ký tự), replace tối đa 1000 ký tự")
	}
	if in.Find == in.Replace && !in.Preview {
		return nil, mcpFindReplaceOut{}, errors.New("find và replace giống nhau")
	}
	out := mcpFindReplaceOut{mcpEditOut: mcpEditOut{Slug: info.Slug, Title: info.Title}, Hits: []mcpHit{}}
	var edits []SectionEdit
	for _, s := range info.Sections {
		if s.Intro {
			continue // lời mở đầu tự tạo từ tên sách, sửa qua update_book_info
		}
		n := strings.Count(s.Text, in.Find)
		if in.IncludeTitles {
			n += strings.Count(s.Title, in.Find)
		}
		if n == 0 {
			continue
		}
		out.Matches += n
		ex := excerpt(s.Text, in.Find)
		if ex == "" {
			ex = s.Title
		}
		out.Hits = append(out.Hits, mcpHit{Index: s.Index + 1, Title: s.Title, Count: n, Excerpt: ex})
		title := s.Title
		if in.IncludeTitles {
			title = strings.ReplaceAll(title, in.Find, in.Replace)
		}
		edits = append(edits, SectionEdit{Index: s.Index, Title: title, Text: strings.ReplaceAll(s.Text, in.Find, in.Replace)})
		out.Sections = append(out.Sections, s.Index+1)
	}
	switch {
	case out.Matches == 0:
		out.Message = fmt.Sprintf("Không thấy %q trong lời đọc (khớp nguyên cụm, phân biệt hoa thường).", in.Find)
		return nil, out, nil
	case in.Preview:
		out.Message = fmt.Sprintf("Có %d chỗ trong %d mục. Chưa thay gì. Gọi lại với preview=false để thay và đọc lại các mục đó.", out.Matches, len(edits))
		return nil, out, nil
	}
	_, e, err := a.rereadWithSnapshot(info, edits, out.Sections, fmt.Sprintf("AI thay %q → %q (%d chỗ, %d mục)", in.Find, in.Replace, out.Matches, len(edits)))
	if err != nil {
		return nil, mcpFindReplaceOut{}, err
	}
	e.Message = fmt.Sprintf("Đã thay %d chỗ trong %d mục, đang đọc lại các mục đó. ", out.Matches, len(edits)) + e.Message
	out.mcpEditOut = e
	return nil, out, nil
}

// ── Thông tin, bìa, giọng ────────────────────────────────────────────────

type mcpInfoInput struct {
	Slug       string  `json:"slug" jsonschema:"slug của cuốn"`
	Title      *string `json:"title,omitempty" jsonschema:"tên sách mới"`
	Author     *string `json:"author,omitempty"`
	Translator *string `json:"translator,omitempty" jsonschema:"dịch giả"`
	Publisher  *string `json:"publisher,omitempty" jsonschema:"nhà xuất bản"`
	Category   *string `json:"category,omitempty" jsonschema:"danh mục trong thư viện; chuỗi rỗng = chưa phân loại"`
	Series     *string `json:"series,omitempty" jsonschema:"tên bộ sách; chuỗi rỗng = sách lẻ"`
	Volume     *int    `json:"volume,omitempty" jsonschema:"số tập trong bộ"`
}

func (a *App) mcpUpdateInfo(_ context.Context, _ *mcp.CallToolRequest, in mcpInfoInput) (*mcp.CallToolResult, mcpEditOut, error) {
	info, err := a.mcpBook(in.Slug)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	next := library.Info{Title: info.Title, Author: info.Author, Translator: info.Translator, Publisher: info.Publisher,
		Category: info.Category, Series: info.Series, Volume: info.Volume}
	set := func(dst *string, v *string) {
		if v != nil {
			*dst = strings.TrimSpace(*v)
		}
	}
	set(&next.Title, in.Title)
	set(&next.Author, in.Author)
	set(&next.Translator, in.Translator)
	set(&next.Publisher, in.Publisher)
	set(&next.Category, in.Category)
	set(&next.Series, in.Series)
	if in.Volume != nil {
		next.Volume = *in.Volume
	}
	snap, err := a.snapshot(info.Slug, "AI sửa thông tin sách")
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	res, err := a.SaveBookInfo(info.Slug, next)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	out := mcpEditOut{Slug: info.Slug, Title: res.Book.Title, snapshot: snap, Message: "Đã lưu thông tin sách."}
	if res.CoverRedrawn {
		out.Message += " Bìa tự vẽ đã vẽ lại theo tên mới."
	}
	if res.IntroEdit != nil {
		if _, err := a.StartReread(info.Slug, []SectionEdit{*res.IntroEdit}); err == nil {
			out.Reread, out.Sections = true, []int{1}
			out.Message += " Đang đọc lại lời giới thiệu ở đầu sách theo tên mới."
		} else {
			out.Message += " Chưa đọc lại được lời giới thiệu (" + err.Error() + ")."
		}
	}
	return nil, out, nil
}

type mcpCoverInput struct {
	Slug        string `json:"slug" jsonschema:"slug của cuốn"`
	ImageBase64 string `json:"image_base64,omitempty" jsonschema:"ảnh bìa jpg / png / webp mã hoá base64 (tối đa 10 MB)"`
	Auto        bool   `json:"auto,omitempty" jsonschema:"true = dùng bìa Sano tự vẽ theo tên sách"`
}

func (a *App) mcpSetCover(_ context.Context, _ *mcp.CallToolRequest, in mcpCoverInput) (*mcp.CallToolResult, mcpEditOut, error) {
	info, err := a.mcpBook(in.Slug)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	var tmp string
	if !in.Auto {
		raw := strings.TrimSpace(in.ImageBase64)
		if i := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && i > 0 {
			raw = raw[i+1:]
		}
		if raw == "" {
			return nil, mcpEditOut{}, errors.New("gửi image_base64 hoặc auto=true")
		}
		if base64.StdEncoding.DecodedLen(len(raw)) > maxCoverBytes+3 {
			return nil, mcpEditOut{}, errors.New("ảnh bìa quá lớn (tối đa 10 MB)")
		}
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, mcpEditOut{}, errors.New("image_base64 không đúng dạng base64")
		}
		ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[http.DetectContentType(data)]
		if ext == "" {
			return nil, mcpEditOut{}, errors.New("ảnh bìa phải là jpg, png hoặc webp")
		}
		dir := filepath.Join(a.lib.Root(), ".tam", "mcp")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, mcpEditOut{}, err
		}
		tmp = filepath.Join(dir, "bia-"+randomID()+ext)
		if err := os.WriteFile(tmp, data, 0o600); err != nil {
			return nil, mcpEditOut{}, err
		}
		defer func() { _ = os.Remove(tmp) }()
	}
	snap, err := a.snapshot(info.Slug, "AI đổi bìa")
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	if in.Auto {
		_, err = a.UseAutoCover(info.Slug)
	} else {
		_, err = a.SetBookCover(info.Slug, tmp)
	}
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	return nil, mcpEditOut{Slug: info.Slug, Title: info.Title, snapshot: snap, Message: "Đã đổi bìa."}, nil
}

type mcpVoiceInput struct {
	Slug  string `json:"slug" jsonschema:"slug của cuốn"`
	Voice string `json:"voice" jsonschema:"tên giọng mới (từ list_voices)"`
}

func (a *App) mcpChangeVoice(_ context.Context, _ *mcp.CallToolRequest, in mcpVoiceInput) (*mcp.CallToolResult, mcpEditOut, error) {
	info, err := a.mcpBook(in.Slug)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	voice := strings.TrimSpace(in.Voice)
	vs, err := a.Voices()
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	ok := false
	for _, v := range vs {
		ok = ok || v.Name == voice
	}
	if !ok {
		return nil, mcpEditOut{}, fmt.Errorf("không có giọng %q: gọi list_voices để xem tên đúng", voice)
	}
	snap, err := a.snapshot(info.Slug, "AI đổi giọng sang "+voice)
	if err != nil {
		return nil, mcpEditOut{}, err
	}
	if _, err := a.StartVoiceChange(info.Slug, voice); err != nil {
		return nil, mcpEditOut{}, err
	}
	return nil, mcpEditOut{Slug: info.Slug, Title: info.Title, Reread: true, snapshot: snap,
		Message: fmt.Sprintf("Đang đọc lại cả cuốn bằng giọng %s (%d mục). Xong hết mới thay giọng cũ. Theo dõi bằng get_edit_status.", voice, len(info.Sections))}, nil
}

// ── Cách đọc một từ ──────────────────────────────────────────────────────

type mcpPronounceInput struct {
	Word    string `json:"word" jsonschema:"từ / chữ viết tắt, ví dụ SePay, KPI"`
	Reading string `json:"reading" jsonschema:"cách đọc viết bằng chữ, ví dụ Xi Pây, ca pê i"`
	Slug    string `json:"slug,omitempty" jsonschema:"chỉ áp cho cuốn này; bỏ trống = từ điển chung (áp cho sách tạo sau)"`
}

type mcpPronounceOut struct {
	Message string `json:"message"`
	Hits    []int  `json:"sections_with_word,omitempty"` // mục có từ này (cần đọc lại mới nghe thấy cách đọc mới)
}

func (a *App) mcpSetPronunciation(_ context.Context, _ *mcp.CallToolRequest, in mcpPronounceInput) (*mcp.CallToolResult, mcpPronounceOut, error) {
	word, reading := strings.TrimSpace(in.Word), strings.TrimSpace(in.Reading)
	if word == "" || reading == "" || utf8.RuneCountInString(word) > 60 || utf8.RuneCountInString(reading) > 200 {
		return nil, mcpPronounceOut{}, errors.New("word, reading phải có chữ (word tối đa 60, reading tối đa 200 ký tự)")
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		if err := a.SetGlobalPronunciation(word, reading); err != nil {
			return nil, mcpPronounceOut{}, err
		}
		return nil, mcpPronounceOut{Message: fmt.Sprintf("Đã thêm vào từ điển chung: %s đọc là %q. Áp cho sách tạo sau; sách đã có muốn đổi thì đọc lại mục có từ này.", word, reading)}, nil
	}
	info, err := a.lib.Edit(slug)
	if err != nil {
		return nil, mcpPronounceOut{}, mcpNotFound(slug, err)
	}
	if err := a.SetBookPronunciation(slug, word, reading); err != nil {
		return nil, mcpPronounceOut{}, err
	}
	out := mcpPronounceOut{}
	for _, s := range info.Sections {
		if strings.Contains(s.Text, word) || strings.Contains(s.Title, word) {
			out.Hits = append(out.Hits, s.Index+1)
		}
	}
	out.Message = fmt.Sprintf("Đã lưu cách đọc %s là %q cho cuốn %q.", word, reading, info.Title)
	if len(out.Hits) > 0 {
		out.Message += fmt.Sprintf(" Có %d mục chứa từ này; muốn nghe cách đọc mới thì gọi update_sections với lời hiện tại của các mục đó để đọc lại.", len(out.Hits))
	}
	return nil, out, nil
}

// ── Tiến độ sửa ──────────────────────────────────────────────────────────

type mcpEditStatusOut struct {
	Status    string  `json:"status"` // none | rereading | voice_change | done | error | cancelled
	Title     string  `json:"title,omitempty"`
	Slug      string  `json:"slug,omitempty"`
	Done      int     `json:"sections_done"`
	Total     int     `json:"sections_total"`
	Percent   float64 `json:"percent"`
	RemainMin int     `json:"remain_min,omitempty"`
	Error     string  `json:"error,omitempty"`
}

func (a *App) mcpEditStatus(context.Context, *mcp.CallToolRequest, mcpNoInput) (*mcp.CallToolResult, mcpEditStatusOut, error) {
	st := a.EditStatus()
	if st == nil {
		return nil, mcpEditStatusOut{Status: "none"}, nil
	}
	out := mcpEditStatusOut{Title: st.BookTitle, Slug: st.Slug, Done: st.Finished, Total: st.Total,
		Percent: progressPercent(st.Finished, st.Total), Error: st.Error}
	if st.Remaining > 0 {
		out.RemainMin = max(1, st.Remaining/60)
	}
	switch {
	case st.Running && st.Kind == editKindVoice:
		out.Status = "voice_change"
	case st.Running:
		out.Status = "rereading"
	case st.Cancelled:
		out.Status = "cancelled"
	case st.Error != "":
		out.Status = "error"
	default:
		out.Status = "done"
	}
	return nil, out, nil
}

// ── Hoàn tác (chỉ người dùng, ở màn MCP) ────────────────────────────────

// MCPUndo đưa cuốn slug về bản cũ id (bản giữ lại trước khi AI sửa).
func (a *App) MCPUndo(slug, id string) (*MCPInfo, error) {
	if st := a.editing(); st != nil && st.Slug == slug {
		return nil, errors.New("cuốn này đang được đọc lại / đổi giọng, đợi xong rồi hoàn tác")
	}
	if info, err := a.lib.Edit(slug); err == nil && info.VoiceJob != nil {
		return nil, errors.New("cuốn này đang đổi giọng dở, đợi xong rồi hoàn tác")
	}
	var snap *library.Snapshot
	for _, s := range a.lib.Snapshots(slug) {
		if s.ID == id {
			s := s
			snap = &s
		}
	}
	if snap == nil {
		return nil, library.ErrSnapshotGone
	}
	if err := a.lib.RestoreSnapshot(slug, id); err != nil {
		return nil, err
	}
	a.logMCP(MCPLogEntry{Client: "Bạn", Tool: "undo", Text: "Hoàn tác: " + snap.Label + " · " + quoteTitle(snap.Title, slug), Slug: slug, Undo: id})
	return a.MCPInfo(), nil
}
