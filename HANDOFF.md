# HANDOFF — Sano (sách nói tiếng Việt)

Cập nhật: 28/09/2026 13:38 · Phiên bản đang phát hành: **0.1.17** (Latest trên GitHub Releases) · trên `main` có 2 sửa lỗi chưa phát hành (mục 3)

Phiên mới: đọc file này + `README.md` + `CHANGELOG.md` là đủ nắm trạng thái.

## 1. Dự án là gì

Phần mềm desktop (Wails: Go + Vue 3 + shadcn-vue + Tailwind) biến file Word thành sách nói tiếng Việt.
Giọng đọc VieNeu-TTS v3 Turbo chạy ngay trên máy, không cần API key. Chạy Windows, macOS, Linux.

| Thư mục | Nội dung |
|---|---|
| `internal/bookmaker/` | Lõi làm sách: đọc docx, chuẩn hoá lời đọc (tiếng Việt), từ điển cách đọc, gọi bộ đọc, đóng gói zip, sửa sách (`rework.go`), bìa (`coverfit.go`) |
| `desktop/` | App Wails. `maker.go` tạo sách, `rework.go` sửa sách, `dict.go` từ điển, `thumb.go` bìa thu nhỏ, `media.go` phục vụ mp3/bìa |
| `desktop/internal/library/` | Thư viện `~/Sano/Sach`: đọc/ghi metadata, gói zip, nhập gói, sửa sách, từ điển |
| `desktop/internal/setup/`, `tts/` | Cài bộ đọc VieNeu (uv + Python + mô hình), kiểm tra, gỡ |
| `desktop/frontend/src/` | Giao diện. `wireframes/` là wireframe đã duyệt (D1–D12), mở bằng `?wireframe=<tên>` ở bản dev |
| `scripts/tts/` | Script Python gọi VieNeu (`audio_gen_batch.py`), khoá phiên bản |
| `scripts/release/` | Build, đóng gói, ký. Phát hành = push tag `vX.Y.Z` → CI tạo **release nháp** → công khai tay |
| `docs/` | Trang hướng dẫn (VitePress, GitHub Pages) |

## 2. Chạy, test, phát hành

- Dev: `make desktop-dev` (wails dev, Vite :5390, trình duyệt gọi được Go ở :34115). Thử không đụng thư viện thật: `SANO_HOME=/thu/muc/tam wails dev`.
- Test: `go vet ./... && go test ./... -race` ở gốc **và** trong `desktop/`; frontend `cd desktop/frontend && npm run build`.
- QA giao diện: `agent-browser` CLI (xem quy tắc chung). Lưu ý: `agent-browser fill` không kích sự kiện input của Vue → đặt `value` + `dispatchEvent(new Event('input'))` bằng `eval`.
- Build RC tại máy: `scripts/release/build.sh 0.1.18-rc.1 darwin/arm64` → `desktop/build/bin/Sano.app` (wails dev dùng chung thư mục này, `-clean` xoá nó).
- Phát hành: cập nhật `VERSION`, `CHANGELOG.md` (ngắn, hộp cập nhật của bản cũ hiện đoạn này), link tải trong `README.md` → commit `chore: phát hành vX.Y.Z` → push `main` → `git tag -a vX.Y.Z` + push tag → chờ CI → `gh release edit vX.Y.Z --draft=false --latest`.
- VirusTotal luôn báo 2 phần mềm nhầm bản Windows (đã có từ 0.1.16, file Go chưa ký), không chặn phát hành.

## 3. Phiên 28/09 — 2 sửa lỗi trên `main`, CHƯA phát hành, CHƯA push

- `0eed810` **Mục lục màn nghe hiện tên chương**: `PlayerView.vue` chèn dòng tên chương (chữ nhỏ, in hoa) trước tiểu mục
  đầu mỗi chương theo cờ `chapterStart` backend đã có; chương đang nghe tô màu chính; chương 1 tiểu mục trùng tên không lặp.
- `3dfdc6d` **Giọng Thiền Tâm Đức đọc "chánh" thành "tránh"** (chánh niệm, chánh kiến…). Gốc: bộ phiên âm sea_g2p của
  VieNeu gộp ch/tr thành cùng âm `tʃ`, model tự chọn; chỉ giọng này chọn sai (anh Việt nghe tai: Hải Đăng, Ngọc Huyền vẫn đúng).
  Sửa: `scripts/tts/audio_gen.py` có bảng `PHONEME_OVERRIDES` **theo giọng** (từ → phiên âm IPA), bọc
  `vieneu.v3turbo.phonemize_text_with_emotions`; hiện chỉ `"Thiền Tâm Đức": {"chánh": "tʃˈeɜɲ"}` (anh Việt đã nghe duyệt).
  Giọng khác không đụng. App tự đồng bộ script khi nội dung khác (`EnsureScripts`).
  **Bài học:** Whisper (mlx-whisper large-v3-turbo, `uvx --offline --from mlx-whisper mlx_whisper`) nghe ch/tr của giọng
  VieNeu KHÔNG tin được — ghi "tránh" cho mọi giọng dù tai người nghe đúng. Kết luận âm phải dựng mp3 gửi anh nghe.
- Bản RC đã build và chạy trên máy: `desktop/build/bin/Sano.app` = 0.1.18-rc.4 (có cả 2 sửa).

### Việc dở dang lúc handoff — phiên sau làm tiếp

1. ~~Đọc lại cuốn "Chánh niệm, nghệ thuật của sự có mặt"~~ **ĐÃ XONG** (13:39): 21/39 mục có chữ "chánh" đã đọc lại
   bằng script mới qua `ApplySection` + `Repack` (mp3, metadata, zip đều cập nhật). Chương trình tạm đã xoá. Còn lại:
   anh Việt nghe thử mục "Vì sao có bản nghe này" trong app xác nhận "chánh niệm" đúng tai.
2. **Phát hành 0.1.18**: `VERSION`, `CHANGELOG.md` (2 mục trên, viết ngắn cho người dùng), link tải `README.md`, commit
   `chore: phát hành v0.1.18`, push, tag, chờ CI, công khai (mục 2). Chưa được lệnh push — hỏi anh trước.

## 4. Đã làm trong phiên 27/09 (0.1.16 → 0.1.17)

- Dấu ":" giữa câu đọc thành ngắt câu (VieNeu chỉ nghỉ ~0,25 s ở ":"), giữ 10:30, 3:1, link.
- Phím tắt khi nghe: Space dừng / nghe tiếp, ← → tiểu mục trước / sau (bỏ qua khi gõ chữ, có hộp thoại, ở Tạo sách / Sửa sách).
- **Sửa sách (D11)**: menu ⋯ ở Thư viện "Sửa sách", bánh răng màn nghe "Sửa mục đang nghe". 4 tab:
  Nội dung (sửa chữ từng mục → Lưu & đọc lại, chấm vàng = chưa lưu, bản sửa giữ trong localStorage),
  Thông tin & bìa (đổi tên → vẽ lại bìa tự vẽ + đọc lại lời giới thiệu), Giọng đọc (đổi giọng cả cuốn,
  đọc vào `.doc-lai/` trong thư mục sách, xong hết mới thay, tắt app thì mở lại tự đọc tiếp), Từ điển.
  Hộp "Sửa thông tin" cũ đã bỏ.
- **Từ điển cách đọc (D12)**: bộ chuẩn 65 từ (nhúng) → từ điển chung `~/Sano/.tu-dien.tsv` → từ điển
  của cuốn `tu-dien.tsv` (trong zip là `pronunciations.tsv`, tuỳ chọn). Dùng ở Nạp file (bấm từ viết tắt
  bị cảnh báo), Nghe thử (bôi đen → "Đọc từ này là…", khung từ điển bên phải), Sửa sách (tab Từ điển,
  đổi cách đọc → đánh dấu mục cần đọc lại), Cài đặt (Từ điển chung, ghi đè từ có sẵn).
- Nghe thử: đổi giọng / tên / lời mở đầu không còn xoá chỗ đã sửa; chỗ sửa không khớp nữa thì báo.
- Ô xác nhận quyền dùng tài liệu nằm cố định trên thanh dưới cạnh nút render.
- Bìa: kệ sách dùng bản thu nhỏ 480px (đệm `~/Sano/.tam/bia`), ảnh lỗi thì thử lại rồi hiện bìa tự vẽ;
  ảnh bìa tự chọn lớn hơn 1200×1600 thì thu nhỏ giữ tỉ lệ.

## 5. Việc còn tồn

- App chưa có nút "Đọc lại cả cuốn" (khi script đọc / bộ đọc đổi, sách cũ không tự đọc lại; hiện phải mượn từ điển
  của cuốn để đánh dấu mục, hoặc chạy tay như mục 3). Tính năng mới → cần wireframe (tab Giọng đọc trong Sửa sách).

- 11 nhánh cũ trên máy + remote (`feat/cach-doc-3-cap`, `feat/thu-vien-bo-sach`, `release/0.1.3`, `release/0.1.4`,
  `fix/linux-audio`…): nhiều nhánh đã gộp vào `main`, cần kiểm rồi mới xoá.
- 3 nhánh Dependabot trên remote (TypeScript 7.0.2, Vite 8.3.0, vue-tsc 3.3.11) chưa gộp.
- Thanh "Đọc từ này là…" ở Nghe thử đang nằm góc trên phải ô chữ, chưa bám sát vị trí từ bôi đen như wireframe D12.
- Từ điển chỉ nhận một từ liền (chữ, số, &), chưa nhận cụm nhiều từ ("Hồ Chí Minh").
- Đổi Từ điển chung không đánh dấu mục cần đọc lại ở sách cũ (phải thêm từ vào từ điển của cuốn).
- Trang hướng dẫn (`docs/`) chưa có bài riêng cho Sửa sách và Từ điển cách đọc.
- Khi app bị tắt cứng (không qua hộp hỏi), tiến trình Python của bộ đọc có thể còn chạy tới hết lượt (có từ trước, cả Tạo sách).

## 6. Ý tưởng chưa làm: giọng đọc tiếng Anh, tiếng Nhật (anh Việt: "chưa cần làm", 27/09)

Nhiều người hỏi. Đã phân tích, **chưa làm**. Khi làm lại thì đọc đoạn này trước.

**Hướng đã thống nhất:** mỗi ngôn ngữ là một **gói bộ đọc cài riêng**. Lần đầu mở Sano hỏi làm sách tiếng nào
(mặc định tiếng Việt → VieNeu như hiện nay); thêm / gỡ tiếng khác ở Cài đặt → Bộ đọc → Thêm ngôn ngữ; khi tạo sách
chọn ngôn ngữ (chỉ hiện tiếng đã cài); mỗi cuốn nhớ ngôn ngữ (`metadata.json` đã có `language: "vi"`).
Bộ đọc gợi ý: **Kokoro-82M** (Apache 2.0, chạy trên máy) — một mô hình dùng chung nhiều tiếng, mỗi tiếng thêm giọng
+ phần riêng (tiếng Nhật cần từ điển Kanji nặng). Số liệu Kokoro chưa kiểm trên máy thật.

**Mức độ dính VieNeu / tiếng Việt trong code:**

| Phần | Việc phải làm | Rủi ro |
|---|---|---|
| Cài bộ đọc (`desktop/internal/setup`, ~3.400 dòng, 1 venv VieNeu, khoá phiên bản, SHA-256 mô hình) | Thành "gói theo ngôn ngữ", mỗi bộ đọc một môi trường Python riêng | **Cao** — 3 hệ điều hành, Mac Intel từng lỗi onnxruntime; hỏng thì người dùng tiếng Việt mất bộ đọc |
| Chuẩn hoá lời đọc (`normalize.go`, `spoken.go`, từ điển) | Bộ chuẩn hoá theo ngôn ngữ | Trung bình |
| Đọc docx (nhận "Mục lục", cảnh báo viết tắt, lời mở đầu tiếng Việt) | Theo ngôn ngữ của sách | Thấp |
| Giọng + giao diện (tab miền Bắc/Trung/Nam, giọng khuyên dùng, "VieNeu-TTS (giọng)" viết cứng) | Theo ngôn ngữ; màn cài lần đầu; chọn ngôn ngữ khi tạo sách (cần wireframe) | Trung bình |
| Chữ chạy theo (`lyrics.ts`) | Chỉnh nhịp ước lượng | Thấp |

Thuận lợi: Go gọi bộ đọc qua script Python theo giao ước đơn giản (file chữ vào → wav ra, in `✨ <stem>_full.wav`).
Kokoro chỉ cần `kokoro_batch.py` cùng giao ước; tạo sách, Sửa sách, M4B, zip, trình phát không đổi.

**Ước lượng (AI làm):** tiếng Anh khoảng 1 ngày viết code, 1–2 ngày thật (chờ CI 15–30 phút/lượt + duyệt wireframe).
Tiếng Nhật thêm 2–3 ngày và không ai trong nhóm kiểm được chất lượng.

**Rủi ro tốc độ AI không giảm được:** không test tay được trên Windows / Mac Intel / Linux (chỉ CI build + smoke);
chất lượng giọng phải người nghe; mỗi bộ đọc thêm là thêm bảo trì lâu dài (khoá phiên bản, vá bảo mật, lỗi từng HĐH).

**Khi làm:** (1) chạy thử Kokoro tiếng Anh trong thư mục tạm, anh nghe + đo dung lượng, tốc độ; (2) wireframe 3 chỗ
(chọn ngôn ngữ lúc cài, Cài đặt → Thêm ngôn ngữ, chọn ngôn ngữ khi tạo sách); (3) không đụng đường cài VieNeu,
gói mới là phần cài thêm tuỳ chọn; (4) cho vài người dùng thật trên Windows + Mac Intel thử trước khi phát hành;
(5) tiếng Anh trước, tiếng Nhật sau; không chạy theo "nhiều model".
