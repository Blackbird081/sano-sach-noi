package setup

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"sano/desktop/internal/tts"
)

// minRAM — dưới mức này đọc giọng rất chậm hoặc hết bộ nhớ (mô hình ~600 MB).
const minRAM = 4 << 30

// Info — thông tin hiện trên màn cài trước khi bấm Cài.
type Info struct {
	OS            string `json:"os"`
	OSVersion     string `json:"osVersion"`
	Arch          string `json:"arch"`
	CPU           string `json:"cpu"`
	Cores         int    `json:"cores"`
	RAMBytes      int64  `json:"ramBytes"`
	FreeBytes     int64  `json:"freeBytes"`
	NeedBytes     int64  `json:"needBytes"`     // dung lượng bộ đọc sau khi cài
	RequiredFree  int64  `json:"requiredFree"`  // cần trống lúc cài
	DownloadBytes int64  `json:"downloadBytes"` // tổng tải về
	DataDir       string `json:"dataDir"`
	UsedBytes     int64  `json:"usedBytes"` // thư mục bộ đọc app đang chiếm
	Enough        bool   `json:"enough"`
	Supported     bool   `json:"supported"`
	Blocked       bool   `json:"blocked"` // hệ điều hành quá cũ: không cài được
	Note          string `json:"note"`
}

// MachineInfo đo máy + thư mục bộ đọc. Không lỗi: thiếu gì thì để trống.
func MachineInfo(l tts.Layout, goos, goarch string) Info {
	inf := Info{
		OS:            osName(goos),
		OSVersion:     osVersion(goos),
		Arch:          goarch,
		CPU:           cpuName(goos),
		Cores:         runtime.NumCPU(),
		RAMBytes:      int64(ramBytes(goos)),
		NeedBytes:     InstalledSize,
		RequiredFree:  RequiredBytes,
		DownloadBytes: DownloadBytes,
		DataDir:       l.Root,
		UsedBytes:     DirSize(l.Root),
		Supported:     officiallySupported(goos, goarch),
	}
	if free, err := freeBytes(existingParent(l.Root)); err == nil {
		inf.FreeBytes = int64(free)
	}
	need := RequiredBytes - inf.UsedBytes
	if need < minFreeBytes {
		need = minFreeBytes
	}
	inf.Enough = inf.FreeBytes == 0 || inf.FreeBytes >= need
	osErr := checkOS(goos, goarch, inf.OSVersion)
	inf.Blocked = osErr != nil
	switch {
	case inf.Blocked:
		inf.Note = "Bộ đọc cần macOS " + strconv.Itoa(minMacOS(goarch)) + " trở lên, máy đang chạy macOS " + inf.OSVersion + " — cập nhật macOS rồi mở lại Sano"
	case !inf.Enough:
		inf.Note = "Ổ đĩa còn trống " + humanGB(inf.FreeBytes) + ", cần khoảng " + humanGB(need)
	case inf.RAMBytes > 0 && inf.RAMBytes < minRAM:
		inf.Note = "Máy dưới 4 GB RAM — đọc giọng sẽ chậm"
	case !inf.Supported:
		inf.Note = "Bộ đọc chưa được thử trên máy " + inf.OS + " " + goarch + " — vẫn cài thử được"
	}
	return inf
}

func osName(goos string) string {
	switch goos {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	}
	return goos
}

// existingParent — thư mục gần nhất đã tồn tại (để đo đĩa trống trước khi tạo).
func existingParent(p string) string {
	for {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(p)
		if parent == p {
			return p
		}
		p = parent
	}
}
