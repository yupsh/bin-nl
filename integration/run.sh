#!/bin/sh
# Integration checks for yup-nl, run inside a Debian (GNU coreutils) container.
#
# parity ARGS... — yup-nl reading the sample on stdin must match GNU `nl`.
# assert WANT ARGS... — yup-nl must produce WANT exactly, used where yup-nl
#                       diverges from GNU by design (see cmd-nl COMPATIBILITY.md).
set -eu

fails=0
# Sample with a blank line, to exercise body-numbering styles.
sample='alpha

beta'

parity() {
	ours=$(printf '%s\n' "$sample" | yup-nl "$@" 2>/dev/null || true)
	gnu=$(printf '%s\n' "$sample" | nl "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  nl %s < stdin\n' "$*"
	else
		printf 'FAIL  parity  nl %s < stdin\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	shift
	got=$(printf '%s\n' "$sample" | yup-nl "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  nl %s < stdin\n' "$*"
	else
		printf 'FAIL  assert  nl %s < stdin\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# --body-numbering a (number all): exact parity with GNU `nl -b a`, including
# the width-6 right-justified field, the TAB separator, and how GNU renders the
# numbered blank line.
parity -b a
# --number-width (-w): narrower field, still matches GNU under -b a.
parity -b a -w 3
# --number-separator (-s): custom separator after the number.
parity -b a -s ': '
# --number-format (-n): left-justified, right-justified, right zero-padded.
parity -b a -n ln
parity -b a -n rn
parity -b a -n rz
# --starting-line-number (-v) and --line-increment (-i).
parity -b a -v 10
parity -b a -i 5
# Combined flags under -b a.
parity -b a -w 4 -n rz -s '|'

# Documented divergences (cmd-nl differs from GNU; assert exact yup-nl output):
#
# 1. Default body style is "a" (number every line) rather than GNU's "t"
#    (number non-empty lines only). With the blank middle line, yup-nl numbers
#    it where GNU would skip it.
assert "$(printf '     1\talpha\n     2\t\n     3\tbeta')"
# 2. Under -b t (number non-empty only), GNU pads the skipped blank line to the
#    number field width; yup-nl emits the bare line.
assert "$(printf '     1\talpha\n\n     2\tbeta')" -b t
# 3. Under -b n (number no lines), GNU emits a width-wide blank field with no
#    separator; yup-nl emits the blank field followed by the separator (TAB).
assert "$(printf '      \talpha\n      \t\n      \tbeta')" -b n

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
