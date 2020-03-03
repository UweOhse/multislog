package main

import (
	"errors"
	"fmt"
	syslog "github.com/RackSec/srslog"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultBuflen int    = 8192
	minMaxSize    uint64 = 20
)

var buflen int = defaultBuflen
var flagExiting bool

const (
	acNone = iota
	acAction
	acSelect
	acDeSelect
	acDone
)
const (
	tsNone = iota
	tsTAI64
	tsEpoch
	tsEpochMs
	tsEpochUs
	tsEpochNs
	tsRFC3339
	tsRFC3339Nano
)

type outputDesc struct {
	open         func(sc *scriptLine)
	quit         func(sc *scriptLine)
	flush        func(sc *scriptLine)
	sendStart    func(*scriptLine, []byte)
	sendContinue func(*scriptLine, []byte)
	sendEnd      func(*scriptLine)
	private      interface{}
}

type scriptLine struct {
	selector  string
	target    string
	typ       int
	timestamp int
	maxSize   uint64 // acDir only
	maxFiles  uint64 // acDir only
	processor string // acDir only

	selected bool // internal
	priority syslog.Priority
	msgid    string

	desc *outputDesc
}

var script []scriptLine
var curFormattedTimestamp []byte
var curTimestamp time.Time

func doit(readChan chan syncReadData, errChan chan error, exitChan, flushChan chan bool) {
	// limit the memory usage and the influence of core dumps inserted into logs.
	// rd := bufio.NewReaderSize(os.Stdin,BUFLEN)
	var gotEOF bool
	var gotReadError bool

	for {
		if gotEOF {
			break
		}
		var ok bool
		var sd syncReadData
		select {
		case <-flushChan:
			for _, sc := range script {
				if sc.desc != nil && sc.desc.flush != nil {
					sc.desc.flush(&sc)
				}
			}
			continue
		case <-exitChan:
			flagExiting = true
		case sd, ok = <-readChan:
			if !ok {
				gotEOF = true
			}
		case err := <-errChan:
			// note: a read error is not an EOF. that is signaled by closing the readc
			if !gotReadError {
				log.Printf("failed to read from stdin: %v\n", err)
				gotReadError = true
			}
		}

		if gotEOF && len(sd.line) == 0 {
			break
		}
		if len(sd.line) == 0 {
			continue
		}
		curTimestamp = time.Now()

		/* part2: select channels */
		selected := true
		var done = false
		for i := 0; i < len(script); i++ {
			if done {
				script[i].selected = false
				continue
			}
			typ := script[i].typ
			switch {
			case typ == acSelect:
				if !selected && match(script[i].selector, string(sd.line)) {
					selected = true
				}
			case typ == acDeSelect:
				if selected && match(script[i].selector, string(sd.line)) {
					selected = false
				}
			case typ == acAction:
				script[i].selected = selected
			case typ == acDone:
				if selected {
					done = true
				}
			}
		}
		/* part3: writing the first part. */
		for i := 0; i < len(script); i++ {
			if script[i].selected {
				curFormattedTimestamp = timestamp(script[i].timestamp, curTimestamp)
				if script[i].desc != nil && script[i].desc.sendStart != nil {
					script[i].desc.sendStart(&script[i], sd.line)
				}
			}
		}
		/* part4: writing later parts. */
		for {
			if sd.isComplete || gotEOF {
				break
			}
			// not wantFlush, not exitChan, not errChan
			sd, ok = <-readChan
			if !ok {
				gotEOF = true
			}
			for i := 0; i < len(script); i++ {
				if script[i].selected {
					if script[i].desc != nil && script[i].desc.sendContinue != nil {
						script[i].desc.sendContinue(&script[i], sd.line)
					}
				}
			}
		}
		/* part 5: write \n */
		for i := 0; i < len(script); i++ {
			if script[i].selected && script[i].typ == acAction {
				if script[i].desc != nil && script[i].desc.sendEnd != nil {
					script[i].desc.sendEnd(&script[i])
				}
			}
		}
	}
}

func setupScript() {
	flagTS := tsNone
	var maxFiles uint64 = 10
	var maxSize uint64 = 99999
	var processor string
	var msgid string
	var err error
	priority := syslog.LOG_USER | syslog.LOG_NOTICE
	for i := 1; i < len(os.Args); i++ {
		var n scriptLine
		ch := os.Args[i][0]
		switch ch {
		case '-':
			n.typ = acDeSelect
			n.selector = os.Args[i][1:]
		case '+':
			n.typ = acSelect
			n.selector = os.Args[i][1:]
		case 'e':
			n.typ = acAction
			n.desc = newStderrOutputDesc()
			n.target = "stderr"
		case '=':
			n.typ = acAction
			n.desc = newFileOutputDesc()
			n.target = os.Args[i][1:]
		case 't':
			switch strings.ToLower(os.Args[i]) {
			case "t":
				flagTS = tsTAI64
			case "t-":
				flagTS = tsNone
			case "tunix":
				flagTS = tsEpoch
			case "tunixms":
				flagTS = tsEpochMs
			case "tunixus":
				flagTS = tsEpochUs
			case "tunixns":
				flagTS = tsEpochNs
			case "trfc3339":
				flagTS = tsRFC3339
			case "trfc3339nano":
				flagTS = tsRFC3339Nano
			case "tt":
				flagTS = tsRFC3339Nano
			default:
				log.Fatalf("unable to understand %s\n", os.Args[i])
			}
		case '.':
			n.typ = acAction
			n.target = os.Args[i]
			n.desc = newDirOutputDesc()
		case '/':
			n.typ = acAction
			n.target = os.Args[i]
			n.desc = newDirOutputDesc()
		case 's':
			maxSize, err = strconv.ParseUint(os.Args[i][1:], 10, 64)
			if err != nil {
				log.Fatalf("failed to parse %s: %v\n", os.Args[i], err)
			}
			if maxSize < minMaxSize {
				maxSize = minMaxSize
			}
			if maxSize > 16777215 {
				maxSize = 16777215
			}
		case 'n':
			maxFiles, err = strconv.ParseUint(os.Args[i][1:], 10, 16)
			if err != nil {
				log.Fatalf("failed to parse %s: %v\n", os.Args[i], err)
			}
			if maxFiles < 2 {
				maxFiles = 2
			}
		case 'i':
			msgid = os.Args[i][1:]
		case '!':
			processor = os.Args[i][1:]
		case '@':
			n.typ = acAction
			n.target = os.Args[i][1:]
			n.desc = newSyslogOutputDesc()
		case 'p':
			priority = parsePriority(os.Args[i][1:])
		case 'd':
			if len(os.Args[i]) > 1 {
				log.Fatalf("unable to understand %s\n", os.Args[i])
			}
			n.typ = acDone
		default:
			log.Fatalf("unable to understand %s\n", os.Args[i])
		}
		if n.typ == acNone {
			continue
		}
		n.processor = processor
		n.timestamp = flagTS
		n.maxFiles = maxFiles
		n.maxSize = maxSize
		n.priority = priority
		n.msgid = msgid
		if n.typ == acAction {
			if n.desc.open != nil {
				n.desc.open(&n)
			}
		}
		script = append(script, n)
	}
}

type syncReadData struct {
	isComplete bool
	line       []byte
}

func makeReadChan(r io.Reader, bufSize int) (datachan chan syncReadData, errchan chan error) {
	readc := make(chan syncReadData, 1)
	errc := make(chan error, 1)
	go func() {
		curbuf := make([]byte, 0)
		readbuf := make([]byte, buflen)
		for {
			sd := new(syncReadData)
			sd.line = []byte(nil)
			n, err := r.Read(readbuf)
			if n != 0 {
				curbuf = append(curbuf, readbuf[:n]...)
			}
			for {
				i := strings.Index(string(curbuf), "\n")
				if -1 == i {
					break
				}
				sd.isComplete = true
				sd.line = curbuf[:i]
				readc <- *sd
				curbuf = curbuf[i+1:]
			}
			if err != nil {
				if len(curbuf) > 0 {
					sd.line = curbuf
					sd.isComplete = false
					readc <- *sd
				}
				close(readc)
				// we signal EOF by closing...
				if !errors.Is(err, io.EOF) {
					errc <- err
				}
				return
			}
			if flagExiting && len(curbuf) == 0 {
				close(readc)
				return
			}
			if len(curbuf) >= buflen {
				sd.line = append(sd.line, curbuf...)
				sd.isComplete = false
				curbuf = curbuf[:0]
				readc <- *sd
				continue
			}
		}
	}()
	return readc, errc
}
func quit() {
	for _, sc := range script {
		if sc.typ == acAction && sc.desc != nil && sc.desc.quit != nil {
			sc.desc.quit(&sc)
		}
	}
}
func main() {
	sigs := make(chan os.Signal, 1)
	exitChan := make(chan bool, 1)
	flushChan := make(chan bool, 1)
	readChan, errChan := makeReadChan(os.Stdin, 8192)

	log.SetFlags(0) // don't want a date, messes up self check
	if len(os.Args) == 2 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			fmt.Printf("multislog %v.\n", versionString)
			return
		}
	}

	// so that the dir* functions have a valid time stamp
	curTimestamp = time.Now()
	curFormattedTimestamp = timestamp(tsTAI64, curTimestamp)

	ev, ok := os.LookupEnv("MULTISLOG_BUFLEN")
	if ok {
		t, err := strconv.ParseInt(ev, 10, 32)
		if err != nil {
			log.Fatalf("failed to parse MULTISLOG_BUFLEN %s: %v\n", ev, err)
		}
		buflen = int(t)
	}

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGALRM)
	go func() {
		for {
			sig := <-sigs
			log.Printf("got signal %v\n", sig)
			if sig == syscall.SIGTERM {
				exitChan <- true
				return
			}
			if sig == syscall.SIGALRM {
				flushChan <- true
			}
		}
	}()
	setupScript()
	doit(readChan, errChan, exitChan, flushChan)
	quit()
}
