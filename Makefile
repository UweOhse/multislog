PREFIX=/usr/local
VERSION=0.2
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
	staticcheck .
	gosec -conf ~/etc/gosec.conf.json ./...
	gocritic check -enable='#diagnostic,#experimental,#performace,#style,#opionionated' \
		-disable octalLiteral ./...
	errcheck

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

clean:
	rm -f cover.out cover.html multislog multislog.test

build-release: all check-release-info check-git-tag
	git archive --prefix multislog-$(VERSION)/ -o tmp.multislog-$(VERSION).tar.gz v$(VERSION)
	rm -rf releasebuild
	mkdir releasebuild
	(cd releasebuild; tar xzf ../tmp.multislog-$(VERSION).tar.gz --strip-components 1 )
	(cd releasebuild ; make check)
	rm -rf releasebuild
	mv tmp.multislog-$(VERSION).tar.gz multislog-$(VERSION).tar.gz

check-release-info:
	@fgrep "Version $(VERSION):" README.md >/dev/null || ( echo "no Release info in README.me" ; exit 1 )

check-git-tag:
	@git tag -l v$(VERSION) | grep '^v$(VERSION)$$' >/dev/null || ( echo "no Release tag in git" ; exit 1 )

