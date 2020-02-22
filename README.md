# multislog

This is a rewrite of Dan Bernsteins [multilog](https://cr.yp.to/daemontools/multilog.html) with a number of extra features. If you don't use them, then the program should be perfectly compatible with multilog.

The main purpose of this tool is to store log messages in a way compatible to multilog, and be able to send the same messages to a central syslog server, too.

## Differences from multilog

multislog recognizes a number of additional script actions.

* @sysloghost:port
  send selected lines to a syslog server (by TCP).

* pfacility.severity
  set syslog facility ("user", "kern", "mail" and so on) and severity ("debug", "info", "crit" ...).

* imsgid
  set the syslog MSGID to `msgid`. Note that MSGID in the lingua of RFC5424 doesn't identify the message, but the type of a message (or rather a kind of classification).
  use a `-` if in doubt (but that's the default anyway).

* t (timestamp)
  does not need to be the first action anymore.

* t-
  turns the timestamping off.

### Recommendations

* do not send multislog (and multilog) a TERM signal if you can avoid it. Just stop the message sender instead.
  There is no way multislog can ever be sure that it has read every message the sender sent.

* do not send messages to syslog only if you value them. Right now there is no way the receiving syslog server can signal it got a message (a protocol limitation). The use of TCP means that a limited number of messages will be lost if a syslog server goes down, but messages can get lost, anyway.

* the use of TCP means that multilog will hang if it cannot write to the remote syslog server. This is the right behaviour in my use case, but possibly not in yours.

## License

[GPLv2](https://www.ohse.de/uwe/licenses/GPL-2)

## Installation
```
make install
```
will compile the program and install it into /usr/local/bin . You need a working go installation for it.

## Security
Well, if i can get you to install this without doing at least a short code audit, you do not need to worry about security. believe me: if you install software just because you found it on github your computers security depends on luck alone.

So, you did the audit? Yeah, my variable naming sucks.

Now continue, and audit the 2 packages imported.

That's too much work? Well, yes. I thought the same and didn't audit them, too. Possibly nobody ever will bother to do that. Maybe i should bite the bullet and not use the prometheus libraries, but right now i fail to find the motivation for that.

Computer security in 2020 sucks.
