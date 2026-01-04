#!/bin/bash
set -e

FUZZTIME="${FUZZTIME:-30s}"

for pkg in $(go list ./... | xargs -I{} sh -c 'go test -list "^Fuzz" {} 2>/dev/null | grep -q "^Fuzz" && echo {}'); do
    for fuzz in $(go test -list "^Fuzz" "$pkg" 2>/dev/null | grep "^Fuzz"); do
        echo "Fuzzing $fuzz in $pkg"
        go test -fuzz="$fuzz" -fuzztime="$FUZZTIME" "$pkg"
    done
done
