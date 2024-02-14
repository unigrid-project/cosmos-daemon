#!/bin/sh

if test -n "$1"; then
    # need -R not -r to copy hidden files
    cp -R "$1/.paxd" /root
fi

mkdir -p /root/log
paxd start --rpc.laddr tcp://0.0.0.0:26657 --api.enable=true --api.swagger=true --api.address tcp://0.0.0.0:1317 --trace
