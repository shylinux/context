#!/usr/bin/env bash

[ `uname` = "Darwin" ] && bench=bin/bench.darwin
[ `uname` = "Linux" ] && bench=bin/bench.linux64
[ -e "$bench" ] || bench=bench

echo "bench: $bench"> var/boot.log

while true; do
    $bench stdio 2>var/error.log && break
    echo "bench run error"
    echo "restarting..."
    sleep 3
done
