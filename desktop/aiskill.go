package main

// Nạp skill cho AI (bước Cách đọc, hộp "Nạp skill cho AI"): lưu về máy file
// skill Claude (zip có SKILL.md) hoặc file hướng dẫn cho ChatGPT / Gemini, dựng từ
// cùng nguồn prompt với app (docs/prompts).

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"sano/docs/prompts"
)

// SaveAIGuide lưu file cho AI: kind "claude" → sano-sach-noi.zip, còn lại →
// sano-huong-dan-ai.txt. Hộp lưu file mặc định thư mục Tải về, lưu xong mở thư
// mục chọn sẵn file. Trả đường dẫn đã lưu; huỷ → "".
func (a *App) SaveAIGuide(kind string) (string, error) {
	name, data, err := aiGuideFile(kind)
	if err != nil {
		return "", err
	}
	path, reveal, err := a.saveTarget(name)
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("lưu %s: %w", name, err)
	}
	if reveal {
		_ = openPath(path, true)
	}
	return path, nil
}

func aiGuideFile(kind string) (string, []byte, error) {
	if kind == "claude" {
		data, err := prompts.SkillZip()
		return prompts.SkillZipName, data, err
	}
	return prompts.GuideFileName, []byte(prompts.GuideText()), nil
}

// saveTarget hỏi nơi lưu file tên name (dev/test: thư mục EnvSampleOut, không hỏi).
func (a *App) saveTarget(name string) (string, bool, error) {
	if v := strings.TrimSpace(os.Getenv(EnvSampleOut)); v != "" {
		return filepath.Join(v, name), false, nil
	}
	if a.ctx == nil {
		return "", false, errors.New("ứng dụng chưa khởi động xong")
	}
	ext := filepath.Ext(name)
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:                "Lưu " + name,
		DefaultDirectory:     downloadsDir(),
		DefaultFilename:      name,
		Filters:              []wruntime.FileFilter{{DisplayName: name, Pattern: "*" + ext}},
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", false, fmt.Errorf("mở hộp lưu file: %w", err)
	}
	if path == "" {
		return "", false, nil
	}
	if !strings.EqualFold(filepath.Ext(path), ext) {
		path += ext
	}
	return path, true, nil
}
