package main

import (
	"time"
	"log"
	"os"
)

type fileOutputPrivate struct {
	fd *os.File

}
func newFileOutputDesc() *outputDesc {
	return &outputDesc{
		open: fileOutputOpen,
		sendStart: fileOutputSendStart,
		private: new(fileOutputPrivate),
	}
}
/*
 * "The action 
 *      =file
 *  replaces the contents of file with (the first 1000 bytes of) each selected line, 
 *  padded with newlines to 1001 bytes. There is no protection of file against power outages."
 */
func fileOutputSendStart(sc *scriptLine, buf []byte) {
	buf2:=make([]byte,0,1001)
	if sc.timestamp!=tsNone {
		buf2=append(buf2,curFormattedTimestamp...)
		buf2=append(buf2,' ')
	}
	for _, c := range buf {
		if len(buf2)==1000 {
			break
		}
		buf2=append(buf2,c)
	}
	for ; len(buf2)<1000; {
		buf2=append(buf2,'\n')
	}
	buf2=append(buf2,'\n')

	priv := (*fileOutputPrivate)(sc.desc.private.(*fileOutputPrivate))

	for {
		_, err:=priv.fd.Seek(0,0);
		if err==nil {
			break
		}
		log.Printf("unable to seek in %s, pausing: %v\n",sc.target, err)
		time.Sleep(time.Second);
	}
	writeLoop(sc.target, priv.fd, buf2);
}

func fileOutputOpen(sc *scriptLine) {
	priv := (*fileOutputPrivate)(sc.desc.private.(*fileOutputPrivate))
	for {
		fd, err := os.OpenFile(sc.target, os.O_WRONLY|os.O_CREATE, 0644)
		if err==nil {
			priv.fd=fd
			closeOnExec(fd);
			return
		}
	}
}

