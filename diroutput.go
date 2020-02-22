package main

import (
	"log"
	"syscall"
	"errors"
	"os"
	"bufio"
	"strings"
	"time"
	"golang.org/x/sys/unix"
)

type dirOutputPrivate struct {
	fd *os.File
	bytesWritten int64
	writer *bufio.Writer

	workDir *os.File
	fdLock *os.File
}
func newDirOutputDesc() *outputDesc {
	return &outputDesc{
		quit: dirQuit,
		flush: dirFlush,
		sendStart: dirOutputSendStart,
		sendContinue: dirOutputSendContinue,
		sendEnd: dirOutputSendEnd,
		open: dirOpen,
		private: new(dirOutputPrivate),
	}
}


func dirFinish(sc *scriptLine, fn, code string) {
	var st os.FileInfo
	var err error
	for {
		st, err = os.Stat(fn)
		if err==nil {
			break
		}
		if errors.Is(err,os.ErrNotExist) {
			return
		}
		log.Printf("unable to stat %s, pausing: %v\n",fn, err)
		time.Sleep(time.Second);
	}
	x:=st.Sys().(*syscall.Stat_t)
	nlink:=x.Nlink
	if nlink == 1 {
		for {
			buf:=make([]byte,0)
			buf=append(buf, curTaiaTimestamp...)

			buf=append(buf,'.');
			buf=append(buf,code[0])
			err=os.Link(fn,string(buf))
			if err==nil {
				break
			}
			log.Printf("unable to link %s to %s, pausing: %v\n",fn, buf, err)
			time.Sleep(time.Second);
		}
	}
	removeLoop(sc, fn)
	for {
		fit, err:=dirFilesfit(sc)
		if err==nil {
			if fit {
				return
			}
		} else {
			log.Printf("unable to read or delete in %s, pausing: %v\n",sc.target, err)
			time.Sleep(time.Second);
		}
	}
}
func dirOpenFinish(priv *dirOutputPrivate, fd *os.File, sz int64) {
	closeOnExec(fd);
	priv.fd=fd
	priv.bytesWritten=sz
	priv.writer=bufio.NewWriterSize(fd,Buflen+4096)
}

func chdirLoop(fd *os.File, fn string) {
	for {
		err := fd.Chdir();
		if err==nil {
			return
		}
		log.Printf("unable to fchdir to %s, pausing: %v\n",fn, err)
		time.Sleep(time.Second);
	}
}
func closeLogger(fd *os.File, fn string) {
	err := fd.Close();
	if err==nil {
		return
	}
	// cannot retry: the kernel may have freed / reused the resource (think other threads).
	log.Printf("unable to close %s: %v\n",fn, err)
}

// log.Fatal really is overused, but there is not much we can do.
func dirOpen(sc *scriptLine) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))

	oldDir, err:=os.Open(".");
	if err!=nil {
		log.Panicf("unable to open current directory: %v\n",err);
	}
	closeOnExec(oldDir);

	err = os.Mkdir(sc.target,0700);
	if err!=nil && !errors.Is(err,os.ErrExist) {
		log.Panicf("unable to mkdir target directory %s: %v\n",sc.target, err);
	}
	priv.workDir, err =os.Open(sc.target);
	if err !=nil{
		log.Panicf("unable to open target directory %s: %v\n",sc.target, err);
	}
	closeOnExec(priv.workDir);
	chdirLoop(priv.workDir, sc.target);

	defer chdirLoop(oldDir, ".")

	fdLock, err := openAppend("lock")
	if err !=nil{
		log.Panicf("unable to create lock %s: %v\n",sc.target, err);
	}
	err = unix.Flock(int(fdLock.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		log.Panicf("unable to lock %s: %v\n",sc.target, err);
        }
	priv.fdLock=fdLock
	closeOnExec(priv.fdLock);

	st, err := os.Stat("current");
	if err!=nil && !errors.Is(err,os.ErrNotExist) {
		log.Panicf("unable to stat %s/current: %v\n",sc.target, err);
	}
	if err == nil {
		if (st.Mode() & 0100)!=0 {
			// reuse old current file
			fd, err := openAppend("current")
			if err != nil {
				log.Panicf("unable to append to %s/current: %v\n",sc.target, err);
			}
			err=fd.Chmod(0644);
			if err!=nil {
				log.Panicf("unable to set mode of %s/current: %v\n", sc.target, err);
			}
			dirOpenFinish(priv, fd, st.Size())
			return
		}
	}

	// why was this done: os.Remove("state");
	removeLoop(sc, "newstate");

	flagProcessed:=false
	st, err = os.Stat("processed");
	if err!=nil && !errors.Is(err,os.ErrNotExist) {
		log.Panicf("unable to stat %s/processed: %v\n",sc.target, err);
	}
	if err==nil && (st.Mode()&0100)!=0 {
		flagProcessed=true
	}
	if flagProcessed {
		removeLoop(sc, "previous")
		dirFinish(sc,"processed","s")
	} else {
		removeLoop(sc, "processed")
		dirFinish(sc,"previous","u")
	}
	dirFinish(sc,"current","u")

	fd, err := openAppend("state");
	if err != nil && errors.Is(err,os.ErrNotExist) {
		fd, err = openTrunc("state");
	}
	if err != nil {
		log.Panicf("unable to open %s/state: %v\n",sc.target, err)
	}
	err = fd.Close()
	if err != nil { // ECANTHAPPEN, really
		log.Panicf("unable to close %s/state: %v\n",sc.target, err)
	}

	fd, err = openAppend("current");
	if err != nil {
		log.Panicf("unable to write to %s/current: %v\n",sc.target, err)
	}
	err = fd.Chmod(0644);
	if err != nil {
		log.Panicf("unable to set mode of %s/current: %v\n",sc.target, err)
	}
	dirOpenFinish(priv, fd, 0)
}

func dirFilesfit(sc *scriptLine) (bool,error) {

	dir, err := os.Open(".");
	if err!=nil { return false, err}
	defer closeLogger(dir, ".");
	var count uint64
	oldest:="@z"

	names, err := dir.Readdirnames(0); // all
	if err!=nil {
		return false, err
	}
	for _, fn := range names {
		if fn[0]=='@' && len(fn)>=25 {
			count++
		}
		if strings.Compare(fn, oldest) <0 {
			oldest=fn
		}
	}
	if count < sc.maxFiles {
		return true, nil
	}
	err = os.Remove(oldest)
	return false,err;
}
func dirWriterLoop(sc *scriptLine, buf[]byte) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))
        for {
                n, err:=priv.fd.Write(buf)
		priv.bytesWritten+=int64(n)
                if err == nil {
                        break;
                }
                log.Printf("unable to write to %s/current, pausing: %v\n",sc.target, err);
                buf=buf[n:]
                time.Sleep(time.Second);
        }
}

func dirOutputSendStart(sc *scriptLine, buf []byte) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))
	if priv.bytesWritten>=int64(sc.maxSize) {
		dirFullCurrent(sc)
	}
	if sc.doTimestamp {
		dirWriterLoop(sc, curTaiaTimestamp)
		dirWriterLoop(sc, []byte{' '});
	}
	dirWriterLoop(sc, buf);
}
func dirOutputSendContinue(sc *scriptLine, buf []byte) {
	dirWriterLoop(sc, buf);
}
func dirOutputSendEnd(sc *scriptLine) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))
	dirWriterLoop(sc, []byte{'\n'});
	dirJustFlush(sc);
	if priv.bytesWritten>=int64(sc.maxSize) {
		dirFullCurrent(sc)
	}
}


func dirStartProcessor(sc *scriptLine) (*os.Process, error) {
/*
	Not done:
		sig_uncatch(sig_term);
		sig_uncatch(sig_alarm);
		sig_unblock(sig_term);
		sig_unblock(sig_alarm);
	Need to fork() and exec() to handle that (fork, uncatch/unblock, exec, like in C), but
	since there is no simple fork() in golang, i really would need to write something 
	like syscall.ForkExec without the output catching, but that's a real mess.
	Signal handling doesn't fit golang.

	Anyway, the OS resets both signals to the default behaviour anyway, so uncatching 
	should not be needed, anyway (at least on any OS go runs one).
	That leaves the blocking.

	2020-02-22: ok, i removed the slightly horrible sigblock-hack i used in main, and
	do not block signals anymore. This should solve the problem, and not open any
	can of worms. multilog did the sigblock so it only had to deal with signals in the
	stdin-read-path.
*/

	args:=make([]string,4)
	args[0]="sh"
	args[1]="-c"
	args[2]=sc.processor

	finfo:=make([]*os.File,0)
	// stdin
	fd, err := openRead("previous")
	if err!=nil {
		return nil, err
	}
	finfo=append(finfo,fd)
	defer closeLogger(fd, "previous")

	// stdout
	fd, err = openTrunc("processed")
	if err!=nil {
		return nil, err
	}
	finfo=append(finfo,fd)
	defer closeLogger(fd, "processed")

	// 2
	finfo=append(finfo,os.Stderr)
	// 3
	finfo=append(finfo,nil)

	//4
	fd, err = openRead("state")
	if err!=nil {
		return nil, err
	}
	finfo=append(finfo,fd)
	defer closeLogger(fd, "state")

	//4
	fd, err = openTrunc("newstate")
	if err!=nil {
		return nil, err
	}
	finfo=append(finfo,fd)
	defer closeLogger(fd, "newstate")

	attr:=new(os.ProcAttr)
	attr.Files=finfo

	p, err := os.StartProcess("/bin/sh",args, attr);
	if err != nil {
		return nil, err
	}
	return p,nil;
}

func dirOpenLoop(name string) *os.File {
	for {
		handle, err:=os.Open(".");
		if err==nil {
			return handle
		}
		log.Printf("unable to open current ., pausing: %v\n",err);
		time.Sleep(time.Second)
	}
}
func dirFullCurrent(sc *scriptLine) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))

	oldDir := dirOpenLoop(".")
	var err error
	defer chdirLoop(oldDir,".")
	closeOnExec(oldDir);

	for  {
		err = priv.workDir.Chdir();
		if err==nil {
			break
		}
		log.Printf("unable to chdir to %s, pausing: %v\n",sc.target, err)
		time.Sleep(time.Second)
	}
	fsyncLoop(sc, "current", priv.fd)
	_ = priv.fd.Close()

	renameLoop(sc,"current","previous");

	fd:=openAppendLoop(sc,"current")
	dirOpenFinish(priv, fd, 0)

	chmodLoop(sc,"current",priv.fd, 0644);
	chmodLoop(sc,"previous",priv.fd, 0744);

	if sc.processor=="" {
		dirFinish(sc,"previous","s");
		return
	}
	// processor handling
	for {
		p, err := dirStartProcessor(sc)
		var state *os.ProcessState
		if err != nil {
			log.Printf("unable to start processor of %s/current, pausing: %v\n",sc.target, err)
			time.Sleep(time.Second)
			continue
		}
		for {
			state, err = p.Wait()
			if err==nil {
				break;
			}
			log.Printf("wait for processor of %s/current failed, pausing: %v\n",sc.target, err)
			time.Sleep(time.Second)
		}
		if state.Success() {
			break
		}
		ec := state.ExitCode()
		if -1==ec {
			log.Printf("processor of %s/current crashed, pausing\n",sc.target)
			time.Sleep(time.Second)
			continue
		}
		log.Printf("processor of %s/current exited with code %d, pausing\n",sc.target, ec)
		time.Sleep(time.Second)
	}

	fd = openAppendLoop(sc,"processed")
	fsyncLoop(sc,"processed",fd)
	chmodLoop(sc,"processed",fd,0744)
	_ = fd.Close()

	fd = openAppendLoop(sc,"newstate")
	fsyncLoop(sc,"newstate",fd)
	_ = fd.Close()

	removeLoop(sc,"previous")
	renameLoop(sc,"newstate", "state")

	dirFinish(sc,"processed","s");
}
func openAppendLoop(sc *scriptLine, fn string) *os.File {
	for {
		fd, err := openAppend(fn);
		if err==nil {
			return fd
		}
		log.Printf("unable to create %s/%s, pausing: %v\n",sc.target,fn, err)
		time.Sleep(time.Second)
	}
}
func fsyncLoop(sc *scriptLine, fn string, fd *os.File) {
	for {
		err := fd.Sync()
		if err==nil {
			return
		}
		log.Printf("unable to fsync %s/%s, pausing: %v\n",sc.target, fn, err)
		time.Sleep(time.Second)
	}
}
func chmodLoop(sc *scriptLine, fn string, fd *os.File, mode os.FileMode) {
	for {
		err := fd.Chmod(mode)
		if err==nil {
			return
		}
		log.Printf("unable to set mode of %s/%s, pausing: %v\n",sc.target, fn, err)
		time.Sleep(time.Second)
	}
}
func removeLoop(sc *scriptLine, fn string) {
	for {
		err := os.Remove(fn)
		if err==nil || errors.Is(err,os.ErrNotExist) {
			return
		}
		log.Printf("unable to unlink %s/%s, pausing: %v\n",sc.target, fn, err)
		time.Sleep(time.Second)
	}
}
func renameLoop(sc *scriptLine, from, to string) {
	for {
		err := os.Rename(from,to)
		if err==nil {
			return
		}
		log.Printf("unable to rename %s/%s to %s, pausing: %v\n",sc.target, from, to, err)
		time.Sleep(time.Second)
	}
}

func dirJustFlush(sc *scriptLine) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))
	for {
		err:=priv.writer.Flush()
		if err==nil {
			break
		}
		log.Printf("unable to flush %s/current, pausing: %v\n",sc.target, err)
		time.Sleep(time.Second)
	}
}
func dirQuit(sc *scriptLine) {
	priv := (*dirOutputPrivate)(sc.desc.private.(*dirOutputPrivate))
	dirJustFlush(sc);
	fsyncLoop(sc, "current", priv.fd)
	chmodLoop(sc, "current", priv.fd, 0744)
}

func dirFlush(sc *scriptLine) {
	dirJustFlush(sc);
	dirFullCurrent(sc)
}

