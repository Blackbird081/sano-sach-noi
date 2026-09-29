// Cam kết trước khi render (wireframe D19): người dùng tick từng ô mới render được. Nội dung
// khớp Điều khoản sử dụng (docs/dieu-khoan-su-dung.md, phiên bản 2).
import { FileCheck2, Scale, Share2, ShieldAlert, UserX } from 'lucide-vue-next'

export const PLEDGES = [
  {
    icon: FileCheck2,
    title: 'Tôi có quyền dùng tài liệu này',
    desc: 'Tài liệu do tôi viết, tác phẩm đã hết thời hạn bảo hộ quyền tác giả, hoặc tôi đã được tác giả / chủ sở hữu cho phép bằng văn bản chuyển thành sách nói.',
  },
  {
    icon: Scale,
    title: 'Nội dung không vi phạm pháp luật Việt Nam',
    desc: 'Không chống phá Nhà nước, xuyên tạc lịch sử; không kích động bạo lực, thù hằn dân tộc, tôn giáo; không đồi truỵ, mê tín, tin giả; không xúc phạm danh dự người khác; không chứa bí mật nhà nước hay thông tin cá nhân của người khác.',
  },
  {
    icon: UserX,
    title: 'Không mạo danh, không lừa đảo',
    desc: 'Không dùng giọng đọc để giả giọng hay mạo danh người thật, cơ quan, tổ chức; không dùng sách nói, video, ảnh tạo ra để lừa đảo.',
  },
  {
    icon: Share2,
    title: 'Không phát tán tác phẩm của người khác',
    desc: 'Không đăng công khai, chia sẻ cho người ngoài, bán hay dùng vào mục đích thương mại sách nói, video, ảnh làm từ tác phẩm của người khác khi chưa có sự đồng ý bằng văn bản của chủ sở hữu.',
  },
  {
    icon: ShieldAlert,
    title: 'Tôi tự chịu hoàn toàn trách nhiệm',
    desc: 'Tôi tự chịu trách nhiệm trước pháp luật về nội dung đưa vào và mọi thứ tạo ra từ Sano. Tác giả, người đóng góp và nhà tài trợ của Sano không liên quan và không chịu trách nhiệm; nếu họ bị khiếu nại vì việc làm của tôi, tôi chịu mọi chi phí và thiệt hại phát sinh.',
  },
]
