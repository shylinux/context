#!/usr/bin/env bash

while true; do
    bench && break
    echo "bench run error"
    echo "restarting..."
    sleep 3
done
