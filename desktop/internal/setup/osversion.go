package setup

import (
	"fmt"
	"strconv"
	"strings"
)

// ErrOSTooOld — hệ điều hành cũ hơn mức thư viện đọc giọng (onnxruntime) có bản dựng.
var ErrOSTooOld = fmt.Errorf("hệ điều hành chưa đủ mới")

// minMacOS — macOS thấp nhất có bản onnxruntime ghim trong uv.lock:
// Apple Silicon dùng 1.24 (bản dựng macOS 14), Mac Intel dùng 1.23.2 (macOS 13).
func minMacOS(goarch string) int {
	if goarch == "amd64" {
		return 13
	}
	return 14
}

// checkOS — lỗi nếu máy không cài được bộ đọc vì hệ điều hành quá cũ.
// ver rỗng (không đọc được) thì cho qua, không chặn.
func checkOS(goos, goarch, ver string) error {
	if goos != "darwin" || ver == "" {
		return nil
	}
	major, err := strconv.Atoi(strings.SplitN(ver, ".", 2)[0])
	if err != nil {
		return nil
	}
	if need := minMacOS(goarch); major < need {
		return fmt.Errorf("%w: bộ đọc cần macOS %d trở lên, máy đang chạy macOS %s", ErrOSTooOld, need, ver)
	}
	return nil
}
