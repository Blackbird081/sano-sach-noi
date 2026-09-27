---
title: Hành trình nghe — thống kê thời gian nghe sách nói
description: 'Xem bạn đã nghe sách nói bao lâu, chuỗi ngày nghe, sách đã nghe xong, giờ hay nghe và đặt mục tiêu nghe mỗi ngày trong Sano. Số liệu chỉ lưu trên máy.'
---

# Hành trình nghe (từ 0.1.14)

Bấm **Hành trình nghe** ở thanh bên để xem thói quen nghe của bạn. Sano ghi số giây nghe thật mỗi ngày: phát bao lâu tính bấy lâu, tua hay nhảy chương không tính. Số liệu bắt đầu có từ bản 0.1.14.

## Xem theo tuần, tháng, năm

Mặc định Sano mở **Tháng này**, lần sau nhớ khoảng bạn chọn. Có thể đổi sang tuần, năm (chọn được năm cũ có số liệu) hoặc tất cả. Mỗi khoảng có:

- **Thời gian nghe**, so với kỳ trước.
- **Chuỗi ngày nghe**: số ngày liền nhau có nghe, kèm kỷ lục.
- **Sách nghe xong** trong khoảng đó.
- **Tiết kiệm nhờ nghe nhanh**: thời gian bớt được khi nghe ở tốc độ trên 1×.
- **Nghe nhiều nhất**, **Nghe xong gần đây**, **Giờ hay nghe** và **Lịch nghe 6 tháng**.

## Mục tiêu nghe mỗi ngày

Bấm **Đặt mục tiêu**, chọn số phút mỗi ngày, ví dụ 15 phút. Trang Hành trình nghe hiện tiến độ hôm nay và báo khi đã đạt. Mục tiêu chỉ để tự nhắc mình, Sano không gửi thông báo. Muốn bỏ thì bấm **Tắt mục tiêu**.

## Dấu đã nghe trong mục lục

Trong trình phát, mục nào bạn nghe thật từ 85% trở lên mới có dấu ✓. Nhảy qua một mục không tính là đã nghe.

## Số liệu lưu ở đâu, xoá thế nào

Số liệu nghe chỉ nằm trên máy bạn, trong file `~/Sano/.nghe.json`, không gửi đi đâu.

- **Xoá số liệu**: vào **Cài đặt** → **Dữ liệu nghe** → **Xoá số liệu**. File số liệu được chuyển vào Thùng rác (lấy lại được nếu lỡ tay). Sách và vị trí đang nghe vẫn giữ nguyên.
- **Xoá lịch sử nghe** ở hàng **Nghe tiếp** trong Thư viện chỉ bỏ vị trí đang nghe dở của các cuốn, không xoá số liệu Hành trình nghe. Xem [Thư viện](./thu-vien#nghe-tiep-va-tat-ca-sach).
