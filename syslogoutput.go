package main

import (
	"time"
	"log"
	syslog "github.com/RackSec/srslog"
)

type syslogOutputPrivate struct {
	fd *syslog.Writer
}


func newSyslogOutputDesc() *outputDesc {
	return &outputDesc{
		sendStart: syslogOutputSendStart,
		open: syslogOpen,
		private: new(syslogOutputPrivate),
	}
}

func syslogOpen(sc *scriptLine) {
	s, err := syslog.Dial("tcp",sc.target,
		syslog.LOG_INFO|syslog.LOG_DAEMON, sc.msgid);
	if err != nil {
		log.Fatalf("failed to open syslog %s: %v\n",sc.target, err)
	}
	priv := (*syslogOutputPrivate)(sc.desc.private.(*syslogOutputPrivate))
	s.SetFormatter(syslog.RFC5424Formatter);
	priv.fd=s

}

func syslogOutputSendStart(sc *scriptLine, buf []byte) {
	priv := (*syslogOutputPrivate)(sc.desc.private.(*syslogOutputPrivate))
	for {
		_, err:=priv.fd.WriteWithPriority(sc.priority, buf)
		if err == nil {
			break;
		}
		log.Printf("unable to write to %s, pausing: %v\n",sc.target, err);
		time.Sleep(time.Second);
	}
}

