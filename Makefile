PREFIX=/usr/local
VERSION=0.1
G=`git rev-list HEAD | head -1`

SRC=main.go version.go match.go parse.go timestamp.go ioutils.go \
    diroutput.go fileoutput.go stderroutput.go syslogoutput.go

all: multislog

multislog: $(SRC)
	go build -o $@ $^
openfds: openfds.go
	go build -o $@ $^

install: multislog
	install -t $(PREFIX)/bin $^ 

version.go: Makefile version.in
	sed -e 's/VVVVV/$(VERSION)/g' -e 's/GGGGG/'$G'/g' <version.in >$@.t
	mv $@.t $@

style:
	go vet $(SRC)
	errcheck
	staticcheck .
	gocritic check -enable='#diagnostic,#experimental,#performace,#style,#opionionated' ./...
	gosec ./...

check test: multislog
	sh test.sh >test.out 
	diff test.expect test.out
cover: cover.out

cover.out: multislog.test
	COVER=1 sh test.sh >/dev/null
	go tool cover -func cover.out

multislog.test: $(SRC)
	go test -coverpkg="./..." -c -tags testrunmain .

coverhtmlupload: cover.html
	scp cover.html uwe@ohse.de:oldweb/uwe/misc/multislog.cover.html

cover.html: cover.out
	go tool cover --html=cover.out -o cover.html
	
