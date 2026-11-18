//go:build cgo && windows

package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
)

func startupFailure(err error) {
	message := fmt.Sprintf("The local sentiment service could not start.\n\n%s\n\nCheck %s beside the executable, then try again.", err, configFileName)
	text := syscall.StringToUTF16Ptr(message)
	title := syscall.StringToUTF16Ptr("Local Sentiment API")
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")
	messageBox.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
	log.Print(err)
	os.Exit(1)
}
