#!/bin/sh

if test -n "$1"; then
    # need -R not -r to copy hidden files
    cp -R "$1/.paxd" /root
fi

# Create a unique log file name with timestamp
# LOG_FILE="/root/paxd_$(date +%Y%m%d_%H%M%S).log"

# LOCAL
HEDGEHOG_URL=${HEDGEHOG_URL:-https://127.0.0.1:40005}

mkdir -p /root/log
paxd start --hedgehog=$HEDGEHOG_URL --rpc.laddr tcp://0.0.0.0:26657 --api.enable=true --api.swagger=true --api.address tcp://0.0.0.0:1317 --trace #>> "$LOG_FILE" 2>&1
