#!/usr/bin/env bash
# test/compliance/test_cat.sh — compare korego cat vs system cat
set -uo pipefail
KOREGO=${KOREGO:-./korego}
PASS=0; FAIL=0
TMPFILE=$(mktemp)
echo -e "line one\nline two\nline three" > "$TMPFILE"
trap 'rm -f "$TMPFILE"' EXIT

run_test() {
    local desc="$1"; shift
    local expected got exp_rc got_rc
    expected=$(cat "$@" 2>&1); exp_rc=$?
    got=$("$KOREGO" cat "$@" 2>&1); got_rc=$?
    if [ "$expected" = "$got" ] && [ "$exp_rc" = "$got_rc" ]; then
        ((PASS++))
        echo "PASS: $desc"
    else
        ((FAIL++))
        echo "FAIL: $desc (exit: expected=$exp_rc got=$got_rc)"
        diff <(echo "$expected") <(echo "$got") || true
    fi
}

run_test "cat file"           "$TMPFILE"
run_test "cat -n file"        -n "$TMPFILE"

expected_rc=1
got_rc=0
"$KOREGO" cat /no/such/file >/dev/null 2>&1 || got_rc=$?
if [ "$got_rc" -eq "$expected_rc" ]; then
    ((PASS++))
    echo "PASS: cat nonexistent"
else
    ((FAIL++))
    echo "FAIL: cat nonexistent (exit: expected=$expected_rc got=$got_rc)"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
