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
