// Prompt gửi cho AI (Claude, ChatGPT, Gemini bản web) cùng file Word, ở bước "Cách đọc".
// Một nguồn duy nhất: docs/prompts/ (hướng dẫn dùng ở docs/lam-muot-tai-lieu.md).
//   - cấp 2 Làm mượt: loi-nhac-mau.txt (giữ tên file cũ vì docs đã dẫn link)
//   - cấp 3 Viết lại thành văn sách nói: viet-lai-sach-noi.txt, soát lại: soat-lai-sach-noi.txt
import smooth from '../../../../docs/prompts/loi-nhac-mau.txt?raw'
import rewrite from '../../../../docs/prompts/viet-lai-sach-noi.txt?raw'
import review from '../../../../docs/prompts/soat-lai-sach-noi.txt?raw'

export const SAMPLE_PROMPT = smooth.trim()
export const REWRITE_PROMPT = rewrite.trim()
export const REVIEW_PROMPT = review.trim()

export type AITool = 'claude' | 'chatgpt' | 'gemini'

// Gemini bản miễn phí không tạo được file Word: hứa tạo file rồi dừng, hoặc trộn
// lời chào vào nội dung (đã thử 27/09). Bản cho Gemini thay phần "Cách trả kết
// quả" bằng yêu cầu trả khối mã ngay từ đầu, để dán vào Sano.
const TEXT_OUTPUT = `Cách trả kết quả:
- Trả nội dung sách trong MỘT khối mã (code block) để tôi sao chép. Không tạo file, không hỏi lại.
- Dòng tên sách mở đầu bằng "% ", dòng chương mở đầu bằng "# ", dòng mục mở đầu bằng "## ". Các dòng khác là đoạn văn thường.
- Trong khối mã chỉ có nội dung sách để đọc to. Danh sách ý, báo cáo, lời chào ghi bên ngoài khối mã.
- Không bảng, không gạch đầu dòng, không in đậm. Tên sách và tiêu đề không viết IN HOA toàn bộ.
- Nếu tài liệu dài quá một lần trả lời, làm lần lượt từng chương, mỗi phần một khối mã. Hết mỗi phần thì dừng và chờ tôi gõ "tiếp".`

function forText(p: string): string {
  const i = p.indexOf('Cách trả kết quả:')
  return i < 0 ? p : `${p.slice(0, i)}${TEXT_OUTPUT}`
}

/** Prompt cho cấp (2 làm mượt, 3 viết lại) và AI đang dùng. */
export function promptFor(level: number, tool: AITool): string {
  const p = level === 2 ? SAMPLE_PROMPT : REWRITE_PROMPT
  return tool === 'gemini' ? forText(p) : p
}

export function reviewPromptFor(tool: AITool): string {
  return tool === 'gemini' ? forText(REVIEW_PROMPT) : REVIEW_PROMPT
}

export const AI_TOOLS: { k: AITool; name: string; url: string }[] = [
  { k: 'claude', name: 'Claude', url: 'https://claude.ai/new' },
  { k: 'chatgpt', name: 'ChatGPT', url: 'https://chatgpt.com/' },
  { k: 'gemini', name: 'Gemini', url: 'https://gemini.google.com/app' },
]
