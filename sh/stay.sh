#!/bin/sh

stay_exit() {
   echo "stay exit"
   kill $pid
   exit 1
}

interval=1
backoff() {
    sleep $interval
    if [ $interval -lt 5 ]; then
        interval=`expr $interval + 1`
    fi
}

trap stay_exit HUP INT QUIT TERM

pname=${1##*/}

while :
do
    pkill $pname
    $* &
    pid=$!
    wait $pid
    backoff
done
