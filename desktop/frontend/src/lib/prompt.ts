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

// Hộp "Nạp skill cho AI": Claude nạp zip skill; ChatGPT / Gemini đính kèm file
// hướng dẫn vào Project / Gem và dán câu ngắn này vào ô Instructions (ô của
// ChatGPT tối đa 8.000 ký tự, không vừa nguyên hướng dẫn).
const SHORT = 'Mỗi khi tôi gửi tài liệu để làm sách nói, làm đúng theo file sano-huong-dan-ai.txt đã đính kèm. Mặc định làm việc A, cấp 3 viết lại thành văn sách nói. Tôi ghi "cấp 2" thì làm việc B, làm mượt.'
export function shortInstruction(tool: AITool): string {
  return tool === 'gemini' ? `${SHORT} Bạn không tạo được file Word, nên luôn trả nội dung sách trong một khối mã như file hướng dẫn dặn.` : SHORT
}

export const SKILL_STEPS: Record<AITool, { file: string; steps: string[]; note: string }> = {
  claude: {
    file: 'sano-sach-noi.zip',
    steps: [
      'Tải file sano-sach-noi.zip về, không cần giải nén.',
      'Mở claude.ai → Cài đặt (Settings) → Capabilities → Skills → Tải lên (Upload skill), chọn file vừa tải.',
      'Từ nay chỉ cần đính kèm file Word và gõ "làm sách nói cấp 3" (hoặc "cấp 2").',
    ],
    note: 'Skill cần gói Claude trả phí (Pro trở lên) và bật chạy code (Code execution). Skill cũng dùng được trong Claude Code.',
  },
  chatgpt: {
    file: 'sano-huong-dan-ai.txt',
    steps: [
      'Tải file sano-huong-dan-ai.txt về.',
      'Mở ChatGPT → tạo Dự án (Project) mới tên "Sano – sách nói", thêm file vừa tải vào phần Tệp (Files) của dự án.',
      'Mở phần Hướng dẫn (Instructions) của dự án, dán câu dưới đây.',
      'Từ nay mở trò chuyện trong dự án đó, đính kèm file Word và gõ "cấp 3" (hoặc "cấp 2").',
    ],
    note: 'Tên các mục trong ChatGPT có thể đổi theo phiên bản; trang hướng dẫn luôn cập nhật cách làm mới nhất.',
  },
  gemini: {
    file: 'sano-huong-dan-ai.txt',
    steps: [
      'Tải file sano-huong-dan-ai.txt về.',
      'Mở gemini.google.com → Gem → Gem mới, đặt tên "Sano – sách nói", thêm file vừa tải ở mục Kiến thức (Knowledge).',
      'Dán câu dưới đây vào ô Hướng dẫn (Instructions) rồi Lưu.',
      'Từ nay mở Gem đó, đính kèm file Word và gõ "cấp 3" (hoặc "cấp 2"). Sao chép khối mã Gem trả về, dán vào Sano.',
    ],
    note: 'Gem dùng được cả với tài khoản Google miễn phí. Tên các mục có thể đổi theo phiên bản; trang hướng dẫn luôn cập nhật cách làm mới nhất.',
  },
}
