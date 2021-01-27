#!/bin/sh
echo PID: $$
n=0
while :; do
	echo $n
	n=$((n+1))
	sleep 1
done
