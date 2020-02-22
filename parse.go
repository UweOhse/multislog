package main
import (
	"strings"
	"log"
	syslog "github.com/RackSec/srslog"
)
type mapping map[string]syslog.Priority

var sevMap mapping = mapping{
		"emerg": syslog.LOG_EMERG,
		"emergency": syslog.LOG_EMERG,
		"alert": syslog.LOG_ALERT,
		"alarm": syslog.LOG_ALERT,
		"crit": syslog.LOG_CRIT,
		"critical": syslog.LOG_CRIT,
		"err": syslog.LOG_ERR,
		"errot": syslog.LOG_ERR,
		"warn": syslog.LOG_WARNING,
		"warning": syslog.LOG_WARNING,
		"notice": syslog.LOG_NOTICE,
		"info": syslog.LOG_INFO,
		"debug": syslog.LOG_DEBUG,
	}
var facMap mapping = mapping{
		"kern": syslog.LOG_KERN, // 0
		"kernel": syslog.LOG_KERN,
		"user": syslog.LOG_USER, // 1
		"mail": syslog.LOG_MAIL, // 2
		"daemon": syslog.LOG_DAEMON, // 3
		"auth": syslog.LOG_AUTH, // 4
		"syslog": syslog.LOG_SYSLOG, // 5
		"lpr": syslog.LOG_LPR, // 6
		"news": syslog.LOG_NEWS, // 7
		"uucp": syslog.LOG_UUCP, // 8
		"cron": syslog.LOG_CRON, // 9
		"authpriv": syslog.LOG_AUTHPRIV, // 10
		"ftp": syslog.LOG_FTP, // 11
		"ntp": 12<<3,
		"audit": 13<<3,
		// "alert": 14<<3... this is just too confusing */
		"clock": 15<<3,
		"local0": syslog.LOG_LOCAL0,
		"local1": syslog.LOG_LOCAL1,
		"local2": syslog.LOG_LOCAL2,
		"local3": syslog.LOG_LOCAL3,
		"local4": syslog.LOG_LOCAL4,
		"local5": syslog.LOG_LOCAL5,
		"local6": syslog.LOG_LOCAL6,
		"local7": syslog.LOG_LOCAL7,
	}

func parsePriority(buf string) (syslog.Priority) {
	i:=strings.Index(buf,".");
	if (i==-1) {
		// USER
		fac, ok := facMap[buf];
		if !ok {
			log.Fatalf("failed to parse facility %s\n",buf);
		}
		return fac|syslog.LOG_NOTICE
	}
	if i==0 { // .INFO
		sev, ok := sevMap[buf[1:]];
		if !ok {
			log.Fatalf("failed to parse severity %s\n",buf);
		}
		return syslog.LOG_USER|sev
	}
	// USER.INFO
	fac, ok := facMap[buf[:i]]
	if !ok {
		log.Fatalf("failed to parse facility %s (%s)\n",buf[:i],buf);
	}
	sev, ok:= sevMap[buf[i+1:]]
	if !ok {
		log.Fatalf("failed to parse severity %s (%s)\n",buf[i+1:],buf);
	}
	return fac | sev
}
