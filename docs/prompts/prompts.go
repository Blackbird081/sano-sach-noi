// Package prompts nhúng các prompt gửi cho AI (một nguồn duy nhất với app và trang
// hướng dẫn) và dựng từ đó "skill" để nạp cho AI một lần, dùng mãi:
//   - Claude: sano-sach-noi.zip (thư mục sano-sach-noi/ chứa SKILL.md)
//   - ChatGPT / Gemini: sano-huong-dan-ai.txt, đính kèm vào Project / GPT / Gem
//     (ô Instructions của ChatGPT tối đa 8.000 ký tự, không vừa cả hai cấp).
package prompts

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed loi-nhac-mau.txt
var smooth string

//go:embed viet-lai-sach-noi.txt
var rewrite string

//go:embed soat-lai-sach-noi.txt
var review string

const (
	SkillZipName  = "sano-sach-noi.zip"
	GuideFileName = "sano-huong-dan-ai.txt"
)

// guideBody — phần chung của skill và file hướng dẫn: chọn việc theo lời người
// dùng rồi làm đúng prompt tương ứng.
func guideBody() string {
	var b strings.Builder
	b.WriteString(`Khi người dùng gửi một tài liệu để làm sách nói bằng Sano, chọn MỘT trong ba việc dưới đây:
- Mặc định, hoặc khi người dùng nói "cấp 3", "viết lại", "văn sách nói": làm VIỆC A.
- Khi người dùng nói "cấp 2", "làm mượt", "giữ nguyên văn": làm VIỆC B.
- Khi người dùng gửi cả bản gốc lẫn bản đã viết lại và nhờ "soát lại": làm VIỆC C.

Về cách trả kết quả: nếu bạn có công cụ tạo file .docx cho người dùng tải về thì trả file Word như từng việc yêu cầu. Nếu không có công cụ đó, đừng hứa tạo file: trả ngay nội dung sách trong một khối mã (code block), dòng tên sách mở đầu "% ", chương "# ", mục "## ", trong khối mã chỉ có nội dung sách.

`)
	for _, s := range []struct{ title, body string }{
		{"VIỆC A. Viết lại thành văn sách nói (cấp 3)", rewrite},
		{"VIỆC B. Làm mượt, giữ nguyên ý (cấp 2)", smooth},
		{"VIỆC C. Soát lại bản viết lại", review},
	} {
		b.WriteString("=== " + s.title + " ===\n\n")
		b.WriteString(strings.TrimSpace(s.body))
		b.WriteString("\n\n")
	}
	return b.String()
}

// GuideText — nội dung sano-huong-dan-ai.txt (ChatGPT, Gemini).
func GuideText() string {
	return "HƯỚNG DẪN LÀM BẢN ĐỌC CHO SÁCH NÓI SANO\n\n" + guideBody()
}

// SkillMD — SKILL.md của skill Claude.
func SkillMD() string {
	return `---
name: sano-sach-noi
description: Biên tập tài liệu thành bản đọc cho sách nói Sano. Làm mượt (cấp 2) hoặc viết lại thành văn sách nói (cấp 3), rồi trả file Word có Title, Heading 1, Heading 2 để nạp vào Sano. Dùng khi người dùng gửi tài liệu và nói làm sách nói, làm mượt để đọc, viết lại để nghe, hoặc soát lại bản sách nói.
---

# Làm bản đọc cho sách nói Sano

` + guideBody()
}

// SkillZip — file zip nạp vào claude.ai → Customize → Skills → Upload a skill.
func SkillZip() ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("sano-sach-noi/SKILL.md")
	if err != nil {
		return nil, fmt.Errorf("zip create SKILL.md: %w", err)
	}
	if _, err := fw.Write([]byte(SkillMD())); err != nil {
		return nil, fmt.Errorf("zip write SKILL.md: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("zip close: %w", err)
	}
	return buf.Bytes(), nil
}
