#!/usr/bin/env bash
# test/compliance/test_ls.sh — compare korego ls vs system ls
set -uo pipefail
KOREGO=${KOREGO:-./korego}
PASS=0; FAIL=0

run_test() {
    local desc="$1"; shift
    local expected got exp_rc got_rc
    expected=$(ls "$@" 2>&1); exp_rc=$?
    got=$("$KOREGO" ls "$@" 2>&1); got_rc=$?
    if [ "$exp_rc" = "$got_rc" ]; then
        ((PASS++))
        echo "PASS: $desc"
    else
        ((FAIL++))
        echo "FAIL: $desc (exit: expected=$exp_rc got=$got_rc)"
    fi
}

run_test "ls /tmp"          /tmp
run_test "ls -a /tmp"       -a /tmp
run_test "ls -1 /tmp"       -1 /tmp
run_test "ls nonexistent"   /this/does/not/exist/korego || true

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
