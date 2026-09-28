//go:build !darwin

package main

import "errors"

// Máy khác: giao diện chép ảnh bằng API clipboard của trình duyệt (WebView2, WebKitGTK).
const copyImageSupported = false

func copyImage([]byte) error { return errors.New("chép ảnh qua phần Go chỉ có trên máy Mac") }
