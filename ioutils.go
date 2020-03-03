package main

import (
	"log"
	"os"
	"syscall"
	"time"
)

func openTrunc(fn string) (*os.File, error) {
	return os.OpenFile(fn, os.O_WRONLY|syscall.O_NDELAY|os.O_TRUNC|os.O_CREATE, 0644)
}
func openAppend(fn string) (*os.File, error) {
	return os.OpenFile(fn, os.O_WRONLY|syscall.O_NDELAY|os.O_APPEND|os.O_CREATE, 0600)
}
func openRead(fn string) (*os.File, error) {
	// really, gosec is stupid.
	//    G304 (CWE-22): Potential file inclusion via variable
	// now, tell me, why should this be a problem? Our user tells us where to work, anyway.
	// #nosec
	return os.Open(fn)
}
func closeOnExec(f *os.File) { syscall.CloseOnExec(int(f.Fd())) }

func writeLoop(target string, fd *os.File, buf []byte) {
	for {
		n, err := fd.Write(buf)
		if err == nil { // all written
			break
		}
		log.Printf("unable to write to %s, pausing: %v\n", target, err)
		buf = buf[n:]
		time.Sleep(time.Second)
	}
}
