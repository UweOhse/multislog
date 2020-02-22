#!/bin/sh

# not tested here:
# multilog handles TERM (tested manually)
# multilog handles ALRM (tested manually)
# multilog handles out-of-memory (next to impossible to do right)
# multilog t produces the right time (next to impossible to do right)
# multilog closes descriptors properly (system dependencies ahead)


PATH=".:$PATH"

mkdir test.tmp.$$ || exit 1
cd test.tmp.$$
function cleanup()
{
	if test "$COVER" = 1 ; then
		mv cover.out ..
	fi
	cd ..
	rm -rf test.tmp.$$
}
trap cleanup EXIT

M="../multislog "
if test "$COVER" = "1" ; then
	M="../multislog.test -test.coverprofile=cover1.out"
	echo | $M -- --version
	mv cover1.out cover.out
fi
function docover()
{
	if test "$COVER" = 1 ; then
		gocovmerge cover.out cover1.out >t.out
		mv t.out cover.out

	fi
}

echo '--- multilog prints nothing with no actions'
( echo one; echo two ) | $M ; echo $?
docover

echo '--- multilog e prints to stderr'
( echo one; echo two ) | $M  e 2>&1; echo $?
docover

echo '--- multilog inserts newline after partial final line'
( echo one; echo two | tr -d '\012' ) | $M  e 2>&1; echo $?
docover

echo '--- multilog handles multiple actions'
( echo one; echo two ) | $M  e e 2>&1; echo $?
docover

echo '--- multilog handles wildcard -'
( echo one; echo two ) | $M  '-*' e 2>&1; echo $?
docover

echo '--- multilog handles literal +'
( echo one; echo two ) | $M  '-*' '+one' e 2>&1; echo $?
docover

echo '--- multilog handles long lines for stderr'
echo 0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678 \
	| $M  e 2>&1; echo $?
docover

echo 01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789 \
	| $M  e 2>&1; echo $?
docover

echo 012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890 \
	| $M  e 2>&1; echo $?
docover

echo '--- multilog handles status files'
rm -f test.status
( echo one; echo two ) | $M  =test.status; echo $?
docover
uniq -c < test.status | sed 's/[        ]*[     ]/_/g'

echo '--- multilog t has the right format on stderr'
( echo ONE; echo TWO ) | $M  t e 2>&1 | sed 's/[0-9a-f]/x/g'
docover

echo '--- multilog t has the right format in status files'
( echo ONE; echo TWO ) | $M  t =test.status ; echo $?
cat test.status | sed 's/[0-9a-f]/x/g' |head -1
docover

echo '--- multilog t ./x has the right format'
( echo ONE; echo TWO ) | $M  t ./x 2>&1 ; echo $?
docover
sed 's/[0-9a-f]/x/g' < x/current
rm -rf x

echo '--- multilog ./x inserts newline after partial final line'
( echo one; echo two | tr -d '\012' ) | $M  ./x ; echo $?
docover
cat x/current
rm -rf x

echo '--- multilog ./x handles overlong lines'
( echo 1234567890abcdef ; echo 123456; echo 123456123456 ; echo 123 ) | MULTISLOG_BUFLEN=6 $M  ./x ; echo $?
docover
cat x/current
rm -rf x

echo '--- multilog t ./x handles overlong lines'
( echo ABCDEFGIHJKLMNOP ; echo ABCDEF; echo ABCDEFABCDEF ; echo ABC ) | MULTISLOG_BUFLEN=6 $M t ./x ; echo $?
docover
sed 's/[0-9a-f]/x/g' < x/current
rm -rf x

echo '--- multilog t ./x handles overlong lines'
( echo ABCDEFGIHJKLMNOP ; echo ABCDEF; echo ABCDEFABCDEF ; echo ABC ) | MULTISLOG_BUFLEN=6 $M t ./x ; echo $?
docover
sed 's/[0-9a-f]/x/g' < x/current
rm -rf x

echo '--- multilog ./x s4096 works'
( for i in `seq 1 257` ; do echo 0123456789abcde ; done ) |  $M s4096 ./x ; echo $?
docover
cat x/current
od -tx1 x/\@*
rm -rf x

echo '--- multilog ./x s4096 n3 works'
( for i in `seq 1 513` ; do echo 0123456789abcde ; done ) |  $M n3 s4096 ./x ; echo $?
docover
cat x/current
od -tx1 x/@*
rm -rf x
echo '--- multilog ./x s4096 n2 works'
( for i in `seq 1 513` ; do echo 0123456789abcde ; done ) |  $M n2 s4096 ./x ; echo $?
docover
cat x/current
od -tx1 x/@*
rm -rf x

echo '--- multilog ./x s8192 n2 works'
( for i in `seq 1 1025` ; do echo 0123456789abcde ; done ) |  $M n2 s8192 ./x ; echo $?
docover
cat x/current
od -tx1 x/@*
rm -rf x

echo '--- multilog ./x s8192 n2 works'
( for i in `seq 1 1025` ; do echo 0123456789abcde ; done ) |  $M n2 s8192 ./x ; echo $?
docover
cat x/current
od -tx1 x/@*
rm -rf x

echo '--- multilog ./x !sed work'
( for i in `seq 1 1025` ; do echo 0123456789abcde ; done ) |  $M n2 s4096 "!sed s/e/f/g" ./x ; echo $?
docover
cat x/current
od -tx1 x/@*

# deliberate change from multilog behaviour: multilog deletes state at start, which at least is
# undocumentated behaviour, but at most possibly a bug.
echo '--- multilog ./x !sed works with state'
rm x/@* x/current
echo 'eeeeee' >x/state
( for i in `seq 1 256` ; do echo 0123456789abcde ; done ) | \
	$M n2 s4096 "!cat ; sed s/e/f/g <&4 >&5" ./x ; echo $?
	docover
echo
echo -n current: ; cat x/current ; echo
echo -n state: ; cat x/state ; echo
echo "od:"
od -tx1 x/@*
rm -rf x

echo '--- multilog ./x ! repeats if processor fails'
( for i in `seq 1 257` ; do echo 0123456789abcde ; done ) | \
	$M n2 s4096 "!if test -f testflag ; then echo done ; exit 0; else echo failed; touch testflag ; exit 1; fi " ./x ; echo $?
	docover
cat x/current
cat x/@*
test -f x/testflag && echo "testflag set"
od -tx1 x/@*
rm -rf x

echo '--- multilog ./x continues cleanly after stop'
( for i in `seq 1 1` ; do echo 0123456789abcde ; done ) | $M n2 s4096 ./x ; echo $?
docover
test -x x/current || echo "current has no x bit"
test -x x/current && echo "current has x bit"
( for i in `seq 1 1` ; do echo 0123456789abcde ; done ) | $M n2 s4096 ./x ; echo $?
docover
od -tx1 x/current
rm -rf x

echo '--- multilog ./x continues cleanly after crash'
( for i in `seq 1 1` ; do echo 0123456789abcde ; done ) | $M n2 s4096 ./x ; echo $?
docover
chmod -x x/current
( for i in `seq 1 1` ; do echo 0123456789abcde ; done ) | $M n2 s4096 ./x ; echo $?
docover
test -x x/current || echo "current has no x bit"
test -x x/current && echo "current has x bit"
echo current:; od -tx1 x/current
echo .u: ; od -tx1 x/\@*.u
rm -rf x

echo '--- multilog n1 is corrected to 2'
( echo 111111111111111; echo aaaaaaaaaaaaa ; echo ZZZZZZZZZZZZZZZ ) | \
	MULTISLOG_BUFSIZE=15 $M n1 s10 ./x 2>&1 ; echo $?
docover
echo current:; od -tx1 x/current
echo .s: ; od -tx1 x/\@*.s

# all options.
echo '--- multilog -a* deselects'
echo 'anything' | $M '-a*' e 2>&1
docover

echo '--- multilog +* selects everything'
echo 'anything' | $M '-*' '+*' e 2>&1
docover

echo '--- multilog handles negative size or number'
echo 'anything' | $M s-999 2>&1 ; echo $?
docover
echo 'anything' | $M n-999 2>&1 ; echo $?
docover
echo '--- multilog parses facility.severity'
echo 'anything' | $M pkern.emerg 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.emer 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.emergx 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.alert 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.aler 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.alertx 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.crit 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.cri 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.critx 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.err 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.er 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.errx 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.warning 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.warnin 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.warningx 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.info 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.inf 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.infox 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.debu 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.debugx 2>&1 ; echo $? ; docover
echo 'anything' | $M pkern.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pker.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pkernx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M puser.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M puse.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M puserx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pmail.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pmai.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pmailx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pdaemon.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pdaemo.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pdaemonx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pauth.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M paut.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pauthx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M psyslog.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M psyslo.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M psyslogx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plpr.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plp.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plprx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pnews.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pnew.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pnewsx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M puucp.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M puuc.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M puucpx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pcron.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pcro.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pcronx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pauthpriv.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pauthpri.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pauthprivx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pftp.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pft.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pftpx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pntp.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pnt.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pntpx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M paudit.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M paudi.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M pauditx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plpr.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plp.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plprx.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plocal0.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plocal.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plocal0x.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M plocal8.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M p.debug 2>&1 ; echo $? ; docover
echo 'anything' | $M p.debu 2>&1 ; echo $? ; docover
echo 'anything' | $M puser 2>&1 ; echo $? ; docover
echo 'anything' | $M puse 2>&1 ; echo $? ; docover

echo '--- multilog t t- works'
echo 0123456789abcde | $M t e 2>&1 | sed 's/[0-9a-f]/x/g' ; echo $?
echo 0123456789abcde | $M t t- e 2>&1 ; echo $?
docover

echo '--- multilog tbad fails'
echo 0123456789abcde | $M tbad e 2>&1 ; echo $?
docover
echo '--- multilog i is known'
echo 0123456789abcde | $M imsgid e 2>&1 ; echo $?
docover

echo '--- multilog $ is unknown'
echo 0123456789abcde | $M \$ e 2>&1 ; echo $?
docover

echo '--- multilog d(one) works'
echo 0123456789abcde | $M d e 2>&1 ; echo $?
docover

echo '--- multilog d(one) works if selected'
echo oldstuff, must not change >test.status
echo 0123456789abcde | $M "-*" +0* e d =test.status 2>&1 ; echo $?
cat test.status | head -1
docover

echo '--- multilog d(one) does not break loop if unselected'
echo oldstuff, must change >test.status
echo 0123456789abcde | $M "-*" e d +0* =test.status 2>&1 ; echo $?
cat test.status | head -1
docover

if test "$COVER" = "1" ; then
	exit 0
fi


