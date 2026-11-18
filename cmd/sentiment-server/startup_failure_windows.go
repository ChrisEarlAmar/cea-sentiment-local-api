//go:build cgo && windows

package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
)

const startupFailureTimeoutMilliseconds = 8000

func startupFailure(err error) {
	message := fmt.Sprintf("The local sentiment service could not start.\n\n%s\n\nIf another copy is already using the selected port, close it or change the port in %s.\n\nThis warning closes automatically in 8 seconds.", err, configFileName)
	text := syscall.StringToUTF16Ptr(message)
	title := syscall.StringToUTF16Ptr("Local Sentiment API")
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxTimeout := user32.NewProc("MessageBoxTimeoutW")
	messageBoxTimeout.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10, 0, startupFailureTimeoutMilliseconds)
	log.Print(err)
	os.Exit(1)
}
