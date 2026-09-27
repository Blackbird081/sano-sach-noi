//go:build !windows

package setup

func cpuName(goos string) string  { return cpuNameUnix(goos) }
func ramBytes(goos string) uint64 { return ramBytesUnix(goos) }

// osVersion — số phiên bản macOS (vd "12.7.6"); hệ khác trả rỗng.
func osVersion(goos string) string {
	if goos != "darwin" {
		return ""
	}
	return sysctl("kern.osproductversion")
}
