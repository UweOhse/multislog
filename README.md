# multislog

This is a rewrite of Dan Bernsteins [multilog](https://cr.yp.to/daemontools/multilog.html) with a number of extra features. If you don't use them, then the program should be perfectly compatible with multilog.

The main purpose of this tool is to store log messages in a way compatible to multilog, and be able to send the same messages to a central syslog server, too.

## Differences from multilog

multislog recognizes a number of additional script actions.

* @tcp:host:port
* @sysloghost:port
  send selected lines to a syslog server (by TCP).
* @udp:host:port
  send selected lines to a syslog server (by UDP).
* @unix
  send selected lines to the local syslog server (by unix domain socket).

* pfacility.severity
  set syslog facility ("user", "kern", "mail" and so on) and severity ("debug", "info", "crit" ...) for the following syslog actions.

* imsgid
  set the syslog MSGID for the following syslog actions to `msgid`. Note that MSGID in the lingua of RFC5424 doesn't identify the message, but the type of a message (or rather a kind of classification).
  use a `-` if in doubt (but that's the default anyway).

* t (timestamp)
  does not need to be the first action anymore.

* t-
  turns the timestamping off.

* tUnix
  switches the timestamp format to seconds since the epoch (1234567890, for 2009-02-14 00:31:30.000000000+01:00)
* tUnixMs
  switches the timestamp format to seconds since the epoch, with milliseconds appended (1234567890123).
* tUnixUs
  switches the timestamp format to seconds since the epoch, with microseconds appended (1234567890123456).
* tUnixNs
  switches the timestamp format to seconds since the epoch, with nanoseconds appended (1234567890123456789).
  This is for interoperability with promtail (grafana loki), not for beauty.
* tRFC3339
  switches the timestamp format to RFC3339. It uses a T as separator, not a space.
* tRFC3339Nano
  switches the timestamp format to RFC3339 with nanoseconds. It uses a T as separator, not a space.
* tT
  is a shorthand notation for tRFC3339Nano.
The t action is case insensitive, by the way.

* d
  ends the processing of the current line.
  use this, for example, to send a line with 
      multislog \
         -* +error: plocal0.error @127.0.0.1:514 d \
         +* plocal0.info @127.0.0.1:514

### Recommendations

* do not send multislog (and multilog) a TERM signal if you can avoid it. Just stop the message sender instead.
  There is no way multislog can ever be sure that it has read every message the sender sent.

* if you value your log entries, do not send them to syslog only. Right now there is no way the receiving syslog server can signal it got a message (a protocol limitation).

* TCP: a limited number of messages will be lost if a syslog server goes down, but messages can get lost, anyway (things can be worse with udp and unix sockets).

* TCP: multilog will hang if it cannot write to the remote syslog server.

* UDP: multilog will NOT hang if it cannot write to the remote syslog server, it will "just" lose the messages. 

* UNIX domain sockets: multilog will NOT hang, it will lose messages instead.

## License

[GPLv2](https://www.ohse.de/uwe/licenses/GPL-2)

## Installation
```
make install
```
will compile the program and install it into /usr/local/bin . You need a working go installation for it.

## Releases

* Version 0.2: 
  - add @unix, @tcp:, @udp:
* Version 0.1: 
  - the original release

## Security
If i can get you to install this without doing at least a short code audit, you do not need to worry about security.
Believe me: your computers security depends on luck alone if you install software just because you found it on github.

So, you did the audit? Yeah, my variable naming sucks.

Now continue, and audit the 2 non-standard packages imported.

That's too much work? Well, yes. I thought the same and didn't audit them, too. Possibly nobody ever will bother to do that.

Computer security in 2020/2021 sucks.
