# Skill làm sách nói cho AI

Skill giúp Claude hoặc ChatGPT biến tài liệu của bạn thành bản đọc cho sách nói, rồi trả về file Word có sẵn chương, mục để nạp vào [Sano](https://sanobook.com). Có ba việc:

- **Cấp độ 3:** viết lại thành văn sách nói, như người kể: chuyện trước, lý thuyết sau, chương ngắn, cuối chương có ba ý cần nhớ.
- **Cấp độ 2:** làm mượt, giữ nguyên ý: đổi bảng, hình, danh sách, chữ viết tắt thành lời.
- **Soát lại:** đối chiếu bản viết lại với bản gốc, bổ sung chỗ mất ý, sửa chỗ tự thêm.

Hướng dẫn đầy đủ, có ảnh: [Ba cách đọc](https://sanobook.com/lam-muot-tai-lieu).

## Tải về

| File | Dùng cho |
|---|---|
| [`sano-sach-noi.zip`](https://sanobook.com/skill/sano-sach-noi.zip) (thư mục [`sano-sach-noi/`](sano-sach-noi/SKILL.md)) | Claude (mọi gói), ChatGPT có mục Skills, Claude Code |
| [`sano-huong-dan-ai.txt`](sano-huong-dan-ai.txt) | Dự án (Project) của ChatGPT chưa có mục Skills |

Trong phần mềm Sano: bước **Tạo sách nói** → **Cách đọc** → chọn cấp 2 hoặc 3 → **Nạp skill cho Claude / ChatGPT** cũng tải được hai file này.

## Cách nạp

- **Claude (claude.ai):** **Customize** → **Skills** → nút **+** → **Create skill** → **Upload a skill**, chọn `sano-sach-noi.zip`.
- **ChatGPT có mục Skills** (gói Business, Enterprise, Edu): **Skills** → **Tạo** → **Tải lên**, chọn `sano-sach-noi.zip`.
- **ChatGPT dùng Dự án:** tạo Dự án tên "Sano – sách nói", thêm `sano-huong-dan-ai.txt` vào phần Tệp, dán câu này vào ô Hướng dẫn (Instructions):

  > Mỗi khi tôi gửi tài liệu để làm sách nói, làm đúng theo file sano-huong-dan-ai.txt đã đính kèm. Mặc định làm việc A, cấp 3 viết lại thành văn sách nói. Tôi ghi "cấp 2" thì làm việc B, làm mượt.

- **Claude Code:** chép thư mục `sano-sach-noi/` vào `~/.claude/skills/`.

Gemini chưa hỗ trợ skill: dùng prompt trong Sano (nút **Sao chép prompt cho Gemini**), mỗi lần dán một lần.

## Câu gõ mỗi lần làm sách

Mở cuộc trò chuyện mới, đính kèm file Word rồi gõ:

| Việc | Claude, ChatGPT có mục Skills | ChatGPT dùng Dự án |
|---|---|---|
| Viết lại thành văn sách nói | Làm file sách nói dùng skill sano-sach-noi (cấp độ 3) | Làm file sách nói theo file sano-huong-dan-ai.txt (cấp độ 3) |
| Làm mượt, giữ nguyên ý | Làm file sách nói dùng skill sano-sach-noi (cấp độ 2) | Làm file sách nói theo file sano-huong-dan-ai.txt (cấp độ 2) |
| Soát lại (gửi kèm bản gốc và bản viết lại) | Soát lại file sách nói dùng skill sano-sach-noi | Soát lại file sách nói theo file sano-huong-dan-ai.txt |

Tải file Word AI trả về, nạp vào Sano ở bước **Nạp file**.

## Cho người đóng góp

Các file ở đây được dựng từ prompt trong [`docs/prompts/`](../docs/prompts/), cùng nguồn với nút **Sao chép prompt** trong phần mềm. Sửa prompt ở đó, rồi chạy từ gốc repo:

```bash
go run ./cmd/sano-skill
```

Test `go test ./docs/prompts/` báo lỗi nếu file trong thư mục này lệch với prompt.
