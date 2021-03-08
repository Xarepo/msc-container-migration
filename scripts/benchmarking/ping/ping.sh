#!/bin/sh

# Send individiual ping packages as fast as possible.
while :; do
	ping -W 0.1 -c 1 172.17.0.1 &
	sleep 0.01
done
