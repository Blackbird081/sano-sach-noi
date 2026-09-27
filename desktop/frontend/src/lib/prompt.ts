// Prompt gửi cho AI (Claude, ChatGPT, Gemini bản web) cùng file Word, ở bước "Cách đọc".
// Một nguồn duy nhất: docs/prompts/ (hướng dẫn dùng ở docs/lam-muot-tai-lieu.md).
//   - cấp 2 Làm mượt: loi-nhac-mau.txt (giữ tên file cũ vì docs đã dẫn link)
//   - cấp 3 Viết lại thành văn sách nói: viet-lai-sach-noi.txt
// Prompt soát lại (soat-lai-sach-noi.txt) không có nút trong app (wireframe D8b): prompt
// chính đã dặn AI tự soát; soát lại chỉ nằm trong skill và trang hướng dẫn.
import smooth from '../../../../docs/prompts/loi-nhac-mau.txt?raw'
import rewrite from '../../../../docs/prompts/viet-lai-sach-noi.txt?raw'

export const SAMPLE_PROMPT = smooth.trim()
export const REWRITE_PROMPT = rewrite.trim()

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

export const AI_TOOLS: { k: AITool; name: string; url: string }[] = [
  { k: 'claude', name: 'Claude', url: 'https://claude.ai/new' },
  { k: 'chatgpt', name: 'ChatGPT', url: 'https://chatgpt.com/' },
  { k: 'gemini', name: 'Gemini', url: 'https://gemini.google.com/app' },
]


// Hộp "Nạp skill" (wireframe D8b), tên file khớp docs/prompts/prompts.go.
// Claude: zip SKILL.md. ChatGPT: Skills chỉ có ở gói Business/Enterprise (dùng được
// luôn zip của Claude), nên đường chính là Dự án + file hướng dẫn + câu ngắn dưới
// đây trong ô Instructions (tối đa 8.000 ký tự, không vừa nguyên hướng dẫn).
// Gemini không có hộp: tin rò rỉ Google bỏ Gem từ 13/10/2026, Skills của Gemini chỉ gói trả phí.
export const SKILL_ZIP = 'sano-sach-noi.zip'
export const GUIDE_TXT = 'sano-huong-dan-ai.txt'

export function shortInstruction(): string {
  return `Mỗi khi tôi gửi tài liệu để làm sách nói, làm đúng theo file ${GUIDE_TXT} đã đính kèm. Mặc định làm việc A, cấp 3 viết lại thành văn sách nói. Tôi ghi "cấp 2" thì làm việc B, làm mượt.`
}

export type ChatGPTPlan = 'skills' | 'project'
export const CHATGPT_PLANS: { k: ChatGPTPlan; title: string; sub: string }[] = [
  { k: 'skills', title: 'Có mục Skills', sub: 'Gói Business, Enterprise, Edu' },
  { k: 'project', title: 'Chưa có mục Skills', sub: 'Gói Free, Plus, Pro · dùng Dự án' },
]

export interface SkillStep {
  text: string
  act: 'file' | 'open' | 'copy'
  file?: string
  label?: string
  url?: string
}
export const SKILL_STEPS: Record<'claude' | ChatGPTPlan, SkillStep[]> = {
  claude: [
    { text: 'Tải file skill về, không cần giải nén.', act: 'file', file: SKILL_ZIP },
    { text: 'Mở Customize → Skills → nút + → Create skill → Upload a skill, chọn file vừa tải.', act: 'open', label: 'Mở Skills của Claude', url: 'https://claude.ai/customize/skills' },
  ],
  skills: [
    { text: 'Tải file skill về, không cần giải nén.', act: 'file', file: SKILL_ZIP },
    { text: 'Mở ChatGPT → Skills → Tạo → Tải lên, chọn file vừa tải.', act: 'open', label: 'Mở ChatGPT', url: 'https://chatgpt.com/' },
  ],
  project: [
    { text: 'Tải file hướng dẫn về.', act: 'file', file: GUIDE_TXT },
    { text: 'Tạo Dự án (Project) tên "Sano – sách nói", thêm file vừa tải vào phần Tệp.', act: 'open', label: 'Mở ChatGPT', url: 'https://chatgpt.com/' },
    { text: 'Dán câu Sano đưa vào ô Hướng dẫn (Instructions) của dự án. Từ nay làm sách trong dự án đó.', act: 'copy' },
  ],
}
