package main

import (
	"time"
	"strconv"
	"fmt"
	"log"
)
const hex = "0123456789abcdef"

func timestamp(mode int, now time.Time) ([]byte) {
	out:=[]byte{}
	switch mode {
	case tsNone:
	case tsTAI64:
		out=timestamp_tai64(now)
	case tsEpoch:
		out=timestamp_epoch(now)
	case tsEpochMs:
		out=timestamp_epochMs(now)
	case tsEpochUs:
		out=timestamp_epochUs(now)
	case tsEpochNs:
		out=timestamp_epochNs(now)
	case tsRFC3339:
		out=timestamp_RFC3339(now)
	default:
		log.Fatalf("ECANTHAPPEN: timestamps mode is %d\n",mode)
	}
	return out
}
func timestamp_epoch(now time.Time) ([]byte) {
	sec:=now.Unix()
	return []byte(strconv.FormatInt(sec,10))
}
func timestamp_epochMs(now time.Time) ([]byte) {
	sec:=now.Unix()
	nano:=now.Nanosecond()
	return []byte(fmt.Sprintf("%d%03d",sec,nano/1000000))
}
func timestamp_epochUs(now time.Time) ([]byte) {
	sec:=now.Unix()
	nano:=now.Nanosecond()
	return []byte(fmt.Sprintf("%d%06d",sec,nano/1000))
}
func timestamp_epochNs(now time.Time) ([]byte) {
	sec:=now.Unix()
	nano:=now.Nanosecond()
	return []byte(fmt.Sprintf("%d%09d",sec,nano))
}
func timestamp_RFC3339(now time.Time) ([]byte) {
	return []byte(time.Format(time.RFC3339))
}


func timestamp_tai64(now time.Time) ([]byte) {
	buf:=make([]byte,12)
	out:=make([]byte,25)
	sec:=now.Unix()
	sec+=4611686018427387914;
	nano:=now.Nanosecond()

	buf[7]=byte(sec&255); sec>>= 8;
	buf[6]=byte(sec&255); sec>>= 8;
	buf[5]=byte(sec&255); sec>>= 8;
	buf[4]=byte(sec&255); sec>>= 8;
	buf[3]=byte(sec&255); sec>>= 8;
	buf[2]=byte(sec&255); sec>>= 8;
	buf[1]=byte(sec&255); sec>>= 8;
	buf[0]=byte(sec&255)
	buf[11]=byte(nano&255) ; nano>>=8
	buf[10]=byte(nano&255) ; nano>>=8
	buf[9]=byte(nano&255) ; nano>>=8
	buf[8]=byte(nano&255)

	out[0]='@'

	for i:=0; i<12; i++ {
		t:=(buf[i] /16)&15;
		out[i*2+1]=hex[t]
		t=(buf[i])&15;
		out[i*2+2]=hex[t]
	}
	return out
}


