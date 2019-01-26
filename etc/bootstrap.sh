#!/usr/bin/env bash

bench=bench
[ -e bin/bench ] && bench=bin/bench

echo "bench: $bench"> var/boot.log

while true; do
    $bench 2>var/error.log && break
    echo "bench run error"
    echo "restarting..."
    sleep 3
done
