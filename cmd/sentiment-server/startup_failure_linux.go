//go:build cgo && linux

package main

import "log"

func startupFailure(err error) {
	log.Fatal(err)
}
