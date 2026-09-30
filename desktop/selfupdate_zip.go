package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Dùng cho bản chạy ngay trên Windows (selfupdate_windows.go); để chung cho
// test chạy được trên mọi máy.

// extractExe lấy đúng Sano.exe trong file .zip bản chạy ngay ra dest.
func extractExe(zipFile, dest string) error {
	found, err := extractFile(zipFile, "Sano.exe", dest)
	if err == nil && !found {
		err = errors.New("file .zip bản mới không có Sano.exe")
	}
	return err
}

// extractFile lấy file name (nằm ở gốc file .zip) ra dest; không có thì found = false.
func extractFile(zipFile, name, dest string) (found bool, err error) {
	zr, err := zip.OpenReader(zipFile)
	if err != nil {
		return false, fmt.Errorf("mở file .zip bản mới: %w", err)
	}
	defer zr.Close()
	tooBig := fmt.Errorf("%s trong file .zip lớn bất thường", name)
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		if f.UncompressedSize64 > maxInstallerBytes {
			return true, tooBig
		}
		rc, err := f.Open()
		if err != nil {
			return true, err
		}
		defer rc.Close()
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return true, err
		}
		n, err := io.Copy(out, io.LimitReader(rc, maxInstallerBytes+1))
		if cerr := out.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return true, err
		}
		if n > maxInstallerBytes {
			return true, tooBig
		}
		return true, nil
	}
	return false, nil
}

// replaceBridge đặt sano-mcp.exe (cầu nối MCP) của bản mới cạnh Sano.exe. Claude đang chạy bản
// cũ thì file bị khoá → đổi tên bản cũ thành .cu-<pid> (Windows cho đổi tên file đang chạy),
// cleanupAfterUpdate xoá sau. Lỗi thì bỏ qua: app vẫn chạy, màn MCP báo thiếu cầu nối.
func replaceBridge(zipFile, dir, pid string) {
	fresh := filepath.Join(dir, "sano-mcp.exe.moi")
	if found, err := extractFile(zipFile, "sano-mcp.exe", fresh); !found || err != nil {
		_ = os.Remove(fresh)
		return
	}
	cur := filepath.Join(dir, "sano-mcp.exe")
	if _, err := os.Stat(cur); err == nil {
		if os.Remove(cur) != nil {
			_ = os.Rename(cur, cur+".cu-"+pid)
		}
	}
	if err := os.Rename(fresh, cur); err != nil {
		_ = os.Remove(fresh)
	}
}
