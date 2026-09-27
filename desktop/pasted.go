package main

// Dán văn bản (bước Cách đọc cấp 2/3): AI không tạo được file Word thì người
// dùng dán văn bản "% / # / ##" AI trả về. Sano đổi thành .docx tạm trong
// ~/Sano/.tam rồi đi tiếp đúng luồng nạp file cũ (mục lục, cảnh báo, render).

import (
	"fmt"
	"os"
	"path/filepath"

	"sano/internal/bookmaker"
)

const pastedDocxName = "Van-ban-dan-tu-AI.docx"

// PastedDocxLabel — tên hiện ở ô file sau khi dán (file tạm không cần lộ đường dẫn).
const PastedDocxLabel = "Văn bản dán từ AI"

// PastedText đổi văn bản dán thành .docx tạm và trả như file người dùng vừa chọn.
func (a *App) PastedText(text string) (*DocxFile, error) {
	dir := filepath.Join(a.lib.Root(), ".tam")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục tạm: %w", err)
	}
	return writePastedDocx(filepath.Join(dir, pastedDocxName), text)
}

func writePastedDocx(path, text string) (*DocxFile, error) {
	data, err := bookmaker.TextToDocx(text)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("ghi văn bản dán: %w", err)
	}
	f, err := describeDocx(path)
	if err != nil {
		return nil, err
	}
	f.Name = PastedDocxLabel
	return f, nil
}
