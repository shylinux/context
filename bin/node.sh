#! /bin/bash

export ctx_box=${ctx_box:="http://localhost:9095"}

[ -f bin/boot.sh ] && source bin/boot.sh || source boot.sh
