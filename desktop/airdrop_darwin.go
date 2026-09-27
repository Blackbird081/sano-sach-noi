//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit
#import <AppKit/AppKit.h>

// Mở bảng AirDrop của macOS cho một file. Chạy trên luồng chính (AppKit).
// Trả 0 khi máy không gửi AirDrop được (tắt Wi-Fi/Bluetooth, máy không hỗ trợ).
static int sanoAirDrop(const char *path) {
	__block int ok = 0;
	NSString *p = [NSString stringWithUTF8String:path];
	void (^run)(void) = ^{
		NSURL *url = [NSURL fileURLWithPath:p];
		NSSharingService *s = [NSSharingService sharingServiceNamed:NSSharingServiceNameSendViaAirDrop];
		if (s != nil && [s canPerformWithItems:@[url]]) {
			[s performWithItems:@[url]];
			ok = 1;
		}
	};
	if ([NSThread isMainThread]) run();
	else dispatch_sync(dispatch_get_main_queue(), run);
	return ok;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

const airDropSupported = true

func airDrop(path string) error {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	if C.sanoAirDrop(cs) == 0 {
		return errors.New("máy này chưa gửi AirDrop được: bật Wi-Fi và Bluetooth rồi thử lại, hoặc dùng cách khác")
	}
	return nil
}
