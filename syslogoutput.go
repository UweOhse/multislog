package main

import (
	syslog "github.com/RackSec/srslog"
	"log"
	"time"
	"strings"
)

type syslogOutputPrivate struct {
	fd *syslog.Writer
}

func newSyslogOutputDesc() *outputDesc {
	return &outputDesc{
		sendStart: syslogOutputSendStart,
		open:      syslogOpen,
		private:   new(syslogOutputPrivate),
	}
}

func syslogParseTarget(sc *scriptLine) (proto, target, formatter, framer string) {
	target=sc.target
	proto="tcp"
	formatter="rfc5424"
	framer=""

	// target:
	//   1234:514
	//   (tcp|udp):1234:514
	//   (tcp|udp)/format:1234:514
	//   (tcp|udp)/format/frame:1234:514
	//   unix
	//   unix/format/frame
	//
	var formatframe string

	if target=="unix" {
		target=""
		proto=""
		formatter="unix";
		framer="";
	} else if target[0:5]=="unix/" {
		formatframe=target[5:]
		target=""
		proto=""
		framer="" // change default
	} else if target[0:4]=="tcp:" {
		target=target[4:]
		proto="tcp"
	} else if target[0:4]=="udp:" {
		target=target[4:]
		proto="udp"
	} else if target[0:4]=="tcp/" || target[0:4]=="udp/" {
		if target[0:4]=="udp/" {
			proto="udp"
		}
		idx:=strings.IndexByte(target, ':')
		formatframe=target[4:idx]
		target=target[idx+1:]
	} else {
		// take as is
	}
	if formatframe != "" {
		idx:=strings.IndexByte(formatframe, '/')
		if idx==-1 {
			formatter=formatframe
		} else {
			formatter=formatframe[0:idx]
			framer=formatframe[idx+1:]
		}
	}
	return proto,target,formatter,framer
}

func syslogOpen(sc *scriptLine) {
	proto, target, formatter, framer := syslogParseTarget(sc)
	log.Printf("in: %s => proto %s target %s formatter %s framer %s", sc.target, proto,target,formatter, framer)

	s, err := syslog.Dial(proto, target,
		syslog.LOG_INFO|syslog.LOG_DAEMON, sc.msgid)
	if err != nil {
		log.Fatalf("failed to open syslog %s: %v\n", sc.target, err)
	}
	priv := (*syslogOutputPrivate)(sc.desc.private.(*syslogOutputPrivate))
	switch formatter {
	case "rfc3164", "3164":
		s.SetFormatter(syslog.RFC3164Formatter)
	case "rfc5424", "5424":
		s.SetFormatter(syslog.RFC5424Formatter)
	case "unix":
		s.SetFormatter(syslog.UnixFormatter)
	case "compat", "":
		s.SetFormatter(syslog.DefaultFormatter)
	default:
		log.Fatalf("unknown formatter '%s' in %s\n", formatter, sc.target)
	}
	switch framer {
	case "rfc5425", "5425":
		s.SetFramer(syslog.RFC5425MessageLengthFramer)
	case "none","":
		s.SetFramer(syslog.DefaultFramer)
	default:
		log.Fatalf("unknown framer '%s' in %s\n", framer, sc.target)
	}
	priv.fd = s

}

func syslogOutputSendStart(sc *scriptLine, buf []byte) {
	priv := (*syslogOutputPrivate)(sc.desc.private.(*syslogOutputPrivate))
	for {
		_, err := priv.fd.WriteWithPriority(sc.priority, buf)
		if err == nil {
			break
		}
		log.Printf("unable to write to %s, pausing: %v\n", sc.target, err)
		time.Sleep(time.Second)
	}
}
