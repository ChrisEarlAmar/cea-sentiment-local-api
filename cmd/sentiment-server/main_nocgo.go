//go:build !cgo

package main

import "log"

func main() {
	log.Fatal("this build has CGO disabled; install a C compiler and rebuild with CGO_ENABLED=1")
}
