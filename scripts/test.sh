#!/usr/bin/env bash

rm coverage.txt
set -e
echo "mode: atomic" > coverage.txt

for d in $(go list ./... | grep -v /vendor/ | grep -v /mnemosynerpc| grep -v /mnemosynetest); do
	go test -coverprofile=profile.out -covermode=atomic $d
	if [ -f profile.out ]; then
		tail -n +2 profile.out >> coverage.txt
		rm profile.out
	fi
done