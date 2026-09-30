---
title: Kết nối AI (MCP)
description: 'Cho Claude Desktop, Claude Code, Codex, Cursor xem, tạo, sửa sách nói trong Sano bằng lời nói thường, qua chuẩn MCP. Chạy trên máy bạn, không có lệnh xoá sách.'
---

# Kết nối AI (MCP)

Từ bản 0.1.21, Sano nói chuyện được với trợ lý AI trên máy bạn: **Claude Desktop, Claude Code, Codex, Cursor**… qua chuẩn **MCP** (Model Context Protocol). Bạn chỉ cần nói bằng lời thường, AI tự làm trong Sano:

- "Trong Sano có những sách nào?"
- "Cuốn Kênh phân phối tập 2 có mục nào nói về đại lý? Tóm tắt giúp tôi."
- "Viết giúp tôi cuốn sách nói ngắn về quản lý dòng tiền, 3 chương, rồi tạo luôn."
- "Trong cuốn X, thay chữ KPI thành chỉ số KPI, cho tôi xem trước."
- "Đổi giọng cuốn Y sang Trúc Ly."

Việc tạo sách (render) vẫn chạy **trên máy bạn**, bằng bộ đọc của Sano. **Sano phải đang mở** thì AI mới làm được.

## Kết nối (làm một lần)

Mở Sano → mục **MCP** ở thanh bên (ngay trên Cài đặt). **Kết nối trong máy** bật sẵn. Chọn phần mềm AI bạn dùng:

| Phần mềm | Cách thêm |
|---|---|
| **Claude Desktop** | Bấm **Thêm vào Claude Desktop**. Sano thêm mình vào cấu hình của Claude (giữ nguyên cấu hình khác, có sao lưu). Thoát hẳn Claude Desktop (⌘Q trên Mac) rồi mở lại. |
| **Claude Code** | Bấm **Chép** ở thẻ Claude Code, dán vào Terminal, chạy một lần. Lệnh dạng `claude mcp add sano -- <đường dẫn sano-mcp>`. |
| **Codex** | Bấm **Chép** ở thẻ Codex, dán vào Terminal, chạy một lần. Lệnh dạng `codex mcp add sano -- <đường dẫn sano-mcp>`. |
| **Phần mềm khác** (Cursor, Windsurf…) | Chép khối cấu hình ở thẻ **Phần mềm khác**, dán vào phần cấu hình MCP của phần mềm (kiểu stdio). |

Xong, hỏi thử AI: *"Trong Sano có những sách nào?"*

::: tip Lệnh khác nhau theo máy
Đường dẫn cầu nối khác nhau trên mỗi máy (Mac: bên trong `Sano.app`; Windows: thư mục cài; Linux: chính file `.AppImage`). Luôn chép lệnh từ mục MCP trong app, đừng gõ tay. Chuyển Sano sang chỗ khác (ví dụ vào thư mục Applications) thì chạy lại lệnh mới.
:::

## Tạo sách qua AI: tạo trước, cam kết sau

1. AI gửi nội dung sách (Tên sách, Chương, Mục) cho Sano, cho bạn xem mục lục và thời lượng.
2. Bạn đồng ý, AI xin tạo. Sano **tạo ngay**, nhiều cuốn thì tạo lần lượt. Bạn không cần ngồi chờ.
3. Tạo xong, sách nằm ở **khu chờ**: chưa vào Thư viện, chưa nghe, xuất hay chia sẻ được.
4. Sano mở popup **cam kết để lưu vào Thư viện**: tick cuốn muốn lưu, tick 6 ô cam kết **một lần** cho các cuốn đã tick. Cuốn đang tạo mà đã tick thì tự lưu khi xong.
5. Chưa tiện? Bấm **Để sau**, sách vẫn ở khu chờ. Mở lại bằng nút *"… cuốn chờ cam kết"* ở thanh bên. Không cam kết thì Sano tự xoá bản ở khu chờ sau **7 ngày**.

Mỗi cuốn được ghi thời điểm bạn cam kết, như khi bạn tự tạo sách trong app. Xem [Điều khoản sử dụng](./dieu-khoan-su-dung).

::: tip Để AI viết hay hơn
Nạp thêm [skill sano-sach-noi](./skill-ai) cho Claude: AI sẽ viết theo văn sách nói (câu ngắn, số và chữ viết tắt viết thành lời, cuối chương có Ba ý cần nhớ) trước khi gửi sang Sano.
:::

## Sửa sách qua AI, hoàn tác được

AI sửa được lời từng mục (rồi đọc lại đúng các mục đó), tìm và thay cả cuốn, tên sách, tác giả, danh mục, bộ sách, bìa, giọng đọc, cách đọc một từ (ví dụ *SePay → Xi Pây*).

Trước mỗi lần AI sửa một cuốn, Sano **giữ bản trước đó**. Vào mục **MCP → Việc AI làm gần đây**, dòng sửa mới nhất của mỗi cuốn có nút **Hoàn tác** (bấm hai lần để chắc). Sano giữ 3 bản gần nhất của mỗi cuốn trong 7 ngày.

## An toàn

- **Không có lệnh xoá.** AI không xoá được sách hay file nào. Xoá sách chỉ làm trong app.
- **Chỉ trong thư viện Sano.** AI gửi nội dung chứ không gửi đường dẫn file, không đọc hay ghi file nào khác trên máy.
- **Chỉ trên máy này.** Kết nối chỉ nhận từ chính máy (127.0.0.1), cần mã bí mật lưu trong `~/Sano/.mcp` mà chỉ tài khoản của bạn đọc được, và chặn yêu cầu từ trang web.
- **Bạn giữ quyền.** Ở mục MCP có thể tắt hẳn kết nối, hoặc bỏ quyền **Tạo và sửa sách** (AI chỉ còn xem).
- **Có nhật ký.** Mọi việc AI làm được ghi ở mục MCP, chỉ lưu trên máy, giữ 7 ngày.

## Dùng từ điện thoại, claude.ai, ChatGPT?

Các trợ lý chạy trên web và điện thoại chỉ kết nối được qua internet. **Kết nối từ xa** đang được làm, sẽ có ở bản sau và mặc định tắt. Hiện tại bạn vẫn dùng [skill sano-sach-noi](./skill-ai) để AI trả về file Word, rồi nạp file vào Sano.

## Gặp lỗi?

| AI báo | Cách xử lý |
|---|---|
| *"app Sano chưa mở trên máy này"* | Mở Sano rồi hỏi lại. |
| *"người dùng đã tắt Kết nối trong máy"* | Vào mục MCP, bật **Kết nối trong máy**. |
| *"người dùng đã tắt quyền Tạo và sửa sách"* | Vào mục MCP, tick **Tạo và sửa sách**. |
| Claude Desktop không thấy Sano | Thoát hẳn Claude Desktop (⌘Q / thoát ở khay hệ thống) rồi mở lại. Vẫn không được thì bấm **Thêm lại** ở mục MCP. |
| Đã chuyển Sano sang thư mục khác | Chạy lại lệnh / bấm Thêm lại ở mục MCP (đường dẫn cầu nối đã đổi). |
