package main

import (
	"os"
)

func newStderrOutputDesc() *outputDesc {
	return &outputDesc{
		sendStart: stderrOutputSendStart,
	}
}



// djb: The action e prints (the first 200 bytes of) each selected line to stderr.
// reality: if is shortens the line then it also appends ...
func stderrOutputSendStart(sc *scriptLine, inbuf []byte) {
	// log.Printf("stderrWriter %v %v\n",sc, buf)
	buf := make([]byte,0,200)
	if sc.doTimestamp {
		buf=append(buf, curTaiaTimestamp...)
		buf=append(buf,' ')
	}
	buf=append(buf,inbuf...)
	if len(buf)>200 {
		buf=buf[:200]
		buf=append(buf,'.');
		buf=append(buf,'.');
		buf=append(buf,'.');
	}
	buf=append(buf,'\n');
	writeLoop(sc.target, os.Stderr, buf);
}
