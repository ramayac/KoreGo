#!/usr/bin/env bash
# test/compliance/test_basename_dirname.sh
set -uo pipefail
KOREGO=${KOREGO:-./korego}
PASS=0; FAIL=0

run_test() {
    local desc="$1" cmd="$2"; shift 2
    local expected got exp_rc got_rc
    expected=$($cmd "$@" 2>&1); exp_rc=$?
    got=$("$KOREGO" "$cmd" "$@" 2>&1); got_rc=$?
    if [ "$expected" = "$got" ] && [ "$exp_rc" = "$got_rc" ]; then
        ((PASS++)); echo "PASS: $desc"
    else
        ((FAIL++)); echo "FAIL: $desc => expected=$expected got=$got"
    fi
}

run_test "basename /path/to/file.txt"    basename /path/to/file.txt
run_test "basename with suffix"          basename /path/to/file.txt .txt
run_test "dirname /path/to/file.txt"     dirname  /path/to/file.txt
run_test "dirname bare name"             dirname  file.txt

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
