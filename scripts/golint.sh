#!/usr/bin/env bash

EXIT=0

# Check gofmt
echo "==> Checking that code complies with golint requirements..."
packages=$@

for package in $@; do
    OUT=$(golint $package)

    if [ "$OUT" != "" ]; then
        echo -e "$OUT"
        EXIT=1
    fi
done

exit $EXIT
