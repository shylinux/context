#!/usr/bin/env bash

bench=bench
[ `uname` = "Darwin" ] && bench=bin/bench.darwin
[ `uname` = "Linux" ] && bench=bin/bench.linux64
[ -e "$bench" ] || bench=bench

while true; do
    $bench stdio && break
    echo "bench run error"
    echo "restarting..."
    sleep 3
done
